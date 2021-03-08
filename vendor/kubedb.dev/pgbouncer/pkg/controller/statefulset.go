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
	"kubedb.dev/apimachinery/pkg/eventer"

	"gomodules.xyz/pointer"
	"gomodules.xyz/x/log"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
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
	UpstreamServerCertMountPath = "/var/run/pgbouncer/tls/upstream/server"
)

func (c *Controller) ensureStatefulSet(
	db *api.PgBouncer,
	pgbouncerVersion *catalog.PgBouncerVersion,
	envList []core.EnvVar,
) (kutil.VerbType, error) {
	if err := c.checkSecret(db); err != nil {
		if kerr.IsNotFound(err) {
			_, err := c.ensureConfigSecret(db)
			if err != nil {
				log.Infoln(err)
				return kutil.VerbUnchanged, err
			}

		} else {
			log.Infoln(err)
			return kutil.VerbUnchanged, err
		}
	}

	if err := c.checkStatefulSet(db); err != nil {
		log.Infoln(err)
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
	image := pgbouncerVersion.Spec.Server.Image

	statefulSet, vt, err := app_util.CreateOrPatchStatefulSet(
		context.TODO(),
		c.Client,
		statefulSetMeta,
		func(in *apps.StatefulSet) *apps.StatefulSet {
			in.Annotations = db.Annotations //TODO: actual annotations
			in.Labels = db.OffshootLabels()
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

			in.Spec.Replicas = pointer.Int32P(replicas)

			in.Spec.ServiceName = db.GoverningServiceName()
			in.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: db.OffshootSelectors(),
			}
			in.Spec.Template.Labels = db.OffshootSelectors()

			var volumes []core.Volume
			secretVolume := core.Volume{
				Name: db.OffshootName(),
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: db.OffshootName(),
					},
				},
			}
			volumes = append(volumes, secretVolume)

			var volumeMounts []core.VolumeMount
			secretVolumeMount := core.VolumeMount{
				Name:      db.OffshootName(),
				MountPath: configMountPath,
			}
			volumeMounts = append(volumeMounts, secretVolumeMount)

			cfgVolume, cfgVolumeMount := c.getVolumeAndVolumeMountForAuthSecret(db)
			volumes = append(volumes, *cfgVolume)
			volumeMounts = append(volumeMounts, *cfgVolumeMount)

			if db.Spec.TLS != nil {
				//TLS is enabled
				//mount client crt (CT is short for client-tls)
				if db.Spec.TLS.IssuerRef != nil {
					servingServerSecretVolume, servingServerSecretVolumeMount := c.getVolumeAndVolumeMountForCertificate(db, api.PgBouncerServerCert)
					volumes = append(volumes, *servingServerSecretVolume)
					volumeMounts = append(volumeMounts, *servingServerSecretVolumeMount)

					servingClientSecretVolume, servingClientVolumeMount := c.getVolumeAndVolumeMountForCertificate(db, api.PgBouncerClientCert)
					volumes = append(volumes, *servingClientSecretVolume)
					volumeMounts = append(volumeMounts, *servingClientVolumeMount)

					//add exporter certificate volume
					exporterSecretVolume, _ := c.getVolumeAndVolumeMountForCertificate(db, api.PgBouncerMetricsExporterCert)
					volumes = append(volumes, *exporterSecretVolume)
				}
			}

			in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(in.Spec.Template.Spec.InitContainers, db.Spec.PodTemplate.Spec.InitContainers)
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
				in.Spec.Template.Spec.Containers,
				core.Container{
					Name: api.ResourceSingularPgBouncer,
					//TODO: decide what to do with Args and Env
					//Args: append([]string{
					//	fmt.Sprintf(`--enable-analytics=%v`, c.EnableAnalytics),
					//}, c.LoggerOptions.ToFlags()...),
					Env: []core.EnvVar{
						{
							Name:  "PGBOUNCER_PORT",
							Value: fmt.Sprintf("%d", *db.Spec.ConnectionPool.Port),
						},
					},

					Image:           image,
					ImagePullPolicy: core.PullIfNotPresent,
					VolumeMounts:    volumeMounts,
					Ports: []core.ContainerPort{
						{
							Name:          api.PgBouncerDatabasePortName,
							ContainerPort: *db.Spec.ConnectionPool.Port,
							Protocol:      core.ProtocolTCP,
						},
					},

					Resources:       db.Spec.PodTemplate.Spec.Container.Resources,
					SecurityContext: db.Spec.PodTemplate.Spec.Container.SecurityContext,
					LivenessProbe:   db.Spec.PodTemplate.Spec.Container.LivenessProbe,
					ReadinessProbe:  db.Spec.PodTemplate.Spec.Container.ReadinessProbe,
					Lifecycle:       db.Spec.PodTemplate.Spec.Container.Lifecycle,
				})
			in = upsertEnv(in, db, envList)
			in.Spec.Template.Spec.Volumes = volumes
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
			in.Spec.Template.Spec.SecurityContext = db.Spec.PodTemplate.Spec.SecurityContext
			in = c.upsertMonitoringContainer(in, db, pgbouncerVersion)

			return in
		},
		metav1.PatchOptions{},
	)

	if err != nil {
		log.Infoln(err)
		return kutil.VerbUnchanged, err
	}

	if vt == kutil.VerbCreated || vt == kutil.VerbPatched {
		// Check StatefulSet Pod status
		if err := c.CheckStatefulSetPodStatus(statefulSet); err != nil {
			return kutil.VerbUnchanged, err
		}

		c.recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet",
			vt,
		)
	}

	// ensure pdb
	if err := c.CreateStatefulSetPodDisruptionBudget(statefulSet); err != nil {
		log.Infoln(err)
		return vt, err
	}

	return vt, nil
}

func (c *Controller) CheckStatefulSetPodStatus(statefulSet *apps.StatefulSet) error {
	err := WaitUntilPodRunningBySelector(
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

func (c *Controller) checkStatefulSet(db *api.PgBouncer) error {
	//Name validation for StatefulSet
	// Check whether PgBouncer's StatefulSet (not managed by KubeDB) already exists
	name := db.OffshootName()
	// SatatefulSet for PgBouncer database
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

func (c *Controller) checkSecret(db *api.PgBouncer) error {
	//Name validation for secret
	// Check whether a non-kubedb managed secret by this name already exists
	name := db.OffshootName()
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
		secret.Labels[meta_util.InstanceLabelKey] != name {
		return fmt.Errorf(`intended secret "%v/%v" already exists`, db.Namespace, name)
	}

	return nil
}

// upsertUserEnv add/overwrite env from user provided env in crd spec
func upsertUserEnv(statefulSet *apps.StatefulSet, db *api.PgBouncer) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPgBouncer {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, db.Spec.PodTemplate.Spec.Container.Env...)
			return statefulSet
		}
	}
	return statefulSet
}

func (c *Controller) upsertMonitoringContainer(statefulSet *apps.StatefulSet, db *api.PgBouncer, pgbouncerVersion *catalog.PgBouncerVersion) *apps.StatefulSet {
	if db.Spec.Monitor != nil && db.Spec.Monitor.Agent.Vendor() == mona.VendorPrometheus {
		var monitorArgs []string
		if db.Spec.Monitor != nil {
			monitorArgs = db.Spec.Monitor.Prometheus.Exporter.Args
		}

		dataSource := fmt.Sprintf("postgres://%s:@localhost:%d/%s?sslmode=disable", api.PgBouncerAdminUsername, *db.Spec.ConnectionPool.Port, pbAdminDatabase)

		var volumeMounts []core.VolumeMount
		if db.Spec.TLS != nil {
			// TLS is enabled
			if db.Spec.TLS.IssuerRef != nil {
				// mount exporter client-cert in exporter container
				_, ctClientVolumeMount := c.getVolumeAndVolumeMountForCertificate(db, api.PgBouncerMetricsExporterCert)
				volumeMounts = append(volumeMounts, *ctClientVolumeMount)

				// update dataSource
				dataSource = fmt.Sprintf("postgres://%s:@localhost:%d/%s?sslmode=verify-full"+
					"&sslrootcert=%s&sslcert=%s&sslkey=%s",
					api.PgBouncerAdminUsername, *db.Spec.ConnectionPool.Port, pbAdminDatabase,
					filepath.Join(ServingCertMountPath, string(api.PgBouncerClientCert), "ca.crt"),
					filepath.Join(ServingCertMountPath, string(api.PgBouncerClientCert), "tls.crt"),
					filepath.Join(ServingCertMountPath, string(api.PgBouncerClientCert), "tls.key"))

			}
		}

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
			Env:             db.Spec.Monitor.Prometheus.Exporter.Env,
			Resources:       db.Spec.Monitor.Prometheus.Exporter.Resources,
			SecurityContext: db.Spec.Monitor.Prometheus.Exporter.SecurityContext,
			VolumeMounts:    volumeMounts,
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

func WaitUntilPodRunningBySelector(kubeClient kubernetes.Interface, namespace string, selector *metav1.LabelSelector, count int) error {
	r, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return err
	}

	return wait.PollImmediate(kutil.RetryInterval, kutil.GCTimeout, func() (bool, error) {
		podList, err := kubeClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: r.String(),
		})
		if err != nil {
			return true, nil
		}

		if len(podList.Items) != count {
			return true, nil
		}

		for _, pod := range podList.Items {
			runningAndReady, _ := core_util.PodRunningAndReady(pod)
			if !runningAndReady {
				return false, nil
			}
		}
		return true, nil
	})
}
