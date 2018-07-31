package controller

import (
	"fmt"
	"path/filepath"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	"github.com/appscode/kutil"
	app_util "github.com/appscode/kutil/apps/v1"
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/eventer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	mon_api "kmodules.xyz/monitoring-agent-api/api"
)

func (c *Controller) ensureStatefulSet(
	postgres *api.Postgres,
	envList []core.EnvVar,
) (kutil.VerbType, error) {

	if err := c.checkStatefulSet(postgres); err != nil {
		return kutil.VerbUnchanged, err
	}

	statefulSetMeta := metav1.ObjectMeta{
		Name:      postgres.OffshootName(),
		Namespace: postgres.Namespace,
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, postgres)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	replicas := int32(1)
	if postgres.Spec.Replicas != nil {
		replicas = types.Int32(postgres.Spec.Replicas)
	}

	statefulSet, vt, err := app_util.CreateOrPatchStatefulSet(c.Client, statefulSetMeta, func(in *apps.StatefulSet) *apps.StatefulSet {
		in.ObjectMeta = core_util.EnsureOwnerReference(in.ObjectMeta, ref)

		in = upsertObjectMeta(in, postgres)

		in.Spec.Replicas = types.Int32P(replicas)

		in.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: in.Labels,
		}

		in.Spec.ServiceName = c.GoverningService
		in.Spec.Template.Labels = in.Labels

		in = c.upsertContainer(in, postgres)
		in = upsertEnv(in, postgres, envList)
		in = upsertUserEnv(in, postgres)
		in = upsertPort(in)

		in.Spec.Template.Spec.NodeSelector = postgres.Spec.NodeSelector
		in.Spec.Template.Spec.Affinity = postgres.Spec.Affinity

		if postgres.Spec.SchedulerName != "" {
			in.Spec.Template.Spec.SchedulerName = postgres.Spec.SchedulerName
		}

		in.Spec.Template.Spec.Tolerations = postgres.Spec.Tolerations
		in.Spec.Template.Spec.ImagePullSecrets = postgres.Spec.ImagePullSecrets

		in = c.upsertMonitoringContainer(in, postgres)
		if postgres.Spec.Archiver != nil {
			archiverStorage := postgres.Spec.Archiver.Storage
			if archiverStorage != nil {
				in = upsertArchiveSecret(in, archiverStorage.StorageSecretName)
			}
		}

		if _, err := meta_util.GetString(postgres.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound {
			if postgres.Spec.Init != nil && postgres.Spec.Init.PostgresWAL != nil {
				in = upsertInitWalSecret(in, postgres.Spec.Init.PostgresWAL.StorageSecretName)
			}
			if postgres.Spec.Init != nil && postgres.Spec.Init.ScriptSource != nil {
				in = upsertInitScript(in, postgres.Spec.Init.ScriptSource.VolumeSource)
			}
		}

		in = upsertDataVolume(in, postgres)
		in = upsertCustomConfig(in, postgres)

		if c.EnableRBAC {
			in.Spec.Template.Spec.ServiceAccountName = postgres.OffshootName()
		}

		in.Spec.UpdateStrategy.Type = apps.RollingUpdateStatefulSetStrategyType

		return in
	})

	if err != nil {
		return kutil.VerbUnchanged, err
	}

	if vt == kutil.VerbCreated || vt == kutil.VerbPatched {
		// Check StatefulSet Pod status
		if err := c.CheckStatefulSetPodStatus(statefulSet); err != nil {
			if ref, rerr := reference.GetReference(clientsetscheme.Scheme, postgres); rerr == nil {
				c.recorder.Eventf(
					ref,
					core.EventTypeWarning,
					eventer.EventReasonFailedToStart,
					`Failed to be running after StatefulSet %v. Reason: %v`,
					vt,
					err,
				)
			}
			return kutil.VerbUnchanged, err
		}

		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, postgres); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully %v StatefulSet",
				vt,
			)
		}
	}
	return vt, nil
}

func (c *Controller) CheckStatefulSetPodStatus(statefulSet *apps.StatefulSet) error {
	err := core_util.WaitUntilPodRunningBySelector(
		c.Client,
		statefulSet.Namespace,
		statefulSet.Spec.Selector,
		int(types.Int32(statefulSet.Spec.Replicas)),
	)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) ensureCombinedNode(postgres *api.Postgres) (kutil.VerbType, error) {
	standbyMode := api.WarmStandby
	streamingMode := api.AsynchronousStreaming

	if postgres.Spec.StandbyMode != nil {
		standbyMode = *postgres.Spec.StandbyMode
	}
	if postgres.Spec.StreamingMode != nil {
		streamingMode = *postgres.Spec.StreamingMode
	}

	envList := []core.EnvVar{
		{
			Name:  "STANDBY",
			Value: string(standbyMode),
		},
		{
			Name:  "STREAMING",
			Value: string(streamingMode),
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
						Value: fmt.Sprintf("s3://%v/%v/archive", archiverStorage.S3.Bucket, filepath.Join(archiverStorage.S3.Prefix, api.DatabaseNamePrefix, postgres.Namespace, postgres.Name)),
					},
				}...,
			)
		}
	}

	if postgres.Spec.Init != nil {
		restoreStorage := postgres.Spec.Init.PostgresWAL
		if restoreStorage != nil {
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
	}

	return c.ensureStatefulSet(postgres, envList)
}

func (c *Controller) checkStatefulSet(postgres *api.Postgres) error {
	name := postgres.OffshootName()
	// SatatefulSet for Postgres database
	statefulSet, err := c.Client.AppsV1().StatefulSets(postgres.Namespace).Get(name, metav1.GetOptions{})
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

func (c *Controller) upsertContainer(statefulSet *apps.StatefulSet, postgres *api.Postgres) *apps.StatefulSet {
	container := core.Container{
		Name:  api.ResourceSingularPostgres,
		Image: c.docker.GetImageWithTag(postgres),
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
			Value: postgres.ServiceName(),
		},
		{
			Name: KeyPostgresPassword,
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: postgres.Spec.DatabaseSecret.SecretName,
					},
					Key: KeyPostgresPassword,
				},
			},
		},
	}

	envList = append(envList, envs...)

	// To do this, Upsert Container first
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPostgres {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envList...)
			return statefulSet
		}
	}

	return statefulSet
}

// upsertUserEnv add/overwrite env from user provided env in crd spec
func upsertUserEnv(statefulSet *apps.StatefulSet, postgress *api.Postgres) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPostgres {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, postgress.Spec.Env...)
			return statefulSet
		}
	}
	return statefulSet
}

func upsertPort(statefulSet *apps.StatefulSet) *apps.StatefulSet {
	getPorts := func() []core.ContainerPort {
		portList := []core.ContainerPort{
			{
				Name:          PostgresPortName,
				ContainerPort: PostgresPort,
				Protocol:      core.ProtocolTCP,
			},
		}
		return portList
	}

	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPostgres {
			statefulSet.Spec.Template.Spec.Containers[i].Ports = getPorts()
			return statefulSet
		}
	}

	return statefulSet
}

func (c *Controller) upsertMonitoringContainer(statefulSet *apps.StatefulSet, postgres *api.Postgres) *apps.StatefulSet {
	if postgres.GetMonitoringVendor() == mon_api.VendorPrometheus {
		container := core.Container{
			Name: "exporter",
			Args: append([]string{
				"export",
				fmt.Sprintf("--address=:%d", api.PrometheusExporterPortNumber),
				"--v=3",
				fmt.Sprintf("--enable-analytics=%v", c.EnableAnalytics),
			}, c.LoggerOptions.ToFlags()...),
			Image:           c.docker.GetOperatorImageWithTag(postgres),
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          api.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: int32(api.PrometheusExporterPortNumber),
				},
			},
			VolumeMounts: []core.VolumeMount{
				{
					Name:      "secret",
					MountPath: ExporterSecretPath,
				},
			},
		}
		containers := statefulSet.Spec.Template.Spec.Containers
		containers = core_util.UpsertContainer(containers, container)
		statefulSet.Spec.Template.Spec.Containers = containers

		volume := core.Volume{
			Name: "secret",
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: postgres.Spec.DatabaseSecret.SecretName,
				},
			},
		}
		volumes := statefulSet.Spec.Template.Spec.Volumes
		volumes = core_util.UpsertVolume(volumes, volume)
		statefulSet.Spec.Template.Spec.Volumes = volumes
	}
	return statefulSet
}

func upsertArchiveSecret(statefulSet *apps.StatefulSet, secretName string) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPostgres {
			volumeMount := core.VolumeMount{
				Name:      "wal-g-archive",
				MountPath: "/srv/wal-g/archive/secrets",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			volume := core.Volume{
				Name: "wal-g-archive",
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: secretName,
					},
				},
			}
			volumes := statefulSet.Spec.Template.Spec.Volumes
			volumes = core_util.UpsertVolume(volumes, volume)
			statefulSet.Spec.Template.Spec.Volumes = volumes
			return statefulSet
		}
	}
	return statefulSet
}

func upsertInitWalSecret(statefulSet *apps.StatefulSet, secretName string) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPostgres {
			volumeMount := core.VolumeMount{
				Name:      "wal-g-restore",
				MountPath: "/srv/wal-g/restore/secrets",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			volume := core.Volume{
				Name: "wal-g-restore",
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: secretName,
					},
				},
			}
			volumes := statefulSet.Spec.Template.Spec.Volumes
			volumes = core_util.UpsertVolume(volumes, volume)
			statefulSet.Spec.Template.Spec.Volumes = volumes
			return statefulSet
		}
	}
	return statefulSet
}

func upsertInitScript(statefulSet *apps.StatefulSet, script core.VolumeSource) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPostgres {
			volumeMount := core.VolumeMount{
				Name:      "initial-script",
				MountPath: "/var/initdb",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			volume := core.Volume{
				Name:         "initial-script",
				VolumeSource: script,
			}
			volumes := statefulSet.Spec.Template.Spec.Volumes
			volumes = core_util.UpsertVolume(volumes, volume)
			statefulSet.Spec.Template.Spec.Volumes = volumes
			return statefulSet
		}
	}
	return statefulSet
}

func upsertDataVolume(statefulSet *apps.StatefulSet, postgres *api.Postgres) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPostgres {
			volumeMount := core.VolumeMount{
				Name:      "data",
				MountPath: "/var/pv",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			pvcSpec := postgres.Spec.Storage
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
				Spec: pvcSpec,
			}
			if pvcSpec.StorageClassName != nil {
				volumeClaim.Annotations = map[string]string{
					"volume.beta.kubernetes.io/storage-class": *pvcSpec.StorageClassName,
				}
			}
			volumeClaims := statefulSet.Spec.VolumeClaimTemplates
			volumeClaims = core_util.UpsertVolumeClaim(volumeClaims, volumeClaim)
			statefulSet.Spec.VolumeClaimTemplates = volumeClaims

			break
		}
	}
	return statefulSet
}

func upsertCustomConfig(statefulSet *apps.StatefulSet, postgres *api.Postgres) *apps.StatefulSet {
	if postgres.Spec.ConfigSource != nil {
		for i, container := range statefulSet.Spec.Template.Spec.Containers {
			if container.Name == api.ResourceSingularPostgres {
				configVolumeMount := core.VolumeMount{
					Name:      "custom-config",
					MountPath: "/etc/config",
				}
				volumeMounts := container.VolumeMounts
				volumeMounts = core_util.UpsertVolumeMount(volumeMounts, configVolumeMount)
				statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

				configVolume := core.Volume{
					Name:         "custom-config",
					VolumeSource: *postgres.Spec.ConfigSource,
				}

				volumes := statefulSet.Spec.Template.Spec.Volumes
				volumes = core_util.UpsertVolume(volumes, configVolume)
				statefulSet.Spec.Template.Spec.Volumes = volumes
				break
			}
		}
	}
	return statefulSet
}
