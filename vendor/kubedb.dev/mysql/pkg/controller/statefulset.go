/*
Copyright AppsCode Inc. and Contributors

Licensed under the PolyForm Noncommercial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/PolyForm-Noncommercial-1.0.0.md

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
	"strconv"
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	"github.com/fatih/structs"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
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
			return kutil.VerbUnchanged, err
		}
		c.recorder.Eventf(
			mysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet",
			vt,
		)
	}

	// ensure pdb
	if err := c.CreateStatefulSetPodDisruptionBudget(statefulSet); err != nil {
		return kutil.VerbUnchanged, err
	}

	return vt, nil
}

func (c *Controller) checkStatefulSet(mysql *api.MySQL) error {
	// SatatefulSet for MySQL database
	statefulSet, err := c.Client.AppsV1().StatefulSets(mysql.Namespace).Get(context.TODO(), mysql.OffshootName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindMySQL ||
		statefulSet.Labels[api.LabelDatabaseName] != mysql.Name {
		return fmt.Errorf(`intended statefulSet "%v/%v" already exists`, mysql.Namespace, mysql.OffshootName())
	}

	return nil
}

func (c *Controller) createStatefulSet(mysql *api.MySQL) (*apps.StatefulSet, kutil.VerbType, error) {
	statefulSetMeta := metav1.ObjectMeta{
		Name:      mysql.OffshootName(),
		Namespace: mysql.Namespace,
	}

	owner := metav1.NewControllerRef(mysql, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

	mysqlVersion, err := c.ExtClient.CatalogV1alpha1().MySQLVersions().Get(context.TODO(), string(mysql.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	return app_util.CreateOrPatchStatefulSet(
		context.TODO(),
		c.Client,
		statefulSetMeta,
		func(in *apps.StatefulSet) *apps.StatefulSet {
			in.Labels = mysql.OffshootLabels()
			in.Annotations = mysql.Spec.PodTemplate.Controller.Annotations
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

			in.Spec.Replicas = mysql.Spec.Replicas
			in.Spec.ServiceName = mysql.GoverningServiceName()
			in.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: mysql.OffshootSelectors(),
			}
			in.Spec.Template.Labels = mysql.OffshootSelectors()
			in.Spec.Template.Annotations = mysql.Spec.PodTemplate.Annotations
			in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(
				in.Spec.Template.Spec.InitContainers,
				append(
					[]core.Container{
						{
							Name:            "remove-lost-found",
							Image:           mysqlVersion.Spec.InitContainer.Image,
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
							Resources: mysql.Spec.PodTemplate.Spec.Resources,
						},
					},
					mysql.Spec.PodTemplate.Spec.InitContainers...,
				),
			)

			container := core.Container{
				Name:            api.ResourceSingularMySQL,
				Image:           mysqlVersion.Spec.DB.Image,
				ImagePullPolicy: core.PullIfNotPresent,
				Args:            mysql.Spec.PodTemplate.Spec.Args,
				Resources:       mysql.Spec.PodTemplate.Spec.Resources,
				LivenessProbe:   mysql.Spec.PodTemplate.Spec.LivenessProbe,
				ReadinessProbe:  mysql.Spec.PodTemplate.Spec.ReadinessProbe,
				Lifecycle:       mysql.Spec.PodTemplate.Spec.Lifecycle,
				Ports: []core.ContainerPort{
					{
						Name:          "db",
						ContainerPort: api.MySQLNodePort,
						Protocol:      core.ProtocolTCP,
					},
				},
				VolumeMounts: []core.VolumeMount{
					{
						Name:      "tmp",
						MountPath: "/tmp",
					},
				},
			}
			if mysql.Spec.Topology != nil && mysql.Spec.Topology.Mode != nil &&
				*mysql.Spec.Topology.Mode == api.MySQLClusterModeGroup {
				container.Command = []string{
					"peer-finder",
				}
				userProvidedArgs := strings.Join(mysql.Spec.PodTemplate.Spec.Args, " ")
				container.Args = []string{
					fmt.Sprintf("-service=%s", mysql.GoverningServiceName()),
					fmt.Sprintf("-on-start=/on-start.sh %s", userProvidedArgs),
				}
				if container.LivenessProbe != nil && structs.IsZero(*container.LivenessProbe) {
					container.LivenessProbe = nil
				}
				if container.ReadinessProbe != nil && structs.IsZero(*container.ReadinessProbe) {
					container.ReadinessProbe = nil
				}
			}

			// TODO: probe for standalone needs to be set from mutator
			probe := core.Probe{
				Handler: core.Handler{
					Exec: &core.ExecAction{
						Command: []string{
							"bash",
							"-c",
							`
export MYSQL_PWD=${MYSQL_ROOT_PASSWORD}
mysql -h localhost -nsLNE -e "select 1;" 2>/dev/null | grep -v "*"
`,
						},
					},
				},
			}
			if mysql.Spec.Topology == nil {
				container.ReadinessProbe = &probe
				container.LivenessProbe = &probe
			}
			if container.ReadinessProbe != nil {
				container.ReadinessProbe.InitialDelaySeconds = 60
				container.ReadinessProbe.PeriodSeconds = 10
				container.ReadinessProbe.TimeoutSeconds = 50
				container.ReadinessProbe.SuccessThreshold = 1
				container.ReadinessProbe.FailureThreshold = 3
			}

			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, container)
			in.Spec.Template.Spec.Volumes = []core.Volume{
				{
					Name: "tmp",
					VolumeSource: core.VolumeSource{
						EmptyDir: &core.EmptyDirVolumeSource{},
					},
				},
			}

			if mysql.GetMonitoringVendor() == mona.VendorPrometheus {
				in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
					Name: "exporter",
					Command: []string{
						"/bin/sh",
					},
					Args: []string{
						"-c",
						// DATA_SOURCE_NAME=user:password@tcp(localhost:5555)/dbname
						// owner: https://github.com/prometheus/mysqld_exporter#setting-the-mysql-servers-data-source-name
						fmt.Sprintf(`export DATA_SOURCE_NAME="${MYSQL_ROOT_USERNAME:-}:${MYSQL_ROOT_PASSWORD:-}@(127.0.0.1:3306)/"
						/bin/mysqld_exporter --web.listen-address=:%v --web.telemetry-path=%v %v`,
							mysql.Spec.Monitor.Prometheus.Exporter.Port, mysql.StatsService().Path(), strings.Join(mysql.Spec.Monitor.Prometheus.Exporter.Args, " ")),
					},
					Image: mysqlVersion.Spec.Exporter.Image,
					Ports: []core.ContainerPort{
						{
							Name:          api.PrometheusExporterPortName,
							Protocol:      core.ProtocolTCP,
							ContainerPort: mysql.Spec.Monitor.Prometheus.Exporter.Port,
						},
					},
					Env:             mysql.Spec.Monitor.Prometheus.Exporter.Env,
					Resources:       mysql.Spec.Monitor.Prometheus.Exporter.Resources,
					SecurityContext: mysql.Spec.Monitor.Prometheus.Exporter.SecurityContext,
				})
			}
			// Set Admin Secret as MYSQL_ROOT_PASSWORD env variable
			in = upsertEnv(in, mysql)
			in = upsertDataVolume(in, mysql)
			in = upsertCustomConfig(in, mysql)

			if mysql.Spec.Init != nil && mysql.Spec.Init.ScriptSource != nil {
				in = upsertInitScript(in, mysql.Spec.Init.ScriptSource.VolumeSource)
			}

			in.Spec.Template.Spec.NodeSelector = mysql.Spec.PodTemplate.Spec.NodeSelector
			in.Spec.Template.Spec.Affinity = mysql.Spec.PodTemplate.Spec.Affinity
			if mysql.Spec.PodTemplate.Spec.SchedulerName != "" {
				in.Spec.Template.Spec.SchedulerName = mysql.Spec.PodTemplate.Spec.SchedulerName
			}
			in.Spec.Template.Spec.Tolerations = mysql.Spec.PodTemplate.Spec.Tolerations
			in.Spec.Template.Spec.ImagePullSecrets = mysql.Spec.PodTemplate.Spec.ImagePullSecrets
			in.Spec.Template.Spec.PriorityClassName = mysql.Spec.PodTemplate.Spec.PriorityClassName
			in.Spec.Template.Spec.Priority = mysql.Spec.PodTemplate.Spec.Priority
			if in.Spec.Template.Spec.SecurityContext == nil {
				in.Spec.Template.Spec.SecurityContext = mysql.Spec.PodTemplate.Spec.SecurityContext
			}
			in.Spec.Template.Spec.ServiceAccountName = mysql.Spec.PodTemplate.Spec.ServiceAccountName
			in.Spec.UpdateStrategy = mysql.Spec.UpdateStrategy
			in = upsertUserEnv(in, mysql)

			return in
		},
		metav1.PatchOptions{},
	)
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
			if mysql.Spec.StorageType == api.StorageTypeEphemeral {
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
					log.Infof(`Using "%v" as AccessModes in mysql.Spec.Storage`, core.ReadWriteOnce)
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

func upsertEnv(statefulSet *apps.StatefulSet, mysql *api.MySQL) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMySQL || container.Name == "exporter" {
			envs := []core.EnvVar{
				{
					Name: "MYSQL_ROOT_PASSWORD",
					ValueFrom: &core.EnvVarSource{
						SecretKeyRef: &core.SecretKeySelector{
							LocalObjectReference: core.LocalObjectReference{
								Name: mysql.Spec.DatabaseSecret.SecretName,
							},
							Key: KeyMySQLPassword,
						},
					},
				},
				{
					Name: "MYSQL_ROOT_USERNAME",
					ValueFrom: &core.EnvVarSource{
						SecretKeyRef: &core.SecretKeySelector{
							LocalObjectReference: core.LocalObjectReference{
								Name: mysql.Spec.DatabaseSecret.SecretName,
							},
							Key: KeyMySQLUser,
						},
					},
				},
			}
			if mysql.Spec.Topology != nil &&
				mysql.Spec.Topology.Mode != nil &&
				*mysql.Spec.Topology.Mode == api.MySQLClusterModeGroup &&
				container.Name == api.ResourceSingularMySQL {
				envs = append(envs, []core.EnvVar{
					{
						Name:  "BASE_NAME",
						Value: mysql.Name,
					},
					{
						Name:  "GOV_SVC",
						Value: mysql.GoverningServiceName(),
					},
					{
						Name: "POD_NAMESPACE",
						ValueFrom: &core.EnvVarSource{
							FieldRef: &core.ObjectFieldSelector{
								FieldPath: "metadata.namespace",
							},
						},
					},
					{
						Name:  "GROUP_NAME",
						Value: mysql.Spec.Topology.Group.Name,
					},
					{
						Name:  "BASE_SERVER_ID",
						Value: strconv.Itoa(int(*mysql.Spec.Topology.Group.BaseServerID)),
					},
				}...)
			}
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envs...)
		}
	}

	return statefulSet
}

// upsertUserEnv add/overwrite env from user provided env in crd spec
func upsertUserEnv(statefulSet *apps.StatefulSet, mysql *api.MySQL) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMySQL {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, mysql.Spec.PodTemplate.Spec.Env...)
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
	return core_util.WaitUntilPodRunningBySelector(
		context.TODO(),
		c.Client,
		statefulSet.Namespace,
		statefulSet.Spec.Selector,
		int(types.Int32(statefulSet.Spec.Replicas)),
	)
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
