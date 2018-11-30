package controller

import (
	"fmt"

	"github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/log"
	"github.com/appscode/kutil"
	dynamic_util "github.com/appscode/kutil/dynamic"
	meta_util "github.com/appscode/kutil/meta"
	"github.com/kubedb/apimachinery/apis"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	validator "github.com/kubedb/redis/pkg/admission"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
)

func (c *Controller) create(redis *api.Redis) error {
	if err := validator.ValidateRedis(c.Client, c.ExtClient, redis, true); err != nil {
		c.recorder.Event(
			redis,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		return nil // user error so just record error and don't retry.
	}

	// Delete Matching DormantDatabase if exists any
	if err := c.deleteMatchingDormantDatabase(redis); err != nil {
		return fmt.Errorf(`failed to delete dormant Database : "%v/%v". Reason: %v`, redis.Namespace, redis.Name, err)
	}

	if redis.Status.Phase == "" {
		rd, err := util.UpdateRedisStatus(c.ExtClient.KubedbV1alpha1(), redis, func(in *api.RedisStatus) *api.RedisStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		}, apis.EnableStatusSubresource)
		if err != nil {
			return err
		}
		redis.Status = rd.Status
	}

	// create Governing Service
	governingService := c.GoverningService
	if err := c.CreateGoverningService(governingService, redis.Namespace); err != nil {
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
		c.recorder.Event(
			redis,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created Redis",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			redis,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched Redis",
		)
	}

	rd, err := util.UpdateRedisStatus(c.ExtClient.KubedbV1alpha1(), redis, func(in *api.RedisStatus) *api.RedisStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = types.NewIntHash(redis.Generation, meta_util.GenerationHash(redis))
		return in
	}, apis.EnableStatusSubresource)
	if err != nil {
		c.recorder.Eventf(
			redis,
			core.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			err.Error(),
		)
		return err
	}
	redis.Status = rd.Status

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(redis); err != nil {
		c.recorder.Eventf(
			redis,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorf("failed to manage monitoring system. Reason: %v", err)
		return nil
	}

	if err := c.manageMonitor(redis); err != nil {
		c.recorder.Eventf(
			redis,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorf("failed to manage monitoring system. Reason: %v", err)
		return nil
	}

	_, err = c.ensureAppBinding(redis)
	if err != nil {
		log.Errorln(err)
		return err
	}
	return nil
}

func (c *Controller) terminate(redis *api.Redis) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis)
	if rerr != nil {
		return rerr
	}

	// If TerminationPolicy is "terminate", keep everything (ie, PVCs,Secrets,Snapshots) intact.
	// In operator, create dormantdatabase
	if redis.Spec.TerminationPolicy == api.TerminationPolicyPause {
		if err := c.removeOwnerReferenceFromOffshoots(redis, ref); err != nil {
			return err
		}

		if _, err := c.createDormantDatabase(redis); err != nil {
			if kerr.IsAlreadyExists(err) {
				// if already exists, check if it is database of another Kind and return error in that case.
				// If the Kind is same, we can safely assume that the DormantDB was not deleted in before,
				// Probably because, User is more faster (create-delete-create-again-delete...) than operator!
				// So reuse that DormantDB!
				ddb, err := c.ExtClient.KubedbV1alpha1().DormantDatabases(redis.Namespace).Get(redis.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if val, _ := meta_util.GetStringValue(ddb.Labels, api.LabelDatabaseKind); val != api.ResourceKindRedis {
					return fmt.Errorf(`DormantDatabase "%v/%v" of kind %v already exists`, redis.Namespace, redis.Name, val)
				}
			} else {
				return fmt.Errorf(`failed to create DormantDatabase: "%v/%v". Reason: %v`, redis.Namespace, redis.Name, err)
			}
		}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshoots(redis, ref); err != nil {
			return err
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

func (c *Controller) setOwnerReferenceToOffshoots(redis *api.Redis, ref *core.ObjectReference) error {
	selector := labels.SelectorFromSet(redis.OffshootSelectors())

	// delete PVC for both "wipeOut" and "delete" TerminationPolicy.
	return dynamic_util.EnsureOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		redis.Namespace,
		selector,
		ref)
}

func (c *Controller) removeOwnerReferenceFromOffshoots(redis *api.Redis, ref *core.ObjectReference) error {
	// First, Get LabelSelector for Other Components
	labelSelector := labels.SelectorFromSet(redis.OffshootSelectors())

	return dynamic_util.RemoveOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		redis.Namespace,
		labelSelector,
		ref)
}
