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
	Port:       3306,
	TargetPort: intstr.FromString("db"),
}

func (c *Controller) ensurePrimaryService(mysql *api.MySQL) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      mysql.OffshootName(),
		Namespace: mysql.Namespace,
	}

	owner := metav1.NewControllerRef(mysql, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

	_, vt, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = mysql.OffshootLabels()
		in.Annotations = mysql.Spec.ServiceTemplate.Annotations

		in.Spec.Selector = mysql.OffshootSelectors()
		//add extra selector to select only primary pod for group replication
		if mysql.UsesGroupReplication() {
			in.Spec.Selector[api.MySQLLabelRole] = api.MySQLPodPrimary
		}

		in.Spec.Ports = ofst.MergeServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{defaultDBPort}),
			mysql.Spec.ServiceTemplate.Spec.Ports,
		)

		if mysql.Spec.ServiceTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = mysql.Spec.ServiceTemplate.Spec.ClusterIP
		}
		if mysql.Spec.ServiceTemplate.Spec.Type != "" {
			in.Spec.Type = mysql.Spec.ServiceTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = mysql.Spec.ServiceTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = mysql.Spec.ServiceTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = mysql.Spec.ServiceTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = mysql.Spec.ServiceTemplate.Spec.ExternalTrafficPolicy
		if mysql.Spec.ServiceTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = mysql.Spec.ServiceTemplate.Spec.HealthCheckNodePort
		}
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.Recorder.Eventf(
			mysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s Service",
			vt,
		)
	}
	return vt, nil
}

func (c *Controller) ensureStatsService(mysql *api.MySQL) (kutil.VerbType, error) {
	// return if monitoring is not prometheus
	if mysql.Spec.Monitor == nil || mysql.Spec.Monitor.Agent.Vendor() != mona.VendorPrometheus {
		log.Infoln("spec.monitor.agent is not provided by prometheus.io")
		return kutil.VerbUnchanged, nil
	}

	owner := metav1.NewControllerRef(mysql, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

	// reconcile stats Service
	meta := metav1.ObjectMeta{
		Name:      mysql.StatsService().ServiceName(),
		Namespace: mysql.Namespace,
	}
	_, vt, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = mysql.StatsServiceLabels()
		in.Spec.Selector = mysql.OffshootSelectors()
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
			{
				Name:       mona.PrometheusExporterPortName,
				Protocol:   core.ProtocolTCP,
				Port:       mysql.Spec.Monitor.Prometheus.Exporter.Port,
				TargetPort: intstr.FromString(mona.PrometheusExporterPortName),
			},
		})
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.Recorder.Eventf(
			mysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s stats service",
			vt,
		)
	}
	return vt, nil
}

func (c *Controller) createMySQLGoverningService(mysql *api.MySQL) (string, error) {
	owner := metav1.NewControllerRef(mysql, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

	service := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mysql.GoverningServiceName(),
			Namespace: mysql.Namespace,
			Labels:    mysql.OffshootLabels(),
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
			Selector: mysql.OffshootSelectors(),
		},
	}
	core_util.EnsureOwnerReference(&service.ObjectMeta, owner)

	_, err := c.Client.CoreV1().Services(mysql.Namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil && !kerr.IsAlreadyExists(err) {
		return "", err
	}
	return service.Name, nil
}

func (c *Controller) ensureSecondaryService(mysql *api.MySQL) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      mysql.SecondaryServiceName(),
		Namespace: mysql.Namespace,
	}

	owner := metav1.NewControllerRef(mysql, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

	_, vt, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = mysql.OffshootLabels()
		in.Annotations = mysql.Spec.ServiceTemplate.Annotations
		in.Spec.Selector = mysql.OffshootSelectors()
		//add extra selector to select only secondary pod
		in.Spec.Selector[api.MySQLLabelRole] = api.MySQLPodSecondary

		in.Spec.Ports = ofst.MergeServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{defaultDBPort}),
			mysql.Spec.ServiceTemplate.Spec.Ports,
		)

		if mysql.Spec.ServiceTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = mysql.Spec.ServiceTemplate.Spec.ClusterIP
		}
		if mysql.Spec.ServiceTemplate.Spec.Type != "" {
			in.Spec.Type = mysql.Spec.ServiceTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = mysql.Spec.ServiceTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = mysql.Spec.ServiceTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = mysql.Spec.ServiceTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = mysql.Spec.ServiceTemplate.Spec.ExternalTrafficPolicy
		if mysql.Spec.ServiceTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = mysql.Spec.ServiceTemplate.Spec.HealthCheckNodePort
		}
		return in
	}, metav1.PatchOptions{})
	return vt, err
}
