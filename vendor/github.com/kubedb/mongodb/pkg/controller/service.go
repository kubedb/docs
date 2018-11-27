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

const (
	MongoDBPort = 27017
)

func (c *Controller) ensureService(mongodb *api.MongoDB) (kutil.VerbType, error) {
	// Check if service name exists
	if err := c.checkService(mongodb, mongodb.ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	// create database Service
	vt, err := c.createService(mongodb)
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			mongodb,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s Service",
			vt,
		)
	}
	return vt, nil
}

func (c *Controller) checkService(mongodb *api.MongoDB, serviceName string) error {
	service, err := c.Client.CoreV1().Services(mongodb.Namespace).Get(serviceName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if service.Labels[api.LabelDatabaseKind] != api.ResourceKindMongoDB ||
		service.Labels[api.LabelDatabaseName] != mongodb.Name {
		return fmt.Errorf(`intended service "%v/%v" already exists`, mongodb.Namespace, serviceName)
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
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = mongodb.OffshootLabels()
		in.Annotations = mongodb.Spec.ServiceTemplate.Annotations

		in.Spec.Selector = mongodb.OffshootSelectors()
		in.Spec.Ports = ofst.MergeServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
				{
					Name:       "db",
					Protocol:   core.ProtocolTCP,
					Port:       MongoDBPort,
					TargetPort: intstr.FromString("db"),
				},
			}),
			mongodb.Spec.ServiceTemplate.Spec.Ports,
		)

		if mongodb.Spec.ServiceTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = mongodb.Spec.ServiceTemplate.Spec.ClusterIP
		}
		if mongodb.Spec.ServiceTemplate.Spec.Type != "" {
			in.Spec.Type = mongodb.Spec.ServiceTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = mongodb.Spec.ServiceTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = mongodb.Spec.ServiceTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = mongodb.Spec.ServiceTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = mongodb.Spec.ServiceTemplate.Spec.ExternalTrafficPolicy
		if mongodb.Spec.ServiceTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = mongodb.Spec.ServiceTemplate.Spec.HealthCheckNodePort
		}
		return in
	})
	return ok, err
}

func (c *Controller) ensureStatsService(mongodb *api.MongoDB) (kutil.VerbType, error) {
	// return if monitoring is not prometheus
	if mongodb.GetMonitoringVendor() != mona.VendorPrometheus {
		log.Infoln("spec.monitor.agent is not coreos-operator or builtin.")
		return kutil.VerbUnchanged, nil
	}

	// Check if stats Service name exists
	if err := c.checkService(mongodb, mongodb.StatsService().ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	// create/patch stats Service
	meta := metav1.ObjectMeta{
		Name:      mongodb.StatsService().ServiceName(),
		Namespace: mongodb.Namespace,
	}
	_, vt, err := core_util.CreateOrPatchService(c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = mongodb.OffshootLabels()
		in.Spec.Selector = mongodb.OffshootSelectors()
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
			{
				Name:       api.PrometheusExporterPortName,
				Protocol:   core.ProtocolTCP,
				Port:       mongodb.Spec.Monitor.Prometheus.Port,
				TargetPort: intstr.FromString(api.PrometheusExporterPortName),
			},
		})
		return in
	})
	if err != nil {
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

func (c *Controller) createMongoDBGoverningService(mongodb *api.MongoDB) (string, error) {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, mongodb)
	if rerr != nil {
		return "", rerr
	}

	service := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mongodb.GoverningServiceName(),
			Namespace: mongodb.Namespace,
			Labels:    mongodb.OffshootLabels(),
			// 'tolerate-unready-endpoints' annotation is deprecated.
			// ref: https://github.com/kubernetes/kubernetes/pull/63742
			Annotations: map[string]string{
				"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
			},
		},
		Spec: core.ServiceSpec{
			Type:                     core.ServiceTypeClusterIP,
			ClusterIP:                core.ClusterIPNone,
			PublishNotReadyAddresses: true,
			Ports: []core.ServicePort{
				{
					Name: "db",
					Port: MongoDBPort,
				},
			},
			Selector: mongodb.OffshootSelectors(),
		},
	}
	core_util.EnsureOwnerReference(&service.ObjectMeta, ref)

	_, err := c.Client.CoreV1().Services(mongodb.Namespace).Create(service)
	if err != nil && !kerr.IsAlreadyExists(err) {
		return "", err
	}
	return service.Name, nil
}
