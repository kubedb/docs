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
	"github.com/kubedb/mysql/pkg/validator"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (c *Controller) create(mysql *api.MySQL) error {
	if err := validator.ValidateMySQL(c.Client, mysql); err != nil {
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
			"Successfully created MySQL",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			mysql.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched MySQL",
		)
	}

	if _, err := meta_util.GetString(mysql.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		mysql.Spec.Init != nil && mysql.Spec.Init.SnapshotSource != nil {

		snapshotSource := mysql.Spec.Init.SnapshotSource

		if mysql.Status.Phase == api.DatabasePhaseInitializing {
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
		if err := c.initialize(mysql); err != nil {
			return fmt.Errorf("failed to complete initialization. Reason: %v", err)
		}
		return nil
	}

	ms, _, err := util.PatchMySQL(c.ExtClient, mysql, func(in *api.MySQL) *api.MySQL {
		in.Status.Phase = api.DatabasePhaseRunning
		return in
	})
	if err != nil {
		c.recorder.Eventf(
			mysql,
			core.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			err.Error(),
		)
		return err
	}
	mysql.Status = ms.Status

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

// Assign Default Monitoring Port if MonitoringSpec Exists
// and the AgentVendor is Prometheus.
func (c *Controller) setMonitoringPort(mysql *api.MySQL) error {
	if mysql.Spec.Monitor != nil &&
		mysql.GetMonitoringVendor() == mon_api.VendorPrometheus {
		if mysql.Spec.Monitor.Prometheus == nil {
			mysql.Spec.Monitor.Prometheus = &mon_api.PrometheusSpec{}
		}
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

	// Check Origin Spec
	drmnOriginSpec := dormantDb.Spec.Origin.Spec.MySQL
	originalSpec := mysql.Spec

	if originalSpec.DatabaseSecret == nil {
		originalSpec.DatabaseSecret = &core.SecretVolumeSource{
			SecretName: mysql.Name + "-auth",
		}
	}

	// Skip checking doNotPause
	drmnOriginSpec.DoNotPause = originalSpec.DoNotPause

	if !meta_util.Equal(drmnOriginSpec, originalSpec) {
		return sendEvent("MySQL spec mismatches with OriginSpec in DormantDatabases")
	}

	if _, err := meta_util.GetString(mysql.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		mysql.Spec.Init != nil &&
		mysql.Spec.Init.SnapshotSource != nil {
		mg, _, err := util.PatchMySQL(c.ExtClient, mysql, func(in *api.MySQL) *api.MySQL {
			in.Annotations = core_util.UpsertMap(in.Annotations, map[string]string{
				api.AnnotationInitialized: "",
			})
			return in
		})
		if err != nil {
			return err
		}
		mysql.Annotations = mg.Annotations
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

func (c *Controller) initialize(mysql *api.MySQL) error {
	mg, _, err := util.PatchMySQL(c.ExtClient, mysql, func(in *api.MySQL) *api.MySQL {
		in.Status.Phase = api.DatabasePhaseInitializing
		return in
	})
	if err != nil {
		c.recorder.Eventf(mysql, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}
	mysql.Status = mg.Status

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

	if err := c.SetJobOwnerReference(snapshot, job); err != nil {
		return err
	}

	return nil
}

func (c *Controller) pause(mysql *api.MySQL) error {
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

func (c *Controller) GetDatabase(meta metav1.ObjectMeta) (runtime.Object, error) {
	mysql, err := c.ExtClient.MySQLs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return mysql, nil
}

func (c *Controller) SetDatabaseStatus(meta metav1.ObjectMeta, phase api.DatabasePhase, reason string) error {
	mysql, err := c.ExtClient.MySQLs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, _, err = util.PatchMySQL(c.ExtClient, mysql, func(in *api.MySQL) *api.MySQL {
		in.Status.Phase = phase
		in.Status.Reason = reason
		return in
	})
	return err
}

func (c *Controller) UpsertDatabaseAnnotation(meta metav1.ObjectMeta, annotation map[string]string) error {
	mysql, err := c.ExtClient.MySQLs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	_, _, err = util.PatchMySQL(c.ExtClient, mysql, func(in *api.MySQL) *api.MySQL {
		in.Annotations = core_util.UpsertMap(mysql.Annotations, annotation)
		return in
	})
	return err
}
