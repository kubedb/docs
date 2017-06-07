package monitor

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	_ "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	prom "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	tapi "github.com/k8sdb/apimachinery/api"
	"github.com/k8sdb/apimachinery/pkg/docker"
	cgerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kapi "k8s.io/kubernetes/pkg/api"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
)

type PrometheusController struct {
	kubeClient        clientset.Interface
	promClient        *v1alpha1.MonitoringV1alpha1Client
	operatorNamespace string
}

func NewPrometheusController(kubeClient clientset.Interface, promClient *v1alpha1.MonitoringV1alpha1Client, operatorNamespace string) Monitor {
	return &PrometheusController{
		kubeClient:        kubeClient,
		promClient:        promClient,
		operatorNamespace: operatorNamespace,
	}
}

func (c *PrometheusController) AddMonitor(meta kapi.ObjectMeta, spec *tapi.MonitorSpec) error {
	if !c.SupportsCoreOSOperator() {
		return errors.New("Cluster does not support CoreOS Prometheus operator")
	}
	return c.ensureServiceMonitor(meta, spec, spec)
}

func (c *PrometheusController) UpdateMonitor(meta kapi.ObjectMeta, old, new *tapi.MonitorSpec) error {
	if !c.SupportsCoreOSOperator() {
		return errors.New("Cluster does not support CoreOS Prometheus operator")
	}
	return c.ensureServiceMonitor(meta, old, new)
}

func (c *PrometheusController) DeleteMonitor(meta kapi.ObjectMeta, spec *tapi.MonitorSpec) error {
	if !c.SupportsCoreOSOperator() {
		return errors.New("Cluster does not support CoreOS Prometheus operator")
	}
	if err := c.promClient.ServiceMonitors(spec.Prometheus.Namespace).Delete(getServiceMonitorName(meta), nil); !cgerr.IsNotFound(err) {
		return err
	}
	return nil
}

func (c *PrometheusController) SupportsCoreOSOperator() bool {
	_, err := c.kubeClient.Extensions().ThirdPartyResources().Get("prometheus." + prom.TPRGroup)
	if err != nil {
		return false
	}
	_, err = c.kubeClient.Extensions().ThirdPartyResources().Get("service-monitor." + prom.TPRGroup)
	if err != nil {
		return false
	}
	return true
}

func (c *PrometheusController) ensureServiceMonitor(meta kapi.ObjectMeta, old, new *tapi.MonitorSpec) error {
	name := getServiceMonitorName(meta)
	if new == nil || old.Prometheus.Namespace != new.Prometheus.Namespace {
		err := c.promClient.ServiceMonitors(old.Prometheus.Namespace).Delete(name, nil)
		if err != nil && !cgerr.IsNotFound(err) {
			return err
		}
		if new == nil {
			return nil
		}
	}

	actual, err := c.promClient.ServiceMonitors(new.Prometheus.Namespace).Get(name)
	if cgerr.IsNotFound(err) {
		return c.createServiceMonitor(meta, new)
	} else if err != nil {
		return err
	}
	if old != nil &&
		(!reflect.DeepEqual(old.Prometheus.Labels, new.Prometheus.Labels) || old.Prometheus.Interval != new.Prometheus.Interval) {
		actual.Labels = new.Prometheus.Labels
		for i := range actual.Spec.Endpoints {
			actual.Spec.Endpoints[i].Interval = new.Prometheus.Interval
		}
		_, err := c.promClient.ServiceMonitors(new.Prometheus.Namespace).Update(actual)
		return err
	}
	return nil
}

func (c *PrometheusController) createServiceMonitor(meta kapi.ObjectMeta, spec *tapi.MonitorSpec) error {
	svc, err := c.kubeClient.Core().Services(meta.Namespace).Get(meta.Name)
	if err != nil {
		return err
	}
	ports := svc.Spec.Ports
	if len(ports) == 0 {
		return errors.New("No port found in database service")
	}

	sm := &prom.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getServiceMonitorName(meta),
			Namespace: spec.Prometheus.Namespace,
			Labels:    spec.Prometheus.Labels,
		},
		Spec: prom.ServiceMonitorSpec{
			NamespaceSelector: prom.NamespaceSelector{
				MatchNames: []string{svc.Namespace},
			},
			Endpoints: []prom.Endpoint{
				{
					Address:  fmt.Sprintf("%s.%s.svc:%d", docker.OperatorName, c.operatorNamespace, docker.OperatorPortNumber),
					Port:     svc.Spec.Ports[0].Name,
					Interval: spec.Prometheus.Interval,
					Path:     fmt.Sprintf("/kubedb.com/v1beta1/namespaces/%s/%s/%s/pods/${__meta_kubernetes_pod_ip}/metrics", meta.Namespace, getTypeFromSelfLink(meta.SelfLink), meta.Name),
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: svc.Spec.Selector,
			},
		},
	}
	if _, err := c.promClient.ServiceMonitors(spec.Prometheus.Namespace).Create(sm); !cgerr.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func getTypeFromSelfLink(selfLink string) string {
	if len(selfLink) == 0 {
		return ""
	}
	s := strings.Split(selfLink, "/")
	return s[len(s)-2]
}

func getServiceMonitorName(meta kapi.ObjectMeta) string {
	return fmt.Sprintf("kubedb-%s-%s", meta.Namespace, meta.Name)
}
