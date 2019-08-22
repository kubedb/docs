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
	"k8s.io/apimachinery/pkg/runtime"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	dynamic_util "kmodules.xyz/client-go/dynamic"
	meta_util "kmodules.xyz/client-go/meta"
	storage "kmodules.xyz/objectstore-api/osm"
	"kubedb.dev/apimachinery/apis"
	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/postgres/pkg/admission"
)

func (c *Controller) create(postgres *api.Postgres) error {
	if err := validator.ValidatePostgres(c.Client, c.ExtClient, postgres, true); err != nil {
		c.recorder.Event(
			postgres,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		// stop Scheduler in case there is any.
		c.cronController.StopBackupScheduling(postgres.ObjectMeta)
		return nil // user error so just record error and don't retry.
	}

	// Delete Matching DormantDatabase if exists any
	if err := c.deleteMatchingDormantDatabase(postgres); err != nil {
		return fmt.Errorf(`failed to delete dormant Database : "%v/%v". Reason: %v`, postgres.Namespace, postgres.Name, err)
	}

	if postgres.Status.Phase == "" {
		pg, err := util.UpdatePostgresStatus(c.ExtClient.KubedbV1alpha1(), postgres, func(in *api.PostgresStatus) *api.PostgresStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		}, apis.EnableStatusSubresource)
		if err != nil {
			return err
		}
		postgres.Status = pg.Status
	}

	// create Governing Service
	governingService := c.GoverningService
	if err := c.CreateGoverningService(governingService, postgres.Namespace); err != nil {
		return fmt.Errorf(`failed to create Service: "%v/%v". Reason: %v`, postgres.Namespace, governingService, err)
	}

	// ensure database Service
	vt1, err := c.ensureService(postgres)
	if err != nil {
		return err
	}

	// ensure database StatefulSet
	postgresVersion, err := c.ExtClient.CatalogV1alpha1().PostgresVersions().Get(string(postgres.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return err
	}
	vt2, err := c.ensurePostgresNode(postgres, postgresVersion)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.recorder.Event(
			postgres,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created Postgres",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			postgres,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched Postgres",
		)
	}

	// ensure appbinding before ensuring Restic scheduler and restore
	_, err = c.ensureAppBinding(postgres, postgresVersion)
	if err != nil {
		log.Errorln(err)
		return err
	}

	if _, err := meta_util.GetString(postgres.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		postgres.Spec.Init != nil &&
		(postgres.Spec.Init.SnapshotSource != nil || postgres.Spec.Init.StashRestoreSession != nil) {

		if postgres.Status.Phase == api.DatabasePhaseInitializing {
			return nil
		}

		// add phase that database is being initialized
		pg, err := util.UpdatePostgresStatus(c.ExtClient.KubedbV1alpha1(), postgres, func(in *api.PostgresStatus) *api.PostgresStatus {
			in.Phase = api.DatabasePhaseInitializing
			return in
		}, apis.EnableStatusSubresource)
		if err != nil {
			return err
		}
		postgres.Status = pg.Status

		init := postgres.Spec.Init
		if init.SnapshotSource != nil {
			err = c.initializeFromSnapshot(postgres)
			if err != nil {
				return fmt.Errorf("failed to complete initialization. Reason: %v", err)
			}
			return err
		} else if init.StashRestoreSession != nil {
			log.Debugf("Postgres %v/%v is waiting for restoreSession to be succeeded", postgres.Namespace, postgres.Name)
			return nil
		}
	}

	pg, err := util.UpdatePostgresStatus(c.ExtClient.KubedbV1alpha1(), postgres, func(in *api.PostgresStatus) *api.PostgresStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = types.NewIntHash(postgres.Generation, meta_util.GenerationHash(postgres))
		return in
	}, apis.EnableStatusSubresource)
	if err != nil {
		return err
	}
	postgres.Status = pg.Status

	// Ensure Schedule backup
	if err := c.ensureBackupScheduler(postgres); err != nil {
		c.recorder.Eventf(
			postgres,
			core.EventTypeWarning,
			eventer.EventReasonFailedToSchedule,
			err.Error(),
		)
		log.Errorln(err)
		// Don't return error. Continue processing rest.
	}

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(postgres); err != nil {
		c.recorder.Eventf(
			postgres,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
		return nil
	}

	if err := c.manageMonitor(postgres); err != nil {
		c.recorder.Eventf(
			postgres,
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

func (c *Controller) ensurePostgresNode(postgres *api.Postgres, postgresVersion *catalog.PostgresVersion) (kutil.VerbType, error) {
	var err error

	if err = c.ensureDatabaseSecret(postgres); err != nil {
		return kutil.VerbUnchanged, err
	}

	if c.EnableRBAC {
		// Ensure Service account, role, rolebinding, and PSP for database statefulsets
		if err := c.ensureDatabaseRBAC(postgres); err != nil {
			return kutil.VerbUnchanged, err
		}
	}

	vt, err := c.ensureCombinedNode(postgres, postgresVersion)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	return vt, nil
}

func (c *Controller) ensureBackupScheduler(postgres *api.Postgres) error {
	postgresVersion, err := c.ExtClient.CatalogV1alpha1().PostgresVersions().Get(string(postgres.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get PostgresVersion %v for %v/%v. Reason: %v", postgres.Spec.Version, postgres.Namespace, postgres.Name, err)
	}
	// Setup Schedule backup
	if postgres.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(postgres, postgres.Spec.BackupSchedule, postgresVersion)
		if err != nil {
			return fmt.Errorf("failed to schedule snapshot. Reason: %v", err)
		}
	} else {
		c.cronController.StopBackupScheduling(postgres.ObjectMeta)
	}
	return nil
}

func (c *Controller) initializeFromSnapshot(postgres *api.Postgres) error {
	snapshotSource := postgres.Spec.Init.SnapshotSource
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
		postgres,
		core.EventTypeNormal,
		eventer.EventReasonInitializing,
		`Initializing from Snapshot: "%v"`,
		snapshotSource.Name,
	)

	namespace := snapshotSource.Namespace
	if namespace == "" {
		namespace = postgres.Namespace
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

	job, err := c.createRestoreJob(postgres, snapshot)
	if err != nil {
		return err
	}

	if err := c.SetJobOwnerReference(snapshot, job); err != nil {
		return err
	}
	return nil
}

func (c *Controller) terminate(postgres *api.Postgres) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, postgres)
	if rerr != nil {
		return rerr
	}

	// If TerminationPolicy is "pause", keep everything (ie, PVCs,Secrets,Snapshots) intact.
	// In operator, create dormantdatabase
	if postgres.Spec.TerminationPolicy == api.TerminationPolicyPause {
		if err := c.removeOwnerReferenceFromOffshoots(postgres, ref); err != nil {
			return err
		}

		if _, err := c.createDormantDatabase(postgres); err != nil {
			if kerr.IsAlreadyExists(err) {
				// if already exists, check if it is database of another Kind and return error in that case.
				// If the Kind is same, we can safely assume that the DormantDB was not deleted in before,
				// Probably because, User is more faster (create-delete-create-again-delete...) than operator!
				// So reuse that DormantDB!
				ddb, err := c.ExtClient.KubedbV1alpha1().DormantDatabases(postgres.Namespace).Get(postgres.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if val, _ := meta_util.GetStringValue(ddb.Labels, api.LabelDatabaseKind); val != api.ResourceKindPostgres {
					return fmt.Errorf(`DormantDatabase "%s/%s" of kind %v already exists`, postgres.Namespace, postgres.Name, val)
				}
			} else {
				return fmt.Errorf(`failed to create DormantDatabase: "%s/%s". Reason: %v`, postgres.Namespace, postgres.Name, err)
			}
		}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots,WAL-data).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets, wal-data intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshoots(postgres, ref); err != nil {
			return err
		}
	}

	c.cronController.StopBackupScheduling(postgres.ObjectMeta)

	if postgres.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(postgres); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}

func (c *Controller) setOwnerReferenceToOffshoots(postgres *api.Postgres, ref *core.ObjectReference) error {
	selector := labels.SelectorFromSet(postgres.OffshootSelectors())

	// If TerminationPolicy is "wipeOut", delete snapshots and secrets,
	// else, keep it intact.
	if postgres.Spec.TerminationPolicy == api.TerminationPolicyWipeOut {
		if err := dynamic_util.EnsureOwnerReferenceForSelector(
			c.DynamicClient,
			api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
			postgres.Namespace,
			selector,
			ref); err != nil {
			return err
		}
		if err := c.wipeOutDatabase(postgres.ObjectMeta, postgres.Spec.GetSecrets(), ref); err != nil {
			return errors.Wrap(err, "error in wiping out database.")
		}
		// if wal archiver was configured, remove wal data from backend
		if postgres.Spec.Archiver != nil {
			return c.wipeOutWalData(postgres.ObjectMeta, &postgres.Spec)
		}
	} else {
		// Make sure snapshot and secret's ownerreference is removed.
		if err := dynamic_util.RemoveOwnerReferenceForSelector(
			c.DynamicClient,
			api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
			postgres.Namespace,
			selector,
			ref); err != nil {
			return err
		}
		if err := dynamic_util.RemoveOwnerReferenceForItems(
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			postgres.Namespace,
			postgres.Spec.GetSecrets(),
			ref); err != nil {
			return err
		}
	}
	// delete PVC for both "wipeOut" and "delete" TerminationPolicy.
	return dynamic_util.EnsureOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		postgres.Namespace,
		selector,
		ref)
}

func (c *Controller) removeOwnerReferenceFromOffshoots(postgres *api.Postgres, ref *core.ObjectReference) error {
	// First, Get LabelSelector for Other Components
	labelSelector := labels.SelectorFromSet(postgres.OffshootSelectors())

	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		c.DynamicClient,
		api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
		postgres.Namespace,
		labelSelector,
		ref); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		postgres.Namespace,
		labelSelector,
		ref); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForItems(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		postgres.Namespace,
		postgres.Spec.GetSecrets(),
		ref); err != nil {
		return err
	}
	return nil
}

func (c *Controller) GetDatabase(meta metav1.ObjectMeta) (runtime.Object, error) {
	postgres, err := c.ExtClient.KubedbV1alpha1().Postgreses(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return postgres, nil
}

func (c *Controller) SetDatabaseStatus(meta metav1.ObjectMeta, phase api.DatabasePhase, reason string) error {
	postgres, err := c.ExtClient.KubedbV1alpha1().Postgreses(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, err = util.UpdatePostgresStatus(c.ExtClient.KubedbV1alpha1(), postgres, func(in *api.PostgresStatus) *api.PostgresStatus {
		in.Phase = phase
		in.Reason = reason
		return in
	}, apis.EnableStatusSubresource)
	return err
}

func (c *Controller) UpsertDatabaseAnnotation(meta metav1.ObjectMeta, annotation map[string]string) error {
	postgres, err := c.ExtClient.KubedbV1alpha1().Postgreses(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	_, _, err = util.PatchPostgres(c.ExtClient.KubedbV1alpha1(), postgres, func(in *api.Postgres) *api.Postgres {
		in.Annotations = core_util.UpsertMap(in.Annotations, annotation)
		return in
	})
	return err
}
