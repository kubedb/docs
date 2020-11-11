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
	kutildb "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (c *Controller) ensureRedisConfig(db *api.Redis) error {
	// TODO: Need to support if user provided config needs to merge with our default config
	if db.Spec.ConfigSecret == nil {
		// Check if secret name exists
		if err := c.checkSecret(db); err != nil {
			return err
		}

		// create configmap for redis
		configSecret, vt, err := c.createSecret(db)
		if err != nil {
			return errors.Wrap(err, "Failed to CreateOrPatch configmap")
		} else if vt != kutil.VerbUnchanged {
			// add configmap to redis.spec.configSource
			rd, _, err := kutildb.PatchRedis(context.TODO(), c.DBClient.KubedbV1alpha2(), db, func(in *api.Redis) *api.Redis {
				in.Spec.ConfigSecret = &core.LocalObjectReference{
					Name: configSecret.Name,
				}
				return in
			}, metav1.PatchOptions{})
			if err != nil {
				return errors.Wrap(err, "Failed to Patch redis while updating redis.spec.configSource")
			}

			db.Spec.ConfigSecret = rd.Spec.ConfigSecret
			c.Recorder.Eventf(
				db,
				core.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully %s Secret",
				vt,
			)
		}
	}

	return nil
}

func (c *Controller) checkSecret(db *api.Redis) error {
	// Secret for Redis configuration
	configmap, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.ConfigSecretName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if configmap.Labels[api.LabelDatabaseKind] != api.ResourceKindRedis ||
		configmap.Labels[api.LabelDatabaseName] != db.Name {
		return fmt.Errorf(`intended configmap "%v" already exists`, db.ConfigSecretName())
	}

	return nil
}

func (c *Controller) createSecret(db *api.Redis) (*core.Secret, kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      db.ConfigSecretName(),
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindRedis))

	return core_util.CreateOrPatchSecret(context.TODO(), c.Client, meta, func(in *core.Secret) *core.Secret {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = db.OffshootSelectors()

		in.StringData = map[string]string{
			RedisConfigKey: redisConfig,
		}

		return in
	}, metav1.PatchOptions{})
}
