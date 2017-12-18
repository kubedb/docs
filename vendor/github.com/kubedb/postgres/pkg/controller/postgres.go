package controller

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/appscode/go/log"
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
	pg, err := kutildb.PatchPostgres(c.ExtClient, postgres, func(in *api.Postgres) *api.Postgres {
		t := metav1.Now()
		in.Status.CreationTime = &t
		in.Status.Phase = api.DatabasePhaseCreating
		return in
	})
	if err != nil {
		c.recorder.Eventf(postgres.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}
	*postgres = *pg

	if err := validator.ValidatePostgres(c.Client, postgres); err != nil {
		c.recorder.Event(postgres.ObjectReference(), core.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}
	// Event for successful validation
	c.recorder.Event(
		postgres.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulValidate,
		"Successfully validate Postgres",
	)

	// Check DormantDatabase
	// return True (as matched) only if Postgres matched with DormantDatabase
	// If matched, It will be resumed
	if matched, err := c.matchDormantDatabase(postgres); err != nil || matched {
		return err
	}

	// Event for notification that kubernetes objects are creating
	c.recorder.Event(postgres.ObjectReference(), core.EventTypeNormal, eventer.EventReasonCreating, "Creating Kubernetes objects")

	// create Governing Service
	governingService := c.opt.GoverningService
	if err := c.CreateGoverningService(governingService, postgres.Namespace); err != nil {
		c.recorder.Eventf(
			postgres.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create Service: "%v". Reason: %v`,
			governingService,
			err,
		)
		return err
	}

	// ensure database Service
	if err := c.ensureService(postgres); err != nil {
		return err
	}

	configMapName := fmt.Sprintf("%v-leader-lock", postgres.OffshootName())
	if err := c.deleteConfigMap(configMapName, postgres.Namespace); err != nil {
		return err
	}

	// ensure database StatefulSet
	if err := c.ensurePostgresNode(postgres); err != nil {
		return err
	}

	if postgres.Spec.Init != nil && postgres.Spec.Init.SnapshotSource != nil {
		pg, err := kutildb.PatchPostgres(c.ExtClient, postgres, func(in *api.Postgres) *api.Postgres {
			in.Status.Phase = api.DatabasePhaseInitializing
			return in
		})
		if err != nil {
			c.recorder.Eventf(postgres, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}
		*postgres = *pg

		if err := c.initialize(postgres); err != nil {
			c.recorder.Eventf(
				postgres.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToInitialize,
				"Failed to initialize. Reason: %v",
				err,
			)
		}

		pg, err = kutildb.PatchPostgres(c.ExtClient, postgres, func(in *api.Postgres) *api.Postgres {
			in.Status.Phase = api.DatabasePhaseRunning
			return in
		})
		if err != nil {
			c.recorder.Eventf(postgres, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}
		*postgres = *pg
	}

	c.recorder.Event(
		postgres.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		"Successfully created Postgres",
	)

	// Ensure Schedule backup
	c.ensureBackupScheduler(postgres)

	if postgres.Spec.Monitor != nil {
		if err := c.addMonitor(postgres); err != nil {
			c.recorder.Eventf(
				postgres.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				"Failed to add monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			postgres.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulCreate,
			"Successfully added monitoring system.",
		)
	}
	return nil
}

func (c *Controller) matchDormantDatabase(postgres *api.Postgres) (bool, error) {
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
			return false, err
		}
		return false, nil
	}

	var sendEvent = func(message string, args ...interface{}) (bool, error) {
		c.recorder.Eventf(
			postgres.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			message,
			args,
		)
		return false, fmt.Errorf(message, args)
	}

	// Check DatabaseKind
	if dormantDb.Labels[api.LabelDatabaseKind] != api.ResourceKindPostgres {
		return sendEvent(`Invalid Postgres: "%v". Exists DormantDatabase "%v" of different Kind`,
			postgres.Name, dormantDb.Name)
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

	pg, err := kutildb.PatchPostgres(c.ExtClient, postgres, func(in *api.Postgres) *api.Postgres {
		// This will ignore processing all kind of Update while creating
		if in.Annotations == nil {
			in.Annotations = make(map[string]string)
		}
		in.Annotations["kubedb.com/ignore"] = "set"
		return in
	})
	if err != nil {
		c.recorder.Eventf(postgres.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return sendEvent(err.Error())
	}
	*postgres = *pg

	if err := c.ExtClient.Postgreses(postgres.Namespace).Delete(postgres.Name, &metav1.DeleteOptions{}); err != nil {
		return sendEvent(`failed to resume Postgres "%v" from DormantDatabase "%v". Error: %v`, postgres.Name, postgres.Name, err)
	}

	_, err = kutildb.PatchDormantDatabase(c.ExtClient, dormantDb, func(in *api.DormantDatabase) *api.DormantDatabase {
		in.Spec.Resume = true
		return in
	})
	if err != nil {
		c.recorder.Eventf(postgres.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return sendEvent(err.Error())
	}

	return true, nil
}

func (c *Controller) ensurePostgresNode(postgres *api.Postgres) error {
	if err := c.ensureDatabaseSecret(postgres); err != nil {
		return err
	}

	if c.opt.EnableRbac {
		// Ensure ClusterRoles for database statefulsets
		if err := c.ensureRBACStuff(postgres); err != nil {
			return err
		}
	}

	if err := c.ensureCombinedNode(postgres); err != nil {
		return err
	}

	pg, err := kutildb.PatchPostgres(c.ExtClient, postgres, func(in *api.Postgres) *api.Postgres {
		in.Status.Phase = api.DatabasePhaseRunning
		return in
	})
	if err != nil {
		c.recorder.Eventf(postgres.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}
	*postgres = *pg

	return nil
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
	if postgres.Annotations != nil {
		if val, found := postgres.Annotations["kubedb.com/ignore"]; found {
			//TODO: Add Event Reason "Ignored"
			c.recorder.Event(postgres.ObjectReference(), core.EventTypeNormal, "Ignored", val)
			return nil
		}
	}

	c.recorder.Event(postgres.ObjectReference(), core.EventTypeNormal, eventer.EventReasonPausing, "Pausing Postgres")

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
		if err := c.deleteMonitor(postgres); err != nil {
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
		c.recorder.Event(
			postgres.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorDelete,
			"Successfully deleted monitoring system.",
		)
	}
	return nil
}

func (c *Controller) update(oldPostgres, updatedPostgres *api.Postgres) error {
	if updatedPostgres.Annotations != nil {
		if _, found := updatedPostgres.Annotations["kubedb.com/ignore"]; found {
			kutildb.PatchPostgres(c.ExtClient, updatedPostgres, func(in *api.Postgres) *api.Postgres {
				delete(in.Annotations, "kubedb.com/ignore")
				return in
			})
			return nil
		}
	}

	if err := validator.ValidatePostgres(c.Client, updatedPostgres); err != nil {
		c.recorder.Event(updatedPostgres.ObjectReference(), core.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}
	// Event for successful validation
	c.recorder.Event(
		updatedPostgres.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulValidate,
		"Successfully validate Postgres",
	)

	if err := c.ensureService(updatedPostgres); err != nil {
		return err
	}

	if err := c.ensurePostgresNode(updatedPostgres); err != nil {
		return err
	}

	if !reflect.DeepEqual(updatedPostgres.Spec.BackupSchedule, oldPostgres.Spec.BackupSchedule) {
		c.ensureBackupScheduler(updatedPostgres)
	}

	if !reflect.DeepEqual(oldPostgres.Spec.Monitor, updatedPostgres.Spec.Monitor) {
		if err := c.updateMonitor(oldPostgres, updatedPostgres); err != nil {
			c.recorder.Eventf(
				updatedPostgres.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				"Failed to update monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			updatedPostgres.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorUpdate,
			"Successfully updated monitoring system.",
		)

	}
	return nil
}

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
