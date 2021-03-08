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
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	kutildb "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

const (
	RedisConfigKey          = "redis.conf"
	RedisConfigRelativePath = "redis.conf"
)

/*
example config:

dir /data								always default config
appendonly yes									.
protected-mode no								.

cluster-enabled yes						if redis in cluster mode
cluster-config-file /data/nodes.conf			  .
cluster-node-timeout 5000						  .
cluster-migration-barrier 1						  .

database 10								data from initial configsecret
maxclient 500										.
.													.
.													.
*/
var redisConfig = `dir /data
appendonly yes
protected-mode no
`
var clusterConfig = `cluster-enabled yes
cluster-config-file /data/nodes.conf
cluster-node-timeout 5000
cluster-migration-barrier 1
`

func (c *Controller) ensureRedisConfig(db *api.Redis) error {
	// TODO: Need to support if user provided config needs to merge with our default config
	if db.Spec.ConfigSecret == nil {
		// Check if secret name exists
		if err := c.checkSecret(db); err != nil {
			return err
		}

		// create secret for redis
		configSecret, vt, err := CreateSecret(c.Client, db, db.ConfigSecretName(), "")
		if err != nil {
			return errors.Wrap(err, "Failed to CreateOrPatch secret")
		} else if vt != kutil.VerbUnchanged {
			// add secret to db.spec.configSource
			rd, _, err := kutildb.PatchRedis(context.TODO(), c.DBClient.KubedbV1alpha2(), db, func(in *api.Redis) *api.Redis {
				in.Spec.ConfigSecret = &core.LocalObjectReference{
					Name: configSecret.Name,
				}
				return in
			}, metav1.PatchOptions{})
			if err != nil {
				return errors.Wrap(err, "Failed to Patch redis while updating db.spec.configSource")
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
	} else {
		// Check if secret name exists
		data, err := c.checkUsedSecret(db)
		if err != nil {
			return err
		} else if data == "" {
			return nil
		}

		// create secret for redis
		_, _, err = CreateSecret(c.Client, db, db.Spec.ConfigSecret.Name, data)
		if err != nil {
			return errors.Wrap(err, "Failed to CreateOrPatch secret")
		}
	}
	return nil
}

func (c *Controller) checkSecret(db *api.Redis) error {
	// Secret for Redis configuration
	configSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.ConfigSecretName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if configSecret.Labels[meta_util.NameLabelKey] != db.ResourceFQN() ||
		configSecret.Labels[meta_util.InstanceLabelKey] != db.Name {
		return fmt.Errorf(`intended secret "%v" already exists`, db.ConfigSecretName())
	}

	return nil
}

func (c *Controller) checkUsedSecret(db *api.Redis) (string, error) {
	// Secret for Redis configuration
	configSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.Spec.ConfigSecret.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if _, found := configSecret.Data[RedisConfigKey]; !found {
		return "", fmt.Errorf(`your secret should have the key "db.conf"`)
	}
	if _, found := configSecret.Labels[meta_util.NameLabelKey]; !found {
		return string(configSecret.Data[RedisConfigKey]), err
	}

	return "", nil
}

func CreateSecret(client kubernetes.Interface, db *api.Redis, name, data string) (*core.Secret, kutil.VerbType, error) {
	dataArray := []string{redisConfig}
	if db.Spec.Mode == api.RedisModeCluster {
		dataArray = append(dataArray, clusterConfig)
	}
	dataArray = append(dataArray, data)
	finalConfig := strings.Join(dataArray, "\n")

	meta := metav1.ObjectMeta{
		Name:      name,
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindRedis))

	return core_util.CreateOrPatchSecret(context.TODO(), client, meta, func(in *core.Secret) *core.Secret {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = db.OffshootSelectors()

		in.StringData = map[string]string{
			RedisConfigKey: finalConfig,
		}

		return in
	}, metav1.PatchOptions{})
}
