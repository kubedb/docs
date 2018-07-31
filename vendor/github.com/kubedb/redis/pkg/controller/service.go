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
	mon_api "kmodules.xyz/monitoring-agent-api/api"
)

func (c *Controller) ensureService(redis *api.Redis) (kutil.VerbType, error) {
	// Check if service name exists
	if err := c.checkService(redis); err != nil {
		return kutil.VerbUnchanged, err
	}
	// create database Service
	vt, err := c.createService(redis)
	if err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				"Failed to create Service. Reason: %v",
				err,
			)
		}
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully %s Service",
				vt,
			)
		}
	}
	return vt, nil
}

func (c *Controller) checkService(redis *api.Redis) error {
	name := redis.OffshootName()
	service, err := c.Client.CoreV1().Services(redis.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if service.Spec.Selector[api.LabelDatabaseName] != name {
		return fmt.Errorf(`intended service "%v" already exists`, name)
	}

	return nil
}

func (c *Controller) createService(redis *api.Redis) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      redis.OffshootName(),
		Namespace: redis.Namespace,
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	_, ok, err := core_util.CreateOrPatchService(c.Client, meta, func(in *core.Service) *core.Service {
		in.ObjectMeta = core_util.EnsureOwnerReference(in.ObjectMeta, ref)
		in.Labels = redis.OffshootLabels()
		in.Spec.Ports = upsertServicePort(in, redis)
		in.Spec.Selector = redis.OffshootLabels()
		return in
	})
	return ok, err
}

func upsertServicePort(service *core.Service, redis *api.Redis) []core.ServicePort {
	desiredPorts := []core.ServicePort{
		{
			Name:       "db",
			Protocol:   core.ProtocolTCP,
			Port:       6379,
			TargetPort: intstr.FromString("db"),
		},
	}
	if redis.GetMonitoringVendor() == mon_api.VendorPrometheus {
		desiredPorts = append(desiredPorts, core.ServicePort{
			Name:       api.PrometheusExporterPortName,
			Protocol:   core.ProtocolTCP,
			Port:       redis.Spec.Monitor.Prometheus.Port,
			TargetPort: intstr.FromString(api.PrometheusExporterPortName),
		})
	}
	return core_util.MergeServicePorts(service.Spec.Ports, desiredPorts)
}
