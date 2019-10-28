package dormantdatabase

import (
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/appscode/go/encoding/json/types"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

func (c *Controller) create(ddb *api.DormantDatabase) error {
	if ddb.Status.Phase == api.DormantDatabasePhasePaused {
		return nil
	}

	drmn, err := util.UpdateDormantDatabaseStatus(c.ExtClient.KubedbV1alpha1(), ddb, func(in *api.DormantDatabaseStatus) *api.DormantDatabaseStatus {
		in.Phase = api.DormantDatabasePhasePausing
		return in
	})
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
	})
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
