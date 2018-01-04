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
	"github.com/kubedb/apimachinery/pkg/docker"
	"github.com/kubedb/apimachinery/pkg/eventer"
	"github.com/kubedb/apimachinery/pkg/storage"
	"github.com/kubedb/elasticsearch/pkg/validator"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) create(elasticsearch *api.Elasticsearch) error {
	if err := validator.ValidateElasticsearch(c.Client, elasticsearch, &c.opt.Docker); err != nil {
		c.recorder.Event(
			elasticsearch.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		return nil // user error so just record error and don't retry.
	}

	if elasticsearch.Status.CreationTime == nil {
		es, _, err := kutildb.PatchElasticsearch(c.ExtClient, elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
			t := metav1.Now()
			in.Status.CreationTime = &t
			in.Status.Phase = api.DatabasePhaseCreating
			return in
		})
		if err != nil {
			c.recorder.Eventf(
				elasticsearch.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return err
		}
		elasticsearch.Status = es.Status
	}

	// Dynamic Defaulting
	// Assign Default Monitoring Port
	if err := c.setMonitoringPort(elasticsearch); err != nil {
		return err
	}

	// Check DormantDatabase
	// It can be used as resumed
	if err := c.matchDormantDatabase(elasticsearch); err != nil {
		return err
	}

	// create Governing Service
	governingService := c.opt.GoverningService
	if err := c.CreateGoverningService(governingService, elasticsearch.Namespace); err != nil {
		c.recorder.Eventf(
			elasticsearch.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create ServiceAccount: "%v". Reason: %v`,
			governingService,
			err,
		)
		return err
	}

	// ensure database Service
	vt1, err := c.ensureService(elasticsearch)
	if err != nil {
		return err
	}

	// ensure database StatefulSet
	vt2, err := c.ensureElasticsearchNode(elasticsearch)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.recorder.Event(
			elasticsearch.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created Elasticsearch",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			elasticsearch.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched Elasticsearch",
		)
	}

	if vt2 == kutil.VerbCreated && elasticsearch.Spec.Init != nil && elasticsearch.Spec.Init.SnapshotSource != nil {
		es, _, err := kutildb.PatchElasticsearch(c.ExtClient, elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
			in.Status.Phase = api.DatabasePhaseInitializing
			return in
		})
		if err != nil {
			c.recorder.Eventf(elasticsearch.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}
		elasticsearch.Status = es.Status

		if err := c.initialize(elasticsearch); err != nil {
			c.recorder.Eventf(
				elasticsearch.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToInitialize,
				"Failed to initialize. Reason: %v",
				err,
			)
		}

		es, _, err = kutildb.PatchElasticsearch(c.ExtClient, elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
			in.Status.Phase = api.DatabasePhaseRunning
			return in
		})
		if err != nil {
			c.recorder.Eventf(elasticsearch.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}
		elasticsearch.Status = es.Status
	}

	// Ensure Schedule backup
	c.ensureBackupScheduler(elasticsearch)

	if err := c.manageMonitor(elasticsearch); err != nil {
		c.recorder.Eventf(
			elasticsearch.ObjectReference(),
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

func (c *Controller) setMonitoringPort(elasticsearch *api.Elasticsearch) error {
	if elasticsearch.Spec.Monitor != nil &&
		elasticsearch.Spec.Monitor.Prometheus != nil {
		if elasticsearch.Spec.Monitor.Prometheus.Port == 0 {
			es, _, err := kutildb.PatchElasticsearch(c.ExtClient, elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
				in.Spec.Monitor.Prometheus.Port = api.PrometheusExporterPortNumber
				return in
			})

			if err != nil {
				c.recorder.Eventf(
					elasticsearch.ObjectReference(),
					core.EventTypeWarning,
					eventer.EventReasonFailedToUpdate,
					err.Error(),
				)
				return err
			}
			elasticsearch.Spec.Monitor = es.Spec.Monitor
		}
	}
	return nil
}

func (c *Controller) matchDormantDatabase(elasticsearch *api.Elasticsearch) error {
	// Check if DormantDatabase exists or not
	dormantDb, err := c.ExtClient.DormantDatabases(elasticsearch.Namespace).Get(elasticsearch.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			c.recorder.Eventf(
				elasticsearch.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToGet,
				`Fail to get DormantDatabase: "%v". Reason: %v`,
				elasticsearch.Name,
				err,
			)
			return err
		}
		return nil
	}

	var sendEvent = func(message string, args ...interface{}) error {
		c.recorder.Eventf(
			elasticsearch.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			message,
			args,
		)
		return fmt.Errorf(message, args)
	}

	// Check DatabaseKind
	if dormantDb.Labels[api.LabelDatabaseKind] != api.ResourceKindElasticsearch {
		return sendEvent(fmt.Sprintf(`Invalid Elasticsearch: "%v". Exists DormantDatabase "%v" of different Kind`,
			elasticsearch.Name, dormantDb.Name))
	}

	// Check InitSpec
	initSpecAnnotationStr := dormantDb.Annotations[api.GenericInitSpec]
	if initSpecAnnotationStr != "" {
		var initSpecAnnotation *api.InitSpec
		if err := json.Unmarshal([]byte(initSpecAnnotationStr), &initSpecAnnotation); err != nil {
			return sendEvent(err.Error())
		}

		if elasticsearch.Spec.Init != nil {
			if !reflect.DeepEqual(initSpecAnnotation, elasticsearch.Spec.Init) {
				return sendEvent("InitSpec mismatches with DormantDatabase annotation")
			}
		}
	}

	// Check Origin Spec
	drmnOriginSpec := dormantDb.Spec.Origin.Spec.Elasticsearch
	originalSpec := elasticsearch.Spec
	originalSpec.Init = nil

	if originalSpec.DatabaseSecret == nil {
		originalSpec.DatabaseSecret = &core.SecretVolumeSource{
			SecretName: elasticsearch.Name + "-auth",
		}
	}

	if originalSpec.CertificateSecret == nil {
		originalSpec.CertificateSecret = &core.SecretVolumeSource{
			SecretName: elasticsearch.Name + "-cert",
		}
	}

	if !reflect.DeepEqual(drmnOriginSpec, &originalSpec) {
		return sendEvent("Elasticsearch spec mismatches with OriginSpec in DormantDatabases")
	}

	return kutildb.DeleteDormantDatabase(c.ExtClient, dormantDb.ObjectMeta)
}

func (c *Controller) ensureElasticsearchNode(elasticsearch *api.Elasticsearch) (kutil.VerbType, error) {
	var err error

	if err = c.ensureCertSecret(elasticsearch); err != nil {
		return kutil.VerbUnchanged, err
	}
	if err = c.ensureDatabaseSecret(elasticsearch); err != nil {
		return kutil.VerbUnchanged, err
	}

	vt := kutil.VerbUnchanged
	topology := elasticsearch.Spec.Topology
	if topology != nil {
		vt1, err := c.ensureClientNode(elasticsearch)
		if err != nil {
			return kutil.VerbUnchanged, err
		}
		vt2, err := c.ensureMasterNode(elasticsearch)
		if err != nil {
			return kutil.VerbUnchanged, err
		}
		vt3, err := c.ensureDataNode(elasticsearch)
		if err != nil {
			return kutil.VerbUnchanged, err
		}

		if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated && vt3 == kutil.VerbCreated {
			vt = kutil.VerbCreated
		} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched || vt3 == kutil.VerbPatched {
			vt = kutil.VerbPatched
		}
	} else {
		vt, err = c.ensureCombinedNode(elasticsearch)
		if err != nil {
			return kutil.VerbUnchanged, err
		}
	}

	// Need some time to build elasticsearch cluster. Nodes will communicate with each other
	// TODO: find better way
	time.Sleep(time.Second * 30)

	es, _, err := kutildb.PatchElasticsearch(c.ExtClient, elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
		in.Status.Phase = api.DatabasePhaseRunning
		return in
	})
	if err != nil {
		c.recorder.Eventf(elasticsearch.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return kutil.VerbUnchanged, err
	}
	elasticsearch.Status = es.Status

	return vt, nil
}

func (c *Controller) ensureBackupScheduler(elasticsearch *api.Elasticsearch) {
	kutildb.AssignTypeKind(elasticsearch)
	// Setup Schedule backup
	if elasticsearch.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(elasticsearch, elasticsearch.ObjectMeta, elasticsearch.Spec.BackupSchedule)
		if err != nil {
			c.recorder.Eventf(
				elasticsearch.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToSchedule,
				"Failed to schedule snapshot. Reason: %v",
				err,
			)
			log.Errorln(err)
		}
	} else {
		c.cronController.StopBackupScheduling(elasticsearch.ObjectMeta)
	}
}

const (
	durationCheckRestoreJob = time.Minute * 30
)

func (c *Controller) initialize(elasticsearch *api.Elasticsearch) error {
	snapshotSource := elasticsearch.Spec.Init.SnapshotSource
	// Event for notification that kubernetes objects are creating
	c.recorder.Eventf(
		elasticsearch.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonInitializing,
		`Initializing from Snapshot: "%v"`,
		snapshotSource.Name,
	)

	namespace := snapshotSource.Namespace
	if namespace == "" {
		namespace = elasticsearch.Namespace
	}
	snapshot, err := c.ExtClient.Snapshots(namespace).Get(snapshotSource.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if err := docker.CheckDockerImageVersion(c.opt.Docker.GetToolsImage(elasticsearch), string(elasticsearch.Spec.Version)); err != nil {
		return fmt.Errorf(`image %s not found`, c.opt.Docker.GetToolsImageWithTag(elasticsearch))
	}

	secret, err := storage.NewOSMSecret(c.Client, snapshot)
	if err != nil {
		return err
	}
	_, err = c.Client.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil && !kerr.IsAlreadyExists(err) {
		return err
	}

	job, err := c.createRestoreJob(elasticsearch, snapshot)
	if err != nil {
		return err
	}

	jobSuccess := c.CheckDatabaseRestoreJob(snapshot, job, elasticsearch, c.recorder, durationCheckRestoreJob)
	if jobSuccess {
		c.recorder.Event(
			elasticsearch.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulInitialize,
			"Successfully completed initialization",
		)
	} else {
		c.recorder.Event(
			elasticsearch.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToInitialize,
			"Failed to complete initialization",
		)
	}
	return nil
}

func (c *Controller) pause(elasticsearch *api.Elasticsearch) error {

	c.recorder.Event(elasticsearch.ObjectReference(), core.EventTypeNormal, eventer.EventReasonPausing, "Pausing Elasticsearch")

	/*
		if elasticsearch.Spec.DoNotPause {
			c.recorder.Eventf(
				elasticsearch.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToPause,
				`Elasticsearch "%v" is locked.`,
				elasticsearch.Name,
			)

			if err := c.reCreateElastic(elasticsearch); err != nil {
				c.recorder.Eventf(
					elasticsearch.ObjectReference(),
					core.EventTypeWarning,
					eventer.EventReasonFailedToCreate,
					`Failed to recreate Elasticsearch: "%v". Reason: %v`,
					elasticsearch.Name,
					err,
				)
				return err
			}
			return nil
		}
	*/

	if _, err := c.createDormantDatabase(elasticsearch); err != nil {
		c.recorder.Eventf(
			elasticsearch.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create DormantDatabase: "%v". Reason: %v`,
			elasticsearch.Name,
			err,
		)
		return err
	}
	c.recorder.Eventf(
		elasticsearch.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		`Successfully created DormantDatabase: "%v"`,
		elasticsearch.Name,
	)

	c.cronController.StopBackupScheduling(elasticsearch.ObjectMeta)

	if elasticsearch.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(elasticsearch); err != nil {
			c.recorder.Eventf(
				elasticsearch.ObjectReference(),
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
func (c *Controller) reCreateElastic(elasticsearch *api.Elasticsearch) error {
	es := &api.Elasticsearch{
		ObjectMeta: metav1.ObjectMeta{
			Name:        elasticsearch.Name,
			Namespace:   elasticsearch.Namespace,
			Labels:      elasticsearch.Labels,
			Annotations: elasticsearch.Annotations,
		},
		Spec:   elasticsearch.Spec,
		Status: elasticsearch.Status,
	}

	if _, err := c.ExtClient.Elasticsearchs(es.Namespace).Create(es); err != nil {
		return err
	}

	return nil
}
*/
