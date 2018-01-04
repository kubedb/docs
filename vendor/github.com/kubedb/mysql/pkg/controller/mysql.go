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
	"github.com/kubedb/mysql/pkg/validator"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) create(mysql *api.MySQL) error {
	if err := validator.ValidateMySQL(c.Client, mysql, &c.opt.Docker); err != nil {
		c.recorder.Event(
			mysql.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error())
		log.Errorln(err)
		return nil
	}

	if mysql.Status.CreationTime == nil {
		ms, _, err := util.PatchMySQL(c.ExtClient, mysql, func(in *api.MySQL) *api.MySQL {
			t := metav1.Now()
			in.Status.CreationTime = &t
			in.Status.Phase = api.DatabasePhaseCreating
			return in
		})
		if err != nil {
			c.recorder.Eventf(
				mysql.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return err
		}
		mysql.Status = ms.Status
	}

	// Dynamic Defaulting
	// Assign Default Monitoring Port
	if err := c.setMonitoringPort(mysql); err != nil {
		return err
	}

	// Check DormantDatabase
	// It can be used as resumed
	if err := c.matchDormantDatabase(mysql); err != nil {
		return err
	}

	// create Governing Service
	governingService := c.opt.GoverningService
	if err := c.CreateGoverningService(governingService, mysql.Namespace); err != nil {
		c.recorder.Eventf(
			mysql.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create Service: "%v". Reason: %v`,
			governingService,
			err,
		)
		return err
	}

	// ensure database Service
	vt1, err := c.ensureService(mysql)
	if err != nil {
		return err
	}

	// ensure database StatefulSet
	vt2, err := c.ensureStatefulSet(mysql)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.recorder.Event(
			mysql.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created Elasticsearch",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			mysql.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched Elasticsearch",
		)
	}

	if vt2 == kutil.VerbCreated && mysql.Spec.Init != nil && mysql.Spec.Init.SnapshotSource != nil {
		es, _, err := util.PatchMySQL(c.ExtClient, mysql, func(in *api.MySQL) *api.MySQL {
			in.Status.Phase = api.DatabasePhaseInitializing
			return in
		})
		if err != nil {
			c.recorder.Eventf(
				mysql.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return err
		}
		mysql.Status = es.Status

		if err := c.initialize(mysql); err != nil {
			c.recorder.Eventf(
				mysql.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToInitialize,
				"Failed to initialize. Reason: %v",
				err,
			)
		}

		es, _, err = util.PatchMySQL(c.ExtClient, mysql, func(in *api.MySQL) *api.MySQL {
			in.Status.Phase = api.DatabasePhaseRunning
			return in
		})
		if err != nil {
			c.recorder.Eventf(
				mysql.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return err
		}
		mysql.Status = es.Status
	}

	// Ensure Schedule backup
	c.ensureBackupScheduler(mysql)

	if err := c.manageMonitor(mysql); err != nil {
		c.recorder.Eventf(
			mysql.ObjectReference(),
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

func (c *Controller) setMonitoringPort(mysql *api.MySQL) error {
	if mysql.Spec.Monitor != nil &&
		mysql.Spec.Monitor.Prometheus != nil {
		if mysql.Spec.Monitor.Prometheus.Port == 0 {
			ms, _, err := util.PatchMySQL(c.ExtClient, mysql, func(in *api.MySQL) *api.MySQL {
				in.Spec.Monitor.Prometheus.Port = api.PrometheusExporterPortNumber
				return in
			})

			if err != nil {
				c.recorder.Eventf(
					mysql.ObjectReference(),
					core.EventTypeWarning,
					eventer.EventReasonFailedToUpdate,
					err.Error(),
				)
				return err
			}
			mysql.Spec = ms.Spec
		}
	}
	return nil
}

func (c *Controller) matchDormantDatabase(mysql *api.MySQL) error {
	// Check if DormantDatabase exists or not
	dormantDb, err := c.ExtClient.DormantDatabases(mysql.Namespace).Get(mysql.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			c.recorder.Eventf(
				mysql.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToGet,
				`Fail to get DormantDatabase: "%v". Reason: %v`,
				mysql.Name,
				err,
			)
			return err
		}
		return nil
	}

	var sendEvent = func(message string, args ...interface{}) error {
		c.recorder.Eventf(
			mysql.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			message,
			args,
		)
		return fmt.Errorf(message, args)
	}

	// Check DatabaseKind
	if dormantDb.Labels[api.LabelDatabaseKind] != api.ResourceKindMySQL {
		return sendEvent(fmt.Sprintf(`Invalid MySQL: "%v". Exists DormantDatabase "%v" of different Kind`,
			mysql.Name, dormantDb.Name))
	}

	// Check InitSpec
	initSpecAnnotationStr := dormantDb.Annotations[api.GenericInitSpec]
	if initSpecAnnotationStr != "" {
		var initSpecAnnotation *api.InitSpec
		if err := json.Unmarshal([]byte(initSpecAnnotationStr), &initSpecAnnotation); err != nil {
			return sendEvent(err.Error())
		}

		if mysql.Spec.Init != nil {
			if !reflect.DeepEqual(initSpecAnnotation, mysql.Spec.Init) {
				return sendEvent("InitSpec mismatches with DormantDatabase annotation")
			}
		}
	}

	// Check Origin Spec
	drmnOriginSpec := dormantDb.Spec.Origin.Spec.MySQL
	originalSpec := mysql.Spec
	originalSpec.Init = nil

	if originalSpec.DatabaseSecret == nil {
		originalSpec.DatabaseSecret = &core.SecretVolumeSource{
			SecretName: mysql.Name + "-auth",
		}
	}

	if !reflect.DeepEqual(drmnOriginSpec, &originalSpec) {
		return sendEvent("MySQL spec mismatches with OriginSpec in DormantDatabases")
	}

	return util.DeleteDormantDatabase(c.ExtClient, dormantDb.ObjectMeta)
}

func (c *Controller) ensureBackupScheduler(mysql *api.MySQL) {
	// Setup Schedule backup
	if mysql.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(mysql, mysql.ObjectMeta, mysql.Spec.BackupSchedule)
		if err != nil {
			c.recorder.Eventf(
				mysql.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToSchedule,
				"Failed to schedule snapshot. Reason: %v",
				err,
			)
			log.Errorln(err)
		}
	} else {
		c.cronController.StopBackupScheduling(mysql.ObjectMeta)
	}
}

const (
	durationCheckRestoreJob = time.Minute * 30
)

func (c *Controller) initialize(mysql *api.MySQL) error {
	snapshotSource := mysql.Spec.Init.SnapshotSource
	// Event for notification that kubernetes objects are creating
	c.recorder.Eventf(
		mysql.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonInitializing,
		`Initializing from Snapshot: "%v"`,
		snapshotSource.Name,
	)

	namespace := snapshotSource.Namespace
	if namespace == "" {
		namespace = mysql.Namespace
	}
	snapshot, err := c.ExtClient.Snapshots(namespace).Get(snapshotSource.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if err := docker.CheckDockerImageVersion(c.opt.Docker.GetToolsImage(mysql), string(mysql.Spec.Version)); err != nil {
		return fmt.Errorf(`image %s not found`, c.opt.Docker.GetToolsImageWithTag(mysql))
	}

	secret, err := storage.NewOSMSecret(c.Client, snapshot)
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

	jobSuccess := c.CheckDatabaseRestoreJob(snapshot, job, mysql, c.recorder, durationCheckRestoreJob)
	if jobSuccess {
		c.recorder.Event(
			mysql.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulInitialize,
			"Successfully completed initialization",
		)
	} else {
		c.recorder.Event(
			mysql.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToInitialize,
			"Failed to complete initialization",
		)
	}
	return nil
}

func (c *Controller) pause(mysql *api.MySQL) error {
	//if mysql.Spec.DoNotPause {
	//	c.recorder.Eventf(
	//		mysql.ObjectReference(),
	//		core.EventTypeWarning,
	//		eventer.EventReasonFailedToPause,
	//		`MySQL "%v" is locked.`,
	//		mysql.Name,
	//	)
	//
	//	if err := c.reCreateMySQL(mysql); err != nil {
	//		c.recorder.Eventf(
	//			mysql.ObjectReference(),
	//			core.EventTypeWarning,
	//			eventer.EventReasonFailedToCreate,
	//			`Failed to recreate MySQL: "%v". Reason: %v`,
	//			mysql.Name,
	//			err,
	//		)
	//		return err
	//	}
	//	return nil
	//}

	if _, err := c.createDormantDatabase(mysql); err != nil {
		c.recorder.Eventf(
			mysql.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create DormantDatabase: "%v". Reason: %v`,
			mysql.Name,
			err,
		)
		return err
	}
	c.recorder.Eventf(
		mysql.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		`Successfully created DormantDatabase: "%v"`,
		mysql.Name,
	)

	c.cronController.StopBackupScheduling(mysql.ObjectMeta)

	if mysql.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(mysql); err != nil {
			c.recorder.Eventf(
				mysql.ObjectReference(),
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
