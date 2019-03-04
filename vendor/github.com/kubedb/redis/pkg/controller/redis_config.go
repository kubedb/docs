package controller

import (
	"fmt"

	"github.com/appscode/go/types"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	kutildb "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
)

const (
	RedisConfigKey          = "redis.conf"
	RedisConfigRelativePath = "redis.conf"
)

var redisConfig = `
cluster-enabled yes
cluster-config-file /data/nodes.conf
cluster-node-timeout 5000
cluster-migration-barrier 1
dir /data
appendonly yes
protected-mode no
`

func (c *Controller) ensureRedisConfig(redis *api.Redis) error {
	// TODO: Need to support if user provided config needs to merge with our default config
	if redis.Spec.ConfigSource == nil {
		// Check if configmap name exists
		if err := c.checkConfigMap(redis); err != nil {
			return err
		}

		// create configmap for redis
		configmap, vt, err := c.createConfigMap(redis)
		if err != nil {
			return errors.Wrap(err, "Failed to CreateOrPatch configmap")
		} else if vt != kutil.VerbUnchanged {
			// add configmap to redis.spec.configSource
			redis.Spec.ConfigSource = &core.VolumeSource{}
			rd, _, err := kutildb.PatchRedis(c.ExtClient.KubedbV1alpha1(), redis, func(in *api.Redis) *api.Redis {
				in.Spec.ConfigSource = &core.VolumeSource{
					ConfigMap: &core.ConfigMapVolumeSource{
						LocalObjectReference: core.LocalObjectReference{
							Name: configmap.Name,
						},
						DefaultMode: types.Int32P(511),
					},
				}
				return in
			})
			if err != nil {
				return errors.Wrap(err, "Failed to Patch redis while updating redis.spec.configSource")
			}

			redis.Spec.ConfigSource = rd.Spec.ConfigSource
			c.recorder.Eventf(
				redis,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully %s ConfigMap",
				vt,
			)
		}
	}

	return nil
}

func (c *Controller) checkConfigMap(redis *api.Redis) error {
	// ConfigMap for Redis configuration
	configmap, err := c.Client.CoreV1().ConfigMaps(redis.Namespace).Get(redis.ConfigMapName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if configmap.Labels[api.LabelDatabaseKind] != api.ResourceKindRedis ||
		configmap.Labels[api.LabelDatabaseName] != redis.Name {
		return fmt.Errorf(`intended configmap "%v" already exists`, redis.ConfigMapName())
	}

	return nil
}

func (c *Controller) createConfigMap(redis *api.Redis) (*core.ConfigMap, kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      redis.ConfigMapName(),
		Namespace: redis.Namespace,
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis)
	if rerr != nil {
		return nil, kutil.VerbUnchanged, rerr
	}

	return core_util.CreateOrPatchConfigMap(c.Client, meta, func(in *core.ConfigMap) *core.ConfigMap {
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = redis.OffshootSelectors()
		in.Annotations = redis.Spec.ServiceTemplate.Annotations

		in.Data = map[string]string{
			RedisConfigKey: redisConfig,
		}

		return in
	})
}
