package controller

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/kutil"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/docker"
	"github.com/kubedb/apimachinery/pkg/eventer"
	"github.com/kubedb/apimachinery/pkg/storage"
	"github.com/kubedb/mongodb/pkg/validator"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) create(mongodb *api.MongoDB) error {
	if err := validator.ValidateMongoDB(c.Client, mongodb, &c.opt.Docker); err != nil {
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
			"Successfully created Elasticsearch",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			mongodb.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched Elasticsearch",
		)
	}

	if vt2 == kutil.VerbCreated && mongodb.Spec.Init != nil && mongodb.Spec.Init.SnapshotSource != nil {
		es, _, err := util.PatchMongoDB(c.ExtClient, mongodb, func(in *api.MongoDB) *api.MongoDB {
			in.Status.Phase = api.DatabasePhaseInitializing
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
		mongodb.Status = es.Status

		if err := c.initialize(mongodb); err != nil {
			c.recorder.Eventf(
				mongodb.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToInitialize,
				"Failed to initialize. Reason: %v",
				err,
			)
		}

		es, _, err = util.PatchMongoDB(c.ExtClient, mongodb, func(in *api.MongoDB) *api.MongoDB {
			in.Status.Phase = api.DatabasePhaseRunning
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
		mongodb.Status = es.Status
	}

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

func (c *Controller) setMonitoringPort(mongodb *api.MongoDB) error {
	if mongodb.Spec.Monitor != nil &&
		mongodb.Spec.Monitor.Prometheus != nil {
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

	// Check InitSpec
	initSpecAnnotationStr := dormantDb.Annotations[api.GenericInitSpec]
	if initSpecAnnotationStr != "" {
		var initSpecAnnotation *api.InitSpec
		if err := json.Unmarshal([]byte(initSpecAnnotationStr), &initSpecAnnotation); err != nil {
			return sendEvent(err.Error())
		}

		if mongodb.Spec.Init != nil {
			if !reflect.DeepEqual(initSpecAnnotation, mongodb.Spec.Init) {
				return sendEvent("InitSpec mismatches with DormantDatabase annotation")
			}
		}
	}

	// Check Origin Spec
	drmnOriginSpec := dormantDb.Spec.Origin.Spec.MongoDB
	originalSpec := mongodb.Spec
	originalSpec.Init = nil

	if originalSpec.DatabaseSecret == nil {
		originalSpec.DatabaseSecret = &core.SecretVolumeSource{
			SecretName: mongodb.Name + "-auth",
		}
	}

	if !reflect.DeepEqual(drmnOriginSpec, &originalSpec) {
		return sendEvent("MongoDB spec mismatches with OriginSpec in DormantDatabases")
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

const (
	durationCheckRestoreJob = time.Minute * 30
)

func (c *Controller) initialize(mongodb *api.MongoDB) error {
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

	if err := docker.CheckDockerImageVersion(c.opt.Docker.GetToolsImage(mongodb), string(mongodb.Spec.Version)); err != nil {
		return fmt.Errorf(`image %s not found`, c.opt.Docker.GetToolsImageWithTag(mongodb))
	}

	secret, err := storage.NewOSMSecret(c.Client, snapshot)
	if err != nil {
		return err
	}
	_, err = c.Client.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil && !kerr.IsAlreadyExists(err) {
		return err
	}

	job, err := c.createRestoreJob(mongodb, snapshot)
	if err != nil {
		return err
	}

	jobSuccess := c.CheckDatabaseRestoreJob(snapshot, job, mongodb, c.recorder, durationCheckRestoreJob)
	if jobSuccess {
		c.recorder.Event(
			mongodb.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulInitialize,
			"Successfully completed initialization",
		)
	} else {
		c.recorder.Event(
			mongodb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToInitialize,
			"Failed to complete initialization",
		)
	}
	return nil
}

func (c *Controller) pause(mongodb *api.MongoDB) error {
	//if mongodb.Spec.DoNotPause {
	//	c.recorder.Eventf(
	//		mongodb.ObjectReference(),
	//		core.EventTypeWarning,
	//		eventer.EventReasonFailedToPause,
	//		`MongoDB "%v" is locked.`,
	//		mongodb.Name,
	//	)
	//
	//	if err := c.reCreateMongoDB(mongodb); err != nil {
	//		c.recorder.Eventf(
	//			mongodb.ObjectReference(),
	//			core.EventTypeWarning,
	//			eventer.EventReasonFailedToCreate,
	//			`Failed to recreate MongoDB: "%v". Reason: %v`,
	//			mongodb.Name,
	//			err,
	//		)
	//		return err
	//	}
	//	return nil
	//}

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
