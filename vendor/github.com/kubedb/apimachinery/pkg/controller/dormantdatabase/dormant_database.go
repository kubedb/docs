package dormantdatabase

import (
	"github.com/appscode/go/encoding/json/types"
	meta_util "github.com/appscode/kutil/meta"
	"github.com/kubedb/apimachinery/apis"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) create(ddb *api.DormantDatabase) error {
	if ddb.Status.Phase == api.DormantDatabasePhasePaused {
		return nil
	}

	drmn, err := util.UpdateDormantDatabaseStatus(c.ExtClient.KubedbV1alpha1(), ddb, func(in *api.DormantDatabaseStatus) *api.DormantDatabaseStatus {
		in.Phase = api.DormantDatabasePhasePausing
		return in
	}, apis.EnableStatusSubresource)
	if err != nil {
		c.recorder.Eventf(ddb, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}
	*ddb = *drmn

	c.recorder.Event(ddb, core.EventTypeNormal, eventer.EventReasonPausing, "Pausing Database")

	// Pause Database workload
	if err := c.deleter.WaitUntilPaused(ddb); err != nil {
		c.recorder.Eventf(
			ddb,
			core.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			"Failed to pause. Reason: %v",
			err,
		)
		return err
	}

	c.recorder.Event(
		ddb,
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulPause,
		"Successfully paused Database workload",
	)

	_, err = util.UpdateDormantDatabaseStatus(c.ExtClient.KubedbV1alpha1(), ddb, func(in *api.DormantDatabaseStatus) *api.DormantDatabaseStatus {
		t := metav1.Now()
		in.PausingTime = &t
		in.Phase = api.DormantDatabasePhasePaused
		in.ObservedGeneration = types.NewIntHash(ddb.Generation, meta_util.GenerationHash(ddb))
		return in
	}, apis.EnableStatusSubresource)
	if err != nil {
		c.recorder.Eventf(ddb, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	return nil
}

func (c *Controller) delete(dormantDb *api.DormantDatabase) error {
	if dormantDb.Spec.WipeOut {
		if err := c.deleter.WipeOutDatabase(dormantDb); err != nil {
			return err
		}
	}
	return nil
}
