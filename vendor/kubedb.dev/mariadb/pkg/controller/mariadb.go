/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/mariadb/pkg/admission"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	kmapi "kmodules.xyz/client-go/api/v1"
	dynamic_util "kmodules.xyz/client-go/dynamic"
)

func (c *Controller) create(db *api.MariaDB) error {

	// validate MariaDB object
	if err := validator.ValidateMariaDB(c.Client, c.DBClient, db, true); err != nil {
		c.Recorder.Event(
			db,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		klog.Errorln(err)
		return nil
	}

	// ensure service account, role, rolebinding, and PSP for database statefulsets
	if err := c.ensureRBACStuff(db); err != nil {
		return fmt.Errorf(`failed to create RBAC's for : "%v/%v". Reason: %v`, db.Namespace, db.Name, err)
	}

	// ensure Governing Service
	if err := c.ensureGoverningService(db); err != nil {
		return fmt.Errorf(`failed to create Governing Service for : "%v/%v". Reason: %v`, db.Namespace, db.Name, err)
	}

	vt1, err := c.ensureService(db)
	if err != nil {
		return fmt.Errorf(`failed to create Service for : "%v/%v". Reason: %v`, db.Namespace, db.Name, err)
	}

	reconciler := NewReconciler(c.Config, c.Controller)
	db, vt2, err := reconciler.ReconcileNodes(db)
	if err != nil {
		return err
	}

	// If both err==nil & mariadb == nil,
	// the object was dropped from the work-queue, to process later.
	// return nil.
	if db == nil {
		return nil
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.Recorder.Event(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created MariaDB",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.Recorder.Event(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched MariaDB",
		)
	}

	_, err = c.ensureAppBinding(db)
	if err != nil {
		klog.Errorln(err)
		return err
	}

	if db.Spec.Init != nil {
		// For MariaDB Cluster (md.spec.replicas > 1), Stash restores the data into some PVCs.
		// Then, KubeDB should create the StatefulSet using those PVCs. So, for clustering mode, we are going to
		// wait for restore process to complete before creating the StatefulSet.
		//======================== Wait for the initial restore =====================================
		if db.Spec.Init.WaitForInitialRestore && db.IsCluster() {
			// Only wait for the first restore.
			// For initial restore, "Provisioned" condition won't exist and "DataRestored" condition either won't exist or will be "False".
			if !kmapi.HasCondition(db.Status.Conditions, api.DatabaseProvisioned) &&
				!kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseDataRestored) {
				// write log indicating that the database is waiting for the data to be restored by external initializer
				klog.Infof("Database %s %s/%s is waiting for data to be restored by external initializer",
					db.Kind,
					db.Namespace,
					db.Name,
				)
				// Rest of the processing will execute after the the restore process completed. So, just return for now.
				return nil
			}
		}

		//======================== Wait for initialize script =====================================
		if db.Spec.Init.Script != nil && !db.IsCluster() {
			if !kmapi.HasCondition(db.Status.Conditions, api.DatabaseDataRestored) &&
				kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseReplicaReady) &&
				kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseAcceptingConnection) &&
				kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseReady) {
				_, err := util.UpdateMariaDBStatus(
					context.TODO(),
					c.DBClient.KubedbV1alpha2(),
					db.ObjectMeta,
					func(in *api.MariaDBStatus) (types.UID, *api.MariaDBStatus) {
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:               api.DatabaseDataRestored,
								Status:             core.ConditionTrue,
								Reason:             api.DatabaseSuccessfullyRestored,
								ObservedGeneration: db.Generation,
								Message:            fmt.Sprintf("Data successfully restored into The MySQL databse: %s/%s", db.Namespace, db.Name),
							})
						return db.UID, in
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					return err
				}
			}
		}
	}

	if _, err := c.ensureStatsService(db); err != nil {
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
	// If spec.Init.WaitForInitialRestore is true, but data wasn't restored successfully,
	// process won't reach here (returned nil at the beginning). As it is here, that means data was restored successfully.
	// No need to check for IsConditionTrue(DataRestored).
	if kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseReplicaReady) &&
		kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseAcceptingConnection) &&
		kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseReady) &&
		!kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseProvisioned) {
		_, err := util.UpdateMariaDBStatus(
			context.TODO(),
			c.DBClient.KubedbV1alpha2(),
			db.ObjectMeta,
			func(in *api.MariaDBStatus) (types.UID, *api.MariaDBStatus) {
				in.Conditions = kmapi.SetCondition(in.Conditions,
					kmapi.Condition{
						Type:               api.DatabaseProvisioned,
						Status:             core.ConditionTrue,
						Reason:             api.DatabaseSuccessfullyProvisioned,
						ObservedGeneration: db.Generation,
						Message:            fmt.Sprintf("The MariaDB: %s/%s is successfully provisioned.", db.Namespace, db.Name),
					})
				return db.UID, in
			},
			metav1.UpdateOptions{},
		)
		if err != nil {
			return err
		}
	}

	// If the database is successfully provisioned,
	// Set spec.Init.Initialized to true, if init!=nil.
	// This will prevent the operator from re-initializing the database.
	if db.Spec.Init != nil &&
		!db.Spec.Init.Initialized &&
		kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseProvisioned) {
		_, _, err := util.CreateOrPatchMariaDB(context.TODO(), c.DBClient.KubedbV1alpha2(), db.ObjectMeta, func(in *api.MariaDB) *api.MariaDB {
			in.Spec.Init.Initialized = true
			return in
		}, metav1.PatchOptions{})

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) halt(db *api.MariaDB) error {
	if db.Spec.TerminationPolicy != api.TerminationPolicyHalt {
		return errors.New("can't halt db. 'spec.terminationPolicy' is not 'Halt'")
	}
	klog.Infof("Halting MariaDB %v/%v", db.Namespace, db.Name)
	if err := c.haltDatabase(db); err != nil {
		return err
	}
	if err := c.waitUntilPaused(db); err != nil {
		return err
	}
	klog.Infof("update status of MariaDB %v/%v to Halted.", db.Namespace, db.Name)
	if _, err := util.UpdateMariaDBStatus(
		context.TODO(),
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.MariaDBStatus) (types.UID, *api.MariaDBStatus) {
			in.Conditions = kmapi.SetCondition(in.Conditions, kmapi.Condition{
				Type:               api.DatabaseHalted,
				Status:             core.ConditionTrue,
				Reason:             api.DatabaseHaltedSuccessfully,
				ObservedGeneration: db.Generation,
				Message:            fmt.Sprintf("MySQL %s/%s successfully halted.", db.Namespace, db.Name),
			})
			// make "AcceptingConnection" and "Ready" conditions false.
			// Because these are handled from health checker at a certain interval,
			// if consecutive halt and un-halt occurs in the meantime,
			// phase might still be on the "Ready" state.
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseAcceptingConnection,
					Status:             core.ConditionFalse,
					Reason:             api.DatabaseHaltedSuccessfully,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The MariaDB: %s/%s is not accepting client requests.", db.Namespace, db.Name),
				})
			in.Conditions = kmapi.SetCondition(in.Conditions,
				kmapi.Condition{
					Type:               api.DatabaseReady,
					Status:             core.ConditionFalse,
					Reason:             api.DatabaseHaltedSuccessfully,
					ObservedGeneration: db.Generation,
					Message:            fmt.Sprintf("The MariaDB: %s/%s is not ready.", db.Namespace, db.Name),
				})
			return db.UID, in
		},
		metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *Controller) terminate(db *api.MariaDB) error {
	// If TerminationPolicy is "halt", keep PVCs and Secrets intact.
	// TerminationPolicyPause is deprecated and will be removed in future.
	if db.Spec.TerminationPolicy == api.TerminationPolicyHalt {
		if err := c.removeOwnerReferenceFromOffshoots(db); err != nil {
			return err
		}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshoots(db); err != nil {
			return err
		}
	}

	if db.Spec.Monitor != nil {
		if err := c.deleteMonitor(db); err != nil {
			klog.Errorln(err)
			return nil
		}
	}
	return nil
}

func (c *Controller) setOwnerReferenceToOffshoots(db *api.MariaDB) error {
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMariaDB))
	selector := labels.SelectorFromSet(db.OffshootSelectors())

	// If TerminationPolicy is "wipeOut", delete snapshots and secrets,
	// else, keep it intact.
	if db.Spec.TerminationPolicy == api.TerminationPolicyWipeOut {
		if err := c.wipeOutDatabase(db.ObjectMeta, db.Spec.GetPersistentSecrets(), owner); err != nil {
			return errors.Wrap(err, "error in wiping out database.")
		}
	} else {
		// Make sure secret's ownerreference is removed.
		if err := dynamic_util.RemoveOwnerReferenceForItems(
			context.TODO(),
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			db.Namespace,
			db.Spec.GetPersistentSecrets(),
			db); err != nil {
			return err
		}
	}
	// delete PVC for both "wipeOut" and "delete" TerminationPolicy.
	return dynamic_util.EnsureOwnerReferenceForSelector(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		db.Namespace,
		selector,
		owner)
}

func (c *Controller) removeOwnerReferenceFromOffshoots(db *api.MariaDB) error {
	// First, Get LabelSelector for Other Components
	labelSelector := labels.SelectorFromSet(db.OffshootSelectors())

	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		db.Namespace,
		labelSelector,
		db); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForItems(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		db.Namespace,
		db.Spec.GetPersistentSecrets(),
		db); err != nil {
		return err
	}
	return nil
}
