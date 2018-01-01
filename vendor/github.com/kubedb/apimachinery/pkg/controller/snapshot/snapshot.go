package snapshot

import (
	"errors"
	"fmt"
	"time"

	"github.com/appscode/go/log"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	"github.com/kubedb/apimachinery/pkg/storage"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	durationCheckSnapshotJob = time.Minute * 30
	sleepDuration            = time.Second
)

func (c *Controller) create(snapshot *api.Snapshot) error {
	_, _, err := util.PatchSnapshot(c.ExtClient, snapshot, func(in *api.Snapshot) *api.Snapshot {
		t := metav1.Now()
		in.Status.StartTime = &t
		return in
	})
	if err != nil {
		c.eventRecorder.Eventf(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	// Validate DatabaseSnapshot spec
	if err := c.snapshotter.ValidateSnapshot(snapshot); err != nil {
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}

	// Check running snapshot
	if err := c.checkRunningSnapshot(snapshot); err != nil {
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, err.Error())
		return err
	}

	runtimeObj, err := c.snapshotter.GetDatabase(snapshot)
	if err != nil {
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToGet, err.Error())
		return err
	}

	snap, _, err := util.PatchSnapshot(c.ExtClient, snapshot, func(in *api.Snapshot) *api.Snapshot {
		in.Labels[api.LabelDatabaseName] = snapshot.Spec.DatabaseName
		in.Labels[api.LabelSnapshotStatus] = string(api.SnapshotPhaseRunning)
		in.Status.Phase = api.SnapshotPhaseRunning
		return in
	})
	if err != nil {
		c.eventRecorder.Eventf(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}
	snapshot.Labels = snap.Labels
	snapshot.Status = snap.Status

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
	if _, err := c.Client.BatchV1().Jobs(snapshot.Namespace).Create(job); err != nil {
		message := fmt.Sprintf("Failed to take snapshot. Reason: %v", err)
		c.eventRecorder.Event(api.ObjectReferenceFor(runtimeObj), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonSnapshotFailed, message)
		return err
	}

	return c.checkSnapshotJob(snapshot, job.Name, durationCheckSnapshotJob)
}

func (c *Controller) delete(snapshot *api.Snapshot) error {
	runtimeObj, err := c.snapshotter.GetDatabase(snapshot)
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

func (c *Controller) checkRunningSnapshot(snapshot *api.Snapshot) error {
	labelMap := map[string]string{
		api.LabelDatabaseKind:   snapshot.Labels[api.LabelDatabaseKind],
		api.LabelDatabaseName:   snapshot.Spec.DatabaseName,
		api.LabelSnapshotStatus: string(api.SnapshotPhaseRunning),
	}

	snapshotList, err := c.ExtClient.Snapshots(snapshot.Namespace).List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelMap).String(),
	})
	if err != nil {
		return err
	}

	if len(snapshotList.Items) > 0 {
		_, _, err = util.PatchSnapshot(c.ExtClient, snapshot, func(in *api.Snapshot) *api.Snapshot {
			t := metav1.Now()
			in.Status.StartTime = &t
			in.Status.CompletionTime = &t
			in.Status.Phase = api.SnapshotPhaseFailed
			in.Status.Reason = "One Snapshot is already Running"
			return in
		})
		if err != nil {
			c.eventRecorder.Eventf(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}

		return errors.New("one Snapshot is already Running")
	}

	return nil
}

func (c *Controller) checkSnapshotJob(snapshot *api.Snapshot, jobName string, checkDuration time.Duration) error {

	var jobSuccess bool = false
	var job *batch.Job
	var err error
	then := time.Now()
	now := time.Now()
	for now.Sub(then) < checkDuration {
		log.Debugln("Checking for Job ", jobName)
		job, err = c.Client.BatchV1().Jobs(snapshot.Namespace).Get(jobName, metav1.GetOptions{})
		if err != nil {
			if kerr.IsNotFound(err) {
				time.Sleep(sleepDuration)
				now = time.Now()
				continue
			}
			c.eventRecorder.Eventf(
				snapshot.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToList,
				"Failed to get Job. Reason: %v",
				err,
			)
			return err
		}
		log.Debugf("Pods Statuses:	%d Running / %d Succeeded / %d Failed",
			job.Status.Active, job.Status.Succeeded, job.Status.Failed)
		// If job is success
		if job.Status.Succeeded > 0 {
			jobSuccess = true
			break
		} else if job.Status.Failed > 0 {
			break
		}

		time.Sleep(sleepDuration)
		now = time.Now()
	}

	if err != nil {
		return err
	}

	c.DeleteJobResources(c.eventRecorder, snapshot, job)

	err = c.Client.CoreV1().Secrets(snapshot.Namespace).Delete(snapshot.OSMSecretName(), &metav1.DeleteOptions{})
	if err != nil && !kerr.IsNotFound(err) {
		c.eventRecorder.Eventf(
			snapshot.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			"Failed to delete Secret. Reason: %v",
			err,
		)
		log.Errorln(err)
	}

	runtimeObj, err := c.snapshotter.GetDatabase(snapshot)
	if err != nil {
		c.eventRecorder.Event(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToGet, err.Error())
		return err
	}

	var snapshotPhase api.SnapshotPhase
	if jobSuccess {
		snapshotPhase = api.SnapshotPhaseSuccessed
		c.eventRecorder.Event(
			api.ObjectReferenceFor(runtimeObj),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulSnapshot,
			"Successfully completed snapshot",
		)
		c.eventRecorder.Event(
			snapshot.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulSnapshot,
			"Successfully completed snapshot",
		)
	} else {
		snapshotPhase = api.SnapshotPhaseFailed
		c.eventRecorder.Event(
			api.ObjectReferenceFor(runtimeObj),
			core.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			"Failed to complete snapshot",
		)
		c.eventRecorder.Event(
			snapshot.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			"Failed to complete snapshot",
		)
	}

	snap, _, err := util.PatchSnapshot(c.ExtClient, snapshot, func(in *api.Snapshot) *api.Snapshot {
		t := metav1.Now()
		in.Status.CompletionTime = &t
		delete(in.Labels, api.LabelSnapshotStatus)
		in.Status.Phase = snapshotPhase
		return in
	})
	if err != nil {
		c.eventRecorder.Eventf(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}
	snapshot.Labels = snap.Labels
	snapshot.Status = snap.Status

	return nil
}
