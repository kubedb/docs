package controller

import (
	"fmt"

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
	mon_api "kmodules.xyz/monitoring-agent-api/api"
)

func (c *Controller) ensureStatefulSet(mysql *api.MySQL) (kutil.VerbType, error) {
	if err := c.checkStatefulSet(mysql); err != nil {
		return kutil.VerbUnchanged, err
	}

	// Create statefulSet for MySQL database
	statefulSet, vt, err := c.createStatefulSet(mysql)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// Check StatefulSet Pod status
	if vt != kutil.VerbUnchanged {
		if err := c.checkStatefulSetPodStatus(statefulSet); err != nil {
			if ref, rerr := reference.GetReference(clientsetscheme.Scheme, mysql); rerr == nil {
				c.recorder.Eventf(
					ref,
					core.EventTypeWarning,
					eventer.EventReasonFailedToStart,
					`Failed to CreateOrPatch StatefulSet. Reason: %v`,
					err,
				)
			}
			return kutil.VerbUnchanged, err
		}
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, mysql); rerr == nil {
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

func (c *Controller) checkStatefulSet(mysql *api.MySQL) error {
	// SatatefulSet for MySQL database
	statefulSet, err := c.Client.AppsV1().StatefulSets(mysql.Namespace).Get(mysql.OffshootName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindMySQL {
		return fmt.Errorf(`Intended statefulSet "%v" already exists`, mysql.OffshootName())
	}

	return nil
}

func (c *Controller) createStatefulSet(mysql *api.MySQL) (*apps.StatefulSet, kutil.VerbType, error) {
	statefulSetMeta := metav1.ObjectMeta{
		Name:      mysql.OffshootName(),
		Namespace: mysql.Namespace,
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, mysql)
	if rerr != nil {
		return nil, kutil.VerbUnchanged, rerr
	}

	return app_util.CreateOrPatchStatefulSet(c.Client, statefulSetMeta, func(in *apps.StatefulSet) *apps.StatefulSet {
		in.ObjectMeta = core_util.EnsureOwnerReference(in.ObjectMeta, ref)
		in.Labels = core_util.UpsertMap(in.Labels, mysql.StatefulSetLabels())
		in.Annotations = core_util.UpsertMap(in.Annotations, mysql.StatefulSetAnnotations())

		in.Spec.Replicas = types.Int32P(1)
		in.Spec.ServiceName = c.GoverningService
		in.Spec.Template.Labels = in.Labels
		in.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: in.Labels,
		}

		in = upsertInitContainer(in)

		in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
			Name:            api.ResourceSingularMySQL,
			Image:           c.docker.GetImageWithTag(mysql),
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          "db",
					ContainerPort: 3306,
					Protocol:      core.ProtocolTCP,
				},
			},
			Resources: mysql.Spec.Resources,
		})
		if mysql.GetMonitoringVendor() == mon_api.VendorPrometheus {
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
				Name: "exporter",
				Args: append([]string{
					"export",
					fmt.Sprintf("--address=:%d", mysql.Spec.Monitor.Prometheus.Port),
					fmt.Sprintf("--enable-analytics=%v", c.EnableAnalytics),
				}, c.LoggerOptions.ToFlags()...),
				Image: c.docker.GetOperatorImageWithTag(mysql),
				Ports: []core.ContainerPort{
					{
						Name:          api.PrometheusExporterPortName,
						Protocol:      core.ProtocolTCP,
						ContainerPort: mysql.Spec.Monitor.Prometheus.Port,
					},
				},
				VolumeMounts: []core.VolumeMount{
					{
						Name:      "db-secret",
						MountPath: ExporterSecretPath,
						ReadOnly:  true,
					},
				},
			})
			in.Spec.Template.Spec.Volumes = core_util.UpsertVolume(
				in.Spec.Template.Spec.Volumes,
				core.Volume{
					Name: "db-secret",
					VolumeSource: core.VolumeSource{
						Secret: &core.SecretVolumeSource{
							SecretName: mysql.Spec.DatabaseSecret.SecretName,
						},
					},
				},
			)
		}
		// Set Admin Secret as MYSQL_ROOT_PASSWORD env variable
		in = upsertEnv(in, mysql)
		in = upsertDataVolume(in, mysql)
		in = upsertCustomConfig(in, mysql)

		if mysql.Spec.Init != nil && mysql.Spec.Init.ScriptSource != nil {
			in = upsertInitScript(in, mysql.Spec.Init.ScriptSource.VolumeSource)
		}

		in.Spec.Template.Spec.NodeSelector = mysql.Spec.NodeSelector
		in.Spec.Template.Spec.Affinity = mysql.Spec.Affinity
		in.Spec.Template.Spec.Tolerations = mysql.Spec.Tolerations
		in.Spec.Template.Spec.ImagePullSecrets = mysql.Spec.ImagePullSecrets
		if mysql.Spec.SchedulerName != "" {
			in.Spec.Template.Spec.SchedulerName = mysql.Spec.SchedulerName
		}

		in.Spec.UpdateStrategy.Type = apps.RollingUpdateStatefulSetStrategyType
		in = upsertUserEnv(in, mysql)
		return in
	})
}

func upsertInitContainer(statefulSet *apps.StatefulSet) *apps.StatefulSet {
	container := core.Container{
		Name:            "remove-lost-found",
		Image:           "busybox",
		ImagePullPolicy: core.PullIfNotPresent,
		Command: []string{
			"rm",
			"-rf",
			"/var/lib/mysql/lost+found",
		},
		VolumeMounts: []core.VolumeMount{
			{
				Name:      "data",
				MountPath: "/var/lib/mysql",
			},
		},
	}
	initContainers := statefulSet.Spec.Template.Spec.InitContainers
	initContainers = core_util.UpsertContainer(initContainers, container)
	statefulSet.Spec.Template.Spec.InitContainers = initContainers
	return statefulSet
}

func upsertDataVolume(statefulSet *apps.StatefulSet, mysql *api.MySQL) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMySQL {
			volumeMount := core.VolumeMount{
				Name:      "data",
				MountPath: "/var/lib/mysql",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			pvcSpec := mysql.Spec.Storage
			if len(pvcSpec.AccessModes) == 0 {
				pvcSpec.AccessModes = []core.PersistentVolumeAccessMode{
					core.ReadWriteOnce,
				}
				log.Infof(`Using "%v" as AccessModes in mysql.Spec.Storage`, core.ReadWriteOnce)
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

func upsertEnv(statefulSet *apps.StatefulSet, mysql *api.MySQL) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMySQL {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, core.EnvVar{
				Name: "MYSQL_ROOT_PASSWORD",
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: mysql.Spec.DatabaseSecret.SecretName,
						},
						Key: KeyMySQLPassword,
					},
				},
			})
			return statefulSet
		}
	}
	return statefulSet
}

// upsertUserEnv add/overwrite env from user provided env in crd spec
func upsertUserEnv(statefulSet *apps.StatefulSet, mysql *api.MySQL) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMySQL {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, mysql.Spec.Env...)
			return statefulSet
		}
	}
	return statefulSet
}

func upsertInitScript(statefulSet *apps.StatefulSet, script core.VolumeSource) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMySQL {
			volumeMount := core.VolumeMount{
				Name:      "initial-script",
				MountPath: "/docker-entrypoint-initdb.d",
			}
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = core_util.UpsertVolumeMount(
				container.VolumeMounts,
				volumeMount,
			)

			volume := core.Volume{
				Name:         "initial-script",
				VolumeSource: script,
			}
			statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(
				statefulSet.Spec.Template.Spec.Volumes,
				volume,
			)
			return statefulSet
		}
	}
	return statefulSet
}

func (c *Controller) checkStatefulSetPodStatus(statefulSet *apps.StatefulSet) error {
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

func upsertCustomConfig(statefulSet *apps.StatefulSet, mysql *api.MySQL) *apps.StatefulSet {
	if mysql.Spec.ConfigSource != nil {
		for i, container := range statefulSet.Spec.Template.Spec.Containers {
			if container.Name == api.ResourceSingularMySQL {
				configVolumeMount := core.VolumeMount{
					Name:      "custom-config",
					MountPath: "/etc/mysql/conf.d",
				}
				volumeMounts := container.VolumeMounts
				volumeMounts = core_util.UpsertVolumeMount(volumeMounts, configVolumeMount)
				statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

				configVolume := core.Volume{
					Name:         "custom-config",
					VolumeSource: *mysql.Spec.ConfigSource,
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
