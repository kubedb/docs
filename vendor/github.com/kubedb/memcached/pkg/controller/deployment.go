package controller

import (
	"fmt"

	"github.com/appscode/go/types"
	mon_api "github.com/appscode/kube-mon/api"
	"github.com/appscode/kutil"
	app_util "github.com/appscode/kutil/apps/v1"
	core_util "github.com/appscode/kutil/core/v1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) ensureDeployment(memcached *api.Memcached) (kutil.VerbType, error) {
	if err := c.checkDeployment(memcached); err != nil {
		return kutil.VerbUnchanged, err
	}

	deploymentMeta := metav1.ObjectMeta{
		Name:      memcached.OffshootName(),
		Namespace: memcached.Namespace,
	}

	replicas := int32(1)
	if memcached.Spec.Replicas != nil {
		replicas = types.Int32(memcached.Spec.Replicas)
	}

	_, vt, err := app_util.CreateOrPatchDeployment(c.Client, deploymentMeta, func(in *apps.Deployment) *apps.Deployment {
		in.Labels = core_util.UpsertMap(in.Labels, memcached.DeploymentLabels())
		in.Annotations = core_util.UpsertMap(in.Annotations, memcached.DeploymentAnnotations())

		in.Spec.Replicas = types.Int32P(replicas)
		in.Spec.Template.Labels = in.Labels
		in.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: in.Labels,
		}

		in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
			Name:            api.ResourceNameMemcached,
			Image:           c.opt.Docker.GetImageWithTag(memcached),
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          "db",
					ContainerPort: 11211,
					Protocol:      core.ProtocolTCP,
				},
			},
			Resources: memcached.Spec.Resources,
		})
		if memcached.GetMonitoringVendor() == mon_api.VendorPrometheus {
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
				Name: "exporter",
				Args: append([]string{
					"export",
					fmt.Sprintf("--address=:%d", memcached.Spec.Monitor.Prometheus.Port),
					fmt.Sprintf("--analytics=%v", c.opt.EnableAnalytics),
				}, c.opt.LoggerOptions.ToFlags()...),
				Image:           c.opt.Docker.GetOperatorImageWithTag(memcached),
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

		in.Spec.Template.Spec.NodeSelector = memcached.Spec.NodeSelector
		in.Spec.Template.Spec.Affinity = memcached.Spec.Affinity
		in.Spec.Template.Spec.SchedulerName = memcached.Spec.SchedulerName
		in.Spec.Template.Spec.Tolerations = memcached.Spec.Tolerations
		in.Spec.Template.Spec.ImagePullSecrets = memcached.Spec.ImagePullSecrets

		return in
	})

	if err != nil {
		return kutil.VerbUnchanged, err
	}
	// Check Deployment Pod status
	if vt != kutil.VerbUnchanged {
		if err := app_util.WaitUntilDeploymentReady(c.Client, deploymentMeta); err != nil {
			c.recorder.Eventf(
				memcached.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToStart,
				`Failed to CreateOrPatch Deployment. Reason: %v`,
				err,
			)
			return kutil.VerbUnchanged, err
		}
		c.recorder.Eventf(
			memcached.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v Deployment",
			vt,
		)
		mg, _, err := util.PatchMemcached(c.ExtClient, memcached, func(in *api.Memcached) *api.Memcached {
			in.Status.Phase = api.DatabasePhaseRunning
			return in
		})
		if err != nil {
			c.recorder.Eventf(
				memcached,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return kutil.VerbUnchanged, err
		}
		memcached.Status = mg.Status
	}
	return vt, nil
}

func (c *Controller) checkDeployment(memcached *api.Memcached) error {
	// Deployment for Memcached database
	dbName := memcached.OffshootName()
	deployment, err := c.Client.AppsV1().Deployments(memcached.Namespace).Get(dbName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		} else {
			return err
		}
	}
	if deployment.Labels[api.LabelDatabaseKind] != api.ResourceKindMemcached || deployment.Labels[api.LabelDatabaseName] != dbName {
		return fmt.Errorf(`intended deployment "%v" already exists`, dbName)
	}
	return nil
}
