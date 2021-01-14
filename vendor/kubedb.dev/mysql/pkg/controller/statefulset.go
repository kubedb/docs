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
	"strconv"
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/pkg/eventer"

	"gomodules.xyz/x/log"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func (c *Controller) ensureStatefulSet(db *api.MySQL) error {
	stsName, _, err := c.findStatefulSet(db)
	if err != nil {
		return err
	}

	// Create statefulSet for MySQL database
	stsNew, vt, err := c.createOrPatchStatefulSet(db, stsName)
	if err != nil {
		return err
	}
	// Check StatefulSet Pod status
	if vt != kutil.VerbUnchanged {
		c.Recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet",
			vt,
		)
		// ensure pdb
		if err := c.CreateStatefulSetPodDisruptionBudget(stsNew); err != nil {
			return err
		}
		log.Info("Successfully created/patched PodDisruptionBudget")
	}

	return nil
}

func (c *Controller) createOrPatchStatefulSet(db *api.MySQL, stsName string) (*apps.StatefulSet, kutil.VerbType, error) {
	statefulSetMeta := metav1.ObjectMeta{
		Name:      stsName,
		Namespace: db.Namespace,
	}
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

	mysqlVersion, err := c.DBClient.CatalogV1alpha1().MySQLVersions().Get(context.TODO(), db.Spec.Version, metav1.GetOptions{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	return app_util.CreateOrPatchStatefulSet(
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
							Resources: db.Spec.PodTemplate.Spec.Resources,
						},
					},
					db.Spec.PodTemplate.Spec.InitContainers...,
				),
			)

			container := core.Container{
				Name:            api.ResourceSingularMySQL,
				Image:           mysqlVersion.Spec.DB.Image,
				ImagePullPolicy: core.PullIfNotPresent,
				Args:            db.Spec.PodTemplate.Spec.Args,
				Resources:       db.Spec.PodTemplate.Spec.Resources,
				LivenessProbe:   db.Spec.PodTemplate.Spec.LivenessProbe,
				ReadinessProbe:  db.Spec.PodTemplate.Spec.ReadinessProbe,
				Lifecycle:       db.Spec.PodTemplate.Spec.Lifecycle,
				Ports: []core.ContainerPort{
					{
						Name:          api.MySQLDatabasePortName,
						ContainerPort: api.MySQLDatabasePort,
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

			// add ssl certs flag into args to configure TLS for standalone
			if db.Spec.Topology == nil && db.Spec.TLS != nil {
				args := container.Args
				tlsArgs := []string{
					"--ssl-capath=/etc/mysql/certs",
					"--ssl-ca=/etc/mysql/certs/ca.crt",
					"--ssl-cert=/etc/mysql/certs/server.crt",
					"--ssl-key=/etc/mysql/certs/server.key",
				}
				args = append(args, tlsArgs...)
				if db.Spec.RequireSSL {
					args = append(args, "--require-secure-transport=ON")
				}
				container.Args = args
			}

			if db.UsesGroupReplication() {
				// replicationModeDetector is used to continuous select primary pod
				// and add label as primary
				replicationModeDetector := core.Container{
					Name:            api.ReplicationModeDetectorContainerName,
					Image:           mysqlVersion.Spec.ReplicationModeDetector.Image,
					ImagePullPolicy: core.PullIfNotPresent,
					Args: append([]string{
						"run",
						fmt.Sprintf("--db-name=%s", db.Name),
						fmt.Sprintf("--db-kind=%s", api.ResourceKindMySQL),
					}, c.LoggerOptions.ToFlags()...),
				}

				in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, replicationModeDetector)

				container.Command = []string{
					"peer-finder",
				}

				args := db.Spec.PodTemplate.Spec.Args

				// add ssl certs flag into args in peer-finder to configure TLS for group replication
				if db.Spec.TLS != nil {
					tlsArgs := []string{
						"--ssl-capath=/etc/mysql/certs",
						"--ssl-ca=/etc/mysql/certs/ca.crt",
						"--ssl-cert=/etc/mysql/certs/server.crt",
						"--ssl-key=/etc/mysql/certs/server.key",
					}
					args = append(args, tlsArgs...)
					if db.Spec.RequireSSL {
						args = append(args, "--require-secure-transport=ON ")
					}
				}

				container.Args = []string{
					fmt.Sprintf("-service=%s", db.GoverningServiceName()),
					"-on-start",
					strings.Join(append([]string{"/on-start.sh"}, args...), " "),
				}
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

			if db.Spec.Monitor != nil && db.Spec.Monitor.Agent.Vendor() == mona.VendorPrometheus {
				var commands []string
				// pass config.my-cnf flag into exporter to configure TLS
				if db.Spec.TLS != nil {
					// ref: https://github.com/prometheus/mysqld_exporter#general-flags
					// https://github.com/prometheus/mysqld_exporter#customizing-configuration-for-a-ssl-connection
					cmd := strings.Join(append([]string{
						"/bin/mysqld_exporter",
						fmt.Sprintf("--web.listen-address=:%d", db.Spec.Monitor.Prometheus.Exporter.Port),
						fmt.Sprintf("--web.telemetry-path=%v", db.StatsService().Path()),
						"--config.my-cnf=/etc/mysql/certs/exporter.cnf",
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
				in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
					Name: api.ContainerExporterName,
					Command: []string{
						"/bin/sh",
					},
					Args: []string{
						"-c",
						script,
					},
					Image: mysqlVersion.Spec.Exporter.Image,
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
				})
			}
			// Set Admin Secret as MYSQL_ROOT_PASSWORD env variable
			in = upsertEnv(in, db, stsName)
			in = upsertDataVolume(in, db)
			in = upsertCustomConfig(in, db)

			if db.Spec.Init != nil && db.Spec.Init.Script != nil {
				in = upsertInitScript(in, db.Spec.Init.Script.VolumeSource)
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
			if in.Spec.Template.Spec.SecurityContext == nil {
				in.Spec.Template.Spec.SecurityContext = db.Spec.PodTemplate.Spec.SecurityContext
			}
			in.Spec.Template.Spec.ServiceAccountName = db.Spec.PodTemplate.Spec.ServiceAccountName
			in.Spec.UpdateStrategy = apps.StatefulSetUpdateStrategy{
				Type: apps.OnDeleteStatefulSetStrategyType,
			}
			in = upsertUserEnv(in, db)

			// configure tls if configured in DB
			in = upsertTLSVolume(in, db)

			in.Spec.Template.Spec.ReadinessGates = core_util.UpsertPodReadinessGateConditionType(in.Spec.Template.Spec.ReadinessGates, core_util.PodConditionTypeReady)

			return in
		}, metav1.PatchOptions{})
}

func upsertDataVolume(statefulSet *apps.StatefulSet, db *api.MySQL) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMySQL {
			volumeMount := core.VolumeMount{
				Name:      "data",
				MountPath: "/var/lib/mysql",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

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

func upsertEnv(statefulSet *apps.StatefulSet, db *api.MySQL, stsName string) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMySQL || container.Name == api.ContainerExporterName || container.Name == api.ReplicationModeDetectorContainerName {
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
			if db.UsesGroupReplication() &&
				container.Name == api.ResourceSingularMySQL {
				envs = append(envs, []core.EnvVar{
					{
						Name:  "BASE_NAME",
						Value: stsName,
					},
					{
						Name:  "GOV_SVC",
						Value: db.GoverningServiceName(),
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
						Value: db.Spec.Topology.Group.Name,
					},
					{
						Name:  "BASE_SERVER_ID",
						Value: strconv.Itoa(int(*db.Spec.Topology.Group.BaseServerID)),
					},
				}...)
			}
			if container.Name == api.ReplicationModeDetectorContainerName {
				envs = append(envs, []core.EnvVar{
					{
						Name: "POD_NAME",
						ValueFrom: &core.EnvVarSource{
							FieldRef: &core.ObjectFieldSelector{
								FieldPath: "metadata.name",
							},
						},
					},
				}...)
			}
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envs...)
		}
	}

	return statefulSet
}

// upsertUserEnv add/overwrite env from user provided env in crd spec
func upsertUserEnv(statefulSet *apps.StatefulSet, db *api.MySQL) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMySQL {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, db.Spec.PodTemplate.Spec.Env...)
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

func upsertCustomConfig(statefulSet *apps.StatefulSet, db *api.MySQL) *apps.StatefulSet {
	if db.Spec.ConfigSecret != nil {
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
					Name: "custom-config",
					VolumeSource: core.VolumeSource{
						Secret: &core.SecretVolumeSource{
							SecretName: db.Spec.ConfigSecret.Name,
						},
					},
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

func (c *Controller) findStatefulSet(db *api.MySQL) (string, *apps.StatefulSet, error) {
	stsList, err := c.Client.AppsV1().StatefulSets(db.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(db.OffshootSelectors()).String(),
	})
	if err != nil {
		return "", nil, err
	}

	count := 0
	var cur *apps.StatefulSet
	for i, sts := range stsList.Items {
		if metav1.IsControlledBy(&sts, db) {
			count++
			cur = &stsList.Items[i]
		}
	}

	switch count {
	case 0:
		return db.OffshootName(), nil, nil
	case 1:
		return cur.Name, cur, nil
	}
	return "", nil, fmt.Errorf("more then one StatefulSet found for MySQL %s/%s", db.Namespace, db.Name)
}

func upsertTLSVolume(sts *apps.StatefulSet, db *api.MySQL) *apps.StatefulSet {
	if db.Spec.TLS != nil {
		volume := core.Volume{
			Name: "tls-volume",
			VolumeSource: core.VolumeSource{
				Projected: &core.ProjectedVolumeSource{
					Sources: []core.VolumeProjection{
						{
							Secret: &core.SecretProjection{
								LocalObjectReference: core.LocalObjectReference{
									Name: db.MustCertSecretName(api.MySQLServerCert),
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
									Name: db.MustCertSecretName(api.MySQLClientCert),
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
									Name: db.MustCertSecretName(api.MySQLMetricsExporterCert),
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
						{
							Secret: &core.SecretProjection{
								LocalObjectReference: core.LocalObjectReference{
									Name: meta_util.NameWithSuffix(db.Name, api.MySQLMetricsExporterConfigSecretSuffix),
								},
								Items: []core.KeyToPath{
									{
										Key:  "exporter.cnf",
										Path: "exporter.cnf",
									},
								},
							},
						},
					},
				},
			},
		}
		for i, container := range sts.Spec.Template.Spec.Containers {
			if container.Name == api.ResourceSingularMySQL {
				volumeMount := core.VolumeMount{
					Name:      "tls-volume",
					MountPath: "/etc/mysql/certs",
				}
				volumeMounts := container.VolumeMounts
				volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
				sts.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts
			}
			if container.Name == api.ContainerExporterName {
				volumeMount := core.VolumeMount{
					Name:      "exporter-tls-volume",
					MountPath: "/etc/mysql/certs",
				}
				volumeMounts := container.VolumeMounts
				volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
				sts.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts
			}
		}
		sts.Spec.Template.Spec.Volumes = core_util.UpsertVolume(
			sts.Spec.Template.Spec.Volumes,
			volume,
			exporterTLSVolume,
		)

	} else {
		for i, container := range sts.Spec.Template.Spec.Containers {
			if container.Name == api.ResourceSingularMySQL {
				sts.Spec.Template.Spec.Containers[i].VolumeMounts = core_util.EnsureVolumeMountDeleted(sts.Spec.Template.Spec.Containers[i].VolumeMounts, "tls-volume")
			}
			if container.Name == api.ContainerExporterName {
				sts.Spec.Template.Spec.Containers[i].VolumeMounts = core_util.EnsureVolumeMountDeleted(sts.Spec.Template.Spec.Containers[i].VolumeMounts, "exporter-tls-volume")
			}
		}
		sts.Spec.Template.Spec.Volumes = core_util.EnsureVolumeDeleted(sts.Spec.Template.Spec.Volumes, "tls-volume")
		sts.Spec.Template.Spec.Volumes = core_util.EnsureVolumeDeleted(sts.Spec.Template.Spec.Volumes, "exporter-tls-volume")
	}

	return sts
}
