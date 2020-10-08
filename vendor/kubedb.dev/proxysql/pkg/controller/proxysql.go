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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/proxysql/pkg/admission"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kutil "kmodules.xyz/client-go"
	dynamic_util "kmodules.xyz/client-go/dynamic"
)

func (c *Controller) create(proxysql *api.ProxySQL) error {
	if err := validator.ValidateProxySQL(c.Client, c.DBClient, proxysql, true); err != nil {
		c.recorder.Event(
			proxysql,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		return nil
	}

	if proxysql.Status.Phase == "" {
		proxysqlUpd, err := util.UpdateProxySQLStatus(
			context.TODO(),
			c.DBClient.KubedbV1alpha2(),
			proxysql.ObjectMeta,
			func(in *api.ProxySQLStatus) *api.ProxySQLStatus {
				in.Phase = api.DatabasePhaseProvisioning
				return in
			},
			metav1.UpdateOptions{},
		)
		if err != nil {
			return err
		}
		proxysql.Status = proxysqlUpd.Status
	}

	// create Governing Service
	if err := c.CreateGoverningService(c.GoverningService, proxysql.Namespace); err != nil {
		return err
	}

	// Ensure ClusterRoles for statefulsets
	if err := c.ensureRBACStuff(proxysql); err != nil {
		return err
	}

	// ensure database Service
	vt1, err := c.ensureService(proxysql)
	if err != nil {
		return err
	}

	if err := c.ensureProxySQLSecret(proxysql); err != nil {
		return err
	}

	// ensure proxysql StatefulSet
	vt2, err := c.ensureProxySQLNode(proxysql)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.recorder.Event(
			proxysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created ProxySQL",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			proxysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched ProxySQL",
		)
	}

	proxysqlUpd, err := util.UpdateProxySQLStatus(
		context.TODO(),
		c.DBClient.KubedbV1alpha2(),
		proxysql.ObjectMeta,
		func(in *api.ProxySQLStatus) *api.ProxySQLStatus {
			in.Phase = api.DatabasePhaseReady
			in.ObservedGeneration = proxysql.Generation
			return in
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}
	proxysql.Status = proxysqlUpd.Status

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(proxysql); err != nil {
		c.recorder.Eventf(
			proxysql,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
		return nil
	}

	if err := c.manageMonitor(proxysql); err != nil {
		c.recorder.Eventf(
			proxysql,
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

func (c *Controller) terminate(proxysql *api.ProxySQL) error {
	owner := metav1.NewControllerRef(proxysql, api.SchemeGroupVersion.WithKind(api.ResourceKindProxySQL))

	// delete PVC
	selector := labels.SelectorFromSet(proxysql.OffshootSelectors())
	if err := dynamic_util.EnsureOwnerReferenceForSelector(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		proxysql.Namespace,
		selector,
		owner); err != nil {
		return err
	}

	if proxysql.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(proxysql); err != nil {
			log.Errorln(err)
			return nil
		}
	}

	return nil
}
