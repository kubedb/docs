package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	app_util "github.com/appscode/kutil/apps/v1beta1"
	core_util "github.com/appscode/kutil/core/v1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/docker"
	"github.com/kubedb/apimachinery/pkg/eventer"
	apps "k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) ensureStatefulSet(
	postgres *api.Postgres,
	envList []core.EnvVar,
) error {

	if err := c.checkStatefulSet(postgres); err != nil {
		return err
	}

	statefulSetMeta := metav1.ObjectMeta{
		Name:      postgres.OffshootName(),
		Namespace: postgres.Namespace,
	}

	replicas := postgres.Spec.Replicas
	if replicas < 0 {
		replicas = 0
	}

	statefulSet, err := app_util.CreateOrPatchStatefulSet(c.Client, statefulSetMeta, func(in *apps.StatefulSet) *apps.StatefulSet {
		in = upsertObjectMeta(in, postgres)

		in.Spec.Replicas = types.Int32P(replicas)
		in.Spec.ServiceName = c.opt.GoverningService
		in.Spec.Template = core.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: in.ObjectMeta.Labels,
			},
		}

		in = upsertContainer(in, postgres)
		in = upsertEnv(in, postgres, envList)
		in = upsertPort(in)

		in.Spec.Template.Spec.NodeSelector = postgres.Spec.NodeSelector
		in.Spec.Template.Spec.Affinity = postgres.Spec.Affinity
		in.Spec.Template.Spec.SchedulerName = postgres.Spec.SchedulerName
		in.Spec.Template.Spec.Tolerations = postgres.Spec.Tolerations

		in = upsertMonitoringContainer(in, postgres, c.opt.ExporterTag)
		in = upsertDatabaseSecret(in, postgres.Spec.DatabaseSecret.SecretName)
		if postgres.Spec.Archiver != nil {
			archiverStorage := postgres.Spec.Archiver.Storage
			if archiverStorage != nil {
				in = upsertArchiveSecret(in, archiverStorage.StorageSecretName)
			}
		}

		if postgres.Spec.Init != nil && postgres.Spec.Init.PostgresWAL != nil {
			in = upsertInitWalSecret(in, postgres.Spec.Init.PostgresWAL.StorageSecretName)
		}
		if postgres.Spec.Init != nil && postgres.Spec.Init.ScriptSource != nil {
			in = upsertInitScript(in, postgres.Spec.Init.ScriptSource.VolumeSource)
		}

		in = upsertDataVolume(in, postgres)

		if c.opt.EnableRbac {
			in.Spec.Template.Spec.ServiceAccountName = postgres.Name
		}

		in.Spec.UpdateStrategy.Type = apps.RollingUpdateStatefulSetStrategyType

		return in
	})

	if err != nil {
		return err
	}

	if replicas > 0 {
		// Check StatefulSet Pod status
		if err := c.CheckStatefulSetPodStatus(statefulSet, durationCheckStatefulSet); err != nil {
			c.recorder.Eventf(
				postgres.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToStart,
				"Failed to create StatefulSet. Reason: %v",
				err,
			)

			return err
		} else {
			c.recorder.Event(
				postgres.ObjectReference(),
				core.EventTypeNormal,
				eventer.EventReasonSuccessfulCreate,
				"Successfully created StatefulSet",
			)
		}
	}

	return nil
}

func (c *Controller) ensureCombinedNode(postgres *api.Postgres) error {
	standby := postgres.Spec.Standby
	streaming := postgres.Spec.Streaming
	if standby == "" {
		standby = "warm"
	}
	if streaming == "" {
		streaming = "asynchronous"
	}

	envList := []core.EnvVar{
		{
			Name:  "STANDBY",
			Value: string(standby),
		},
		{
			Name:  "STREAMING",
			Value: string(streaming),
		},
	}

	if postgres.Spec.Archiver != nil {
		archiverStorage := postgres.Spec.Archiver.Storage
		if archiverStorage != nil {
			envList = append(envList,
				[]core.EnvVar{
					{
						Name:  "ARCHIVE",
						Value: "wal-g",
					},
					{
						Name:  "ARCHIVE_S3_PREFIX",
						Value: fmt.Sprintf("s3://%v/%v", archiverStorage.S3.Bucket, archiverStorage.S3.Prefix),
					},
				}...,
			)
		}
	}

	if postgres.Spec.Init != nil && postgres.Spec.Init.PostgresWAL != nil {
		wal := postgres.Spec.Init.PostgresWAL
		envList = append(envList,
			[]core.EnvVar{
				{
					Name:  "RESTORE",
					Value: "true",
				},
				{
					Name:  "RESTORE_S3_PREFIX",
					Value: fmt.Sprintf("s3://%v/%v", wal.S3.Bucket, wal.S3.Prefix),
				},
			}...,
		)
	}

	return c.ensureStatefulSet(postgres, envList)
}

func (c *Controller) checkStatefulSet(postgres *api.Postgres) error {
	name := postgres.OffshootName()
	// SatatefulSet for Postgres database
	statefulSet, err := c.Client.AppsV1beta1().StatefulSets(postgres.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		} else {
			return err
		}
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindPostgres ||
		statefulSet.Labels[api.LabelDatabaseName] != name {
		return fmt.Errorf(`intended statefulSet "%v" already exists`, name)
	}

	return nil
}

func upsertObjectMeta(statefulSet *apps.StatefulSet, postgres *api.Postgres) *apps.StatefulSet {
	statefulSet.Labels = core_util.UpsertMap(statefulSet.Labels, postgres.StatefulSetLabels())
	statefulSet.Annotations = core_util.UpsertMap(statefulSet.Annotations, postgres.StatefulSetAnnotations())
	return statefulSet
}

func upsertContainer(statefulSet *apps.StatefulSet, postgres *api.Postgres) *apps.StatefulSet {
	container := core.Container{
		Name:            api.ResourceNamePostgres,
		Image:           fmt.Sprintf("%v:%v-db", docker.ImagePostgres, postgres.Spec.Version),
		ImagePullPolicy: core.PullIfNotPresent,
		SecurityContext: &core.SecurityContext{
			Privileged: types.BoolP(false),
			Capabilities: &core.Capabilities{
				Add: []core.Capability{"IPC_LOCK", "SYS_RESOURCE"},
			},
		},
	}
	containers := statefulSet.Spec.Template.Spec.Containers
	containers = core_util.UpsertContainer(containers, container)
	statefulSet.Spec.Template.Spec.Containers = containers
	return statefulSet
}

func upsertEnv(statefulSet *apps.StatefulSet, postgres *api.Postgres, envs []core.EnvVar) *apps.StatefulSet {

	envList := []core.EnvVar{
		{
			Name: "NAMESPACE",
			ValueFrom: &core.EnvVarSource{
				FieldRef: &core.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name:  "PRIMARY_HOST",
			Value: postgres.PrimaryName(),
		},
	}

	envList = append(envList, envs...)

	// To do this, Upsert Container first
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceNamePostgres {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envList...)
			return statefulSet
		}
	}

	return statefulSet
}

func upsertPort(statefulSet *apps.StatefulSet) *apps.StatefulSet {
	getPorts := func() []core.ContainerPort {
		portList := []core.ContainerPort{
			{
				Name:          "api",
				ContainerPort: 5432,
				Protocol:      core.ProtocolTCP,
			},
		}
		return portList
	}

	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceNamePostgres {
			statefulSet.Spec.Template.Spec.Containers[i].Ports = getPorts()
			return statefulSet
		}
	}

	return statefulSet
}

func upsertMonitoringContainer(statefulSet *apps.StatefulSet, postgres *api.Postgres, tag string) *apps.StatefulSet {
	if postgres.Spec.Monitor != nil &&
		postgres.Spec.Monitor.Agent == api.AgentCoreosPrometheus &&
		postgres.Spec.Monitor.Prometheus != nil {
		container := core.Container{
			Name: "exporter",
			Args: []string{
				"export",
				fmt.Sprintf("--address=:%d", api.PrometheusExporterPortNumber),
				"--v=3",
			},
			Image:           docker.ImageOperator + ":" + tag,
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          api.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: int32(api.PrometheusExporterPortNumber),
				},
			},
		}
		containers := statefulSet.Spec.Template.Spec.Containers
		containers = core_util.UpsertContainer(containers, container)
		statefulSet.Spec.Template.Spec.Containers = containers
	}
	return statefulSet
}

func upsertDatabaseSecret(statefulset *apps.StatefulSet, secretName string) *apps.StatefulSet {
	for i, container := range statefulset.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceNamePostgres {
			volumeMount := core.VolumeMount{
				Name:      "secret",
				MountPath: "/srv/" + api.ResourceNamePostgres + "/secrets",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulset.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			volume := core.Volume{
				Name: "secret",
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: secretName,
					},
				},
			}
			volumes := statefulset.Spec.Template.Spec.Volumes
			volumes = core_util.UpsertVolume(volumes, volume)
			statefulset.Spec.Template.Spec.Volumes = volumes
			return statefulset
		}
	}
	return statefulset
}

func upsertArchiveSecret(statefulset *apps.StatefulSet, secretName string) *apps.StatefulSet {
	for i, container := range statefulset.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceNamePostgres {
			volumeMount := core.VolumeMount{
				Name:      "wal-g-archive",
				MountPath: "/srv/wal-g/archive/secrets",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulset.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			volume := core.Volume{
				Name: "wal-g-archive",
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: secretName,
					},
				},
			}
			volumes := statefulset.Spec.Template.Spec.Volumes
			volumes = core_util.UpsertVolume(volumes, volume)
			statefulset.Spec.Template.Spec.Volumes = volumes
			return statefulset
		}
	}
	return statefulset
}

func upsertInitWalSecret(statefulset *apps.StatefulSet, secretName string) *apps.StatefulSet {
	for i, container := range statefulset.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceNamePostgres {
			volumeMount := core.VolumeMount{
				Name:      "wal-g-restore",
				MountPath: "/srv/wal-g/restore/secrets",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulset.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			volume := core.Volume{
				Name: "wal-g-restore",
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: secretName,
					},
				},
			}
			volumes := statefulset.Spec.Template.Spec.Volumes
			volumes = core_util.UpsertVolume(volumes, volume)
			statefulset.Spec.Template.Spec.Volumes = volumes
			return statefulset
		}
	}
	return statefulset
}

func upsertInitScript(statefulset *apps.StatefulSet, script core.VolumeSource) *apps.StatefulSet {
	for i, container := range statefulset.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceNamePostgres {
			volumeMount := core.VolumeMount{
				Name:      "initial-script",
				MountPath: "/var/initdb",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulset.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			volume := core.Volume{
				Name:         "initial-script",
				VolumeSource: script,
			}
			volumes := statefulset.Spec.Template.Spec.Volumes
			volumes = core_util.UpsertVolume(volumes, volume)
			statefulset.Spec.Template.Spec.Volumes = volumes
			return statefulset
		}
	}
	return statefulset
}

func upsertDataVolume(statefulSet *apps.StatefulSet, postgres *api.Postgres) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceNamePostgres {
			volumeMount := core.VolumeMount{
				Name:      "data",
				MountPath: "/var/pv",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			pvcSpec := postgres.Spec.Storage
			if pvcSpec != nil {
				if len(pvcSpec.AccessModes) == 0 {
					pvcSpec.AccessModes = []core.PersistentVolumeAccessMode{
						core.ReadWriteOnce,
					}
					log.Infof(`Using "%v" as AccessModes in postgres.Spec.Storage`, core.ReadWriteOnce)
				}

				volumeClaim := core.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: *pvcSpec,
				}
				if pvcSpec.StorageClassName != nil {
					volumeClaim.Annotations = map[string]string{
						"volume.beta.kubernetes.io/storage-class": *pvcSpec.StorageClassName,
					}
				}
				volumeClaims := statefulSet.Spec.VolumeClaimTemplates
				volumeClaims = core_util.UpsertVolumeClaim(volumeClaims, volumeClaim)
				statefulSet.Spec.VolumeClaimTemplates = volumeClaims
			} else {
				volume := core.Volume{
					Name: "data",
					VolumeSource: core.VolumeSource{
						EmptyDir: &core.EmptyDirVolumeSource{},
					},
				}
				volumes := statefulSet.Spec.Template.Spec.Volumes
				volumes = core_util.UpsertVolume(volumes, volume)
				statefulSet.Spec.Template.Spec.Volumes = volumes
				return statefulSet
			}
			break
		}
	}
	return statefulSet
}
