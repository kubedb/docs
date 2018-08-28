package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	"github.com/appscode/kutil"
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	validator "github.com/kubedb/etcd/pkg/admission"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwatch "k8s.io/apimachinery/pkg/watch"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
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

	if err := validator.ValidateEtcd(c.Client, c.ExtClient, etcd); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd); rerr == nil {
			c.recorder.Event(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonInvalid,
				err.Error())
		}
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

	db, err := util.UpdateEtcdStatus(c.ExtClient, etcd, func(in *api.EtcdStatus) *api.EtcdStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = etcd.Generation
		in.ObservedGenerationHash = meta_util.GenerationHash(etcd)
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
	// Setup Schedule backup
	if etcd.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(etcd, etcd.ObjectMeta, etcd.Spec.BackupSchedule)
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
	db, err := util.UpdateEtcdStatus(c.ExtClient, etcd, func(in *api.EtcdStatus) *api.EtcdStatus {
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

func (c *Controller) pause(etcd *api.Etcd) error {
	if _, err := c.createDormantDatabase(etcd); err != nil {
		if kerr.IsAlreadyExists(err) {
			// if already exists, check if it is database of another Kind and return error in that case.
			// If the Kind is same, we can safely assume that the DormantDB was not deleted in before,
			// Probably because, User is more faster (create-delete-create-again-delete...) than operator!
			// So reuse that DormantDB!
			ddb, err := c.ExtClient.DormantDatabases(etcd.Namespace).Get(etcd.Name, metav1.GetOptions{})
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

	c.cronController.StopBackupScheduling(etcd.ObjectMeta)

	if etcd.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(etcd); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}
