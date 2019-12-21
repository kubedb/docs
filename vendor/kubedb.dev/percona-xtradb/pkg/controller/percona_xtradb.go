/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controller

import (
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/percona-xtradb/pkg/admission"

	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kutil "kmodules.xyz/client-go"
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
		c.cronController.StopBackupScheduling(px.ObjectMeta)
		return nil
	}

	// Delete Matching DormantDatabase if exists any
	if err := c.deleteMatchingDormantDatabase(px); err != nil {
		return fmt.Errorf(`failed to delete dormant Database : "%v/%v". Reason: %v`, px.Namespace, px.Name, err)
	}

	if px.Status.Phase == "" {
		perconaxtradb, err := util.UpdatePerconaXtraDBStatus(c.ExtClient.KubedbV1alpha1(), px, func(in *api.PerconaXtraDBStatus) *api.PerconaXtraDBStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		})
		if err != nil {
			return err
		}
		px.Status = perconaxtradb.Status
	}

	// Set status as "Initializing" until specified restoresession object be succeeded, if provided
	if _, err := meta_util.GetString(px.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		px.Spec.Init != nil && px.Spec.Init.StashRestoreSession != nil {

		if px.Status.Phase == api.DatabasePhaseInitializing {
			return nil
		}

		perconaxtradb, err := util.UpdatePerconaXtraDBStatus(c.ExtClient.KubedbV1alpha1(), px, func(in *api.PerconaXtraDBStatus) *api.PerconaXtraDBStatus {
			in.Phase = api.DatabasePhaseInitializing
			return in
		})
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

	per, err := util.UpdatePerconaXtraDBStatus(c.ExtClient.KubedbV1alpha1(), px, func(in *api.PerconaXtraDBStatus) *api.PerconaXtraDBStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = px.Generation
		return in
	})
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

	_, err = c.ensureAppBinding(px)
	if err != nil {
		log.Errorln(err)
		return err
	}

	return nil
}

func (c *Controller) terminate(px *api.PerconaXtraDB) error {
	// If TerminationPolicy is "pause", keep everything (ie, PVCs,Secrets,Snapshots) intact.
	// In operator, create dormantdatabase
	if px.Spec.TerminationPolicy == api.TerminationPolicyPause {
		if err := c.removeOwnerReferenceFromOffshoots(px); err != nil {
			return err
		}

		if _, err := c.createDormantDatabase(px); err != nil {
			if kerr.IsAlreadyExists(err) {
				// if already exists, check if it is database of another Kind and return error in that case.
				// If the Kind is same, we can safely assume that the DormantDB was not deleted in before,
				// Probably because, User is more faster (create-delete-create-again-delete...) than operator!
				// So reuse that DormantDB!
				ddb, err := c.ExtClient.KubedbV1alpha1().DormantDatabases(px.Namespace).Get(px.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if val, _ := meta_util.GetStringValue(ddb.Labels, api.LabelDatabaseKind); val != api.ResourceKindPerconaXtraDB {
					return fmt.Errorf(`DormantDatabase "%v" of kind %v already exists`, px.Name, val)
				}
			} else {
				return fmt.Errorf(`failed to create DormantDatabase: "%v/%v". Reason: %v`, px.Namespace, px.Name, err)
			}
		}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshoots(px); err != nil {
			return err
		}
	}

	c.cronController.StopBackupScheduling(px.ObjectMeta)

	if px.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(px); err != nil {
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
		if err := dynamic_util.EnsureOwnerReferenceForSelector(
			c.DynamicClient,
			api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
			px.Namespace,
			selector,
			owner); err != nil {
			return err
		}
		if err := c.wipeOutDatabase(px.ObjectMeta, px.Spec.GetSecrets(), owner); err != nil {
			return errors.Wrap(err, "error in wiping out database.")
		}
	} else {
		// Make sure snapshot and secret's ownerreference is removed.
		if err := dynamic_util.RemoveOwnerReferenceForSelector(
			c.DynamicClient,
			api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
			px.Namespace,
			selector,
			px); err != nil {
			return err
		}
		if err := dynamic_util.RemoveOwnerReferenceForItems(
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
		c.DynamicClient,
		api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
		px.Namespace,
		labelSelector,
		px); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		px.Namespace,
		labelSelector,
		px); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForItems(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		px.Namespace,
		px.Spec.GetSecrets(),
		px); err != nil {
		return err
	}
	return nil
}
