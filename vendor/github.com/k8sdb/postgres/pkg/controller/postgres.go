package controller

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/appscode/log"
	tapi "github.com/k8sdb/apimachinery/api"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
	"github.com/k8sdb/apimachinery/pkg/eventer"
	kapi "k8s.io/kubernetes/pkg/api"
	k8serr "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

func (c *Controller) create(postgres *tapi.Postgres) error {
	var err error
	if postgres, err = c.ExtClient.Postgreses(postgres.Namespace).Get(postgres.Name); err != nil {
		return err
	}

	t := unversioned.Now()
	postgres.Status.CreationTime = &t
	postgres.Status.Phase = tapi.DatabasePhaseCreating
	if _, err = c.ExtClient.Postgreses(postgres.Namespace).Update(postgres); err != nil {
		c.eventRecorder.Eventf(
			postgres,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			`Fail to update Postgres: "%v". Reason: %v`,
			postgres.Name,
			err,
		)
		log.Errorln(err)
	}

	if err := c.validatePostgres(postgres); err != nil {
		c.eventRecorder.Event(postgres, kapi.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}
	// Event for successful validation
	c.eventRecorder.Event(
		postgres,
		kapi.EventTypeNormal,
		eventer.EventReasonSuccessfulValidate,
		"Successfully validate Postgres",
	)

	// Check DormantDatabase
	if err := c.findDormantDatabase(postgres); err != nil {
		return err
	}

	// Event for notification that kubernetes objects are creating
	c.eventRecorder.Event(postgres, kapi.EventTypeNormal, eventer.EventReasonCreating, "Creating Kubernetes objects")

	// create Governing Service
	governingService := c.opt.GoverningService
	if err := c.CreateGoverningService(governingService, postgres.Namespace); err != nil {
		c.eventRecorder.Eventf(
			postgres,
			kapi.EventTypeWarning,
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

	// ensure database StatefulSet
	if err := c.ensureStatefulSet(postgres); err != nil {
		return err
	}

	c.eventRecorder.Event(
		postgres,
		kapi.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		"Successfully created Postgres",
	)

	// Ensure Schedule backup
	c.ensureBackupScheduler(postgres)

	if postgres.Spec.Monitor != nil {
		if err := c.addMonitor(postgres); err != nil {
			c.eventRecorder.Eventf(
				postgres,
				kapi.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				"Failed to add monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.eventRecorder.Event(
			postgres,
			kapi.EventTypeNormal,
			eventer.EventReasonSuccessfulCreate,
			"Successfully added monitoring system.",
		)
	}
	return nil
}

func (c *Controller) findDormantDatabase(postgres *tapi.Postgres) error {
	// Check if DormantDatabase exists or not
	dormantDb, err := c.ExtClient.DormantDatabases(postgres.Namespace).Get(postgres.Name)
	if err != nil {
		if !k8serr.IsNotFound(err) {
			c.eventRecorder.Eventf(
				postgres,
				kapi.EventTypeWarning,
				eventer.EventReasonFailedToGet,
				`Fail to get DormantDatabase: "%v". Reason: %v`,
				postgres.Name,
				err,
			)
			return err
		}
	} else {
		var message string
		if dormantDb.Labels[amc.LabelDatabaseKind] != tapi.ResourceKindPostgres {
			message = fmt.Sprintf(`Invalid Postgres: "%v". Exists DormantDatabase "%v" of different Kind`,
				postgres.Name, dormantDb.Name)
		} else {
			message = fmt.Sprintf(`Resume from DormantDatabase: "%v"`, dormantDb.Name)
		}
		c.eventRecorder.Event(
			postgres,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			message,
		)
		return errors.New(message)
	}
	return nil
}

func (c *Controller) ensureService(postgres *tapi.Postgres) error {
	// Check if service name exists
	found, err := c.findService(postgres.Name, postgres.Namespace)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	// create database Service
	if err := c.createService(postgres.Name, postgres.Namespace); err != nil {
		c.eventRecorder.Eventf(
			postgres,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to create Service. Reason: %v",
			err,
		)
		return err
	}
	return nil
}

func (c *Controller) ensureStatefulSet(postgres *tapi.Postgres) error {
	found, err := c.findStatefulSet(postgres)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	// Create statefulSet for Postgres database
	statefulSet, err := c.createStatefulSet(postgres)
	if err != nil {
		c.eventRecorder.Eventf(
			postgres,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to create StatefulSet. Reason: %v",
			err,
		)
		return err
	}

	// Check StatefulSet Pod status
	if err := c.CheckStatefulSetPodStatus(statefulSet, durationCheckStatefulSet); err != nil {
		c.eventRecorder.Eventf(
			postgres,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToStart,
			`Failed to create StatefulSet. Reason: %v`,
			err,
		)
		return err
	} else {
		c.eventRecorder.Event(
			postgres,
			kapi.EventTypeNormal,
			eventer.EventReasonSuccessfulCreate,
			"Successfully created StatefulSet",
		)
	}

	if postgres.Spec.Init != nil && postgres.Spec.Init.SnapshotSource != nil {
		if postgres, err = c.ExtClient.Postgreses(postgres.Namespace).Get(postgres.Name); err != nil {
			return err
		}

		postgres.Status.Phase = tapi.DatabasePhaseInitializing
		if _, err = c.ExtClient.Postgreses(postgres.Namespace).Update(postgres); err != nil {
			c.eventRecorder.Eventf(
				postgres,
				kapi.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				`Fail to update Postgres: "%v". Reason: %v`,
				postgres.Name,
				err,
			)
			log.Errorln(err)
		}

		if err := c.initialize(postgres); err != nil {
			c.eventRecorder.Eventf(
				postgres,
				kapi.EventTypeWarning,
				eventer.EventReasonFailedToInitialize,
				"Failed to initialize. Reason: %v",
				err,
			)
		}
	}

	if postgres, err = c.ExtClient.Postgreses(postgres.Namespace).Get(postgres.Name); err != nil {
		return err
	}

	postgres.Status.Phase = tapi.DatabasePhaseRunning
	if _, err = c.ExtClient.Postgreses(postgres.Namespace).Update(postgres); err != nil {
		c.eventRecorder.Eventf(
			postgres,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			`Failed to update Postgres: "%v". Reason: %v`,
			postgres.Name,
			err,
		)
		log.Errorln(err)
	}

	return nil
}

func (c *Controller) ensureBackupScheduler(postgres *tapi.Postgres) {
	// Setup Schedule backup
	if postgres.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(postgres, postgres.ObjectMeta, postgres.Spec.BackupSchedule)
		if err != nil {
			c.eventRecorder.Eventf(
				postgres,
				kapi.EventTypeWarning,
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

func (c *Controller) initialize(postgres *tapi.Postgres) error {
	snapshotSource := postgres.Spec.Init.SnapshotSource
	// Event for notification that kubernetes objects are creating
	c.eventRecorder.Eventf(
		postgres,
		kapi.EventTypeNormal,
		eventer.EventReasonInitializing,
		`Initializing from Snapshot: "%v"`,
		snapshotSource.Name,
	)

	namespace := snapshotSource.Namespace
	if namespace == "" {
		namespace = postgres.Namespace
	}
	snapshot, err := c.ExtClient.Snapshots(namespace).Get(snapshotSource.Name)
	if err != nil {
		return err
	}

	job, err := c.createRestoreJob(postgres, snapshot)
	if err != nil {
		return err
	}

	jobSuccess := c.CheckDatabaseRestoreJob(job, postgres, c.eventRecorder, durationCheckRestoreJob)
	if jobSuccess {
		c.eventRecorder.Event(
			postgres,
			kapi.EventTypeNormal,
			eventer.EventReasonSuccessfulInitialize,
			"Successfully completed initialization",
		)
	} else {
		c.eventRecorder.Event(
			postgres,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToInitialize,
			"Failed to complete initialization",
		)
	}
	return nil
}

func (c *Controller) pause(postgres *tapi.Postgres) error {
	c.eventRecorder.Event(postgres, kapi.EventTypeNormal, eventer.EventReasonPausing, "Pausing Postgres")

	if postgres.Spec.DoNotPause {
		c.eventRecorder.Eventf(
			postgres,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToPause,
			`Postgres "%v" is locked.`,
			postgres.Name,
		)

		if err := c.reCreatePostgres(postgres); err != nil {
			c.eventRecorder.Eventf(
				postgres,
				kapi.EventTypeWarning,
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
		c.eventRecorder.Eventf(
			postgres,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create DormantDatabase: "%v". Reason: %v`,
			postgres.Name,
			err,
		)
		return err
	}
	c.eventRecorder.Eventf(
		postgres,
		kapi.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		`Successfully created DormantDatabase: "%v"`,
		postgres.Name,
	)

	c.cronController.StopBackupScheduling(postgres.ObjectMeta)

	if postgres.Spec.Monitor != nil {
		if err := c.deleteMonitor(postgres); err != nil {
			c.eventRecorder.Eventf(
				postgres,
				kapi.EventTypeWarning,
				eventer.EventReasonFailedToDelete,
				"Failed to delete monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.eventRecorder.Event(
			postgres,
			kapi.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorDelete,
			"Successfully deleted monitoring system.",
		)
	}
	return nil
}

func (c *Controller) update(oldPostgres, updatedPostgres *tapi.Postgres) error {

	if err := c.validatePostgres(updatedPostgres); err != nil {
		c.eventRecorder.Event(updatedPostgres, kapi.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}
	// Event for successful validation
	c.eventRecorder.Event(
		updatedPostgres,
		kapi.EventTypeNormal,
		eventer.EventReasonSuccessfulValidate,
		"Successfully validate Postgres",
	)

	// Check DormantDatabase
	if err := c.findDormantDatabase(updatedPostgres); err != nil {
		return err
	}

	if err := c.ensureService(updatedPostgres); err != nil {
		return err
	}
	if err := c.ensureStatefulSet(updatedPostgres); err != nil {
		return err
	}

	if !reflect.DeepEqual(updatedPostgres.Spec.BackupSchedule, oldPostgres.Spec.BackupSchedule) {
		c.ensureBackupScheduler(updatedPostgres)
	}

	if !reflect.DeepEqual(oldPostgres.Spec.Monitor, updatedPostgres.Spec.Monitor) {
		if err := c.updateMonitor(oldPostgres, updatedPostgres); err != nil {
			c.eventRecorder.Eventf(
				updatedPostgres,
				kapi.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				"Failed to update monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.eventRecorder.Event(
			updatedPostgres,
			kapi.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorUpdate,
			"Successfully updated monitoring system.",
		)

	}
	return nil
}
