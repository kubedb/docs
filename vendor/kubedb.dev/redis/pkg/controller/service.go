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
	Name:       "db",
	Protocol:   core.ProtocolTCP,
	Port:       6379,
	TargetPort: intstr.FromString("db"),
}

func (c *Controller) ensureService(redis *api.Redis) (kutil.VerbType, error) {
	// Check if service name exists
	if err := c.checkService(redis, redis.ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	// create database Service
	vt, err := c.createService(redis)
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			redis,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s Service",
			vt,
		)
	}
	return vt, nil
}

func (c *Controller) checkService(redis *api.Redis, serviceName string) error {
	service, err := c.Client.CoreV1().Services(redis.Namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if service.Labels[api.LabelDatabaseKind] != api.ResourceKindRedis ||
		service.Labels[api.LabelDatabaseName] != redis.Name {
		return fmt.Errorf(`intended service "%v/%v" already exists`, redis.Namespace, serviceName)
	}

	return nil
}

func (c *Controller) createService(redis *api.Redis) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      redis.OffshootName(),
		Namespace: redis.Namespace,
	}

	owner := metav1.NewControllerRef(redis, api.SchemeGroupVersion.WithKind(api.ResourceKindRedis))

	_, ok, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = redis.OffshootSelectors()
		in.Annotations = redis.Spec.ServiceTemplate.Annotations

		in.Spec.Selector = redis.OffshootSelectors()
		in.Spec.Ports = ofst.MergeServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{defaultDBPort}),
			redis.Spec.ServiceTemplate.Spec.Ports,
		)

		if redis.Spec.ServiceTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = redis.Spec.ServiceTemplate.Spec.ClusterIP
		}
		if redis.Spec.ServiceTemplate.Spec.Type != "" {
			in.Spec.Type = redis.Spec.ServiceTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = redis.Spec.ServiceTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = redis.Spec.ServiceTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = redis.Spec.ServiceTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = redis.Spec.ServiceTemplate.Spec.ExternalTrafficPolicy
		if redis.Spec.ServiceTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = redis.Spec.ServiceTemplate.Spec.HealthCheckNodePort
		}
		return in
	}, metav1.PatchOptions{})
	return ok, err
}

func (c *Controller) ensureStatsService(redis *api.Redis) (kutil.VerbType, error) {
	// return if monitoring is not prometheus
	if redis.GetMonitoringVendor() != mona.VendorPrometheus {
		log.Infoln("spec.monitor.agent is not operator or builtin.")
		return kutil.VerbUnchanged, nil
	}

	// Check if stats Service name exists
	if err := c.checkService(redis, redis.StatsService().ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	owner := metav1.NewControllerRef(redis, api.SchemeGroupVersion.WithKind(api.ResourceCodeRedis))

	// reconcile stats Service
	meta := metav1.ObjectMeta{
		Name:      redis.StatsService().ServiceName(),
		Namespace: redis.Namespace,
	}
	_, vt, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = redis.StatsServiceLabels()
		in.Spec.Selector = redis.OffshootSelectors()
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
			{
				Name:       api.PrometheusExporterPortName,
				Protocol:   core.ProtocolTCP,
				Port:       redis.Spec.Monitor.Prometheus.Exporter.Port,
				TargetPort: intstr.FromString(api.PrometheusExporterPortName),
			},
		})
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			redis,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s stats service",
			vt,
		)
	}
	return vt, nil
}
