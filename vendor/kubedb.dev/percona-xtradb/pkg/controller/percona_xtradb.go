/*
Copyright AppsCode Inc. and Contributors

Licensed under the PolyForm Noncommercial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/PolyForm-Noncommercial-1.0.0.md

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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/percona-xtradb/pkg/admission"

	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	dynamic_util "kmodules.xyz/client-go/dynamic"
	meta_util "kmodules.xyz/client-go/meta"
)

func (c *Controller) create(px *api.PerconaXtraDB) error {
	if err := validator.ValidatePerconaXtraDB(c.Client, c.ExtClient, px, true); err != nil {
		c.recorder.Event(
			px,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		// stop Scheduler in case there is any.
		return nil
	}

	if px.Status.Phase == "" {
		perconaxtradb, err := util.UpdatePerconaXtraDBStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), px.ObjectMeta, func(in *api.PerconaXtraDBStatus) *api.PerconaXtraDBStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		}, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		px.Status = perconaxtradb.Status
	}

	// For Percona XtraDB Cluster (px.spec.replicas > 1),
	// Set status as "Initializing" until specified restoresession object be succeeded, if provided
	if _, err := meta_util.GetString(px.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		px.IsCluster() && px.Spec.Init != nil && px.Spec.Init.StashRestoreSession != nil {

		if px.Status.Phase == api.DatabasePhaseInitializing {
			return nil
		}

		perconaxtradb, err := util.UpdatePerconaXtraDBStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), px.ObjectMeta, func(in *api.PerconaXtraDBStatus) *api.PerconaXtraDBStatus {
			in.Phase = api.DatabasePhaseInitializing
			return in
		}, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		px.Status = perconaxtradb.Status

		log.Debugf("PerconaXtraDB %v/%v is waiting for restoreSession to be succeeded", px.Namespace, px.Name)
		return nil
	}

	// create Governing Service
	governingService, err := c.createPerconaXtraDBGoverningService(px)
	if err != nil {
		return fmt.Errorf(`failed to create Service: "%v/%v". Reason: %v`, px.Namespace, governingService, err)
	}
	c.GoverningService = governingService

	// Ensure ClusterRoles for statefulsets
	if err := c.ensureRBACStuff(px); err != nil {
		return err
	}

	// ensure database Service
	vt1, err := c.ensureService(px)
	if err != nil {
		return err
	}

	if err := c.ensureDatabaseSecret(px); err != nil {
		return err
	}

	// ensure database StatefulSet
	vt2, err := c.ensurePerconaXtraDB(px)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.recorder.Event(
			px,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created PerconaXtraDB",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			px,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched PerconaXtraDB",
		)
	}

	_, err = c.ensureAppBinding(px)
	if err != nil {
		log.Errorln(err)
		return err
	}

	// For Standalone Percona XtraDB (px.spec.replicas = 1),
	// Set status as "Initializing" until specified restoresession object be succeeded, if provided
	if _, err := meta_util.GetString(px.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		!px.IsCluster() && px.Spec.Init != nil && px.Spec.Init.StashRestoreSession != nil {

		if px.Status.Phase == api.DatabasePhaseInitializing {
			return nil
		}

		// add phase that database is being initialized
		perconaxtradb, err := util.UpdatePerconaXtraDBStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), px.ObjectMeta, func(in *api.PerconaXtraDBStatus) *api.PerconaXtraDBStatus {
			in.Phase = api.DatabasePhaseInitializing
			return in
		}, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		px.Status = perconaxtradb.Status

		log.Debugf("PerconaXtraDB %v/%v is waiting for restoreSession to be succeeded", px.Namespace, px.Name)
		return nil
	}

	per, err := util.UpdatePerconaXtraDBStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), px.ObjectMeta, func(in *api.PerconaXtraDBStatus) *api.PerconaXtraDBStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = px.Generation
		return in
	}, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	px.Status = per.Status

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(px); err != nil {
		c.recorder.Eventf(
			px,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
		return nil
	}

	if err := c.manageMonitor(px); err != nil {
		c.recorder.Eventf(
			px,
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
	if _, err := util.UpdatePerconaXtraDBStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), db.ObjectMeta, func(in *api.PerconaXtraDBStatus) *api.PerconaXtraDBStatus {
		in.Phase = api.DatabasePhaseHalted
		in.ObservedGeneration = db.Generation
		return in
	}, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *Controller) terminate(px *api.PerconaXtraDB) error {
	// If TerminationPolicy is "halt", keep PVCs and Secrets intact.
	// TerminationPolicyPause is deprecated and will be removed in future.
	if px.Spec.TerminationPolicy == api.TerminationPolicyHalt || px.Spec.TerminationPolicy == api.TerminationPolicyPause {
		if err := c.removeOwnerReferenceFromOffshoots(px); err != nil {
			return err
		}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshoots(px); err != nil {
			return err
		}
	}

	if px.Spec.Monitor != nil {
		if err := c.deleteMonitor(px); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}

func (c *Controller) setOwnerReferenceToOffshoots(px *api.PerconaXtraDB) error {
	owner := metav1.NewControllerRef(px, api.SchemeGroupVersion.WithKind(api.ResourceKindPerconaXtraDB))
	selector := labels.SelectorFromSet(px.OffshootSelectors())

	// If TerminationPolicy is "wipeOut", delete snapshots and secrets,
	// else, keep it intact.
	if px.Spec.TerminationPolicy == api.TerminationPolicyWipeOut {
		if err := c.wipeOutDatabase(px.ObjectMeta, px.Spec.GetSecrets(), owner); err != nil {
			return errors.Wrap(err, "error in wiping out database.")
		}
	} else {
		// Make sure secret's ownerreference is removed.
		if err := dynamic_util.RemoveOwnerReferenceForItems(
			context.TODO(),
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			px.Namespace,
			px.Spec.GetSecrets(),
			px); err != nil {
			return err
		}
	}
	// delete PVC for both "wipeOut" and "delete" TerminationPolicy.
	return dynamic_util.EnsureOwnerReferenceForSelector(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		px.Namespace,
		selector,
		owner)
}

func (c *Controller) removeOwnerReferenceFromOffshoots(px *api.PerconaXtraDB) error {
	// First, Get LabelSelector for Other Components
	labelSelector := labels.SelectorFromSet(px.OffshootSelectors())

	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		px.Namespace,
		labelSelector,
		px); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForItems(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		px.Namespace,
		px.Spec.GetSecrets(),
		px); err != nil {
		return err
	}
	return nil
}

func (c *Controller) GetDatabase(meta metav1.ObjectMeta) (runtime.Object, error) {
	px, err := c.pxLister.PerconaXtraDBs(meta.Namespace).Get(meta.Name)
	if err != nil {
		return nil, err
	}

	return px, nil
}

func (c *Controller) SetDatabaseStatus(meta metav1.ObjectMeta, phase api.DatabasePhase, reason string) error {
	px, err := c.pxLister.PerconaXtraDBs(meta.Namespace).Get(meta.Name)
	if err != nil {
		return err
	}
	_, err = util.UpdatePerconaXtraDBStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), px.ObjectMeta, func(in *api.PerconaXtraDBStatus) *api.PerconaXtraDBStatus {
		in.Phase = phase
		in.Reason = reason
		return in
	}, metav1.UpdateOptions{})
	return err
}

func (c *Controller) UpsertDatabaseAnnotation(meta metav1.ObjectMeta, annotation map[string]string) error {
	px, err := c.pxLister.PerconaXtraDBs(meta.Namespace).Get(meta.Name)
	if err != nil {
		return err
	}

	_, _, err = util.PatchPerconaXtraDB(context.TODO(), c.ExtClient.KubedbV1alpha1(), px, func(in *api.PerconaXtraDB) *api.PerconaXtraDB {
		in.Annotations = core_util.UpsertMap(in.Annotations, annotation)
		return in
	}, metav1.PatchOptions{})
	return err
}
