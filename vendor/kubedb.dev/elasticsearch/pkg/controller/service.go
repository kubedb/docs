/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controller

import (
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

var (
	NodeRoleMaster = "node.role.master"
	NodeRoleClient = "node.role.client"
	NodeRoleData   = "node.role.data"
	NodeRoleSet    = "set"

	defaultClientPort = core.ServicePort{
		Name:       api.ElasticsearchRestPortName,
		Port:       api.ElasticsearchRestPort,
		TargetPort: intstr.FromString(api.ElasticsearchRestPortName),
	}
	defaultPeerPort = core.ServicePort{
		Name:       api.ElasticsearchNodePortName,
		Port:       api.ElasticsearchNodePort,
		TargetPort: intstr.FromString(api.ElasticsearchNodePortName),
	}
)

func (c *Controller) ensureService(elasticsearch *api.Elasticsearch) (kutil.VerbType, error) {
	// Check if service name exists
	err := c.checkService(elasticsearch, elasticsearch.OffshootName())
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	// create database Service
	vt1, err := c.createService(elasticsearch)
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt1 != kutil.VerbUnchanged {
		c.recorder.Eventf(
			elasticsearch,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s Service",
			vt1,
		)
	}

	// Check if service name exists
	err = c.checkService(elasticsearch, elasticsearch.MasterServiceName())
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	// create database Service
	vt2, err := c.createMasterService(elasticsearch)
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt2 != kutil.VerbUnchanged {
		c.recorder.Eventf(
			elasticsearch,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s Service",
			vt2,
		)
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		return kutil.VerbCreated, nil
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		return kutil.VerbPatched, nil
	}

	return kutil.VerbUnchanged, nil
}

func (c *Controller) checkService(elasticsearch *api.Elasticsearch, name string) error {
	service, err := c.Client.CoreV1().Services(elasticsearch.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if service.Labels[api.LabelDatabaseKind] != api.ResourceKindElasticsearch ||
		service.Labels[api.LabelDatabaseName] != elasticsearch.Name {
		return fmt.Errorf(`intended service "%v/%v" already exists`, elasticsearch.Namespace, name)
	}

	return nil
}

func (c *Controller) createService(elasticsearch *api.Elasticsearch) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      elasticsearch.OffshootName(),
		Namespace: elasticsearch.Namespace,
	}

	owner := metav1.NewControllerRef(elasticsearch, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))

	_, ok, err := core_util.CreateOrPatchService(c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = elasticsearch.OffshootLabels()
		in.Annotations = elasticsearch.Spec.ServiceTemplate.Annotations

		in.Spec.Selector = elasticsearch.OffshootSelectors()
		in.Spec.Selector[NodeRoleClient] = NodeRoleSet
		in.Spec.Ports = ofst.MergeServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{defaultClientPort}),
			elasticsearch.Spec.ServiceTemplate.Spec.Ports,
		)

		if elasticsearch.Spec.ServiceTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = elasticsearch.Spec.ServiceTemplate.Spec.ClusterIP
		}
		if elasticsearch.Spec.ServiceTemplate.Spec.Type != "" {
			in.Spec.Type = elasticsearch.Spec.ServiceTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = elasticsearch.Spec.ServiceTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = elasticsearch.Spec.ServiceTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = elasticsearch.Spec.ServiceTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = elasticsearch.Spec.ServiceTemplate.Spec.ExternalTrafficPolicy
		if elasticsearch.Spec.ServiceTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = elasticsearch.Spec.ServiceTemplate.Spec.HealthCheckNodePort
		}
		return in
	})
	return ok, err
}

func (c *Controller) createMasterService(elasticsearch *api.Elasticsearch) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      elasticsearch.MasterServiceName(),
		Namespace: elasticsearch.Namespace,
	}

	owner := metav1.NewControllerRef(elasticsearch, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))

	_, ok, err := core_util.CreateOrPatchService(c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = elasticsearch.OffshootLabels()
		in.Annotations = elasticsearch.Spec.ServiceTemplate.Annotations

		in.Spec.Selector = elasticsearch.OffshootSelectors()
		in.Spec.Selector[NodeRoleMaster] = NodeRoleSet
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{defaultPeerPort})
		return in
	})
	return ok, err
}

func (c *Controller) ensureStatsService(elasticsearch *api.Elasticsearch) (kutil.VerbType, error) {
	// return if monitoring is not prometheus
	if elasticsearch.GetMonitoringVendor() != mona.VendorPrometheus {
		log.Infoln("spec.monitor.agent is not coreos-operator or builtin.")
		return kutil.VerbUnchanged, nil
	}

	// Check if statsService name exists
	if err := c.checkService(elasticsearch, elasticsearch.StatsService().ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	owner := metav1.NewControllerRef(elasticsearch, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))

	// reconcile statsService
	meta := metav1.ObjectMeta{
		Name:      elasticsearch.StatsService().ServiceName(),
		Namespace: elasticsearch.Namespace,
	}
	_, vt, err := core_util.CreateOrPatchService(c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = elasticsearch.StatsServiceLabels()
		in.Spec.Selector = elasticsearch.OffshootSelectors()
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
			{
				Name:       api.PrometheusExporterPortName,
				Protocol:   core.ProtocolTCP,
				Port:       elasticsearch.Spec.Monitor.Prometheus.Port,
				TargetPort: intstr.FromString(api.PrometheusExporterPortName),
			},
		})
		return in
	})
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			elasticsearch,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s stats service",
			vt,
		)
	}
	return vt, nil
}
