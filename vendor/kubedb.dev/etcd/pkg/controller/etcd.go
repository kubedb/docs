package controller

import (
	"fmt"

	. "github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kwatch "k8s.io/apimachinery/pkg/watch"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	dynamic_util "kmodules.xyz/client-go/dynamic"
	meta_util "kmodules.xyz/client-go/meta"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/etcd/pkg/admission"
)

func (c *Controller) syncEtcd(etcd *api.Etcd) error {
	ev := &Event{
		Type:   kwatch.Added,
		Object: etcd,
	}

	if _, ok := c.clusters[etcd.Name]; ok {
		ev.Type = kwatch.Modified
	}

	return c.handleEtcdEvent(ev)

}

func (c *Controller) handleEtcdEvent(event *Event) error {
	etcd := event.Object
	if etcd.Status.Phase == api.DatabasePhaseFailed {
		if event.Type == kwatch.Deleted {
			delete(c.clusters, etcd.Name)
			return nil
		}
		return fmt.Errorf("ignore failed cluster (%s). Please delete its CR", etcd.Name)
	}

	if err := validator.ValidateEtcd(c.Client, c.ExtClient, etcd, true); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd); rerr == nil {
			c.recorder.Event(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonInvalid,
				err.Error(),
			)
		}
		// stop Scheduler in case there is any.
		c.cronController.StopBackupScheduling(etcd.ObjectMeta)
		log.Errorln(err)
		return nil
	}

	// Delete Matching DormantDatabase if exists any
	if err := c.deleteMatchingDormantDatabase(etcd); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to delete dormant Database : "%v". Reason: %v`,
				etcd.Name,
				err,
			)
		}
		return err
	}

	switch event.Type {
	case kwatch.Added:
		if _, ok := c.clusters[etcd.Name]; ok {
			return fmt.Errorf("unsafe state. cluster (%s) was created before but we received event (%s)", etcd.Name, event.Type)
		}
		c.NewCluster(etcd)
		//c.clusters[etcd.Name] = cluster
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd); rerr == nil {
			c.recorder.Event(
				ref,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully created Etcd",
			)
		}
	case kwatch.Modified:
		if _, ok := c.clusters[etcd.Name]; !ok {
			return fmt.Errorf("unsafe state. cluster (%s) was never created but we received event (%s)", etcd.Name, event.Type)
		}
		c.clusters[etcd.Name].Update(etcd)
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd); rerr == nil {
			c.recorder.Event(
				ref,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully patched Etcd",
			)
		}

	case kwatch.Deleted:
		if _, ok := c.clusters[etcd.Name]; !ok {
			return fmt.Errorf("unsafe state. cluster (%s) was never created but we received event (%s)", etcd.Name, event.Type)
		}
		c.clusters[etcd.Name].Delete()
		delete(c.clusters, etcd.Name)
	}

	if err := core_util.WaitUntilPodRunningBySelector(
		c.Client,
		etcd.Namespace,
		&metav1.LabelSelector{
			MatchLabels: etcd.OffshootLabels(),
		},
		int(types.Int32(etcd.Spec.Replicas)),
	); err != nil {
		return err
	}

	//return nil
	if _, err := meta_util.GetString(etcd.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		etcd.Spec.Init != nil && etcd.Spec.Init.SnapshotSource != nil {
		//snapshotSource := etcd.Spec.Init.SnapshotSource

		etcd.Annotations = core_util.UpsertMap(etcd.Annotations, map[string]string{
			api.AnnotationInitialized: "",
		})
	}

	db, err := util.UpdateEtcdStatus(c.ExtClient.KubedbV1alpha1(), etcd, func(in *api.EtcdStatus) *api.EtcdStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = NewIntHash(etcd.Generation, meta_util.GenerationHash(etcd))
		return in
	})

	if err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
		}
		return err
	}
	etcd.Status = db.Status

	// Ensure Schedule backup
	c.ensureBackupScheduler(etcd)

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(etcd); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd); rerr == nil {
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

	if err := c.manageMonitor(etcd); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd); rerr == nil {
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

func (c *Controller) ensureBackupScheduler(etcd *api.Etcd) {
	etcdVersion, err := c.ExtClient.CatalogV1alpha1().EtcdVersions().Get(string(etcd.Spec.Version), metav1.GetOptions{})
	if err != nil {
		c.recorder.Eventf(
			etcd,
			core.EventTypeWarning,
			eventer.EventReasonFailedToSchedule,
			"Failed to get EtcdVersion for %v. Reason: %v",
			etcd.Spec.Version, err,
		)
		log.Errorln(err)
		return
	}
	// Setup Schedule backup
	if etcd.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(etcd, etcd.Spec.BackupSchedule, etcdVersion)
		if err != nil {
			log.Errorln(err)
			if ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd); rerr == nil {
				c.recorder.Eventf(
					ref,
					core.EventTypeWarning,
					eventer.EventReasonFailedToSchedule,
					"Failed to schedule snapshot. Reason: %v",
					err,
				)
			}
		}
	} else {
		c.cronController.StopBackupScheduling(etcd.ObjectMeta)
	}
}

func (c *Controller) initialize(etcd *api.Etcd) error {
	db, err := util.UpdateEtcdStatus(c.ExtClient.KubedbV1alpha1(), etcd, func(in *api.EtcdStatus) *api.EtcdStatus {
		in.Phase = api.DatabasePhaseInitializing
		return in
	})
	if err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
		}
		return err
	}
	etcd.Status = db.Status

	return nil
}

func (c *Controller) terminate(etcd *api.Etcd) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd)
	if rerr != nil {
		return rerr
	}

	// If TerminationPolicy is "terminate", keep everything (ie, PVCs,Secrets,Snapshots) intact.
	// In operator, create dormantdatabase
	if etcd.Spec.TerminationPolicy == api.TerminationPolicyPause {
		if err := c.removeOwnerReferenceFromOffshoots(etcd, ref); err != nil {
			return err
		}
		if _, err := c.createDormantDatabase(etcd); err != nil {
			if kerr.IsAlreadyExists(err) {
				// if already exists, check if it is database of another Kind and return error in that case.
				// If the Kind is same, we can safely assume that the DormantDB was not deleted in before,
				// Probably because, User is more faster (create-delete-create-again-delete...) than operator!
				// So reuse that DormantDB!
				ddb, err := c.ExtClient.KubedbV1alpha1().DormantDatabases(etcd.Namespace).Get(etcd.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if val, _ := meta_util.GetStringValue(ddb.Labels, api.LabelDatabaseKind); val != api.ResourceKindEtcd {
					return fmt.Errorf(`DormantDatabase "%v" of kind %v already exists`, etcd.Name, val)
				}
			} else {
				return fmt.Errorf(`Failed to create DormantDatabase: "%v". Reason: %v`, etcd.Name, err)
			}
		}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshoots(etcd, ref); err != nil {
			return err
		}
	}

	c.cronController.StopBackupScheduling(etcd.ObjectMeta)

	if etcd.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(etcd); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}

func (c *Controller) setOwnerReferenceToOffshoots(etcd *api.Etcd, ref *core.ObjectReference) error {
	selector := labels.SelectorFromSet(etcd.OffshootSelectors())

	// If TerminationPolicy is "wipeOut", delete snapshots and secrets,
	// else, keep it intact.
	if etcd.Spec.TerminationPolicy == api.TerminationPolicyWipeOut {
		if err := dynamic_util.EnsureOwnerReferenceForSelector(
			c.DynamicClient,
			api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
			etcd.Namespace,
			selector,
			ref); err != nil {
			return err
		}
	} else {
		// Make sure snapshot and secret's ownerreference is removed.
		if err := dynamic_util.RemoveOwnerReferenceForSelector(
			c.DynamicClient,
			api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
			etcd.Namespace,
			selector,
			ref); err != nil {
			return err
		}
	}
	// delete PVC for both "wipeOut" and "delete" TerminationPolicy.
	return dynamic_util.EnsureOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		etcd.Namespace,
		selector,
		ref)
}

func (c *Controller) removeOwnerReferenceFromOffshoots(etcd *api.Etcd, ref *core.ObjectReference) error {
	// First, Get LabelSelector for Other Components
	labelSelector := labels.SelectorFromSet(etcd.OffshootSelectors())

	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		c.DynamicClient,
		api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
		etcd.Namespace,
		labelSelector,
		ref); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		etcd.Namespace,
		labelSelector,
		ref); err != nil {
		return err
	}
	return nil
}
