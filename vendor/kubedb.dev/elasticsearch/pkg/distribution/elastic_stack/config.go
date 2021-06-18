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
	"context"
	"errors"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"github.com/blang/semver"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

const (
	ConfigFileName = "elasticsearch.yml"
)

var elasticsearch_node_roles = `
node.roles: '${NODE_ROLES}'
`

var xpack_security_enabled = `
xpack.security.enabled: true

xpack.security.transport.ssl.enabled: true
xpack.security.transport.ssl.verification_mode: certificate
xpack.security.transport.ssl.key: certs/transport/tls.key
xpack.security.transport.ssl.certificate: certs/transport/tls.crt 
xpack.security.transport.ssl.certificate_authorities: [ "certs/transport/ca.crt" ]
`

var xpack_security_disabled = `
xpack.security.enabled: false
`

var https_enabled = `
xpack.security.http.ssl.enabled: true
xpack.security.http.ssl.key:  certs/http/tls.key
xpack.security.http.ssl.certificate: certs/http/tls.crt
xpack.security.http.ssl.certificate_authorities: [ "certs/http/ca.crt" ]
`

var https_disabled = `
xpack.security.http.ssl.enabled: false
`

func (es *Elasticsearch) EnsureDefaultConfig() error {
	secret, err := es.findSecret(es.db.ConfigSecretName())
	if err != nil {
		return err
	}
	// If the secret already exists
	// and not controlled by the ES object.
	// i.e. user provided custom secret file.
	if secret != nil && !metav1.IsControlledBy(secret, es.db) {
		// check whether it contains "elasticsearch.yml" file or not.
		if value, ok := secret.Data[ConfigFileName]; !ok || len(value) == 0 {
			return errors.New("elasticsearch.yml is missing")
		}
		return nil
	}

	// config secret isn't created yet.
	// let's create it.
	owner := metav1.NewControllerRef(es.db, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))
	secretMeta := metav1.ObjectMeta{
		Name:      es.db.ConfigSecretName(),
		Namespace: es.db.Namespace,
	}

	var config string

	// For Elasticsearch version >= 7.9.x
	// The legacy node role setting (ie. node.master=true) is deprecated,
	// So, for newer versions, node roles will be managed by elasticsearch.yml file.
	dbVersion, err := semver.Parse(es.esVersion.Spec.Version)
	if err != nil {
		return err
	}
	if dbVersion.Major > 7 || (dbVersion.Major == 7 && dbVersion.Minor >= 9) {
		config += elasticsearch_node_roles
	}

	if !es.db.Spec.DisableSecurity {
		config += xpack_security_enabled

		// If rest layer is secured with certs
		if es.db.Spec.EnableSSL {
			config += https_enabled
		} else {
			config += https_disabled
		}

	} else {
		config += xpack_security_disabled
	}

	if _, _, err := core_util.CreateOrPatchSecret(context.TODO(), es.kClient, secretMeta, func(in *core.Secret) *core.Secret {
		in.Labels = core_util.UpsertMap(in.Labels, es.db.OffshootLabels())
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Data = map[string][]byte{
			ConfigFileName: []byte(config),
		}
		return in
	}, metav1.PatchOptions{}); err != nil {
		return err
	}

	return nil
}
