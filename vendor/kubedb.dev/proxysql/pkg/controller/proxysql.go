/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/proxysql/pkg/admission"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	dynamic_util "kmodules.xyz/client-go/dynamic"
)

func (c *Controller) create(db *api.ProxySQL) error {
	if err := validator.ValidateProxySQL(c.Client, c.DBClient, db, true); err != nil {
		c.recorder.Event(
			db,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		klog.Errorln(err)
		return nil
	}

	if db.Status.Phase == "" {
		proxysqlUpd, err := util.UpdateProxySQLStatus(
			context.TODO(),
			c.DBClient.KubedbV1alpha2(),
			db.ObjectMeta,
			func(in *api.ProxySQLStatus) (types.UID, *api.ProxySQLStatus) {
				in.Phase = api.DatabasePhaseProvisioning
				return db.UID, in
			},
			metav1.UpdateOptions{},
		)
		if err != nil {
			return err
		}
		db.Status = proxysqlUpd.Status
	}

	// ensure Governing Service
	if err := c.ensureGoverningService(db); err != nil {
		return fmt.Errorf(`failed to create governing Service for : "%v/%v". Reason: %v`, db.Namespace, db.Name, err)
	}

	// Ensure ClusterRoles for statefulsets
	if err := c.ensureRBACStuff(db); err != nil {
		return err
	}

	// ensure database Service
	vt1, err := c.ensureService(db)
	if err != nil {
		return err
	}

	if err := c.ensureAuthSecret(db); err != nil {
		return err
	}

	// ensure proxysql StatefulSet
	vt2, err := c.ensureProxySQLNode(db)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.recorder.Event(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created ProxySQL",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched ProxySQL",
		)
	}

	proxysqlUpd, err := util.UpdateProxySQLStatus(
		context.TODO(),
		c.DBClient.KubedbV1alpha2(),
		db.ObjectMeta,
		func(in *api.ProxySQLStatus) (types.UID, *api.ProxySQLStatus) {
			in.Phase = api.DatabasePhaseReady
			in.ObservedGeneration = db.Generation
			return db.UID, in
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}
	db.Status = proxysqlUpd.Status

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(db); err != nil {
		c.recorder.Eventf(
			db,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		klog.Errorln(err)
		return nil
	}

	if err := c.manageMonitor(db); err != nil {
		c.recorder.Eventf(
			db,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		klog.Errorln(err)
		return nil
	}

	return nil
}

func (c *Controller) terminate(db *api.ProxySQL) error {
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindProxySQL))

	// delete PVC
	selector := labels.SelectorFromSet(db.OffshootSelectors())
	if err := dynamic_util.EnsureOwnerReferenceForSelector(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		db.Namespace,
		selector,
		owner); err != nil {
		return err
	}

	if db.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(db); err != nil {
			klog.Errorln(err)
			return nil
		}
	}

	return nil
}
