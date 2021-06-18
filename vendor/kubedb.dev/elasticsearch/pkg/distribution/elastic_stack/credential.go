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
	authSecret := es.db.Spec.AuthSecret
	if authSecret == nil {
		var err error
		if authSecret, err = es.createAdminCredSecret(); err != nil {
			return err
		}
		newES, _, err := util.PatchElasticsearch(context.TODO(), es.extClient.KubedbV1alpha2(), es.db, func(in *api.Elasticsearch) *api.Elasticsearch {
			in.Spec.AuthSecret = authSecret
			return in
		}, metav1.PatchOptions{})
		if err != nil {
			return err
		}
		// Note: Instead of updating the whole DB object,
		// we've just updated the spec.AuthSecret part.
		// We are making this package independent of current state of DB object,
		// it will only depend on the given (input) DB object, and make decision based on that.
		// Necessary for, KubeDB enterprise, will reconcile based on the given DB object instead of the current DB object.
		es.db.Spec.AuthSecret = newES.Spec.AuthSecret
	} else {
		// Get the secret and validate it.
		dbSecret, err := es.kClient.CoreV1().Secrets(es.db.Namespace).Get(context.TODO(), authSecret.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to get credential secret: %s/%s", es.db.Namespace, authSecret.Name))
		}

		err = es.validateAndSyncLabels(dbSecret)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to validate/sync secret: %s/%s", dbSecret.Namespace, dbSecret.Name))
		}
	}
	return nil
}

func (es *Elasticsearch) createAdminCredSecret() (*core.LocalObjectReference, error) {
	dbSecret, err := es.findSecret(es.db.DefaultUserCredSecretName(string(api.ElasticsearchInternalUserElastic)))
	if err != nil {
		return nil, err
	}

	// if a secret already exist with the given name.
	// Validate it, whether it contains the following keys:
	//	- username
	// 	- password
	// If the secret is owned by this object, sync the labels.
	if dbSecret != nil {
		err = es.validateAndSyncLabels(dbSecret)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to validate/sync secret: %s/%s", dbSecret.Namespace, dbSecret.Name))
		}
		return &core.LocalObjectReference{
			Name: dbSecret.Name,
		}, nil
	}

	// Create new secret new random password
	pass := password.Generate(api.DefaultPasswordLength)
	var data = map[string][]byte{
		core.BasicAuthUsernameKey: []byte(api.ElasticsearchInternalUserElastic),
		core.BasicAuthPasswordKey: []byte(pass),
	}

	secret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   es.db.DefaultUserCredSecretName(string(api.ElasticsearchInternalUserElastic)),
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

	return &core.LocalObjectReference{
		Name: secret.Name,
	}, nil
}

func (es *Elasticsearch) validateAndSyncLabels(secret *core.Secret) error {
	if secret == nil {
		return errors.New("secret is empty")
	}

	if value, exist := secret.Data[core.BasicAuthUsernameKey]; !exist || len(value) == 0 {
		return errors.New("username is missing")
	}

	if value, exist := secret.Data[core.BasicAuthPasswordKey]; !exist || len(value) == 0 {
		return errors.New("password is missing")
	}

	// If secret is owned by this elasticsearch object,
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
