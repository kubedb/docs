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

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"gomodules.xyz/pointer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
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

func getMajorPgVersion(postgresVersion *catalog.PostgresVersion) (uint64, error) {
	ver, err := semver.NewVersion(postgresVersion.Spec.Version)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to get postgres major.")
	}
	return ver.Major(), nil
}

func (c *Reconciler) ensureStatefulSet(
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
			in.Spec.Template.Spec.InitContainers = getInitContainers(in, db, postgresVersion)

			in.Spec.Template.Spec.Containers = getContainers(in, db, postgresVersion)

			in = upsertEnv(in, db, postgresVersion, envList)
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
			// No need to upsert volumes
			// Everytime the volume list is generated from the YAML file,
			// so it will contain all required volumes. As there is no support for user provided volume for now,
			// we don't need to use upsert here.
			volumes, pvc := getVolumes(db)
			in.Spec.Template.Spec.Volumes = volumes
			// Upsert volumeClaimTemplates if any
			if pvc != nil {
				in.Spec.VolumeClaimTemplates = core_util.UpsertVolumeClaim(in.Spec.VolumeClaimTemplates, *pvc)
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

	// ensure pdb
	if err := c.CreateStatefulSetPodDisruptionBudget(statefulSet); err != nil {
		return vt, err
	}
	return vt, nil
}

func (c *Reconciler) ensureValidUserForPostgreSQL(db *api.Postgres) error {
	if db.Spec.PodTemplate.Spec.ContainerSecurityContext != nil &&
		db.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsUser != nil &&
		db.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsGroup != nil {
		if pointer.Int64(db.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsUser) == 0 || pointer.Int64(db.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsGroup) == 0 {
			return fmt.Errorf("container's securityContext RunAsUser or RunAsGroup can't be 0")
		} else {
			return nil
		}
	} else {
		return fmt.Errorf("container's securityContext RunAsUser or RunAsGroup can't be null")
	}
}

func (c *Reconciler) ensureCombinedNode(db *api.Postgres, postgresVersion *catalog.PostgresVersion) (kutil.VerbType, error) {
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

func (c *Reconciler) checkStatefulSet(db *api.Postgres) error {
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

func upsertEnv(statefulSet *apps.StatefulSet, db *api.Postgres, postgresVersion *catalog.PostgresVersion, envs []core.EnvVar) *apps.StatefulSet {
	majorPGVersion, err := getMajorPgVersion(postgresVersion)
	if err != nil {
		klog.Error("couldn't get version's major part")
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
			Value: postgresVersion.Spec.Version,
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
			Name:  "PV",
			Value: "/var/pv",
		},
		{
			Name:  "DB_UID",
			Value: strconv.FormatInt(pointer.Int64(db.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsUser), 10),
		},
		{
			Name:  "DB_GID",
			Value: strconv.FormatInt(pointer.Int64(db.Spec.PodTemplate.Spec.ContainerSecurityContext.RunAsGroup), 10),
		},
		{
			Name:  "PGDATA",
			Value: "/var/pv/data",
		},
		{
			Name:  "INITDB",
			Value: "/var/initdb",
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
	} else {
		tlEnv := []core.EnvVar{
			{
				Name:  "SSL",
				Value: "OFF",
			},
		}
		envList = append(envList, tlEnv...)
	}

	// To do this, Upsert Container first
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPostgres || container.Name == api.PostgresCoordinatorContainerName {
			if container.Name == api.ResourceSingularPostgres {
				env := core.EnvVar{
					Name:  "SHARED_BUFFERS",
					Value: api.GetSharedBufferSizeForPostgres(db.Spec.PodTemplate.Spec.Resources.Requests.Memory()),
				}
				container.Env = core_util.UpsertEnvVars(container.Env, env)
			}
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

func (c *Reconciler) upsertMonitoringContainer(statefulSet *apps.StatefulSet, db *api.Postgres, postgresVersion *catalog.PostgresVersion) *apps.StatefulSet {

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
		var volumeMounts []core.VolumeMount
		if db.Spec.TLS != nil {
			volumeMount := core.VolumeMount{

				Name:      exporterTlsVolumeName,
				MountPath: clientTlsVolumeMountPath,
			}

			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
		}
		container.VolumeMounts = volumeMounts

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

func upsertShm(volumes []core.Volume) []core.Volume {
	configVolume := core.Volume{
		Name: "shared-memory",
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{
				Medium: core.StorageMediumMemory,
			},
		},
	}
	volumes = core_util.UpsertVolume(volumes, configVolume)
	return volumes
}

func upsertInitScript(volumes []core.Volume, script core.VolumeSource) []core.Volume {
	volume := core.Volume{
		Name:         "initial-script",
		VolumeSource: script,
	}
	volumes = core_util.UpsertVolume(volumes, volume)
	return volumes
}

func upsertDataVolume(volumes []core.Volume, db *api.Postgres) ([]core.Volume, *core.PersistentVolumeClaim) {
	var pvc *core.PersistentVolumeClaim
	pvcSpec := db.Spec.Storage
	if db.Spec.StorageType == api.StorageTypeEphemeral {
		ed := core.EmptyDirVolumeSource{}
		if pvcSpec != nil {
			if sz, found := pvcSpec.Resources.Requests[core.ResourceStorage]; found {
				ed.SizeLimit = &sz
			}
		}
		volumes = core_util.UpsertVolume(
			volumes,
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
			klog.Infof(`Using "%v" as AccessModes in postgres.Spec.Storage`, core.ReadWriteOnce)
		}

		pvc = &core.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: "data",
			},
			Spec: *pvcSpec,
		}
		if pvcSpec.StorageClassName != nil {
			pvc.Annotations = map[string]string{
				"volume.beta.kubernetes.io/storage-class": *pvcSpec.StorageClassName,
			}
		}
	}

	return volumes, pvc
}

func upsertCustomConfig(volumes []core.Volume, db *api.Postgres) []core.Volume {
	if db.Spec.ConfigSecret != nil {
		configVolume := core.Volume{
			Name: "custom-config",
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: db.Spec.ConfigSecret.Name,
				},
			},
		}
		volumes = core_util.UpsertVolume(volumes, configVolume)
	}
	return volumes
}

func upsertSharedRunScriptsVolume(volumes []core.Volume) []core.Volume {
	configVolume := core.Volume{
		Name: "run-scripts",
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}
	volumes = core_util.UpsertVolume(volumes, configVolume)
	return volumes
}
func upsertSharedScriptsVolume(volumes []core.Volume) []core.Volume {

	configVolume := core.Volume{
		Name: "scripts",
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}
	volumes = core_util.UpsertVolume(volumes, configVolume)

	return volumes
}
func upsertCertificatesVolume(volumes []core.Volume) []core.Volume {

	configVolume := core.Volume{
		Name: sharedTlsVolumeName,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}

	volumes = core_util.UpsertVolume(volumes, configVolume)

	return volumes
}

func getInitContainers(statefulSet *apps.StatefulSet, db *api.Postgres, postgresVersion *catalog.PostgresVersion) []core.Container {

	volumeMounts := []core.VolumeMount{
		{
			Name:      "data",
			MountPath: "/var/pv",
		},
		{
			Name:      "run-scripts",
			MountPath: "/run_scripts",
		},
		{
			Name:      "scripts",
			MountPath: "/scripts",
		},
	}
	if db.Spec.TLS != nil {
		tlsVolumeMounts := []core.VolumeMount{
			{
				Name:      sharedTlsVolumeName,
				MountPath: sharedTlsVolumeMountPath,
			},
			{
				Name:      serverTlsVolumeName,
				MountPath: serverTlsVolumeMountPath,
				ReadOnly:  false,
			},
			{
				Name:      clientTlsVolumeName,
				MountPath: clientTlsVolumeMountPath,
				ReadOnly:  false,
			},
		}
		volumeMounts = core_util.UpsertVolumeMount(volumeMounts, tlsVolumeMounts...)
	}

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
			VolumeMounts: volumeMounts,
		})
	return statefulSet.Spec.Template.Spec.InitContainers
}

func getContainers(statefulSet *apps.StatefulSet, db *api.Postgres, postgresVersion *catalog.PostgresVersion) []core.Container {
	pgLifeCycle := &core.Lifecycle{
		PreStop: &core.Handler{
			Exec: &core.ExecAction{
				Command: []string{"pg_ctl", "-m", "immediate", "-w", "stop"},
			},
		},
	}
	volumeMounts := []core.VolumeMount{
		{
			Name:      "shared-memory",
			MountPath: "/dev/shm",
		},
		{
			Name:      "data",
			MountPath: "/var/pv",
		},
		{
			Name:      "run-scripts",
			MountPath: "/run_scripts",
		},
		{
			Name:      "scripts",
			MountPath: "/scripts",
		},
	}
	if !kmapi.HasCondition(db.Status.Conditions, api.DatabaseDataRestored) {
		initSource := db.Spec.Init
		if initSource != nil && initSource.Script != nil {
			volumeMount := core.VolumeMount{
				Name:      "initial-script",
				MountPath: "/var/initdb",
			}
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
		}
	}
	if db.Spec.ConfigSecret != nil {
		volumeMount := core.VolumeMount{
			Name:      "custom-config",
			MountPath: "/etc/config",
		}
		volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
	}
	if db.Spec.TLS != nil {
		volumeMount := core.VolumeMount{
			Name:      sharedTlsVolumeName,
			MountPath: sharedTlsVolumeMountPath,
		}
		volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
	}

	statefulSet.Spec.Template.Spec.Containers = core_util.UpsertContainer(
		statefulSet.Spec.Template.Spec.Containers,
		core.Container{
			Name: api.ResourceSingularPostgres,
			Command: []string{
				"/scripts/tini",
				"--",
			},
			Args: []string{
				"/scripts/run.sh",
			},
			Image:           postgresVersion.Spec.DB.Image,
			Resources:       db.Spec.PodTemplate.Spec.Resources,
			SecurityContext: db.Spec.PodTemplate.Spec.ContainerSecurityContext,
			LivenessProbe:   db.Spec.PodTemplate.Spec.LivenessProbe,
			ReadinessProbe:  db.Spec.PodTemplate.Spec.ReadinessProbe,
			Lifecycle:       pgLifeCycle,
			VolumeMounts:    volumeMounts,
		})

	coordinatorVolumeMounts := []core.VolumeMount{
		{
			Name:      "data",
			MountPath: "/var/pv",
		},
		{
			Name:      "run-scripts",
			MountPath: "/run_scripts",
		},
	}
	if db.Spec.TLS != nil {
		volumeMount := core.VolumeMount{
			Name:      coordinatorTlsVolumeName,
			MountPath: clientTlsVolumeMountPath,
		}
		coordinatorVolumeMounts = core_util.UpsertVolumeMount(coordinatorVolumeMounts, volumeMount)
	}
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
			VolumeMounts: coordinatorVolumeMounts,
		})
	return statefulSet.Spec.Template.Spec.Containers
}

func getVolumes(db *api.Postgres) ([]core.Volume, *core.PersistentVolumeClaim) {
	var volumes []core.Volume
	var pvc *core.PersistentVolumeClaim

	if !kmapi.HasCondition(db.Status.Conditions, api.DatabaseDataRestored) {
		initSource := db.Spec.Init
		if initSource != nil && initSource.Script != nil {
			volumes = upsertInitScript(volumes, db.Spec.Init.Script.VolumeSource)
		}
	}

	volumes = upsertShm(volumes)
	volumes, pvc = upsertDataVolume(volumes, db)
	volumes = upsertCustomConfig(volumes, db)
	volumes = upsertSharedRunScriptsVolume(volumes)
	volumes = upsertSharedScriptsVolume(volumes)
	if db.Spec.TLS != nil {
		volumes = upsertTLSVolume(volumes, db)
		volumes = upsertCertificatesVolume(volumes)

	}
	return volumes, pvc
}

// adding tls key , cert and ca-cert
func upsertTLSVolume(volumes []core.Volume, db *api.Postgres) []core.Volume {
	serverVolume := core.Volume{
		Name: serverTlsVolumeName,
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: db.GetCertSecretName(api.PostgresServerCert),
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
				SecretName: db.GetCertSecretName(api.PostgresClientCert),
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
				SecretName:  db.GetCertSecretName(api.PostgresMetricsExporterCert),
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
				SecretName:  db.GetCertSecretName(api.PostgresClientCert),
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

	volumes = core_util.UpsertVolume(
		volumes,
		serverVolume,
		clientVolume,
		exporterTLSVolume,
		coordinatorTLSVolume,
	)

	return volumes
}
