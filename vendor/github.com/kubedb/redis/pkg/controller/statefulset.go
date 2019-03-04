package controller

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/eventer"
	configure_cluster "github.com/kubedb/redis/pkg/configure-cluster"
	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

const (
	CONFIG_MOUNT_PATH = "/usr/local/etc/redis/"
)

func (c *Controller) ensureStatefulSet(redis *api.Redis, statefulSetName string, removeSlave bool) (kutil.VerbType, error) {
	err := c.checkStatefulSet(redis, statefulSetName)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// Create statefulSet for Redis database
	statefulSet, vt, err := c.createStatefulSet(redis, statefulSetName, removeSlave)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// Check StatefulSet Pod status
	if vt != kutil.VerbUnchanged {
		if err := c.checkStatefulSetPodStatus(statefulSet); err != nil {
			return kutil.VerbUnchanged, errors.Wrap(err, "Failed to CreateOrPatch StatefulSet")
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

func (c *Controller) ensureRedisNodes(redis *api.Redis) (kutil.VerbType, error) {
	var (
		vt  kutil.VerbType
		err error
	)

	if redis.Spec.Mode == api.RedisModeStandalone {
		vt, err = c.ensureStatefulSet(redis, redis.OffshootName(), false)
		if err != nil {
			return vt, err
		}
	} else {
		for i := 0; i < int(*redis.Spec.Cluster.Master); i++ {
			vt, err = c.ensureStatefulSet(redis, redis.StatefulSetNameWithShard(i), false)
			if err != nil {
				return vt, err
			}
		}

		statefulSets, err := c.Client.AppsV1().StatefulSets(redis.Namespace).List(metav1.ListOptions{
			LabelSelector: labels.Set{
				api.LabelDatabaseKind: api.ResourceKindRedis,
				api.LabelDatabaseName: redis.Name,
			}.String(),
		})
		if err != nil {
			return vt, err
		}

		pods := make([][]*core.Pod, len(statefulSets.Items))
		for i := 0; i < len(statefulSets.Items); i++ {
			stsIndex, _ := strconv.Atoi(statefulSets.Items[i].Name[len(redis.BaseNameForShard()):])
			pods[stsIndex] = make([]*core.Pod, *statefulSets.Items[i].Spec.Replicas)
			for j := 0; j < int(*statefulSets.Items[i].Spec.Replicas); j++ {
				podName := fmt.Sprintf("%s-%d", statefulSets.Items[i].Name, j)
				pods[stsIndex][j], err = c.Client.CoreV1().Pods(redis.Namespace).Get(podName, metav1.GetOptions{})
				if err != nil {
					return vt, err
				}
			}
		}

		if err := configure_cluster.ConfigureRedisCluster(c.ClientConfig, redis, pods); err != nil {
			return vt, errors.Wrap(err, "failed to configure required cluster")
		}

		log.Infoln("Cluster configured")
		log.Infoln("Checking for removing master(s)...")
		statefulSets, err = c.Client.AppsV1().StatefulSets(redis.Namespace).List(metav1.ListOptions{
			LabelSelector: labels.Set{
				api.LabelDatabaseKind: api.ResourceKindRedis,
				api.LabelDatabaseName: redis.Name,
			}.String(),
		})
		if err != nil {
			return vt, err
		}
		if len(statefulSets.Items) > int(*redis.Spec.Cluster.Master) {
			log.Infoln("Removing masters...")

			foregroundPolicy := metav1.DeletePropagationForeground
			for i := int(*redis.Spec.Cluster.Master); i < len(statefulSets.Items); i++ {
				err = c.Client.AppsV1().
					StatefulSets(redis.Namespace).
					Delete(redis.StatefulSetNameWithShard(i), &metav1.DeleteOptions{
						PropagationPolicy: &foregroundPolicy,
					})
				if err != nil {
					return vt, err
				}
			}
		}

		log.Infoln("Checking for removing slave(s)...")

		// update the the statefulSets with reduced replicas as some of their slaves have been
		// removed when redis.spec.cluster.replicas field is reduced
		statefulSets, err = c.Client.AppsV1().StatefulSets(redis.Namespace).List(metav1.ListOptions{
			LabelSelector: labels.Set{
				api.LabelDatabaseKind: api.ResourceKindRedis,
				api.LabelDatabaseName: redis.Name,
			}.String(),
		})
		if err != nil {
			return vt, err
		}
		for i := 0; i < int(*redis.Spec.Cluster.Master); i++ {
			if int(*statefulSets.Items[i].Spec.Replicas) > int(*redis.Spec.Cluster.Replicas)+1 {
				log.Infoln("Removing slaves...")

				vt, err = c.ensureStatefulSet(redis, redis.StatefulSetNameWithShard(i), true)
				if err != nil {
					return vt, err
				}
			}
		}
	}

	return vt, nil
}

func (c *Controller) checkStatefulSet(redis *api.Redis, statefulSetName string) error {
	// SatatefulSet for Redis database
	statefulSet, err := c.Client.AppsV1().StatefulSets(redis.Namespace).Get(statefulSetName, metav1.GetOptions{})
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

func (c *Controller) createStatefulSet(redis *api.Redis, statefulSetName string, removeSlave bool) (*apps.StatefulSet, kutil.VerbType, error) {
	statefulSetMeta := metav1.ObjectMeta{
		Name:      statefulSetName,
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

		if redis.Spec.Mode == api.RedisModeStandalone {
			in.Spec.Replicas = types.Int32P(1)
		} else if redis.Spec.Mode == api.RedisModeCluster {
			// while creating first time, in.Spec.Replicas is 'nil'
			if in.Spec.Replicas == nil ||
				// while adding slave(s), (*in.Spec.Replicas < *redis.Spec.Cluster.Replicas + 1) is true
				*in.Spec.Replicas < *redis.Spec.Cluster.Replicas+1 ||
				// removeSlave is true only after deleting slave node(s) in the stage of configuring redis cluster
				removeSlave {
				in.Spec.Replicas = types.Int32P(*redis.Spec.Cluster.Replicas + 1)
			}
		}
		in.Spec.ServiceName = c.GoverningService

		labels := redis.OffshootSelectors()
		if redis.Spec.Mode == api.RedisModeCluster {
			labels = core_util.UpsertMap(labels, map[string]string{
				api.RedisShardKey: statefulSetName[len(redis.BaseNameForShard()):],
			})
		}
		in.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: labels,
		}
		in.Spec.Template.Labels = labels

		in.Spec.Template.Annotations = redis.Spec.PodTemplate.Annotations
		in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(
			in.Spec.Template.Spec.InitContainers,
			redis.Spec.PodTemplate.Spec.InitContainers,
		)

		var (
			ports = []core.ContainerPort{
				{
					Name:          "db",
					ContainerPort: api.RedisNodePort,
					Protocol:      core.ProtocolTCP,
				},
			}
		)
		if redis.Spec.Mode == api.RedisModeCluster {
			ports = append(ports, core.ContainerPort{
				Name:          "gossip",
				ContainerPort: api.RedisGossipPort,
			})
		}
		in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
			Name:            api.ResourceSingularRedis,
			Image:           redisVersion.Spec.DB.Image,
			ImagePullPolicy: core.PullIfNotPresent,
			Args:            redis.Spec.PodTemplate.Spec.Args,
			Ports:           ports,
			Env: []core.EnvVar{
				{
					Name: "POD_IP",
					ValueFrom: &core.EnvVarSource{
						FieldRef: &core.ObjectFieldSelector{
							FieldPath: "status.podIP",
						},
					},
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
				Args: append([]string{
					fmt.Sprintf("--web.listen-address=:%v", redis.Spec.Monitor.Prometheus.Port),
					fmt.Sprintf("--web.telemetry-path=%v", redis.StatsService().Path()),
				}, redis.Spec.Monitor.Args...),
				Image:           redisVersion.Spec.Exporter.Image,
				ImagePullPolicy: core.PullIfNotPresent,
				Ports: []core.ContainerPort{
					{
						Name:          api.PrometheusExporterPortName,
						Protocol:      core.ProtocolTCP,
						ContainerPort: redis.Spec.Monitor.Prometheus.Port,
					},
				},
				Env:             redis.Spec.Monitor.Env,
				Resources:       redis.Spec.Monitor.Resources,
				SecurityContext: redis.Spec.Monitor.SecurityContext,
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

		if c.EnableRBAC {
			in.Spec.Template.Spec.ServiceAccountName = redis.OffshootName()
		}

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
				configPath := filepath.Join(CONFIG_MOUNT_PATH, RedisConfigRelativePath)
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
