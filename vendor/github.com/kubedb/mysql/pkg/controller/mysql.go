package controller

import (
	"fmt"

	"github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/log"
	"github.com/kubedb/apimachinery/apis"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	validator "github.com/kubedb/mysql/pkg/admission"
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
)

func (c *Controller) create(mysql *api.MySQL) error {
	if err := validator.ValidateMySQL(c.Client, c.ExtClient, mysql, true); err != nil {
		c.recorder.Event(
			mysql,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		// stop Scheduler in case there is any.
		c.cronController.StopBackupScheduling(mysql.ObjectMeta)
		return nil
	}

	// Delete Matching DormantDatabase if exists any
	if err := c.deleteMatchingDormantDatabase(mysql); err != nil {
		return fmt.Errorf(`failed to delete dormant Database : "%v/%v". Reason: %v`, mysql.Namespace, mysql.Name, err)
	}

	if mysql.Status.Phase == "" {
		my, err := util.UpdateMySQLStatus(c.ExtClient.KubedbV1alpha1(), mysql, func(in *api.MySQLStatus) *api.MySQLStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		}, apis.EnableStatusSubresource)
		if err != nil {
			return err
		}
		mysql.Status = my.Status
	}

	// create Governing Service
	governingService := c.GoverningService
	if err := c.CreateGoverningService(governingService, mysql.Namespace); err != nil {
		return fmt.Errorf(`failed to create Service: "%v/%v". Reason: %v`, mysql.Namespace, governingService, err)
	}

	if c.EnableRBAC {
		// Ensure ClusterRoles for statefulsets
		if err := c.ensureRBACStuff(mysql); err != nil {
			return err
		}
	}

	// ensure database Service
	vt1, err := c.ensureService(mysql)
	if err != nil {
		return err
	}

	if err := c.ensureDatabaseSecret(mysql); err != nil {
		return err
	}

	// ensure database StatefulSet
	vt2, err := c.ensureStatefulSet(mysql)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.recorder.Event(
			mysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created MySQL",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			mysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched MySQL",
		)
	}

	if _, err := meta_util.GetString(mysql.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		mysql.Spec.Init != nil && mysql.Spec.Init.SnapshotSource != nil {

		snapshotSource := mysql.Spec.Init.SnapshotSource

		if mysql.Status.Phase == api.DatabasePhaseInitializing {
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
		if err := c.initialize(mysql); err != nil {
			return fmt.Errorf("failed to complete initialization for %v/%v. Reason: %v", mysql.Namespace, mysql.Name, err)
		}
		return nil
	}

	my, err := util.UpdateMySQLStatus(c.ExtClient.KubedbV1alpha1(), mysql, func(in *api.MySQLStatus) *api.MySQLStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = types.NewIntHash(mysql.Generation, meta_util.GenerationHash(mysql))
		return in
	}, apis.EnableStatusSubresource)
	if err != nil {
		return err
	}
	mysql.Status = my.Status

	// Ensure Schedule backup
	if err := c.ensureBackupScheduler(mysql); err != nil {
		c.recorder.Eventf(
			mysql,
			core.EventTypeWarning,
			eventer.EventReasonFailedToSchedule,
			err.Error(),
		)
		log.Errorln(err)
		// Don't return error. Continue processing rest.
	}

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(mysql); err != nil {
		c.recorder.Eventf(
			mysql,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
		return nil
	}

	if err := c.manageMonitor(mysql); err != nil {
		c.recorder.Eventf(
			mysql,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
		return nil
	}

	_, err = c.ensureAppBinding(mysql)
	if err != nil {
		log.Errorln(err)
		return err
	}

	return nil
}

func (c *Controller) ensureBackupScheduler(mysql *api.MySQL) error {
	mysqlVersion, err := c.ExtClient.CatalogV1alpha1().MySQLVersions().Get(string(mysql.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get MySQLVersion %v for %v/%v. Reason: %v", mysql.Spec.Version, mysql.Namespace, mysql.Name, err)
	}
	// Setup Schedule backup
	if mysql.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(mysql, mysql.Spec.BackupSchedule, mysqlVersion)
		if err != nil {
			return fmt.Errorf("failed to schedule snapshot for %v/%v. Reason: %v", mysql.Namespace, mysql.Name, err)
		}
	} else {
		c.cronController.StopBackupScheduling(mysql.ObjectMeta)
	}
	return nil
}

func (c *Controller) initialize(mysql *api.MySQL) error {
	my, err := util.UpdateMySQLStatus(c.ExtClient.KubedbV1alpha1(), mysql, func(in *api.MySQLStatus) *api.MySQLStatus {
		in.Phase = api.DatabasePhaseInitializing
		return in
	}, apis.EnableStatusSubresource)
	if err != nil {
		return err
	}
	mysql.Status = my.Status

	snapshotSource := mysql.Spec.Init.SnapshotSource
	// Event for notification that kubernetes objects are creating
	c.recorder.Eventf(
		mysql,
		core.EventTypeNormal,
		eventer.EventReasonInitializing,
		`Initializing from Snapshot: "%v"`,
		snapshotSource.Name,
	)

	namespace := snapshotSource.Namespace
	if namespace == "" {
		namespace = mysql.Namespace
	}
	snapshot, err := c.ExtClient.KubedbV1alpha1().Snapshots(namespace).Get(snapshotSource.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	secret, err := storage.NewOSMSecret(c.Client, snapshot.OSMSecretName(), snapshot.Namespace, snapshot.Spec.Backend)
	if err != nil {
		return err
	}
	_, err = c.Client.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil && !kerr.IsAlreadyExists(err) {
		return err
	}

	job, err := c.createRestoreJob(mysql, snapshot)
	if err != nil {
		return err
	}

	if err := c.SetJobOwnerReference(snapshot, job); err != nil {
		return err
	}

	return nil
}

func (c *Controller) terminate(mysql *api.MySQL) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, mysql)
	if rerr != nil {
		return rerr
	}

	// If TerminationPolicy is "pause", keep everything (ie, PVCs,Secrets,Snapshots) intact.
	// In operator, create dormantdatabase
	if mysql.Spec.TerminationPolicy == api.TerminationPolicyPause {
		if err := c.removeOwnerReferenceFromOffshoots(mysql, ref); err != nil {
			return err
		}

		if _, err := c.createDormantDatabase(mysql); err != nil {
			if kerr.IsAlreadyExists(err) {
				// if already exists, check if it is database of another Kind and return error in that case.
				// If the Kind is same, we can safely assume that the DormantDB was not deleted in before,
				// Probably because, User is more faster (create-delete-create-again-delete...) than operator!
				// So reuse that DormantDB!
				ddb, err := c.ExtClient.KubedbV1alpha1().DormantDatabases(mysql.Namespace).Get(mysql.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if val, _ := meta_util.GetStringValue(ddb.Labels, api.LabelDatabaseKind); val != api.ResourceKindMySQL {
					return fmt.Errorf(`DormantDatabase "%v" of kind %v already exists`, mysql.Name, val)
				}
			} else {
				return fmt.Errorf(`failed to create DormantDatabase: "%v/%v". Reason: %v`, mysql.Namespace, mysql.Name, err)
			}
		}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshoots(mysql, ref); err != nil {
			return err
		}
	}

	c.cronController.StopBackupScheduling(mysql.ObjectMeta)

	if mysql.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(mysql); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}

func (c *Controller) setOwnerReferenceToOffshoots(mysql *api.MySQL, ref *core.ObjectReference) error {
	selector := labels.SelectorFromSet(mysql.OffshootSelectors())

	// If TerminationPolicy is "wipeOut", delete snapshots and secrets,
	// else, keep it intact.
	if mysql.Spec.TerminationPolicy == api.TerminationPolicyWipeOut {
		if err := dynamic_util.EnsureOwnerReferenceForSelector(
			c.DynamicClient,
			api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
			mysql.Namespace,
			selector,
			ref); err != nil {
			return err
		}
		if err := c.wipeOutDatabase(mysql.ObjectMeta, mysql.Spec.GetSecrets(), ref); err != nil {
			return errors.Wrap(err, "error in wiping out database.")
		}
	} else {
		// Make sure snapshot and secret's ownerreference is removed.
		if err := dynamic_util.RemoveOwnerReferenceForSelector(
			c.DynamicClient,
			api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
			mysql.Namespace,
			selector,
			ref); err != nil {
			return err
		}
		if err := dynamic_util.RemoveOwnerReferenceForItems(
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			mysql.Namespace,
			mysql.Spec.GetSecrets(),
			ref); err != nil {
			return err
		}
	}
	// delete PVC for both "wipeOut" and "delete" TerminationPolicy.
	return dynamic_util.EnsureOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		mysql.Namespace,
		selector,
		ref)
}

func (c *Controller) removeOwnerReferenceFromOffshoots(mysql *api.MySQL, ref *core.ObjectReference) error {
	// First, Get LabelSelector for Other Components
	labelSelector := labels.SelectorFromSet(mysql.OffshootSelectors())

	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		c.DynamicClient,
		api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
		mysql.Namespace,
		labelSelector,
		ref); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		mysql.Namespace,
		labelSelector,
		ref); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForItems(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		mysql.Namespace,
		mysql.Spec.GetSecrets(),
		ref); err != nil {
		return err
	}
	return nil
}
