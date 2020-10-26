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
	validator "kubedb.dev/percona-xtradb/pkg/admission"

	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kutil "kmodules.xyz/client-go"
	kmapi "kmodules.xyz/client-go/api/v1"
	dynamic_util "kmodules.xyz/client-go/dynamic"
)

func (c *Controller) create(db *api.PerconaXtraDB) error {
	if err := validator.ValidatePerconaXtraDB(c.Client, c.DBClient, db, true); err != nil {
		c.Recorder.Event(
			db,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		// stop Scheduler in case there is any.
		return nil
	}

	if db.Status.Phase == "" {
		perconaxtradb, err := util.UpdatePerconaXtraDBStatus(context.TODO(), c.DBClient.KubedbV1alpha2(), db.ObjectMeta, func(in *api.PerconaXtraDBStatus) *api.PerconaXtraDBStatus {
			in.Phase = api.DatabasePhaseProvisioning
			return in
		}, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		db.Status = perconaxtradb.Status
	}

	// For Percona XtraDB Cluster (px.spec.replicas > 1), Stash restores the data into some PVCs.
	// Then, KubeDB should create the StatefulSet using those PVCs. So, for clustering mode, we are going to
	// wait for restore process to complete before creating the StatefulSet.
	//======================== Wait for the initial restore =====================================
	if db.Spec.Init != nil && db.Spec.Init.WaitForInitialRestore && db.IsCluster() {
		// Only wait for the first restore.
		// For initial restore, "Provisioned" condition won't exist and "DataRestored" condition either won't exist or will be "False".
		if !kmapi.HasCondition(db.Status.Conditions, api.DatabaseProvisioned) &&
			!kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseDataRestored) {
			// write log indicating that the database is waiting for the data to be restored by external initializer
			log.Infof("Database %s %s/%s is waiting for data to be restored by external initializer",
				db.Kind,
				db.Namespace,
				db.Name,
			)
			// Rest of the processing will execute after the the restore process completed. So, just return for now.
			return nil
		}
	}

	// ensure Governing Service
	if err := c.ensureGoverningService(db); err != nil {
		return fmt.Errorf(`failed to create governing Service for : "%v/%v". Reason: %v`, db.Namespace, db.Name, err)
	}

	// Ensure ClusterRoles for statefulsets
	if err := c.ensureRBACStuff(db); err != nil {
		return err
	}

	// ensure database Service
	vt1, err := c.ensureService(db)
	if err != nil {
		return err
	}

	if err := c.ensureAuthSecret(db); err != nil {
		return err
	}

	// ensure database StatefulSet
	vt2, err := c.ensurePerconaXtraDB(db)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.Recorder.Event(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created PerconaXtraDB",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.Recorder.Event(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched PerconaXtraDB",
		)
	}

	_, err = c.ensureAppBinding(db)
	if err != nil {
		log.Errorln(err)
		return err
	}

	// For Standalone Percona XtraDB (px.spec.replicas = 1),, Stash directly restore into the database.
	// So, for standalone mode, we are going to wait for restore process to complete after creating the StatefulSet.
	//======================== Wait for the initial restore =====================================
	if db.Spec.Init != nil && db.Spec.Init.WaitForInitialRestore && !db.IsCluster() {
		// Only wait for the first restore.
		// For initial restore, "Provisioned" condition won't exist and "DataRestored" condition either won't exist or will be "False".
		if !kmapi.HasCondition(db.Status.Conditions, api.DatabaseProvisioned) &&
			!kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseDataRestored) {
			// write log indicating that the database is waiting for the data to be restored by external initializer
			log.Infof("Database %s %s/%s is waiting for data to be restored by external initializer",
				db.Kind,
				db.Namespace,
				db.Name,
			)
			// Rest of the processing will execute after the the restore process completed. So, just return for now.
			return nil
		}
	}

	per, err := util.UpdatePerconaXtraDBStatus(context.TODO(), c.DBClient.KubedbV1alpha2(), db.ObjectMeta, func(in *api.PerconaXtraDBStatus) *api.PerconaXtraDBStatus {
		in.Phase = api.DatabasePhaseReady
		in.ObservedGeneration = db.Generation
		return in
	}, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	db.Status = per.Status

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(db); err != nil {
		c.Recorder.Eventf(
			db,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
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
		log.Errorln(err)
		return nil
	}

	return nil
}

func (c *Controller) halt(db *api.PerconaXtraDB) error {
	if db.Spec.Halted && db.Spec.TerminationPolicy != api.TerminationPolicyHalt {
		return errors.New("can't halt db. 'spec.terminationPolicy' is not 'Halt'")
	}
	log.Infof("Halting PerconaXtraDB %v/%v", db.Namespace, db.Name)
	if err := c.haltDatabase(db); err != nil {
		return err
	}
	if err := c.waitUntilPaused(db); err != nil {
		return err
	}
	log.Infof("update status of PerconaXtraDB %v/%v to Halted.", db.Namespace, db.Name)
	if _, err := util.UpdatePerconaXtraDBStatus(context.TODO(), c.DBClient.KubedbV1alpha2(), db.ObjectMeta, func(in *api.PerconaXtraDBStatus) *api.PerconaXtraDBStatus {
		in.Phase = api.DatabasePhaseHalted
		in.ObservedGeneration = db.Generation
		return in
	}, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *Controller) terminate(db *api.PerconaXtraDB) error {
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
			log.Errorln(err)
			return nil
		}
	}
	return nil
}

func (c *Controller) setOwnerReferenceToOffshoots(db *api.PerconaXtraDB) error {
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindPerconaXtraDB))
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

func (c *Controller) removeOwnerReferenceFromOffshoots(db *api.PerconaXtraDB) error {
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
