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
	Port:       3306,
	TargetPort: intstr.FromString("db"),
}

func (c *Controller) checkService(db *api.MySQL, name string) error {
	service, err := c.Client.CoreV1().Services(db.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	// check services labels which are related to MySQL database or not
	if service.Labels[api.LabelDatabaseKind] != api.ResourceKindMySQL ||
		service.Labels[api.LabelDatabaseName] != db.Name {
		return fmt.Errorf(`intended service "%v/%v" already exists`, db.Namespace, name)
	}

	return nil
}

func (c *Controller) ensureMySQLGoverningService(db *api.MySQL) error {
	meta := metav1.ObjectMeta{
		Name:      db.GoverningServiceName(),
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

	// Check if service name exists with different db kind
	if err := c.checkService(db, db.GoverningServiceName()); err != nil {
		return err
	}

	_, vt, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = db.OffshootLabels()
		// 'tolerate-unready-endpoints' annotation is deprecated.
		// owner: https://github.com/kubernetes/kubernetes/pull/63742
		in.Annotations = map[string]string{
			"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
		}
		in.Spec.Selector = db.OffshootSelectors()
		in.Spec.Type = core.ServiceTypeClusterIP
		in.Spec.ClusterIP = core.ClusterIPNone
		in.Spec.Ports = []core.ServicePort{
			defaultDBPort,
		}
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
	if err := c.checkService(db, db.ServiceName()); err != nil {
		return err
	}
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
		if err := c.checkService(db, db.SecondaryServiceName()); err != nil {
			return err
		}
		vt, err := c.ensureSecondaryService(db)
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

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

	_, vt, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = db.OffshootLabels()
		in.Annotations = db.Spec.ServiceTemplate.Annotations

		in.Spec.Selector = db.OffshootSelectors()
		//add extra selector to select only primary pod for group replication
		if db.UsesGroupReplication() {
			in.Spec.Selector[api.MySQLLabelRole] = api.MySQLPodPrimary
		}

		in.Spec.Ports = ofst.MergeServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{defaultDBPort}),
			db.Spec.ServiceTemplate.Spec.Ports,
		)

		if db.Spec.ServiceTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = db.Spec.ServiceTemplate.Spec.ClusterIP
		}
		if db.Spec.ServiceTemplate.Spec.Type != "" {
			in.Spec.Type = db.Spec.ServiceTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = db.Spec.ServiceTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = db.Spec.ServiceTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = db.Spec.ServiceTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = db.Spec.ServiceTemplate.Spec.ExternalTrafficPolicy
		if db.Spec.ServiceTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = db.Spec.ServiceTemplate.Spec.HealthCheckNodePort
		}
		return in
	}, metav1.PatchOptions{})

	return vt, err
}

func (c *Controller) ensureSecondaryService(db *api.MySQL) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      db.SecondaryServiceName(),
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

	_, vt, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = db.OffshootLabels()
		in.Annotations = db.Spec.ServiceTemplate.Annotations
		in.Spec.Selector = db.OffshootSelectors()
		//add extra selector to select only secondary pod
		in.Spec.Selector[api.MySQLLabelRole] = api.MySQLPodSecondary

		in.Spec.Ports = ofst.MergeServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{defaultDBPort}),
			db.Spec.ServiceTemplate.Spec.Ports,
		)

		if db.Spec.ServiceTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = db.Spec.ServiceTemplate.Spec.ClusterIP
		}
		if db.Spec.ServiceTemplate.Spec.Type != "" {
			in.Spec.Type = db.Spec.ServiceTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = db.Spec.ServiceTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = db.Spec.ServiceTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = db.Spec.ServiceTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = db.Spec.ServiceTemplate.Spec.ExternalTrafficPolicy
		if db.Spec.ServiceTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = db.Spec.ServiceTemplate.Spec.HealthCheckNodePort
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

	stateServiceName := db.StatsService().ServiceName()

	if err := c.checkService(db, stateServiceName); err != nil {
		return kutil.VerbUnchanged, err
	}

	// stats Service
	meta := metav1.ObjectMeta{
		Name:      stateServiceName,
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

	_, vt, err := core_util.CreateOrPatchService(context.TODO(), c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = db.StatsServiceLabels()
		in.Spec.Selector = db.OffshootSelectors()
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
			{
				Name:       mona.PrometheusExporterPortName,
				Protocol:   core.ProtocolTCP,
				Port:       db.Spec.Monitor.Prometheus.Exporter.Port,
				TargetPort: intstr.FromString(mona.PrometheusExporterPortName),
			},
		})
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
