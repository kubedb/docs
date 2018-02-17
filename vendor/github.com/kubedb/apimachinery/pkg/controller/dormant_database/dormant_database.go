package dormant_database

import (
	"github.com/appscode/go/log"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) create(dormantDb *api.DormantDatabase) error {
	if dormantDb.Spec.WipeOut {
		return c.wipeOut(dormantDb)
	}

	if dormantDb.Spec.Resume {
		if dormantDb.Status.Phase == api.DormantDatabasePhasePaused {
			return c.resume(dormantDb)
		}
		message := "Failed to resume Database. " +
			"Only DormantDatabase of \"Paused\" Phase can be resumed"
		c.recorder.Event(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			message,
		)
		return nil
	}

	_, _, err := util.PatchDormantDatabase(c.ExtClient, dormantDb, func(in *api.DormantDatabase) *api.DormantDatabase {
		t := metav1.Now()
		in.Status.CreationTime = &t
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	// Check if DB TPR object exists
	found, err := c.deleter.Exists(&dormantDb.ObjectMeta)
	if err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToPause,
			"Failed to pause Database. Reason: %v",
			err,
		)
		return err
	}

	if found {
		message := "failed to pause Database. Delete Database TPR object first"
		c.recorder.Event(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToPause,
			message,
		)

		// Delete DormantDatabase object
		if err := c.ExtClient.DormantDatabases(dormantDb.Namespace).Delete(dormantDb.Name, &metav1.DeleteOptions{}); err != nil {
			c.recorder.Eventf(
				dormantDb.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToDelete,
				"Failed to delete DormantDatabase. Reason: %v",
				err,
			)
			log.Errorln(err)
		}
		return errors.New(message)
	}

	_, _, err = util.PatchDormantDatabase(c.ExtClient, dormantDb, func(in *api.DormantDatabase) *api.DormantDatabase {
		in.Status.Phase = api.DormantDatabasePhasePausing
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	c.recorder.Event(dormantDb, core.EventTypeNormal, eventer.EventReasonPausing, "Pausing Database")

	// Pause Database workload
	if err := c.deleter.PauseDatabase(dormantDb); err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			"Failed to pause. Reason: %v",
			err,
		)
		return err
	}

	c.recorder.Event(
		dormantDb.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulPause,
		"Successfully paused Database workload",
	)

	_, _, err = util.PatchDormantDatabase(c.ExtClient, dormantDb, func(in *api.DormantDatabase) *api.DormantDatabase {
		t := metav1.Now()
		in.Status.PausingTime = &t
		in.Status.Phase = api.DormantDatabasePhasePaused
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	return nil
}

func (c *Controller) delete(dormantDb *api.DormantDatabase) error {

	exists, err := c.deleter.Exists(&dormantDb.ObjectMeta)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	phase := dormantDb.Status.Phase
	if phase != api.DormantDatabasePhaseResuming && phase != api.DormantDatabasePhaseWipedOut {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			`DormantDatabase "%v" is not %v.`,
			dormantDb.Name,
			api.DormantDatabasePhaseWipedOut,
		)

		if err := c.reCreateDormantDatabase(dormantDb); err != nil {
			c.recorder.Eventf(
				dormantDb.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to recreate DormantDatabase: "%v". Reason: %v`,
				dormantDb.Name,
				err,
			)
			return err
		}
	}
	return nil
}

func (c *Controller) wipeOut(dormantDb *api.DormantDatabase) error {
	// Check if DB TPR object exists
	found, err := c.deleter.Exists(&dormantDb.ObjectMeta)
	if err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			"Failed to wipeOut Database. Reason: %v",
			err,
		)
		return err
	}

	if found {
		message := "failed to wipeOut Database. Delete Database TPR object first"
		c.recorder.Event(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToWipeOut,
			message,
		)

		// Delete DormantDatabase object
		if err := c.ExtClient.DormantDatabases(dormantDb.Namespace).Delete(dormantDb.Name, &metav1.DeleteOptions{}); err != nil {
			c.recorder.Eventf(
				dormantDb.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToDelete,
				"Failed to delete DormantDatabase. Reason: %v",
				err,
			)
			log.Errorln(err)
		}
		return errors.New(message)
	}

	_, _, err = util.PatchDormantDatabase(c.ExtClient, dormantDb, func(in *api.DormantDatabase) *api.DormantDatabase {
		in.Status.Phase = api.DormantDatabasePhaseWipingOut
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	// Wipe out Database workload
	c.recorder.Event(dormantDb, core.EventTypeNormal, eventer.EventReasonWipingOut, "Wiping out Database")
	if err := c.deleter.WipeOutDatabase(dormantDb); err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToWipeOut,
			"Failed to wipeOut. Reason: %v",
			err,
		)
		return err
	}

	c.recorder.Event(
		dormantDb.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulWipeOut,
		"Successfully wiped out Database workload",
	)

	_, _, err = util.PatchDormantDatabase(c.ExtClient, dormantDb, func(in *api.DormantDatabase) *api.DormantDatabase {
		t := metav1.Now()
		in.Status.WipeOutTime = &t
		in.Status.Phase = api.DormantDatabasePhaseWipedOut
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	return nil
}

func (c *Controller) resume(dormantDb *api.DormantDatabase) error {
	c.recorder.Event(
		dormantDb.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonResuming,
		"Resuming DormantDatabase",
	)

	// Check if DB TPR object exists
	found, err := c.deleter.Exists(&dormantDb.ObjectMeta)
	if err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToResume,
			"Failed to resume DormantDatabase. Reason: %v",
			err,
		)
		return err
	}

	if found {
		message := "failed to resume DormantDatabase. One Database TPR object exists with same name"
		c.recorder.Event(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToResume,
			message,
		)
		return errors.New(message)
	}

	_, _, err = util.PatchDormantDatabase(c.ExtClient, dormantDb, func(in *api.DormantDatabase) *api.DormantDatabase {
		in.Status.Phase = api.DormantDatabasePhaseResuming
		return in
	})
	if err != nil {
		c.recorder.Eventf(dormantDb, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	if err = c.ExtClient.DormantDatabases(dormantDb.Namespace).Delete(dormantDb.Name, &metav1.DeleteOptions{}); err != nil {
		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			"Failed to delete DormantDatabase. Reason: %v",
			err,
		)
		return err
	}

	if err = c.deleter.ResumeDatabase(dormantDb); err != nil {
		if err := c.reCreateDormantDatabase(dormantDb); err != nil {
			c.recorder.Eventf(
				dormantDb.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to recreate DormantDatabase: "%v". Reason: %v`,
				dormantDb.Name,
				err,
			)
			return err
		}

		c.recorder.Eventf(
			dormantDb.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToResume,
			"Failed to resume Database. Reason: %v",
			err,
		)
		return err
	}
	return nil
}

func (c *Controller) reCreateDormantDatabase(dormantDatabase *api.DormantDatabase) error {
	dormantDb := &api.DormantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:        dormantDatabase.Name,
			Namespace:   dormantDatabase.Namespace,
			Labels:      dormantDatabase.Labels,
			Annotations: dormantDatabase.Annotations,
		},
		Spec:   dormantDatabase.Spec,
		Status: dormantDatabase.Status,
	}

	if _, err := c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb); err != nil {
		return err
	}

	return nil
}
