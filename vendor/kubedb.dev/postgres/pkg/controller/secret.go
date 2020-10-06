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
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"

	passgen "gomodules.xyz/password-generator"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

const (
	PostgresUser     = "POSTGRES_USER"
	PostgresPassword = "POSTGRES_PASSWORD"
)

func (c *Controller) ensureDatabaseSecret(postgres *api.Postgres) error {
	databaseSecretVolume := postgres.Spec.DatabaseSecret
	if databaseSecretVolume == nil {
		var err error
		if databaseSecretVolume, err = c.createDatabaseSecret(postgres); err != nil {
			return err
		}
		pg, _, err := util.PatchPostgres(context.TODO(), c.DBClient.KubedbV1alpha1(), postgres, func(in *api.Postgres) *api.Postgres {
			in.Spec.DatabaseSecret = databaseSecretVolume
			return in
		}, metav1.PatchOptions{})
		if err != nil {
			return err
		}
		postgres.Spec.DatabaseSecret = pg.Spec.DatabaseSecret
		return nil
	}
	return c.upgradeDatabaseSecret(postgres)
}

func (c *Controller) findDatabaseSecret(postgres *api.Postgres) (*core.Secret, error) {
	name := postgres.OffshootName() + "-auth"

	secret, err := c.Client.CoreV1().Secrets(postgres.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	if secret.Labels[api.LabelDatabaseKind] != api.ResourceKindPostgres ||
		secret.Labels[api.LabelDatabaseName] != postgres.Name {
		return nil, fmt.Errorf(`intended secret "%v/%v" already exists`, postgres.Namespace, name)
	}

	return secret, nil
}

func (c *Controller) createDatabaseSecret(postgres *api.Postgres) (*core.SecretVolumeSource, error) {
	databaseSecret, err := c.findDatabaseSecret(postgres)
	if err != nil {
		return nil, err
	}
	if databaseSecret != nil {
		return &core.SecretVolumeSource{
			SecretName: databaseSecret.Name,
		}, nil
	}

	name := fmt.Sprintf("%v-auth", postgres.OffshootName())
	secret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: postgres.OffshootLabels(),
		},
		Type: core.SecretTypeOpaque,
		Data: map[string][]byte{
			PostgresUser:     []byte("postgres"),
			PostgresPassword: []byte(passgen.Generate(api.DefaultPasswordLength)),
		},
	}
	if _, err := c.Client.CoreV1().Secrets(postgres.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
		return nil, err
	}

	return &core.SecretVolumeSource{
		SecretName: secret.Name,
	}, nil
}

// This is done to fix 0.8.0 -> 0.9.0 upgrade due to
// https://github.com/kubedb/postgres/pull/179/files#diff-10ddaf307bbebafda149db10a28b9c24R20 commit
func (c *Controller) upgradeDatabaseSecret(postgres *api.Postgres) error {
	meta := metav1.ObjectMeta{
		Name:      postgres.Spec.DatabaseSecret.SecretName,
		Namespace: postgres.Namespace,
	}

	_, _, err := core_util.CreateOrPatchSecret(context.TODO(), c.Client, meta, func(in *core.Secret) *core.Secret {
		if _, ok := in.Data[PostgresUser]; !ok {
			in.StringData = map[string]string{PostgresUser: "postgres"}
		}
		return in
	}, metav1.PatchOptions{})
	return err
}
