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
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

const (
	PgBouncerPortName = "api"
)

func (c *Controller) ensureService(pgbouncer *api.PgBouncer) (kutil.VerbType, error) {
	// Check if service name exists
	err := c.checkService(pgbouncer, pgbouncer.OffshootName())
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	// create database Service
	vt1, err := c.createOrPatchService(pgbouncer)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	if vt1 != kutil.VerbUnchanged {
		c.recorder.Eventf(
			pgbouncer,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s Service",
			vt1,
		)
	}

	return vt1, nil
}

func (c *Controller) checkService(pgbouncer *api.PgBouncer, name string) error {
	//returns error if Service already exists
	service, err := c.Client.CoreV1().Services(pgbouncer.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if service.Labels[api.LabelDatabaseKind] != api.ResourceKindPgBouncer ||
		service.Labels[api.LabelDatabaseName] != pgbouncer.Name {
		return fmt.Errorf(`intended service "%v/%v" already exists`, pgbouncer.Namespace, name)
	}

	return nil
}

func (c *Controller) createOrPatchService(pgbouncer *api.PgBouncer) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      pgbouncer.OffshootName(),
		Namespace: pgbouncer.Namespace,
	}

	_, ok, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		ref := metav1.NewControllerRef(pgbouncer, api.SchemeGroupVersion.WithKind(api.ResourceKindPgBouncer))
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = pgbouncer.OffshootLabels()

		in.Spec.Selector = pgbouncer.OffshootSelectors()
		in.Spec.Ports = upsertServicePort(in, pgbouncer)

		if pgbouncer.Spec.ServiceTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = pgbouncer.Spec.ServiceTemplate.Spec.ClusterIP
		}
		if pgbouncer.Spec.ServiceTemplate.Spec.Type != "" {
			in.Spec.Type = pgbouncer.Spec.ServiceTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = pgbouncer.Spec.ServiceTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = pgbouncer.Spec.ServiceTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = pgbouncer.Spec.ServiceTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = pgbouncer.Spec.ServiceTemplate.Spec.ExternalTrafficPolicy
		if pgbouncer.Spec.ServiceTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = pgbouncer.Spec.ServiceTemplate.Spec.HealthCheckNodePort
		}
		return in
	}, metav1.PatchOptions{})
	return ok, err
}

func upsertServicePort(in *core.Service, pgbouncer *api.PgBouncer) []core.ServicePort {
	if pgbouncer.Spec.ConnectionPool == nil {
		return ofst.MergeServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{}),
			pgbouncer.Spec.ServiceTemplate.Spec.Ports,
		)
	}
	defaultDBPort := core.ServicePort{
		Name:       PgBouncerPortName,
		Port:       *pgbouncer.Spec.ConnectionPool.Port,
		TargetPort: intstr.FromString(PgBouncerPortName),
	}
	return ofst.MergeServicePorts(
		core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{defaultDBPort}),
		pgbouncer.Spec.ServiceTemplate.Spec.Ports,
	)
}

func (c *Controller) ensureStatsService(pgbouncer *api.PgBouncer) (kutil.VerbType, error) {
	// return if monitoring is not prometheus
	if pgbouncer.GetMonitoringVendor() != mona.VendorPrometheus {
		log.Infoln("pgbouncer.spec.monitor.agent is not coreos-operator or builtin.")
		return kutil.VerbUnchanged, nil
	}

	// Check if statsService name exists
	if err := c.checkService(pgbouncer, pgbouncer.StatsService().ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	// reconcile stats service
	meta := metav1.ObjectMeta{
		Name:      pgbouncer.StatsService().ServiceName(),
		Namespace: pgbouncer.Namespace,
	}
	_, vt, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		ref := metav1.NewControllerRef(pgbouncer, api.SchemeGroupVersion.WithKind(api.ResourceKindPgBouncer))
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = pgbouncer.StatsServiceLabels()
		in.Spec.Selector = pgbouncer.OffshootSelectors()
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
			{
				Name:       api.PrometheusExporterPortName,
				Protocol:   core.ProtocolTCP,
				Port:       pgbouncer.Spec.Monitor.Prometheus.Exporter.Port,
				TargetPort: intstr.FromString(api.PrometheusExporterPortName),
			},
		})
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			pgbouncer,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s stats service",
			vt,
		)
	}
	return vt, nil
}

func (c *Controller) PgBouncerForService(s *core.Service) (*api.PgBouncer, error) {
	pgbouncers, err := c.pbLister.PgBouncers(s.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, pgbouncer := range pgbouncers {
		if metav1.IsControlledBy(s, pgbouncer) {
			return pgbouncer, nil
		}
	}

	return nil, nil
}
