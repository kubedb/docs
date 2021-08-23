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
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	"github.com/pkg/errors"
	"gomodules.xyz/password-generator"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

func (es *Elasticsearch) EnsureAuthSecret() error {
	if es.db.Spec.DisableSecurity {
		return nil
	}

	// Cases:
	//	1. If user doesn't provide anything: Create secret with default name and random password
	//	2. If user provides secret name:
	//		- If secret exists: use this
	//		- If secret doesn't exist: create secret with the given name and random password

	// Algo:
	//	If secretName is empty:
	//		- set default secret name to `db.spec.authSecret`: <db-name>-elastic-cred
	//	end if
	//	Get the secret with the given name
	//	If secret is found:
	//		- validate: username(== elastic), password exist
	//		- If secret is owned by DB, Sync the labels
	//	else
	//		# secret not found
	//		- Create secret with the given name, random generated password
	//	end if
	//	Patch Elasticsearch `db.spec.authSecret`

	authSecret, err := es.createOrSyncAdminCredSecret()
	if err != nil {
		return err
	}
	_, _, err = util.PatchElasticsearch(context.TODO(), es.extClient.KubedbV1alpha2(), es.db, func(in *api.Elasticsearch) *api.Elasticsearch {
		in.Spec.AuthSecret = authSecret
		return in
	}, metav1.PatchOptions{})
	return err
}

func (es *Elasticsearch) createOrSyncAdminCredSecret() (*core.LocalObjectReference, error) {
	// If secret name is not provided:
	//		set default secretName
	if es.db.Spec.AuthSecret == nil || (es.db.Spec.AuthSecret != nil && len(es.db.Spec.AuthSecret.Name) == 0) {
		es.db.Spec.AuthSecret = &core.LocalObjectReference{
			Name: es.db.DefaultUserCredSecretName(string(api.ElasticsearchInternalUserElastic)),
		}
	}

	dbSecret, err := es.findSecret(es.db.Spec.AuthSecret.Name)
	if err != nil {
		return nil, err
	}

	// if a secret already exist with the given name.
	// Validate it, whether it contains the following keys:
	//	- username & (username == given username)
	// 	- password
	// If the secret is owned by this object, sync the labels.
	if dbSecret != nil {
		err = es.validateAndSyncLabels(dbSecret, string(api.ElasticsearchInternalUserElastic))
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to validate/sync secret: %s/%s", dbSecret.Namespace, dbSecret.Name))
		}
		return es.db.Spec.AuthSecret, nil
	}

	// Create new secret new random password
	pass := password.Generate(api.DefaultPasswordLength)
	var data = map[string][]byte{
		core.BasicAuthUsernameKey: []byte(api.ElasticsearchInternalUserElastic),
		core.BasicAuthPasswordKey: []byte(pass),
	}

	secret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   es.db.Spec.AuthSecret.Name,
			Labels: es.db.OffshootLabels(),
		},
		Type: core.SecretTypeBasicAuth,
		Data: data,
	}

	// add owner reference
	owner := metav1.NewControllerRef(es.db, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))
	core_util.EnsureOwnerReference(&secret.ObjectMeta, owner)

	if _, err := es.kClient.CoreV1().Secrets(es.db.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
		return nil, err
	}

	return es.db.Spec.AuthSecret, nil
}

func (es *Elasticsearch) validateAndSyncLabels(secret *core.Secret, username string) error {
	if secret == nil {
		return errors.New("secret is empty")
	}

	if value, exist := secret.Data[core.BasicAuthUsernameKey]; !exist || len(value) == 0 {
		return errors.New("username is missing")
	} else if username != "" && string(value) != username {
		return errors.Errorf("username must be %s", username)
	}

	if value, exist := secret.Data[core.BasicAuthPasswordKey]; !exist || len(value) == 0 {
		return errors.New("password is missing")
	}

	// If secret is owned by this elasticsearch object, update the labels.
	// Labels may hold important information, should be synced.
	owned, _ := core_util.IsOwnedBy(secret, es.db)
	if owned {
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
