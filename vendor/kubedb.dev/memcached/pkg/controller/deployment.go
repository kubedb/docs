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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

const (
	CONFIG_SOURCE_VOLUME           = "custom-config"
	CONFIG_SOURCE_VOLUME_MOUNTPATH = "/usr/config/"
	DATA_SOURCE_VOLUME             = "data-volume"
	DATA_SOURCE_VOLUME_MOUNTPATH   = "/data"
)

func (c *Controller) ensureDeployment(memcached *api.Memcached) (kutil.VerbType, error) {
	if err := c.checkDeployment(memcached); err != nil {
		return kutil.VerbUnchanged, err
	}

	// Create statefulSet for Memcached database
	deployment, vt, err := c.createDeployment(memcached)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	// Check Deployment Pod status
	if vt != kutil.VerbUnchanged {
		if err := app_util.WaitUntilDeploymentReady(context.TODO(), c.Client, deployment.ObjectMeta); err != nil {
			return kutil.VerbUnchanged, err
		}
		c.recorder.Eventf(
			memcached,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet",
			vt,
		)
	}

	// ensure pdb
	if err := c.CreateDeploymentPodDisruptionBudget(deployment); err != nil {
		return vt, err
	}
	return vt, nil
}

func (c *Controller) checkDeployment(memcached *api.Memcached) error {
	// Deployment for Memcached database
	deployment, err := c.Client.AppsV1().Deployments(memcached.Namespace).Get(context.TODO(), memcached.OffshootName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}
	if deployment.Labels[api.LabelDatabaseKind] != api.ResourceKindMemcached ||
		deployment.Labels[api.LabelDatabaseName] != memcached.Name {
		return fmt.Errorf(`intended deployment "%v/%v" already exists`, memcached.Namespace, memcached.OffshootName())
	}
	return nil
}

func (c *Controller) createDeployment(memcached *api.Memcached) (*apps.Deployment, kutil.VerbType, error) {
	deploymentMeta := metav1.ObjectMeta{
		Name:      memcached.OffshootName(),
		Namespace: memcached.Namespace,
	}

	owner := metav1.NewControllerRef(memcached, api.SchemeGroupVersion.WithKind(api.ResourceKindMemcached))

	memcachedVersion, err := c.ExtClient.CatalogV1alpha1().MemcachedVersions().Get(context.TODO(), string(memcached.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	return app_util.CreateOrPatchDeployment(context.TODO(), c.Client, deploymentMeta, func(in *apps.Deployment) *apps.Deployment {
		in.Labels = memcached.OffshootLabels()
		in.Annotations = memcached.Spec.PodTemplate.Controller.Annotations
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

		in.Spec.Replicas = memcached.Spec.Replicas
		in.Spec.Template.Labels = in.Labels
		in.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: memcached.OffshootSelectors(),
		}
		in.Spec.Template.Labels = memcached.OffshootSelectors()
		in.Spec.Template.Annotations = memcached.Spec.PodTemplate.Annotations
		in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(in.Spec.Template.Spec.InitContainers, memcached.Spec.PodTemplate.Spec.InitContainers)
		in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
			Name:            api.ResourceSingularMemcached,
			Image:           memcachedVersion.Spec.DB.Image,
			ImagePullPolicy: core.PullIfNotPresent,
			Args:            memcached.Spec.PodTemplate.Spec.Args,
			Ports: []core.ContainerPort{
				{
					Name:          "db",
					ContainerPort: 11211,
					Protocol:      core.ProtocolTCP,
				},
			},
			Resources:      memcached.Spec.PodTemplate.Spec.Resources,
			LivenessProbe:  memcached.Spec.PodTemplate.Spec.LivenessProbe,
			ReadinessProbe: memcached.Spec.PodTemplate.Spec.ReadinessProbe,
			Lifecycle:      memcached.Spec.PodTemplate.Spec.Lifecycle,
		})
		if memcached.GetMonitoringVendor() == mona.VendorPrometheus {
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
				Name: "exporter",
				Args: append([]string{
					fmt.Sprintf("--web.listen-address=:%v", memcached.Spec.Monitor.Prometheus.Exporter.Port),
					fmt.Sprintf("--web.telemetry-path=%v", memcached.StatsService().Path()),
				}, memcached.Spec.Monitor.Prometheus.Exporter.Args...),
				Image:           memcachedVersion.Spec.Exporter.Image,
				ImagePullPolicy: core.PullIfNotPresent,
				Ports: []core.ContainerPort{
					{
						Name:          api.PrometheusExporterPortName,
						Protocol:      core.ProtocolTCP,
						ContainerPort: memcached.Spec.Monitor.Prometheus.Exporter.Port,
					},
				},
				Env:             memcached.Spec.Monitor.Prometheus.Exporter.Env,
				Resources:       memcached.Spec.Monitor.Prometheus.Exporter.Resources,
				SecurityContext: memcached.Spec.Monitor.Prometheus.Exporter.SecurityContext,
			})
		}
		in = upsertUserEnv(in, memcached)
		in = upsertCustomConfig(in, memcached)
		in = upsertDataVolume(in, memcached, memcachedVersion.Spec.DB.Image)

		in.Spec.Template.Spec.NodeSelector = memcached.Spec.PodTemplate.Spec.NodeSelector
		in.Spec.Template.Spec.Affinity = memcached.Spec.PodTemplate.Spec.Affinity
		if memcached.Spec.PodTemplate.Spec.SchedulerName != "" {
			in.Spec.Template.Spec.SchedulerName = memcached.Spec.PodTemplate.Spec.SchedulerName
		}
		in.Spec.Template.Spec.Tolerations = memcached.Spec.PodTemplate.Spec.Tolerations
		in.Spec.Template.Spec.ImagePullSecrets = memcached.Spec.PodTemplate.Spec.ImagePullSecrets
		in.Spec.Template.Spec.PriorityClassName = memcached.Spec.PodTemplate.Spec.PriorityClassName
		in.Spec.Template.Spec.Priority = memcached.Spec.PodTemplate.Spec.Priority
		in.Spec.Template.Spec.SecurityContext = memcached.Spec.PodTemplate.Spec.SecurityContext
		in.Spec.Template.Spec.ServiceAccountName = memcached.Spec.PodTemplate.Spec.ServiceAccountName
		in.Spec.Strategy = apps.DeploymentStrategy{
			Type: apps.RollingUpdateDeploymentStrategyType,
		}

		return in
	}, metav1.PatchOptions{})
}

// upsertUserEnv add/overwrite env from user provided env in crd spec
func upsertUserEnv(deployment *apps.Deployment, memcached *api.Memcached) *apps.Deployment {
	for i, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMemcached {
			deployment.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, memcached.Spec.PodTemplate.Spec.Env...)
			return deployment
		}
	}
	return deployment
}

// upsertCustomConfig insert custom configuration volume if provided.
func upsertCustomConfig(deployment *apps.Deployment, memcached *api.Memcached) *apps.Deployment {
	if memcached.Spec.ConfigSource != nil {
		for i, container := range deployment.Spec.Template.Spec.Containers {
			if container.Name == api.ResourceSingularMemcached {

				configSourceVolumeMount := core.VolumeMount{
					Name:      CONFIG_SOURCE_VOLUME,
					MountPath: CONFIG_SOURCE_VOLUME_MOUNTPATH,
				}

				volumeMounts := container.VolumeMounts
				volumeMounts = core_util.UpsertVolumeMount(volumeMounts, configSourceVolumeMount)
				deployment.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

				configSourceVolume := core.Volume{
					Name:         CONFIG_SOURCE_VOLUME,
					VolumeSource: *memcached.Spec.ConfigSource,
				}

				volumes := deployment.Spec.Template.Spec.Volumes
				volumes = core_util.UpsertVolume(volumes, configSourceVolume)
				deployment.Spec.Template.Spec.Volumes = volumes
				break
			}
		}
	}

	return deployment
}

// upsertDataVolume insert additional data volume if provided and ensures that it is useable
// by memcached.
func upsertDataVolume(deployment *apps.Deployment, memcached *api.Memcached, memcachedImage string) *apps.Deployment {
	if memcached.Spec.DataVolume != nil {
		dataVolumeMount := core.VolumeMount{
			Name:      DATA_SOURCE_VOLUME,
			MountPath: DATA_SOURCE_VOLUME_MOUNTPATH,
		}

		for i, container := range deployment.Spec.Template.Spec.Containers {
			if container.Name == api.ResourceSingularMemcached {
				volumeMounts := container.VolumeMounts
				volumeMounts = core_util.UpsertVolumeMount(volumeMounts, dataVolumeMount)
				deployment.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

				dataVolume := core.Volume{
					Name:         DATA_SOURCE_VOLUME,
					VolumeSource: *memcached.Spec.DataVolume,
				}

				volumes := deployment.Spec.Template.Spec.Volumes
				volumes = core_util.UpsertVolume(volumes, dataVolume)
				deployment.Spec.Template.Spec.Volumes = volumes
				break
			}
		}

		// The volume will be created as owned by root, but
		// memcached will run as user "memcache". Changing fsGroup is broken
		// for ephemeral inline volumes (https://github.com/kubernetes/kubernetes/issues/89290)
		// and we don't know the uid of the "memcache" user, so instead of relying
		// on fsGroup we run a "chown" inside an init container which uses the same
		// image as the daemon.
		var root int64
		privileged := true
		deployment.Spec.Template.Spec.InitContainers = core_util.UpsertContainer(deployment.Spec.Template.Spec.InitContainers,
			core.Container{
				Name:            "data-volume-owner",
				Image:           memcachedImage,
				ImagePullPolicy: core.PullIfNotPresent,
				Command: []string{
					"/bin/chown",
					"memcache",
					DATA_SOURCE_VOLUME_MOUNTPATH,
				},
				SecurityContext: &core.SecurityContext{
					RunAsUser:  &root,
					Privileged: &privileged,
				},
				VolumeMounts: []core.VolumeMount{
					dataVolumeMount,
				},
			})
	}

	return deployment
}
