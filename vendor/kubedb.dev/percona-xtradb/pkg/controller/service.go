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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
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
	Port:       api.MySQLNodePort,
	TargetPort: intstr.FromInt(api.MySQLNodePort),
}

func (c *Controller) ensureService(px *api.PerconaXtraDB) (kutil.VerbType, error) {
	// Check if service name exists
	if err := c.checkService(px, px.ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	// create database Service
	vt, err := c.createService(px)
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.Recorder.Eventf(
			px,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s Service",
			vt,
		)
	}
	return vt, nil
}

func (c *Controller) checkService(px *api.PerconaXtraDB, serviceName string) error {
	service, err := c.Client.CoreV1().Services(px.Namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if service.Labels[api.LabelDatabaseKind] != api.ResourceKindPerconaXtraDB ||
		service.Labels[api.LabelDatabaseName] != px.Name {
		return fmt.Errorf(`intended service "%v/%v" already exists`, px.Namespace, serviceName)
	}

	return nil
}

func (c *Controller) createService(px *api.PerconaXtraDB) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      px.OffshootName(),
		Namespace: px.Namespace,
	}

	owner := metav1.NewControllerRef(px, api.SchemeGroupVersion.WithKind(api.ResourceKindPerconaXtraDB))

	_, ok, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = px.OffshootLabels()
		in.Annotations = px.Spec.ServiceTemplate.Annotations

		in.Spec.Selector = px.OffshootSelectors()
		in.Spec.Ports = ofst.MergeServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{defaultDBPort}),
			px.Spec.ServiceTemplate.Spec.Ports,
		)

		if px.Spec.ServiceTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = px.Spec.ServiceTemplate.Spec.ClusterIP
		}
		if px.Spec.ServiceTemplate.Spec.Type != "" {
			in.Spec.Type = px.Spec.ServiceTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = px.Spec.ServiceTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = px.Spec.ServiceTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = px.Spec.ServiceTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = px.Spec.ServiceTemplate.Spec.ExternalTrafficPolicy
		if px.Spec.ServiceTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = px.Spec.ServiceTemplate.Spec.HealthCheckNodePort
		}
		return in
	}, metav1.PatchOptions{})
	return ok, err
}

func (c *Controller) ensureStatsService(px *api.PerconaXtraDB) (kutil.VerbType, error) {
	// return if monitoring is not prometheus
	if px.Spec.Monitor == nil || px.Spec.Monitor.Agent.Vendor() != mona.VendorPrometheus {
		log.Infoln("spec.monitor.agent is not provided by prometheus.io")
		return kutil.VerbUnchanged, nil
	}

	// Check if statsService name exists
	if err := c.checkService(px, px.StatsService().ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	owner := metav1.NewControllerRef(px, api.SchemeGroupVersion.WithKind(api.ResourceKindPerconaXtraDB))

	// reconcile stats Service
	meta := metav1.ObjectMeta{
		Name:      px.StatsService().ServiceName(),
		Namespace: px.Namespace,
	}
	_, vt, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = px.StatsServiceLabels()
		in.Spec.Selector = px.OffshootSelectors()
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
			{
				Name:       mona.PrometheusExporterPortName,
				Protocol:   core.ProtocolTCP,
				Port:       px.Spec.Monitor.Prometheus.Exporter.Port,
				TargetPort: intstr.FromString(mona.PrometheusExporterPortName),
			},
		})
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.Recorder.Eventf(
			px,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s stats service",
			vt,
		)
	}
	return vt, nil
}

func (c *Controller) createPerconaXtraDBGoverningService(px *api.PerconaXtraDB) (string, error) {
	owner := metav1.NewControllerRef(px, api.SchemeGroupVersion.WithKind(api.ResourceKindPerconaXtraDB))

	service := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      px.GoverningServiceName(),
			Namespace: px.Namespace,
			Labels:    px.OffshootLabels(),
			// 'tolerate-unready-endpoints' annotation is deprecated.
			// owner: https://github.com/kubernetes/kubernetes/pull/63742
			Annotations: map[string]string{
				"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
			},
		},
		Spec: core.ServiceSpec{
			Type:                     core.ServiceTypeClusterIP,
			ClusterIP:                core.ClusterIPNone,
			PublishNotReadyAddresses: true,
			Ports: []core.ServicePort{
				{
					Name: "db",
					Port: api.MySQLNodePort,
				},
			},
			Selector: px.OffshootSelectors(),
		},
	}
	core_util.EnsureOwnerReference(&service.ObjectMeta, owner)

	_, err := c.Client.CoreV1().Services(px.Namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil && !kerr.IsAlreadyExists(err) {
		return "", err
	}
	return service.Name, nil
}
