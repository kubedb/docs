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
	cs "kubedb.dev/apimachinery/client/clientset/versioned"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
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
port 6379
`
var clusterConfig = `cluster-enabled yes
cluster-config-file /data/nodes.conf
cluster-node-timeout 5000
cluster-migration-barrier 1
`
var redisSentinelConfig = `replica-announce-port 6379
`

func (c *Controller) ensureRedisConfig(db *api.Redis) error {
	var err error
	data := ""
	if db.Spec.ConfigSecret == nil {
		// Check if secret name exists
		if err := c.checkSecret(db); err != nil {
			return err
		}
	} else {
		// Check if secret name exists
		data, err = c.getCustomSecretData(db)
		if err != nil {
			return err
		}
	}
	// create secret for redis
	_, _, err = CreateConfigSecret(c.Client, c.DBClient, db, db.ConfigSecretName(), data)
	if err != nil {
		return errors.Wrap(err, "Failed to CreateOrPatch secret")
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

func (c *Controller) getCustomSecretData(db *api.Redis) (string, error) {
	// Secret for Redis configuration
	configSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.Spec.ConfigSecret.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(configSecret.Data[api.RedisConfigKey]), err
}

// If Custom Config Data is empty then start create or patch the secret with default config (default.conf)
// If custom config not empty then create with two keys:
//		1. default config (default.conf)
// 		2. for custom config (redis.conf)
func CreateConfigSecret(kubeClient kubernetes.Interface, dbClient cs.Interface, db *api.Redis, name, CustomConfigData string) (*core.Secret, kutil.VerbType, error) {
	dataArray := []string{redisConfig}
	if db.Spec.Mode == api.RedisModeCluster {
		dataArray = append(dataArray, clusterConfig)
	} else if db.Spec.Mode == api.RedisModeSentinel {
		dataArray = append(dataArray, redisSentinelConfig)
	}
	passwordConfigData, err := AddAuthParamsInConfigSecret(kubeClient, dbClient, db)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	dataArray = append(dataArray, passwordConfigData)
	if CustomConfigData != "" {
		dataArray = append(dataArray, fmt.Sprintf("include %s%s\n", CONFIG_MOUNT_PATH, api.RedisConfigKey))
	}

	finalDefaultConfig := strings.Join(dataArray, "\n")

	inputData := map[string][]byte{
		api.DefaultConfigKey: []byte(finalDefaultConfig),
	}

	if CustomConfigData != "" {
		inputData[api.RedisConfigKey] = []byte(CustomConfigData)
	}

	meta := metav1.ObjectMeta{
		Name:      name,
		Namespace: db.Namespace,
	}

	configSecret, err := kubeClient.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.ConfigSecretName(), metav1.GetOptions{})
	if err != nil && !kerr.IsNotFound(err) {
		return nil, kutil.VerbUnchanged, err
	}
	if err != nil && kerr.IsNotFound(err) {
		owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindRedis))
		return core_util.CreateOrPatchSecret(context.TODO(), kubeClient, meta, func(in *core.Secret) *core.Secret {
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
			in.Labels = db.OffshootSelectors()
			in.Data = inputData

			return in
		}, metav1.PatchOptions{})
	} else {
		return core_util.PatchSecret(context.TODO(), kubeClient, configSecret, func(in *core.Secret) *core.Secret {
			in.Data = inputData
			return in
		}, metav1.PatchOptions{})
	}

}
func AddAuthParamsInConfigSecret(kubeClient kubernetes.Interface, dbClient cs.Interface, db *api.Redis) (finalData string, err error) {
	redisVersion, err := dbClient.CatalogV1alpha1().RedisVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	curVersion, err := semver.NewVersion(redisVersion.Spec.Version)
	if err != nil {
		return "", fmt.Errorf("can't get the version from RedisVersion spec")
	}
	if curVersion.Major() > 4 {
		authSecret, err := kubeClient.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.Spec.AuthSecret.Name, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		var dataArray []string
		dataArray = append(dataArray, fmt.Sprintf("requirepass %s\n", string(authSecret.Data[core.BasicAuthPasswordKey])))
		dataArray = append(dataArray, fmt.Sprintf("masterauth %s\n", string(authSecret.Data[core.BasicAuthPasswordKey])))
		finalData = strings.Join(dataArray, "\n")
	}

	return finalData, nil
}
