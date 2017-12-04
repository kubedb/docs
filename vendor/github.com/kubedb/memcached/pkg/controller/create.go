package controller

import (
	"fmt"
	"time"

	"github.com/appscode/go/types"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/docker"
	apps "k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// Duration in Minute
	// Check whether pod under Deployment is running or not
	// Continue checking for this duration until failure
	durationCheckDeployment = time.Minute * 30
)

func (c *Controller) findService(memcached *api.Memcached) (bool, error) {
	name := memcached.OffshootName()
	service, err := c.Client.CoreV1().Services(memcached.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	if service.Spec.Selector[api.LabelDatabaseName] != name {
		return false, fmt.Errorf(`Intended service "%v" already exists`, name)
	}

	return true, nil
}

func (c *Controller) createService(memcached *api.Memcached) error {
	svc := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   memcached.OffshootName(),
			Labels: memcached.OffshootLabels(),
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				{
					Name:       "db",
					Port:       11211,
					TargetPort: intstr.FromString("db"),
				},
			},
			Selector: memcached.OffshootLabels(),
		},
	}
	if memcached.Spec.Monitor != nil &&
		memcached.Spec.Monitor.Agent == api.AgentCoreosPrometheus &&
		memcached.Spec.Monitor.Prometheus != nil {
		svc.Spec.Ports = append(svc.Spec.Ports, core.ServicePort{
			Name:       api.PrometheusExporterPortName,
			Port:       memcached.Spec.Monitor.Prometheus.Port,
			TargetPort: intstr.FromString(api.PrometheusExporterPortName),
		})
	}

	if _, err := c.Client.CoreV1().Services(memcached.Namespace).Create(svc); err != nil {
		return err
	}

	return nil
}

func (c *Controller) findDeployment(memcached *api.Memcached) (bool, error) {
	// Deployment for Memcached database
	deployment, err := c.Client.AppsV1beta1().Deployments(memcached.Namespace).Get(memcached.OffshootName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	if deployment.Labels[api.LabelDatabaseKind] != api.ResourceKindMemcached {
		return false, fmt.Errorf(`intended deployment "%v" already exists`, memcached.OffshootName())
	}

	return true, nil
}

func (c *Controller) createDeployment(memcached *api.Memcached) (*apps.Deployment, error) {
	// Deployment for Memcached database
	if memcached.Spec.Replicas == 0 {
		memcached.Spec.Replicas = 1
	}
	deployment := &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        memcached.OffshootName(),
			Namespace:   memcached.Namespace,
			Labels:      memcached.DeploymentLabels(),
			Annotations: memcached.DeploymentAnnotations(),
		},
		Spec: apps.DeploymentSpec{
			Replicas: types.Int32P(memcached.Spec.Replicas),
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: memcached.OffshootLabels(),
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:            api.ResourceNameMemcached,
							Image:           fmt.Sprintf("%s:%s", docker.ImageMemcached, memcached.Spec.Version),
							ImagePullPolicy: core.PullIfNotPresent,
							Ports: []core.ContainerPort{
								{
									Name:          "db",
									ContainerPort: 11211,
								},
							},
							Resources: memcached.Spec.Resources,
						},
					},
					NodeSelector:  memcached.Spec.NodeSelector,
					Affinity:      memcached.Spec.Affinity,
					SchedulerName: memcached.Spec.SchedulerName,
					Tolerations:   memcached.Spec.Tolerations,
				},
			},
		},
	}

	if memcached.Spec.Monitor != nil &&
		memcached.Spec.Monitor.Agent == api.AgentCoreosPrometheus &&
		memcached.Spec.Monitor.Prometheus != nil {
		exporter := core.Container{
			Name: "exporter",
			Args: []string{
				"export",
				fmt.Sprintf("--address=:%d", memcached.Spec.Monitor.Prometheus.Port),
				"--v=3",
			},
			Image:           docker.ImageOperator + ":" + c.opt.ExporterTag,
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          api.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: memcached.Spec.Monitor.Prometheus.Port,
				},
			},
		}
		deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, exporter)
	}

	if c.opt.EnableRbac {
		// Ensure ClusterRoles for database deployment
		if err := c.createRBACStuff(memcached); err != nil {
			return nil, err
		}

		deployment.Spec.Template.Spec.ServiceAccountName = memcached.Name
	}

	if _, err := c.Client.AppsV1beta1().Deployments(deployment.Namespace).Create(deployment); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (c *Controller) createDormantDatabase(memcached *api.Memcached) (*api.DormantDatabase, error) {
	dormantDb := &api.DormantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      memcached.Name,
			Namespace: memcached.Namespace,
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindMemcached,
			},
		},
		Spec: api.DormantDatabaseSpec{
			Origin: api.Origin{
				ObjectMeta: metav1.ObjectMeta{
					Name:        memcached.Name,
					Namespace:   memcached.Namespace,
					Labels:      memcached.Labels,
					Annotations: memcached.Annotations,
				},
				Spec: api.OriginSpec{
					Memcached: &memcached.Spec,
				},
			},
		},
	}

	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}

func (c *Controller) reCreateMemcached(memcached *api.Memcached) error {
	_memcached := &api.Memcached{
		ObjectMeta: metav1.ObjectMeta{
			Name:        memcached.Name,
			Namespace:   memcached.Namespace,
			Labels:      memcached.Labels,
			Annotations: memcached.Annotations,
		},
		Spec:   memcached.Spec,
		Status: memcached.Status,
	}

	if _, err := c.ExtClient.Memcacheds(_memcached.Namespace).Create(_memcached); err != nil {
		return err
	}

	return nil
}
