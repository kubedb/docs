package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	mon_api "github.com/appscode/kube-mon/api"
	"github.com/appscode/kutil"
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	"github.com/kubedb/apimachinery/pkg/storage"
	"github.com/kubedb/mongodb/pkg/validator"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (c *Controller) create(mongodb *api.MongoDB) error {
	if err := validator.ValidateMongoDB(c.Client, mongodb); err != nil {
		c.recorder.Event(
			mongodb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error())
		log.Errorln(err)
		return nil
	}

	if mongodb.Status.CreationTime == nil {
		mg, _, err := util.PatchMongoDB(c.ExtClient, mongodb, func(in *api.MongoDB) *api.MongoDB {
			t := metav1.Now()
			in.Status.CreationTime = &t
			in.Status.Phase = api.DatabasePhaseCreating
			return in
		})
		if err != nil {
			c.recorder.Eventf(
				mongodb.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return err
		}
		mongodb.Status = mg.Status
	}

	// Dynamic Defaulting
	// Assign Default Monitoring Port
	if err := c.setMonitoringPort(mongodb); err != nil {
		return err
	}

	// Check DormantDatabase
	// It can be used as resumed
	if err := c.matchDormantDatabase(mongodb); err != nil {
		return err
	}

	// create Governing Service
	governingService := c.opt.GoverningService
	if err := c.CreateGoverningService(governingService, mongodb.Namespace); err != nil {
		c.recorder.Eventf(
			mongodb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create Service: "%v". Reason: %v`,
			governingService,
			err,
		)
		return err
	}

	// ensure database Service
	vt1, err := c.ensureService(mongodb)
	if err != nil {
		return err
	}

	// ensure database StatefulSet
	vt2, err := c.ensureStatefulSet(mongodb)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.recorder.Event(
			mongodb.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created MongoDB",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			mongodb.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched MongoDB",
		)
	}

	if _, err := meta_util.GetString(mongodb.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		mongodb.Spec.Init != nil && mongodb.Spec.Init.SnapshotSource != nil {

		snapshotSource := mongodb.Spec.Init.SnapshotSource

		if mongodb.Status.Phase == api.DatabasePhaseInitializing {
			return nil
		}
		jobName := fmt.Sprintf("%s-%s", api.DatabaseNamePrefix, snapshotSource.Name)
		if _, err := c.Client.BatchV1().Jobs(snapshotSource.Namespace).Get(jobName, metav1.GetOptions{}); err != nil {
			if kerr.IsAlreadyExists(err) {
				return nil
			} else if !kerr.IsNotFound(err) {
				return err
			}
		}
		if err := c.initialize(mongodb); err != nil {
			return fmt.Errorf("failed to complete initialization. Reason: %v", err)
		}
		return nil
	}

	ms, _, err := util.PatchMongoDB(c.ExtClient, mongodb, func(in *api.MongoDB) *api.MongoDB {
		in.Status.Phase = api.DatabasePhaseRunning
		return in
	})
	if err != nil {
		c.recorder.Eventf(
			mongodb,
			core.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			err.Error(),
		)
		return err
	}
	mongodb.Status = ms.Status

	// Ensure Schedule backup
	c.ensureBackupScheduler(mongodb)

	if err := c.manageMonitor(mongodb); err != nil {
		c.recorder.Eventf(
			mongodb.ObjectReference(),
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

// Assign Default Monitoring Port if MonitoringSpec Exists
// and the AgentVendor is Prometheus.
func (c *Controller) setMonitoringPort(mongodb *api.MongoDB) error {
	if mongodb.Spec.Monitor != nil &&
		mongodb.GetMonitoringVendor() == mon_api.VendorPrometheus {
		if mongodb.Spec.Monitor.Prometheus == nil {
			mongodb.Spec.Monitor.Prometheus = &mon_api.PrometheusSpec{}
		}
		if mongodb.Spec.Monitor.Prometheus.Port == 0 {
			mg, _, err := util.PatchMongoDB(c.ExtClient, mongodb, func(in *api.MongoDB) *api.MongoDB {
				in.Spec.Monitor.Prometheus.Port = api.PrometheusExporterPortNumber
				return in
			})

			if err != nil {
				c.recorder.Eventf(
					mongodb.ObjectReference(),
					core.EventTypeWarning,
					eventer.EventReasonFailedToUpdate,
					err.Error(),
				)
				return err
			}
			mongodb.Spec = mg.Spec
		}
	}
	return nil
}

func (c *Controller) matchDormantDatabase(mongodb *api.MongoDB) error {
	// Check if DormantDatabase exists or not
	dormantDb, err := c.ExtClient.DormantDatabases(mongodb.Namespace).Get(mongodb.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			c.recorder.Eventf(
				mongodb.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToGet,
				`Fail to get DormantDatabase: "%v". Reason: %v`,
				mongodb.Name,
				err,
			)
			return err
		}
		return nil
	}

	var sendEvent = func(message string, args ...interface{}) error {
		c.recorder.Eventf(
			mongodb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			message,
			args,
		)
		return fmt.Errorf(message, args)
	}

	// Check DatabaseKind
	if dormantDb.Labels[api.LabelDatabaseKind] != api.ResourceKindMongoDB {
		return sendEvent(fmt.Sprintf(`Invalid MongoDB: "%v". Exists DormantDatabase "%v" of different Kind`,
			mongodb.Name, dormantDb.Name))
	}

	// Check Origin Spec
	drmnOriginSpec := dormantDb.Spec.Origin.Spec.MongoDB
	originalSpec := mongodb.Spec

	if originalSpec.DatabaseSecret == nil {
		originalSpec.DatabaseSecret = &core.SecretVolumeSource{
			SecretName: mongodb.Name + "-auth",
		}
	}

	// Skip checking doNotPause
	drmnOriginSpec.DoNotPause = originalSpec.DoNotPause

	if !meta_util.Equal(drmnOriginSpec, originalSpec) {
		return sendEvent("MongoDB spec mismatches with OriginSpec in DormantDatabases")
	}

	if _, err := meta_util.GetString(mongodb.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		mongodb.Spec.Init != nil &&
		mongodb.Spec.Init.SnapshotSource != nil {
		mg, _, err := util.PatchMongoDB(c.ExtClient, mongodb, func(in *api.MongoDB) *api.MongoDB {
			in.Annotations = core_util.UpsertMap(in.Annotations, map[string]string{
				api.AnnotationInitialized: "",
			})
			return in
		})
		if err != nil {
			return err
		}
		mongodb.Annotations = mg.Annotations
	}

	return util.DeleteDormantDatabase(c.ExtClient, dormantDb.ObjectMeta)
}

func (c *Controller) ensureBackupScheduler(mongodb *api.MongoDB) {
	// Setup Schedule backup
	if mongodb.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(mongodb, mongodb.ObjectMeta, mongodb.Spec.BackupSchedule)
		if err != nil {
			c.recorder.Eventf(
				mongodb.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToSchedule,
				"Failed to schedule snapshot. Reason: %v",
				err,
			)
			log.Errorln(err)
		}
	} else {
		c.cronController.StopBackupScheduling(mongodb.ObjectMeta)
	}
}

func (c *Controller) initialize(mongodb *api.MongoDB) error {
	mg, _, err := util.PatchMongoDB(c.ExtClient, mongodb, func(in *api.MongoDB) *api.MongoDB {
		in.Status.Phase = api.DatabasePhaseInitializing
		return in
	})
	if err != nil {
		c.recorder.Eventf(mongodb, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}
	mongodb.Status = mg.Status

	snapshotSource := mongodb.Spec.Init.SnapshotSource
	// Event for notification that kubernetes objects are creating
	c.recorder.Eventf(
		mongodb.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonInitializing,
		`Initializing from Snapshot: "%v"`,
		snapshotSource.Name,
	)

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
		c.recorder.Eventf(
			mongodb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create DormantDatabase: "%v". Reason: %v`,
			mongodb.Name,
			err,
		)
		return err
	}
	c.recorder.Eventf(
		mongodb.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		`Successfully created DormantDatabase: "%v"`,
		mongodb.Name,
	)

	c.cronController.StopBackupScheduling(mongodb.ObjectMeta)

	if mongodb.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(mongodb); err != nil {
			c.recorder.Eventf(
				mongodb.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToDelete,
				"Failed to delete monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
	}
	return nil
}

func (c *Controller) GetDatabase(meta metav1.ObjectMeta) (runtime.Object, error) {
	mongodb, err := c.ExtClient.MongoDBs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return mongodb, nil
}

func (c *Controller) SetDatabaseStatus(meta metav1.ObjectMeta, phase api.DatabasePhase, reason string) error {
	mongodb, err := c.ExtClient.MongoDBs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, _, err = util.PatchMongoDB(c.ExtClient, mongodb, func(in *api.MongoDB) *api.MongoDB {
		in.Status.Phase = phase
		in.Status.Reason = reason
		return in
	})
	return err
}

func (c *Controller) UpsertDatabaseAnnotation(meta metav1.ObjectMeta, annotation map[string]string) error {
	mongodb, err := c.ExtClient.MongoDBs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	_, _, err = util.PatchMongoDB(c.ExtClient, mongodb, func(in *api.MongoDB) *api.MongoDB {
		in.Annotations = core_util.UpsertMap(mongodb.Annotations, annotation)
		return in
	})
	return err
}
