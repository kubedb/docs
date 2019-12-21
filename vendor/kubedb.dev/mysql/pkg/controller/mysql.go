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
	validator "kubedb.dev/mysql/pkg/admission"

	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kutil "kmodules.xyz/client-go"
	dynamic_util "kmodules.xyz/client-go/dynamic"
	meta_util "kmodules.xyz/client-go/meta"
	storage "kmodules.xyz/objectstore-api/osm"
)

func (c *Controller) create(mysql *api.MySQL) error {
	if err := validator.ValidateMySQL(c.Client, c.ExtClient, mysql, true); err != nil {
		c.recorder.Event(
			mysql,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		// stop Scheduler in case there is any.
		c.cronController.StopBackupScheduling(mysql.ObjectMeta)
		return nil
	}

	// Delete Matching DormantDatabase if exists any
	if err := c.deleteMatchingDormantDatabase(mysql); err != nil {
		return fmt.Errorf(`failed to delete dormant Database : "%v/%v". Reason: %v`, mysql.Namespace, mysql.Name, err)
	}

	if mysql.Status.Phase == "" {
		my, err := util.UpdateMySQLStatus(c.ExtClient.KubedbV1alpha1(), mysql, func(in *api.MySQLStatus) *api.MySQLStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		})
		if err != nil {
			return err
		}
		mysql.Status = my.Status
	}

	// create Governing Service
	governingService, err := c.createMySQLGoverningService(mysql)
	if err != nil {
		return fmt.Errorf(`failed to create Service: "%v/%v". Reason: %v`, mysql.Namespace, governingService, err)
	}
	c.GoverningService = governingService

	// Ensure Service account, role, rolebinding, and PSP for database statefulsets
	if err := c.ensureDatabaseRBAC(mysql); err != nil {
		return err
	}

	// ensure database Service
	vt1, err := c.ensureService(mysql)
	if err != nil {
		return err
	}

	if err := c.ensureDatabaseSecret(mysql); err != nil {
		return err
	}

	// ensure database StatefulSet
	vt2, err := c.ensureStatefulSet(mysql)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.recorder.Event(
			mysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created MySQL",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			mysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched MySQL",
		)
	}

	// ensure appbinding before ensuring Restic scheduler and restore
	_, err = c.ensureAppBinding(mysql)
	if err != nil {
		log.Errorln(err)
		return err
	}

	if _, err := meta_util.GetString(mysql.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		mysql.Spec.Init != nil &&
		(mysql.Spec.Init.SnapshotSource != nil || mysql.Spec.Init.StashRestoreSession != nil) {

		if mysql.Status.Phase == api.DatabasePhaseInitializing {
			return nil
		}

		// add phase that database is being initialized
		my, err := util.UpdateMySQLStatus(c.ExtClient.KubedbV1alpha1(), mysql, func(in *api.MySQLStatus) *api.MySQLStatus {
			in.Phase = api.DatabasePhaseInitializing
			return in
		})
		if err != nil {
			return err
		}
		mysql.Status = my.Status

		init := mysql.Spec.Init
		if init.SnapshotSource != nil {
			err = c.initializeFromSnapshot(mysql)
			if err != nil {
				return fmt.Errorf("failed to complete initialization. Reason: %v", err)
			}
			return err
		} else if init.StashRestoreSession != nil {
			log.Debugf("MySQL %v/%v is waiting for restoreSession to be succeeded", mysql.Namespace, mysql.Name)
			return nil
		}
	}

	my, err := util.UpdateMySQLStatus(c.ExtClient.KubedbV1alpha1(), mysql, func(in *api.MySQLStatus) *api.MySQLStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = mysql.Generation
		return in
	})
	if err != nil {
		return err
	}
	mysql.Status = my.Status

	// Ensure Schedule backup
	if err := c.ensureBackupScheduler(mysql); err != nil {
		c.recorder.Eventf(
			mysql,
			core.EventTypeWarning,
			eventer.EventReasonFailedToSchedule,
			err.Error(),
		)
		log.Errorln(err)
		// Don't return error. Continue processing rest.
	}

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(mysql); err != nil {
		c.recorder.Eventf(
			mysql,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
		return nil
	}

	if err := c.manageMonitor(mysql); err != nil {
		c.recorder.Eventf(
			mysql,
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

func (c *Controller) ensureBackupScheduler(mysql *api.MySQL) error {
	mysqlVersion, err := c.ExtClient.CatalogV1alpha1().MySQLVersions().Get(string(mysql.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get MySQLVersion %v for %v/%v. Reason: %v", mysql.Spec.Version, mysql.Namespace, mysql.Name, err)
	}
	// Setup Schedule backup
	if mysql.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(mysql, mysql.Spec.BackupSchedule, mysqlVersion)
		if err != nil {
			return fmt.Errorf("failed to schedule snapshot for %v/%v. Reason: %v", mysql.Namespace, mysql.Name, err)
		}
	} else {
		c.cronController.StopBackupScheduling(mysql.ObjectMeta)
	}
	return nil
}

func (c *Controller) initializeFromSnapshot(mysql *api.MySQL) error {
	snapshotSource := mysql.Spec.Init.SnapshotSource
	jobName := fmt.Sprintf("%s-%s", api.DatabaseNamePrefix, snapshotSource.Name)
	if _, err := c.Client.BatchV1().Jobs(snapshotSource.Namespace).Get(jobName, metav1.GetOptions{}); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	} else {
		return nil
	}

	// Event for notification that kubernetes objects are creating
	c.recorder.Eventf(
		mysql,
		core.EventTypeNormal,
		eventer.EventReasonInitializing,
		`Initializing from Snapshot: "%v"`,
		snapshotSource.Name,
	)

	namespace := snapshotSource.Namespace
	if namespace == "" {
		namespace = mysql.Namespace
	}
	snapshot, err := c.ExtClient.KubedbV1alpha1().Snapshots(namespace).Get(snapshotSource.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	secret, err := storage.NewOSMSecret(c.Client, snapshot.OSMSecretName(), snapshot.Namespace, snapshot.Spec.Backend)
	if err != nil {
		return err
	}
	_, err = c.Client.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil && !kerr.IsAlreadyExists(err) {
		return err
	}

	job, err := c.createRestoreJob(mysql, snapshot)
	if err != nil {
		return err
	}

	if err := c.SetJobOwnerReference(snapshot, job); err != nil {
		return err
	}

	return nil
}

func (c *Controller) terminate(mysql *api.MySQL) error {
	owner := metav1.NewControllerRef(mysql, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

	// If TerminationPolicy is "pause", keep everything (ie, PVCs,Secrets,Snapshots) intact.
	// In operator, create dormantdatabase
	if mysql.Spec.TerminationPolicy == api.TerminationPolicyPause {
		if err := c.removeOwnerReferenceFromOffshoots(mysql); err != nil {
			return err
		}

		if _, err := c.createDormantDatabase(mysql); err != nil {
			if kerr.IsAlreadyExists(err) {
				// if already exists, check if it is database of another Kind and return error in that case.
				// If the Kind is same, we can safely assume that the DormantDB was not deleted in before,
				// Probably because, User is more faster (create-delete-create-again-delete...) than operator!
				// So reuse that DormantDB!
				ddb, err := c.ExtClient.KubedbV1alpha1().DormantDatabases(mysql.Namespace).Get(mysql.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if val, _ := meta_util.GetStringValue(ddb.Labels, api.LabelDatabaseKind); val != api.ResourceKindMySQL {
					return fmt.Errorf(`DormantDatabase "%v" of kind %v already exists`, mysql.Name, val)
				}
			} else {
				return fmt.Errorf(`failed to create DormantDatabase: "%v/%v". Reason: %v`, mysql.Namespace, mysql.Name, err)
			}
		}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshoots(mysql, owner); err != nil {
			return err
		}
	}

	c.cronController.StopBackupScheduling(mysql.ObjectMeta)

	if mysql.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(mysql); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}

func (c *Controller) setOwnerReferenceToOffshoots(mysql *api.MySQL, owner *metav1.OwnerReference) error {
	selector := labels.SelectorFromSet(mysql.OffshootSelectors())

	// If TerminationPolicy is "wipeOut", delete snapshots and secrets,
	// else, keep it intact.
	if mysql.Spec.TerminationPolicy == api.TerminationPolicyWipeOut {
		if err := dynamic_util.EnsureOwnerReferenceForSelector(
			c.DynamicClient,
			api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
			mysql.Namespace,
			selector,
			owner); err != nil {
			return err
		}
		if err := c.wipeOutDatabase(mysql.ObjectMeta, mysql.Spec.GetSecrets(), owner); err != nil {
			return errors.Wrap(err, "error in wiping out database.")
		}
	} else {
		// Make sure snapshot and secret's ownerreference is removed.
		if err := dynamic_util.RemoveOwnerReferenceForSelector(
			c.DynamicClient,
			api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
			mysql.Namespace,
			selector,
			mysql); err != nil {
			return err
		}
		if err := dynamic_util.RemoveOwnerReferenceForItems(
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			mysql.Namespace,
			mysql.Spec.GetSecrets(),
			mysql); err != nil {
			return err
		}
	}
	// delete PVC for both "wipeOut" and "delete" TerminationPolicy.
	return dynamic_util.EnsureOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		mysql.Namespace,
		selector,
		owner)
}

func (c *Controller) removeOwnerReferenceFromOffshoots(mysql *api.MySQL) error {
	// First, Get LabelSelector for Other Components
	labelSelector := labels.SelectorFromSet(mysql.OffshootSelectors())

	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		c.DynamicClient,
		api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
		mysql.Namespace,
		labelSelector,
		mysql); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		mysql.Namespace,
		labelSelector,
		mysql); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForItems(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		mysql.Namespace,
		mysql.Spec.GetSecrets(),
		mysql); err != nil {
		return err
	}
	return nil
}
