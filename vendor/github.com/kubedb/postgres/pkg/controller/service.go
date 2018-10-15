package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/kutil"
	core_util "github.com/appscode/kutil/core/v1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/eventer"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
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

func (c *Controller) ensureService(postgres *api.Postgres) (kutil.VerbType, error) {
	// Check if service name exists
	err := c.checkService(postgres, postgres.OffshootName())
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	// create database Service
	vt1, err := c.createService(postgres)
	if err != nil {
		c.recorder.Eventf(
			postgres,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to createOrPatch Service. Reason: %v",
			err,
		)
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
		c.recorder.Eventf(
			postgres,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to createOrPatch Service. Reason: %v",
			err,
		)
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
	service, err := c.Client.CoreV1().Services(postgres.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if service.Labels[api.LabelDatabaseKind] != api.ResourceKindPostgres ||
		service.Labels[api.LabelDatabaseName] != postgres.Name {
		return fmt.Errorf(`intended service "%v" already exists`, name)
	}

	return nil
}

func (c *Controller) createService(postgres *api.Postgres) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      postgres.OffshootName(),
		Namespace: postgres.Namespace,
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, postgres)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	_, ok, err := core_util.CreateOrPatchService(c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
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
	})
	return ok, err
}

func upsertServicePort(in *core.Service, postgres *api.Postgres) []core.ServicePort {
	return ofst.MergeServicePorts(
		core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
			{
				Name:       PostgresPortName,
				Port:       PostgresPort,
				TargetPort: intstr.FromString(PostgresPortName),
			},
		}),
		postgres.Spec.ServiceTemplate.Spec.Ports,
	)
}

func (c *Controller) createReplicasService(postgres *api.Postgres) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      postgres.ReplicasServiceName(),
		Namespace: postgres.Namespace,
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, postgres)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	_, ok, err := core_util.CreateOrPatchService(c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = postgres.OffshootSelectors()
		in.Spec.Selector = postgres.OffshootSelectors()
		in.Spec.Ports = upsertServicePort(in, postgres)
		return in
	})
	return ok, err
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

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, postgres)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	// reconcile stats service
	meta := metav1.ObjectMeta{
		Name:      postgres.StatsService().ServiceName(),
		Namespace: postgres.Namespace,
	}
	_, vt, err := core_util.CreateOrPatchService(c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = postgres.OffshootLabels()
		in.Spec.Selector = postgres.OffshootSelectors()
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
			{
				Name:       api.PrometheusExporterPortName,
				Protocol:   core.ProtocolTCP,
				Port:       postgres.Spec.Monitor.Prometheus.Port,
				TargetPort: intstr.FromString(api.PrometheusExporterPortName),
			},
		})
		return in
	})
	if err != nil {
		c.recorder.Eventf(
			ref,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to reconcile stats service. Reason: %v",
			err,
		)
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			ref,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s stats service",
			vt,
		)
	}
	return vt, nil
}
