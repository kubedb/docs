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
	"kubedb.dev/apimachinery/pkg/eventer"

	"gomodules.xyz/x/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

func (c *Controller) ensureGoverningService(db *api.PgBouncer) error {
	meta := metav1.ObjectMeta{
		Name:      db.GoverningServiceName(),
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindPgBouncer))

	_, vt, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = db.OffshootLabels()

		in.Spec.Type = core.ServiceTypeClusterIP
		// create headless service
		in.Spec.ClusterIP = core.ClusterIPNone
		// create pod dns records
		in.Spec.Selector = db.OffshootSelectors()
		in.Spec.PublishNotReadyAddresses = true
		// create SRV records with pod DNS name as service provider
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
			{
				Name:       api.PgBouncerDatabasePortName,
				Port:       api.PgBouncerDatabasePort,
				TargetPort: intstr.FromString(api.PgBouncerDatabasePortName),
			},
		})

		return in
	}, metav1.PatchOptions{})
	if err == nil && (vt == kutil.VerbCreated || vt == kutil.VerbPatched) {
		c.Recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s governing service",
			vt,
		)
	}

	return err
}

func (c *Controller) ensurePrimaryService(db *api.PgBouncer) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      db.OffshootName(),
		Namespace: db.Namespace,
	}
	svcTemplate := api.GetServiceTemplate(db.Spec.ServiceTemplates, api.PrimaryServiceAlias)
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindPgBouncer))

	_, ok, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = db.OffshootLabels()
		in.Annotations = svcTemplate.Annotations

		in.Spec.Selector = db.OffshootSelectors()
		in.Spec.Ports = ofst.PatchServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
				{
					Name:       api.PgBouncerPrimaryServicePortName,
					Port:       api.PgBouncerDatabasePort,
					TargetPort: intstr.FromString(api.PgBouncerDatabasePortName),
				},
			}),
			svcTemplate.Spec.Ports,
		)

		if svcTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = svcTemplate.Spec.ClusterIP
		}
		if svcTemplate.Spec.Type != "" {
			in.Spec.Type = svcTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = svcTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = svcTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = svcTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = svcTemplate.Spec.ExternalTrafficPolicy
		if svcTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = svcTemplate.Spec.HealthCheckNodePort
		}
		return in
	}, metav1.PatchOptions{})
	return ok, err
}

func (c *Controller) ensureStatsService(db *api.PgBouncer) (kutil.VerbType, error) {
	// return if monitoring is not prometheus
	if db.Spec.Monitor == nil || db.Spec.Monitor.Agent.Vendor() != mona.VendorPrometheus {
		log.Infoln("pgbouncer.spec.monitor.agent is not provided by prometheus.io")
		return kutil.VerbUnchanged, nil
	}

	// reconcile stats service
	meta := metav1.ObjectMeta{
		Name:      db.StatsService().ServiceName(),
		Namespace: db.Namespace,
	}
	svcTemplate := api.GetServiceTemplate(db.Spec.ServiceTemplates, api.StatsServiceAlias)
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindPgBouncer))
	_, vt, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = db.StatsServiceLabels()
		in.Annotations = svcTemplate.Annotations

		in.Spec.Selector = db.OffshootSelectors()
		in.Spec.Ports = ofst.PatchServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
				{
					Name:       mona.PrometheusExporterPortName,
					Port:       db.Spec.Monitor.Prometheus.Exporter.Port,
					TargetPort: intstr.FromString(mona.PrometheusExporterPortName),
				},
			}),
			svcTemplate.Spec.Ports,
		)
		if svcTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = svcTemplate.Spec.ClusterIP
		}
		if svcTemplate.Spec.Type != "" {
			in.Spec.Type = svcTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = svcTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = svcTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = svcTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = svcTemplate.Spec.ExternalTrafficPolicy
		if svcTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = svcTemplate.Spec.HealthCheckNodePort
		}

		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s stats service",
			vt,
		)
	}
	return vt, nil
}
