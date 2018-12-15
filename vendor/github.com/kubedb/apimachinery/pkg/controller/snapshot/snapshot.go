package snapshot

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/kutil"
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
				eventer.EventReasonSnapshotError,
				err.Error(),
			)
			return err
		}
		*snapshot = *snap
	}

	// Do not process "completed", aka "failed" or "succeeded", snapshots.
	if snapshot.Status.Phase == api.SnapshotPhaseFailed || snapshot.Status.Phase == api.SnapshotPhaseSucceeded {
		// Although the snapshot is "completed", yet the snapshot CRD can contain annotation "snapshot.kubedb.com/status: Running",
		// due to failure of operator in critical moment. So, Make sure the "completed" snapshot doesn't have
		// annotation "snapshot.kubedb.com/status: Running".
		// Error may occur when the operator just restarted and the Webhooks is not in working state yet. So retry for "InternalError".
		if _, _, err := util.PatchSnapshot(c.ExtClient.KubedbV1alpha1(), snapshot, func(in *api.Snapshot) *api.Snapshot {
			delete(in.Labels, api.LabelSnapshotStatus)
			return in
		}); err != nil {
			c.eventRecorder.Eventf(
				snapshot,
				core.EventTypeWarning,
				eventer.EventReasonSnapshotError,
				err.Error(),
			)
			return err
		}
		return nil
	}

	if _, _, err := util.PatchSnapshot(c.ExtClient.KubedbV1alpha1(), snapshot, func(in *api.Snapshot) *api.Snapshot {
		in.Labels[api.LabelDatabaseName] = snapshot.Spec.DatabaseName
		return in
	}); err != nil {
		c.eventRecorder.Eventf(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotError,
			err.Error(),
		)
		return err
	}

	// Validate DatabaseSnapshot
	if err := c.snapshotter.ValidateSnapshot(snapshot); err != nil {
		if kutil.IsRequestRetryable(err) {
			return err
		}

		if _, e2 := util.MarkAsFailedSnapshot(c.ExtClient.KubedbV1alpha1(), snapshot, err.Error(), apis.EnableStatusSubresource); e2 != nil {
			return e2 // retry if retryable error
		}
		c.eventRecorder.Event(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			err.Error(),
		)
		return nil // stop retry
	}

	// Check running snapshot
	running, err := c.isSnapshotRunning(snapshot)
	if err != nil {
		c.eventRecorder.Event(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotError,
			err.Error(),
		)
		return err
	}
	if running {
		msg := "One Snapshot is already Running"
		if _, e2 := util.MarkAsFailedSnapshot(c.ExtClient.KubedbV1alpha1(), snapshot, msg, apis.EnableStatusSubresource); e2 != nil {
			return e2 // retry if retryable error
		}
		c.eventRecorder.Event(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			msg,
		)
		return nil
	}

	db, err := c.snapshotter.GetDatabase(metav1.ObjectMeta{Name: snapshot.Spec.DatabaseName, Namespace: snapshot.Namespace})
	if err != nil {
		c.eventRecorder.Event(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotError,
			err.Error(),
		)
		return err
	}

	secret, err := osm.NewOSMSecret(c.Client, snapshot.OSMSecretName(), snapshot.Namespace, snapshot.Spec.Backend)
	if err != nil {
		msg := fmt.Sprintf("Failed to generate osm secret. Reason: %v", err)
		if _, e2 := util.MarkAsFailedSnapshot(c.ExtClient.KubedbV1alpha1(), snapshot, msg, apis.EnableStatusSubresource); e2 != nil {
			return e2 // retry if retryable error
		}
		c.eventRecorder.Event(
			db,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			msg,
		)
		c.eventRecorder.Event(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotFailed,
			msg,
		)
		return nil // don't retry
	}

	if _, err = c.Client.CoreV1().Secrets(secret.Namespace).Create(secret); err != nil && !kerr.IsAlreadyExists(err) {
		message := fmt.Sprintf("Failed to create osm secret. Reason: %v", err)
		c.eventRecorder.Event(
			db,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotError,
			message,
		)
		c.eventRecorder.Event(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotError,
			message,
		)
		return err
	}

	// Do not check bucket access for local volume
	if snapshot.Spec.Local == nil {
		if err := osm.CheckBucketAccess(c.Client, snapshot.Spec.Backend, snapshot.Namespace); err != nil {
			if _, e2 := util.MarkAsFailedSnapshot(c.ExtClient.KubedbV1alpha1(), snapshot, err.Error(), apis.EnableStatusSubresource); e2 != nil {
				return e2 // retry if retryable error
			}
			c.eventRecorder.Eventf(
				snapshot,
				core.EventTypeWarning,
				eventer.EventReasonSnapshotFailed,
				err.Error(),
			)
			return nil // don't retry
		}
	}

	job, err := c.snapshotter.GetSnapshotter(snapshot)
	if err != nil {
		message := fmt.Sprintf("Failed to create Snapshotter Job. Reason: %v", err)
		c.eventRecorder.Event(
			db,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotError,
			message,
		)
		c.eventRecorder.Event(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotError,
			message,
		)

		// If error is not retryable then mark the snapshot as failed and don't retry
		if !kutil.IsRequestRetryable(err) {
			if _, e2 := util.MarkAsFailedSnapshot(c.ExtClient.KubedbV1alpha1(), snapshot, err.Error(), apis.EnableStatusSubresource); e2 != nil {
				return e2 // retry if retryable error
			}
			return nil // don't retry
		}

		return err
	}

	if _, err := util.UpdateSnapshotStatus(c.ExtClient.KubedbV1alpha1(), snapshot, func(in *api.SnapshotStatus) *api.SnapshotStatus {
		in.Phase = api.SnapshotPhaseRunning
		return in
	}, apis.EnableStatusSubresource); err != nil {
		c.eventRecorder.Eventf(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotError,
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
			eventer.EventReasonSnapshotError,
			err.Error(),
		)
		return err
	}

	c.eventRecorder.Event(
		db,
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

	job, err = c.Client.BatchV1().Jobs(snapshot.Namespace).Create(job)
	if err != nil {
		message := fmt.Sprintf("Failed to create snapshot job. Reason: %v", err)
		c.eventRecorder.Event(
			db,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotError,
			message,
		)
		c.eventRecorder.Event(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotError,
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
	db, err := c.snapshotter.GetDatabase(metav1.ObjectMeta{Name: snapshot.Spec.DatabaseName, Namespace: snapshot.Namespace})
	// Database may not exists while dormantdb is deleted and wipedout.
	// So, skip error if not found. But process the rest.
	if err != nil && !kerr.IsNotFound(err) {
		c.eventRecorder.Event(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonSnapshotError,
			err.Error(),
		)
		return err
	}

	if db != nil {
		c.eventRecorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonWipingOut,
			"Wiping out Snapshot: %v",
			snapshot.Name,
		)
	}

	if err := c.snapshotter.WipeOutSnapshot(snapshot); err != nil {
		if db != nil {
			c.eventRecorder.Eventf(
				db,
				core.EventTypeWarning,
				eventer.EventReasonFailedToWipeOut,
				"Failed to wipeOut. Reason: %v",
				err,
			)
		}
		return err
	}

	if db != nil {
		c.eventRecorder.Eventf(
			db,
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
