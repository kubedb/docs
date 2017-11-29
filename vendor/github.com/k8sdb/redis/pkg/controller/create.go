package controller

import (
	"fmt"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	api "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/k8sdb/apimachinery/pkg/docker"
	apps "k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// Duration in Minute
	// Check whether pod under StatefulSet is running or not
	// Continue checking for this duration until failure
	durationCheckStatefulSet = time.Minute * 30
)

func (c *Controller) findService(redis *api.Redis) (bool, error) {
	name := redis.OffshootName()
	service, err := c.Client.CoreV1().Services(redis.Namespace).Get(name, metav1.GetOptions{})
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

func (c *Controller) createService(redis *api.Redis) error {
	svc := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   redis.OffshootName(),
			Labels: redis.OffshootLabels(),
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				{
					Name:       "db",
					Port:       6379,
					TargetPort: intstr.FromString("db"),
				},
			},
			Selector: redis.OffshootLabels(),
		},
	}
	if redis.Spec.Monitor != nil &&
		redis.Spec.Monitor.Agent == api.AgentCoreosPrometheus &&
		redis.Spec.Monitor.Prometheus != nil {
		svc.Spec.Ports = append(svc.Spec.Ports, core.ServicePort{
			Name:       api.PrometheusExporterPortName,
			Port:       redis.Spec.Monitor.Prometheus.Port,
			TargetPort: intstr.FromString(api.PrometheusExporterPortName),
		})
	}

	if _, err := c.Client.CoreV1().Services(redis.Namespace).Create(svc); err != nil {
		return err
	}

	return nil
}

func (c *Controller) findStatefulSet(redis *api.Redis) (bool, error) {
	// SatatefulSet for Redis database
	statefulSet, err := c.Client.AppsV1beta1().StatefulSets(redis.Namespace).Get(redis.OffshootName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindRedis {
		return false, fmt.Errorf(`Intended statefulSet "%v" already exists`, redis.OffshootName())
	}

	return true, nil
}

func (c *Controller) createStatefulSet(redis *api.Redis) (*apps.StatefulSet, error) {
	// SatatefulSet for Redis database
	statefulSet := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        redis.OffshootName(),
			Namespace:   redis.Namespace,
			Labels:      redis.StatefulSetLabels(),
			Annotations: redis.StatefulSetAnnotations(),
		},
		Spec: apps.StatefulSetSpec{
			Replicas:    types.Int32P(1),
			ServiceName: c.opt.GoverningService,
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: redis.OffshootLabels(),
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:            api.ResourceNameRedis,
							Image:           fmt.Sprintf("%s:%s", docker.ImageRedis, redis.Spec.Version),
							ImagePullPolicy: core.PullIfNotPresent,
							Ports: []core.ContainerPort{
								{
									Name:          "db",
									ContainerPort: 6379,
								},
							},
							Resources: redis.Spec.Resources,
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "data",
									MountPath: "/data",
								},
							},
						},
					},
					NodeSelector:  redis.Spec.NodeSelector,
					Affinity:      redis.Spec.Affinity,
					SchedulerName: redis.Spec.SchedulerName,
					Tolerations:   redis.Spec.Tolerations,
				},
			},
		},
	}

	if redis.Spec.Monitor != nil &&
		redis.Spec.Monitor.Agent == api.AgentCoreosPrometheus &&
		redis.Spec.Monitor.Prometheus != nil {
		exporter := core.Container{
			Name: "exporter",
			Args: []string{
				"export",
				fmt.Sprintf("--address=:%d", redis.Spec.Monitor.Prometheus.Port),
				"--v=3",
			},
			Image:           docker.ImageOperator + ":" + c.opt.ExporterTag,
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          api.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: redis.Spec.Monitor.Prometheus.Port,
				},
			},
		}
		statefulSet.Spec.Template.Spec.Containers = append(statefulSet.Spec.Template.Spec.Containers, exporter)
	}

	// Add Data volume for StatefulSet
	addDataVolume(statefulSet, redis.Spec.Storage)

	if c.opt.EnableRbac {
		// Ensure ClusterRoles for database statefulsets
		if err := c.createRBACStuff(redis); err != nil {
			return nil, err
		}

		statefulSet.Spec.Template.Spec.ServiceAccountName = redis.Name
	}

	if _, err := c.Client.AppsV1beta1().StatefulSets(statefulSet.Namespace).Create(statefulSet); err != nil {
		return nil, err
	}

	return statefulSet, nil
}

func addDataVolume(statefulSet *apps.StatefulSet, pvcSpec *core.PersistentVolumeClaimSpec) {
	if pvcSpec != nil {
		if len(pvcSpec.AccessModes) == 0 {
			pvcSpec.AccessModes = []core.PersistentVolumeAccessMode{
				core.ReadWriteOnce,
			}
			log.Infof(`Using "%v" as AccessModes in "%v"`, core.ReadWriteOnce, *pvcSpec)
		}
		// volume claim templates
		// Dynamically attach volume
		statefulSet.Spec.VolumeClaimTemplates = []core.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "data",
					Annotations: map[string]string{
						"volume.beta.kubernetes.io/storage-class": *pvcSpec.StorageClassName,
					},
				},
				Spec: *pvcSpec,
			},
		}
	} else {
		// Attach Empty directory
		statefulSet.Spec.Template.Spec.Volumes = append(
			statefulSet.Spec.Template.Spec.Volumes,
			core.Volume{
				Name: "data",
				VolumeSource: core.VolumeSource{
					EmptyDir: &core.EmptyDirVolumeSource{},
				},
			},
		)
	}
}

func (c *Controller) createDormantDatabase(redis *api.Redis) (*api.DormantDatabase, error) {
	dormantDb := &api.DormantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redis.Name,
			Namespace: redis.Namespace,
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindRedis,
			},
		},
		Spec: api.DormantDatabaseSpec{
			Origin: api.Origin{
				ObjectMeta: metav1.ObjectMeta{
					Name:        redis.Name,
					Namespace:   redis.Namespace,
					Labels:      redis.Labels,
					Annotations: redis.Annotations,
				},
				Spec: api.OriginSpec{
					Redis: &redis.Spec,
				},
			},
		},
	}

	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}

func (c *Controller) reCreateRedis(redis *api.Redis) error {
	_redis := &api.Redis{
		ObjectMeta: metav1.ObjectMeta{
			Name:        redis.Name,
			Namespace:   redis.Namespace,
			Labels:      redis.Labels,
			Annotations: redis.Annotations,
		},
		Spec:   redis.Spec,
		Status: redis.Status,
	}

	if _, err := c.ExtClient.Redises(_redis.Namespace).Create(_redis); err != nil {
		return err
	}

	return nil
}
