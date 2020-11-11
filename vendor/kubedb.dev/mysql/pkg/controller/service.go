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

	"gomodules.xyz/x/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

func (c *Controller) ensureGoverningService(db *api.MySQL) error {
	meta := metav1.ObjectMeta{
		Name:      db.GoverningServiceName(),
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

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
				Name:       api.MySQLDatabasePortName,
				Port:       api.MySQLDatabasePort,
				TargetPort: intstr.FromString(api.MySQLDatabasePortName),
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

func (c *Controller) ensureService(db *api.MySQL) error {
	// 1. ensure primary service:
	// using primary service, user always have both read and write operation permission.
	// for MySQL standalone, it will be used for selecting standalone pod
	// and for the group replication, it will be used for selecting primary pod.
	// Check if service name exists with different db kind, name
	// then create/patch the service
	vt, err := c.ensurePrimaryService(db)
	if err != nil {
		return err
	}
	if vt == kutil.VerbCreated {
		c.Recorder.Event(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created service for primary/standalone",
		)
	} else if vt == kutil.VerbPatched {
		c.Recorder.Event(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched service for primary/standalone replica",
		)
	}

	// 2. ensure secondary service:
	// using secondary service, user only have the read operation permission.
	// it will be used only for selecting the replicas/secondaries
	// Check if service name exists with different db kind, name
	// the create/patch the service
	if db.UsesGroupReplication() {
		vt, err := c.ensureStandbyService(db)
		if err != nil {
			return err
		}

		if vt == kutil.VerbCreated {
			c.Recorder.Event(
				db,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully created service for secondary replicas",
			)
		} else if vt == kutil.VerbPatched {
			c.Recorder.Event(
				db,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully patched service for secondary replicas",
			)
		}
	}

	return err
}

func (c *Controller) ensurePrimaryService(db *api.MySQL) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      db.ServiceName(),
		Namespace: db.Namespace,
	}
	svcTemplate := api.GetServiceTemplate(db.Spec.ServiceTemplates, api.PrimaryServiceAlias)
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

	_, vt, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = db.OffshootLabels()
		in.Annotations = svcTemplate.Annotations

		in.Spec.Selector = db.OffshootSelectors()
		//add extra selector to select only primary pod for group replication
		if db.UsesGroupReplication() {
			in.Spec.Selector[api.LabelRole] = api.DatabasePodPrimary
		}

		in.Spec.Ports = ofst.PatchServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
				{
					Name:       api.MySQLPrimaryServicePortName,
					Port:       api.MySQLDatabasePort,
					TargetPort: intstr.FromString(api.MySQLDatabasePortName),
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

	return vt, err
}

func (c *Controller) ensureStandbyService(db *api.MySQL) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      db.StandbyServiceName(),
		Namespace: db.Namespace,
	}
	svcTemplate := api.GetServiceTemplate(db.Spec.ServiceTemplates, api.StandbyServiceAlias)
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

	_, vt, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = db.OffshootLabels()
		in.Annotations = svcTemplate.Annotations
		in.Spec.Selector = db.OffshootSelectors()
		//add extra selector to select only secondary pod
		in.Spec.Selector[api.LabelRole] = api.DatabasePodStandby

		in.Spec.Ports = ofst.PatchServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
				{
					Name:       api.MySQLStandbyServicePortName,
					Port:       api.MySQLDatabasePort,
					TargetPort: intstr.FromString(api.MySQLDatabasePortName),
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
	return vt, err
}

func (c *Controller) ensureStatsService(db *api.MySQL) (kutil.VerbType, error) {
	// return if monitoring is not prometheus
	if db.Spec.Monitor == nil || db.Spec.Monitor.Agent.Vendor() != mona.VendorPrometheus {
		log.Infoln("spec.monitor.agent is not provided by prometheus.io")
		return kutil.VerbUnchanged, nil
	}

	// stats Service
	meta := metav1.ObjectMeta{
		Name:      db.StatsService().ServiceName(),
		Namespace: db.Namespace,
	}
	svcTemplate := api.GetServiceTemplate(db.Spec.ServiceTemplates, api.StatsServiceAlias)
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

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
		c.Recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s stats service",
			vt,
		)
	}
	return vt, nil
}
