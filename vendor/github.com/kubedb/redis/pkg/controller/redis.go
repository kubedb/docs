package controller

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/appscode/go/log"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	"github.com/kubedb/redis/pkg/validator"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) create(redis *api.Redis) error {
	_, err := util.TryPatchRedis(c.ExtClient, redis.ObjectMeta, func(in *api.Redis) *api.Redis {
		t := metav1.Now()
		in.Status.CreationTime = &t
		in.Status.Phase = api.DatabasePhaseCreating
		return in
	})

	if err != nil {
		c.recorder.Eventf(redis.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	if err := validator.ValidateRedis(c.Client, redis); err != nil {
		c.recorder.Event(redis.ObjectReference(), core.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}
	// Event for successful validation
	c.recorder.Event(
		redis.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulValidate,
		"Successfully validate Redis",
	)

	// Check DormantDatabase
	matched, err := c.matchDormantDatabase(redis)
	if err != nil {
		return err
	}
	if matched {
		//TODO: Use Annotation Key
		redis.Annotations = map[string]string{
			"kubedb.com/ignore": "",
		}
		if err := c.ExtClient.Redises(redis.Namespace).Delete(redis.Name, &metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf(
				`Failed to resume Redis "%v" from DormantDatabase "%v". Error: %v`,
				redis.Name,
				redis.Name,
				err,
			)
		}

		_, err := util.TryPatchDormantDatabase(c.ExtClient, redis.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
			in.Spec.Resume = true
			return in
		})
		if err != nil {
			c.recorder.Eventf(redis.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}
		return nil
	}

	// Event for notification that kubernetes objects are creating
	c.recorder.Event(redis.ObjectReference(), core.EventTypeNormal, eventer.EventReasonCreating, "Creating Kubernetes objects")

	// create Governing Service
	governingService := c.opt.GoverningService
	if err := c.CreateGoverningService(governingService, redis.Namespace); err != nil {
		c.recorder.Eventf(
			redis.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create Service: "%v". Reason: %v`,
			governingService,
			err,
		)
		return err
	}

	// ensure database Service
	if err := c.ensureService(redis); err != nil {
		return err
	}

	// ensure database StatefulSet
	if err := c.ensureStatefulSet(redis); err != nil {
		return err
	}

	c.recorder.Event(
		redis.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		"Successfully created Redis",
	)

	if redis.Spec.Monitor != nil {
		if err := c.addMonitor(redis); err != nil {
			c.recorder.Eventf(
				redis.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				"Failed to add monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			redis.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulCreate,
			"Successfully added monitoring system.",
		)
	}
	return nil
}

func (c *Controller) matchDormantDatabase(redis *api.Redis) (bool, error) {
	// Check if DormantDatabase exists or not
	dormantDb, err := c.ExtClient.DormantDatabases(redis.Namespace).Get(redis.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			c.recorder.Eventf(
				redis.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToGet,
				`Fail to get DormantDatabase: "%v". Reason: %v`,
				redis.Name,
				err,
			)
			return false, err
		}
		return false, nil
	}

	var sendEvent = func(message string) (bool, error) {
		c.recorder.Event(
			redis.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			message,
		)
		return false, errors.New(message)
	}

	// Check DatabaseKind
	if dormantDb.Labels[api.LabelDatabaseKind] != api.ResourceKindRedis {
		return sendEvent(fmt.Sprintf(`Invalid Redis: "%v". Exists DormantDatabase "%v" of different Kind`,
			redis.Name, dormantDb.Name))
	}

	// Check Origin Spec
	drmnOriginSpec := dormantDb.Spec.Origin.Spec.Redis
	originalSpec := redis.Spec

	if !reflect.DeepEqual(drmnOriginSpec, &originalSpec) {
		return sendEvent("Redis spec mismatches with OriginSpec in DormantDatabases")
	}

	return true, nil
}

func (c *Controller) ensureService(redis *api.Redis) error {
	// Check if service name exists
	found, err := c.findService(redis)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	// create database Service
	if err := c.createService(redis); err != nil {
		c.recorder.Eventf(
			redis.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to create Service. Reason: %v",
			err,
		)
		return err
	}
	return nil
}

func (c *Controller) ensureStatefulSet(redis *api.Redis) error {
	found, err := c.findStatefulSet(redis)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	// Create statefulSet for Redis database
	statefulSet, err := c.createStatefulSet(redis)
	if err != nil {
		c.recorder.Eventf(
			redis.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to create StatefulSet. Reason: %v",
			err,
		)
		return err
	}

	_redis, err := c.ExtClient.Redises(redis.Namespace).Get(redis.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	redis = _redis

	// Check StatefulSet Pod status
	if err := c.CheckStatefulSetPodStatus(statefulSet, durationCheckStatefulSet); err != nil {
		c.recorder.Eventf(
			redis.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToStart,
			`Failed to create StatefulSet. Reason: %v`,
			err,
		)
		return err
	} else {
		c.recorder.Event(
			redis.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulCreate,
			"Successfully created StatefulSet",
		)
	}

	_, err = util.TryPatchRedis(c.ExtClient, redis.ObjectMeta, func(in *api.Redis) *api.Redis {
		in.Status.Phase = api.DatabasePhaseRunning
		return in
	})
	if err != nil {
		c.recorder.Eventf(redis, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}
	return nil
}

func (c *Controller) pause(redis *api.Redis) error {
	if redis.Annotations != nil {
		if val, found := redis.Annotations["kubedb.com/ignore"]; found {
			//TODO: Add Event Reason "Ignored"
			c.recorder.Event(redis.ObjectReference(), core.EventTypeNormal, "Ignored", val)
			return nil
		}
	}

	c.recorder.Event(redis.ObjectReference(), core.EventTypeNormal, eventer.EventReasonPausing, "Pausing Redis")

	if redis.Spec.DoNotPause {
		c.recorder.Eventf(
			redis.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToPause,
			`Redis "%v" is locked.`,
			redis.Name,
		)

		if err := c.reCreateRedis(redis); err != nil {
			c.recorder.Eventf(
				redis.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to recreate Redis: "%v". Reason: %v`,
				redis.Name,
				err,
			)
			return err
		}
		return nil
	}

	if _, err := c.createDormantDatabase(redis); err != nil {
		c.recorder.Eventf(
			redis.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create DormantDatabase: "%v". Reason: %v`,
			redis.Name,
			err,
		)
		return err
	}
	c.recorder.Eventf(
		redis.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		`Successfully created DormantDatabase: "%v"`,
		redis.Name,
	)

	if redis.Spec.Monitor != nil {
		if err := c.deleteMonitor(redis); err != nil {
			c.recorder.Eventf(
				redis.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToDelete,
				"Failed to delete monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			redis.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorDelete,
			"Successfully deleted monitoring system.",
		)
	}
	return nil
}

func (c *Controller) update(oldRedis, updatedRedis *api.Redis) error {
	if err := validator.ValidateRedis(c.Client, updatedRedis); err != nil {
		c.recorder.Event(updatedRedis.ObjectReference(), core.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}
	// Event for successful validation
	c.recorder.Event(
		updatedRedis.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulValidate,
		"Successfully validate Redis",
	)

	if err := c.ensureService(updatedRedis); err != nil {
		return err
	}
	if err := c.ensureStatefulSet(updatedRedis); err != nil {
		return err
	}

	if !reflect.DeepEqual(oldRedis.Spec.Monitor, updatedRedis.Spec.Monitor) {
		if err := c.updateMonitor(oldRedis, updatedRedis); err != nil {
			c.recorder.Eventf(
				updatedRedis.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				"Failed to update monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			updatedRedis.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorUpdate,
			"Successfully updated monitoring system.",
		)

	}
	return nil
}
