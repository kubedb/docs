package controller

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/kutil"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	kutildb "github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	"github.com/kubedb/apimachinery/pkg/storage"
	"github.com/kubedb/postgres/pkg/validator"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) create(postgres *api.Postgres) error {
	if err := validator.ValidatePostgres(c.Client, postgres, c.opt.Docker); err != nil {
		c.recorder.Event(
			postgres.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		return nil // user error so just record error and don't retry.
	}

	if postgres.Status.CreationTime == nil {
		es, _, err := kutildb.PatchPostgres(c.ExtClient, postgres, func(in *api.Postgres) *api.Postgres {
			t := metav1.Now()
			in.Status.CreationTime = &t
			in.Status.Phase = api.DatabasePhaseCreating
			return in
		})
		if err != nil {
			c.recorder.Eventf(
				postgres.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return err
		}
		postgres.Status = es.Status
	}

	// Dynamic Defaulting
	// Assign Default Monitoring Port
	if err := c.setMonitoringPort(postgres); err != nil {
		return err
	}

	// Check DormantDatabase
	// It can be used as resumed
	if err := c.matchDormantDatabase(postgres); err != nil {
		return err
	}

	// create Governing Service
	governingService := c.opt.GoverningService
	if err := c.CreateGoverningService(governingService, postgres.Namespace); err != nil {
		c.recorder.Eventf(
			postgres.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create ServiceAccount: "%v". Reason: %v`,
			governingService,
			err,
		)
		return err
	}

	// ensure database Service
	vt1, err := c.ensureService(postgres)
	if err != nil {
		return err
	}

	// ensure database StatefulSet
	vt2, err := c.ensurePostgresNode(postgres)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.recorder.Event(
			postgres.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created Postgres",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			postgres.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched Postgres",
		)
	}

	if vt2 == kutil.VerbCreated && postgres.Spec.Init != nil && postgres.Spec.Init.SnapshotSource != nil {
		pg, _, err := kutildb.PatchPostgres(c.ExtClient, postgres, func(in *api.Postgres) *api.Postgres {
			in.Status.Phase = api.DatabasePhaseInitializing
			return in
		})
		if err != nil {
			c.recorder.Eventf(postgres, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}
		postgres.Status = pg.Status

		if err := c.initialize(postgres); err != nil {
			c.recorder.Eventf(
				postgres.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToInitialize,
				"Failed to initialize. Reason: %v",
				err,
			)
		}

		pg, _, err = kutildb.PatchPostgres(c.ExtClient, postgres, func(in *api.Postgres) *api.Postgres {
			in.Status.Phase = api.DatabasePhaseRunning
			return in
		})
		if err != nil {
			c.recorder.Eventf(postgres, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}
		postgres.Status = pg.Status
	}

	// Ensure Schedule backup
	c.ensureBackupScheduler(postgres)

	if err := c.manageMonitor(postgres); err != nil {
		c.recorder.Eventf(
			postgres.ObjectReference(),
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

func (c *Controller) setMonitoringPort(postgres *api.Postgres) error {
	if postgres.Spec.Monitor != nil &&
		postgres.Spec.Monitor.Prometheus != nil {
		if postgres.Spec.Monitor.Prometheus.Port == 0 {
			es, _, err := kutildb.PatchPostgres(c.ExtClient, postgres, func(in *api.Postgres) *api.Postgres {
				in.Spec.Monitor.Prometheus.Port = api.PrometheusExporterPortNumber
				return in
			})

			if err != nil {
				c.recorder.Eventf(
					postgres.ObjectReference(),
					core.EventTypeWarning,
					eventer.EventReasonFailedToUpdate,
					err.Error(),
				)
				return err
			}
			postgres.Spec.Monitor = es.Spec.Monitor
		}
	}
	return nil
}

func (c *Controller) matchDormantDatabase(postgres *api.Postgres) error {
	// Check if DormantDatabase exists or not
	dormantDb, err := c.ExtClient.DormantDatabases(postgres.Namespace).Get(postgres.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			c.recorder.Eventf(
				postgres.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToGet,
				`Fail to get DormantDatabase: "%v". Reason: %v`,
				postgres.Name,
				err,
			)
			return err
		}
		return nil
	}

	var sendEvent = func(message string, args ...interface{}) error {
		c.recorder.Eventf(
			postgres.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			message,
			args,
		)
		return fmt.Errorf(message, args)
	}

	// Check DatabaseKind
	if dormantDb.Labels[api.LabelDatabaseKind] != api.ResourceKindPostgres {
		return sendEvent(fmt.Sprintf(`Invalid Postgres: "%v". Exists DormantDatabase "%v" of different Kind`,
			postgres.Name, dormantDb.Name))
	}

	// Check InitSpec
	initSpecAnnotationStr := dormantDb.Annotations[api.PostgresInitSpec]
	if initSpecAnnotationStr != "" {
		var initSpecAnnotation *api.InitSpec
		if err := json.Unmarshal([]byte(initSpecAnnotationStr), &initSpecAnnotation); err != nil {
			return sendEvent(err.Error())
		}

		if postgres.Spec.Init != nil {
			if !reflect.DeepEqual(initSpecAnnotation, postgres.Spec.Init) {
				return sendEvent("InitSpec mismatches with DormantDatabase annotation")
			}
		}
	}

	// Check Origin Spec
	drmnOriginSpec := dormantDb.Spec.Origin.Spec.Postgres
	originalSpec := postgres.Spec
	originalSpec.Init = nil

	if originalSpec.DatabaseSecret == nil {
		originalSpec.DatabaseSecret = &core.SecretVolumeSource{
			SecretName: postgres.Name + "-auth",
		}
	}

	if !reflect.DeepEqual(drmnOriginSpec, &originalSpec) {
		return sendEvent("Postgres spec mismatches with OriginSpec in DormantDatabases")
	}

	return kutildb.DeleteDormantDatabase(c.ExtClient, dormantDb.ObjectMeta)
}

func (c *Controller) ensurePostgresNode(postgres *api.Postgres) (kutil.VerbType, error) {
	var err error

	if err = c.ensureDatabaseSecret(postgres); err != nil {
		return kutil.VerbUnchanged, err
	}

	if c.opt.EnableRbac {
		// Ensure ClusterRoles for database statefulsets
		if err := c.ensureRBACStuff(postgres); err != nil {
			return kutil.VerbUnchanged, err
		}
	}

	vt, err := c.ensureCombinedNode(postgres)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	pg, _, err := kutildb.PatchPostgres(c.ExtClient, postgres, func(in *api.Postgres) *api.Postgres {
		in.Status.Phase = api.DatabasePhaseRunning
		return in
	})
	if err != nil {
		c.recorder.Eventf(postgres.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return kutil.VerbUnchanged, err
	}
	postgres.Status = pg.Status

	return vt, nil
}

func (c *Controller) ensureBackupScheduler(postgres *api.Postgres) {
	kutildb.AssignTypeKind(postgres)
	// Setup Schedule backup
	if postgres.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(postgres, postgres.ObjectMeta, postgres.Spec.BackupSchedule)
		if err != nil {
			c.recorder.Eventf(
				postgres.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToSchedule,
				"Failed to schedule snapshot. Reason: %v",
				err,
			)
			log.Errorln(err)
		}
	} else {
		c.cronController.StopBackupScheduling(postgres.ObjectMeta)
	}
}

const (
	durationCheckRestoreJob = time.Minute * 30
)

func (c *Controller) initialize(postgres *api.Postgres) error {
	snapshotSource := postgres.Spec.Init.SnapshotSource
	// Event for notification that kubernetes objects are creating
	c.recorder.Eventf(
		postgres.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonInitializing,
		`Initializing from Snapshot: "%v"`,
		snapshotSource.Name,
	)

	namespace := snapshotSource.Namespace
	if namespace == "" {
		namespace = postgres.Namespace
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

	job, err := c.createRestoreJob(postgres, snapshot)
	if err != nil {
		return err
	}

	jobSuccess := c.CheckDatabaseRestoreJob(snapshot, job, postgres, c.recorder, durationCheckRestoreJob)
	if jobSuccess {
		c.recorder.Event(
			postgres.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulInitialize,
			"Successfully completed initialization",
		)
	} else {
		c.recorder.Event(
			postgres.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToInitialize,
			"Failed to complete initialization",
		)
	}
	return nil
}

func (c *Controller) pause(postgres *api.Postgres) error {

	c.recorder.Event(postgres.ObjectReference(), core.EventTypeNormal, eventer.EventReasonPausing, "Pausing Postgres")

	/*
		if postgres.Spec.DoNotPause {
			c.recorder.Eventf(
				postgres.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToPause,
				`Postgres "%v" is locked.`,
				postgres.Name,
			)

			if err := c.reCreatePostgres(postgres); err != nil {
				c.recorder.Eventf(
					postgres.ObjectReference(),
					core.EventTypeWarning,
					eventer.EventReasonFailedToCreate,
					`Failed to recreate Postgres: "%v". Reason: %v`,
					postgres.Name,
					err,
				)
				return err
			}
			return nil
		}
	*/

	if _, err := c.createDormantDatabase(postgres); err != nil {
		c.recorder.Eventf(
			postgres.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create DormantDatabase: "%v". Reason: %v`,
			postgres.Name,
			err,
		)
		return err
	}
	c.recorder.Eventf(
		postgres.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		`Successfully created DormantDatabase: "%v"`,
		postgres.Name,
	)

	c.cronController.StopBackupScheduling(postgres.ObjectMeta)

	if postgres.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(postgres); err != nil {
			c.recorder.Eventf(
				postgres.ObjectReference(),
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

/*
func (c *Controller) reCreatePostgres(postgres *api.Postgres) error {
	pg := &api.Postgres{
		ObjectMeta: metav1.ObjectMeta{
			Name:        postgres.Name,
			Namespace:   postgres.Namespace,
			Labels:      postgres.Labels,
			Annotations: postgres.Annotations,
		},
		Spec:   postgres.Spec,
		Status: postgres.Status,
	}

	if _, err := c.ExtClient.Postgreses(pg.Namespace).Create(pg); err != nil {
		return err
	}

	return nil
}
*/
