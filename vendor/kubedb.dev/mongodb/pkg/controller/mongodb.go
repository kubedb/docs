package controller

import (
	"fmt"

	"github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	kutil "kmodules.xyz/client-go"
	dynamic_util "kmodules.xyz/client-go/dynamic"
	meta_util "kmodules.xyz/client-go/meta"
	storage "kmodules.xyz/objectstore-api/osm"
	"kubedb.dev/apimachinery/apis"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/mongodb/pkg/admission"
)

func (c *Controller) create(mongodb *api.MongoDB) error {
	if err := validator.ValidateMongoDB(c.Client, c.ExtClient, mongodb, true); err != nil {
		c.recorder.Event(
			mongodb,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		// stop Scheduler in case there is any.
		c.cronController.StopBackupScheduling(mongodb.ObjectMeta)
		return nil
	}

	// Delete Matching DormantDatabase if exists any
	if err := c.deleteMatchingDormantDatabase(mongodb); err != nil {
		return fmt.Errorf(`failed to delete dormant Database : "%v/%v". Reason: %v`, mongodb.Namespace, mongodb.Name, err)
	}

	if mongodb.Status.Phase == "" {
		mg, err := util.UpdateMongoDBStatus(c.ExtClient.KubedbV1alpha1(), mongodb, func(in *api.MongoDBStatus) *api.MongoDBStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		}, apis.EnableStatusSubresource)
		if err != nil {
			return err
		}
		mongodb.Status = mg.Status
	}

	// create Governing Service
	if err := c.ensureMongoGvrSvc(mongodb); err != nil {
		return fmt.Errorf(`failed to create governing Service for "%v/%v". Reason: %v`, mongodb.Namespace, mongodb.Name, err)
	}

	if c.EnableRBAC {
		// Ensure Service account, role, rolebinding, and PSP for database statefulsets
		if err := c.ensureDatabaseRBAC(mongodb); err != nil {
			return err
		}
	}

	// ensure database Service
	vt1, err := c.ensureService(mongodb)
	if err != nil {
		return err
	}

	if err := c.ensureDatabaseSecret(mongodb); err != nil {
		return err
	}

	// ensure certificate or keyfile for cluster
	sslMode := mongodb.Spec.SSLMode
	if (sslMode != api.SSLModeDisabled && sslMode != "") ||
		mongodb.Spec.ReplicaSet != nil || mongodb.Spec.ShardTopology != nil {
		if err := c.ensureCertSecret(mongodb); err != nil {
			return err
		}
	}

	// ensure database StatefulSet
	vt2, err := c.ensureMongoDBNode(mongodb)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.recorder.Event(
			mongodb,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created MongoDB",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			mongodb,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched MongoDB",
		)
	}

	// ensure appbinding before ensuring Restic scheduler and restore
	_, err = c.ensureAppBinding(mongodb)
	if err != nil {
		log.Errorln(err)
		return err
	}

	if _, err := meta_util.GetString(mongodb.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		mongodb.Spec.Init != nil &&
		(mongodb.Spec.Init.SnapshotSource != nil || mongodb.Spec.Init.StashRestoreSession != nil) {

		if mongodb.Status.Phase == api.DatabasePhaseInitializing {
			return nil
		}

		// add phase that database is being initialized
		mg, err := util.UpdateMongoDBStatus(c.ExtClient.KubedbV1alpha1(), mongodb, func(in *api.MongoDBStatus) *api.MongoDBStatus {
			in.Phase = api.DatabasePhaseInitializing
			return in
		}, apis.EnableStatusSubresource)
		if err != nil {
			return err
		}
		mongodb.Status = mg.Status

		init := mongodb.Spec.Init
		if init.SnapshotSource != nil {
			err = c.initializeFromSnapshot(mongodb)
			if err != nil {
				return fmt.Errorf("failed to complete initialization. Reason: %v", err)
			}
			return err
		} else if init.StashRestoreSession != nil {
			log.Debugf("MongoDB %v/%v is waiting for restoreSession to be succeeded", mongodb.Namespace, mongodb.Name)
			return nil
		}
	}

	mg, err := util.UpdateMongoDBStatus(c.ExtClient.KubedbV1alpha1(), mongodb, func(in *api.MongoDBStatus) *api.MongoDBStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = types.NewIntHash(mongodb.Generation, meta_util.GenerationHash(mongodb))
		return in
	}, apis.EnableStatusSubresource)
	if err != nil {
		return err
	}
	mongodb.Status = mg.Status

	// Ensure Schedule backup
	if err := c.ensureBackupScheduler(mongodb); err != nil {
		c.recorder.Eventf(
			mongodb,
			core.EventTypeWarning,
			eventer.EventReasonFailedToSchedule,
			err.Error(),
		)
		log.Errorln(err)
		// Don't return error. Continue processing rest.
	}

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(mongodb); err != nil {
		c.recorder.Eventf(
			mongodb,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorf("failed to manage monitoring system. Reason: %v", err)
		return nil
	}

	if err := c.manageMonitor(mongodb); err != nil {
		c.recorder.Eventf(
			mongodb,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorf("failed to manage monitoring system. Reason: %v", err)
		return nil
	}

	return nil
}

func (c *Controller) ensureBackupScheduler(mongodb *api.MongoDB) error {
	mongodbVersion, err := c.ExtClient.CatalogV1alpha1().MongoDBVersions().Get(string(mongodb.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get MongoDBVersion %v for %v/%v. Reason: %v", mongodb.Spec.Version, mongodb.Namespace, mongodb.Name, err)
	}
	// Setup Schedule backup
	if mongodb.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(mongodb, mongodb.Spec.BackupSchedule, mongodbVersion)
		if err != nil {
			return fmt.Errorf("failed to schedule snapshot for %v/%v. Reason: %v", mongodb.Namespace, mongodb.Name, err)
		}
	} else {
		c.cronController.StopBackupScheduling(mongodb.ObjectMeta)
	}
	return nil
}

func (c *Controller) initializeFromSnapshot(mongodb *api.MongoDB) error {
	snapshotSource := mongodb.Spec.Init.SnapshotSource
	jobName := fmt.Sprintf("%s-%s", api.DatabaseNamePrefix, snapshotSource.Name)
	if _, err := c.Client.BatchV1().Jobs(snapshotSource.Namespace).Get(jobName, metav1.GetOptions{}); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	} else {
		return nil
	}

	// Event for notification that kubernetes objects are creating
	c.recorder.Eventf(
		mongodb,
		core.EventTypeNormal,
		eventer.EventReasonInitializing,
		`Initializing from Snapshot: "%v"`,
		snapshotSource.Name,
	)

	namespace := snapshotSource.Namespace
	if namespace == "" {
		namespace = mongodb.Namespace
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
	if err != nil && !kerr.IsAlreadyExists(err) {
		return err
	}

	job, err := c.createRestoreJob(mongodb, snapshot)
	if err != nil {
		return err
	}

	if err := c.SetJobOwnerReference(snapshot, job); err != nil {
		return err
	}

	return nil
}

func (c *Controller) terminate(mongodb *api.MongoDB) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb)
	if rerr != nil {
		return rerr
	}

	// If TerminationPolicy is "pause", keep everything (ie, PVCs,Secrets,Snapshots) intact.
	// In operator, create dormantdatabase
	if mongodb.Spec.TerminationPolicy == api.TerminationPolicyPause {
		if err := c.removeOwnerReferenceFromOffshoots(mongodb, ref); err != nil {
			return err
		}

		if _, err := c.createDormantDatabase(mongodb); err != nil {
			if kerr.IsAlreadyExists(err) {
				// if already exists, check if it is database of another Kind and return error in that case.
				// If the Kind is same, we can safely assume that the DormantDB was not deleted in before,
				// Probably because, User is more faster (create-delete-create-again-delete...) than operator!
				// So reuse that DormantDB!
				ddb, err := c.ExtClient.KubedbV1alpha1().DormantDatabases(mongodb.Namespace).Get(mongodb.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if val, _ := meta_util.GetStringValue(ddb.Labels, api.LabelDatabaseKind); val != api.ResourceKindMongoDB {
					return fmt.Errorf(`DormantDatabase "%v/%v" of kind %v already exists`, mongodb.Namespace, mongodb.Name, val)
				}
			} else {
				return fmt.Errorf(`failed to create DormantDatabase: "%v/%v". Reason: %v`, mongodb.Namespace, mongodb.Name, err)
			}
		}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshoots(mongodb, ref); err != nil {
			return err
		}
	}

	c.cronController.StopBackupScheduling(mongodb.ObjectMeta)

	if mongodb.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(mongodb); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}

func (c *Controller) setOwnerReferenceToOffshoots(mongodb *api.MongoDB, ref *core.ObjectReference) error {
	selector := labels.SelectorFromSet(mongodb.OffshootSelectors())

	// If TerminationPolicy is "wipeOut", delete snapshots and secrets,
	// else, keep it intact.
	if mongodb.Spec.TerminationPolicy == api.TerminationPolicyWipeOut {
		if err := dynamic_util.EnsureOwnerReferenceForSelector(
			c.DynamicClient,
			api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
			mongodb.Namespace,
			selector,
			ref); err != nil {
			return err
		}
		if err := c.wipeOutDatabase(mongodb.ObjectMeta, mongodb.Spec.GetSecrets(), ref); err != nil {
			return errors.Wrap(err, "error in wiping out database.")
		}
	} else {
		// Make sure snapshot and secret's ownerreference is removed.
		if err := dynamic_util.RemoveOwnerReferenceForSelector(
			c.DynamicClient,
			api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
			mongodb.Namespace,
			selector,
			ref); err != nil {
			return err
		}
		if err := dynamic_util.RemoveOwnerReferenceForItems(
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			mongodb.Namespace,
			mongodb.Spec.GetSecrets(),
			ref); err != nil {
			return err
		}
	}
	// delete PVC for both "wipeOut" and "delete" TerminationPolicy.
	return dynamic_util.EnsureOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		mongodb.Namespace,
		selector,
		ref)
}

func (c *Controller) removeOwnerReferenceFromOffshoots(mongodb *api.MongoDB, ref *core.ObjectReference) error {
	// First, Get LabelSelector for Other Components
	labelSelector := labels.SelectorFromSet(mongodb.OffshootSelectors())

	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		c.DynamicClient,
		api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
		mongodb.Namespace,
		labelSelector,
		ref); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		mongodb.Namespace,
		labelSelector,
		ref); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForItems(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		mongodb.Namespace,
		mongodb.Spec.GetSecrets(),
		ref); err != nil {
		return err
	}
	return nil
}
