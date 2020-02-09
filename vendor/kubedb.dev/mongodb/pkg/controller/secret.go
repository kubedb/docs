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
	"encoding/base64"
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"

	"github.com/appscode/go/crypto/rand"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	mongodbUser = "root"

	KeyMongoDBUser     = "username"
	KeyMongoDBPassword = "password"
	KeyForKeyFile      = "key.txt"

	DatabaseSecretSuffix = "-auth"
)

func (c *Controller) ensureDatabaseSecret(mongodb *api.MongoDB) error {
	if mongodb.Spec.DatabaseSecret == nil {
		secretVolumeSource, err := c.createDatabaseSecret(mongodb)
		if err != nil {
			return err
		}

		ms, _, err := util.PatchMongoDB(c.ExtClient.KubedbV1alpha1(), mongodb, func(in *api.MongoDB) *api.MongoDB {
			in.Spec.DatabaseSecret = secretVolumeSource
			return in
		})
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
		if _, err := c.Client.CoreV1().Secrets(mongodb.Namespace).Create(secret); err != nil {
			return err
		}
	}

	keyFile := &core.SecretVolumeSource{
		SecretName: secretName,
	}
	_, _, err = util.PatchMongoDB(c.ExtClient.KubedbV1alpha1(), mongodb, func(in *api.MongoDB) *api.MongoDB {
		in.Spec.KeyFile = keyFile
		return in
	})
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
		randPassword := ""

		// if the password starts with "-" it will cause error in bash scripts (in mongodb-tools)
		for randPassword = rand.GeneratePassword(); randPassword[0] == '-'; randPassword = rand.GeneratePassword() {
		}

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   authSecretName,
				Labels: mongodb.OffshootLabels(),
			},
			Type: core.SecretTypeOpaque,
			StringData: map[string]string{
				KeyMongoDBUser:     mongodbUser,
				KeyMongoDBPassword: randPassword,
			},
		}
		if _, err := c.Client.CoreV1().Secrets(mongodb.Namespace).Create(secret); err != nil {
			return nil, err
		}
	}
	return &core.SecretVolumeSource{
		SecretName: authSecretName,
	}, nil
}

func (c *Controller) checkSecret(secretName string, mongodb *api.MongoDB) (*core.Secret, error) {
	secret, err := c.Client.CoreV1().Secrets(mongodb.Namespace).Get(secretName, metav1.GetOptions{})
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

func (c *Controller) MongoDBForSecret(s *core.Secret) (*api.MongoDB, error) {
	dbs, err := c.mgLister.MongoDBs(s.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, db := range dbs {
		if metav1.IsControlledBy(s, db) {
			return db, nil
		}
	}

	return nil, nil
}
