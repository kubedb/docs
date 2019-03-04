package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/eventer"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

const TolerateUnreadyEndpointsAnnotation = "service.alpha.kubernetes.io/tolerate-unready-endpoints"

var (
	defaultClientPort = v1.ServicePort{
		Name:       "client",
		Port:       EtcdClientPort,
		TargetPort: intstr.FromInt(EtcdClientPort),
		Protocol:   v1.ProtocolTCP,
	}
	defaultPeerPort = v1.ServicePort{
		Name:       "peer",
		Port:       EtcdPeerPort,
		TargetPort: intstr.FromInt(EtcdPeerPort),
		Protocol:   v1.ProtocolTCP,
	}
)

func (c *Controller) CreateClientService(cl *Cluster) error {
	ports := []v1.ServicePort{defaultClientPort}

	return createService(c.Controller.Client, cl.cluster.ClientServiceName(), "", ports, cl.cluster)
}

func (c *Controller) CreatePeerService(cl *Cluster) error {
	ports := []v1.ServicePort{defaultClientPort, defaultPeerPort}

	return createService(c.Controller.Client, cl.cluster.PeerServiceName(), v1.ClusterIPNone, ports, cl.cluster)
}

func createService(kubecli kubernetes.Interface, svcName, clusterIP string, ports []v1.ServicePort, etcd *api.Etcd) error {
	meta := metav1.ObjectMeta{
		Name:      svcName,
		Namespace: etcd.Namespace,
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd)
	if rerr != nil {
		return rerr
	}

	_, _, err := core_util.CreateOrPatchService(kubecli, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = etcd.OffshootLabels()
		in.Annotations = etcd.Spec.ServiceTemplate.Annotations
		if in.Annotations == nil {
			in.Annotations = map[string]string{}
		}
		in.Annotations[TolerateUnreadyEndpointsAnnotation] = "true"

		in.Spec.Selector = etcd.OffshootSelectors()
		in.Spec.Ports = ofst.MergeServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, ports),
			etcd.Spec.ServiceTemplate.Spec.Ports,
		)

		in.Spec.ClusterIP = clusterIP
		if etcd.Spec.ServiceTemplate.Spec.Type != "" {
			in.Spec.Type = etcd.Spec.ServiceTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = etcd.Spec.ServiceTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = etcd.Spec.ServiceTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = etcd.Spec.ServiceTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = etcd.Spec.ServiceTemplate.Spec.ExternalTrafficPolicy
		if etcd.Spec.ServiceTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = etcd.Spec.ServiceTemplate.Spec.HealthCheckNodePort
		}
		return in
	})
	return err
}

func (c *Controller) checkService(etcd *api.Etcd, serviceName string) error {
	service, err := c.Client.CoreV1().Services(etcd.Namespace).Get(serviceName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if service.Labels[api.LabelDatabaseKind] != api.ResourceKindEtcd ||
		service.Labels[api.LabelDatabaseName] != etcd.Name {
		return fmt.Errorf(`intended service "%v" already exists`, serviceName)
	}

	return nil
}

func (c *Controller) ensureStatsService(etcd *api.Etcd) (kutil.VerbType, error) {
	// return if monitoring is not prometheus
	if etcd.GetMonitoringVendor() != mona.VendorPrometheus {
		log.Infoln("spec.monitor.agent is not coreos-operator or builtin.")
		return kutil.VerbUnchanged, nil
	}

	// Check if statsService name exists
	if err := c.checkService(etcd, etcd.StatsService().ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	// reconcile stats Service
	meta := metav1.ObjectMeta{
		Name:      etcd.StatsService().ServiceName(),
		Namespace: etcd.Namespace,
	}
	_, vt, err := core_util.CreateOrPatchService(c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = etcd.OffshootLabels()
		in.Spec.Selector = etcd.OffshootSelectors()
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
			{
				Name:       api.PrometheusExporterPortName,
				Protocol:   core.ProtocolTCP,
				Port:       etcd.Spec.Monitor.Prometheus.Port,
				TargetPort: intstr.FromString("client"),
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
