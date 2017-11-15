package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/appscode/go/log"
	api "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	kutildb "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"github.com/k8sdb/apimachinery/pkg/docker"
	"github.com/k8sdb/apimachinery/pkg/eventer"
	"github.com/k8sdb/apimachinery/pkg/storage"
	"github.com/k8sdb/elasticsearch/pkg/validator"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) create(elastic *api.Elasticsearch) error {
	_, err := kutildb.TryPatchElasticsearch(c.ExtClient, elastic.ObjectMeta, func(in *api.Elasticsearch) *api.Elasticsearch {
		t := metav1.Now()
		in.Status.CreationTime = &t
		in.Status.Phase = api.DatabasePhaseCreating
		return in
	})
	if err != nil {
		c.recorder.Eventf(elastic.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	if err := validator.ValidateElastic(c.Client, elastic); err != nil {
		c.recorder.Event(elastic.ObjectReference(), core.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}
	// Validate DiscoveryTag
	if err := docker.CheckDockerImageVersion(docker.ImageElasticOperator, c.opt.DiscoveryTag); err != nil {
		return fmt.Errorf(`Image %v:%v not found`, docker.ImageElasticOperator, c.opt.DiscoveryTag)
	}

	// Event for successful validation
	c.recorder.Event(
		elastic.ObjectReference(),
		core.EventTypeNormal,
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

		_, err := kutildb.TryPatchDormantDatabase(c.ExtClient, elastic.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
			in.Spec.Resume = true
			return in
		})
		if err != nil {
			c.recorder.Eventf(elastic.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}
		return nil
	}

	// Event for notification that kubernetes objects are creating
	c.recorder.Event(elastic.ObjectReference(), core.EventTypeNormal, eventer.EventReasonCreating, "Creating Kubernetes objects")

	// create Governing Service
	governingService := c.opt.GoverningService
	if err := c.CreateGoverningService(governingService, elastic.Namespace); err != nil {
		c.recorder.Eventf(
			elastic.ObjectReference(),
			core.EventTypeWarning,
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
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		"Successfully created Elasticsearch",
	)

	// Ensure Schedule backup
	c.ensureBackupScheduler(elastic)

	if elastic.Spec.Monitor != nil {
		if err := c.addMonitor(elastic); err != nil {
			c.recorder.Eventf(
				elastic.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToAddMonitor,
				"Failed to add monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			elastic.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorAdd,
			"Successfully added monitoring system.",
		)
	}
	return nil
}

func (c *Controller) matchDormantDatabase(elastic *api.Elasticsearch) (bool, error) {
	// Check if DormantDatabase exists or not
	dormantDb, err := c.ExtClient.DormantDatabases(elastic.Namespace).Get(elastic.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			c.recorder.Eventf(
				elastic.ObjectReference(),
				core.EventTypeWarning,
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
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			message,
		)
		return false, errors.New(message)
	}

	// Check DatabaseKind
	if dormantDb.Labels[api.LabelDatabaseKind] != api.ResourceKindElasticsearch {
		return sendEvent(fmt.Sprintf(`Invalid Elasticsearch: "%v". Exists DormantDatabase "%v" of different Kind`,
			elastic.Name, dormantDb.Name))
	}

	// Check InitSpec
	initSpecAnnotationStr := dormantDb.Annotations[api.ElasticsearchInitSpec]
	if initSpecAnnotationStr != "" {
		var initSpecAnnotation *api.InitSpec
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

func (c *Controller) ensureService(elastic *api.Elasticsearch) error {
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
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to create Service. Reason: %v",
			err,
		)
		return err
	}
	return nil
}

func (c *Controller) ensureStatefulSet(elastic *api.Elasticsearch) error {
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
			core.EventTypeWarning,
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
				core.EventTypeWarning,
				eventer.EventReasonFailedToStart,
				"Failed to create StatefulSet. Reason: %v",
				err,
			)
			return err
		} else {
			c.recorder.Event(
				elastic.ObjectReference(),
				core.EventTypeNormal,
				eventer.EventReasonSuccessfulCreate,
				"Successfully created StatefulSet",
			)
		}
	}

	if elastic.Spec.Init != nil && elastic.Spec.Init.SnapshotSource != nil {
		_, err := kutildb.TryPatchElasticsearch(c.ExtClient, elastic.ObjectMeta, func(in *api.Elasticsearch) *api.Elasticsearch {
			in.Status.Phase = api.DatabasePhaseInitializing
			return in
		})
		if err != nil {
			c.recorder.Eventf(elastic.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}

		if err := c.initialize(elastic); err != nil {
			c.recorder.Eventf(
				elastic.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToInitialize,
				"Failed to initialize. Reason: %v",
				err,
			)
		}
	}

	_, err = kutildb.TryPatchElasticsearch(c.ExtClient, elastic.ObjectMeta, func(in *api.Elasticsearch) *api.Elasticsearch {
		in.Status.Phase = api.DatabasePhaseRunning
		return in
	})
	if err != nil {
		c.recorder.Eventf(elastic.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	return nil
}

func (c *Controller) ensureBackupScheduler(elastic *api.Elasticsearch) {
	// Setup Schedule backup
	if elastic.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(elastic, elastic.ObjectMeta, elastic.Spec.BackupSchedule)
		if err != nil {
			c.recorder.Eventf(
				elastic.ObjectReference(),
				core.EventTypeWarning,
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

func (c *Controller) initialize(elastic *api.Elasticsearch) error {
	snapshotSource := elastic.Spec.Init.SnapshotSource
	// Event for notification that kubernetes objects are creating
	c.recorder.Eventf(
		elastic.ObjectReference(),
		core.EventTypeNormal,
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
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulInitialize,
			"Successfully completed initialization",
		)
	} else {
		c.recorder.Event(
			elastic.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToInitialize,
			"Failed to complete initialization",
		)
	}
	return nil
}

func (c *Controller) pause(elastic *api.Elasticsearch) error {
	if elastic.Annotations != nil {
		if val, found := elastic.Annotations["kubedb.com/ignore"]; found {
			//TODO: Add Event Reason "Ignored"
			c.recorder.Event(elastic.ObjectReference(), core.EventTypeNormal, "Ignored", val)
			return nil
		}
	}

	c.recorder.Event(elastic.ObjectReference(), core.EventTypeNormal, eventer.EventReasonPausing, "Pausing Elasticsearch")

	if elastic.Spec.DoNotPause {
		c.recorder.Eventf(
			elastic.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToPause,
			`Elasticsearch "%v" is locked.`,
			elastic.Name,
		)

		if err := c.reCreateElastic(elastic); err != nil {
			c.recorder.Eventf(
				elastic.ObjectReference(),
				core.EventTypeWarning,
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
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create DormantDatabase: "%v". Reason: %v`,
			elastic.Name,
			err,
		)
		return err
	}
	c.recorder.Eventf(
		elastic.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		`Successfully created DormantDatabase: "%v"`,
		elastic.Name,
	)

	c.cronController.StopBackupScheduling(elastic.ObjectMeta)

	if elastic.Spec.Monitor != nil {
		if err := c.deleteMonitor(elastic); err != nil {
			c.recorder.Eventf(
				elastic.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToDeleteMonitor,
				"Failed to delete monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			elastic.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorDelete,
			"Successfully deleted monitoring system.",
		)
	}
	return nil
}

func (c *Controller) update(oldElastic, updatedElastic *api.Elasticsearch) error {

	if err := validator.ValidateElastic(c.Client, updatedElastic); err != nil {
		c.recorder.Event(updatedElastic, core.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}
	// Event for successful validation
	c.recorder.Event(
		updatedElastic.ObjectReference(),
		core.EventTypeNormal,
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
				core.EventTypeNormal,
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
				core.EventTypeNormal,
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
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdateMonitor,
				"Failed to update monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			updatedElastic.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorUpdate,
			"Successfully updated monitoring system.",
		)
	}

	return nil
}
