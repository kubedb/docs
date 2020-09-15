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

package open_distro

import (
	"context"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	distapi "kubedb.dev/elasticsearch/pkg/distribution/api"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	api_util "kmodules.xyz/client-go/api/v1"
)

type Elasticsearch struct {
	kClient       kubernetes.Interface
	extClient     cs.Interface
	elasticsearch *api.Elasticsearch
	esVersion     *catalog.ElasticsearchVersion
}

var _ distapi.ElasticsearchInterface = &Elasticsearch{}

func New(kc kubernetes.Interface, extClient cs.Interface, es *api.Elasticsearch, esVersion *catalog.ElasticsearchVersion) *Elasticsearch {
	return &Elasticsearch{
		kClient:       kc,
		extClient:     extClient,
		elasticsearch: es,
		esVersion:     esVersion,
	}
}

func (es *Elasticsearch) UpdatedElasticsearch() *api.Elasticsearch {
	return es.elasticsearch
}

func (es *Elasticsearch) IsAllRequiredSecretAvailable() bool {
	if !es.elasticsearch.Spec.DisableSecurity {
		tls := es.elasticsearch.Spec.TLS

		// check transport layer cert
		sName, exist := api_util.GetCertificateSecretName(tls.Certificates, string(api.ElasticsearchTransportCert))
		if exist {
			_, err := es.getSecret(sName, es.elasticsearch.Namespace)
			if err != nil {
				return false
			}
		} else {
			return false
		}

		if es.elasticsearch.Spec.EnableSSL {
			// check http layer cert
			sName, exist := api_util.GetCertificateSecretName(tls.Certificates, string(api.ElasticsearchHTTPCert))
			if exist {
				_, err := es.getSecret(sName, es.elasticsearch.Namespace)
				if err != nil {
					return false
				}
			} else {
				return false
			}

			// check admin cert
			sName, exist = api_util.GetCertificateSecretName(tls.Certificates, string(api.ElasticsearchAdminCert))
			if exist {
				_, err := es.getSecret(sName, es.elasticsearch.Namespace)
				if err != nil {
					return false
				}
			} else {
				return false
			}

		}

		// check user credentials secret
		// admin credentials
		_, err := es.getSecret(es.elasticsearch.Spec.DatabaseSecret.SecretName, es.elasticsearch.Namespace)
		if err != nil {
			return false
		}

		// other credentials secrets
		userList := es.elasticsearch.Spec.InternalUsers
		for username := range userList {
			if username == string(api.ElasticsearchInternalUserAdmin) {
				continue
			}

			_, err := es.getSecret(es.elasticsearch.UserCredSecretName(username), es.elasticsearch.Namespace)
			if err != nil {
				return false
			}
		}

	}

	return true
}

func (es *Elasticsearch) getSecret(name, namespace string) (*corev1.Secret, error) {
	secret, err := es.kClient.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	return secret, err
}
