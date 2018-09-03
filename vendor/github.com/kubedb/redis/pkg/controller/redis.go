package controller

import (
	"fmt"

	"github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/log"
	"github.com/appscode/kutil"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	validator "github.com/kubedb/redis/pkg/admission"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
)

func (c *Controller) create(redis *api.Redis) error {
	if err := validator.ValidateRedis(c.Client, c.ExtClient, redis); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis); rerr == nil {
			c.recorder.Event(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonInvalid,
				err.Error(),
			)
		}
		log.Errorln(err)
		return nil // user error so just record error and don't retry.
	}

	// Delete Matching DormantDatabase if exists any
	if err := c.deleteMatchingDormantDatabase(redis); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to delete dormant Database : "%v". Reason: %v`,
				redis.Name,
				err,
			)
		}
		return err
	}

	if redis.Status.Phase == "" {
		rd, err := util.UpdateRedisStatus(c.ExtClient, redis, func(in *api.RedisStatus) *api.RedisStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		}, api.EnableStatusSubresource)
		if err != nil {
			if ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis); rerr == nil {
				c.recorder.Eventf(
					ref,
					core.EventTypeWarning,
					eventer.EventReasonFailedToUpdate,
					err.Error(),
				)
			}
			return err
		}
		redis.Status = rd.Status
	}

	// create Governing Service
	governingService := c.GoverningService
	if err := c.CreateGoverningService(governingService, redis.Namespace); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to create Service: "%v". Reason: %v`,
				governingService,
				err,
			)
		}
		return err
	}

	// ensure database Service
	vt1, err := c.ensureService(redis)
	if err != nil {
		return err
	}

	// ensure database StatefulSet
	vt2, err := c.ensureStatefulSet(redis)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis); rerr == nil {
			c.recorder.Event(
				ref,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully created Redis",
			)
		}
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis); rerr == nil {
			c.recorder.Event(
				ref,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully patched Redis",
			)
		}
	}

	rd, err := util.UpdateRedisStatus(c.ExtClient, redis, func(in *api.RedisStatus) *api.RedisStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = types.NewIntHash(redis.Generation, meta_util.GenerationHash(redis))
		return in
	}, api.EnableStatusSubresource)
	if err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
		}
		return err
	}
	redis.Status = rd.Status

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(redis); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				"Failed to manage monitoring system. Reason: %v",
				err,
			)
		}
		log.Errorln(err)
		return nil
	}

	if err := c.manageMonitor(redis); err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				"Failed to manage monitoring system. Reason: %v",
				err,
			)
		}
		log.Errorln(err)
		return nil
	}

	return nil
}

func (c *Controller) pause(redis *api.Redis) error {
	if _, err := c.createDormantDatabase(redis); err != nil {
		if kerr.IsAlreadyExists(err) {
			// if already exists, check if it is database of another Kind and return error in that case.
			// If the Kind is same, we can safely assume that the DormantDB was not deleted in before,
			// Probably because, User is more faster (create-delete-create-again-delete...) than operator!
			// So reuse that DormantDB!
			ddb, err := c.ExtClient.DormantDatabases(redis.Namespace).Get(redis.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if val, _ := meta_util.GetStringValue(ddb.Labels, api.LabelDatabaseKind); val != api.ResourceKindRedis {
				return fmt.Errorf(`DormantDatabase "%v" of kind %v already exists`, redis.Name, val)
			}
		} else {
			return fmt.Errorf(`failed to create DormantDatabase: "%v". Reason: %v`, redis.Name, err)
		}
	}

	if redis.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(redis); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}
