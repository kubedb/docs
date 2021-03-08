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
	"strings"

	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/pkg/eventer"

	"gomodules.xyz/x/log"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func (c *Controller) ensureStatefulSet(db *api.MariaDB) (kutil.VerbType, error) {
	dbVersion, err := c.DBClient.CatalogV1alpha1().MariaDBVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	statefulSetMeta := metav1.ObjectMeta{
		Name:      db.OffshootName(),
		Namespace: db.Namespace,
	}
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMariaDB))

	stsNew, vt, err := app_util.CreateOrPatchStatefulSet(
		context.TODO(),
		c.Client,
		statefulSetMeta,
		func(in *apps.StatefulSet) *apps.StatefulSet {
			in.Labels = db.OffshootLabels()
			in.Annotations = db.Spec.PodTemplate.Controller.Annotations
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
			in.Spec.Replicas = db.Spec.Replicas
			in.Spec.ServiceName = db.GoverningServiceName()
			in.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: db.OffshootSelectors(),
			}
			in.Spec.Template.Labels = db.OffshootSelectors()
			in.Spec.Template.Annotations = db.Spec.PodTemplate.Annotations
			in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(
				in.Spec.Template.Spec.InitContainers,
				append(lostAndFoundCleaner(db, dbVersion), db.Spec.PodTemplate.Spec.InitContainers...),
			)
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
				in.Spec.Template.Spec.Containers,
				mariaDBContainer(db, dbVersion),
			)
			if db.Spec.Monitor != nil && db.Spec.Monitor.Agent.Vendor() == mona.VendorPrometheus {
				in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
					in.Spec.Template.Spec.Containers,
					exporterContainer(db, dbVersion),
				)
			}

			in.Spec.Template.Spec.Volumes = core_util.UpsertVolume(in.Spec.Template.Spec.Volumes, initScriptVolume(db)...)
			in = upsertEnv(in, db)
			in = upsertVolumes(in, db)

			if db.Spec.ConfigSecret != nil {
				in.Spec.Template = upsertCustomConfig(in.Spec.Template, db.Spec.ConfigSecret)
			}

			in.Spec.Template.Spec.NodeSelector = db.Spec.PodTemplate.Spec.NodeSelector
			in.Spec.Template.Spec.Affinity = db.Spec.PodTemplate.Spec.Affinity
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
			in.Spec.Template.Spec.SecurityContext = db.Spec.PodTemplate.Spec.SecurityContext
			in.Spec.Template.Spec.ServiceAccountName = db.OffshootName()
			in.Spec.UpdateStrategy = apps.StatefulSetUpdateStrategy{
				Type: apps.OnDeleteStatefulSetStrategyType,
			}
			in.Spec.Template.Spec.ReadinessGates = core_util.UpsertPodReadinessGateConditionType(in.Spec.Template.Spec.ReadinessGates, core_util.PodConditionTypeReady)

			return in
		},
		metav1.PatchOptions{},
	)

	if err != nil {
		return kutil.VerbUnchanged, err
	}

	if vt != kutil.VerbUnchanged {
		c.Recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet %v/%v",
			vt, db.Namespace, db.Name,
		)
		if err := c.CreateStatefulSetPodDisruptionBudget(stsNew); err != nil {
			return kutil.VerbUnchanged, err
		}
		log.Info("successfully created/patched PodDisruptonBudget")
	}

	return vt, nil
}

func upsertCustomConfig(
	template core.PodTemplateSpec, configSecret *core.LocalObjectReference) core.PodTemplateSpec {
	for i, container := range template.Spec.Containers {
		if container.Name == api.ResourceSingularMariaDB {
			configVolumeMount := core.VolumeMount{
				Name:      "custom-config",
				MountPath: "/etc/mysql/conf.d",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, configVolumeMount)
			template.Spec.Containers[i].VolumeMounts = volumeMounts

			configVolume := core.Volume{
				Name: "custom-config",
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: configSecret.Name,
					},
				},
			}

			volumes := template.Spec.Volumes
			volumes = core_util.UpsertVolume(volumes, configVolume)
			template.Spec.Volumes = volumes
			break
		}
	}
	return template
}

func lostAndFoundCleaner(db *api.MariaDB, dbVersion *v1alpha1.MariaDBVersion) []core.Container {
	return []core.Container{
		{
			Name:            "remove-lost-found",
			Image:           dbVersion.Spec.InitContainer.Image,
			ImagePullPolicy: core.PullIfNotPresent,
			Command: []string{
				"rm",
				"-rf",
				"/var/lib/mysql/lost+found",
			},
			VolumeMounts: []core.VolumeMount{
				{
					Name:      "data",
					MountPath: api.MariaDBDataMountPath,
				},
			},
			Resources: db.Spec.PodTemplate.Spec.Resources,
		},
	}

}

func mariaDBContainer(db *api.MariaDB, dbVersion *v1alpha1.MariaDBVersion) core.Container {
	return core.Container{
		Name:            api.ResourceSingularMariaDB,
		Image:           dbVersion.Spec.DB.Image,
		ImagePullPolicy: core.PullIfNotPresent,
		Command:         getCmdsForMariaDBContainer(db),
		Args:            getArgsForMariaDBContainer(db),
		Ports: []core.ContainerPort{
			{
				Name:          api.MySQLDatabasePortName,
				ContainerPort: api.MySQLDatabasePort,
				Protocol:      core.ProtocolTCP,
			},
		},
		Env:             core_util.UpsertEnvVars(db.Spec.PodTemplate.Spec.Env, getEnvsForMariaDBContainer(db)...),
		Resources:       db.Spec.PodTemplate.Spec.Resources,
		SecurityContext: db.Spec.PodTemplate.Spec.ContainerSecurityContext,
		LivenessProbe:   db.Spec.PodTemplate.Spec.LivenessProbe,
		ReadinessProbe:  db.Spec.PodTemplate.Spec.ReadinessProbe,
		Lifecycle:       db.Spec.PodTemplate.Spec.Lifecycle,
		VolumeMounts:    initScriptVolumeMount(db),
	}
}
func exporterContainer(db *api.MariaDB, dbVersion *v1alpha1.MariaDBVersion) core.Container {
	var commands []string
	// pass config.my-cnf flag into exporter to configure TLS
	if db.Spec.TLS != nil {
		// ref: https://github.com/prometheus/mysqld_exporter#general-flags
		// https://github.com/prometheus/mysqld_exporter#customizing-configuration-for-a-ssl-connection
		cmd := strings.Join(append([]string{
			"/bin/mysqld_exporter",
			fmt.Sprintf("--web.listen-address=:%d", db.Spec.Monitor.Prometheus.Exporter.Port),
			fmt.Sprintf("--web.telemetry-path=%v", db.StatsService().Path()),
			"--config.my-cnf=/etc/mysql/config/exporter/exporter.cnf",
		}, db.Spec.Monitor.Prometheus.Exporter.Args...), " ")
		commands = []string{cmd}
	} else {
		// DATA_SOURCE_NAME=user:password@tcp(localhost:5555)/dbname
		// ref: https://github.com/prometheus/mysqld_exporter#setting-the-mysql-servers-data-source-name
		cmd := strings.Join(append([]string{
			"/bin/mysqld_exporter",
			fmt.Sprintf("--web.listen-address=:%d", db.Spec.Monitor.Prometheus.Exporter.Port),
			fmt.Sprintf("--web.telemetry-path=%v", db.StatsService().Path()),
		}, db.Spec.Monitor.Prometheus.Exporter.Args...), " ")
		commands = []string{
			`export DATA_SOURCE_NAME="${MYSQL_ROOT_USERNAME:-}:${MYSQL_ROOT_PASSWORD:-}@(127.0.0.1:3306)/"`,
			cmd,
		}
	}
	script := strings.Join(commands, ";")

	return core.Container{
		Name: api.ContainerExporterName,
		Command: []string{
			"/bin/sh",
		},
		Args: []string{
			"-c",
			script,
		},
		Image: dbVersion.Spec.Exporter.Image,
		Ports: []core.ContainerPort{
			{
				Name:          mona.PrometheusExporterPortName,
				Protocol:      core.ProtocolTCP,
				ContainerPort: db.Spec.Monitor.Prometheus.Exporter.Port,
			},
		},
		Env:             db.Spec.Monitor.Prometheus.Exporter.Env,
		Resources:       db.Spec.Monitor.Prometheus.Exporter.Resources,
		SecurityContext: db.Spec.Monitor.Prometheus.Exporter.SecurityContext,
	}
}

func getTLSArgsForMariaDBContainer(db *api.MariaDB) []string {
	args := []string{
		"--ssl-capath=/etc/mysql/certs/server",
		"--ssl-ca=/etc/mysql/certs/server/ca.crt",
		"--ssl-cert=/etc/mysql/certs/server/tls.crt",
		"--ssl-key=/etc/mysql/certs/server/tls.key",
	}
	if db.IsCluster() {
		args = append(args, "--wsrep-provider-options='socket.ssl_key=/etc/mysql/certs/server/tls.key;socket.ssl_cert=/etc/mysql/certs/server/tls.crt;socket.ssl_ca=/etc/mysql/certs/server/ca.crt'")
	}
	if db.Spec.RequireSSL {
		args = append(args, "--require-secure-transport=ON")
	}
	return args
}

func getArgsForMariaDBContainer(db *api.MariaDB) []string {
	var args, tempArgs []string
	if db.IsCluster() {
		args = []string{
			fmt.Sprintf("-service=%s", db.GoverningServiceName()),
			"-on-start",
		}
		tempArgs = append(tempArgs, "/on-start.sh")
	}
	// adding user provided arguments
	tempArgs = append(tempArgs, db.Spec.PodTemplate.Spec.Args...)

	// Adding arguments for TLS setup
	if db.Spec.TLS != nil {
		tempArgs = append(tempArgs, getTLSArgsForMariaDBContainer(db)...)
	}
	if tempArgs != nil {
		if db.IsCluster() {
			args = append(args, strings.Join(tempArgs, " "))
		} else {
			args = append(args, tempArgs...)
		}
	}
	return args
}

func getCmdsForMariaDBContainer(db *api.MariaDB) []string {
	var cmds []string
	if db.IsCluster() {
		cmds = []string{
			"peer-finder",
		}
	}
	return cmds
}

func getEnvsForMariaDBContainer(db *api.MariaDB) []core.EnvVar {
	var envList []core.EnvVar
	if db.IsCluster() {
		envList = append(envList, core.EnvVar{
			Name:  "CLUSTER_NAME",
			Value: db.OffshootName(),
		})
	}
	return envList
}

func initScriptVolume(db *api.MariaDB) []core.Volume {
	var volumes []core.Volume
	if !db.IsCluster() && db.Spec.Init != nil && db.Spec.Init.Script != nil {
		volumes = append(volumes, core.Volume{
			Name:         "initial-script",
			VolumeSource: db.Spec.Init.Script.VolumeSource,
		})
	}
	return volumes
}

func initScriptVolumeMount(db *api.MariaDB) []core.VolumeMount {
	var volumeMounts []core.VolumeMount
	if !db.IsCluster() && db.Spec.Init != nil && db.Spec.Init.Script != nil {
		volumeMounts = append(volumeMounts, core.VolumeMount{
			Name:      "initial-script",
			MountPath: api.MariaDBInitDBMountPath,
		})
	}
	return volumeMounts
}

func upsertVolumes(statefulSet *apps.StatefulSet, db *api.MariaDB) *apps.StatefulSet {

	// Add DataVolume
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMariaDB {
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = core_util.UpsertVolumeMount(statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts, core.VolumeMount{
				Name:      "data",
				MountPath: api.MariaDBDataMountPath,
			})
			pvcSpec := db.Spec.Storage
			if db.Spec.StorageType == api.StorageTypeEphemeral {
				ed := core.EmptyDirVolumeSource{}
				if pvcSpec != nil {
					if sz, found := pvcSpec.Resources.Requests[core.ResourceStorage]; found {
						ed.SizeLimit = &sz
					}
				}
				statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(statefulSet.Spec.Template.Spec.Volumes, core.Volume{
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
					log.Infof(`Using "%v" as AccessModes in .spec.storage`, core.ReadWriteOnce)
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
	// upsert TLSConfig volumes
	if db.Spec.TLS != nil {
		statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(
			statefulSet.Spec.Template.Spec.Volumes,
			[]core.Volume{
				{
					Name: "tls-server-config",
					VolumeSource: core.VolumeSource{
						Secret: &core.SecretVolumeSource{
							SecretName: db.MustCertSecretName(api.MariaDBServerCert),
							Items: []core.KeyToPath{
								{
									Key:  "ca.crt",
									Path: "ca.crt",
								},
								{
									Key:  "tls.crt",
									Path: "tls.crt",
								},
								{
									Key:  "tls.key",
									Path: "tls.key",
								},
							},
						},
					},
				},
				{
					Name: "tls-client-config",
					VolumeSource: core.VolumeSource{
						Secret: &core.SecretVolumeSource{
							SecretName: db.MustCertSecretName(api.MariaDBArchiverCert),
							Items: []core.KeyToPath{
								{
									Key:  "ca.crt",
									Path: "ca.crt",
								},
								{
									Key:  "tls.crt",
									Path: "tls.crt",
								},
								{
									Key:  "tls.key",
									Path: "tls.key",
								},
							},
						},
					},
				},
				{
					Name: "tls-exporter-config",
					VolumeSource: core.VolumeSource{
						Secret: &core.SecretVolumeSource{
							SecretName: db.MustCertSecretName(api.MariaDBMetricsExporterCert),
							Items: []core.KeyToPath{
								{
									Key:  "ca.crt",
									Path: "ca.crt",
								},
								{
									Key:  "tls.crt",
									Path: "tls.crt",
								},
								{
									Key:  "tls.key",
									Path: "tls.key",
								},
							},
						},
					},
				},
				{
					Name: "tls-metrics-exporter-config",
					VolumeSource: core.VolumeSource{
						Secret: &core.SecretVolumeSource{
							SecretName: meta_util.NameWithSuffix(db.Name, api.MySQLMetricsExporterConfigSecretSuffix),
							Items: []core.KeyToPath{
								{
									Key:  "exporter.cnf",
									Path: "exporter.cnf",
								},
							},
						},
					},
				},
			}...)

		for i, container := range statefulSet.Spec.Template.Spec.Containers {
			if container.Name == api.ResourceSingularMariaDB {
				statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = core_util.UpsertVolumeMount(statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts,
					[]core.VolumeMount{
						{
							Name:      "tls-server-config",
							MountPath: "/etc/mysql/certs/server",
						},
						{
							Name:      "tls-client-config",
							MountPath: "/etc/mysql/certs/client",
						},
					}...)
			}
			if container.Name == api.ContainerExporterName {
				statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = core_util.UpsertVolumeMount(statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts,
					[]core.VolumeMount{
						{
							Name:      "tls-exporter-config",
							MountPath: "/etc/mysql/certs/exporter",
						},
						{
							Name:      "tls-metrics-exporter-config",
							MountPath: "/etc/mysql/config/exporter",
						},
					}...)
			}
		}
	}

	return statefulSet
}

// upsertUserEnv add/overwrite env from user provided env in crd spec
func upsertEnv(statefulSet *apps.StatefulSet, db *api.MariaDB) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMariaDB || container.Name == "exporter" {
			envs := []core.EnvVar{
				{
					Name: "MYSQL_ROOT_PASSWORD",
					ValueFrom: &core.EnvVarSource{
						SecretKeyRef: &core.SecretKeySelector{
							LocalObjectReference: core.LocalObjectReference{
								Name: db.Spec.AuthSecret.Name,
							},
							Key: core.BasicAuthPasswordKey,
						},
					},
				},
				{
					Name: "MYSQL_ROOT_USERNAME",
					ValueFrom: &core.EnvVarSource{
						SecretKeyRef: &core.SecretKeySelector{
							LocalObjectReference: core.LocalObjectReference{
								Name: db.Spec.AuthSecret.Name,
							},
							Key: core.BasicAuthUsernameKey,
						},
					},
				},
			}

			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envs...)
		}
	}

	return statefulSet
}

func requiredSecretList(db *api.MariaDB) []string {
	var secretList []string
	for _, cert := range db.Spec.TLS.Certificates {
		secretList = append(secretList, cert.SecretName)
	}

	if db.Spec.Monitor != nil {
		secretList = append(secretList, meta_util.NameWithSuffix(db.Name, api.MySQLMetricsExporterConfigSecretSuffix))
	}
	return secretList
}
