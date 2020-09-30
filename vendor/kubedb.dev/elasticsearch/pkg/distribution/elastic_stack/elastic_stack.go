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

package elastic_stack

import (
	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	distapi "kubedb.dev/elasticsearch/pkg/distribution/api"

	"k8s.io/client-go/kubernetes"
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

func (es *Elasticsearch) RequiredCertSecretNames() []string {
	if !es.elasticsearch.Spec.DisableSecurity {
		var sNames []string
		// transport layer is always secured with certificate
		sNames = append(sNames, es.elasticsearch.MustCertSecretName(api.ElasticsearchTransportCert))

		// If SSL is enabled for REST layer
		if es.elasticsearch.Spec.EnableSSL {
			// http server certificate
			sNames = append(sNames, es.elasticsearch.MustCertSecretName(api.ElasticsearchHTTPCert))
			// archiver certificate
			sNames = append(sNames, es.elasticsearch.MustCertSecretName(api.ElasticsearchArchiverCert))
			// metrics exporter certificate, if monitoring is enabled
			if es.elasticsearch.Spec.Monitor != nil {
				sNames = append(sNames, es.elasticsearch.MustCertSecretName(api.ElasticsearchMetricsExporterCert))
			}
		}
		return sNames
	}
	return nil
}
