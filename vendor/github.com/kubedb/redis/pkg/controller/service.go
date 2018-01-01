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
)

func (c *Controller) ensureService(redis *api.Redis) (kutil.VerbType, error) {
	// Check if service name exists
	if err := c.checkService(redis); err != nil {
		return kutil.VerbUnchanged, err
	}
	// create database Service
	vt, err := c.createService(redis)
	if err != nil {
		c.recorder.Eventf(
			redis.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to createOrPatch Service. Reason: %v",
			err,
		)
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			redis.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s Service",
			vt,
		)
	}
	return vt, nil
}

func (c *Controller) checkService(redis *api.Redis) error {
	name := redis.OffshootName()
	service, err := c.Client.CoreV1().Services(redis.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		} else {
			return err
		}
	}

	if service.Spec.Selector[api.LabelDatabaseName] != name {
		return fmt.Errorf(`Intended service "%v" already exists`, name)
	}

	return nil
}

func (c *Controller) createService(redis *api.Redis) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      redis.OffshootName(),
		Namespace: redis.Namespace,
	}

	_, vt, err := core_util.CreateOrPatchService(c.Client, meta, func(in *core.Service) *core.Service {
		in.Labels = redis.OffshootLabels()
		in.Spec.Ports = upsertServicePort(in, redis)
		in.Spec.Selector = redis.OffshootLabels()
		return in
	})
	return vt, err
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
	if redis.Spec.Monitor != nil &&
		redis.Spec.Monitor.Agent == api.AgentCoreosPrometheus &&
		redis.Spec.Monitor.Prometheus != nil {
		desiredPorts = append(desiredPorts, core.ServicePort{
			Name:       api.PrometheusExporterPortName,
			Protocol:   core.ProtocolTCP,
			Port:       redis.Spec.Monitor.Prometheus.Port,
			TargetPort: intstr.FromString(api.PrometheusExporterPortName),
		})
	}
	return core_util.MergeServicePorts(service.Spec.Ports, desiredPorts)
}

func (c *Controller) deleteService(name, namespace string) error {
	service, err := c.Client.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		} else {
			return err
		}
	}

	if service.Spec.Selector[api.LabelDatabaseName] != name {
		return nil
	}

	return c.Client.CoreV1().Services(namespace).Delete(name, nil)
}
