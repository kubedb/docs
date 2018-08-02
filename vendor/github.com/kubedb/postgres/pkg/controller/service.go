package controller

import (
	"fmt"

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
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, postgres); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				"Failed to createOrPatch Service. Reason: %v",
				err,
			)
		}
		return kutil.VerbUnchanged, err
	} else if vt1 != kutil.VerbUnchanged {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, postgres); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully %s Service",
				vt1,
			)
		}
	}

	// Check if service name exists
	err = c.checkService(postgres, postgres.ReplicasServiceName())
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	// create database Service
	vt2, err := c.createReplicasService(postgres)
	if err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, postgres); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				"Failed to createOrPatch Service. Reason: %v",
				err,
			)
		}
		return kutil.VerbUnchanged, err
	} else if vt2 != kutil.VerbUnchanged {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, postgres); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully %s Service",
				vt2,
			)
		}
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

	if service.Spec.Selector[api.LabelDatabaseName] != postgres.OffshootName() {
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
		in.ObjectMeta = core_util.EnsureOwnerReference(in.ObjectMeta, ref)
		in.Labels = postgres.OffshootLabels()
		in.Spec.Selector = postgres.OffshootLabels()
		in.Spec.Ports = upsertServicePort(in, postgres)
		in.Spec.Selector[NodeRole] = "primary"
		return in
	})
	return ok, err
}

func upsertServicePort(service *core.Service, postgres *api.Postgres) []core.ServicePort {
	desiredPorts := []core.ServicePort{
		{
			Name:       PostgresPortName,
			Port:       PostgresPort,
			TargetPort: intstr.FromString(PostgresPortName),
		},
	}
	if postgres.GetMonitoringVendor() == mona.VendorPrometheus {
		desiredPorts = append(desiredPorts, core.ServicePort{
			Name:       api.PrometheusExporterPortName,
			Protocol:   core.ProtocolTCP,
			Port:       postgres.Spec.Monitor.Prometheus.Port,
			TargetPort: intstr.FromString(api.PrometheusExporterPortName),
		})
	}
	return core_util.MergeServicePorts(service.Spec.Ports, desiredPorts)
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
		in.ObjectMeta = core_util.EnsureOwnerReference(in.ObjectMeta, ref)
		in.Labels = postgres.OffshootLabels()
		in.Spec.Selector = postgres.OffshootLabels()
		in.Spec.Ports = upsertServicePort(in, postgres)
		return in
	})
	return ok, err
}
