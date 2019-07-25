package controller

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/client-go/tools/analytics"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"
	"kubedb.dev/postgres/pkg/leader_election"
)

func (c *Controller) ensureStatefulSet(
	postgres *api.Postgres,
	postgresVersion *catalog.PostgresVersion,
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
		in.Labels = postgres.OffshootLabels()
		in.Annotations = postgres.Spec.PodTemplate.Controller.Annotations
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)

		in.Spec.Replicas = types.Int32P(replicas)

		in.Spec.ServiceName = c.GoverningService
		in.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: postgres.OffshootSelectors(),
		}
		in.Spec.Template.Labels = postgres.OffshootSelectors()
		in.Spec.Template.Annotations = postgres.Spec.PodTemplate.Annotations
		in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(in.Spec.Template.Spec.InitContainers, postgres.Spec.PodTemplate.Spec.InitContainers)
		in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
			in.Spec.Template.Spec.Containers,
			core.Container{
				Name: api.ResourceSingularPostgres,
				Args: append([]string{
					"leader_election",
					fmt.Sprintf(`--enable-analytics=%v`, c.EnableAnalytics),
				}, c.LoggerOptions.ToFlags()...),
				Env: []core.EnvVar{
					{
						Name:  analytics.Key,
						Value: c.AnalyticsClientID,
					},
				},
				Image:          postgresVersion.Spec.DB.Image,
				Resources:      postgres.Spec.PodTemplate.Spec.Resources,
				LivenessProbe:  postgres.Spec.PodTemplate.Spec.LivenessProbe,
				ReadinessProbe: postgres.Spec.PodTemplate.Spec.ReadinessProbe,
				Lifecycle:      postgres.Spec.PodTemplate.Spec.Lifecycle,
				SecurityContext: &core.SecurityContext{
					Privileged: types.BoolP(false),
					Capabilities: &core.Capabilities{
						Add: []core.Capability{"IPC_LOCK", "SYS_RESOURCE"},
					},
				},
			})
		in = upsertEnv(in, postgres, envList)
		in = upsertUserEnv(in, postgres)
		in = upsertPort(in)

		in.Spec.Template.Spec.NodeSelector = postgres.Spec.PodTemplate.Spec.NodeSelector
		in.Spec.Template.Spec.Affinity = postgres.Spec.PodTemplate.Spec.Affinity
		if postgres.Spec.PodTemplate.Spec.SchedulerName != "" {
			in.Spec.Template.Spec.SchedulerName = postgres.Spec.PodTemplate.Spec.SchedulerName
		}
		in.Spec.Template.Spec.Tolerations = postgres.Spec.PodTemplate.Spec.Tolerations
		in.Spec.Template.Spec.ImagePullSecrets = postgres.Spec.PodTemplate.Spec.ImagePullSecrets
		in.Spec.Template.Spec.PriorityClassName = postgres.Spec.PodTemplate.Spec.PriorityClassName
		in.Spec.Template.Spec.Priority = postgres.Spec.PodTemplate.Spec.Priority
		in.Spec.Template.Spec.SecurityContext = postgres.Spec.PodTemplate.Spec.SecurityContext

		in = c.upsertMonitoringContainer(in, postgres, postgresVersion)
		if postgres.Spec.Archiver != nil {
			if postgres.Spec.Archiver.Storage != nil {
				//Creating secret for cloud providers
				archiverStorage := postgres.Spec.Archiver.Storage
				if archiverStorage.Local == nil {
					in = upsertArchiveSecret(in, archiverStorage.StorageSecretName)
				}
			}
		}

		if _, err := meta_util.GetString(postgres.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound {
			initSource := postgres.Spec.Init
			if initSource != nil && initSource.PostgresWAL != nil && initSource.PostgresWAL.Local == nil {
				//Getting secret for cloud providers
				in = upsertInitWalSecret(in, postgres.Spec.Init.PostgresWAL.StorageSecretName)
			}
			if initSource != nil && initSource.ScriptSource != nil {
				in = upsertInitScript(in, postgres.Spec.Init.ScriptSource.VolumeSource)
			}
		}

		in = upsertShm(in, postgres)
		in = upsertDataVolume(in, postgres)
		in = upsertCustomConfig(in, postgres)

		if c.EnableRBAC {
			in.Spec.Template.Spec.ServiceAccountName = postgres.Spec.PodTemplate.Spec.ServiceAccountName
		}
		in.Spec.UpdateStrategy = postgres.Spec.UpdateStrategy

		return in
	})

	if err != nil {
		return kutil.VerbUnchanged, err
	}

	if vt == kutil.VerbCreated || vt == kutil.VerbPatched {
		// Check StatefulSet Pod status
		if err := c.CheckStatefulSetPodStatus(statefulSet); err != nil {
			return kutil.VerbUnchanged, err
		}

		c.recorder.Eventf(
			postgres,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet",
			vt,
		)
	}

	// ensure pdb
	if err := c.CreateStatefulSetPodDisruptionBudget(statefulSet); err != nil {
		return vt, err
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

func (c *Controller) ensureCombinedNode(postgres *api.Postgres, postgresVersion *catalog.PostgresVersion) (kutil.VerbType, error) {
	standbyMode := api.WarmPostgresStandbyMode
	streamingMode := api.AsynchronousPostgresStreamingMode

	if postgres.Spec.StandbyMode != nil {
		standbyMode = *postgres.Spec.StandbyMode
	}
	if postgres.Spec.StreamingMode != nil {
		streamingMode = *postgres.Spec.StreamingMode
	}

	envList := []core.EnvVar{
		{
			Name:  "STANDBY",
			Value: strings.ToLower(string(standbyMode)),
		},
		{
			Name:  "STREAMING",
			Value: strings.ToLower(string(streamingMode)),
		},
	}

	if postgres.Spec.LeaderElection != nil {
		envList = append(envList, []core.EnvVar{
			{
				Name:  leader_election.LeaseDurationEnv,
				Value: strconv.Itoa(int(postgres.Spec.LeaderElection.LeaseDurationSeconds)),
			},
			{
				Name:  leader_election.RenewDeadlineEnv,
				Value: strconv.Itoa(int(postgres.Spec.LeaderElection.RenewDeadlineSeconds)),
			},
			{
				Name:  leader_election.RetryPeriodEnv,
				Value: strconv.Itoa(int(postgres.Spec.LeaderElection.RetryPeriodSeconds)),
			},
		}...)
	}

	if postgres.Spec.Archiver != nil {
		archiverStorage := postgres.Spec.Archiver.Storage
		if archiverStorage != nil {
			envList = append(envList,
				core.EnvVar{
					Name:  "ARCHIVE",
					Value: "wal-g",
				},
			)
			if archiverStorage.S3 != nil {
				envList = append(envList,
					core.EnvVar{
						Name:  "ARCHIVE_S3_PREFIX",
						Value: fmt.Sprintf("s3://%v/%v", archiverStorage.S3.Bucket, WalDataDir(postgres)),
					},
				)
				if archiverStorage.S3.Endpoint != "" {
					envList = append(envList,
						core.EnvVar{
							Name:  "ARCHIVE_S3_ENDPOINT",
							Value: archiverStorage.S3.Endpoint,
						},
					)
				}
			} else if archiverStorage.GCS != nil {
				envList = append(envList,
					core.EnvVar{
						Name:  "ARCHIVE_GS_PREFIX",
						Value: fmt.Sprintf("gs://%v/%v", archiverStorage.GCS.Bucket, WalDataDir(postgres)),
					},
				)
			} else if archiverStorage.Azure != nil {
				envList = append(envList,
					core.EnvVar{
						Name:  "ARCHIVE_AZ_PREFIX",
						Value: fmt.Sprintf("azure://%v/%v", archiverStorage.Azure.Container, WalDataDir(postgres)),
					},
				)
			} else if archiverStorage.Swift != nil {
				envList = append(envList,
					core.EnvVar{
						Name:  "ARCHIVE_SWIFT_PREFIX",
						Value: fmt.Sprintf("swift://%v/%v", archiverStorage.Swift.Container, WalDataDir(postgres)),
					},
				)
			} else if archiverStorage.Local != nil {
				envList = append(envList,
					core.EnvVar{
						Name:  "ARCHIVE_FILE_PREFIX",
						Value: archiverStorage.Local.MountPath,
					},
				)
			}
		}
	}

	if postgres.Spec.Init != nil {
		wal := postgres.Spec.Init.PostgresWAL
		if wal != nil {
			envList = append(envList, walRecoveryConfig(wal)...)
		}
	}

	return c.ensureStatefulSet(postgres, postgresVersion, envList)
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
		return fmt.Errorf(`intended statefulSet "%v/%v" already exists`, postgres.Namespace, name)
	}

	return nil
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
			Name: PostgresUser,
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: postgres.Spec.DatabaseSecret.SecretName,
					},
					Key: PostgresUser,
				},
			},
		},
		{
			Name: PostgresPassword,
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: postgres.Spec.DatabaseSecret.SecretName,
					},
					Key: PostgresPassword,
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
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, postgress.Spec.PodTemplate.Spec.Env...)
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

func (c *Controller) upsertMonitoringContainer(statefulSet *apps.StatefulSet, postgres *api.Postgres, postgresVersion *catalog.PostgresVersion) *apps.StatefulSet {
	if postgres.GetMonitoringVendor() == mona.VendorPrometheus {
		container := core.Container{
			Name: "exporter",
			Args: append([]string{
				"--log.level=info",
			}, postgres.Spec.Monitor.Args...),
			Image:           postgresVersion.Spec.Exporter.Image,
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          api.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: int32(api.PrometheusExporterPortNumber),
				},
			},
			Env:             postgres.Spec.Monitor.Env,
			Resources:       postgres.Spec.Monitor.Resources,
			SecurityContext: postgres.Spec.Monitor.SecurityContext,
		}

		envList := []core.EnvVar{
			{
				Name:  "DATA_SOURCE_URI",
				Value: fmt.Sprintf("localhost:%d/?sslmode=disable", PostgresPort),
			},
			{
				Name: "DATA_SOURCE_USER",
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: postgres.Spec.DatabaseSecret.SecretName,
						},
						Key: PostgresUser,
					},
				},
			},
			{
				Name: "DATA_SOURCE_PASS",
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: postgres.Spec.DatabaseSecret.SecretName,
						},
						Key: PostgresPassword,
					},
				},
			},
			{
				Name:  "PG_EXPORTER_WEB_LISTEN_ADDRESS",
				Value: fmt.Sprintf(":%d", api.PrometheusExporterPortNumber),
			},
			{
				Name:  "PG_EXPORTER_WEB_TELEMETRY_PATH",
				Value: postgres.StatsService().Path(),
			},
		}

		container.Env = core_util.UpsertEnvVars(container.Env, envList...)
		containers := statefulSet.Spec.Template.Spec.Containers
		containers = core_util.UpsertContainer(containers, container)
		statefulSet.Spec.Template.Spec.Containers = containers
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

func upsertShm(statefulSet *apps.StatefulSet, postgress *api.Postgres) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPostgres {
			volumeMount := core.VolumeMount{
				Name:      "shared-memory",
				MountPath: "/dev/shm",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			configVolume := core.Volume{
				Name: "shared-memory",
				VolumeSource: core.VolumeSource{
					EmptyDir: &core.EmptyDirVolumeSource{
						Medium: core.StorageMediumMemory,
					},
				},
			}
			volumes := statefulSet.Spec.Template.Spec.Volumes
			volumes = core_util.UpsertVolume(volumes, configVolume)
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
	if postgres.Spec.Archiver != nil || postgres.Spec.Init != nil {
		// Add a PV
		if postgres.Spec.Archiver != nil &&
			postgres.Spec.Archiver.Storage != nil &&
			postgres.Spec.Archiver.Storage.Local != nil {
			pgLocalVol := postgres.Spec.Archiver.Storage.Local
			podSpec := statefulSet.Spec.Template.Spec
			if pgLocalVol != nil {
				volume := core.Volume{
					Name:         "local-archive",
					VolumeSource: pgLocalVol.VolumeSource,
				}
				statefulSet.Spec.Template.Spec.Volumes = append(podSpec.Volumes, volume)

				statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, core.VolumeMount{
					Name:      "local-archive",
					MountPath: pgLocalVol.MountPath,
					//SubPath:  use of SubPath is discouraged
					// due to the contrasting natures of PV claim and wal-g directory
				})
			}
		}
		if postgres.Spec.Init != nil &&
			postgres.Spec.Init.PostgresWAL != nil &&
			postgres.Spec.Init.PostgresWAL.Local != nil {
			pgLocalVol := postgres.Spec.Init.PostgresWAL.Local
			podSpec := statefulSet.Spec.Template.Spec
			if pgLocalVol != nil {
				volume := core.Volume{
					Name:         "local-init",
					VolumeSource: pgLocalVol.VolumeSource,
				}
				statefulSet.Spec.Template.Spec.Volumes = append(podSpec.Volumes, volume)

				statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, core.VolumeMount{
					Name:      "local-init",
					MountPath: pgLocalVol.MountPath,
					//SubPath: is used to locate existing archive
					//from given mountPath, therefore isn't mounted.
				})
			}
		}
	}

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
			if postgres.Spec.StorageType == api.StorageTypeEphemeral {
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
					log.Infof(`Using "%v" as AccessModes in postgres.Spec.Storage`, core.ReadWriteOnce)
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

func walRecoveryConfig(wal *api.PostgresWALSourceSpec) []core.EnvVar {
	envList := []core.EnvVar{
		{
			Name:  "RESTORE",
			Value: "true",
		},
	}

	if wal.S3 != nil {
		envList = append(envList,
			core.EnvVar{
				Name:  "RESTORE_S3_PREFIX",
				Value: fmt.Sprintf("s3://%v/%v", wal.S3.Bucket, wal.S3.Prefix),
			},
		)
		if wal.S3.Endpoint != "" {
			envList = append(envList,
				core.EnvVar{
					Name:  "RESTORE_S3_ENDPOINT",
					Value: wal.S3.Endpoint,
				},
			)
		}
	} else if wal.GCS != nil {
		envList = append(envList,
			core.EnvVar{
				Name:  "RESTORE_GS_PREFIX",
				Value: fmt.Sprintf("gs://%v/%v", wal.GCS.Bucket, wal.GCS.Prefix),
			},
		)
	} else if wal.Azure != nil {
		envList = append(envList,
			core.EnvVar{
				Name:  "RESTORE_AZ_PREFIX",
				Value: fmt.Sprintf("azure://%v/%v", wal.Azure.Container, wal.Azure.Prefix),
			},
		)
	} else if wal.Swift != nil {
		envList = append(envList,
			core.EnvVar{
				Name:  "RESTORE_SWIFT_PREFIX",
				Value: fmt.Sprintf("swift://%v/%v", wal.Swift.Container, wal.Swift.Prefix),
			},
		)
	} else if wal.Local != nil {
		archiveSource := path.Join("/", wal.Local.MountPath, wal.Local.SubPath)
		envList = append(envList,
			core.EnvVar{
				Name:  "RESTORE_FILE_PREFIX",
				Value: archiveSource,
			},
		)
	}

	if wal.PITR != nil {
		envList = append(envList,
			[]core.EnvVar{
				{
					Name:  "PITR",
					Value: "true",
				},
				{
					Name:  "TARGET_INCLUSIVE",
					Value: fmt.Sprintf("%t", *wal.PITR.TargetInclusive),
				},
			}...)
		if wal.PITR.TargetTime != "" {
			envList = append(envList,
				[]core.EnvVar{
					{
						Name:  "TARGET_TIME",
						Value: wal.PITR.TargetTime,
					},
				}...)
		}
		if wal.PITR.TargetTimeline != "" {
			envList = append(envList,
				[]core.EnvVar{
					{
						Name:  "TARGET_TIMELINE",
						Value: wal.PITR.TargetTimeline,
					},
				}...)
		}
		if wal.PITR.TargetXID != "" {
			envList = append(envList,
				[]core.EnvVar{
					{
						Name:  "TARGET_XID",
						Value: wal.PITR.TargetXID,
					},
				}...)
		}
	}
	return envList
}
