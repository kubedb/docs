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
	"encoding/base64"
	"fmt"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	passgen "gomodules.xyz/password-generator"
	"gomodules.xyz/x/crypto/rand"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

func (c *Reconciler) ensureAuthSecret(db *api.MongoDB) error {
	if db.Spec.AuthSecret == nil {
		authSecret, err := c.createAuthSecret(db)
		if err != nil {
			return err
		}

		ms, _, err := util.PatchMongoDB(context.TODO(), c.DBClient.KubedbV1alpha2(), db, func(in *api.MongoDB) *api.MongoDB {
			in.Spec.AuthSecret = authSecret
			return in
		}, metav1.PatchOptions{})
		if err != nil {
			return err
		}
		db.Spec.AuthSecret = ms.Spec.AuthSecret
	}

	return nil
}

func (c *Reconciler) ensureKeyFileSecret(db *api.MongoDB) error {
	if !db.KeyFileRequired() {
		return nil
	}

	secretName := db.Name + api.MongoDBKeyFileSecretSuffix
	if db.Spec.KeyFileSecret != nil && db.Spec.KeyFileSecret.Name != "" {
		secretName = db.Spec.KeyFileSecret.Name
	}

	secret, err := c.checkSecret(secretName, db)
	if err != nil {
		return err
	}
	if secret == nil {
		randToken := rand.GenerateTokenWithLength(756)
		base64Token := base64.StdEncoding.EncodeToString([]byte(randToken))

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   secretName,
				Labels: db.OffshootLabels(),
			},
			Type: core.SecretTypeOpaque,
			StringData: map[string]string{
				api.MongoDBKeyForKeyFile: base64Token,
			},
		}
		if _, err := c.Client.CoreV1().Secrets(db.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
			return err
		}
	}

	keyFile := &core.LocalObjectReference{
		Name: secretName,
	}
	_, _, err = util.PatchMongoDB(context.TODO(), c.DBClient.KubedbV1alpha2(), db, func(in *api.MongoDB) *api.MongoDB {
		in.Spec.KeyFileSecret = keyFile
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return err
	}

	db.Spec.KeyFileSecret = keyFile
	return nil
}

func (c *Reconciler) createAuthSecret(db *api.MongoDB) (*core.LocalObjectReference, error) {
	authSecretName := db.Name + api.MongoDBAuthSecretSuffix

	sc, err := c.checkSecret(authSecretName, db)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   authSecretName,
				Labels: db.OffshootLabels(),
			},
			Type: core.SecretTypeOpaque,
			StringData: map[string]string{
				core.BasicAuthUsernameKey: api.MongodbUser,
				core.BasicAuthPasswordKey: passgen.Generate(api.DefaultPasswordLength),
			},
		}
		if _, err := c.Client.CoreV1().Secrets(db.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
			return nil, err
		}
	}
	return &core.LocalObjectReference{
		Name: authSecretName,
	}, nil
}

func (c *Reconciler) checkSecret(secretName string, db *api.MongoDB) (*core.Secret, error) {
	secret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if secret.Labels[meta_util.NameLabelKey] != db.ResourceFQN() ||
		secret.Labels[meta_util.InstanceLabelKey] != db.Name {
		return nil, fmt.Errorf(`intended secret "%v/%v" already exists`, db.Namespace, secretName)
	}
	return secret, nil
}

func (c *Controller) MongoDBForSecret(s *core.Secret) cache.ExplicitKey {
	ctrl := metav1.GetControllerOf(s)
	ok, err := core_util.IsOwnerOfGroupKind(ctrl, kubedb.GroupName, api.ResourceKindMongoDB)
	if err != nil || !ok {
		return ""
	}
	// Owner ref is set by the enterprise operator
	return cache.ExplicitKey(s.Namespace + "/" + ctrl.Name)
}

// Ensure keyfile for cluster
func (c *Reconciler) EnsureKeyFileSecret(db *api.MongoDB) error {
	sslMode := db.Spec.SSLMode
	if (sslMode != api.SSLModeDisabled && sslMode != "") ||
		db.Spec.ReplicaSet != nil || db.Spec.ShardTopology != nil {
		if err := c.ensureKeyFileSecret(db); err != nil {
			return err
		}
	}

	return nil
}
