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
	"sort"
	"strings"

	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

const (
	caFile   = "/etc/mysql/certs/ca.crt"
	certFile = "/etc/mysql/certs/server.crt"
	keyFile  = "/etc/mysql/certs/server.key"
)

func (c *Reconciler) ensureStatefulSet(db *api.MySQL) (kutil.VerbType, error) {
	// Create statefulSet for MySQL database
	stsNew, vt, err := c.createOrPatchStatefulSet(db)
	if err != nil {
		return kutil.VerbUnchanged, err
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
			return kutil.VerbUnchanged, err
		}
		klog.Info("Successfully created/patched PodDisruptionBudget")
	}

	return vt, nil
}

func (c *Reconciler) createOrPatchStatefulSet(db *api.MySQL) (*apps.StatefulSet, kutil.VerbType, error) {
	statefulSetMeta := metav1.ObjectMeta{
		Name:      db.OffshootName(),
		Namespace: db.Namespace,
	}
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))

	mysqlVersion, err := c.DBClient.CatalogV1alpha1().MySQLVersions().Get(context.TODO(), db.Spec.Version, metav1.GetOptions{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	var source *api.MySQL
	if db.IsReadReplica() {
		ns := db.Spec.Topology.ReadReplica.SourceRef.Namespace
		name := db.Spec.Topology.ReadReplica.SourceRef.Name
		source, err = c.DBClient.KubedbV1alpha2().MySQLs(ns).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return nil, kutil.VerbUnchanged, errors.Wrap(err, "unable to get source db object")
		}
	}

	return app_util.CreateOrPatchStatefulSet(
		context.TODO(),
		c.Client,
		statefulSetMeta,
		func(in *apps.StatefulSet) *apps.StatefulSet {
			in.Labels = db.PodControllerLabels()
			in.Annotations = db.Spec.PodTemplate.Controller.Annotations
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
			in.Spec.Replicas = db.Spec.Replicas
			in.Spec.ServiceName = db.GoverningServiceName()
			in.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: db.OffshootSelectors(),
			}

			in.Spec.Template.Labels = db.PodLabels()
			in.Spec.Template.Annotations = db.Spec.PodTemplate.Annotations
			in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(
				in.Spec.Template.Spec.InitContainers,
				append(getInitContainers(in, mysqlVersion), db.Spec.PodTemplate.Spec.InitContainers...),
			)

			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
				nil, getMySQLContainer(db, mysqlVersion))

			if db.UsesGroupReplication() || db.IsInnoDBCluster() {
				in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
					in.Spec.Template.Spec.Containers, getMySQLCoordinatorContainer(db, mysqlVersion))
			}

			if db.Spec.Monitor != nil && db.Spec.Monitor.Agent.Vendor() == mona.VendorPrometheus {
				in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
					in.Spec.Template.Spec.Containers, getMySQLExporterContainer(db, mysqlVersion))
			}

			in.Spec.Template.Spec.Volumes = []core.Volume{
				{
					Name: "tmp",
					VolumeSource: core.VolumeSource{
						EmptyDir: &core.EmptyDirVolumeSource{},
					},
				},
			}

			if db.IsReadReplica() && c.SourceHasSSL(db) {
				///test for secret exists
				in.Spec.Template.Spec.Volumes = append(in.Spec.Template.Spec.Volumes, []core.Volume{
					{
						Name: "source-ca",
						VolumeSource: core.VolumeSource{
							Secret: &core.SecretVolumeSource{
								SecretName: meta_util.NameWithSuffix(db.Name, "source-tls-secret"),
							},
						},
					},
				}...)
			}

			// Set Admin Secret as MYSQL_ROOT_PASSWORD env variable
			// TODO:
			//		- while creating the container: make sure it has the envs, dataMount, security, etc.
			//		- containers, err := GetMySqlContainers()
			//		- in.spec...containers = upsert(in.spec...containers, containers)

			in = c.updateStatefulSetEnv(in, db, source)
			in = c.upsertSharedScriptsVolume(in, db)
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
			in.Spec.Template.Spec.HostNetwork = db.Spec.PodTemplate.Spec.HostNetwork
			if in.Spec.Template.Spec.DNSPolicy == "" {
				in.Spec.Template.Spec.DNSPolicy = db.Spec.PodTemplate.Spec.DNSPolicy
			}
			in.Spec.Template.Spec.HostPID = db.Spec.PodTemplate.Spec.HostPID
			in.Spec.Template.Spec.HostIPC = db.Spec.PodTemplate.Spec.HostIPC
			if in.Spec.Template.Spec.SecurityContext == nil {
				in.Spec.Template.Spec.SecurityContext = db.Spec.PodTemplate.Spec.SecurityContext
			}
			in.Spec.Template.Spec.ServiceAccountName = db.Spec.PodTemplate.Spec.ServiceAccountName
			in.Spec.UpdateStrategy = apps.StatefulSetUpdateStrategy{
				Type: apps.OnDeleteStatefulSetStrategyType,
			}

			// if we use `IP` as podIdentity, we have to set hostNetwork to `True` and
			// dnsPolicy to `ClusterFirstWithHostNet` for using host IP
			// https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy
			if db.Spec.UseAddressType.IsIP() {
				in.Spec.Template.Spec.HostNetwork = true
				in.Spec.Template.Spec.DNSPolicy = core.DNSClusterFirstWithHostNet
			}

			in = upsertUserEnv(in, db)
			// configure tls if configured in DB
			in = c.upsertTLSVolume(in, db)
			// in.Spec.PodManagementPolicy = apps.ParallelPodManagement
			in.Spec.Template.Spec.ReadinessGates = nil
			return in
		}, metav1.PatchOptions{})
}

func getMySQLExporterContainer(db *api.MySQL, mysqlVersion *v1alpha1.MySQLVersion) core.Container {
	// ref: https://github.com/prometheus/mysqld_exporter#general-flags
	args := []string{
		"/bin/mysqld_exporter",
		fmt.Sprintf("--web.listen-address=:%d", db.Spec.Monitor.Prometheus.Exporter.Port),
		fmt.Sprintf("--web.telemetry-path=%v", db.StatsService().Path()),
	}
	if db.Spec.TLS != nil {
		// pass config.my-cnf flag into exporter to configure TLS
		// https://github.com/prometheus/mysqld_exporter#customizing-configuration-for-a-ssl-connection
		args = append(args, "--config.my-cnf=/etc/mysql/certs/exporter.cnf")
	}
	if db.UsesGroupReplication() || db.IsInnoDBCluster() {
		groupReplicationArgs := []string{
			"--collect.perf_schema.replication_group_members",
			"--collect.perf_schema.replication_group_member_stats",
			"--collect.perf_schema.replication_applier_status_by_worker",
			"--collect.group_replication.custom_query",
		}
		args = append(args, groupReplicationArgs...)
	}

	args = append(args, db.Spec.Monitor.Prometheus.Exporter.Args...)

	joinedArgs := strings.Join(args, " ")

	var exporterArgs []string
	if db.Spec.TLS == nil {
		// DATA_SOURCE_NAME=user:password@tcp(localhost:5555)/dbname
		// ref: https://github.com/prometheus/mysqld_exporter#setting-the-mysql-servers-data-source-name
		exporterArgs = append(exporterArgs, `export DATA_SOURCE_NAME="${MYSQL_ROOT_USERNAME:-}:${MYSQL_ROOT_PASSWORD:-}@(127.0.0.1:3306)/"`)
	}
	exporterArgs = append(exporterArgs, joinedArgs)

	script := strings.Join(exporterArgs, ";")
	return core.Container{
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
	}
}

func getMySQLCoordinatorContainer(db *api.MySQL, mysqlVersion *v1alpha1.MySQLVersion) core.Container {
	return core.Container{
		Name:            "mysql-coordinator",
		Image:           mysqlVersion.Spec.Coordinator.Image,
		ImagePullPolicy: core.PullIfNotPresent,
		Env:             core_util.UpsertEnvVars(db.Spec.PodTemplate.Spec.Env, getEnvsForMySQLCoordinatorContainer(db)...),
		Args:            []string{"run"},
		Resources:       db.Spec.Coordinator.Resources,
		SecurityContext: db.Spec.Coordinator.SecurityContext,
	}
}

func getEnvsForMySQLCoordinatorContainer(db *api.MySQL) []core.EnvVar {
	var envList []core.EnvVar
	envList = append(envList, core.EnvVar{
		Name:  "DB_NAME",
		Value: db.OffshootName(),
	})
	envList = append(envList, core.EnvVar{
		Name:  "NAMESPACE",
		Value: db.Namespace,
	})
	envList = append(envList, core.EnvVar{
		Name:  "GOVERNING_SERVICE_NAME",
		Value: db.GoverningServiceName(),
	})
	if db.UsesGroupReplication() {
		envList = append(envList, core.EnvVar{
			Name:  "TOPOLOGY",
			Value: string(api.MySQLModeGroupReplication),
		})
	} else if db.IsInnoDBCluster() {
		envList = append(envList, core.EnvVar{
			Name:  "TOPOLOGY",
			Value: string(api.MySQLModeInnoDBCluster),
		})
	}

	if db.Spec.TLS != nil {
		envList = append(envList, core.EnvVar{
			Name:  "SSL_ENABLED",
			Value: "ENABLED",
		})
	}
	return envList
}

func getMySQLContainer(db *api.MySQL, mysqlVersion *v1alpha1.MySQLVersion) core.Container {
	container := core.Container{
		Name:            api.ResourceSingularMySQL,
		Image:           mysqlVersion.Spec.DB.Image,
		ImagePullPolicy: core.PullIfNotPresent,
		Command:         getCmdsForMySQLContainer(db),
		Args:            getArgsForMysqlContainer(db, mysqlVersion),
		Resources:       db.Spec.PodTemplate.Spec.Resources,
		SecurityContext: db.Spec.PodTemplate.Spec.ContainerSecurityContext,
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

	return container
}

func getCmdsForMySQLContainer(db *api.MySQL) []string {
	var cmds []string
	if db.UsesGroupReplication() || db.IsInnoDBCluster() || db.IsReadReplica() {
		cmds = []string{
			"/scripts/tini",
			"-g",
			"--",
		}
	}
	return cmds
}

func getArgsForMysqlContainer(db *api.MySQL, mysqlVersion *v1alpha1.MySQLVersion) []string {
	var args []string
	if db.IsReadReplica() {
		args = append([]string{"scripts/run_read_only.sh"}, args...)
	}
	// add ssl certs flag into args to configure TLS for standalone
	if db.Spec.Topology == nil || db.IsReadReplica() {
		// args = append(db.Spec.PodTemplate.Spec.Args, args...)
		args = append(args, db.Spec.PodTemplate.Spec.Args...)
		if db.Spec.TLS != nil {
			tlsArgs := []string{
				"--ssl-capath=/etc/mysql/certs",
				"--ssl-ca=" + caFile,
				"--ssl-cert=" + certFile,
				"--ssl-key=" + keyFile,
			}
			args = append(args, tlsArgs...)
			if db.Spec.RequireSSL {
				args = append(args, "--require-secure-transport=ON")
			}
		}
	}

	if db.UsesGroupReplication() || db.IsInnoDBCluster() {

		userArgs := meta_util.ParseArgumentListToMap(db.Spec.PodTemplate.Spec.Args)

		specArgs := map[string]string{}
		// add ssl certs flag into args in peer-finder to configure TLS for group replication
		if db.Spec.TLS != nil {
			// https://dev.mysql.com/doc/refman/8.0/en/group-replication-secure-socket-layer-support-ssl.html
			// Host name identity verification with VERIFY_IDENTITY does not work with self-signed certificate
			// specArgs["loose-group_replication_ssl_mode"] = "VERIFY_IDENTITY"
			specArgs["loose-group_replication_ssl_mode"] = "VERIFY_CA"
			// the configuration for Group Replication's group communication connections is taken from the server's SSL configuration
			// https://dev.mysql.com/doc/refman/8.0/en/group-replication-secure-socket-layer-support-ssl.html
			specArgs["ssl-capath"] = "/etc/mysql/certs"
			specArgs["ssl-ca"] = caFile
			specArgs["ssl-cert"] = certFile
			specArgs["ssl-key"] = keyFile
			// By default, distributed recovery connections do not use SSL, even if we activated SSL for group communication connections,
			// and the server SSL options are not applied for distributed recovery connections. we must configure these connections separately
			// https://dev.mysql.com/doc/refman/8.0/en/group-replication-configuring-ssl-for-recovery.html
			specArgs["loose-group_replication_recovery_ssl_ca"] = caFile
			specArgs["loose-group_replication_recovery_ssl_cert"] = certFile
			specArgs["loose-group_replication_recovery_ssl_key"] = keyFile

			refVersion := semver.MustParse("8.0.17")

			curVersion := semver.MustParse(mysqlVersion.Spec.Version)
			if curVersion.Compare(refVersion) != -1 {
				// https://dev.mysql.com/doc/refman/8.0/en/clone-plugin-remote.html
				specArgs["loose-clone_ssl_ca"] = caFile
				specArgs["loose-clone_ssl_cert"] = certFile
				specArgs["loose-clone_ssl_key"] = keyFile
			}

			if db.Spec.RequireSSL {
				specArgs["require-secure-transport"] = "ON"
			}
		}
		// Argument priority (lowest to highest): recommendedArgs, userArgs, specArgs
		args := meta_util.BuildArgumentListFromMap(meta_util.OverwriteKeys(recommendedArgs(db, mysqlVersion), userArgs), specArgs)
		sort.Strings(args)

		// in peer-finder, we have to form peers either using pod IP or DNS. if podIdentity is set to `IP` then we have to use pod IP from pod status
		// otherwise, we have to use pod `DNS` using govern service.
		// That's why we have to pass either `selector` to select IP's of the pod or `service` to find the DNS of the pod.
		//peerFinderArgs := []string{
		//	fmt.Sprintf("-address-type=%s", db.Spec.UseAddressType),
		//}
		//if db.Spec.UseAddressType.IsIP() {
		//	peerFinderArgs = append(peerFinderArgs, fmt.Sprintf("-selector=%s", labels.Set(db.OffshootSelectors()).String()))
		//} else {
		//	peerFinderArgs = append(peerFinderArgs, fmt.Sprintf("-service=%s", db.GoverningServiceName()))
		//}

		if db.IsInnoDBCluster() {
			args = append([]string{"scripts/run_innodb.sh"}, args...)
		} else {
			args = append([]string{"scripts/run.sh"}, args...)
		}
		return args
	}

	if db.Spec.Topology == nil && db.Spec.AllowedReadReplicas != nil {
		// args = append(args, "--datadir=/var/lib/mysql/data")
		args = append(args, "--gtid-mode=ON", "--enforce_gtid_consistency=ON")
	}

	return args
}

func getInitContainers(statefulSet *apps.StatefulSet, mysqlVersion *v1alpha1.MySQLVersion) []core.Container {
	statefulSet.Spec.Template.Spec.InitContainers = core_util.UpsertContainer(
		statefulSet.Spec.Template.Spec.InitContainers,
		core.Container{
			Name:  "mysql-init",
			Image: mysqlVersion.Spec.InitContainer.Image,
			VolumeMounts: []core.VolumeMount{
				{
					Name:      "data",
					MountPath: "/var/lib/mysql",
				},
			},
		})
	return statefulSet.Spec.Template.Spec.InitContainers
}

func (c Reconciler) upsertSharedScriptsVolume(statefulSet *apps.StatefulSet, db *api.MySQL) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMySQL {
			configVolumeMount := core.VolumeMount{
				Name:      "init-scripts",
				MountPath: "/scripts",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, configVolumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts
			if db.IsReadReplica() && c.SourceHasSSL(db) {

				caVolumeMount := core.VolumeMount{
					Name:      "source-ca",
					MountPath: "/etc/mysql/server/certs",
				}
				volumeMounts = core_util.UpsertVolumeMount(volumeMounts, caVolumeMount)
				statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts
			}

		}
		if container.Name == "mysql-coordinator" {
			configVolumeMount := core.VolumeMount{
				Name:      "init-scripts",
				MountPath: "/scripts",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, configVolumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

		}
	}
	for i, initContainer := range statefulSet.Spec.Template.Spec.InitContainers {
		if initContainer.Name == "mysql-init" {
			configVolumeMount := core.VolumeMount{
				Name:      "init-scripts",
				MountPath: "/scripts",
			}
			volumeMounts := initContainer.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, configVolumeMount)
			statefulSet.Spec.Template.Spec.InitContainers[i].VolumeMounts = volumeMounts

		}
	}

	configVolume := core.Volume{
		Name: "init-scripts",
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}

	volumes := statefulSet.Spec.Template.Spec.Volumes
	volumes = core_util.UpsertVolume(volumes, configVolume)
	statefulSet.Spec.Template.Spec.Volumes = volumes
	return statefulSet
}

func upsertDataVolume(statefulSet *apps.StatefulSet, db *api.MySQL) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {

		// upsert data volume claim in mysql-coordinator container
		for i, container := range statefulSet.Spec.Template.Spec.Containers {
			if container.Name == "mysql-coordinator" {
				statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = core_util.UpsertVolumeMount(statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts, core.VolumeMount{
					Name:      "data",
					MountPath: "var/lib/mysql",
				})
			}
		}

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
					klog.Infof(`Using "%v" as AccessModes in mysql.Spec.Storage`, core.ReadWriteOnce)
				}

				claim := core.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: *pvcSpec,
				}
				if pvcSpec.StorageClassName != nil {
					pvcAnnotation := map[string]string{
						"volume.beta.kubernetes.io/storage-class": *pvcSpec.StorageClassName,
					}
					claim.Annotations = meta_util.OverwriteKeys(claim.Annotations, pvcAnnotation)
				}
				statefulSet.Spec.VolumeClaimTemplates = core_util.UpsertVolumeClaim(statefulSet.Spec.VolumeClaimTemplates, claim)
			}
			break
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

func (c Reconciler) upsertTLSVolume(sts *apps.StatefulSet, db *api.MySQL) *apps.StatefulSet {
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

func recommendedArgs(db *api.MySQL, myVersion *v1alpha1.MySQLVersion) map[string]string {
	recommendedArgs := map[string]string{}
	// https://dev.mysql.com/doc/refman/5.7/en/innodb-buffer-pool-resize.html
	// recommended innodb_buffer_pool_size value is 50 to 75 percent of system memory
	// Buffer pool size must always be equal to or a multiple of innodb_buffer_pool_chunk_size * innodb_buffer_pool_instances

	available := db.Spec.PodTemplate.Spec.Resources.Limits.Memory()

	var innodbBufferPoolSize, groupReplicationMessageSize int64
	const mb = 1024 * 1024
	if available.Cmp(resource.MustParse("0.75Gi")) <= 0 {
		innodbBufferPoolSize = 128 * mb
		groupReplicationMessageSize = 128 * mb
	} else if available.Cmp(resource.MustParse("1.5Gi")) <= 0 {
		innodbBufferPoolSize = 256 * mb
		groupReplicationMessageSize = 256 * mb
	} else if available.Cmp(resource.MustParse("4Gi")) <= 0 {
		allocateAbleBytes := available.Value()
		allocateAbleBytes -= 1024 * mb
		innodbBufferPoolSize = (allocateAbleBytes / (128 * mb)) * (128 * mb)
		groupReplicationMessageSize = 256 * mb
	} else {
		allocateAbleBytes := float64(available.Value())
		// allocate 70% of the available memory for innodb buffer pool size
		innodbBufferPoolSize = int64((allocateAbleBytes*0.70)/(128*mb)) * 128 * mb
		groupReplicationMessageSize = int64((allocateAbleBytes*.25-256*mb)*0.40/(128*mb)) * 128 * mb
	}
	recommendedArgs["innodb-buffer-pool-size"] = fmt.Sprintf("%d", innodbBufferPoolSize)

	// allocate rest of the memory for group replication cache size
	// https://dev.mysql.com/doc/refman/8.0/en/group-replication-options.html#sysvar_group_replication_message_cache_size
	// recommended minimum loose-group-replication-message-cache-size is 128mb=134217728byte from version 8.0.21
	refVersion := semver.MustParse("8.0.21")
	curVersion := semver.MustParse(myVersion.Spec.Version)
	if curVersion.Compare(refVersion) != -1 {
		recommendedArgs["loose-group-replication-message-cache-size"] = fmt.Sprintf("%d", groupReplicationMessageSize)
	}

	// Sets the binary log expiration period in seconds. After their expiration period ends, binary log files can be automatically removed.
	// Possible removals happen at startup and when the binary log is flushed
	// https://dev.mysql.com/doc/refman/8.0/en/replication-options-binary-log.html#sysvar_binlog_expire_logs_seconds
	// https://mydbops.wordpress.com/2017/04/13/binlog-expiry-now-in-seconds-mysql-8-0/
	refVersion = semver.MustParse("8.0.1")
	curVersion = semver.MustParse(myVersion.Spec.Version)
	if curVersion.Compare(refVersion) != -1 {
		recommendedArgs["binlog-expire-logs-seconds"] = fmt.Sprintf("%d", 7*24*60*60) // 7 days
	} else {
		recommendedArgs["expire-logs-days"] = fmt.Sprintf("%d", 7) // 7 days
	}
	return recommendedArgs
}

func (c Reconciler) SourceHasSSL(db *api.MySQL) bool {
	sourceName := db.Spec.Topology.ReadReplica.SourceRef.Name
	sourceNameSpace := db.Spec.Topology.ReadReplica.SourceRef.Namespace
	dbObj, err := c.DBClient.KubedbV1alpha2().MySQLs(sourceNameSpace).Get(context.TODO(), sourceName, metav1.GetOptions{})
	if err != nil {
		klog.Error("unable to get source mysql object", err)
		return false
	}
	if dbObj.Spec.RequireSSL {
		return true
	}
	return false
}
