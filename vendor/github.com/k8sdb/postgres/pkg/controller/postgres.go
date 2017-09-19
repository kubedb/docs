package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/appscode/go/log"
	kutildb "github.com/appscode/kutil/kubedb/v1alpha1"
	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/k8sdb/apimachinery/pkg/eventer"
	"github.com/k8sdb/apimachinery/pkg/storage"
	"github.com/k8sdb/postgres/pkg/validator"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func (c *Controller) create(postgres *tapi.Postgres) error {
	_, err := kutildb.TryPatchPostgres(c.ExtClient, postgres.ObjectMeta, func(in *tapi.Postgres) *tapi.Postgres {
		t := metav1.Now()
		in.Status.CreationTime = &t
		in.Status.Phase = tapi.DatabasePhaseCreating
		return in
	})
	if err != nil {
		c.recorder.Eventf(postgres.ObjectReference(), apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	if err := validator.ValidatePostgres(c.Client, postgres); err != nil {
		c.recorder.Event(postgres.ObjectReference(), apiv1.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}
	// Event for successful validation
	c.recorder.Event(
		postgres.ObjectReference(),
		apiv1.EventTypeNormal,
		eventer.EventReasonSuccessfulValidate,
		"Successfully validate Postgres",
	)

	// Check DormantDatabase
	matched, err := c.matchDormantDatabase(postgres)
	if err != nil {
		return err
	}
	if matched {
		//TODO: Use Annotation Key
		postgres.Annotations = map[string]string{
			"kubedb.com/ignore": "",
		}
		if err := c.ExtClient.Postgreses(postgres.Namespace).Delete(postgres.Name, &metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf(
				`Failed to resume Postgres "%v" from DormantDatabase "%v". Error: %v`,
				postgres.Name,
				postgres.Name,
				err,
			)
		}

		_, err := kutildb.TryPatchDormantDatabase(c.ExtClient, postgres.ObjectMeta, func(in *tapi.DormantDatabase) *tapi.DormantDatabase {
			in.Spec.Resume = true
			return in
		})
		if err != nil {
			c.recorder.Eventf(postgres.ObjectReference(), apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}

		return nil
	}

	// Event for notification that kubernetes objects are creating
	c.recorder.Event(postgres.ObjectReference(), apiv1.EventTypeNormal, eventer.EventReasonCreating, "Creating Kubernetes objects")

	// create Governing Service
	governingService := c.opt.GoverningService
	if err := c.CreateGoverningService(governingService, postgres.Namespace); err != nil {
		c.recorder.Eventf(
			postgres.ObjectReference(),
			apiv1.EventTypeWarning,
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

	c.recorder.Event(
		postgres.ObjectReference(),
		apiv1.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		"Successfully created Postgres",
	)

	// Ensure Schedule backup
	c.ensureBackupScheduler(postgres)

	if postgres.Spec.Monitor != nil {
		if err := c.addMonitor(postgres); err != nil {
			c.recorder.Eventf(
				postgres.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				"Failed to add monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			postgres.ObjectReference(),
			apiv1.EventTypeNormal,
			eventer.EventReasonSuccessfulCreate,
			"Successfully added monitoring system.",
		)
	}
	return nil
}

func (c *Controller) matchDormantDatabase(postgres *tapi.Postgres) (bool, error) {
	// Check if DormantDatabase exists or not
	dormantDb, err := c.ExtClient.DormantDatabases(postgres.Namespace).Get(postgres.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			c.recorder.Eventf(
				postgres.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToGet,
				`Fail to get DormantDatabase: "%v". Reason: %v`,
				postgres.Name,
				err,
			)
			return false, err
		}
		return false, nil
	}

	var sendEvent = func(message string) (bool, error) {
		c.recorder.Event(
			postgres.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			message,
		)
		return false, errors.New(message)
	}

	// Check DatabaseKind
	if dormantDb.Labels[tapi.LabelDatabaseKind] != tapi.ResourceKindPostgres {
		return sendEvent(fmt.Sprintf(`Invalid Postgres: "%v". Exists DormantDatabase "%v" of different Kind`,
			postgres.Name, dormantDb.Name))
	}

	// Check InitSpec
	initSpecAnnotationStr := dormantDb.Annotations[tapi.PostgresInitSpec]
	if initSpecAnnotationStr != "" {
		var initSpecAnnotation *tapi.InitSpec
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
		originalSpec.DatabaseSecret = &apiv1.SecretVolumeSource{
			SecretName: postgres.Name + "-admin-auth",
		}
	}

	if !reflect.DeepEqual(drmnOriginSpec, &originalSpec) {
		return sendEvent("Postgres spec mismatches with OriginSpec in DormantDatabases")
	}

	return true, nil
}

func (c *Controller) ensureService(postgres *tapi.Postgres) error {
	// Check if service name exists
	found, err := c.findService(postgres)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	// create database Service
	if err := c.createService(postgres); err != nil {
		c.recorder.Eventf(
			postgres.ObjectReference(),
			apiv1.EventTypeWarning,
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
		c.recorder.Eventf(
			postgres.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to create StatefulSet. Reason: %v",
			err,
		)
		return err
	}

	// Check StatefulSet Pod status
	if err := c.CheckStatefulSetPodStatus(statefulSet, durationCheckStatefulSet); err != nil {
		c.recorder.Eventf(
			postgres.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToStart,
			`Failed to create StatefulSet. Reason: %v`,
			err,
		)
		return err
	} else {
		c.recorder.Event(
			postgres.ObjectReference(),
			apiv1.EventTypeNormal,
			eventer.EventReasonSuccessfulCreate,
			"Successfully created StatefulSet",
		)
	}

	if postgres.Spec.Init != nil && postgres.Spec.Init.SnapshotSource != nil {
		_, err := kutildb.TryPatchPostgres(c.ExtClient, postgres.ObjectMeta, func(in *tapi.Postgres) *tapi.Postgres {
			in.Status.Phase = tapi.DatabasePhaseInitializing
			return in
		})
		if err != nil {
			c.recorder.Eventf(postgres, apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}

		if err := c.initialize(postgres); err != nil {
			c.recorder.Eventf(
				postgres.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToInitialize,
				"Failed to initialize. Reason: %v",
				err,
			)
		}
	}

	_, err = kutildb.TryPatchPostgres(c.ExtClient, postgres.ObjectMeta, func(in *tapi.Postgres) *tapi.Postgres {
		in.Status.Phase = tapi.DatabasePhaseRunning
		return in
	})
	if err != nil {
		c.recorder.Eventf(postgres, apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}
	return nil
}

func (c *Controller) ensureBackupScheduler(postgres *tapi.Postgres) {
	// Setup Schedule backup
	if postgres.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(postgres, postgres.ObjectMeta, postgres.Spec.BackupSchedule)
		if err != nil {
			c.recorder.Eventf(
				postgres.ObjectReference(),
				apiv1.EventTypeWarning,
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
	c.recorder.Eventf(
		postgres.ObjectReference(),
		apiv1.EventTypeNormal,
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
	if err != nil {
		return err
	}

	job, err := c.createRestoreJob(postgres, snapshot)
	if err != nil {
		return err
	}

	jobSuccess := c.CheckDatabaseRestoreJob(job, postgres, c.recorder, durationCheckRestoreJob)
	if jobSuccess {
		c.recorder.Event(
			postgres.ObjectReference(),
			apiv1.EventTypeNormal,
			eventer.EventReasonSuccessfulInitialize,
			"Successfully completed initialization",
		)
	} else {
		c.recorder.Event(
			postgres.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToInitialize,
			"Failed to complete initialization",
		)
	}
	return nil
}

func (c *Controller) pause(postgres *tapi.Postgres) error {
	if postgres.Annotations != nil {
		if val, found := postgres.Annotations["kubedb.com/ignore"]; found {
			//TODO: Add Event Reason "Ignored"
			c.recorder.Event(postgres.ObjectReference(), apiv1.EventTypeNormal, "Ignored", val)
			return nil
		}
	}

	c.recorder.Event(postgres.ObjectReference(), apiv1.EventTypeNormal, eventer.EventReasonPausing, "Pausing Postgres")

	if postgres.Spec.DoNotPause {
		c.recorder.Eventf(
			postgres.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToPause,
			`Postgres "%v" is locked.`,
			postgres.Name,
		)

		if err := c.reCreatePostgres(postgres); err != nil {
			c.recorder.Eventf(
				postgres.ObjectReference(),
				apiv1.EventTypeWarning,
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
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create DormantDatabase: "%v". Reason: %v`,
			postgres.Name,
			err,
		)
		return err
	}
	c.recorder.Eventf(
		postgres.ObjectReference(),
		apiv1.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		`Successfully created DormantDatabase: "%v"`,
		postgres.Name,
	)

	c.cronController.StopBackupScheduling(postgres.ObjectMeta)

	if postgres.Spec.Monitor != nil {
		if err := c.deleteMonitor(postgres); err != nil {
			c.recorder.Eventf(
				postgres.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToDelete,
				"Failed to delete monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			postgres.ObjectReference(),
			apiv1.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorDelete,
			"Successfully deleted monitoring system.",
		)
	}
	return nil
}

func (c *Controller) update(oldPostgres, updatedPostgres *tapi.Postgres) error {

	if err := validator.ValidatePostgres(c.Client, updatedPostgres); err != nil {
		c.recorder.Event(updatedPostgres.ObjectReference(), apiv1.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}
	// Event for successful validation
	c.recorder.Event(
		updatedPostgres.ObjectReference(),
		apiv1.EventTypeNormal,
		eventer.EventReasonSuccessfulValidate,
		"Successfully validate Postgres",
	)

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
			c.recorder.Eventf(
				updatedPostgres.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				"Failed to update monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			updatedPostgres.ObjectReference(),
			apiv1.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorUpdate,
			"Successfully updated monitoring system.",
		)

	}
	return nil
}
