/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

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
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

const (
	configMountPath             = "/etc/config"
	UserListMountPath           = "/var/run/pgbouncer/secret"
	ServingServerCertMountPath  = "/var/run/pgbouncer/tls/serving/server"
	ServingClientCertMountPath  = "/var/run/pgbouncer/tls/serving/client"
	UpstreamServerCertMountPath = "/var/run/pgbouncer/tls/upstream/server"
)

func (c *Controller) ensureStatefulSet(
	pgbouncer *api.PgBouncer,
	pgbouncerVersion *catalog.PgBouncerVersion,
	envList []core.EnvVar,
) (kutil.VerbType, error) {
	if err := c.checkConfigMap(pgbouncer); err != nil {
		if kerr.IsNotFound(err) {
			_, err := c.ensureConfigMapFromCRD(pgbouncer)
			if err != nil {
				log.Infoln(err)
				return kutil.VerbUnchanged, err
			}

		} else {
			log.Infoln(err)
			return kutil.VerbUnchanged, err
		}
	}

	if err := c.checkStatefulSet(pgbouncer); err != nil {
		log.Infoln(err)
		return kutil.VerbUnchanged, err
	}

	statefulSetMeta := metav1.ObjectMeta{
		Name:      pgbouncer.OffshootName(),
		Namespace: pgbouncer.Namespace,
	}

	owner := metav1.NewControllerRef(pgbouncer, api.SchemeGroupVersion.WithKind(api.ResourceKindPgBouncer))

	replicas := int32(1)
	if pgbouncer.Spec.Replicas != nil {
		replicas = types.Int32(pgbouncer.Spec.Replicas)
	}
	image := pgbouncerVersion.Spec.Server.Image

	statefulSet, vt, err := app_util.CreateOrPatchStatefulSet(
		context.TODO(),
		c.Client,
		statefulSetMeta,
		func(in *apps.StatefulSet) *apps.StatefulSet {
			in.Annotations = pgbouncer.Annotations //TODO: actual annotations
			in.Labels = pgbouncer.OffshootLabels()
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

			in.Spec.Replicas = types.Int32P(replicas)

			in.Spec.ServiceName = c.GoverningService
			in.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: pgbouncer.OffshootSelectors(),
			}
			in.Spec.Template.Labels = pgbouncer.OffshootSelectors()

			var volumes []core.Volume
			configMapVolume := core.Volume{
				Name: pgbouncer.OffshootName(),
				VolumeSource: core.VolumeSource{
					ConfigMap: &core.ConfigMapVolumeSource{
						LocalObjectReference: core.LocalObjectReference{
							Name: pgbouncer.OffshootName(),
						},
					},
				},
			}
			volumes = append(volumes, configMapVolume)

			var volumeMounts []core.VolumeMount
			configMapVolumeMount := core.VolumeMount{
				Name:      pgbouncer.OffshootName(),
				MountPath: configMountPath,
			}
			volumeMounts = append(volumeMounts, configMapVolumeMount)

			secretVolume, secretVolumeMount, err := c.getVolumeAndVolumeMountForDefaultUserList(pgbouncer)
			if err == nil {
				volumes = append(volumes, *secretVolume)
				volumeMounts = append(volumeMounts, *secretVolumeMount)
			}

			if pgbouncer.Spec.TLS != nil {
				//TLS is enabled
				//mount client crt (CT is short for client-tls)
				if pgbouncer.Spec.TLS.IssuerRef != nil {
					servingServerSecretVolume, servingServerSecretVolumeMount, err := c.getVolumeAndVolumeMountForServingServerCertificate(pgbouncer)
					if err == nil {
						volumes = append(volumes, *servingServerSecretVolume)
						volumeMounts = append(volumeMounts, *servingServerSecretVolumeMount)
					}
					servingClientSecretVolume, servingClientVolumeMount, err := c.getVolumeAndVolumeMountForServingClientCertificate(pgbouncer)
					if err == nil {
						volumes = append(volumes, *servingClientSecretVolume)
						volumeMounts = append(volumeMounts, *servingClientVolumeMount)
					}
					//add exporter certificate volume
					exporterSecretVolume, _, err := c.getVolumeAndVolumeMountForExporterClientCertificate(pgbouncer)
					if err == nil {
						volumes = append(volumes, *exporterSecretVolume)
					}

				}
			}

			in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(in.Spec.Template.Spec.InitContainers, pgbouncer.Spec.PodTemplate.Spec.InitContainers)
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
							Value: fmt.Sprintf("%d", *pgbouncer.Spec.ConnectionPool.Port),
						},
					},

					Image:           image,
					ImagePullPolicy: core.PullIfNotPresent,
					VolumeMounts:    volumeMounts,

					Resources:      pgbouncer.Spec.PodTemplate.Spec.Resources,
					LivenessProbe:  pgbouncer.Spec.PodTemplate.Spec.LivenessProbe,
					ReadinessProbe: pgbouncer.Spec.PodTemplate.Spec.ReadinessProbe,
					Lifecycle:      pgbouncer.Spec.PodTemplate.Spec.Lifecycle,
				})
			in = upsertEnv(in, pgbouncer, envList)
			in.Spec.Template.Spec.Volumes = volumes
			in = upsertUserEnv(in, pgbouncer)
			in = upsertPort(in, pgbouncer)
			in.Spec.Template.Spec.NodeSelector = pgbouncer.Spec.PodTemplate.Spec.NodeSelector
			in.Spec.Template.Spec.Affinity = pgbouncer.Spec.PodTemplate.Spec.Affinity
			in.Spec.Template.Spec.Tolerations = pgbouncer.Spec.PodTemplate.Spec.Tolerations
			in.Spec.Template.Spec.ImagePullSecrets = pgbouncer.Spec.PodTemplate.Spec.ImagePullSecrets
			in.Spec.Template.Spec.PriorityClassName = pgbouncer.Spec.PodTemplate.Spec.PriorityClassName
			in.Spec.Template.Spec.Priority = pgbouncer.Spec.PodTemplate.Spec.Priority
			if in.Spec.Template.Spec.SecurityContext != nil {
				in.Spec.Template.Spec.SecurityContext = pgbouncer.Spec.PodTemplate.Spec.SecurityContext
			}
			in = c.upsertMonitoringContainer(in, pgbouncer, pgbouncerVersion)

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
			pgbouncer,
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
		int(types.Int32(statefulSet.Spec.Replicas)),
	)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) checkStatefulSet(pgbouncer *api.PgBouncer) error {
	//Name validation for StatefulSet
	// Check whether PgBouncer's StatefulSet (not managed by KubeDB) already exists
	name := pgbouncer.OffshootName()
	// SatatefulSet for PgBouncer database
	statefulSet, err := c.Client.AppsV1().StatefulSets(pgbouncer.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		} else {
			return err
		}
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindPgBouncer ||
		statefulSet.Labels[api.LabelDatabaseName] != name {
		return fmt.Errorf(`intended statefulSet "%v/%v" already exists`, pgbouncer.Namespace, name)
	}

	return nil
}

func (c *Controller) checkConfigMap(pgbouncer *api.PgBouncer) error {
	//Name validation for configMap
	// Check whether a non-kubedb managed configMap by this name already exists
	name := pgbouncer.OffshootName()
	// configMap for PgBouncer
	configMap, err := c.Client.CoreV1().ConfigMaps(pgbouncer.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		} else {
			return err
		}
	}

	if configMap.Labels[api.LabelDatabaseKind] != api.ResourceKindPgBouncer ||
		configMap.Labels[api.LabelDatabaseName] != name {
		return fmt.Errorf(`intended configMap "%v/%v" already exists`, pgbouncer.Namespace, name)
	}

	return nil
}

// upsertUserEnv add/overwrite env from user provided env in crd spec
func upsertUserEnv(statefulSet *apps.StatefulSet, pgbouncer *api.PgBouncer) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPgBouncer {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, pgbouncer.Spec.PodTemplate.Spec.Env...)
			return statefulSet
		}
	}
	return statefulSet
}

func upsertPort(statefulSet *apps.StatefulSet, pgbouncer *api.PgBouncer) *apps.StatefulSet {
	getPorts := func() []core.ContainerPort {
		portList := []core.ContainerPort{
			{
				Name:          PgBouncerPortName,
				ContainerPort: *pgbouncer.Spec.ConnectionPool.Port,
				Protocol:      core.ProtocolTCP,
			},
		}
		return portList
	}

	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularPgBouncer {
			statefulSet.Spec.Template.Spec.Containers[i].Ports = getPorts()
			return statefulSet
		}
	}

	return statefulSet
}

func (c *Controller) upsertMonitoringContainer(statefulSet *apps.StatefulSet, pgbouncer *api.PgBouncer, pgbouncerVersion *catalog.PgBouncerVersion) *apps.StatefulSet {
	if pgbouncer.GetMonitoringVendor() == mona.VendorPrometheus {
		var monitorArgs []string
		if pgbouncer.Spec.Monitor != nil {
			monitorArgs = pgbouncer.Spec.Monitor.Prometheus.Exporter.Args
		}

		adminSecretSpec := c.GetDefaultSecretSpec(pgbouncer)
		err := c.isSecretExists(adminSecretSpec.ObjectMeta)
		if err != nil {
			log.Infoln(err)
			return statefulSet //Dont make changes if error occurs
		}

		dataSource := fmt.Sprintf("postgres://%s:@localhost:%d/%s?sslmode=disable", pbAdminUser, *pgbouncer.Spec.ConnectionPool.Port, pbAdminDatabase)

		var volumeMounts []core.VolumeMount
		if pgbouncer.Spec.TLS != nil {
			// TLS is enabled
			if pgbouncer.Spec.TLS.IssuerRef != nil {
				// mount exporter client-cert in exporter container
				_, ctClientVolumeMount, err := c.getVolumeAndVolumeMountForExporterClientCertificate(pgbouncer)
				if err == nil {
					volumeMounts = append(volumeMounts, *ctClientVolumeMount)
				}
				// update dataSource
				dataSource = fmt.Sprintf("postgres://%s:@localhost:%d/%s?sslmode=verify-full"+
					"&sslrootcert=%s&sslcert=%s&sslkey=%s",
					pbAdminUser, *pgbouncer.Spec.ConnectionPool.Port, pbAdminDatabase,
					filepath.Join(ServingClientCertMountPath, "ca.crt"),
					filepath.Join(ServingClientCertMountPath, "tls.crt"),
					filepath.Join(ServingClientCertMountPath, "tls.key"))

			}
		}

		container := core.Container{
			Name: "exporter",
			Args: append([]string{
				fmt.Sprintf("--web.listen-address=:%d", pgbouncer.Spec.Monitor.Prometheus.Exporter.Port),
			}, monitorArgs...),
			Image:           pgbouncerVersion.Spec.Exporter.Image,
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          api.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: pgbouncer.Spec.Monitor.Prometheus.Exporter.Port,
				},
			},
			Env:             pgbouncer.Spec.Monitor.Prometheus.Exporter.Env,
			Resources:       pgbouncer.Spec.Monitor.Prometheus.Exporter.Resources,
			SecurityContext: pgbouncer.Spec.Monitor.Prometheus.Exporter.SecurityContext,
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
							Name: adminSecretSpec.Name,
						},
						Key: pbAdminPassword,
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

func upsertEnv(statefulSet *apps.StatefulSet, pgbouncer *api.PgBouncer, envs []core.EnvVar) *apps.StatefulSet {
	envList := []core.EnvVar{
		{
			Name:  "NAMESPACE",
			Value: pgbouncer.Namespace,
		},
		{
			Name:  "PRIMARY_HOST",
			Value: pgbouncer.ServiceName(),
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
