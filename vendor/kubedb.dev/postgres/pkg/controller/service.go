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

var (
	NodeRole = "kubedb.com/role"
)

const (
	PostgresPort     = 5432
	PostgresPortName = "api"
)

var (
	defaultDBPort = core.ServicePort{
		Name:       PostgresPortName,
		Port:       PostgresPort,
		TargetPort: intstr.FromString(PostgresPortName),
	}
)

func (c *Controller) ensureService(postgres *api.Postgres) (kutil.VerbType, error) {
	// Check if service name exists
	err := c.checkService(postgres, postgres.OffshootName())
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	// create database Service
	vt1, err := c.createService(postgres)
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt1 != kutil.VerbUnchanged {
		c.recorder.Eventf(
			postgres,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s Service",
			vt1,
		)
	}

	// Check if service name exists
	err = c.checkService(postgres, postgres.ReplicasServiceName())
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	// create database Service
	vt2, err := c.createReplicasService(postgres)
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt2 != kutil.VerbUnchanged {
		c.recorder.Eventf(
			postgres,
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

func (c *Controller) checkService(postgres *api.Postgres, name string) error {
	service, err := c.Client.CoreV1().Services(postgres.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if service.Labels[api.LabelDatabaseKind] != api.ResourceKindPostgres ||
		service.Labels[api.LabelDatabaseName] != postgres.Name {
		return fmt.Errorf(`intended service "%v/%v" already exists`, postgres.Namespace, name)
	}

	return nil
}

func (c *Controller) createService(postgres *api.Postgres) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      postgres.OffshootName(),
		Namespace: postgres.Namespace,
	}

	owner := metav1.NewControllerRef(postgres, api.SchemeGroupVersion.WithKind(api.ResourceKindPostgres))

	_, ok, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = postgres.OffshootLabels()
		in.Annotations = postgres.Spec.ServiceTemplate.Annotations

		in.Spec.Selector = postgres.OffshootSelectors()
		in.Spec.Selector[NodeRole] = "primary"
		in.Spec.Ports = upsertServicePort(in, postgres)

		if postgres.Spec.ServiceTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = postgres.Spec.ServiceTemplate.Spec.ClusterIP
		}
		if postgres.Spec.ServiceTemplate.Spec.Type != "" {
			in.Spec.Type = postgres.Spec.ServiceTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = postgres.Spec.ServiceTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = postgres.Spec.ServiceTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = postgres.Spec.ServiceTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = postgres.Spec.ServiceTemplate.Spec.ExternalTrafficPolicy
		if postgres.Spec.ServiceTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = postgres.Spec.ServiceTemplate.Spec.HealthCheckNodePort
		}
		return in
	}, metav1.PatchOptions{})
	return ok, err
}

func upsertServicePort(in *core.Service, postgres *api.Postgres) []core.ServicePort {
	return ofst.MergeServicePorts(
		core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{defaultDBPort}),
		postgres.Spec.ServiceTemplate.Spec.Ports,
	)
}

func (c *Controller) createReplicasService(postgres *api.Postgres) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      postgres.ReplicasServiceName(),
		Namespace: postgres.Namespace,
	}

	owner := metav1.NewControllerRef(postgres, api.SchemeGroupVersion.WithKind(api.ResourceKindPostgres))

	_, ok, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = postgres.OffshootLabels()
		in.Annotations = postgres.Spec.ReplicaServiceTemplate.Annotations

		in.Spec.Selector = postgres.OffshootSelectors()
		in.Spec.Selector[NodeRole] = "replica"
		in.Spec.Ports = upsertReplicaServicePort(in, postgres)

		if postgres.Spec.ReplicaServiceTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = postgres.Spec.ReplicaServiceTemplate.Spec.ClusterIP
		}
		if postgres.Spec.ReplicaServiceTemplate.Spec.Type != "" {
			in.Spec.Type = postgres.Spec.ReplicaServiceTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = postgres.Spec.ReplicaServiceTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = postgres.Spec.ReplicaServiceTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = postgres.Spec.ReplicaServiceTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = postgres.Spec.ReplicaServiceTemplate.Spec.ExternalTrafficPolicy
		if postgres.Spec.ReplicaServiceTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = postgres.Spec.ReplicaServiceTemplate.Spec.HealthCheckNodePort
		}
		return in
	}, metav1.PatchOptions{})
	return ok, err
}

func upsertReplicaServicePort(in *core.Service, postgres *api.Postgres) []core.ServicePort {
	return ofst.MergeServicePorts(
		core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{defaultDBPort}),
		postgres.Spec.ReplicaServiceTemplate.Spec.Ports,
	)
}

func (c *Controller) ensureStatsService(postgres *api.Postgres) (kutil.VerbType, error) {
	// return if monitoring is not prometheus
	if postgres.GetMonitoringVendor() != mona.VendorPrometheus {
		log.Infoln("postgres.spec.monitor.agent is not coreos-operator or builtin.")
		return kutil.VerbUnchanged, nil
	}

	// Check if statsService name exists
	if err := c.checkService(postgres, postgres.StatsService().ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	owner := metav1.NewControllerRef(postgres, api.SchemeGroupVersion.WithKind(api.ResourceKindPostgres))

	// reconcile stats service
	meta := metav1.ObjectMeta{
		Name:      postgres.StatsService().ServiceName(),
		Namespace: postgres.Namespace,
	}
	_, vt, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = postgres.StatsServiceLabels()
		in.Spec.Selector = postgres.OffshootSelectors()
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
			{
				Name:       api.PrometheusExporterPortName,
				Protocol:   core.ProtocolTCP,
				Port:       postgres.Spec.Monitor.Prometheus.Exporter.Port,
				TargetPort: intstr.FromString(api.PrometheusExporterPortName),
			},
		})
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			postgres,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s stats service",
			vt,
		)
	}
	return vt, nil
}
