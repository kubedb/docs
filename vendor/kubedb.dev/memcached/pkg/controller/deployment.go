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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
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

func (c *Controller) ensureStatefulSet(memcached *api.Memcached) (kutil.VerbType, error) {
	if err := c.checkStatefulSet(memcached); err != nil {
		return kutil.VerbUnchanged, err
	}

	// Create statefulSet for Memcached database
	sts, vt, err := c.createStatefulSet(memcached)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	// Check StatefulSet Pod status
	if vt != kutil.VerbUnchanged {
		if err := app_util.WaitUntilStatefulSetReady(context.TODO(), c.Client, sts.ObjectMeta); err != nil {
			return kutil.VerbUnchanged, err
		}
		c.Recorder.Eventf(
			memcached,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet",
			vt,
		)
	}

	// ensure pdb
	if err := c.CreateStatefulSetPodDisruptionBudget(sts); err != nil {
		return vt, err
	}
	return vt, nil
}

func (c *Controller) checkStatefulSet(memcached *api.Memcached) error {
	// StatefulSet for Memcached database
	sts, err := c.Client.AppsV1().StatefulSets(memcached.Namespace).Get(context.TODO(), memcached.OffshootName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}
	if sts.Labels[api.LabelDatabaseKind] != api.ResourceKindMemcached ||
		sts.Labels[api.LabelDatabaseName] != memcached.Name {
		return fmt.Errorf(`intended sts "%v/%v" already exists`, memcached.Namespace, memcached.OffshootName())
	}
	return nil
}

func (c *Controller) createStatefulSet(db *api.Memcached) (*apps.StatefulSet, kutil.VerbType, error) {
	stsMeta := metav1.ObjectMeta{
		Name:      db.OffshootName(),
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMemcached))

	memcachedVersion, err := c.DBClient.CatalogV1alpha1().MemcachedVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	return app_util.CreateOrPatchStatefulSet(context.TODO(), c.Client, stsMeta, func(in *apps.StatefulSet) *apps.StatefulSet {
		in.Labels = db.OffshootLabels()
		in.Annotations = db.Spec.PodTemplate.Controller.Annotations
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

		in.Spec.Replicas = db.Spec.Replicas
		in.Spec.Template.Labels = in.Labels
		in.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: db.OffshootSelectors(),
		}
		in.Spec.Template.Labels = db.OffshootSelectors()
		in.Spec.Template.Annotations = db.Spec.PodTemplate.Annotations
		in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(in.Spec.Template.Spec.InitContainers, db.Spec.PodTemplate.Spec.InitContainers)
		in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
			Name:            api.ResourceSingularMemcached,
			Image:           memcachedVersion.Spec.DB.Image,
			ImagePullPolicy: core.PullIfNotPresent,
			Args:            db.Spec.PodTemplate.Spec.Args,
			Ports: []core.ContainerPort{
				{
					Name:          "db",
					ContainerPort: 11211,
					Protocol:      core.ProtocolTCP,
				},
			},
			Resources:      db.Spec.PodTemplate.Spec.Resources,
			LivenessProbe:  db.Spec.PodTemplate.Spec.LivenessProbe,
			ReadinessProbe: db.Spec.PodTemplate.Spec.ReadinessProbe,
			Lifecycle:      db.Spec.PodTemplate.Spec.Lifecycle,
		})
		if db.Spec.Monitor != nil && db.Spec.Monitor.Agent.Vendor() == mona.VendorPrometheus {
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
				Name: "exporter",
				Args: append([]string{
					fmt.Sprintf("--web.listen-address=:%v", db.Spec.Monitor.Prometheus.Exporter.Port),
					fmt.Sprintf("--web.telemetry-path=%v", db.StatsService().Path()),
				}, db.Spec.Monitor.Prometheus.Exporter.Args...),
				Image:           memcachedVersion.Spec.Exporter.Image,
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
			})
		}
		in = upsertUserEnv(in, db)
		in = upsertCustomConfig(in, db)
		in = upsertDataVolume(in, db, memcachedVersion.Spec.DB.Image)

		in.Spec.Template.Spec.NodeSelector = db.Spec.PodTemplate.Spec.NodeSelector
		in.Spec.Template.Spec.Affinity = db.Spec.PodTemplate.Spec.Affinity
		if db.Spec.PodTemplate.Spec.SchedulerName != "" {
			in.Spec.Template.Spec.SchedulerName = db.Spec.PodTemplate.Spec.SchedulerName
		}
		in.Spec.Template.Spec.Tolerations = db.Spec.PodTemplate.Spec.Tolerations
		in.Spec.Template.Spec.ImagePullSecrets = db.Spec.PodTemplate.Spec.ImagePullSecrets
		in.Spec.Template.Spec.PriorityClassName = db.Spec.PodTemplate.Spec.PriorityClassName
		in.Spec.Template.Spec.Priority = db.Spec.PodTemplate.Spec.Priority
		in.Spec.Template.Spec.SecurityContext = db.Spec.PodTemplate.Spec.SecurityContext
		in.Spec.Template.Spec.ServiceAccountName = db.Spec.PodTemplate.Spec.ServiceAccountName
		in.Spec.UpdateStrategy = apps.StatefulSetUpdateStrategy{
			Type: apps.OnDeleteStatefulSetStrategyType,
		}

		return in
	}, metav1.PatchOptions{})
}

// upsertUserEnv add/overwrite env from user provided env in crd spec
func upsertUserEnv(sts *apps.StatefulSet, memcached *api.Memcached) *apps.StatefulSet {
	for i, container := range sts.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMemcached {
			sts.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, memcached.Spec.PodTemplate.Spec.Env...)
			return sts
		}
	}
	return sts
}

// upsertCustomConfig insert custom configuration volume if provided.
func upsertCustomConfig(sts *apps.StatefulSet, memcached *api.Memcached) *apps.StatefulSet {
	if memcached.Spec.ConfigSecret != nil {
		for i, container := range sts.Spec.Template.Spec.Containers {
			if container.Name == api.ResourceSingularMemcached {

				configSourceVolumeMount := core.VolumeMount{
					Name:      CONFIG_SOURCE_VOLUME,
					MountPath: CONFIG_SOURCE_VOLUME_MOUNTPATH,
				}

				volumeMounts := container.VolumeMounts
				volumeMounts = core_util.UpsertVolumeMount(volumeMounts, configSourceVolumeMount)
				sts.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

				configSourceVolume := core.Volume{
					Name: CONFIG_SOURCE_VOLUME,
					VolumeSource: core.VolumeSource{
						Secret: &core.SecretVolumeSource{
							SecretName: memcached.Spec.ConfigSecret.Name,
						},
					},
				}

				volumes := sts.Spec.Template.Spec.Volumes
				volumes = core_util.UpsertVolume(volumes, configSourceVolume)
				sts.Spec.Template.Spec.Volumes = volumes
				break
			}
		}
	}

	return sts
}

// upsertDataVolume insert additional data volume if provided and ensures that it is useable
// by memcached.
func upsertDataVolume(sts *apps.StatefulSet, memcached *api.Memcached, memcachedImage string) *apps.StatefulSet {
	if memcached.Spec.DataVolume != nil {
		dataVolumeMount := core.VolumeMount{
			Name:      DATA_SOURCE_VOLUME,
			MountPath: DATA_SOURCE_VOLUME_MOUNTPATH,
		}

		for i, container := range sts.Spec.Template.Spec.Containers {
			if container.Name == api.ResourceSingularMemcached {
				volumeMounts := container.VolumeMounts
				volumeMounts = core_util.UpsertVolumeMount(volumeMounts, dataVolumeMount)
				sts.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

				dataVolume := core.Volume{
					Name:         DATA_SOURCE_VOLUME,
					VolumeSource: *memcached.Spec.DataVolume,
				}

				volumes := sts.Spec.Template.Spec.Volumes
				volumes = core_util.UpsertVolume(volumes, dataVolume)
				sts.Spec.Template.Spec.Volumes = volumes
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
		sts.Spec.Template.Spec.InitContainers = core_util.UpsertContainer(sts.Spec.Template.Spec.InitContainers,
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

	return sts
}
