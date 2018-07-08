package controller

import (
	"fmt"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/kutil"
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	kutildb "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	"github.com/kubedb/apimachinery/pkg/storage"
	validator "github.com/kubedb/elasticsearch/pkg/admission"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
)

func (c *Controller) create(elasticsearch *api.Elasticsearch) error {
	if err := validator.ValidateElasticsearch(c.Client, c.ExtClient, elasticsearch); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, elasticsearch); rerr == nil {
			c.recorder.Event(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonInvalid,
				err.Error(),
			)
		}
		log.Errorln(err)
		return nil // user error so just record error and don't retry.
	}

	// Delete Matching DormantDatabase if exists any
	if err := c.deleteMatchingDormantDatabase(elasticsearch); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, elasticsearch); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to delete dormant Database : "%v". Reason: %v`,
				elasticsearch.Name,
				err,
			)
		}
		return err
	}

	if elasticsearch.Status.CreationTime == nil {
		es, err := kutildb.UpdateElasticsearchStatus(c.ExtClient, elasticsearch, func(in *api.ElasticsearchStatus) *api.ElasticsearchStatus {
			t := metav1.Now()
			in.CreationTime = &t
			in.Phase = api.DatabasePhaseCreating
			return in
		}, api.EnableStatusSubresource)
		if err != nil {
			if ref, rerr := reference.GetReference(clientsetscheme.Scheme, elasticsearch); rerr == nil {
				c.recorder.Eventf(
					ref,
					core.EventTypeWarning,
					eventer.EventReasonFailedToUpdate,
					err.Error(),
				)
			}
			return err
		}
		elasticsearch.Status = es.Status
	}

	// create Governing Service
	governingService := c.GoverningService
	if err := c.CreateGoverningService(governingService, elasticsearch.Namespace); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, elasticsearch); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to create ServiceAccount: "%v". Reason: %v`,
				governingService,
				err,
			)
		}
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
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, elasticsearch); rerr == nil {
			c.recorder.Event(
				ref,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully created Elasticsearch",
			)
		}
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, elasticsearch); rerr == nil {
			c.recorder.Event(
				ref,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully patched Elasticsearch",
			)
		}
	}

	if _, err := meta_util.GetString(elasticsearch.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		elasticsearch.Spec.Init != nil &&
		elasticsearch.Spec.Init.SnapshotSource != nil {

		snapshotSource := elasticsearch.Spec.Init.SnapshotSource

		if elasticsearch.Status.Phase == api.DatabasePhaseInitializing {
			return nil
		}

		jobName := fmt.Sprintf("%s-%s", api.DatabaseNamePrefix, snapshotSource.Name)
		if _, err := c.Client.BatchV1().Jobs(snapshotSource.Namespace).Get(jobName, metav1.GetOptions{}); err != nil {
			if !kerr.IsNotFound(err) {
				return err
			}
		} else {
			return nil
		}
		err = c.initialize(elasticsearch)
		if err != nil {
			return fmt.Errorf("failed to complete initialization. Reason: %v", err)
		}
		return nil
	}

	es, err := kutildb.UpdateElasticsearchStatus(c.ExtClient, elasticsearch, func(in *api.ElasticsearchStatus) *api.ElasticsearchStatus {
		in.Phase = api.DatabasePhaseRunning
		return in
	}, api.EnableStatusSubresource)
	if err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, elasticsearch); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
		}
		return err
	}
	elasticsearch.Status = es.Status

	// Ensure Schedule backup
	c.ensureBackupScheduler(elasticsearch)

	if err := c.manageMonitor(elasticsearch); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, elasticsearch); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				"Failed to manage monitoring system. Reason: %v",
				err,
			)
		}
		log.Errorln(err)
		return nil
	}
	return nil
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

	return vt, nil
}

func (c *Controller) ensureBackupScheduler(elasticsearch *api.Elasticsearch) {
	kutildb.AssignTypeKind(elasticsearch)
	// Setup Schedule backup
	if elasticsearch.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(elasticsearch, elasticsearch.ObjectMeta, elasticsearch.Spec.BackupSchedule)
		if err != nil {
			if ref, rerr := reference.GetReference(clientsetscheme.Scheme, elasticsearch); rerr == nil {
				c.recorder.Eventf(
					ref,
					core.EventTypeWarning,
					eventer.EventReasonFailedToSchedule,
					"Failed to schedule snapshot. Reason: %v",
					err,
				)
			}
			log.Errorln(err)
		}
	} else {
		c.cronController.StopBackupScheduling(elasticsearch.ObjectMeta)
	}
}

func (c *Controller) initialize(elasticsearch *api.Elasticsearch) error {
	es, err := kutildb.UpdateElasticsearchStatus(c.ExtClient, elasticsearch, func(in *api.ElasticsearchStatus) *api.ElasticsearchStatus {
		in.Phase = api.DatabasePhaseInitializing
		return in
	}, api.EnableStatusSubresource)
	if err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, elasticsearch); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
		}
		return err
	}
	elasticsearch.Status = es.Status

	snapshotSource := elasticsearch.Spec.Init.SnapshotSource
	// Event for notification that kubernetes objects are creating
	if ref, rerr := reference.GetReference(clientsetscheme.Scheme, elasticsearch); rerr == nil {
		c.recorder.Eventf(
			ref,
			core.EventTypeNormal,
			eventer.EventReasonInitializing,
			`Initializing from Snapshot: "%v"`,
			snapshotSource.Name,
		)
	}

	namespace := snapshotSource.Namespace
	if namespace == "" {
		namespace = elasticsearch.Namespace
	}
	snapshot, err := c.ExtClient.Snapshots(namespace).Get(snapshotSource.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	secret, err := storage.NewOSMSecret(c.Client, snapshot)
	if err != nil {
		return err
	}
	secret, err = c.Client.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil {
		return err
	}

	job, err := c.createRestoreJob(elasticsearch, snapshot)
	if err != nil {
		return err
	}

	if err := c.SetJobOwnerReference(snapshot, job); err != nil {
		return err
	}
	return nil
}

func (c *Controller) pause(elasticsearch *api.Elasticsearch) error {

	if _, err := c.createDormantDatabase(elasticsearch); err != nil {
		if kerr.IsAlreadyExists(err) {
			// if already exists, check if it is database of another Kind and return error in that case.
			// If the Kind is same, we can safely assume that the DormantDB was not deleted in before,
			// Probably because, User is more faster (create-delete-create-again-delete...) than operator!
			// So reuse that DormantDB!
			ddb, err := c.ExtClient.DormantDatabases(elasticsearch.Namespace).Get(elasticsearch.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if val, _ := meta_util.GetStringValue(ddb.Labels, api.LabelDatabaseKind); val != api.ResourceKindElasticsearch {
				return fmt.Errorf(`DormantDatabase "%v" of kind %v already exists`, elasticsearch.Name, val)
			}
		} else {
			return fmt.Errorf(`failed to create DormantDatabase: "%v". Reason: %v`, elasticsearch.Name, err)
		}
	}

	c.cronController.StopBackupScheduling(elasticsearch.ObjectMeta)

	if elasticsearch.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(elasticsearch); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}

func (c *Controller) GetDatabase(meta metav1.ObjectMeta) (runtime.Object, error) {
	elasticsearch, err := c.ExtClient.Elasticsearches(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return elasticsearch, nil
}

func (c *Controller) SetDatabaseStatus(meta metav1.ObjectMeta, phase api.DatabasePhase, reason string) error {
	elasticsearch, err := c.ExtClient.Elasticsearches(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, err = kutildb.UpdateElasticsearchStatus(c.ExtClient, elasticsearch, func(in *api.ElasticsearchStatus) *api.ElasticsearchStatus {
		in.Phase = phase
		in.Reason = reason
		return in
	}, api.EnableStatusSubresource)
	return err
}

func (c *Controller) UpsertDatabaseAnnotation(meta metav1.ObjectMeta, annotation map[string]string) error {
	elasticsearch, err := c.ExtClient.Elasticsearches(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	_, _, err = kutildb.PatchElasticsearch(c.ExtClient, elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
		in.Annotations = core_util.UpsertMap(elasticsearch.Annotations, annotation)
		return in
	})
	return err
}
