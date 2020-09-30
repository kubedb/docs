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
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"

	"github.com/appscode/go/crypto/rand"
	passgen "gomodules.xyz/password-generator"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	core_util "kmodules.xyz/client-go/core/v1"
)

const (
	mongodbUser = "root"

	KeyForKeyFile = "key.txt"

	DatabaseSecretSuffix = "-auth"
)

func (c *Controller) ensureDatabaseSecret(mongodb *api.MongoDB) error {
	if mongodb.Spec.DatabaseSecret == nil {
		secretVolumeSource, err := c.createDatabaseSecret(mongodb)
		if err != nil {
			return err
		}

		ms, _, err := util.PatchMongoDB(context.TODO(), c.ExtClient.KubedbV1alpha1(), mongodb, func(in *api.MongoDB) *api.MongoDB {
			in.Spec.DatabaseSecret = secretVolumeSource
			return in
		}, metav1.PatchOptions{})
		if err != nil {
			return err
		}
		mongodb.Spec.DatabaseSecret = ms.Spec.DatabaseSecret
	}

	return nil
}

func (c *Controller) ensureKeyFileSecret(mongodb *api.MongoDB) error {
	if !mongodb.KeyFileRequired() {
		return nil
	}

	secretName := mongodb.Name + api.MongoDBKeyFileSecretSuffix
	if mongodb.Spec.KeyFile != nil && mongodb.Spec.KeyFile.SecretName != "" {
		secretName = mongodb.Spec.KeyFile.SecretName
	}

	secret, err := c.checkSecret(secretName, mongodb)
	if err != nil {
		return err
	}
	if secret == nil {
		randToken := rand.GenerateTokenWithLength(756)
		base64Token := base64.StdEncoding.EncodeToString([]byte(randToken))

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   secretName,
				Labels: mongodb.OffshootLabels(),
			},
			Type: core.SecretTypeOpaque,
			StringData: map[string]string{
				KeyForKeyFile: base64Token,
			},
		}
		if _, err := c.Client.CoreV1().Secrets(mongodb.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
			return err
		}
	}

	keyFile := &core.SecretVolumeSource{
		SecretName: secretName,
	}
	_, _, err = util.PatchMongoDB(context.TODO(), c.ExtClient.KubedbV1alpha1(), mongodb, func(in *api.MongoDB) *api.MongoDB {
		in.Spec.KeyFile = keyFile
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return err
	}

	mongodb.Spec.KeyFile = keyFile
	return nil
}

func (c *Controller) createDatabaseSecret(mongodb *api.MongoDB) (*core.SecretVolumeSource, error) {
	authSecretName := mongodb.Name + DatabaseSecretSuffix

	sc, err := c.checkSecret(authSecretName, mongodb)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   authSecretName,
				Labels: mongodb.OffshootLabels(),
			},
			Type: core.SecretTypeOpaque,
			StringData: map[string]string{
				core.BasicAuthUsernameKey: mongodbUser,
				core.BasicAuthPasswordKey: passgen.Generate(api.DefaultPasswordLength),
			},
		}
		if _, err := c.Client.CoreV1().Secrets(mongodb.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
			return nil, err
		}
	}
	return &core.SecretVolumeSource{
		SecretName: authSecretName,
	}, nil
}

func (c *Controller) checkSecret(secretName string, mongodb *api.MongoDB) (*core.Secret, error) {
	secret, err := c.Client.CoreV1().Secrets(mongodb.Namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if secret.Labels[api.LabelDatabaseKind] != api.ResourceKindMongoDB ||
		secret.Labels[api.LabelDatabaseName] != mongodb.Name {
		return nil, fmt.Errorf(`intended secret "%v/%v" already exists`, mongodb.Namespace, secretName)
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
