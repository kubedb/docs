package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	mon_api "github.com/appscode/kube-mon/api"
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
)

func (c *Controller) ensureStatefulSet(redis *api.Redis) (kutil.VerbType, error) {
	if err := c.checkStatefulSet(redis); err != nil {
		return kutil.VerbUnchanged, err
	}

	// Create statefulSet for Redis database
	statefulSet, vt, err := c.createStatefulSet(redis)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// Check StatefulSet Pod status
	if vt != kutil.VerbUnchanged {
		if err := c.checkStatefulSetPodStatus(statefulSet); err != nil {
			if ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis); rerr == nil {
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
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis); rerr == nil {
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

func (c *Controller) checkStatefulSet(redis *api.Redis) error {
	// SatatefulSet for Redis database
	statefulSet, err := c.Client.AppsV1().StatefulSets(redis.Namespace).Get(redis.OffshootName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindRedis {
		return fmt.Errorf(`Intended statefulSet "%v" already exists`, redis.OffshootName())
	}

	return nil
}

func (c *Controller) createStatefulSet(redis *api.Redis) (*apps.StatefulSet, kutil.VerbType, error) {
	statefulSetMeta := metav1.ObjectMeta{
		Name:      redis.OffshootName(),
		Namespace: redis.Namespace,
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis)
	if rerr != nil {
		return nil, kutil.VerbUnchanged, rerr
	}

	return app_util.CreateOrPatchStatefulSet(c.Client, statefulSetMeta, func(in *apps.StatefulSet) *apps.StatefulSet {
		in.ObjectMeta = core_util.EnsureOwnerReference(in.ObjectMeta, ref)
		in.Labels = core_util.UpsertMap(in.Labels, redis.StatefulSetLabels())
		in.Annotations = core_util.UpsertMap(in.Annotations, redis.StatefulSetAnnotations())

		in.Spec.Replicas = types.Int32P(1)
		in.Spec.ServiceName = c.GoverningService
		in.Spec.Template.Labels = in.Labels
		in.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: in.Labels,
		}

		in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
			Name:            api.ResourceSingularRedis,
			Image:           c.docker.GetImageWithTag(redis),
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          "db",
					ContainerPort: 6379,
					Protocol:      core.ProtocolTCP,
				},
			},
			Resources: redis.Spec.Resources,
		})
		if redis.GetMonitoringVendor() == mon_api.VendorPrometheus {
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
				Name: "exporter",
				Args: append([]string{
					"export",
					fmt.Sprintf("--address=:%d", redis.Spec.Monitor.Prometheus.Port),
					fmt.Sprintf("--enable-analytics=%v", c.EnableAnalytics),
				}, c.LoggerOptions.ToFlags()...),
				Image:           c.docker.GetOperatorImageWithTag(redis),
				ImagePullPolicy: core.PullIfNotPresent,
				Ports: []core.ContainerPort{
					{
						Name:          api.PrometheusExporterPortName,
						Protocol:      core.ProtocolTCP,
						ContainerPort: redis.Spec.Monitor.Prometheus.Port,
					},
				},
			})
		}

		in = upsertDataVolume(in, redis)

		in.Spec.Template.Spec.NodeSelector = redis.Spec.NodeSelector
		in.Spec.Template.Spec.Affinity = redis.Spec.Affinity
		in.Spec.Template.Spec.SchedulerName = redis.Spec.SchedulerName
		in.Spec.Template.Spec.Tolerations = redis.Spec.Tolerations
		in.Spec.Template.Spec.ImagePullSecrets = redis.Spec.ImagePullSecrets

		in.Spec.UpdateStrategy.Type = apps.RollingUpdateStatefulSetStrategyType
		return in
	})
}

func upsertDataVolume(statefulSet *apps.StatefulSet, redis *api.Redis) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularRedis {
			volumeMount := core.VolumeMount{
				Name:      "data",
				MountPath: "/data",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			pvcSpec := redis.Spec.Storage
			if pvcSpec != nil {
				if len(pvcSpec.AccessModes) == 0 {
					pvcSpec.AccessModes = []core.PersistentVolumeAccessMode{
						core.ReadWriteOnce,
					}
					log.Infof(`Using "%v" as AccessModes in redis.Spec.Storage`, core.ReadWriteOnce)
				}

				volumeClaim := core.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: *pvcSpec,
				}
				if pvcSpec.StorageClassName != nil {
					volumeClaim.Annotations = map[string]string{
						"volume.beta.kubernetes.io/storage-class": *pvcSpec.StorageClassName,
					}
				}
				volumeClaims := statefulSet.Spec.VolumeClaimTemplates
				volumeClaims = core_util.UpsertVolumeClaim(volumeClaims, volumeClaim)
				statefulSet.Spec.VolumeClaimTemplates = volumeClaims
			} else {
				volume := core.Volume{
					Name: "data",
					VolumeSource: core.VolumeSource{
						EmptyDir: &core.EmptyDirVolumeSource{},
					},
				}
				volumes := statefulSet.Spec.Template.Spec.Volumes
				volumes = core_util.UpsertVolume(volumes, volume)
				statefulSet.Spec.Template.Spec.Volumes = volumes
				return statefulSet
			}
			break
		}
	}
	return statefulSet
}

func (c *Controller) checkStatefulSetPodStatus(statefulSet *apps.StatefulSet) error {
	return core_util.WaitUntilPodRunningBySelector(
		c.Client,
		statefulSet.Namespace,
		statefulSet.Spec.Selector,
		int(types.Int32(statefulSet.Spec.Replicas)),
	)
}
