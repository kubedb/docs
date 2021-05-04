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

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/postgres/pkg/admission"

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

func (c *Controller) create(db *api.Postgres) error {
	if err := validator.ValidatePostgres(c.Client, c.DBClient, db, true); err != nil {
		c.Recorder.Event(
			db,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		klog.Errorln(err)
		return nil // user error so just record error and don't retry.
	}

	//if db.Status.Phase == "" {
	//	pg, err := util.UpdatePostgresStatus(context.TODO(), c.DBClient.KubedbV1alpha2(), db.ObjectMeta, func(in *api.PostgresStatus) (types.UID, *api.PostgresStatus) {
	//		in.Phase = api.DatabasePhaseProvisioning
	//		return db.UID, in
	//	}, metav1.UpdateOptions{})
	//	if err != nil {
	//		return err
	//	}
	//	db.Status = pg.Status
	//}

	// ensure Governing Service
	if err := c.ensureGoverningService(db); err != nil {
		return fmt.Errorf(`failed to create governing Service for : "%v/%v". Reason: %v`, db.Namespace, db.Name, err)
	}

	// ensure database Service
	vt1, err := c.ensureService(db)
	if err != nil {
		return err
	}
	// wait for  Certificates secrets
	if db.Spec.TLS != nil {
		ok, err := dynamic_util.ResourcesExists(
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			db.Namespace,
			db.GetCertSecretName(api.PostgresServerCert),
			db.GetCertSecretName(api.PostgresClientCert),
			db.GetCertSecretName(api.PostgresMetricsExporterCert),
		)
		if err != nil {
			return err
		}
		if !ok {
			klog.Infof("wait for all certificate secrets for Postgres %s/%s", db.Namespace, db.Name)
			return nil
		}
	}
	// ensure database StatefulSet
	postgresVersion, err := c.DBClient.CatalogV1alpha1().PostgresVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return err
	}
	vt2, err := c.ensurePostgresNode(db, postgresVersion)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.Recorder.Event(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created Postgres",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.Recorder.Event(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched Postgres",
		)
	}

	// ensure appbinding before ensuring Restic scheduler and restore
	_, err = c.ensureAppBinding(db, postgresVersion)
	if err != nil {
		klog.Errorln(err)
		return err
	}

	//======================== Wait for the initial restore =====================================
	if db.Spec.Init != nil && db.Spec.Init.WaitForInitialRestore {
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

	pg, err := util.UpdatePostgresStatus(context.TODO(), c.DBClient.KubedbV1alpha2(), db.ObjectMeta, func(in *api.PostgresStatus) (types.UID, *api.PostgresStatus) {
		in.Phase = api.DatabasePhaseReady
		in.ObservedGeneration = db.Generation
		return db.UID, in
	}, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	db.Status = pg.Status

	// ensure StatsService for desired monitoring
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
		_, err := util.UpdatePostgresStatus(
			context.TODO(),
			c.DBClient.KubedbV1alpha2(),
			db.ObjectMeta,
			func(in *api.PostgresStatus) (types.UID, *api.PostgresStatus) {
				in.Conditions = kmapi.SetCondition(in.Conditions,
					kmapi.Condition{
						Type:               api.DatabaseProvisioned,
						Status:             core.ConditionTrue,
						Reason:             api.DatabaseSuccessfullyProvisioned,
						ObservedGeneration: db.Generation,
						Message:            fmt.Sprintf("The PostgreSQL: %s/%s is successfully provisioned.", db.Namespace, db.Name),
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
		_, _, err := util.CreateOrPatchPostgres(context.TODO(), c.DBClient.KubedbV1alpha2(), db.ObjectMeta, func(in *api.Postgres) *api.Postgres {
			in.Spec.Init.Initialized = true
			return in
		}, metav1.PatchOptions{})

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) ensurePostgresNode(db *api.Postgres, postgresVersion *catalog.PostgresVersion) (kutil.VerbType, error) {
	var err error

	if err = c.ensureAuthSecret(db); err != nil {
		return kutil.VerbUnchanged, err
	}

	// Ensure Service account, role, rolebinding, and PSP for database statefulsets
	if err := c.ensureDatabaseRBAC(db); err != nil {
		return kutil.VerbUnchanged, err
	}
	if err = c.ensureValidUserForPostgreSQL(db); err != nil {
		return kutil.VerbUnchanged, err
	}
	vt, err := c.ensureCombinedNode(db, postgresVersion)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	return vt, nil
}

func (c *Controller) halt(db *api.Postgres) error {
	if db.Spec.Halted && db.Spec.TerminationPolicy != api.TerminationPolicyHalt {
		return errors.New("can't halt db. 'spec.terminationPolicy' is not 'Halt'")
	}
	klog.Infof("Halting Postgres %v/%v", db.Namespace, db.Name)
	if err := c.haltDatabase(db); err != nil {
		return err
	}
	if err := c.waitUntilPaused(db); err != nil {
		return err
	}
	klog.Infof("update status of Postgres %v/%v to Halted.", db.Namespace, db.Name)
	if _, err := util.UpdatePostgresStatus(context.TODO(), c.DBClient.KubedbV1alpha2(), db.ObjectMeta, func(in *api.PostgresStatus) (types.UID, *api.PostgresStatus) {
		in.Conditions = kmapi.SetCondition(in.Conditions, kmapi.Condition{
			Type:               api.DatabaseHalted,
			Status:             core.ConditionTrue,
			Reason:             api.DatabaseHaltedSuccessfully,
			ObservedGeneration: db.Generation,
			Message:            fmt.Sprintf("PostgreSQL %s/%s successfully halted.", db.Namespace, db.Name),
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
				Message:            fmt.Sprintf("The PostgreSQL: %s/%s is not accepting client requests.", db.Namespace, db.Name),
			})
		in.Conditions = kmapi.SetCondition(in.Conditions,
			kmapi.Condition{
				Type:               api.DatabaseReady,
				Status:             core.ConditionFalse,
				Reason:             api.DatabaseHaltedSuccessfully,
				ObservedGeneration: db.Generation,
				Message:            fmt.Sprintf("The PostgreSQL: %s/%s is not ready.", db.Namespace, db.Name),
			})

		return db.UID, in
	}, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *Controller) terminate(db *api.Postgres) error {
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindPostgres))

	// If TerminationPolicy is "halt", keep PVCs and Secrets intact.
	// TerminationPolicyPause is deprecated and will be removed in future.
	if db.Spec.TerminationPolicy == api.TerminationPolicyHalt {
		if err := c.removeOwnerReferenceFromOffshoots(db); err != nil {
			return err
		}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots,WAL-data).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets, wal-data intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshoots(db, owner); err != nil {
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

func (c *Controller) setOwnerReferenceToOffshoots(db *api.Postgres, owner *metav1.OwnerReference) error {
	selector := labels.SelectorFromSet(db.OffshootSelectors())

	// If TerminationPolicy is "wipeOut", delete snapshots and secrets,
	// else, keep it intact.
	if db.Spec.TerminationPolicy == api.TerminationPolicyWipeOut {
		if err := c.wipeOutDatabase(db.ObjectMeta, db.Spec.GetPersistentSecrets(), owner); err != nil {
			return errors.Wrap(err, "error in wiping out database.")
		}

	} else {
		secrets := db.Spec.GetPersistentSecrets()
		secrets = append(secrets, c.GetPostgresSecrets(db)...)
		// Make sure secret's ownerreference is removed.
		if err := dynamic_util.RemoveOwnerReferenceForItems(
			context.TODO(),
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			db.Namespace,
			secrets,
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

func (c *Controller) removeOwnerReferenceFromOffshoots(db *api.Postgres) error {

	secrets := db.Spec.GetPersistentSecrets()
	secrets = append(secrets, c.GetPostgresSecrets(db)...)

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
		secrets,
		db); err != nil {
		return err
	}
	return nil
}
