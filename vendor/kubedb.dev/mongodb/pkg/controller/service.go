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

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

func (r *Reconciler) ensureService(db *api.MongoDB) (kutil.VerbType, error) {
	// create database Service
	vt, err := r.ensurePrimaryService(db)
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		r.Recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s Primary Service",
			vt,
		)
	}
	return vt, nil
}

func (r *Reconciler) ensurePrimaryService(db *api.MongoDB) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      db.OffshootName(),
		Namespace: db.Namespace,
	}
	svcTemplate := api.GetServiceTemplate(db.Spec.ServiceTemplates, api.PrimaryServiceAlias)
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMongoDB))

	selector := db.OffshootSelectors()
	if db.Spec.ShardTopology != nil {
		selector = db.MongosSelectors()
	}

	_, ok, err := core_util.CreateOrPatchService(
		context.TODO(),
		r.Client,
		meta,
		func(in *core.Service) *core.Service {
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
			in.Labels = db.ServiceLabels(api.PrimaryServiceAlias, svcTemplate.Labels)
			in.Annotations = svcTemplate.Annotations

			in.Spec.Selector = selector
			if db.Spec.ReplicaSet != nil {
				in.Spec.Selector[api.LabelRole] = api.DatabasePodPrimary
			}
			in.Spec.Ports = ofst.PatchServicePorts(
				core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
					{
						Name:       api.MongoDBPrimaryServicePortName,
						Port:       api.MongoDBDatabasePort,
						TargetPort: intstr.FromString(api.MongoDBDatabasePortName),
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
		},
		metav1.PatchOptions{},
	)
	return ok, err
}

func (r *Reconciler) ensureStatsService(db *api.MongoDB) (kutil.VerbType, error) {
	// return if monitoring is not prometheus
	if db.Spec.Monitor == nil || db.Spec.Monitor.Agent.Vendor() != mona.VendorPrometheus {
		klog.Infoln("spec.monitor.agent is not provided by prometheus.io")
		return kutil.VerbUnchanged, nil
	}

	// create/patch stats Service
	meta := metav1.ObjectMeta{
		Name:      db.StatsService().ServiceName(),
		Namespace: db.Namespace,
	}
	svcTemplate := api.GetServiceTemplate(db.Spec.ServiceTemplates, api.StatsServiceAlias)
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMongoDB))
	_, vt, err := core_util.CreateOrPatchService(
		context.TODO(),
		r.Client,
		meta,
		func(in *core.Service) *core.Service {
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
			in.Labels = db.StatsServiceLabels()
			in.Annotations = meta_util.OverwriteKeys(in.Annotations, svcTemplate.Annotations)

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
		},
		metav1.PatchOptions{},
	)
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		r.Recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s stats service",
			vt,
		)
	}
	return vt, nil
}

func (r *Reconciler) EnsureGoverningService(db *api.MongoDB) error {
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMongoDB))

	svcFunc := func(svcName string, labels, selectors map[string]string) error {

		meta := metav1.ObjectMeta{
			Name:      svcName,
			Namespace: db.Namespace,
		}

		_, vt, err := core_util.CreateOrPatchService(
			context.TODO(),
			r.Client,
			meta,
			func(in *core.Service) *core.Service {
				core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
				in.Labels = labels

				in.Spec.Type = core.ServiceTypeClusterIP
				// create headless service
				in.Spec.ClusterIP = core.ClusterIPNone
				// create pod dns records
				in.Spec.Selector = selectors
				in.Spec.PublishNotReadyAddresses = true
				// create SRV records with pod DNS name as service provider
				in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
					{
						Name:       api.MongoDBDatabasePortName,
						Port:       api.MongoDBDatabasePort,
						TargetPort: intstr.FromString(api.MongoDBDatabasePortName),
					},
				})

				return in
			},
			metav1.PatchOptions{},
		)

		if err == nil && vt != kutil.VerbUnchanged {
			r.Recorder.Eventf(
				db,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully %s governing service",
				vt,
			)
		}
		return err
	}

	if db.Spec.ShardTopology != nil {
		topology := db.Spec.ShardTopology
		// create shard governing service
		for i := int32(0); i < topology.Shard.Shards; i++ {
			if err := svcFunc(db.GoverningServiceName(
				db.ShardNodeName(i)),
				db.ShardLabels(i),
				db.ShardSelectors(i),
			); err != nil {
				return err
			}
		}
		// create configsvr governing service
		if err := svcFunc(db.GoverningServiceName(
			db.ConfigSvrNodeName()),
			db.ConfigSvrLabels(),
			db.ConfigSvrSelectors(),
		); err != nil {
			return err
		}

		// create mongos governing service
		return svcFunc(db.GoverningServiceName(
			db.MongosNodeName()),
			db.MongosLabels(),
			db.MongosSelectors(),
		)
	}
	// create mongodb governing service
	return svcFunc(db.GoverningServiceName(
		db.OffshootName()),
		db.OffshootLabels(),
		db.OffshootSelectors(),
	)
}
