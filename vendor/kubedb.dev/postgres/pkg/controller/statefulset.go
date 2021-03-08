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

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/pkg/errors"
	"gomodules.xyz/pointer"
	"gomodules.xyz/version"
	"gomodules.xyz/x/log"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kutil "kmodules.xyz/client-go"
	kmapi "kmodules.xyz/client-go/api/v1"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

const (
	PostgresInitContainerName = "postgres-init-container"

	sharedTlsVolumeMountPath = "/tls/certs"
	clientTlsVolumeMountPath = "/certs/client"
	serverTlsVolumeMountPath = "/certs/server"
	serverTlsVolumeName      = "tls-volume-server"
	clientTlsVolumeName      = "tls-volume-client"
	coordinatorTlsVolumeName = "coordinator-tls-volume"
	sharedTlsVolumeName      = "certs"
	exporterTlsVolumeName    = "exporter-tls-volume"
	TLS_CERT                 = "tls.crt"
	TLS_KEY                  = "tls.key"
	TLS_CA_CERT              = "ca.crt"
	CLIENT_CERT              = "client.crt"
	CLIENT_KEY               = "client.key"
	SERVER_CERT              = "server.crt"
	SERVER_KEY               = "server.key"
)

func getMajorPgVersion(postgres *api.Postgres) (int64, error) {
	ver, err := version.NewVersion(postgres.Spec.Version)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to get postgres major.")
	}
	return ver.Major(), nil
}

func (c *Controller) ensureStatefulSet(
	db *api.Postgres,
	postgresVersion *catalog.PostgresVersion,
	envList []core.EnvVar,
) (kutil.VerbType, error) {

	if err := c.checkStatefulSet(db); err != nil {
		return kutil.VerbUnchanged, err
	}

	statefulSetMeta := metav1.ObjectMeta{
		Name:      db.OffshootName(),
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindPostgres))

	replicas := int32(1)
	if db.Spec.Replicas != nil {
		replicas = pointer.Int32(db.Spec.Replicas)
	}

	statefulSet, vt, err := app_util.CreateOrPatchStatefulSet(
		context.TODO(),
		c.Client,
		statefulSetMeta,
		func(in *apps.StatefulSet) *apps.StatefulSet {
			in.Labels = db.OffshootLabels()
			in.Annotations = db.Spec.PodTemplate.Controller.Annotations
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

			in.Spec.Replicas = pointer.Int32P(replicas)

			in.Spec.ServiceName = db.GoverningServiceName()
			in.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: db.OffshootSelectors(),
			}
			in.Spec.Template.Labels = db.OffshootSelectors()
			in.Spec.Template.Annotations = db.Spec.PodTemplate.Annotations
			in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(in.Spec.Template.Spec.InitContainers, db.Spec.PodTemplate.Spec.InitContainers)
			in.Spec.Template.Spec.InitContainers = getInitContainers(in, postgresVersion)

			in.Spec.Template.Spec.Containers = getContainers(in, db, postgresVersion)

			in = upsertEnv(in, db, envList)
			in = upsertUserEnv(in, db)
			in = upsertPort(in)

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

			in = c.upsertMonitoringContainer(in, db, postgresVersion)

			if !kmapi.HasCondition(db.Status.Conditions, api.DatabaseDataRestored) {
				initSource := db.Spec.Init
				if initSource != nil && initSource.Script != nil {
					in = upsertInitScript(in, db.Spec.Init.Script.VolumeSource)
				}
			}

			in = upsertShm(in)
			in = upsertDataVolume(in, db)
			in = upsertCustomConfig(in, db)
			in = upsertSharedScriptsVolume(in)
			if db.Spec.TLS != nil {
				in = upsertTLSVolume(in, db)
				in = upsertCertificatesVolume(in)

			}

			in.Spec.Template.Spec.ServiceAccountName = db.Spec.PodTemplate.Spec.ServiceAccountName
			in.Spec.UpdateStrategy = apps.StatefulSetUpdateStrategy{
				Type: apps.OnDeleteStatefulSetStrategyType,
			}

			return in
		},
		metav1.PatchOptions{},
	)

	if err != nil {
		return kutil.VerbUnchanged, err
	}

	if vt == kutil.VerbCreated || vt == kutil.VerbPatched {
		// Check StatefulSet Pod status
		if err := c.CheckStatefulSetPodStatus(statefulSet); err != nil {
			return kutil.VerbUnchanged, err
		}

		c.Recorder.Eventf(
			db,
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
		context.TODO(),
		c.Client,
		statefulSet.Namespace,
		statefulSet.Spec.Selector,
		int(pointer.Int32(statefulSet.Spec.Replicas)),
	)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) ensureCombinedNode(db *api.Postgres, postgresVersion *catalog.PostgresVersion) (kutil.VerbType, error) {
	standbyMode := api.WarmPostgresStandbyMode
	streamingMode := api.AsynchronousPostgresStreamingMode

	if db.Spec.StandbyMode != nil {
		standbyMode = *db.Spec.StandbyMode
	}
	if db.Spec.StreamingMode != nil {
		streamingMode = *db.Spec.StreamingMode
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

	return c.ensureStatefulSet(db, postgresVersion, envList)
}

func (c *Controller) checkStatefulSet(db *api.Postgres) error {
	name := db.OffshootName()
	// SatatefulSet for Postgres database
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

func upsertEnv(statefulSet *apps.StatefulSet, db *api.Postgres, envs []core.EnvVar) *apps.StatefulSet {
	majorPGVersion, err := getMajorPgVersion(db)
	if err != nil {
		log.Error("couldn't get version's major part")
	}
	sslMode := db.Spec.SSLMode
	if sslMode == "" {
		if db.Spec.TLS != nil {
			sslMode = api.PostgresSSLModeVerifyFull
		} else {
			sslMode = api.PostgresSSLModeDisable
		}
	}
	clientAuthMode := db.Spec.ClientAuthMode
	if clientAuthMode == "" {
		clientAuthMode = api.ClientAuthModeMD5
	}
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
			Value: db.ServiceName(),
		},
		{
			Name:  "MAX_LAG_BEFORE_FAILOVER",
			Value: strconv.FormatUint(db.Spec.LeaderElection.MaximumLagBeforeFailover, 10),
		},
		{
			Name:  "PERIOD",
			Value: db.Spec.LeaderElection.Period.Duration.String(),
		},
		{
			Name:  "ELECTION_TICK",
			Value: strconv.Itoa(int(db.Spec.LeaderElection.ElectionTick)),
		},
		{
			Name:  "HEARTBEAT_TICK",
			Value: strconv.Itoa(int(db.Spec.LeaderElection.HeartbeatTick)),
		},
		{
			Name: EnvPostgresUser,
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: db.Spec.AuthSecret.Name,
					},
					Key: core.BasicAuthUsernameKey,
				},
			},
		},
		{
			Name: EnvPostgresPassword,
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
			Name:  "PG_VERSION",
			Value: db.Spec.Version,
		},
		{
			Name:  "MAJOR_PG_VERSION",
			Value: strconv.Itoa(int(majorPGVersion)),
		},
		{
			Name:  "CLIENT_AUTH_MODE",
			Value: string(clientAuthMode),
		},
		{
			Name:  "SSL_MODE",
			Value: string(sslMode),
		},
	}

	envList = append(envList, envs...)

	if db.Spec.TLS != nil {
		tlEnv := []core.EnvVar{
			{
				Name:  "SSL",
				Value: "ON",
			},
		}
		envList = append(envList, tlEnv...)
	}

	// To do this, Upsert Container first
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPostgres || container.Name == api.PostgresCoordinatorContainerName {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envList...)
		}
	}
	for i, initContainer := range statefulSet.Spec.Template.Spec.InitContainers {
		if initContainer.Name == PostgresInitContainerName {
			statefulSet.Spec.Template.Spec.InitContainers[i].Env = core_util.UpsertEnvVars(initContainer.Env, envList...)
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
	getPostgresPorts := func() []core.ContainerPort {
		portList := []core.ContainerPort{
			{
				Name:          api.PostgresDatabasePortName,
				ContainerPort: api.PostgresDatabasePort,
				Protocol:      core.ProtocolTCP,
			},
		}
		return portList
	}
	getCoordinatorPorts := func() []core.ContainerPort {
		portList := []core.ContainerPort{
			{
				Name:          api.PostgresCoordinatorPortName,
				ContainerPort: api.PostgresCoordinatorPort, // 2380
				Protocol:      core.ProtocolTCP,
			},
			{
				Name:          api.PostgresCoordinatorClientPortName,
				ContainerPort: api.PostgresCoordinatorClientPort, // 2379
				Protocol:      core.ProtocolTCP,
			},
		}
		return portList
	}

	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPostgres {
			statefulSet.Spec.Template.Spec.Containers[i].Ports = getPostgresPorts()
		} else if container.Name == api.PostgresCoordinatorContainerName {
			statefulSet.Spec.Template.Spec.Containers[i].Ports = getCoordinatorPorts()
		}
	}

	return statefulSet
}

func (c *Controller) upsertMonitoringContainer(statefulSet *apps.StatefulSet, db *api.Postgres, postgresVersion *catalog.PostgresVersion) *apps.StatefulSet {

	if db.Spec.Monitor != nil && db.Spec.Monitor.Agent.Vendor() == mona.VendorPrometheus {
		sslMode := string(db.Spec.SSLMode)
		if sslMode == string(api.PostgresSSLModePrefer) || sslMode == string(api.PostgresSSLModeAllow) {
			sslMode = string(api.PostgresSSLModeRequire)
		}
		cnnstr := fmt.Sprintf("user=${POSTGRES_SOURCE_USER} password='${POSTGRES_SOURCE_PASS}' host=%s port=%d sslmode=%s", api.LocalHost, api.PostgresDatabasePort, sslMode)

		if db.Spec.TLS != nil {
			if db.Spec.SSLMode == api.PostgresSSLModeVerifyCA || db.Spec.SSLMode == api.PostgresSSLModeVerifyFull {
				cnnstr = fmt.Sprintf("%s sslrootcert=%s/ca.crt", cnnstr, clientTlsVolumeMountPath)
			}
			if db.Spec.ClientAuthMode == api.ClientAuthModeCert {
				cnnstr = fmt.Sprintf("%s sslcert=%s/tls.crt sslkey=%s/tls.key", cnnstr,
					clientTlsVolumeMountPath, clientTlsVolumeMountPath)
			}
		}

		cmd := strings.Join(append([]string{
			"/bin/postgres_exporter",
			"--log.level=info",
		}, db.Spec.Monitor.Prometheus.Exporter.Args...), " ")

		commands := []string{
			fmt.Sprintf(`export DATA_SOURCE_NAME="%s"`, cnnstr),
			cmd,
		}
		command := strings.Join(commands, ";")

		container := core.Container{
			Name: "exporter",
			Command: []string{
				"/bin/sh",
			},
			Args: []string{
				"-c",
				command,
			},
			Image:           postgresVersion.Spec.Exporter.Image,
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          mona.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: int32(db.Spec.Monitor.Prometheus.Exporter.Port),
				},
			},
			Env:       db.Spec.Monitor.Prometheus.Exporter.Env,
			Resources: db.Spec.Monitor.Prometheus.Exporter.Resources,
			// Run the container with default User as root user. As when we mount secret it's default owner is root and
			// it has default mode(0600) we can't change that mode to global or higher than 0600.
			// As postgres server don't allow certs file have the global permission.
			// So User must need to be root to have read permission for the certs files. For this reason, We have set UID = 0, 0 is for root user.
			SecurityContext: &core.SecurityContext{
				RunAsUser: pointer.Int64P(0),
			},
		}
		envList := []core.EnvVar{
			{
				Name: "POSTGRES_SOURCE_USER",
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: db.Spec.AuthSecret.Name,
						},
						Key: core.BasicAuthUsernameKey,
					},
				},
			},
			{
				Name: "POSTGRES_SOURCE_PASS",
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
				Name:  "PG_EXPORTER_WEB_LISTEN_ADDRESS",
				Value: fmt.Sprintf(":%d", db.Spec.Monitor.Prometheus.Exporter.Port),
			},
			{
				Name:  "PG_EXPORTER_WEB_TELEMETRY_PATH",
				Value: db.StatsService().Path(),
			},
		}
		container.Env = core_util.UpsertEnvVars(container.Env, envList...)
		containers := statefulSet.Spec.Template.Spec.Containers
		containers = core_util.UpsertContainer(containers, container)
		statefulSet.Spec.Template.Spec.Containers = containers
	}
	return statefulSet
}

func upsertShm(statefulSet *apps.StatefulSet) *apps.StatefulSet {
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

func upsertDataVolume(statefulSet *apps.StatefulSet, db *api.Postgres) *apps.StatefulSet {

	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPostgres || container.Name == api.PostgresCoordinatorContainerName {
			volumeMount := core.VolumeMount{
				Name:      "data",
				MountPath: "/var/pv",
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
			//	break
		}
	}
	return statefulSet
}

func upsertCustomConfig(statefulSet *apps.StatefulSet, db *api.Postgres) *apps.StatefulSet {
	if db.Spec.ConfigSecret != nil {
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

func upsertSharedScriptsVolume(statefulSet *apps.StatefulSet) *apps.StatefulSet {

	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPostgres || container.Name == api.PostgresCoordinatorContainerName {
			configVolumeMount := core.VolumeMount{
				Name:      "scripts",
				MountPath: "/run_scripts",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, configVolumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

		}
	}
	for i, initContainer := range statefulSet.Spec.Template.Spec.InitContainers {
		if initContainer.Name == PostgresInitContainerName {
			configVolumeMount := core.VolumeMount{
				Name:      "scripts",
				MountPath: "/run_scripts",
			}
			volumeMounts := initContainer.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, configVolumeMount)
			statefulSet.Spec.Template.Spec.InitContainers[i].VolumeMounts = volumeMounts

		}
	}

	configVolume := core.Volume{
		Name: "scripts",
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}

	volumes := statefulSet.Spec.Template.Spec.Volumes
	volumes = core_util.UpsertVolume(volumes, configVolume)
	statefulSet.Spec.Template.Spec.Volumes = volumes

	return statefulSet
}

func upsertCertificatesVolume(statefulSet *apps.StatefulSet) *apps.StatefulSet {

	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPostgres {
			configVolumeMount := core.VolumeMount{
				Name:      sharedTlsVolumeName,
				MountPath: sharedTlsVolumeMountPath,
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, configVolumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

		}
	}
	for i, initContainer := range statefulSet.Spec.Template.Spec.InitContainers {
		if initContainer.Name == PostgresInitContainerName {
			configVolumeMount := core.VolumeMount{
				Name:      sharedTlsVolumeName,
				MountPath: sharedTlsVolumeMountPath,
			}
			volumeMounts := initContainer.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, configVolumeMount)
			statefulSet.Spec.Template.Spec.InitContainers[i].VolumeMounts = volumeMounts

		}
	}

	configVolume := core.Volume{
		Name: sharedTlsVolumeName,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}

	volumes := statefulSet.Spec.Template.Spec.Volumes
	volumes = core_util.UpsertVolume(volumes, configVolume)
	statefulSet.Spec.Template.Spec.Volumes = volumes

	return statefulSet
}

func getInitContainers(statefulSet *apps.StatefulSet, postgresVersion *catalog.PostgresVersion) []core.Container {
	statefulSet.Spec.Template.Spec.InitContainers = core_util.UpsertContainer(
		statefulSet.Spec.Template.Spec.InitContainers,
		core.Container{
			Name:  PostgresInitContainerName,
			Image: postgresVersion.Spec.InitContainer.Image,
			Resources: core.ResourceRequirements{
				Limits: core.ResourceList{
					core.ResourceCPU:    resource.MustParse(".200"),
					core.ResourceMemory: resource.MustParse("128Mi"),
				},
				Requests: core.ResourceList{
					core.ResourceCPU:    resource.MustParse(".200"),
					core.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
		})
	return statefulSet.Spec.Template.Spec.InitContainers
}

func getContainers(statefulSet *apps.StatefulSet, postgres *api.Postgres, postgresVersion *catalog.PostgresVersion) []core.Container {
	lifeCycle := &core.Lifecycle{
		PreStop: &core.Handler{
			Exec: &core.ExecAction{
				Command: []string{"pg_ctl", "-m", "immediate", "-w", "stop"},
			},
		},
	}

	statefulSet.Spec.Template.Spec.Containers = core_util.UpsertContainer(
		statefulSet.Spec.Template.Spec.Containers,
		core.Container{
			Name:            api.ResourceSingularPostgres,
			Image:           postgresVersion.Spec.DB.Image,
			Resources:       postgres.Spec.PodTemplate.Spec.Resources,
			SecurityContext: postgres.Spec.PodTemplate.Spec.ContainerSecurityContext,
			LivenessProbe:   postgres.Spec.PodTemplate.Spec.LivenessProbe,
			ReadinessProbe:  postgres.Spec.PodTemplate.Spec.ReadinessProbe,
			Lifecycle:       lifeCycle,
		})
	statefulSet.Spec.Template.Spec.Containers = core_util.UpsertContainer(
		statefulSet.Spec.Template.Spec.Containers,
		core.Container{
			Name:  api.PostgresCoordinatorContainerName,
			Image: postgresVersion.Spec.Coordinator.Image,
			Resources: core.ResourceRequirements{
				Limits: core.ResourceList{
					core.ResourceCPU:    resource.MustParse(".500"),
					core.ResourceMemory: resource.MustParse("256Mi"),
				},
				Requests: core.ResourceList{
					core.ResourceCPU:    resource.MustParse(".500"),
					core.ResourceMemory: resource.MustParse("256Mi"),
				},
			},
		})
	return statefulSet.Spec.Template.Spec.Containers
}

// adding tls key , cert and ca-cert
func upsertTLSVolume(sts *apps.StatefulSet, db *api.Postgres) *apps.StatefulSet {
	for i, container := range sts.Spec.Template.Spec.Containers {
		if container.Name == "exporter" {
			volumeMount := core.VolumeMount{
				Name:      exporterTlsVolumeName,
				MountPath: clientTlsVolumeMountPath,
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			sts.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

		} else if container.Name == api.PostgresCoordinatorContainerName {
			volumeMount := core.VolumeMount{
				Name:      coordinatorTlsVolumeName,
				MountPath: clientTlsVolumeMountPath,
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			sts.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

		}
	}
	for i, initContainer := range sts.Spec.Template.Spec.InitContainers {
		if initContainer.Name == PostgresInitContainerName {
			volumeMount := core.VolumeMount{
				Name:      serverTlsVolumeName,
				MountPath: serverTlsVolumeMountPath,
				ReadOnly:  false,
			}
			volumeMounts := initContainer.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)

			clientVolumeMount := core.VolumeMount{
				Name:      clientTlsVolumeName,
				MountPath: clientTlsVolumeMountPath,
				ReadOnly:  false,
			}
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, clientVolumeMount)
			sts.Spec.Template.Spec.InitContainers[i].VolumeMounts = volumeMounts

		}
	}
	serverVolume := core.Volume{
		Name: serverTlsVolumeName,
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: db.MustCertSecretName(api.PostgresServerCert),
				Items: []core.KeyToPath{
					{
						Key:  TLS_CA_CERT,
						Path: TLS_CA_CERT,
					},
					{
						Key:  TLS_CERT,
						Path: SERVER_CERT,
					},
					{
						Key:  TLS_KEY,
						Path: SERVER_KEY,
					},
				},
			},
		},
	}
	clientVolume := core.Volume{
		Name: clientTlsVolumeName,
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: db.MustCertSecretName(api.PostgresClientCert),
				Items: []core.KeyToPath{
					{
						Key:  TLS_CA_CERT,
						Path: TLS_CA_CERT,
					},
					{
						Key:  TLS_CERT,
						Path: CLIENT_CERT,
					},
					{
						Key:  TLS_KEY,
						Path: CLIENT_KEY,
					},
				},
			},
		},
	}

	exporterTLSVolume := core.Volume{
		Name: exporterTlsVolumeName,
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				DefaultMode: pointer.Int32P(0600),
				SecretName:  db.MustCertSecretName(api.PostgresMetricsExporterCert),
				Items: []core.KeyToPath{
					{
						Key:  TLS_CA_CERT,
						Path: TLS_CA_CERT,
					},
					{
						Key:  TLS_CERT,
						Path: TLS_CERT,
					},
					{
						Key:  TLS_KEY,
						Path: TLS_KEY,
					},
				},
			},
		},
	}
	coordinatorTLSVolume := core.Volume{
		Name: coordinatorTlsVolumeName,
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				DefaultMode: pointer.Int32P(0600),
				SecretName:  db.MustCertSecretName(api.PostgresClientCert),
				Items: []core.KeyToPath{
					{
						Key:  TLS_CA_CERT,
						Path: TLS_CA_CERT,
					},
					{
						Key:  TLS_CERT,
						Path: CLIENT_CERT,
					},
					{
						Key:  TLS_KEY,
						Path: CLIENT_KEY,
					},
				},
			},
		},
	}

	sts.Spec.Template.Spec.Volumes = core_util.UpsertVolume(
		sts.Spec.Template.Spec.Volumes,
		serverVolume,
		clientVolume,
		exporterTLSVolume,
		coordinatorTLSVolume,
	)

	return sts
}
