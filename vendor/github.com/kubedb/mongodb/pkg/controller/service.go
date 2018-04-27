package controller

import (
	"fmt"

	mon_api "github.com/appscode/kube-mon/api"
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
)

func (c *Controller) ensureService(mongodb *api.MongoDB) (kutil.VerbType, error) {
	// Check if service name exists
	if err := c.checkService(mongodb); err != nil {
		return kutil.VerbUnchanged, err
	}

	// create database Service
	vt, err := c.createService(mongodb)
	if err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb); rerr == nil {
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
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb); rerr == nil {
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

func (c *Controller) checkService(mongodb *api.MongoDB) error {
	name := mongodb.OffshootName()
	service, err := c.Client.CoreV1().Services(mongodb.Namespace).Get(name, metav1.GetOptions{})
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

func (c *Controller) createService(mongodb *api.MongoDB) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      mongodb.OffshootName(),
		Namespace: mongodb.Namespace,
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	_, ok, err := core_util.CreateOrPatchService(c.Client, meta, func(in *core.Service) *core.Service {
		in.ObjectMeta = core_util.EnsureOwnerReference(in.ObjectMeta, ref)
		in.Labels = mongodb.OffshootLabels()
		in.Spec.Ports = upsertServicePort(in, mongodb)
		in.Spec.Selector = mongodb.OffshootLabels()
		return in
	})
	return ok, err
}

func upsertServicePort(service *core.Service, mongodb *api.MongoDB) []core.ServicePort {
	desiredPorts := []core.ServicePort{
		{
			Name:       "db",
			Protocol:   core.ProtocolTCP,
			Port:       27017,
			TargetPort: intstr.FromString("db"),
		},
	}
	if mongodb.GetMonitoringVendor() == mon_api.VendorPrometheus {
		desiredPorts = append(desiredPorts, core.ServicePort{
			Name:       api.PrometheusExporterPortName,
			Protocol:   core.ProtocolTCP,
			Port:       mongodb.Spec.Monitor.Prometheus.Port,
			TargetPort: intstr.FromString(api.PrometheusExporterPortName),
		})
	}
	return core_util.MergeServicePorts(service.Spec.Ports, desiredPorts)
}
