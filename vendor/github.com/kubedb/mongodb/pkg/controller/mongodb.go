package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/kutil"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	"github.com/kubedb/apimachinery/pkg/storage"
	validator "github.com/kubedb/mongodb/pkg/admission"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
)

func (c *Controller) create(mongodb *api.MongoDB) error {
	if err := validator.ValidateMongoDB(c.Client, c.ExtClient, mongodb); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb); rerr == nil {
			c.recorder.Event(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonInvalid,
				err.Error())
		}
		log.Errorln(err)
		return nil
	}

	// Delete Matching DormantDatabase if exists any
	if err := c.deleteMatchingDormantDatabase(mongodb); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to delete dormant Database : "%v". Reason: %v`,
				mongodb.Name,
				err,
			)
		}
		return err
	}

	if mongodb.Status.CreationTime == nil {
		mg, err := util.UpdateMongoDBStatus(c.ExtClient, mongodb, func(in *api.MongoDBStatus) *api.MongoDBStatus {
			t := metav1.Now()
			in.CreationTime = &t
			in.Phase = api.DatabasePhaseCreating
			return in
		}, api.EnableStatusSubresource)
		if err != nil {
			if ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb); rerr == nil {
				c.recorder.Eventf(
					ref,
					core.EventTypeWarning,
					eventer.EventReasonFailedToUpdate,
					err.Error(),
				)
			}
			return err
		}
		mongodb.Status = mg.Status
	}

	// create Governing Service
	governingService := c.GoverningService
	if err := c.CreateGoverningService(governingService, mongodb.Namespace); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to create Service: "%v". Reason: %v`,
				governingService,
				err,
			)
		}
		return err
	}

	// ensure database Service
	vt1, err := c.ensureService(mongodb)
	if err != nil {
		return err
	}

	if err := c.ensureDatabaseSecret(mongodb); err != nil {
		return err
	}

	// ensure database StatefulSet
	vt2, err := c.ensureStatefulSet(mongodb)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb); rerr == nil {
			c.recorder.Event(
				ref,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully created MongoDB",
			)
		}
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb); rerr == nil {
			c.recorder.Event(
				ref,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully patched MongoDB",
			)
		}
	}

	if _, err := meta_util.GetString(mongodb.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		mongodb.Spec.Init != nil && mongodb.Spec.Init.SnapshotSource != nil {

		snapshotSource := mongodb.Spec.Init.SnapshotSource

		if mongodb.Status.Phase == api.DatabasePhaseInitializing {
			return nil
		}
		jobName := fmt.Sprintf("%s-%s", api.DatabaseNamePrefix, snapshotSource.Name)
		if _, err := c.Client.BatchV1().Jobs(snapshotSource.Namespace).Get(jobName, metav1.GetOptions{}); err != nil {
			if !kerr.IsNotFound(err) {
				return err
			}
		} else {
			return nil
		}

		if err := c.initialize(mongodb); err != nil {
			return fmt.Errorf("failed to complete initialization. Reason: %v", err)
		}
		return nil
	}

	ms, err := util.UpdateMongoDBStatus(c.ExtClient, mongodb, func(in *api.MongoDBStatus) *api.MongoDBStatus {
		in.Phase = api.DatabasePhaseRunning
		return in
	}, api.EnableStatusSubresource)
	if err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
		}
		return err
	}
	mongodb.Status = ms.Status

	// Ensure Schedule backup
	c.ensureBackupScheduler(mongodb)

	if err := c.manageMonitor(mongodb); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				"Failed to manage monitoring system. Reason: %v",
				err,
			)
		}
		log.Errorln(err)
		return nil
	}

	return nil
}

func (c *Controller) ensureBackupScheduler(mongodb *api.MongoDB) {
	// Setup Schedule backup
	if mongodb.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(mongodb, mongodb.ObjectMeta, mongodb.Spec.BackupSchedule)
		if err != nil {
			log.Errorln(err)
			if ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb); rerr == nil {
				c.recorder.Eventf(
					ref,
					core.EventTypeWarning,
					eventer.EventReasonFailedToSchedule,
					"Failed to schedule snapshot. Reason: %v",
					err,
				)
			}
		}
	} else {
		c.cronController.StopBackupScheduling(mongodb.ObjectMeta)
	}
}

func (c *Controller) initialize(mongodb *api.MongoDB) error {
	mg, err := util.UpdateMongoDBStatus(c.ExtClient, mongodb, func(in *api.MongoDBStatus) *api.MongoDBStatus {
		in.Phase = api.DatabasePhaseInitializing
		return in
	}, api.EnableStatusSubresource)
	if err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
		}
		return err
	}
	mongodb.Status = mg.Status

	snapshotSource := mongodb.Spec.Init.SnapshotSource
	// Event for notification that kubernetes objects are creating
	if ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb); rerr == nil {
		c.recorder.Eventf(
			ref,
			core.EventTypeNormal,
			eventer.EventReasonInitializing,
			`Initializing from Snapshot: "%v"`,
			snapshotSource.Name,
		)
	}

	namespace := snapshotSource.Namespace
	if namespace == "" {
		namespace = mongodb.Namespace
	}
	snapshot, err := c.ExtClient.Snapshots(namespace).Get(snapshotSource.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	secret, err := storage.NewOSMSecret(c.Client, snapshot)
	if err != nil {
		return err
	}
	secret, err = c.Client.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil && !kerr.IsAlreadyExists(err) {
		return err
	}

	job, err := c.createRestoreJob(mongodb, snapshot)
	if err != nil {
		return err
	}

	if err := c.SetJobOwnerReference(snapshot, job); err != nil {
		return err
	}

	return nil
}

func (c *Controller) pause(mongodb *api.MongoDB) error {
	if _, err := c.createDormantDatabase(mongodb); err != nil {
		if kerr.IsAlreadyExists(err) {
			// if already exists, check if it is database of another Kind and return error in that case.
			// If the Kind is same, we can safely assume that the DormantDB was not deleted in before,
			// Probably because, User is more faster (create-delete-create-again-delete...) than operator!
			// So reuse that DormantDB!
			ddb, err := c.ExtClient.DormantDatabases(mongodb.Namespace).Get(mongodb.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if val, _ := meta_util.GetStringValue(ddb.Labels, api.LabelDatabaseKind); val != api.ResourceKindMongoDB {
				return fmt.Errorf(`DormantDatabase "%v" of kind %v already exists`, mongodb.Name, val)
			}
		} else {
			return fmt.Errorf(`Failed to create DormantDatabase: "%v". Reason: %v`, mongodb.Name, err)
		}
	}

	c.cronController.StopBackupScheduling(mongodb.ObjectMeta)

	if mongodb.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(mongodb); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}
