/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

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

	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

const (
	ConfigFileName          = "elasticsearch.yml"
	ConfigFileMountPathSG   = "/elasticsearch/config"
	ConfigFileMountPath     = "/usr/share/elasticsearch/config"
	TempConfigFileMountPath = "/elasticsearch/temp-config"
	DatabaseConfigMapSuffix = `config`
)

var xpack_config = `
xpack.security.enabled: true

xpack.security.transport.ssl.enabled: true
xpack.security.transport.ssl.verification_mode: certificate
xpack.security.transport.ssl.keystore.path: /usr/share/elasticsearch/config/certs/node.jks
xpack.security.transport.ssl.keystore.password: ${KEY_PASS}
xpack.security.transport.ssl.truststore.path: /usr/share/elasticsearch/config/certs/root.jks
xpack.security.transport.ssl.truststore.password: ${KEY_PASS}

xpack.security.http.ssl.keystore.path: /usr/share/elasticsearch/config/certs/client.jks
xpack.security.http.ssl.keystore.password: ${KEY_PASS}
xpack.security.http.ssl.truststore.path: /usr/share/elasticsearch/config/certs/root.jks
xpack.security.http.ssl.truststore.password: ${KEY_PASS}
`

func (c *Controller) ensureDatabaseConfigForXPack(elasticsearch *api.Elasticsearch) error {
	esVersion, err := c.esVersionLister.Get(string(elasticsearch.Spec.Version))
	if err != nil {
		return err
	}
	if esVersion.Spec.AuthPlugin != v1alpha1.ElasticsearchAuthPluginXpack {
		return nil
	}
	if !elasticsearch.Spec.DisableSecurity {
		if err := c.findDatabaseConfig(elasticsearch); err != nil {
			return err
		}

		cmMeta := metav1.ObjectMeta{
			Name:      fmt.Sprintf("%v-%v", elasticsearch.OffshootName(), DatabaseConfigMapSuffix),
			Namespace: elasticsearch.Namespace,
		}
		owner := metav1.NewControllerRef(elasticsearch, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))

		if _, _, err := core_util.CreateOrPatchConfigMap(context.TODO(), c.Client, cmMeta, func(in *core.ConfigMap) *core.ConfigMap {
			in.Labels = core_util.UpsertMap(in.Labels, elasticsearch.OffshootLabels())
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
			in.Data = map[string]string{
				ConfigFileName: xpack_config,
			}
			return in
		}, metav1.PatchOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) findDatabaseConfig(elasticsearch *api.Elasticsearch) error {
	cmName := fmt.Sprintf("%v-%v", elasticsearch.OffshootName(), DatabaseConfigMapSuffix)

	configMap, err := c.Client.CoreV1().ConfigMaps(elasticsearch.Namespace).Get(context.TODO(), cmName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		} else {
			return err
		}
	}

	if configMap.Labels[api.LabelDatabaseKind] != api.ResourceKindElasticsearch &&
		configMap.Labels[api.LabelDatabaseName] != elasticsearch.Name {
		return fmt.Errorf(`intended configMap "%v/%v" already exists`, elasticsearch.Namespace, cmName)
	}

	return nil
}
