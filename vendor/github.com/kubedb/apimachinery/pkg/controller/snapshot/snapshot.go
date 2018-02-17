package snapshot

import (
	"fmt"

	"github.com/appscode/go/log"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	"github.com/kubedb/apimachinery/pkg/storage"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *Controller) create(snapshot *api.Snapshot) error {
	snap, _, err := util.PatchSnapshot(c.ExtClient, snapshot, func(in *api.Snapshot) *api.Snapshot {
		t := metav1.Now()
		in.Status.StartTime = &t
		return in
	})
	if err != nil {
		c.eventRecorder.Eventf(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}
	snapshot.Status = snap.Status

	// Validate DatabaseSnapshot spec
	if err := c.snapshotter.ValidateSnapshot(snapshot); err != nil {
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		_, _, err = util.PatchSnapshot(c.ExtClient, snapshot, func(in *api.Snapshot) *api.Snapshot {
			t := metav1.Now()
			in.Status.CompletionTime = &t
			in.Labels[api.LabelDatabaseName] = snapshot.Spec.DatabaseName
			in.Status.Phase = api.SnapshotPhaseFailed
			in.Status.Reason = "Invalid Snapshot"
			return in
		})
		if err != nil {
			c.eventRecorder.Eventf(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}
		log.Errorln(err)
		return nil
	}

	// Check running snapshot
	running, err := c.isSnapshotRunning(snapshot)
	if err != nil {
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, err.Error())
		return err
	}
	if running {
		_, _, err = util.PatchSnapshot(c.ExtClient, snapshot, func(in *api.Snapshot) *api.Snapshot {
			t := metav1.Now()
			in.Status.CompletionTime = &t
			in.Status.Phase = api.SnapshotPhaseFailed
			in.Status.Reason = "One Snapshot is already Running"
			return in
		})
		if err != nil {
			c.eventRecorder.Eventf(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}
		return nil
	}

	runtimeObj, err := c.snapshotter.GetDatabase(metav1.ObjectMeta{Name: snapshot.Spec.DatabaseName, Namespace: snapshot.Namespace})
	if err != nil {
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToGet, err.Error())
		return err
	}

	c.eventRecorder.Event(api.ObjectReferenceFor(runtimeObj), core.EventTypeNormal, eventer.EventReasonStarting, "Backup running")
	c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeNormal, eventer.EventReasonStarting, "Backup running")

	secret, err := storage.NewOSMSecret(c.Client, snapshot)
	if err != nil {
		message := fmt.Sprintf("Failed to generate osm secret. Reason: %v", err)
		c.eventRecorder.Event(api.ObjectReferenceFor(runtimeObj), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		return err
	}
	_, err = c.Client.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil && !kerr.IsAlreadyExists(err) {
		message := fmt.Sprintf("Failed to create osm secret. Reason: %v", err)
		c.eventRecorder.Event(api.ObjectReferenceFor(runtimeObj), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		return err
	}

	job, err := c.snapshotter.GetSnapshotter(snapshot)
	if err != nil {
		message := fmt.Sprintf("Failed to take snapshot. Reason: %v", err)
		c.eventRecorder.Event(api.ObjectReferenceFor(runtimeObj), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		return err
	}

	_, _, err = util.PatchSnapshot(c.ExtClient, snapshot, func(in *api.Snapshot) *api.Snapshot {
		in.Labels[api.LabelDatabaseName] = snapshot.Spec.DatabaseName
		in.Labels[api.LabelSnapshotStatus] = string(api.SnapshotPhaseRunning)
		in.Status.Phase = api.SnapshotPhaseRunning
		return in
	})
	if err != nil {
		c.eventRecorder.Eventf(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	job, err = c.Client.BatchV1().Jobs(snapshot.Namespace).Create(job)
	if err != nil {
		message := fmt.Sprintf("Failed to take snapshot. Reason: %v", err)
		c.eventRecorder.Event(api.ObjectReferenceFor(runtimeObj), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		return err
	}

	if err := c.SetJobOwnerReference(snapshot, job); err != nil {
		log.Errorln(err)
	}

	return nil
}

func (c *Controller) delete(snapshot *api.Snapshot) error {
	runtimeObj, err := c.snapshotter.GetDatabase(metav1.ObjectMeta{Name: snapshot.Spec.DatabaseName, Namespace: snapshot.Namespace})
	if err != nil {
		if !kerr.IsNotFound(err) {
			c.eventRecorder.Event(
				snapshot.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToGet,
				err.Error(),
			)
			return err
		}
	}

	if runtimeObj != nil {
		c.eventRecorder.Eventf(
			api.ObjectReferenceFor(runtimeObj),
			core.EventTypeNormal,
			eventer.EventReasonWipingOut,
			"Wiping out Snapshot: %v",
			snapshot.Name,
		)
	}

	if err := c.snapshotter.WipeOutSnapshot(snapshot); err != nil {
		if runtimeObj != nil {
			c.eventRecorder.Eventf(
				api.ObjectReferenceFor(runtimeObj),
				core.EventTypeWarning,
				eventer.EventReasonFailedToWipeOut,
				"Failed to  wipeOut. Reason: %v",
				err,
			)
		}
		return err
	}

	if runtimeObj != nil {
		c.eventRecorder.Eventf(
			api.ObjectReferenceFor(runtimeObj),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulWipeOut,
			"Successfully wiped out Snapshot: %v",
			snapshot.Name,
		)
	}
	return nil
}

func (c *Controller) isSnapshotRunning(snapshot *api.Snapshot) (bool, error) {
	labelMap := map[string]string{
		api.LabelDatabaseKind:   snapshot.Labels[api.LabelDatabaseKind],
		api.LabelDatabaseName:   snapshot.Spec.DatabaseName,
		api.LabelSnapshotStatus: string(api.SnapshotPhaseRunning),
	}

	snapshotList, err := c.ExtClient.Snapshots(snapshot.Namespace).List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelMap).String(),
	})
	if err != nil {
		return false, err
	}

	if len(snapshotList.Items) > 0 {
		return true, nil
	}

	return false, nil
}
