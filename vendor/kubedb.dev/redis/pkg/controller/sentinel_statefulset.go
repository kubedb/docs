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
	"fmt"

	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/Masterminds/semver/v3"
	"gomodules.xyz/pointer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

const TlsOff = "OFF"

func (c *Controller) ensureSentinelStatefulSet(db *api.RedisSentinel) (kutil.VerbType, error) {
	if err := c.checkSentinelStatefulSet(db); err != nil {
		return kutil.VerbUnchanged, err
	}

	// Create statefulSet for Redis Sentinel database
	statefulSet, vt, err := c.createSentinelStatefulSet(db)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// ensure PDB for statefulSet
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
			"Successfully %v Sentinel StatefulSet",
			vt,
		)
	}

	return vt, nil
}

func (c *Controller) checkSentinelStatefulSet(db *api.RedisSentinel) error {
	name := db.OffshootName()
	// SatatefulSet for Sentinel
	statefulSet, err := c.Client.AppsV1().StatefulSets(db.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		} else {
			return err
		}
	}

	if statefulSet.Labels[meta_util.NameLabelKey] != db.ResourceFQN() ||
		statefulSet.Labels[meta_util.InstanceLabelKey] != name {
		return fmt.Errorf(`intended statefulSet "%v/%v" already exists`, db.Namespace, name)
	}

	return nil
}

func (c *Controller) createSentinelStatefulSet(db *api.RedisSentinel) (*apps.StatefulSet, kutil.VerbType, error) {
	var authSecret *core.Secret
	var err error
	if !db.Spec.DisableAuth {
		authSecret, err = c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.Spec.AuthSecret.Name, metav1.GetOptions{})
		if err != nil {
			return nil, "", err
		}
	}
	statefulSetMeta := metav1.ObjectMeta{
		Name:      db.Name,
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindRedisSentinel))

	redisVersion, curVersion, err := c.getRedisSentinelVersion(db)
	if err != nil {
		return nil, kutil.VerbUnchanged, fmt.Errorf("can't get the version from RedisVersion spec")
	}
	var affinity *core.Affinity

	replicas := int32(1)
	if db.Spec.Replicas != nil {
		replicas = pointer.Int32(db.Spec.Replicas)
	}

	return app_util.CreateOrPatchStatefulSet(context.TODO(), c.Client, statefulSetMeta, func(in *apps.StatefulSet) *apps.StatefulSet {
		in.Labels = db.PodControllerLabels()
		in.Annotations = db.Spec.PodTemplate.Controller.Annotations
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

		in.Spec.Replicas = pointer.Int32P(replicas)
		in.Spec.ServiceName = db.GoverningServiceName()

		in.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: db.OffshootSelectors(),
		}
		in.Spec.Template.Labels = db.PodLabels()

		in.Spec.Template.Annotations = db.Spec.PodTemplate.Annotations
		in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(
			in.Spec.Template.Spec.InitContainers,
			db.Spec.PodTemplate.Spec.InitContainers,
		)
		in.Spec.Template.Spec.InitContainers = core_util.UpsertContainer(in.Spec.Template.Spec.InitContainers, upsertSentinelInitContainer(redisVersion))

		// upsert the container
		in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, upsertSentinelContainer(redisVersion, db, curVersion))

		if db.Spec.Monitor != nil && db.Spec.Monitor.Agent.Vendor() == mona.VendorPrometheus {
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, upsertSentinelMonitorContainer(redisVersion, db, authSecret))
		}

		in = upsertSentinelDataVolume(in, db)
		in = upsertSentinelVolume(in)

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
		in = upsertSentinelUserEnv(in, db)

		// configure tls volume
		in = upsertSentinelTLSVolume(in, db)

		return in
	}, metav1.PatchOptions{})
}

// upsertUserEnv add/overwrite env from user provided env in crd spec
func upsertSentinelUserEnv(statefulset *apps.StatefulSet, db *api.RedisSentinel) *apps.StatefulSet {
	for i, container := range statefulset.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularRedisSentinel {
			statefulset.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, db.Spec.PodTemplate.Spec.Env...)
			return statefulset
		}
	}
	return statefulset
}

func upsertSentinelInitContainer(redisVersion *v1alpha1.RedisVersion) core.Container {
	container := core.Container{
		Name:            "sentinel-init",
		Image:           redisVersion.Spec.InitContainer.Image,
		ImagePullPolicy: core.PullIfNotPresent,
		VolumeMounts: []core.VolumeMount{
			{
				Name:      api.RedisScriptVolumeName,
				MountPath: api.RedisScriptVolumePath,
			},
		},
	}
	return container
}

func upsertSentinelContainer(redisVersion *v1alpha1.RedisVersion, db *api.RedisSentinel, curVersion *semver.Version) core.Container {
	ports := []core.ContainerPort{
		{
			Name:          api.RedisSentinelPortName,
			ContainerPort: api.RedisSentinelPort,
			Protocol:      core.ProtocolTCP,
		},
	}
	tlsValue := TlsOff
	if db.Spec.TLS != nil {
		tlsValue = "ON"
	}

	volumeMounts := []core.VolumeMount{
		{
			Name:      "data",
			MountPath: "/data",
		},
		{
			Name:      api.RedisScriptVolumeName,
			MountPath: api.RedisScriptVolumePath,
		},
	}

	if db.Spec.TLS != nil {
		volumeMounts = core_util.UpsertVolumeMount(volumeMounts,
			core.VolumeMount{
				Name:      "tls-volume",
				MountPath: "/certs",
			})
	}

	container := core.Container{
		Name:            api.ResourceSingularRedisSentinel,
		Image:           redisVersion.Spec.DB.Image,
		ImagePullPolicy: core.PullIfNotPresent,
		Command: []string{
			"/scripts/sentinel.sh",
		},
		Args:  db.Spec.PodTemplate.Spec.Args,
		Ports: ports,
		Env: []core.EnvVar{
			{
				Name: "NAMESPACE",
				ValueFrom: &core.EnvVarSource{
					FieldRef: &core.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
			{
				Name:  "GOVERNING_SERVICE",
				Value: fmt.Sprintf("%s.%s.svc", db.GoverningServiceName(), db.Namespace),
			},
			{
				Name:  "TLS",
				Value: tlsValue,
			},
		},
		VolumeMounts: volumeMounts,
		Resources:    db.Spec.PodTemplate.Spec.Resources,
		SecurityContext: &core.SecurityContext{
			RunAsUser: pointer.Int64P(0),
		},
		LivenessProbe:  db.Spec.PodTemplate.Spec.LivenessProbe,
		ReadinessProbe: db.Spec.PodTemplate.Spec.ReadinessProbe,
		Lifecycle:      db.Spec.PodTemplate.Spec.Lifecycle,
	}
	if !db.Spec.DisableAuth {
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
	args := container.Args
	if db.Spec.TLS != nil {
		// tls arguments for redis cluster
		tlsArgs := GetTLSArgs(curVersion, false, api.RedisSentinelPort)
		args = append(args, tlsArgs...)
	}
	container.Args = args
	return container
}

func upsertSentinelMonitorContainer(redisVersion *v1alpha1.RedisVersion, db *api.RedisSentinel, authSecret *core.Secret) core.Container {
	args := []string{
		fmt.Sprintf("--web.listen-address=:%v", db.Spec.Monitor.Prometheus.Exporter.Port),
		fmt.Sprintf("--web.telemetry-path=%v", db.StatsService().Path()),
	}
	if !db.Spec.DisableAuth && authSecret != nil {
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
	} else {
		nonTLSArgs := []string{
			"--redis.addr=redis://localhost:26379",
		}
		args = append(args, nonTLSArgs...)
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

func upsertSentinelVolume(statefulSet *apps.StatefulSet) *apps.StatefulSet {
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

func upsertSentinelDataVolume(statefulSet *apps.StatefulSet, db *api.RedisSentinel) *apps.StatefulSet {
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
			klog.Infof(`Using "%v" as AccessModes in redisSentinel.spec.storage`, core.ReadWriteOnce)
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

// adding tls key , cert and ca-cert
func upsertSentinelTLSVolume(sts *apps.StatefulSet, db *api.RedisSentinel) *apps.StatefulSet {
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
