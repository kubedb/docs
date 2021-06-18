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

package search_guard

import (
	"context"
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"
	"kubedb.dev/elasticsearch/pkg/lib/user"

	"github.com/pkg/errors"
	"gomodules.xyz/password-generator"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

func (es *Elasticsearch) EnsureAuthSecret() error {
	if es.db.Spec.DisableSecurity {
		return nil
	}
	if len(es.db.Spec.InternalUsers) == 0 {
		return errors.New("spec.internalUsers[] cannot be empty")
	}
	if !user.HasUser(es.db.Spec.InternalUsers, api.ElasticsearchInternalUserAdmin) {
		return errors.New("spec.internalUsers[] is missing admin user")
	}

	for username, userSpec := range es.db.Spec.InternalUsers {
		secretName := userSpec.SecretName
		if secretName == "" {
			return errors.New(fmt.Sprintf("secretName cannot be empty for user: %s", username))
		}

		// Get secret
		secret, err := es.kClient.CoreV1().Secrets(es.db.Namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil && !kerr.IsNotFound(err) {
			return errors.Wrap(err, "failed to get secret")
		}

		// If secret already exist;
		// Validate the secret, if owned, synced the labels too.
		if err == nil {
			err = es.validateAndSyncCredSecret(secret, username)
			if err != nil {
				return errors.Wrapf(err, "failed to validate or sync secret: %s/%s", es.db.Namespace, secretName)
			}
		} else {
			// Create new secret
			// Generate random complex password of length 16
			password := password.Generate(api.DefaultPasswordLength)

			// Secret Data
			var data = map[string][]byte{
				core.BasicAuthUsernameKey: []byte(username),
				core.BasicAuthPasswordKey: []byte(password),
			}

			secret = &core.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:   secretName,
					Labels: es.db.OffshootLabels(),
				},
				Type: core.SecretTypeBasicAuth,
				Data: data,
			}

			// add owner reference
			owner := metav1.NewControllerRef(es.db, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))
			core_util.EnsureOwnerReference(&secret.ObjectMeta, owner)

			_, err = es.kClient.CoreV1().Secrets(es.db.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to create secret")
			}

			// update the ES object,
			// Add admin credential secret name to Spec.AuthSecret.
			if username == string(api.ElasticsearchInternalUserAdmin) && es.db.Spec.AuthSecret == nil {
				newES, _, err := util.PatchElasticsearch(context.TODO(), es.extClient.KubedbV1alpha2(), es.db, func(in *api.Elasticsearch) *api.Elasticsearch {
					in.Spec.AuthSecret = &core.LocalObjectReference{
						Name: secretName,
					}
					return in
				}, metav1.PatchOptions{})
				if err != nil {
					return errors.Wrap(err, "failed to patch Elasticsearch")
				}

				// Note: Instead of updating the whole DB object,
				// we've just updated the spec.AuthSecret part.
				// We are making this package independent of current state of DB object,
				// it will only depend on the given (input) DB object, and make decision based on that.
				// Necessary for, KubeDB enterprise, will reconcile based on the given DB object instead of the current DB object.
				es.db.Spec.AuthSecret = newES.Spec.AuthSecret
			}
		}
	}

	return nil
}

func (es *Elasticsearch) validateAndSyncCredSecret(secret *core.Secret, userName string) error {
	// validate secret data
	data := secret.Data
	if username, exists := data[core.BasicAuthUsernameKey]; exists {
		if string(username) != userName {
			return errors.New(fmt.Sprintf("username must be: %s", userName))
		}
	} else {
		return errors.New("username is missing")
	}

	if _, exists := data[core.BasicAuthPasswordKey]; !exists {
		return errors.New("password is missing")
	}

	// If the secret is owned by the DB object, upsert the new labels,
	// Otherwise do nothing.
	if owned, _ := core_util.IsOwnedBy(secret, es.db); owned {
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
