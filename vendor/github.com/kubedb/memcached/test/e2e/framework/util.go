package framework

import (
	"bytes"
	"fmt"
	"strings"

	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func deleteInBackground() *metav1.DeleteOptions {
	policy := metav1.DeletePropagationBackground
	return &metav1.DeleteOptions{PropagationPolicy: &policy}
}

func deleteInForeground() *metav1.DeleteOptions {
	policy := metav1.DeletePropagationForeground
	return &metav1.DeleteOptions{PropagationPolicy: &policy}
}

func (fi *Invocation) GetPod(meta metav1.ObjectMeta) (*core.Pod, error) {
	podList, err := fi.kubeClient.CoreV1().Pods(meta.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, pod := range podList.Items {
		if bytes.HasPrefix([]byte(pod.Name), []byte(meta.Name)) {
			return &pod, nil
		}
	}
	return nil, fmt.Errorf("no pod found for workload %v", meta.Name)
}

type MemcdConfig struct {
	Name  string
	Value string
	Alias string
}

func (f *Invocation) GetCustomConfig(configs []MemcdConfig) *core.ConfigMap {
	data := make([]string, 0)
	for _, cfg := range configs {
		data = append(data, cfg.Name+" = "+cfg.Value)
	}
	return &core.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.app,
			Namespace: f.namespace,
		},
		Data: map[string]string{
			"memcached.conf": strings.Join(data, "\n"),
		},
	}
}

func (f *Invocation) CreateConfigMap(obj *core.ConfigMap) error {
	_, err := f.kubeClient.CoreV1().ConfigMaps(obj.Namespace).Create(obj)
	return err
}

func (f *Invocation) DeleteConfigMap(meta metav1.ObjectMeta) error {
	err := f.kubeClient.CoreV1().ConfigMaps(meta.Namespace).Delete(meta.Name, deleteInForeground())
	if err != nil && !kerr.IsNotFound(err) {
		return err
	}
	return nil
}

func (f *Invocation) GetTestService(meta metav1.ObjectMeta) *core.Service {
	return &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name + "-test-svc",
			Namespace: meta.Namespace,
		},
		Spec: core.ServiceSpec{
			Type: core.ServiceTypeNodePort,
			Ports: []core.ServicePort{
				{
					Name:       "db",
					Protocol:   core.ProtocolTCP,
					Port:       11211,
					NodePort:   32757,
					TargetPort: intstr.FromString("db"),
				},
			},
			Selector: map[string]string{
				api.LabelDatabaseName: meta.Name,
				api.LabelDatabaseKind: api.ResourceKindMemcached,
			},
		},
	}
}

func (f *Invocation) CreateService(obj *core.Service) error {
	_, err := f.kubeClient.CoreV1().Services(obj.Namespace).Create(obj)
	return err
}

func (f *Invocation) DeleteService(meta metav1.ObjectMeta) error {
	err := f.kubeClient.CoreV1().Services(meta.Namespace).Delete(meta.Name, deleteInForeground())
	if err != nil && !kerr.IsNotFound(err) {
		return err
	}
	return nil
}
