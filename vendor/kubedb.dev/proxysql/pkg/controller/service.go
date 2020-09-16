/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

var defaultDBPort = core.ServicePort{
	Name:       "mysql",
	Protocol:   core.ProtocolTCP,
	Port:       api.ProxySQLMySQLNodePort,
	TargetPort: intstr.FromInt(api.ProxySQLMySQLNodePort),
}

func (c *Controller) ensureService(proxysql *api.ProxySQL) (kutil.VerbType, error) {
	// Check if service name exists
	if err := c.checkService(proxysql, proxysql.ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	// create database Service
	vt, err := c.createService(proxysql)
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			proxysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s Service",
			vt,
		)
	}
	return vt, nil
}

func (c *Controller) checkService(proxysql *api.ProxySQL, serviceName string) error {
	service, err := c.Client.CoreV1().Services(proxysql.Namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if service.Labels[api.LabelDatabaseKind] != api.ResourceKindProxySQL ||
		service.Labels[api.LabelProxySQLName] != proxysql.Name {
		return fmt.Errorf(`intended service "%v/%v" already exists`, proxysql.Namespace, serviceName)
	}

	return nil
}

func (c *Controller) createService(proxysql *api.ProxySQL) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      proxysql.OffshootName(),
		Namespace: proxysql.Namespace,
	}

	owner := metav1.NewControllerRef(proxysql, api.SchemeGroupVersion.WithKind(api.ResourceKindProxySQL))

	_, ok, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = proxysql.OffshootLabels()
		in.Annotations = proxysql.Spec.ServiceTemplate.Annotations

		in.Spec.Selector = proxysql.OffshootSelectors()
		in.Spec.Ports = ofst.MergeServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{defaultDBPort}),
			proxysql.Spec.ServiceTemplate.Spec.Ports,
		)

		if proxysql.Spec.ServiceTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = proxysql.Spec.ServiceTemplate.Spec.ClusterIP
		}
		if proxysql.Spec.ServiceTemplate.Spec.Type != "" {
			in.Spec.Type = proxysql.Spec.ServiceTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = proxysql.Spec.ServiceTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = proxysql.Spec.ServiceTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = proxysql.Spec.ServiceTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = proxysql.Spec.ServiceTemplate.Spec.ExternalTrafficPolicy
		if proxysql.Spec.ServiceTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = proxysql.Spec.ServiceTemplate.Spec.HealthCheckNodePort
		}
		return in
	}, metav1.PatchOptions{})
	return ok, err
}

func (c *Controller) ensureStatsService(proxysql *api.ProxySQL) (kutil.VerbType, error) {
	// return if monitoring is not prometheus
	if proxysql.GetMonitoringVendor() != mona.VendorPrometheus {
		log.Infoln("spec.monitor.agent is not coreos-operator or builtin.")
		return kutil.VerbUnchanged, nil
	}

	// Check if statsService name exists
	if err := c.checkService(proxysql, proxysql.StatsService().ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	owner := metav1.NewControllerRef(proxysql, api.SchemeGroupVersion.WithKind(api.ResourceKindProxySQL))

	// reconcile stats Service
	meta := metav1.ObjectMeta{
		Name:      proxysql.StatsService().ServiceName(),
		Namespace: proxysql.Namespace,
	}
	_, vt, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = proxysql.StatsServiceLabels()
		in.Spec.Selector = proxysql.OffshootSelectors()
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
			{
				Name:       api.PrometheusExporterPortName,
				Protocol:   core.ProtocolTCP,
				Port:       proxysql.Spec.Monitor.Prometheus.Port,
				TargetPort: intstr.FromString(api.PrometheusExporterPortName),
			},
		})
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			proxysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s stats service",
			vt,
		)
	}
	return vt, nil
}
