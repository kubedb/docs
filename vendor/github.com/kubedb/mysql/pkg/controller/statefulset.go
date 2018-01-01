package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	"github.com/appscode/kutil"
	app_util "github.com/appscode/kutil/apps/v1beta1"
	core_util "github.com/appscode/kutil/core/v1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	apps "k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) ensureStatefulSet(mysql *api.MySQL) (kutil.VerbType, error) {
	if err := c.checkStatefulSet(mysql); err != nil {
		return kutil.VerbUnchanged, err
	}

	if err := c.ensureDatabaseSecret(mysql); err != nil {
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
			c.recorder.Eventf(
				mysql.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToStart,
				`Failed to CreateOrPatch StatefulSet. Reason: %v`,
				err,
			)
			return kutil.VerbUnchanged, err
		}
		c.recorder.Eventf(
			mysql.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet",
			vt,
		)

		ms, _, err := util.PatchMySQL(c.ExtClient, mysql, func(in *api.MySQL) *api.MySQL {
			in.Status.Phase = api.DatabasePhaseRunning
			return in
		})
		if err != nil {
			c.recorder.Eventf(
				mysql,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return kutil.VerbUnchanged, err
		}
		mysql.Status = ms.Status
	}
	return vt, nil
}

func (c *Controller) checkStatefulSet(mysql *api.MySQL) error {
	// SatatefulSet for MySQL database
	statefulSet, err := c.Client.AppsV1beta1().StatefulSets(mysql.Namespace).Get(mysql.OffshootName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		} else {
			return err
		}
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
	return app_util.CreateOrPatchStatefulSet(c.Client, statefulSetMeta, func(in *apps.StatefulSet) *apps.StatefulSet {
		in.Labels = core_util.UpsertMap(in.Labels, mysql.StatefulSetLabels())
		in.Annotations = core_util.UpsertMap(in.Annotations, mysql.StatefulSetAnnotations())

		in.Spec.Replicas = types.Int32P(1)
		in.Spec.Template = core.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: in.ObjectMeta.Labels,
			},
		}

		in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
			Name:            api.ResourceNameMySQL,
			Image:           c.opt.Docker.GetImageWithTag(mysql),
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
		if mysql.Spec.Monitor != nil &&
			mysql.Spec.Monitor.Agent == api.AgentCoreosPrometheus &&
			mysql.Spec.Monitor.Prometheus != nil {
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
				Name: "exporter",
				Args: []string{
					"export",
					fmt.Sprintf("--address=:%d", mysql.Spec.Monitor.Prometheus.Port),
					"--v=3",
				},
				Image:           c.opt.Docker.GetOperatorImageWithTag(mysql),
				ImagePullPolicy: core.PullIfNotPresent,
				Ports: []core.ContainerPort{
					{
						Name:          api.PrometheusExporterPortName,
						Protocol:      core.ProtocolTCP,
						ContainerPort: mysql.Spec.Monitor.Prometheus.Port,
					},
				},
			})
		}
		// Set Admin Secret as MYSQL_ROOT_PASSWORD env variable
		in = upsertEnv(in, mysql)
		in = upsertDataVolume(in, mysql)
		if mysql.Spec.Init != nil && mysql.Spec.Init.ScriptSource != nil {
			in = upsertInitScript(in, mysql.Spec.Init.ScriptSource.VolumeSource)
		}

		in.Spec.Template.Spec.NodeSelector = mysql.Spec.NodeSelector
		in.Spec.Template.Spec.Affinity = mysql.Spec.Affinity
		in.Spec.Template.Spec.SchedulerName = mysql.Spec.SchedulerName
		in.Spec.Template.Spec.Tolerations = mysql.Spec.Tolerations
		if c.opt.EnableRbac {
			in.Spec.Template.Spec.ServiceAccountName = mysql.Name
		}

		in.Spec.UpdateStrategy.Type = apps.RollingUpdateStatefulSetStrategyType

		return in
	})
}

func upsertDataVolume(statefulSet *apps.StatefulSet, mysql *api.MySQL) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceNameMySQL {
			volumeMount := core.VolumeMount{
				Name:      "data",
				MountPath: "/var/lib/mysql",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			pvcSpec := mysql.Spec.Storage
			if pvcSpec != nil {
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

func upsertEnv(statefulSet *apps.StatefulSet, mysql *api.MySQL) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceNameMySQL {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, core.EnvVar{
				Name: "MYSQL_ROOT_PASSWORD",
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: mysql.Spec.DatabaseSecret.SecretName,
						},
						Key: ".admin",
					},
				},
			})
			return statefulSet
		}
	}
	return statefulSet
}

func upsertInitScript(statefulSet *apps.StatefulSet, script core.VolumeSource) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceNameMySQL {
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
