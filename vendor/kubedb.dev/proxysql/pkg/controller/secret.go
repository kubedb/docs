/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	passgen "gomodules.xyz/password-generator"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

const (
	proxysqlUser = "proxysql"
)

func (c *Controller) ensureAuthSecret(db *api.ProxySQL) error {
	if db.Spec.AuthSecret == nil {
		authSecret, err := c.createAuthSecret(db)
		if err != nil {
			return err
		}

		proxysqlPathced, _, err := util.PatchProxySQL(
			context.TODO(),
			c.DBClient.KubedbV1alpha2(),
			db,
			func(in *api.ProxySQL) *api.ProxySQL {
				in.Spec.AuthSecret = authSecret
				return in
			},
			metav1.PatchOptions{},
		)
		if err != nil {
			return err
		}
		db.Spec.AuthSecret = proxysqlPathced.Spec.AuthSecret
	}

	return nil
}

func (c *Controller) createAuthSecret(db *api.ProxySQL) (*core.LocalObjectReference, error) {
	authSecretName := db.Name + "-auth"

	sc, err := c.checkSecret(authSecretName, db)
	if err != nil {
		return nil, err
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindProxySQL))

	if sc == nil {
		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      authSecretName,
				Namespace: db.Namespace,
				Labels:    db.OffshootSelectors(),
			},
			Type: core.SecretTypeOpaque,
			StringData: map[string]string{
				core.BasicAuthUsernameKey: proxysqlUser,
				core.BasicAuthPasswordKey: passgen.Generate(api.DefaultPasswordLength),
			},
		}

		core_util.EnsureOwnerReference(&secret.ObjectMeta, owner)

		if _, err := c.Client.CoreV1().Secrets(db.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
			return nil, err
		}
	}
	return &core.LocalObjectReference{
		Name: authSecretName,
	}, nil
}

func (c *Controller) checkSecret(secretName string, db *api.ProxySQL) (*core.Secret, error) {
	secret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	if secret.Labels[api.LabelDatabaseKind] != api.ResourceKindProxySQL ||
		secret.Labels[api.LabelProxySQLName] != db.Name {
		return nil, fmt.Errorf(`intended secret "%v/%v" already exists`, db.Namespace, secretName)
	}
	return secret, nil
}
