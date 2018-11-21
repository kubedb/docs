package controller

import (
	"fmt"
	"path/filepath"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	"github.com/appscode/kutil"
	app_util "github.com/appscode/kutil/apps/v1"
	core_util "github.com/appscode/kutil/core/v1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/eventer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

const (
	CONFIG_MOUNT_PATH = "/usr/local/etc/redis/"
)

func (c *Controller) ensureStatefulSet(redis *api.Redis) (kutil.VerbType, error) {
	if err := c.checkStatefulSet(redis); err != nil {
		return kutil.VerbUnchanged, err
	}

	// Create statefulSet for Redis database
	statefulSet, vt, err := c.createStatefulSet(redis)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// Check StatefulSet Pod status
	if vt != kutil.VerbUnchanged {
		if err := c.checkStatefulSetPodStatus(statefulSet); err != nil {
			return kutil.VerbUnchanged, err
		}
		c.recorder.Eventf(
			redis,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet",
			vt,
		)
	}
	return vt, nil
}

func (c *Controller) checkStatefulSet(redis *api.Redis) error {
	// SatatefulSet for Redis database
	statefulSet, err := c.Client.AppsV1().StatefulSets(redis.Namespace).Get(redis.OffshootName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindRedis ||
		statefulSet.Labels[api.LabelDatabaseName] != redis.Name {
		return fmt.Errorf(`intended statefulSet "%v/%v" already exists`, redis.Namespace, redis.OffshootName())
	}

	return nil
}

func (c *Controller) createStatefulSet(redis *api.Redis) (*apps.StatefulSet, kutil.VerbType, error) {
	statefulSetMeta := metav1.ObjectMeta{
		Name:      redis.OffshootName(),
		Namespace: redis.Namespace,
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis)
	if rerr != nil {
		return nil, kutil.VerbUnchanged, rerr
	}

	redisVersion, err := c.ExtClient.CatalogV1alpha1().RedisVersions().Get(string(redis.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	return app_util.CreateOrPatchStatefulSet(c.Client, statefulSetMeta, func(in *apps.StatefulSet) *apps.StatefulSet {
		in.Labels = redis.OffshootLabels()
		in.Annotations = redis.Spec.PodTemplate.Controller.Annotations
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)

		in.Spec.Replicas = types.Int32P(1)
		in.Spec.ServiceName = c.GoverningService
		in.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: redis.OffshootSelectors(),
		}
		in.Spec.Template.Labels = redis.OffshootSelectors()
		in.Spec.Template.Annotations = redis.Spec.PodTemplate.Annotations
		in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(
			in.Spec.Template.Spec.InitContainers,
			redis.Spec.PodTemplate.Spec.InitContainers,
		)
		in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
			Name:            api.ResourceSingularRedis,
			Image:           redisVersion.Spec.DB.Image,
			ImagePullPolicy: core.PullIfNotPresent,
			Args:            redis.Spec.PodTemplate.Spec.Args,
			Ports: []core.ContainerPort{
				{
					Name:          "db",
					ContainerPort: 6379,
					Protocol:      core.ProtocolTCP,
				},
			},
			Resources:      redis.Spec.PodTemplate.Spec.Resources,
			LivenessProbe:  redis.Spec.PodTemplate.Spec.LivenessProbe,
			ReadinessProbe: redis.Spec.PodTemplate.Spec.ReadinessProbe,
			Lifecycle:      redis.Spec.PodTemplate.Spec.Lifecycle,
		})
		if redis.GetMonitoringVendor() == mona.VendorPrometheus {
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
				Name: "exporter",
				Args: []string{
					fmt.Sprintf("--web.listen-address=:%v", redis.Spec.Monitor.Prometheus.Port),
					fmt.Sprintf("--web.telemetry-path=%v", redis.StatsService().Path()),
				},
				Image:           redisVersion.Spec.Exporter.Image,
				ImagePullPolicy: core.PullIfNotPresent,
				Ports: []core.ContainerPort{
					{
						Name:          api.PrometheusExporterPortName,
						Protocol:      core.ProtocolTCP,
						ContainerPort: redis.Spec.Monitor.Prometheus.Port,
					},
				},
				Resources: redis.Spec.Monitor.Resources,
			})
		}

		in = upsertDataVolume(in, redis)

		in.Spec.Template.Spec.NodeSelector = redis.Spec.PodTemplate.Spec.NodeSelector
		in.Spec.Template.Spec.Affinity = redis.Spec.PodTemplate.Spec.Affinity
		if redis.Spec.PodTemplate.Spec.SchedulerName != "" {
			in.Spec.Template.Spec.SchedulerName = redis.Spec.PodTemplate.Spec.SchedulerName
		}
		in.Spec.Template.Spec.Tolerations = redis.Spec.PodTemplate.Spec.Tolerations
		in.Spec.Template.Spec.ImagePullSecrets = redis.Spec.PodTemplate.Spec.ImagePullSecrets
		in.Spec.Template.Spec.PriorityClassName = redis.Spec.PodTemplate.Spec.PriorityClassName
		in.Spec.Template.Spec.Priority = redis.Spec.PodTemplate.Spec.Priority
		in.Spec.Template.Spec.SecurityContext = redis.Spec.PodTemplate.Spec.SecurityContext

		in.Spec.UpdateStrategy = redis.Spec.UpdateStrategy
		in = upsertUserEnv(in, redis)
		in = upsertCustomConfig(in, redis)
		return in
	})
}

func upsertDataVolume(statefulSet *apps.StatefulSet, redis *api.Redis) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularRedis {
			volumeMount := core.VolumeMount{
				Name:      "data",
				MountPath: "/data",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			pvcSpec := redis.Spec.Storage
			if redis.Spec.StorageType == api.StorageTypeEphemeral {
				ed := core.EmptyDirVolumeSource{}
				if pvcSpec != nil {
					if sz, found := pvcSpec.Resources.Requests[core.ResourceStorage]; found {
						ed.SizeLimit = &sz
					}
				}
				statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(
					statefulSet.Spec.Template.Spec.Volumes,
					core.Volume{
						Name: "data",
						VolumeSource: core.VolumeSource{
							EmptyDir: &ed,
						},
					})
			} else {
				if len(pvcSpec.AccessModes) == 0 {
					pvcSpec.AccessModes = []core.PersistentVolumeAccessMode{
						core.ReadWriteOnce,
					}
					log.Infof(`Using "%v" as AccessModes in redis.spec.storage`, core.ReadWriteOnce)
				}

				claim := core.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: *pvcSpec,
				}
				if pvcSpec.StorageClassName != nil {
					claim.Annotations = map[string]string{
						"volume.beta.kubernetes.io/storage-class": *pvcSpec.StorageClassName,
					}
				}
				statefulSet.Spec.VolumeClaimTemplates = core_util.UpsertVolumeClaim(statefulSet.Spec.VolumeClaimTemplates, claim)
			}
		}
		break
	}
	return statefulSet
}

func (c *Controller) checkStatefulSetPodStatus(statefulSet *apps.StatefulSet) error {
	return core_util.WaitUntilPodRunningBySelector(
		c.Client,
		statefulSet.Namespace,
		statefulSet.Spec.Selector,
		int(types.Int32(statefulSet.Spec.Replicas)),
	)
}

// upsertUserEnv add/overwrite env from user provided env in crd spec
func upsertUserEnv(statefulset *apps.StatefulSet, redis *api.Redis) *apps.StatefulSet {
	for i, container := range statefulset.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularRedis {
			statefulset.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, redis.Spec.PodTemplate.Spec.Env...)
			return statefulset
		}
	}
	return statefulset
}

func upsertCustomConfig(statefulSet *apps.StatefulSet, redis *api.Redis) *apps.StatefulSet {
	if redis.Spec.ConfigSource != nil {
		for i, container := range statefulSet.Spec.Template.Spec.Containers {
			if container.Name == api.ResourceSingularRedis {
				configVolumeMount := core.VolumeMount{
					Name:      "custom-config",
					MountPath: CONFIG_MOUNT_PATH,
				}
				volumeMounts := container.VolumeMounts
				volumeMounts = core_util.UpsertVolumeMount(volumeMounts, configVolumeMount)
				statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

				configVolume := core.Volume{
					Name:         "custom-config",
					VolumeSource: *redis.Spec.ConfigSource,
				}

				volumes := statefulSet.Spec.Template.Spec.Volumes
				volumes = core_util.UpsertVolume(volumes, configVolume)
				statefulSet.Spec.Template.Spec.Volumes = volumes

				// send custom config file path as argument
				configPath := filepath.Join(CONFIG_MOUNT_PATH, "redis.conf")
				args := statefulSet.Spec.Template.Spec.Containers[i].Args
				if len(args) == 0 || args[len(args)-1] != configPath {
					args = append(args, configPath)
				}
				statefulSet.Spec.Template.Spec.Containers[i].Args = args
				break
			}
		}
	}
	return statefulSet
}
