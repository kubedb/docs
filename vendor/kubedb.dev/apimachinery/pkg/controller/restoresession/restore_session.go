package restoresession

import (
	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"
	"stash.appscode.dev/stash/apis/stash/v1beta1"
)

func (c *Controller) handleRestoreSession(rs *v1beta1.RestoreSession) error {
	if rs.Status.Phase != v1beta1.RestoreSessionSucceeded && rs.Status.Phase != v1beta1.RestoreSessionFailed {
		log.Debugf("restoreSession %v/%v is not any of 'succeeded' or 'failed' ", rs.Namespace, rs.Name)
		return nil
	}

	if rs.Spec.Target == nil {
		log.Debugf("restoreSession %v/%v does not have spec.target set. ", rs.Namespace, rs.Name)
		return nil
	}

	meta := metav1.ObjectMeta{
		Name:      rs.Spec.Target.Ref.Name,
		Namespace: rs.Namespace,
	}

	var phase api.DatabasePhase
	var reason string
	if rs.Status.Phase == v1beta1.RestoreSessionSucceeded {
		phase = api.DatabasePhaseRunning
		if err := c.snapshotter.UpsertDatabaseAnnotation(meta, map[string]string{
			api.AnnotationInitialized: "",
		}); err != nil {
			return err
		}
	} else {
		phase = api.DatabasePhaseFailed
		reason = "Failed to complete initialization"
	}
	if err := c.snapshotter.SetDatabaseStatus(meta, phase, reason); err != nil {
		return err
	}

	runtimeObj, err := c.snapshotter.GetDatabase(meta)
	if err != nil {
		log.Errorln(err)
		return nil
	}
	if rs.Status.Phase == v1beta1.RestoreSessionSucceeded {
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
