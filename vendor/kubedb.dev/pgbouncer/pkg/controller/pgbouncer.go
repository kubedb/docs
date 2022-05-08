/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	"kubedb.dev/apimachinery/pkg/phase"
	validator "kubedb.dev/pgbouncer/pkg/admission"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

func (c *Controller) runPgBouncer(key string) error {
	klog.V(5).Infoln("started processing, key:", key)
	obj, exists, err := c.pbInformer.GetIndexer().GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		klog.Infof("PgBouncer %s does not exist anymore\n", key)
		klog.V(5).Infof("PgBouncer %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a PgBouncer was recreated with the same name
		db := obj.(*api.PgBouncer).DeepCopy()

		if db.DeletionTimestamp != nil {
			if core_util.HasFinalizer(db.ObjectMeta, kubedb.GroupName) {
				if err := c.terminate(db); err != nil {
					klog.Errorln(err)
					return err
				}
				_, _, err := util.PatchPgBouncer(context.TODO(), c.DBClient.KubedbV1alpha2(), db, func(in *api.PgBouncer) *api.PgBouncer {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, kubedb.GroupName)
					return in
				}, metav1.PatchOptions{})
				return err
			}
		} else {
			if !core_util.HasFinalizer(db.ObjectMeta, kubedb.GroupName) {
				db, _, err = util.PatchPgBouncer(context.TODO(), c.DBClient.KubedbV1alpha2(), db, func(in *api.PgBouncer) *api.PgBouncer {
					in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, kubedb.GroupName)
					return in
				}, metav1.PatchOptions{})
				if err != nil {
					return err
				}
			}
			// Get PgBouncer Phase from condition
			// if new phase is not equal to old phase
			// update PgBouncer phase
			phase := phase.PhaseFromCondition(db.Status.Conditions)
			if db.Status.Phase != phase {
				_, err := util.UpdatePgBouncerStatus(
					context.TODO(),
					c.DBClient.KubedbV1alpha2(),
					db.ObjectMeta,
					func(in *api.PgBouncerStatus) (types.UID, *api.PgBouncerStatus) {
						in.Phase = phase
						in.ObservedGeneration = db.Generation
						return db.UID, in
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					c.pushFailureEvent(db, err.Error())
					return err
				}

				// drop the object from queue,
				// the object will be enqueued again from this update event.
				return nil
			}

			// if conditions are empty, set initial condition "ProvisioningStarted" to "true"
			if !kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseProvisioningStarted) {
				_, err := util.UpdatePgBouncerStatus(
					context.TODO(),
					c.DBClient.KubedbV1alpha2(),
					db.ObjectMeta,
					func(in *api.PgBouncerStatus) (types.UID, *api.PgBouncerStatus) {
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:    api.DatabaseProvisioningStarted,
								Status:  core.ConditionTrue,
								Reason:  api.DatabaseProvisioningStartedSuccessfully,
								Message: fmt.Sprintf("The KubeDB operator has started the provisioning of PgBouncer: %s/%s", db.Namespace, db.Name),
							})
						return db.UID, in
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					return err
				}
				// drop the object from queue,
				// the object will be enqueued again from this update event.
				return nil
			}

			if kmapi.IsConditionTrue(db.Status.Conditions, api.DatabasePaused) {
				return nil
			}

			// process db object
			if err := c.syncPgBouncer(db); err != nil {
				klog.Errorln(err)
				c.pushFailureEvent(db, err.Error())
				return err
			}
		}
	}
	return nil
}

func (c *Controller) syncPgBouncer(db *api.PgBouncer) error {
	if err := c.manageValidation(db); err != nil {
		klog.Infoln(err)
		return nil // user err, dont' retry.
	}

	// ensure Governing Service
	if err := c.ensureGoverningService(db); err != nil {
		return fmt.Errorf(`failed to create governing Service for : "%v/%v". Reason: %v`, db.Namespace, db.Name, err)
	}
	// create or patch Service
	vt1, err := c.ensureService(db)
	if err != nil {
		klog.Infoln(err)
		return err
	}

	reconciler := NewReconciler(c.Config, c.Controller)
	db, vt2, err := reconciler.ReconcileNodes(db)
	if err != nil {
		return err
	}

	if db == nil {
		return nil
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.Recorder.Event(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created PgBouncer",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.Recorder.Event(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched PgBouncer",
		)
	}

	// get pgBouncer version
	pgBouncerVersion, err := c.DBClient.CatalogV1alpha1().PgBouncerVersions().Get(context.TODO(), db.Spec.Version, metav1.GetOptions{})
	if err != nil {
		klog.Infoln(err)
		return err
	}

	// ensure appbinding
	_, err = c.ensureAppBinding(db, pgBouncerVersion)
	if err != nil {
		klog.Errorln(err)
		return err
	}

	// create or patch Stat service
	if err := c.syncStatService(db); err != nil {
		c.Recorder.Eventf(
			db,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		klog.Infoln(err)
		return nil
	}

	if err := c.manageMonitor(db); err != nil {
		c.Recorder.Eventf(
			db,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		klog.Errorln(err)
		return nil
	}

	// Check: ReplicaReady --> AcceptingConnection --> Ready --> Provisioned
	if kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseReplicaReady) &&
		kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseAcceptingConnection) &&
		kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseReady) &&
		!kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseProvisioned) {
		_, err := util.UpdatePgBouncerStatus(
			context.TODO(),
			c.DBClient.KubedbV1alpha2(),
			db.ObjectMeta,
			func(in *api.PgBouncerStatus) (types.UID, *api.PgBouncerStatus) {
				in.Conditions = kmapi.SetCondition(in.Conditions,
					kmapi.Condition{
						Type:               api.DatabaseProvisioned,
						Status:             core.ConditionTrue,
						Reason:             api.DatabaseSuccessfullyProvisioned,
						ObservedGeneration: db.Generation,
						Message:            fmt.Sprintf("The PgBouncer: %s/%s is successfully provisioned.", db.Namespace, db.Name),
					})
				return db.UID, in
			},
			metav1.UpdateOptions{},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) manageValidation(db *api.PgBouncer) error {
	if err := validator.ValidatePgBouncer(c.Client, c.DBClient, db, true); err != nil {
		c.Recorder.Event(
			db,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		klog.Errorln(err)
		return err // user error so just record error and don't retry.
	}

	// Check if userList is absent.
	if db.Spec.UserListSecretRef != nil && db.Spec.UserListSecretRef.Name != "" {
		if db.Spec.ConnectionPool != nil && db.Spec.ConnectionPool.AuthType != api.PgBouncerClientAuthModeAny {
			if _, err := c.Client.CoreV1().Secrets(db.GetNamespace()).Get(context.TODO(), db.Spec.UserListSecretRef.Name, metav1.GetOptions{}); err != nil {
				c.Recorder.Eventf(
					db,
					core.EventTypeWarning,
					"UserListMissing",
					"user-list secret %s not found", db.Spec.UserListSecretRef.Name)
			}
		}
	}

	return nil
}

func (c *Controller) ensureService(db *api.PgBouncer) (kutil.VerbType, error) {
	vt, err := c.ensurePrimaryService(db)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	if vt != kutil.VerbUnchanged {
		c.Recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s Service",
			vt,
		)
	}
	if vt != kutil.VerbUnchanged {
		klog.Infoln("Service ", vt)
	}
	return vt, nil
}

func (c *Controller) syncStatService(db *api.PgBouncer) error {
	statServiceVerb, err := c.ensureStatsService(db)
	if err != nil {
		return err
	}
	if statServiceVerb == kutil.VerbCreated {
		c.Recorder.Event(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created Stat Service",
		)
	} else if statServiceVerb == kutil.VerbPatched {
		c.Recorder.Event(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched Stat Service",
		)
	}
	if statServiceVerb != kutil.VerbUnchanged {
		klog.Infoln("Stat Service ", statServiceVerb)
	}
	return nil
}

func (c *Controller) PgBouncerExists(db *api.PgBouncer) bool {
	_, err := c.DBClient.KubedbV1alpha2().PgBouncers(db.Namespace).Get(context.TODO(), db.Name, metav1.GetOptions{})
	return err == nil
}
