package controller

import (
	"fmt"

	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/eventer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

const (
	CONFIG_SOURCE_VOLUME           = "custom-config"
	CONFIG_SOURCE_VOLUME_MOUNTPATH = "/usr/config/"
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
		if err := app_util.WaitUntilDeploymentReady(c.Client, deployment.ObjectMeta); err != nil {
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
	return vt, nil
}

func (c *Controller) checkDeployment(memcached *api.Memcached) error {
	// Deployment for Memcached database
	deployment, err := c.Client.AppsV1().Deployments(memcached.Namespace).Get(memcached.OffshootName(), metav1.GetOptions{})
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

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, memcached)
	if rerr != nil {
		return nil, kutil.VerbUnchanged, rerr
	}

	memcachedVersion, err := c.ExtClient.CatalogV1alpha1().MemcachedVersions().Get(string(memcached.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	return app_util.CreateOrPatchDeployment(c.Client, deploymentMeta, func(in *apps.Deployment) *apps.Deployment {
		in.Labels = memcached.OffshootLabels()
		in.Annotations = memcached.Spec.PodTemplate.Controller.Annotations
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)

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
					fmt.Sprintf("--web.listen-address=:%v", memcached.Spec.Monitor.Prometheus.Port),
					fmt.Sprintf("--web.telemetry-path=%v", memcached.StatsService().Path()),
				}, memcached.Spec.Monitor.Args...),
				Image:           memcachedVersion.Spec.Exporter.Image,
				ImagePullPolicy: core.PullIfNotPresent,
				Ports: []core.ContainerPort{
					{
						Name:          api.PrometheusExporterPortName,
						Protocol:      core.ProtocolTCP,
						ContainerPort: memcached.Spec.Monitor.Prometheus.Port,
					},
				},
				Env:             memcached.Spec.Monitor.Env,
				Resources:       memcached.Spec.Monitor.Resources,
				SecurityContext: memcached.Spec.Monitor.SecurityContext,
			})
		}
		in = upsertUserEnv(in, memcached)
		in = upsertCustomConfig(in, memcached)

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

		if c.EnableRBAC {
			in.Spec.Template.Spec.ServiceAccountName = memcached.OffshootName()
		}

		in.Spec.Strategy = memcached.Spec.UpdateStrategy

		return in
	})
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
