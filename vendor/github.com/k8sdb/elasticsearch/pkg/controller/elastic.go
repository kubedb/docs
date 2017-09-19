package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	kutildb "github.com/appscode/kutil/kubedb/v1alpha1"
	"github.com/appscode/log"
	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/k8sdb/apimachinery/pkg/docker"
	"github.com/k8sdb/apimachinery/pkg/eventer"
	"github.com/k8sdb/apimachinery/pkg/storage"
	"github.com/k8sdb/elasticsearch/pkg/validator"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func (c *Controller) create(elastic *tapi.Elasticsearch) error {
	_, err := c.UpdateElasticsearch(elastic.ObjectMeta, func(in tapi.Elasticsearch) tapi.Elasticsearch {
		t := metav1.Now()
		in.Status.CreationTime = &t
		in.Status.Phase = tapi.DatabasePhaseCreating
		return in
	})
	if err != nil {
		c.recorder.Eventf(elastic.ObjectReference(), apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	if err := validator.ValidateElastic(c.Client, elastic); err != nil {
		c.recorder.Event(elastic.ObjectReference(), apiv1.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}
	// Validate DiscoveryTag
	if err := docker.CheckDockerImageVersion(docker.ImageElasticOperator, c.opt.DiscoveryTag); err != nil {
		return fmt.Errorf(`Image %v:%v not found`, docker.ImageElasticOperator, c.opt.DiscoveryTag)
	}

	// Event for successful validation
	c.recorder.Event(
		elastic.ObjectReference(),
		apiv1.EventTypeNormal,
		eventer.EventReasonSuccessfulValidate,
		"Successfully validate Elasticsearch",
	)

	// Check DormantDatabase
	matched, err := c.matchDormantDatabase(elastic)
	if err != nil {
		return err
	}
	if matched {
		//TODO: Use Annotation Key
		elastic.Annotations = map[string]string{
			"kubedb.com/ignore": "",
		}
		if err := c.ExtClient.Elasticsearchs(elastic.Namespace).Delete(elastic.Name, &metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf(
				`failed to resume Elasticsearch "%v" from DormantDatabase "%v". Error: %v`,
				elastic.Name,
				elastic.Name,
				err,
			)
		}

		_, err := kutildb.TryPatchDormantDatabase(c.ExtClient, elastic.ObjectMeta, func(in *tapi.DormantDatabase) *tapi.DormantDatabase {
			in.Spec.Resume = true
			return in
		})
		if err != nil {
			c.recorder.Eventf(elastic.ObjectReference(), apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}
		return nil
	}

	// Event for notification that kubernetes objects are creating
	c.recorder.Event(elastic.ObjectReference(), apiv1.EventTypeNormal, eventer.EventReasonCreating, "Creating Kubernetes objects")

	// create Governing Service
	governingService := c.opt.GoverningService
	if err := c.CreateGoverningService(governingService, elastic.Namespace); err != nil {
		c.recorder.Eventf(
			elastic.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create ServiceAccount: "%v". Reason: %v`,
			governingService,
			err,
		)
		return err
	}

	// ensure database Service
	if err := c.ensureService(elastic); err != nil {
		return err
	}

	// ensure database StatefulSet
	if err := c.ensureStatefulSet(elastic); err != nil {
		return err
	}

	c.recorder.Event(
		elastic.ObjectReference(),
		apiv1.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		"Successfully created Elasticsearch",
	)

	// Ensure Schedule backup
	c.ensureBackupScheduler(elastic)

	if elastic.Spec.Monitor != nil {
		if err := c.addMonitor(elastic); err != nil {
			c.recorder.Eventf(
				elastic.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToAddMonitor,
				"Failed to add monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			elastic.ObjectReference(),
			apiv1.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorAdd,
			"Successfully added monitoring system.",
		)
	}
	return nil
}

func (c *Controller) matchDormantDatabase(elastic *tapi.Elasticsearch) (bool, error) {
	// Check if DormantDatabase exists or not
	dormantDb, err := c.ExtClient.DormantDatabases(elastic.Namespace).Get(elastic.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			c.recorder.Eventf(
				elastic.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToGet,
				`Fail to get DormantDatabase: "%v". Reason: %v`,
				elastic.Name,
				err,
			)
			return false, err
		}
		return false, nil
	}

	var sendEvent = func(message string) (bool, error) {
		c.recorder.Event(
			elastic.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			message,
		)
		return false, errors.New(message)
	}

	// Check DatabaseKind
	if dormantDb.Labels[tapi.LabelDatabaseKind] != tapi.ResourceKindElasticsearch {
		return sendEvent(fmt.Sprintf(`Invalid Elasticsearch: "%v". Exists DormantDatabase "%v" of different Kind`,
			elastic.Name, dormantDb.Name))
	}

	// Check InitSpec
	initSpecAnnotationStr := dormantDb.Annotations[tapi.ElasticsearchInitSpec]
	if initSpecAnnotationStr != "" {
		var initSpecAnnotation *tapi.InitSpec
		if err := json.Unmarshal([]byte(initSpecAnnotationStr), &initSpecAnnotation); err != nil {
			return sendEvent(err.Error())
		}

		if elastic.Spec.Init != nil {
			if !reflect.DeepEqual(initSpecAnnotation, elastic.Spec.Init) {
				return sendEvent("InitSpec mismatches with DormantDatabase annotation")
			}
		}
	}

	// Check Origin Spec
	drmnOriginSpec := dormantDb.Spec.Origin.Spec.Elasticsearch
	originalSpec := elastic.Spec
	originalSpec.Init = nil

	if !reflect.DeepEqual(drmnOriginSpec, &originalSpec) {
		return sendEvent("Elasticsearch spec mismatches with OriginSpec in DormantDatabases")
	}

	return true, nil
}

func (c *Controller) ensureService(elastic *tapi.Elasticsearch) error {
	// Check if service name exists
	found, err := c.findService(elastic)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	// create database Service
	if err := c.createService(elastic); err != nil {
		c.recorder.Eventf(
			elastic.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to create Service. Reason: %v",
			err,
		)
		return err
	}
	return nil
}

func (c *Controller) ensureStatefulSet(elastic *tapi.Elasticsearch) error {
	found, err := c.findStatefulSet(elastic)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	// Create statefulSet for Elasticsearch database
	statefulSet, err := c.createStatefulSet(elastic)
	if err != nil {
		c.recorder.Eventf(
			elastic.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to create StatefulSet. Reason: %v",
			err,
		)
		return err
	}

	// Check StatefulSet Pod status
	if elastic.Spec.Replicas > 0 {
		if err := c.CheckStatefulSetPodStatus(statefulSet, durationCheckStatefulSet); err != nil {
			c.recorder.Eventf(
				elastic.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToStart,
				"Failed to create StatefulSet. Reason: %v",
				err,
			)
			return err
		} else {
			c.recorder.Event(
				elastic.ObjectReference(),
				apiv1.EventTypeNormal,
				eventer.EventReasonSuccessfulCreate,
				"Successfully created StatefulSet",
			)
		}
	}

	if elastic.Spec.Init != nil && elastic.Spec.Init.SnapshotSource != nil {
		_, err := c.UpdateElasticsearch(elastic.ObjectMeta, func(in tapi.Elasticsearch) tapi.Elasticsearch {
			in.Status.Phase = tapi.DatabasePhaseInitializing
			return in
		})
		if err != nil {
			c.recorder.Eventf(elastic.ObjectReference(), apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}

		if err := c.initialize(elastic); err != nil {
			c.recorder.Eventf(
				elastic.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToInitialize,
				"Failed to initialize. Reason: %v",
				err,
			)
		}
	}

	_, err = c.UpdateElasticsearch(elastic.ObjectMeta, func(in tapi.Elasticsearch) tapi.Elasticsearch {
		in.Status.Phase = tapi.DatabasePhaseRunning
		return in
	})
	if err != nil {
		c.recorder.Eventf(elastic.ObjectReference(), apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	return nil
}

func (c *Controller) ensureBackupScheduler(elastic *tapi.Elasticsearch) {
	// Setup Schedule backup
	if elastic.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(elastic, elastic.ObjectMeta, elastic.Spec.BackupSchedule)
		if err != nil {
			c.recorder.Eventf(
				elastic.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToSchedule,
				"Failed to schedule snapshot. Reason: %v",
				err,
			)
			log.Errorln(err)
		}
	} else {
		c.cronController.StopBackupScheduling(elastic.ObjectMeta)
	}
}

const (
	durationCheckRestoreJob = time.Minute * 30
)

func (c *Controller) initialize(elastic *tapi.Elasticsearch) error {
	snapshotSource := elastic.Spec.Init.SnapshotSource
	// Event for notification that kubernetes objects are creating
	c.recorder.Eventf(
		elastic.ObjectReference(),
		apiv1.EventTypeNormal,
		eventer.EventReasonInitializing,
		`Initializing from Snapshot: "%v"`,
		snapshotSource.Name,
	)

	namespace := snapshotSource.Namespace
	if namespace == "" {
		namespace = elastic.Namespace
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

	job, err := c.createRestoreJob(elastic, snapshot)
	if err != nil {
		return err
	}

	jobSuccess := c.CheckDatabaseRestoreJob(job, elastic, c.recorder, durationCheckRestoreJob)
	if jobSuccess {
		c.recorder.Event(
			elastic.ObjectReference(),
			apiv1.EventTypeNormal,
			eventer.EventReasonSuccessfulInitialize,
			"Successfully completed initialization",
		)
	} else {
		c.recorder.Event(
			elastic.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToInitialize,
			"Failed to complete initialization",
		)
	}
	return nil
}

func (c *Controller) pause(elastic *tapi.Elasticsearch) error {
	if elastic.Annotations != nil {
		if val, found := elastic.Annotations["kubedb.com/ignore"]; found {
			//TODO: Add Event Reason "Ignored"
			c.recorder.Event(elastic.ObjectReference(), apiv1.EventTypeNormal, "Ignored", val)
			return nil
		}
	}

	c.recorder.Event(elastic.ObjectReference(), apiv1.EventTypeNormal, eventer.EventReasonPausing, "Pausing Elasticsearch")

	if elastic.Spec.DoNotPause {
		c.recorder.Eventf(
			elastic.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToPause,
			`Elasticsearch "%v" is locked.`,
			elastic.Name,
		)

		if err := c.reCreateElastic(elastic); err != nil {
			c.recorder.Eventf(
				elastic.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to recreate Elasticsearch: "%v". Reason: %v`,
				elastic.Name,
				err,
			)
			return err
		}
		return nil
	}

	if _, err := c.createDormantDatabase(elastic); err != nil {
		c.recorder.Eventf(
			elastic.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create DormantDatabase: "%v". Reason: %v`,
			elastic.Name,
			err,
		)
		return err
	}
	c.recorder.Eventf(
		elastic.ObjectReference(),
		apiv1.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		`Successfully created DormantDatabase: "%v"`,
		elastic.Name,
	)

	c.cronController.StopBackupScheduling(elastic.ObjectMeta)

	if elastic.Spec.Monitor != nil {
		if err := c.deleteMonitor(elastic); err != nil {
			c.recorder.Eventf(
				elastic.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToDeleteMonitor,
				"Failed to delete monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			elastic.ObjectReference(),
			apiv1.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorDelete,
			"Successfully deleted monitoring system.",
		)
	}
	return nil
}

func (c *Controller) update(oldElastic, updatedElastic *tapi.Elasticsearch) error {

	if err := validator.ValidateElastic(c.Client, updatedElastic); err != nil {
		c.recorder.Event(updatedElastic, apiv1.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}
	// Event for successful validation
	c.recorder.Event(
		updatedElastic.ObjectReference(),
		apiv1.EventTypeNormal,
		eventer.EventReasonSuccessfulValidate,
		"Successfully validate Elasticsearch",
	)

	if err := c.ensureService(updatedElastic); err != nil {
		return err
	}
	if err := c.ensureStatefulSet(updatedElastic); err != nil {
		return err
	}

	if (updatedElastic.Spec.Replicas != oldElastic.Spec.Replicas) && updatedElastic.Spec.Replicas >= 0 {
		statefulSetName := updatedElastic.OffshootName()
		statefulSet, err := c.Client.AppsV1beta1().StatefulSets(updatedElastic.Namespace).Get(statefulSetName, metav1.GetOptions{})
		if err != nil {
			c.recorder.Eventf(
				updatedElastic.ObjectReference(),
				apiv1.EventTypeNormal,
				eventer.EventReasonFailedToGet,
				`Failed to get StatefulSet: "%v". Reason: %v`,
				statefulSetName,
				err,
			)
			return err
		}
		statefulSet.Spec.Replicas = &updatedElastic.Spec.Replicas
		if _, err := c.Client.AppsV1beta1().StatefulSets(statefulSet.Namespace).Update(statefulSet); err != nil {
			c.recorder.Eventf(
				updatedElastic.ObjectReference(),
				apiv1.EventTypeNormal,
				eventer.EventReasonFailedToUpdate,
				`Failed to update StatefulSet: "%v". Reason: %v`,
				statefulSetName,
				err,
			)
			log.Errorln(err)
		}
	}

	if !reflect.DeepEqual(updatedElastic.Spec.BackupSchedule, oldElastic.Spec.BackupSchedule) {
		c.ensureBackupScheduler(updatedElastic)
	}

	if !reflect.DeepEqual(oldElastic.Spec.Monitor, updatedElastic.Spec.Monitor) {
		if err := c.updateMonitor(oldElastic, updatedElastic); err != nil {
			c.recorder.Eventf(
				updatedElastic.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToUpdateMonitor,
				"Failed to update monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			updatedElastic.ObjectReference(),
			apiv1.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorUpdate,
			"Successfully updated monitoring system.",
		)
	}

	return nil
}
