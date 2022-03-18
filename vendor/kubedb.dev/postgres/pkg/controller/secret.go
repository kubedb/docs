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

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	passgen "gomodules.xyz/password-generator"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

func (c *Reconciler) ensureAuthSecret(db *api.Postgres) error {
	authSecret := db.Spec.AuthSecret
	if authSecret == nil {
		var err error
		if authSecret, err = c.createAuthSecret(db); err != nil {
			return err
		}
		pg, _, err := util.PatchPostgres(context.TODO(), c.DBClient.KubedbV1alpha2(), db, func(in *api.Postgres) *api.Postgres {
			in.Spec.AuthSecret = authSecret
			return in
		}, metav1.PatchOptions{})
		if err != nil {
			return err
		}
		db.Spec.AuthSecret = pg.Spec.AuthSecret
		return nil
	}
	return c.upgradeAuthSecret(db)
}

func (c *Reconciler) findAuthSecret(db *api.Postgres) (*core.Secret, error) {
	name := db.OffshootName() + "-auth"

	secret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	if secret.Labels[meta_util.NameLabelKey] != db.ResourceFQN() ||
		secret.Labels[meta_util.InstanceLabelKey] != db.Name {
		return nil, fmt.Errorf(`intended secret "%v/%v" already exists`, db.Namespace, name)
	}

	return secret, nil
}

func (c *Reconciler) createAuthSecret(db *api.Postgres) (*core.LocalObjectReference, error) {
	databaseSecret, err := c.findAuthSecret(db)
	if err != nil {
		return nil, err
	}
	if databaseSecret != nil {
		return &core.LocalObjectReference{
			Name: databaseSecret.Name,
		}, nil
	}

	name := fmt.Sprintf("%v-auth", db.OffshootName())
	secret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: db.OffshootLabels(),
		},
		Type: core.SecretTypeBasicAuth,
		Data: map[string][]byte{
			core.BasicAuthUsernameKey: []byte("postgres"),
			core.BasicAuthPasswordKey: []byte(passgen.Generate(api.DefaultPasswordLength)),
		},
	}
	if _, err := c.Client.CoreV1().Secrets(db.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
		return nil, err
	}

	return &core.LocalObjectReference{
		Name: secret.Name,
	}, nil
}

// This is done to fix 0.8.0 -> 0.9.0 upgrade due to
// https://github.com/kubedb/postgres/pull/179/files#diff-10ddaf307bbebafda149db10a28b9c24R20 commit
func (c *Reconciler) upgradeAuthSecret(db *api.Postgres) error {
	meta := metav1.ObjectMeta{
		Name:      db.Spec.AuthSecret.Name,
		Namespace: db.Namespace,
	}

	_, _, err := core_util.CreateOrPatchSecret(context.TODO(), c.Client, meta, func(in *core.Secret) *core.Secret {
		if _, ok := in.Data[core.BasicAuthUsernameKey]; !ok {
			if in.Data == nil {
				in.Data = map[string][]byte{}
			}
			if _, ok := in.Data[api.EnvPostgresUser]; ok {
				in.Data[core.BasicAuthUsernameKey] = in.Data[api.EnvPostgresUser]
			} else {
				in.Data[core.BasicAuthUsernameKey] = []byte("postgres")
			}
		}
		if _, ok := in.Data[core.BasicAuthPasswordKey]; !ok {
			if _, ok := in.Data[api.EnvPostgresPassword]; ok {
				if in.Data == nil {
					in.Data = map[string][]byte{}
				}
				in.Data[core.BasicAuthPasswordKey] = in.Data[api.EnvPostgresPassword]
			}
		}
		return in
	}, metav1.PatchOptions{})
	return err
}

func (c *Controller) PostgresForSecret(s *core.Secret) cache.ExplicitKey {
	ctrl := metav1.GetControllerOf(s)
	ok, err := core_util.IsOwnerOfGroupKind(ctrl, kubedb.GroupName, api.ResourceKindPostgres)
	if err != nil || !ok {
		return ""
	}
	// Owner ref is set by the enterprise operator
	return cache.ExplicitKey(s.Namespace + "/" + ctrl.Name)
}

func (c *Controller) GetPostgresSecrets(db *api.Postgres) []string {
	if db.Spec.TLS != nil {
		return []string{
			db.GetCertSecretName(api.PostgresServerCert),
			db.GetCertSecretName(api.PostgresClientCert),
			db.GetCertSecretName(api.PostgresMetricsExporterCert),
		}
	}
	return nil
}
func (r *Reconciler) RequiredCertSecretNames(db *api.Postgres) []string {
	// wait for  Certificates secrets
	if db.Spec.TLS != nil {

		var sNames []string
		// server certificate
		sNames = append(sNames, db.GetCertSecretName(api.PostgresServerCert))
		// client certificate
		sNames = append(sNames, db.GetCertSecretName(api.PostgresClientCert))
		// metrics exporter certificate, if monitoring is enabled
		if db.Spec.Monitor != nil {
			// monitor certificate
			sNames = append(sNames, db.GetCertSecretName(api.PostgresMetricsExporterCert))
		}
		return sNames
	}
	return nil
}
