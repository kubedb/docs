package controller

import (
	"fmt"
	"time"

	"github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/log"
	"github.com/appscode/kutil"
	core_util "github.com/appscode/kutil/core/v1"
	dynamic_util "github.com/appscode/kutil/dynamic"
	meta_util "github.com/appscode/kutil/meta"
	"github.com/kubedb/apimachinery/apis"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	validator "github.com/kubedb/elasticsearch/pkg/admission"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	storage "kmodules.xyz/objectstore-api/osm"
)

func (c *Controller) create(elasticsearch *api.Elasticsearch) error {
	if err := validator.ValidateElasticsearch(c.Client, c.ExtClient, elasticsearch); err != nil {
		c.recorder.Event(
			elasticsearch,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		return nil // user error so just record error and don't retry.
	}

	// Check if elasticsearchVersion is deprecated.
	// If deprecated, add event and return nil (stop processing.)
	elasticsearchVersion, err := c.ExtClient.CatalogV1alpha1().ElasticsearchVersions().Get(string(elasticsearch.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return err
	}
	if elasticsearchVersion.Spec.Deprecated {
		c.recorder.Eventf(
			elasticsearch,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			"ElasticsearchVersion %v is deprecated. Skipped processing.",
			elasticsearchVersion.Name,
		)
		log.Errorf("Elasticsearch %s/%s is using deprecated version %v. Skipped processing.",
			elasticsearch.Namespace, elasticsearch.Name, elasticsearchVersion.Name)
		return nil
	}

	// Delete Matching DormantDatabase if exists any
	if err := c.deleteMatchingDormantDatabase(elasticsearch); err != nil {
		c.recorder.Eventf(
			elasticsearch,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to delete dormant Database : "%v". Reason: %v`,
			elasticsearch.Name,
			err,
		)
		return err
	}

	if elasticsearch.Status.Phase == "" {
		es, err := util.UpdateElasticsearchStatus(c.ExtClient.KubedbV1alpha1(), elasticsearch, func(in *api.ElasticsearchStatus) *api.ElasticsearchStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		}, apis.EnableStatusSubresource)
		if err != nil {
			c.recorder.Eventf(
				elasticsearch,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return err
		}
		elasticsearch.Status = es.Status
	}

	// create Governing Service
	governingService := c.GoverningService
	if err := c.CreateGoverningService(governingService, elasticsearch.Namespace); err != nil {
		c.recorder.Eventf(
			elasticsearch,
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
			elasticsearch,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created Elasticsearch",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			elasticsearch,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched Elasticsearch",
		)

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

	es, err := util.UpdateElasticsearchStatus(c.ExtClient.KubedbV1alpha1(), elasticsearch, func(in *api.ElasticsearchStatus) *api.ElasticsearchStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = types.NewIntHash(elasticsearch.Generation, meta_util.GenerationHash(elasticsearch))
		return in
	}, apis.EnableStatusSubresource)
	if err != nil {
		c.recorder.Eventf(
			elasticsearch,
			core.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			err.Error(),
		)

		return err
	}
	elasticsearch.Status = es.Status

	// Ensure Schedule backup
	c.ensureBackupScheduler(elasticsearch)

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(elasticsearch); err != nil {
		c.recorder.Eventf(
			elasticsearch,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)

		log.Errorln(err)
		return nil
	}

	if err := c.manageMonitor(elasticsearch); err != nil {
		c.recorder.Eventf(
			elasticsearch,
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
	// Setup Schedule backup
	if elasticsearch.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(elasticsearch, elasticsearch.ObjectMeta, elasticsearch.Spec.BackupSchedule)
		if err != nil {
			c.recorder.Eventf(
				elasticsearch,
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

func (c *Controller) initialize(elasticsearch *api.Elasticsearch) error {
	es, err := util.UpdateElasticsearchStatus(c.ExtClient.KubedbV1alpha1(), elasticsearch, func(in *api.ElasticsearchStatus) *api.ElasticsearchStatus {
		in.Phase = api.DatabasePhaseInitializing
		return in
	}, apis.EnableStatusSubresource)
	if err != nil {
		c.recorder.Eventf(
			elasticsearch,
			core.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			err.Error(),
		)

		return err
	}
	elasticsearch.Status = es.Status

	snapshotSource := elasticsearch.Spec.Init.SnapshotSource
	// Event for notification that kubernetes objects are creating
	c.recorder.Eventf(
		elasticsearch,
		core.EventTypeNormal,
		eventer.EventReasonInitializing,
		`Initializing from Snapshot: "%v"`,
		snapshotSource.Name,
	)

	namespace := snapshotSource.Namespace
	if namespace == "" {
		namespace = elasticsearch.Namespace
	}
	snapshot, err := c.ExtClient.KubedbV1alpha1().Snapshots(namespace).Get(snapshotSource.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	secret, err := storage.NewOSMSecret(c.Client, snapshot.OSMSecretName(), snapshot.Namespace, snapshot.Spec.Backend)
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

func (c *Controller) terminate(elasticsearch *api.Elasticsearch) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, elasticsearch)
	if rerr != nil {
		return rerr
	}

	// If TerminationPolicy is "pause", keep everything (ie, PVCs,Secrets,Snapshots) intact.
	// In operator, create dormantdatabase
	if elasticsearch.Spec.TerminationPolicy == api.TerminationPolicyPause {
		if err := c.removeOwnerReferenceFromOffshoots(elasticsearch, ref); err != nil {
			return err
		}

		if _, err := c.createDormantDatabase(elasticsearch); err != nil {
			if kerr.IsAlreadyExists(err) {
				// if already exists, check if it is database of another Kind and return error in that case.
				// If the Kind is same, we can safely assume that the DormantDB was not deleted in before,
				// Probably because, User is more faster (create-delete-create-again-delete...) than operator!
				// So reuse that DormantDB!
				ddb, err := c.ExtClient.KubedbV1alpha1().DormantDatabases(elasticsearch.Namespace).Get(elasticsearch.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if val, _ := meta_util.GetStringValue(ddb.Labels, api.LabelDatabaseKind); val != api.ResourceKindElasticsearch {
					return fmt.Errorf(`DormantDatabase "%s/%s" of kind %v already exists`, elasticsearch.Namespace, elasticsearch.Name, val)
				}
			} else {
				return fmt.Errorf(`failed to create DormantDatabase: "%s/%s". Reason: %v`, elasticsearch.Namespace, elasticsearch.Name, err)
			}
		}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshoots(elasticsearch, ref); err != nil {
			return err
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

func (c *Controller) setOwnerReferenceToOffshoots(elasticsearch *api.Elasticsearch, ref *core.ObjectReference) error {
	selector := labels.SelectorFromSet(elasticsearch.OffshootSelectors())

	// If TerminationPolicy is "wipeOut", delete snapshots and secrets,
	// else, keep it intact.
	if elasticsearch.Spec.TerminationPolicy == api.TerminationPolicyWipeOut {
		if err := dynamic_util.EnsureOwnerReferenceForSelector(
			c.DynamicClient,
			api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
			elasticsearch.Namespace,
			selector,
			ref); err != nil {
			return err
		}
		if err := c.wipeOutDatabase(elasticsearch.ObjectMeta, elasticsearch.Spec.GetSecrets(), ref); err != nil {
			return errors.Wrap(err, "error in wiping out database.")
		}
	} else {
		// Make sure snapshot and secret's ownerreference is removed.
		if err := dynamic_util.RemoveOwnerReferenceForSelector(
			c.DynamicClient,
			api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
			elasticsearch.Namespace,
			selector,
			ref); err != nil {
			return err
		}
		if err := dynamic_util.RemoveOwnerReferenceForItems(
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			elasticsearch.Namespace,
			elasticsearch.Spec.GetSecrets(),
			ref); err != nil {
			return err
		}
	}
	// delete PVC for both "wipeOut" and "delete" TerminationPolicy.
	return dynamic_util.EnsureOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		elasticsearch.Namespace,
		selector,
		ref)
}

func (c *Controller) removeOwnerReferenceFromOffshoots(elasticsearch *api.Elasticsearch, ref *core.ObjectReference) error {
	// First, Get LabelSelector for Other Components
	labelSelector := labels.SelectorFromSet(elasticsearch.OffshootSelectors())

	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		c.DynamicClient,
		api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
		elasticsearch.Namespace,
		labelSelector,
		ref); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		elasticsearch.Namespace,
		labelSelector,
		ref); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForItems(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		elasticsearch.Namespace,
		elasticsearch.Spec.GetSecrets(),
		ref); err != nil {
		return err
	}
	return nil
}

func (c *Controller) GetDatabase(meta metav1.ObjectMeta) (runtime.Object, error) {
	elasticsearch, err := c.ExtClient.KubedbV1alpha1().Elasticsearches(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return elasticsearch, nil
}

func (c *Controller) SetDatabaseStatus(meta metav1.ObjectMeta, phase api.DatabasePhase, reason string) error {
	elasticsearch, err := c.ExtClient.KubedbV1alpha1().Elasticsearches(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, err = util.UpdateElasticsearchStatus(c.ExtClient.KubedbV1alpha1(), elasticsearch, func(in *api.ElasticsearchStatus) *api.ElasticsearchStatus {
		in.Phase = phase
		in.Reason = reason
		return in
	}, apis.EnableStatusSubresource)
	return err
}

func (c *Controller) UpsertDatabaseAnnotation(meta metav1.ObjectMeta, annotation map[string]string) error {
	elasticsearch, err := c.ExtClient.KubedbV1alpha1().Elasticsearches(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	_, _, err = util.PatchElasticsearch(c.ExtClient.KubedbV1alpha1(), elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
		in.Annotations = core_util.UpsertMap(elasticsearch.Annotations, annotation)
		return in
	})
	return err
}
