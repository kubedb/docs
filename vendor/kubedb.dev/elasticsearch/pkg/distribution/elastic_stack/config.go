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

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

const (
	ConfigFileName = "elasticsearch.yml"
)

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

	if secret != nil {
		// If the secret already exists,
		// check whether it contains "elasticsearch.yml" file or not.
		if value, ok := secret.Data[ConfigFileName]; !ok || len(value) == 0 {
			return errors.New("elasticsearch.yml is missing")
		}

		// If secret is owned by the elasticsearch object,
		// update the labels.
		// Labels hold information like elasticsearch version,
		// should be synced.
		ctrl := metav1.GetControllerOf(secret)
		if ctrl != nil &&
			ctrl.Kind == api.ResourceKindElasticsearch && ctrl.Name == es.db.Name {

			// sync labels
			if _, _, err := core_util.CreateOrPatchSecret(context.TODO(), es.kClient, secret.ObjectMeta, func(in *core.Secret) *core.Secret {
				in.Labels = core_util.UpsertMap(in.Labels, es.db.OffshootLabels())
				return in
			}, metav1.PatchOptions{}); err != nil {
				return err
			}
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

	if !es.db.Spec.DisableSecurity {
		config = xpack_security_enabled

		// If rest layer is secured with certs
		if es.db.Spec.EnableSSL {
			config += https_enabled
		} else {
			config += https_disabled
		}

	} else {
		config = xpack_security_disabled
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
