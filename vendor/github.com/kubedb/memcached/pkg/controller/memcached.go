package controller

import (
	"fmt"

	"github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/log"
	"github.com/appscode/kutil"
	meta_util "github.com/appscode/kutil/meta"
	"github.com/kubedb/apimachinery/apis"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	validator "github.com/kubedb/memcached/pkg/admission"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) create(memcached *api.Memcached) error {
	if err := validator.ValidateMemcached(c.Client, c.ExtClient, memcached, true); err != nil {
		c.recorder.Event(
			memcached,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		return nil // user error so just record error and don't retry.
	}

	// Delete Matching DormantDatabase if exists any
	if err := c.deleteMatchingDormantDatabase(memcached); err != nil {
		c.recorder.Eventf(
			memcached,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to delete dormant Database : "%v". Reason: %v`,
			memcached.Name,
			err,
		)
		return err
	}

	if memcached.Status.Phase == "" {
		mc, err := util.UpdateMemcachedStatus(c.ExtClient.KubedbV1alpha1(), memcached, func(in *api.MemcachedStatus) *api.MemcachedStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		}, apis.EnableStatusSubresource)
		if err != nil {
			c.recorder.Eventf(
				memcached,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return err
		}
		memcached.Status = mc.Status
	}

	// ensure database Service
	vt1, err := c.ensureService(memcached)
	if err != nil {
		return err
	}

	// ensure database Deployment
	vt2, err := c.ensureDeployment(memcached)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.recorder.Event(
			memcached,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created Memcached",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			memcached,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched Memcached",
		)
	}

	mc, err := util.UpdateMemcachedStatus(c.ExtClient.KubedbV1alpha1(), memcached, func(in *api.MemcachedStatus) *api.MemcachedStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = types.NewIntHash(memcached.Generation, meta_util.GenerationHash(memcached))
		return in
	}, apis.EnableStatusSubresource)
	if err != nil {
		c.recorder.Eventf(
			memcached,
			core.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			err.Error(),
		)
		return err
	}
	memcached.Status = mc.Status

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(memcached); err != nil {
		c.recorder.Eventf(
			memcached,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
		return nil
	}

	if err := c.manageMonitor(memcached); err != nil {
		c.recorder.Eventf(
			memcached,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
		return nil
	}

	return nil
}

func (c *Controller) terminate(memcached *api.Memcached) error {
	// If TerminationPolicy is "terminate", keep everything (ie, PVCs,Secrets,Snapshots) intact.
	// In operator, create dormantdatabase
	if memcached.Spec.TerminationPolicy == api.TerminationPolicyPause {

		if _, err := c.createDormantDatabase(memcached); err != nil {
			if kerr.IsAlreadyExists(err) {
				// if already exists, check if it is database of another Kind and return error in that case.
				// If the Kind is same, we can safely assume that the DormantDB was not deleted in before,
				// Probably because, User is more faster (create-delete-create-again-delete...) than operator!
				// So reuse that DormantDB!
				ddb, err := c.ExtClient.KubedbV1alpha1().DormantDatabases(memcached.Namespace).Get(memcached.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if val, _ := meta_util.GetStringValue(ddb.Labels, api.LabelDatabaseKind); val != api.ResourceKindMemcached {
					return fmt.Errorf(`DormantDatabase "%v" of kind %v already exists`, memcached.Name, val)
				}
			} else {
				return fmt.Errorf(`failed to create DormantDatabase: "%v". Reason: %v`, memcached.Name, err)
			}
		}
	}
	// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots).
	// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets intact.
	// In both these cases, don't create dormantdatabase

	// At this moment, No elements of memcached to wipe out.
	// In future. if we add any secrets or other component, handle here

	if memcached.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(memcached); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}
