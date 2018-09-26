package snapshot

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/kubedb/apimachinery/apis"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"kmodules.xyz/objectstore-api/osm"
)

func (c *Controller) create(snapshot *api.Snapshot) error {
	if snapshot.Status.StartTime == nil {
		snap, err := util.UpdateSnapshotStatus(c.ExtClient.KubedbV1alpha1(), snapshot, func(in *api.SnapshotStatus) *api.SnapshotStatus {
			t := metav1.Now()
			in.StartTime = &t
			return in
		}, apis.EnableStatusSubresource)
		if err != nil {
			c.eventRecorder.Eventf(
				snapshot,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return err
		}
		*snapshot = *snap
	}

	if snapshot.Status.Phase == api.SnapshotPhaseFailed || snapshot.Status.Phase == api.SnapshotPhaseSucceeded {
		return nil
	}

	// Validate DatabaseSnapshot
	if err := c.snapshotter.ValidateSnapshot(snapshot); err != nil {
		log.Errorln(err)
		c.eventRecorder.Event(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)

		if _, err := util.UpdateSnapshotStatus(c.ExtClient.KubedbV1alpha1(), snapshot, func(in *api.SnapshotStatus) *api.SnapshotStatus {
			t := metav1.Now()
			in.CompletionTime = &t
			in.Phase = api.SnapshotPhaseFailed
			in.Reason = "Invalid Snapshot"
			return in
		}, apis.EnableStatusSubresource); err != nil {
			log.Errorln(err)
			c.eventRecorder.Eventf(
				snapshot,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return err
		}

		if _, _, err = util.PatchSnapshot(c.ExtClient.KubedbV1alpha1(), snapshot, func(in *api.Snapshot) *api.Snapshot {
			in.Labels[api.LabelDatabaseName] = snapshot.Spec.DatabaseName
			return in
		}); err != nil {
			log.Errorln(err)
			c.eventRecorder.Eventf(
				snapshot,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return err
		}
		return nil
	}

	// Check running snapshot
	running, err := c.isSnapshotRunning(snapshot)
	if err != nil {
		c.eventRecorder.Event(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			err.Error(),
		)
		return err
	}
	if running {
		if _, err := util.UpdateSnapshotStatus(c.ExtClient.KubedbV1alpha1(), snapshot, func(in *api.SnapshotStatus) *api.SnapshotStatus {
			t := metav1.Now()
			in.CompletionTime = &t
			in.Phase = api.SnapshotPhaseFailed
			in.Reason = "One Snapshot is already Running"
			return in
		}, apis.EnableStatusSubresource); err != nil {
			c.eventRecorder.Eventf(
				snapshot,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return err
		}
		return nil
	}

	runtimeObj, err := c.snapshotter.GetDatabase(metav1.ObjectMeta{Name: snapshot.Spec.DatabaseName, Namespace: snapshot.Namespace})
	if err != nil {
		c.eventRecorder.Event(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonFailedToGet,
			err.Error(),
		)
		return err
	}

	c.eventRecorder.Event(
		runtimeObj,
		core.EventTypeNormal,
		eventer.EventReasonStarting,
		"Backup running",
	)
	c.eventRecorder.Event(
		snapshot,
		core.EventTypeNormal,
		eventer.EventReasonStarting,
		"Backup running",
	)
	secret, err := osm.NewOSMSecret(c.Client, snapshot.OSMSecretName(), snapshot.Namespace, snapshot.Spec.Backend)
	if err != nil {
		message := fmt.Sprintf("Failed to generate osm secret. Reason: %v", err)
		c.eventRecorder.Event(
			runtimeObj,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			message,
		)
		c.eventRecorder.Event(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			message,
		)
		return err
	}
	_, err = c.Client.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil && !kerr.IsAlreadyExists(err) {
		message := fmt.Sprintf("Failed to create osm secret. Reason: %v", err)
		c.eventRecorder.Event(
			runtimeObj,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			message,
		)
		c.eventRecorder.Event(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			message,
		)
		return err
	}

	// Do not check bucket access for local volume
	if snapshot.Spec.Local == nil {
		if err := osm.CheckBucketAccess(c.Client, snapshot.Spec.Backend, snapshot.Namespace); err != nil {
			return err
		}
	}

	job, err := c.snapshotter.GetSnapshotter(snapshot)
	if err != nil {
		message := fmt.Sprintf("Failed to take snapshot. Reason: %v", err)
		c.eventRecorder.Event(
			runtimeObj,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			message,
		)
		c.eventRecorder.Event(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			message,
		)
		return err
	}

	if _, err := util.UpdateSnapshotStatus(c.ExtClient.KubedbV1alpha1(), snapshot, func(in *api.SnapshotStatus) *api.SnapshotStatus {
		in.Phase = api.SnapshotPhaseRunning
		return in
	}, apis.EnableStatusSubresource); err != nil {
		c.eventRecorder.Eventf(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			err.Error(),
		)
		return err
	}

	if _, _, err = util.PatchSnapshot(c.ExtClient.KubedbV1alpha1(), snapshot, func(in *api.Snapshot) *api.Snapshot {
		in.Labels[api.LabelDatabaseName] = snapshot.Spec.DatabaseName
		in.Labels[api.LabelSnapshotStatus] = string(api.SnapshotPhaseRunning)
		return in
	}); err != nil {
		c.eventRecorder.Eventf(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			err.Error(),
		)
		return err
	}

	job, err = c.Client.BatchV1().Jobs(snapshot.Namespace).Create(job)
	if err != nil {
		message := fmt.Sprintf("Failed to take snapshot. Reason: %v", err)
		c.eventRecorder.Event(
			runtimeObj,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			message,
		)
		c.eventRecorder.Event(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			message,
		)
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
				snapshot,
				core.EventTypeWarning,
				eventer.EventReasonFailedToGet,
				err.Error(),
			)
			return err
		}
	}

	if runtimeObj != nil {
		c.eventRecorder.Eventf(
			runtimeObj,
			core.EventTypeNormal,
			eventer.EventReasonWipingOut,
			"Wiping out Snapshot: %v",
			snapshot.Name,
		)
	}

	if err := c.snapshotter.WipeOutSnapshot(snapshot); err != nil {
		if runtimeObj != nil {
			c.eventRecorder.Eventf(
				runtimeObj,
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
			runtimeObj,
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

	snapshotList, err := c.snLister.List(labels.SelectorFromSet(labelMap))
	if err != nil {
		return false, err
	}

	if len(snapshotList) > 0 {
		return true, nil
	}

	return false, nil
}
