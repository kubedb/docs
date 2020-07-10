/*
Copyright AppsCode Inc. and Contributors

Licensed under the PolyForm Noncommercial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/PolyForm-Noncommercial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"errors"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/memcached/pkg/admission"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kutil "kmodules.xyz/client-go"
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

	if memcached.Status.Phase == "" {
		mc, err := util.UpdateMemcachedStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), memcached.ObjectMeta, func(in *api.MemcachedStatus) *api.MemcachedStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		}, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		memcached.Status = mc.Status
	}

	// Ensure ClusterRoles for deployments
	if err := c.ensureRBACStuff(memcached); err != nil {
		return err
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

	mc, err := util.UpdateMemcachedStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), memcached.ObjectMeta, func(in *api.MemcachedStatus) *api.MemcachedStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = memcached.Generation
		return in
	}, metav1.UpdateOptions{})
	if err != nil {
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
		log.Errorf("failed to manage monitoring system. Reason: %v", err)
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
		log.Errorf("failed to manage monitoring system. Reason: %v", err)
		return nil
	}

	_, err = c.ensureAppBinding(memcached)
	if err != nil {
		log.Errorln(err)
		return err
	}
	return nil
}

func (c *Controller) halt(db *api.Memcached) error {
	if db.Spec.Halted && db.Spec.TerminationPolicy != api.TerminationPolicyHalt {
		return errors.New("can't halt db. 'spec.terminationPolicy' is not 'Halt'")
	}
	log.Infof("Halting Memcached %v/%v", db.Namespace, db.Name)
	if err := c.haltDatabase(db); err != nil {
		return err
	}
	if err := c.waitUntilPaused(db); err != nil {
		return err
	}
	log.Infof("update status of Memcached %v/%v to Halted.", db.Namespace, db.Name)
	if _, err := util.UpdateMemcachedStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), db.ObjectMeta, func(in *api.MemcachedStatus) *api.MemcachedStatus {
		in.Phase = api.DatabasePhaseHalted
		in.ObservedGeneration = db.Generation
		return in
	}, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *Controller) terminate(memcached *api.Memcached) error {
	// If TerminationPolicy is "terminate", keep everything (ie, PVCs,Secrets,Snapshots) intact

	// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots).
	// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets intact.
	// In both these cases, don't create dormantdatabase

	// At this moment, No elements of memcached to wipe out.
	// In future. if we add any secrets or other component, handle here

	if memcached.Spec.Monitor != nil {
		if err := c.deleteMonitor(memcached); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}
