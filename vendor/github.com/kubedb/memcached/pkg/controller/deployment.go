package controller

import (
	"fmt"

	"github.com/appscode/kutil"
	app_util "github.com/appscode/kutil/apps/v1"
	core_util "github.com/appscode/kutil/core/v1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/eventer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
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
			if ref, rerr := reference.GetReference(clientsetscheme.Scheme, memcached); rerr == nil {
				c.recorder.Eventf(
					ref,
					core.EventTypeWarning,
					eventer.EventReasonFailedToStart,
					`Failed to CreateOrPatch StatefulSet. Reason: %v`,
					err,
				)
			}
			return kutil.VerbUnchanged, err
		}
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, memcached); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully %v StatefulSet",
				vt,
			)
		}
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
		return fmt.Errorf(`intended deployment "%v" already exists`, memcached.OffshootName())
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

	memcachedVersion, err := c.ExtClient.MemcachedVersions().Get(string(memcached.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	return app_util.CreateOrPatchDeployment(c.Client, deploymentMeta, func(in *apps.Deployment) *apps.Deployment {
		in.Labels = memcached.OffshootLabels()
		in.Annotations = memcached.Spec.PodTemplate.Controller.Annotations
		in.ObjectMeta = core_util.EnsureOwnerReference(in.ObjectMeta, ref)

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
			Resources: memcached.Spec.PodTemplate.Spec.Resources,
		})
		if memcached.GetMonitoringVendor() == mona.VendorPrometheus {
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
				Name: "exporter",
				Args: append([]string{
					"export",
					fmt.Sprintf("--address=:%d", memcached.Spec.Monitor.Prometheus.Port),
					fmt.Sprintf("--enable-analytics=%v", c.EnableAnalytics),
				}, c.LoggerOptions.ToFlags()...),
				Image:           memcachedVersion.Spec.Exporter.Image,
				ImagePullPolicy: core.PullIfNotPresent,
				Ports: []core.ContainerPort{
					{
						Name:          api.PrometheusExporterPortName,
						Protocol:      core.ProtocolTCP,
						ContainerPort: memcached.Spec.Monitor.Prometheus.Port,
					},
				},
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
