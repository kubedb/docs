/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

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
	"path/filepath"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"gomodules.xyz/pointer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
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
	configMountPath             = "/etc/config"
	UserListMountPath           = "/var/run/pgbouncer/secret"
	ServingCertMountPath        = "/var/run/pgbouncer/tls/serving"
	TemporaryCertMountPath      = "/tmp/certs"
	PgBouncerInitContainerName  = "pgbouncer-init-container"
	sharedTlsVolumeName         = "certs"
	UpstreamServerCertMountPath = "/var/run/pgbouncer/tls/upstream/server"
)

func (r *Reconciler) ensureStatefulSet(
	db *api.PgBouncer,
	pgbouncerVersion *catalog.PgBouncerVersion,
	envList []core.EnvVar,
) (kutil.VerbType, error) {
	if err := r.checkPBConfigSecret(db); err != nil {
		if kerr.IsNotFound(err) {
			_, err := r.ensureConfigSecret(db)
			if err != nil {
				klog.Infoln(err)
				return kutil.VerbUnchanged, err
			}

		} else {
			klog.Infoln(err)
			return kutil.VerbUnchanged, err
		}
	}

	if err := r.checkStatefulSet(db); err != nil {
		klog.Infoln(err)
		return kutil.VerbUnchanged, err
	}

	statefulSetMeta := metav1.ObjectMeta{
		Name:      db.OffshootName(),
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindPgBouncer))

	replicas := int32(1)
	if db.Spec.Replicas != nil {
		replicas = pointer.Int32(db.Spec.Replicas)
	}

	statefulSet, vt, err := app_util.CreateOrPatchStatefulSet(
		context.TODO(),
		r.Client,
		statefulSetMeta,
		func(in *apps.StatefulSet) *apps.StatefulSet {
			in.Labels = db.PodControllerLabels()
			in.Annotations = db.Spec.PodTemplate.Controller.Annotations
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

			in.Spec.Replicas = pointer.Int32P(replicas)

			in.Spec.ServiceName = db.GoverningServiceName()
			in.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: db.OffshootSelectors(),
			}
			in.Spec.Template.Labels = db.PodLabels()
			in.Spec.Template.Annotations = db.Spec.PodTemplate.Annotations

			in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(in.Spec.Template.Spec.InitContainers, db.Spec.PodTemplate.Spec.InitContainers)
			in.Spec.Template.Spec.InitContainers = getInitContainers(in, db, pgbouncerVersion)
			in.Spec.Template.Spec.Containers = getPBContainer(in, db, pgbouncerVersion)
			in = upsertEnv(in, db, envList)
			volumes := getVolumes(db)
			in.Spec.Template.Spec.Volumes = core_util.MustReplaceVolumes(in.Spec.Template.Spec.Volumes, volumes...)
			in = upsertUserEnv(in, db)

			in.Spec.Template.Spec.NodeSelector = db.Spec.PodTemplate.Spec.NodeSelector
			in.Spec.Template.Spec.Affinity = db.Spec.PodTemplate.Spec.Affinity
			in.Spec.Template.Spec.Tolerations = db.Spec.PodTemplate.Spec.Tolerations
			in.Spec.Template.Spec.ImagePullSecrets = db.Spec.PodTemplate.Spec.ImagePullSecrets
			in.Spec.Template.Spec.PriorityClassName = db.Spec.PodTemplate.Spec.PriorityClassName
			in.Spec.Template.Spec.Priority = db.Spec.PodTemplate.Spec.Priority
			in.Spec.Template.Spec.HostNetwork = db.Spec.PodTemplate.Spec.HostNetwork
			in.Spec.Template.Spec.HostPID = db.Spec.PodTemplate.Spec.HostPID
			in.Spec.Template.Spec.HostIPC = db.Spec.PodTemplate.Spec.HostIPC
			if in.Spec.Template.Spec.SecurityContext == nil {
				in.Spec.Template.Spec.SecurityContext = db.Spec.PodTemplate.Spec.SecurityContext
			}
			in = upsertMonitoringContainer(in, db, pgbouncerVersion)

			return in
		},
		metav1.PatchOptions{},
	)
	if err != nil {
		klog.Infoln(err)
		return kutil.VerbUnchanged, err
	}

	// ensure pdb
	if err := r.SyncStatefulSetPodDisruptionBudget(statefulSet); err != nil {
		klog.Infoln(err)
		return vt, err
	}

	return vt, nil
}

func (r *Reconciler) checkStatefulSet(db *api.PgBouncer) error {
	// Name validation for StatefulSet
	// Check whether PgBouncer's StatefulSet (not managed by KubeDB) already exists
	name := db.OffshootName()
	// SatatefulSet for PgBouncer database
	statefulSet, err := r.Client.AppsV1().StatefulSets(db.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
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

func (c *Reconciler) checkPBConfigSecret(db *api.PgBouncer) error {
	// Name validation for secret
	// Check whether a non-kubedb managed secret by this name already exists
	name := db.ConfigSecretName()
	// secret for PgBouncer
	secret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		} else {
			return err
		}
	}

	if secret.Labels[meta_util.NameLabelKey] != db.ResourceFQN() ||
		secret.Labels[meta_util.InstanceLabelKey] != db.OffshootName() {
		return fmt.Errorf(`intended secret "%v/%v" already exists`, db.Namespace, name)
	}

	return nil
}

// upsertUserEnv add/overwrite env from user provided env in crd spec
func upsertUserEnv(statefulSet *apps.StatefulSet, db *api.PgBouncer) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPgBouncer {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, db.Spec.PodTemplate.Spec.Env...)
			return statefulSet
		}
	}
	return statefulSet
}

func upsertMonitoringContainer(statefulSet *apps.StatefulSet, db *api.PgBouncer, pgbouncerVersion *catalog.PgBouncerVersion) *apps.StatefulSet {
	if db.Spec.Monitor != nil && db.Spec.Monitor.Agent.Vendor() == mona.VendorPrometheus {
		var monitorArgs []string
		if db.Spec.Monitor != nil {
			monitorArgs = db.Spec.Monitor.Prometheus.Exporter.Args
		}

		dataSource := fmt.Sprintf("postgres://%s:@localhost:%d/%s?sslmode=%s", api.PgBouncerAdminUsername, *db.Spec.ConnectionPool.Port, pbAdminDatabase, db.Spec.SSLMode)

		var volumeMounts []core.VolumeMount
		if db.Spec.TLS != nil {
			// TLS is enabled
			if db.Spec.TLS.IssuerRef != nil {
				// mount exporter client-cert in exporter container
				ctClientVolumeMount := getVolumeMountForSharedTls()
				//_, ctClientVolumeMount := (db, api.PgBouncerMetricsExporterCert)
				volumeMounts = append(volumeMounts, *ctClientVolumeMount)

				// update dataSource
				dataSource = fmt.Sprintf("postgres://%s:@localhost:%d/%s?sslmode=%s"+
					"&sslrootcert=%s&sslcert=%s&sslkey=%s",
					api.PgBouncerAdminUsername, *db.Spec.ConnectionPool.Port, pbAdminDatabase, db.Spec.SSLMode,
					filepath.Join(ServingCertMountPath, string(api.PgBouncerMetricsExporterCert), "ca.crt"),
					filepath.Join(ServingCertMountPath, string(api.PgBouncerMetricsExporterCert), "tls.crt"),
					filepath.Join(ServingCertMountPath, string(api.PgBouncerMetricsExporterCert), "tls.key"))

			}
		}
		_ = pgbouncerVersion.Spec.Exporter.Image
		container := core.Container{
			Name: "exporter",
			Args: append([]string{
				fmt.Sprintf("--web.listen-address=:%d", db.Spec.Monitor.Prometheus.Exporter.Port),
			}, monitorArgs...),
			Image:           pgbouncerVersion.Spec.Exporter.Image,
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          mona.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: db.Spec.Monitor.Prometheus.Exporter.Port,
				},
			},
			Env:       db.Spec.Monitor.Prometheus.Exporter.Env,
			Resources: db.Spec.Monitor.Prometheus.Exporter.Resources,
			// we have set the permission for exporter certificate for 70 userid
			// that's why we need to set RunAsUser and RunAsGroup 70
			SecurityContext: &core.SecurityContext{
				RunAsUser:  pointer.Int64P(70),
				RunAsGroup: pointer.Int64P(70),
			},
			VolumeMounts: volumeMounts,
		}

		envList := []core.EnvVar{
			{
				Name:  "DATA_SOURCE_NAME",
				Value: dataSource,
			},
			{
				Name: "PGPASSWORD",
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: db.AuthSecretName(),
						},
						Key: pbAdminPasswordKey,
					},
				},
			},
		}
		container.Env = core_util.UpsertEnvVars(container.Env, envList...)
		containers := statefulSet.Spec.Template.Spec.Containers
		containers = core_util.UpsertContainer(containers, container)
		statefulSet.Spec.Template.Spec.Containers = containers
	}

	return statefulSet
}

func upsertEnv(statefulSet *apps.StatefulSet, db *api.PgBouncer, envs []core.EnvVar) *apps.StatefulSet {
	envList := []core.EnvVar{
		{
			Name:  "NAMESPACE",
			Value: db.Namespace,
		},
		{
			Name:  "PRIMARY_HOST",
			Value: db.ServiceName(),
		},
	}

	envList = append(envList, envs...)

	// To do this, upsert Container first
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPgBouncer {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envList...)
			return statefulSet
		}
	}

	return statefulSet
}

// getInitContainers upsert default initContainer for PgBouncer and return InitContainer
func getInitContainers(statefulSet *apps.StatefulSet, db *api.PgBouncer, pgbouncerVersion *catalog.PgBouncerVersion) []core.Container {
	var envList []core.EnvVar

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

	var volumeMounts []core.VolumeMount
	if db.Spec.TLS != nil {
		tlsVolumeMounts := []core.VolumeMount{
			{
				Name:      sharedTlsVolumeName,
				MountPath: ServingCertMountPath,
			},
			{
				Name:      db.GetCertSecretName(api.PgBouncerClientCert),
				MountPath: filepath.Join(TemporaryCertMountPath, string(api.PgBouncerClientCert)),
			},
			{
				Name:      db.GetCertSecretName(api.PgBouncerServerCert),
				MountPath: filepath.Join(TemporaryCertMountPath, string(api.PgBouncerServerCert)),
			},
			{
				Name:      db.GetCertSecretName(api.PgBouncerMetricsExporterCert),
				MountPath: filepath.Join(TemporaryCertMountPath, string(api.PgBouncerMetricsExporterCert)),
			},
		}
		volumeMounts = core_util.UpsertVolumeMount(volumeMounts, tlsVolumeMounts...)
	}

	statefulSet.Spec.Template.Spec.InitContainers = core_util.UpsertContainer(
		statefulSet.Spec.Template.Spec.InitContainers,
		core.Container{
			Name:            PgBouncerInitContainerName,
			Image:           pgbouncerVersion.Spec.InitContainer.Image,
			ImagePullPolicy: core.PullIfNotPresent,

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
			Env:          envList,
			SecurityContext: &core.SecurityContext{
				RunAsUser: pointer.Int64P(0),
			},
		})
	return statefulSet.Spec.Template.Spec.InitContainers
}

// getPBContainer return default container for pgbouncer statefulset
func getPBContainer(statefulSet *apps.StatefulSet, db *api.PgBouncer, pgbouncerVersion *catalog.PgBouncerVersion) []core.Container {
	image := pgbouncerVersion.Spec.PgBouncer.Image
	volumeMounts := getVolumeMountForPBContainer(db)

	statefulSet.Spec.Template.Spec.Containers = core_util.UpsertContainer(
		statefulSet.Spec.Template.Spec.Containers,
		core.Container{
			Name: api.ResourceSingularPgBouncer,
			// TODO: decide what to do with Args and Env

			Env: []core.EnvVar{
				{
					Name:  "PGBOUNCER_LISTEN_PORT",
					Value: fmt.Sprintf("%d", *db.Spec.ConnectionPool.Port),
				},
			},

			Image:           image,
			ImagePullPolicy: core.PullIfNotPresent,

			VolumeMounts: volumeMounts,
			Ports: []core.ContainerPort{
				{
					Name:          api.PgBouncerDatabasePortName,
					ContainerPort: *db.Spec.ConnectionPool.Port,
					Protocol:      core.ProtocolTCP,
				},
			},

			Resources:       db.Spec.PodTemplate.Spec.Resources,
			SecurityContext: db.Spec.PodTemplate.Spec.ContainerSecurityContext,
			LivenessProbe:   db.Spec.PodTemplate.Spec.LivenessProbe,
			ReadinessProbe:  db.Spec.PodTemplate.Spec.ReadinessProbe,
			Lifecycle:       db.Spec.PodTemplate.Spec.Lifecycle,
		})
	return statefulSet.Spec.Template.Spec.Containers
}
