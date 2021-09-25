/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"

	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/pkg/eventer"
	configure_cluster "kubedb.dev/redis/pkg/configure-cluster"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"gomodules.xyz/envsubst"
	"gomodules.xyz/pointer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

const (
	CONFIG_MOUNT_PATH = "/usr/local/etc/redis/"
)

var ErrStsNotReady = fmt.Errorf("statefulSet is not updated yet")

func (c *Controller) ensureStatefulSet(db *api.Redis, statefulSetName string, removeSlave bool) (kutil.VerbType, error) {
	err := c.checkStatefulSet(db, statefulSetName)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// Create statefulSet for Redis database
	statefulSet, vt, err := c.createStatefulSet(db, statefulSetName, removeSlave)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	//ensure PDB for statefulSet
	if err := c.CreateStatefulSetPodDisruptionBudget(statefulSet); err != nil {
		return kutil.VerbUnchanged, err
	}

	// Check StatefulSet Pod status
	if vt != kutil.VerbUnchanged {
		if !app_util.IsStatefulSetReady(statefulSet) {
			return "", ErrStsNotReady
		}

		c.Recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet",
			vt,
		)
	}

	return vt, nil
}

func (c *Controller) ensureRedisNodes(db *api.Redis) (kutil.VerbType, error) {
	var (
		vt  kutil.VerbType
		err error
	)

	if db.Spec.Mode == api.RedisModeStandalone {
		vt, err = c.ensureStatefulSet(db, db.OffshootName(), false)
		if err != nil {
			return vt, err
		}
	} else if db.Spec.Mode == api.RedisModeCluster {
		for i := 0; i < int(*db.Spec.Cluster.Master); i++ {
			vt, err = c.ensureStatefulSet(db, db.StatefulSetNameWithShard(i), false)
			if err != nil {
				return vt, err
			}
		}

		statefulSets, err := c.Client.AppsV1().StatefulSets(db.Namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: labels.Set(db.OffshootSelectors()).String(),
		})
		if err != nil {
			return vt, err
		}

		pods := make([][]*core.Pod, len(statefulSets.Items))
		for i := 0; i < len(statefulSets.Items); i++ {
			stsIndex, _ := strconv.Atoi(statefulSets.Items[i].Name[len(db.BaseNameForShard()):])
			pods[stsIndex] = make([]*core.Pod, *statefulSets.Items[i].Spec.Replicas)
			for j := 0; j < int(*statefulSets.Items[i].Spec.Replicas); j++ {
				podName := fmt.Sprintf("%s-%d", statefulSets.Items[i].Name, j)
				pods[stsIndex][j], err = c.Client.CoreV1().Pods(db.Namespace).Get(context.TODO(), podName, metav1.GetOptions{})
				if err != nil {
					return vt, err
				}
			}
		}

		redisVersion, err := c.DBClient.CatalogV1alpha1().RedisVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
		if err != nil {
			return vt, err
		}
		if err := configure_cluster.ConfigureRedisCluster(c.ClientConfig, db, redisVersion.Spec.Version, pods); err != nil {
			return vt, errors.Wrap(err, "failed to configure required cluster")
		}

		klog.Infoln("Cluster configured")
		klog.Infoln("Checking for removing master(s)...")
		statefulSets, err = c.Client.AppsV1().StatefulSets(db.Namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: labels.Set(db.OffshootSelectors()).String(),
		})
		if err != nil {
			return vt, err
		}
		if len(statefulSets.Items) > int(*db.Spec.Cluster.Master) {
			klog.Infoln("Removing masters...")

			foregroundPolicy := metav1.DeletePropagationForeground
			for i := int(*db.Spec.Cluster.Master); i < len(statefulSets.Items); i++ {
				err = c.Client.AppsV1().
					StatefulSets(db.Namespace).
					Delete(context.TODO(), db.StatefulSetNameWithShard(i), metav1.DeleteOptions{
						PropagationPolicy: &foregroundPolicy,
					})
				if err != nil {
					return vt, err
				}
			}
		}

		klog.Infoln("Checking for removing slave(s)...")

		// update the the statefulSets with reduced replicas as some of their slaves have been
		// removed when redis.spec.cluster.replicas field is reduced
		statefulSets, err = c.Client.AppsV1().StatefulSets(db.Namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: labels.Set(db.OffshootSelectors()).String(),
		})
		if err != nil {
			return vt, err
		}
		for i := 0; i < int(*db.Spec.Cluster.Master); i++ {
			if int(*statefulSets.Items[i].Spec.Replicas) > int(*db.Spec.Cluster.Replicas)+1 {
				klog.Infoln("Removing slaves...")

				vt, err = c.ensureStatefulSet(db, db.StatefulSetNameWithShard(i), true)
				if err != nil {
					return vt, err
				}
			}
		}
	} else if db.Spec.Mode == api.RedisModeSentinel {
		vt, err = c.ensureStatefulSet(db, db.OffshootName(), true)
		if err != nil {
			return vt, err
		}
	}

	return vt, nil
}

func (c *Controller) checkStatefulSet(db *api.Redis, statefulSetName string) error {
	// SatatefulSet for Redis
	statefulSet, err := c.Client.AppsV1().StatefulSets(db.Namespace).Get(context.TODO(), statefulSetName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if statefulSet.Labels[meta_util.NameLabelKey] != db.ResourceFQN() ||
		statefulSet.Labels[meta_util.InstanceLabelKey] != db.Name {
		return fmt.Errorf(`intended statefulSet "%v/%v" already exists`, db.Namespace, db.OffshootName())
	}

	return nil
}

func (c *Controller) createStatefulSet(db *api.Redis, statefulSetName string, removeSlave bool) (*apps.StatefulSet, kutil.VerbType, error) {
	authSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.Spec.AuthSecret.Name, metav1.GetOptions{})
	if err != nil {
		return nil, "", err
	}
	statefulSetMeta := metav1.ObjectMeta{
		Name:      statefulSetName,
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindRedis))

	redisVersion, err := c.DBClient.CatalogV1alpha1().RedisVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	curVersion, err := semver.NewVersion(redisVersion.Spec.Version)
	if err != nil {
		return nil, kutil.VerbUnchanged, fmt.Errorf("can't get the version from RedisVersion spec")
	}
	var (
		affinity      *core.Affinity
		redisModeEnvs []core.EnvVar
	)

	if db.Spec.Mode == api.RedisModeCluster {
		// https://play.golang.org/p/TZPjGg8T0Zn
		re := regexp.MustCompile(fmt.Sprintf(`^%s(\d+)$`, db.BaseNameForShard()))
		matches := re.FindStringSubmatch(statefulSetName)
		if len(matches) == 0 {
			return nil, kutil.VerbUnchanged, fmt.Errorf("failed to detect shard index for statefulset %s", statefulSetName)
		}
		affinity, err = parseAffinityTemplate(db.Spec.PodTemplate.Spec.Affinity.DeepCopy(), matches[1])
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
	} else if db.Spec.Mode == api.RedisModeSentinel {
		sentinel, err := c.DBClient.KubedbV1alpha2().RedisSentinels(db.Spec.SentinelRef.Namespace).Get(context.TODO(), db.Spec.SentinelRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, kutil.VerbUnchanged, err
		}
		sentinelEnvs := []core.EnvVar{
			{
				Name: "SENTINEL_PASSWORD",
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: sentinel.Spec.AuthSecret.Name,
						},
						Key: core.BasicAuthPasswordKey,
					},
				},
			},
			{
				Name:  "SENTINEL_GOVERNING_SERVICE",
				Value: fmt.Sprintf("%s.%s.svc", sentinel.GoverningServiceName(), db.Namespace),
			},
			{
				Name:  "SENTINEL_NAME",
				Value: sentinel.Name,
			},
			{
				Name:  "SENTINEL_REPLICAS",
				Value: strconv.Itoa(int(pointer.Int32(sentinel.Spec.Replicas))),
			},
		}
		redisModeEnvs = core_util.UpsertEnvVars(redisModeEnvs, sentinelEnvs...)
	}

	return app_util.CreateOrPatchStatefulSet(context.TODO(), c.Client, statefulSetMeta, func(in *apps.StatefulSet) *apps.StatefulSet {
		in.Labels = db.OffshootLabels()
		in.Annotations = db.Spec.PodTemplate.Controller.Annotations
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

		if db.Spec.Mode == api.RedisModeStandalone {
			in.Spec.Replicas = pointer.Int32P(1)
		} else if db.Spec.Mode == api.RedisModeCluster {
			// while creating first time, in.Spec.Replicas is 'nil'
			if in.Spec.Replicas == nil ||
				// while adding slave(s), (*in.Spec.Replicas < *redis.Spec.Cluster.Replicas + 1) is true
				*in.Spec.Replicas < *db.Spec.Cluster.Replicas+1 ||
				// removeSlave is true only after deleting slave node(s) in the stage of configuring redis cluster
				removeSlave {
				in.Spec.Replicas = pointer.Int32P(*db.Spec.Cluster.Replicas + 1)
			}
		} else if db.Spec.Mode == api.RedisModeSentinel {
			in.Spec.Replicas = pointer.Int32P(1)
			if db.Spec.Replicas != nil {
				in.Spec.Replicas = pointer.Int32P(*db.Spec.Replicas)
			}
		}
		in.Spec.ServiceName = db.GoverningServiceName()

		labels := db.OffshootSelectors()
		if db.Spec.Mode == api.RedisModeCluster {
			labels = core_util.UpsertMap(labels, map[string]string{
				api.RedisShardKey: statefulSetName[len(db.BaseNameForShard()):],
			})
		}
		in.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: labels,
		}
		in.Spec.Template.Labels = labels

		in.Spec.Template.Annotations = db.Spec.PodTemplate.Annotations
		in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(
			in.Spec.Template.Spec.InitContainers,
			db.Spec.PodTemplate.Spec.InitContainers,
		)

		if db.Spec.Mode == api.RedisModeSentinel {
			in.Spec.Template.Spec.InitContainers = core_util.UpsertContainer(
				in.Spec.Template.Spec.InitContainers,
				getRedisInitContainer(redisVersion))
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers,
				getCoordinatorContainer(redisVersion, curVersion, db, redisModeEnvs))
		}

		//upsert the container
		in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, getRedisContainer(db, redisVersion, curVersion, redisModeEnvs))

		if db.Spec.Monitor != nil && db.Spec.Monitor.Agent.Vendor() == mona.VendorPrometheus {
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, getMonitorContainer(db, redisVersion, curVersion, authSecret))
		}

		in = upsertDataVolume(in, db)
		if db.Spec.Mode == api.RedisModeSentinel {
			in = upsertRedisSentinelVolume(in)
		}

		in.Spec.Template.Spec.NodeSelector = db.Spec.PodTemplate.Spec.NodeSelector
		in.Spec.Template.Spec.Affinity = affinity
		if db.Spec.PodTemplate.Spec.SchedulerName != "" {
			in.Spec.Template.Spec.SchedulerName = db.Spec.PodTemplate.Spec.SchedulerName
		}
		in.Spec.Template.Spec.Tolerations = db.Spec.PodTemplate.Spec.Tolerations
		in.Spec.Template.Spec.ImagePullSecrets = db.Spec.PodTemplate.Spec.ImagePullSecrets
		in.Spec.Template.Spec.PriorityClassName = db.Spec.PodTemplate.Spec.PriorityClassName
		in.Spec.Template.Spec.Priority = db.Spec.PodTemplate.Spec.Priority
		in.Spec.Template.Spec.HostNetwork = db.Spec.PodTemplate.Spec.HostNetwork
		in.Spec.Template.Spec.HostPID = db.Spec.PodTemplate.Spec.HostPID
		in.Spec.Template.Spec.HostIPC = db.Spec.PodTemplate.Spec.HostIPC
		if in.Spec.Template.Spec.SecurityContext == nil {
			in.Spec.Template.Spec.SecurityContext = db.Spec.PodTemplate.Spec.SecurityContext
		}
		in.Spec.Template.Spec.ServiceAccountName = db.Spec.PodTemplate.Spec.ServiceAccountName
		in.Spec.UpdateStrategy = apps.StatefulSetUpdateStrategy{
			Type: apps.OnDeleteStatefulSetStrategyType,
		}
		if in.Spec.Template.Spec.SecurityContext == nil {
			in.Spec.Template.Spec.SecurityContext = db.Spec.PodTemplate.Spec.SecurityContext
		}
		in = upsertUserEnv(in, db)
		in = upsertCustomConfig(in, db)

		// configure tls volume
		in = upsertTLSVolume(in, db)

		return in
	}, metav1.PatchOptions{})
}

func getMonitorContainer(db *api.Redis, redisVersion *v1alpha1.RedisVersion, curVersion *semver.Version, authSecret *core.Secret) core.Container {

	args := []string{
		fmt.Sprintf("--web.listen-address=:%v", db.Spec.Monitor.Prometheus.Exporter.Port),
		fmt.Sprintf("--web.telemetry-path=%v", db.StatsService().Path()),
	}
	if curVersion.Major() > 4 {
		args = append(args, []string{
			fmt.Sprintf("--redis.password=%s", authSecret.Data[core.BasicAuthPasswordKey]),
		}...)
	}
	if db.Spec.TLS != nil {
		tlsArgs := []string{
			"--redis.addr=rediss://localhost:26379",
			"--tls-client-cert-file=/certs/exporter.crt",
			"--tls-client-key-file=/certs/exporter.key",
			"--tls-ca-cert-file=/certs/ca.crt",
		}
		args = append(args, tlsArgs...)
	}
	var volumeMounts []core.VolumeMount
	if db.Spec.TLS != nil {
		volumeMount := core.VolumeMount{
			Name:      "exporter-tls-volume",
			MountPath: "/certs",
		}

		volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
	}

	container := core.Container{
		Name:            "exporter",
		Args:            append(args, db.Spec.Monitor.Prometheus.Exporter.Args...),
		Image:           redisVersion.Spec.Exporter.Image,
		ImagePullPolicy: core.PullIfNotPresent,
		Ports: []core.ContainerPort{
			{
				Name:          mona.PrometheusExporterPortName,
				Protocol:      core.ProtocolTCP,
				ContainerPort: db.Spec.Monitor.Prometheus.Exporter.Port,
			},
		},
		Env:             db.Spec.Monitor.Prometheus.Exporter.Env,
		VolumeMounts:    volumeMounts,
		Resources:       db.Spec.Monitor.Prometheus.Exporter.Resources,
		SecurityContext: db.Spec.Monitor.Prometheus.Exporter.SecurityContext,
	}
	return container
}
func getRedisInitContainer(redisVersion *v1alpha1.RedisVersion) core.Container {
	container := core.Container{
		Name:            "redis-init",
		Image:           redisVersion.Spec.InitContainer.Image,
		ImagePullPolicy: "IfNotPresent",
		VolumeMounts: []core.VolumeMount{
			{
				Name:      api.RedisScriptVolumeName,
				MountPath: api.RedisScriptVolumePath,
			},
		},
	}
	return container
}
func getRedisContainer(db *api.Redis, redisVersion *v1alpha1.RedisVersion, curVersion *semver.Version, redisEnvs []core.EnvVar) core.Container {
	var (
		ports = []core.ContainerPort{
			{
				Name:          api.RedisDatabasePortName,
				ContainerPort: api.RedisDatabasePort,
				Protocol:      core.ProtocolTCP,
			},
		}
	)
	if db.Spec.Mode == api.RedisModeCluster {
		ports = append(ports, core.ContainerPort{
			Name:          api.RedisGossipPortName,
			ContainerPort: api.RedisGossipPort,
			Protocol:      core.ProtocolTCP,
		})
	}
	volumeMounts := []core.VolumeMount{
		{
			Name:      "data",
			MountPath: "/data",
		},
	}

	if db.Spec.Mode == api.RedisModeSentinel {
		volumeMounts = core_util.UpsertVolumeMount(volumeMounts,
			[]core.VolumeMount{
				{
					Name:      api.RedisScriptVolumeName,
					MountPath: api.RedisScriptVolumePath,
				},
			}...)
	}

	if db.Spec.TLS != nil {
		volumeMounts = core_util.UpsertVolumeMount(volumeMounts,
			core.VolumeMount{
				Name:      "tls-volume",
				MountPath: "/certs",
			})
	}

	container := core.Container{
		Name:            api.ResourceSingularRedis,
		Image:           redisVersion.Spec.DB.Image,
		ImagePullPolicy: core.PullIfNotPresent,
		Args:            db.Spec.PodTemplate.Spec.Args,
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
		VolumeMounts:    volumeMounts,
		Resources:       db.Spec.PodTemplate.Spec.Resources,
		SecurityContext: db.Spec.PodTemplate.Spec.ContainerSecurityContext,
		LivenessProbe:   db.Spec.PodTemplate.Spec.LivenessProbe,
		ReadinessProbe:  db.Spec.PodTemplate.Spec.ReadinessProbe,
		Lifecycle:       db.Spec.PodTemplate.Spec.Lifecycle,
	}
	if curVersion.Major() > 4 {
		container.Env = core_util.UpsertEnvVars(container.Env, []core.EnvVar{
			{
				Name: api.EnvRedisPassword,
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: db.Spec.AuthSecret.Name,
						},
						Key: core.BasicAuthPasswordKey,
					},
				},
			},
		}...)
	}
	if db.Spec.Mode == api.RedisModeStandalone {

		args := container.Args
		// for backup redis data
		customArgs := []string{
			"--appendonly yes",
		}
		args = append(args, customArgs...)

		if db.Spec.TLS != nil {
			// tls arguments for redis standalone
			tlsArgs := GetTLSArgs(curVersion, true, api.RedisDatabasePort)
			args = append(args, tlsArgs...)
		}
		container.Args = args

	} else if db.Spec.Mode == api.RedisModeCluster {
		args := container.Args
		// for enabling redis cluster
		customArgs := []string{
			"--cluster-enabled yes",
		}
		args = append(args, customArgs...)
		if db.Spec.TLS != nil {
			// tls arguments for redis cluster
			tlsArgs := GetTLSArgs(curVersion, false, api.RedisDatabasePort)
			tlsArgs = append(tlsArgs, "--tls-cluster yes")
			args = append(args, tlsArgs...)
		}

		container.Args = args
	} else if db.Spec.Mode == api.RedisModeSentinel {
		args := container.Args
		tlsValue := TlsOff
		if db.Spec.TLS != nil {
			tlsValue = "ON"
			// tls arguments for redis cluster
			tlsArgs := GetTLSArgs(curVersion, false, api.RedisDatabasePort)
			args = append(args, tlsArgs...)
		}
		envs := []core.EnvVar{
			{
				Name: "NAMESPACE",
				ValueFrom: &core.EnvVarSource{
					FieldRef: &core.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
			{
				Name:  "TLS",
				Value: tlsValue,
			},
			{
				Name:  "STATEFULSET_NAME",
				Value: db.Name,
			},
			{
				Name:  "REDIS_GOVERNING_SERVICE",
				Value: fmt.Sprintf("%s.%s.svc", db.GoverningServiceName(), db.Namespace),
			},
		}
		container.Env = core_util.UpsertEnvVars(container.Env, envs...)
		container.Command = []string{
			"/scripts/on-start.sh",
		}
		container.Args = args

	}
	container.Env = core_util.UpsertEnvVars(container.Env, redisEnvs...)
	return container
}
func getCoordinatorContainer(redisVersion *v1alpha1.RedisVersion, curVersion *semver.Version, db *api.Redis, redisEnvs []core.EnvVar) core.Container {
	var volumeMounts []core.VolumeMount

	if db.Spec.TLS != nil {
		volumeMounts = core_util.UpsertVolumeMount(volumeMounts,
			core.VolumeMount{
				Name:      "tls-volume",
				MountPath: "/certs",
			})
	}

	container := core.Container{
		Name:  "rd-coordinator",
		Image: redisVersion.Spec.Coordinator.Image,
		Args: []string{
			"run",
		},
		ImagePullPolicy: core.PullIfNotPresent,
		VolumeMounts:    volumeMounts,
		Resources:       db.Spec.Coordinator.Resources,
		SecurityContext: db.Spec.Coordinator.SecurityContext,
	}
	if db.Spec.Mode == api.RedisModeSentinel {
		tlsValue := TlsOff
		if db.Spec.TLS != nil {
			tlsValue = "ON"
		}
		envs := []core.EnvVar{
			{
				Name: "NAMESPACE",
				ValueFrom: &core.EnvVarSource{
					FieldRef: &core.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
			{
				Name:  "REDIS_NAME",
				Value: db.Name,
			},
			{
				Name:  "REDIS_TLS",
				Value: tlsValue,
			},
			{
				Name:  "REDIS_GOVERNING_SERVICE",
				Value: fmt.Sprintf("%s.%s.svc", db.GoverningServiceName(), db.Namespace),
			},
			{
				Name:  "SENTINEL_NAME",
				Value: db.Spec.SentinelRef.Name,
			},
			{
				Name:  "SENTINEL_NAMESPACE",
				Value: db.Spec.SentinelRef.Namespace,
			},
			{
				Name:  "SENTINEL_TLS",
				Value: tlsValue,
			},
			{
				Name:  "SENTINEL_GOVERNING_SERVICE",
				Value: fmt.Sprintf("%s.%s.svc", db.Spec.SentinelRef.Name, db.Spec.SentinelRef.Namespace),
			},
			{
				Name: api.EnvRedisPassword,
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: db.Spec.AuthSecret.Name,
						},
						Key: core.BasicAuthPasswordKey,
					},
				},
			},
		}

		container.Env = core_util.UpsertEnvVars(container.Env, envs...)

	}
	container.Env = core_util.UpsertEnvVars(container.Env, redisEnvs...)

	if curVersion.Major() > 4 {

		container.Env = core_util.UpsertEnvVars(container.Env, []core.EnvVar{
			{
				Name: api.EnvRedisPassword,
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: db.Spec.AuthSecret.Name,
						},
						Key: core.BasicAuthPasswordKey,
					},
				},
			},
		}...)
	}
	return container
}
func upsertDataVolume(statefulSet *apps.StatefulSet, db *api.Redis) *apps.StatefulSet {
	pvcSpec := db.Spec.Storage
	if db.Spec.StorageType == api.StorageTypeEphemeral {
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
			klog.Infof(`Using "%v" as AccessModes in redis.spec.storage`, core.ReadWriteOnce)
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
	return statefulSet
}
func upsertRedisSentinelVolume(statefulSet *apps.StatefulSet) *apps.StatefulSet {
	Volumes := []core.Volume{
		{
			Name: api.RedisScriptVolumeName,
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		},
	}
	statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(statefulSet.Spec.Template.Spec.Volumes, Volumes...)
	return statefulSet
}

// adding tls key , cert and ca-cert
func upsertTLSVolume(sts *apps.StatefulSet, db *api.Redis) *apps.StatefulSet {
	if db.Spec.TLS != nil {
		volume := core.Volume{
			Name: "tls-volume",
			VolumeSource: core.VolumeSource{
				Projected: &core.ProjectedVolumeSource{
					Sources: []core.VolumeProjection{
						{
							Secret: &core.SecretProjection{
								LocalObjectReference: core.LocalObjectReference{
									Name: db.GetCertSecretName(api.RedisServerCert),
								},
								Items: []core.KeyToPath{
									{
										Key:  "ca.crt",
										Path: "ca.crt",
									},
									{
										Key:  "tls.crt",
										Path: "server.crt",
									},
									{
										Key:  "tls.key",
										Path: "server.key",
									},
								},
							},
						},
						{
							Secret: &core.SecretProjection{
								LocalObjectReference: core.LocalObjectReference{
									Name: db.GetCertSecretName(api.RedisClientCert),
								},
								Items: []core.KeyToPath{
									{
										Key:  "tls.crt",
										Path: "client.crt",
									},
									{
										Key:  "tls.key",
										Path: "client.key",
									},
								},
							},
						},
					},
				},
			},
		}

		exporterTLSVolume := core.Volume{
			Name: "exporter-tls-volume",
			VolumeSource: core.VolumeSource{
				Projected: &core.ProjectedVolumeSource{
					Sources: []core.VolumeProjection{
						{
							Secret: &core.SecretProjection{
								LocalObjectReference: core.LocalObjectReference{
									Name: db.GetCertSecretName(api.RedisMetricsExporterCert),
								},
								Items: []core.KeyToPath{
									{
										Key:  "ca.crt",
										Path: "ca.crt",
									},
									{
										Key:  "tls.crt",
										Path: "exporter.crt",
									},
									{
										Key:  "tls.key",
										Path: "exporter.key",
									},
								},
							},
						},
					},
				},
			},
		}

		sts.Spec.Template.Spec.Volumes = core_util.UpsertVolume(
			sts.Spec.Template.Spec.Volumes,
			volume,
			exporterTLSVolume,
		)
	} else {
		sts.Spec.Template.Spec.Volumes = core_util.EnsureVolumeDeleted(sts.Spec.Template.Spec.Volumes, "tls-volume")
		sts.Spec.Template.Spec.Volumes = core_util.EnsureVolumeDeleted(sts.Spec.Template.Spec.Volumes, "exporter-tls-volume")
	}

	return sts
}

// upsertUserEnv add/overwrite env from user provided env in crd spec
func upsertUserEnv(statefulset *apps.StatefulSet, db *api.Redis) *apps.StatefulSet {
	for i, container := range statefulset.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularRedis {
			statefulset.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, db.Spec.PodTemplate.Spec.Env...)
			return statefulset
		}
	}
	return statefulset
}

func upsertCustomConfig(statefulSet *apps.StatefulSet, db *api.Redis) *apps.StatefulSet {
	if db.Spec.ConfigSecret != nil {
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
					Name: "custom-config",
					VolumeSource: core.VolumeSource{
						Secret: &core.SecretVolumeSource{
							SecretName:  db.Spec.ConfigSecret.Name,
							DefaultMode: pointer.Int32P(0777),
						},
					},
				}

				volumes := statefulSet.Spec.Template.Spec.Volumes
				volumes = core_util.UpsertVolume(volumes, configVolume)
				statefulSet.Spec.Template.Spec.Volumes = volumes

				args := statefulSet.Spec.Template.Spec.Containers[i].Args
				if db.Spec.Mode != api.RedisModeSentinel {
					// send custom config file path as argument
					configPath := filepath.Join(CONFIG_MOUNT_PATH, RedisConfigRelativePath)
					if len(args) == 0 || args[0] != configPath {
						args = append([]string{configPath}, args...)
					}
				}

				statefulSet.Spec.Template.Spec.Containers[i].Args = args
				break
			}
		}
	}
	return statefulSet
}

func parseAffinityTemplate(affinity *core.Affinity, shardIndex string) (*core.Affinity, error) {
	if affinity == nil {
		return affinity, nil
	}

	templateMap := map[string]string{
		api.RedisShardAffinityTemplateVar: shardIndex,
	}

	jsonObj, err := json.Marshal(affinity)
	if err != nil {
		return affinity, err
	}

	resolved, err := envsubst.EvalMap(string(jsonObj), templateMap)
	if err != nil {
		return affinity, err
	}

	err = json.Unmarshal([]byte(resolved), affinity)
	return affinity, err
}
