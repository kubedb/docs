package dormantdatabase

import (
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
)

func (c *Controller) create(ddb *api.DormantDatabase) error {
	if ddb.Status.CreationTime == nil {
		_, err := util.UpdateDormantDatabaseStatus(c.ExtClient, ddb, func(in *api.DormantDatabaseStatus) *api.DormantDatabaseStatus {
			t := metav1.Now()
			in.CreationTime = &t
			return in
		}, api.EnableStatusSubresource)
		if err != nil {
			if ref, rerr := reference.GetReference(clientsetscheme.Scheme, ddb); rerr == nil {
				c.recorder.Eventf(
					ref,
					core.EventTypeWarning,
					eventer.EventReasonFailedToUpdate,
					err.Error(),
				)
			}
			return err
		}
	}

	if ddb.Status.Phase == api.DormantDatabasePhasePaused {
		return nil
	}

	_, err := util.UpdateDormantDatabaseStatus(c.ExtClient, ddb, func(in *api.DormantDatabaseStatus) *api.DormantDatabaseStatus {
		in.Phase = api.DormantDatabasePhasePausing
		return in
	}, api.EnableStatusSubresource)
	if err != nil {
		c.recorder.Eventf(ddb, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	c.recorder.Event(ddb, core.EventTypeNormal, eventer.EventReasonPausing, "Pausing Database")

	// Pause Database workload
	if err := c.deleter.WaitUntilPaused(ddb); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, ddb); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToDelete,
				"Failed to pause. Reason: %v",
				err,
			)
		}
		return err
	}

	if ref, rerr := reference.GetReference(clientsetscheme.Scheme, ddb); rerr == nil {
		c.recorder.Event(
			ref,
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulPause,
			"Successfully paused Database workload",
		)
	}

	_, err = util.UpdateDormantDatabaseStatus(c.ExtClient, ddb, func(in *api.DormantDatabaseStatus) *api.DormantDatabaseStatus {
		t := metav1.Now()
		in.PausingTime = &t
		in.Phase = api.DormantDatabasePhasePaused
		return in
	}, api.EnableStatusSubresource)
	if err != nil {
		c.recorder.Eventf(ddb, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	return nil
}
