package job

import (
	"fmt"

	"github.com/appscode/go/log"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubedb.dev/apimachinery/apis"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/apimachinery/pkg/eventer"
)

func (c *Controller) completeJob(job *batch.Job) error {
	deletePolicy := metav1.DeletePropagationBackground
	err := c.Client.BatchV1().Jobs(job.Namespace).Delete(job.Name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})

	if err != nil && !kerr.IsNotFound(err) {
		return fmt.Errorf("failed to delete job: %s, reason: %s", job.Name, err)
	}

	jobType := job.Labels[api.AnnotationJobType]
	if jobType == api.JobTypeBackup {
		return c.handleBackupJob(job)
	} else if jobType == api.JobTypeRestore {
		return c.handleRestoreJob(job)
	}

	return nil
}

func (c *Controller) handleBackupJob(job *batch.Job) error {
	for _, o := range job.OwnerReferences {
		if o.Kind == api.ResourceKindSnapshot {
			snapshot, err := c.ExtClient.KubedbV1alpha1().Snapshots(job.Namespace).Get(o.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			jobSucceeded := job.Status.Succeeded > 0

			if _, err := util.UpdateSnapshotStatus(c.ExtClient.KubedbV1alpha1(), snapshot, func(in *api.SnapshotStatus) *api.SnapshotStatus {
				if jobSucceeded {
					in.Phase = api.SnapshotPhaseSucceeded
				} else {
					in.Phase = api.SnapshotPhaseFailed
				}
				t := metav1.Now()
				in.CompletionTime = &t
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

			if _, _, err := util.PatchSnapshot(c.ExtClient.KubedbV1alpha1(), snapshot, func(in *api.Snapshot) *api.Snapshot {
				delete(in.Labels, api.LabelSnapshotStatus)
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

			runtimeObj, err := c.snapshotter.GetDatabase(metav1.ObjectMeta{Name: snapshot.Spec.DatabaseName, Namespace: snapshot.Namespace})
			if err != nil {
				return nil
			}
			if jobSucceeded {
				c.eventRecorder.Event(
					runtimeObj,
					core.EventTypeNormal,
					eventer.EventReasonSuccessfulSnapshot,
					"Successfully completed snapshot",
				)
				c.eventRecorder.Event(
					snapshot,
					core.EventTypeNormal,
					eventer.EventReasonSuccessfulSnapshot,
					"Successfully completed snapshot",
				)
			} else {
				c.eventRecorder.Event(
					runtimeObj,
					core.EventTypeWarning,
					eventer.EventReasonSnapshotFailed,
					"Failed to complete snapshot",
				)
				c.eventRecorder.Event(
					snapshot,
					core.EventTypeWarning,
					eventer.EventReasonSnapshotFailed,
					"Failed to complete snapshot",
				)
			}

			return nil
		}
	}

	log.Errorf(`resource Job "%s/%s" doesn't have OwnerReference for Snapshot`, job.Namespace, job.Name)
	return nil
}

func (c *Controller) handleRestoreJob(job *batch.Job) error {
	for _, o := range job.OwnerReferences {
		if o.Kind == job.Labels[api.LabelDatabaseKind] {
			jobSucceeded := job.Status.Succeeded > 0

			var phase api.DatabasePhase
			var reason string
			objectMeta := metav1.ObjectMeta{Name: o.Name, Namespace: job.Namespace}

			if jobSucceeded {
				phase = api.DatabasePhaseRunning
				if err := c.snapshotter.UpsertDatabaseAnnotation(objectMeta, map[string]string{
					api.AnnotationInitialized: "",
				}); err != nil {
					return err
				}
			} else {
				phase = api.DatabasePhaseFailed
				reason = "Failed to complete initialization"
			}
			if err := c.snapshotter.SetDatabaseStatus(objectMeta, phase, reason); err != nil {
				return err
			}
			runtimeObj, err := c.snapshotter.GetDatabase(objectMeta)
			if err != nil {
				log.Errorln(err)
				return nil
			}
			if jobSucceeded {
				c.eventRecorder.Event(
					runtimeObj,
					core.EventTypeNormal,
					eventer.EventReasonSuccessfulInitialize,
					"Successfully completed initialization",
				)
			} else {
				c.eventRecorder.Event(
					runtimeObj,
					core.EventTypeWarning,
					eventer.EventReasonFailedToInitialize,
					"Failed to complete initialization",
				)
			}
			return nil
		}
	}
	log.Errorf(`resource Job "%s/%s" doesn't have OwnerReference for %s`, job.Namespace, job.Name, job.Labels[api.LabelDatabaseKind])
	return nil
}
