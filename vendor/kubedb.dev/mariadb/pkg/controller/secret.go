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

const (
	mysqlUser = "root"
)

func (c *Controller) ensureAuthSecret(db *api.MariaDB) error {
	if db.Spec.AuthSecret == nil {
		authSecret, err := c.createAuthSecret(db)
		if err != nil {
			return err
		}

		per, _, err := util.PatchMariaDB(context.TODO(), c.DBClient.KubedbV1alpha2(), db, func(in *api.MariaDB) *api.MariaDB {
			in.Spec.AuthSecret = authSecret
			return in
		}, metav1.PatchOptions{})
		if err != nil {
			return err
		}
		db.Spec.AuthSecret = per.Spec.AuthSecret
		return nil
	}
	return c.upgradeAuthSecret(db)
}

func (c *Controller) createAuthSecret(db *api.MariaDB) (*core.LocalObjectReference, error) {
	authSecretName := db.Name + "-auth"

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
			Type: core.SecretTypeBasicAuth,
			StringData: map[string]string{
				core.BasicAuthUsernameKey: mysqlUser,
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

// This is done to fix 0.8.0 -> 0.9.0 upgrade due to
// https://github.com/kubedb/mariadb/pull/115/files#diff-10ddaf307bbebafda149db10a28b9c24R17 commit
func (c *Controller) upgradeAuthSecret(db *api.MariaDB) error {
	meta := metav1.ObjectMeta{
		Name:      db.Spec.AuthSecret.Name,
		Namespace: db.Namespace,
	}

	_, _, err := core_util.CreateOrPatchSecret(context.TODO(), c.Client, meta, func(in *core.Secret) *core.Secret {
		if _, ok := in.Data[core.BasicAuthUsernameKey]; !ok {
			if val, ok2 := in.Data["user"]; ok2 {
				in.StringData = map[string]string{core.BasicAuthUsernameKey: string(val)}
			}
		}
		return in
	}, metav1.PatchOptions{})
	return err
}

func (c *Controller) checkSecret(secretName string, db *api.MariaDB) (*core.Secret, error) {
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

func (c *Controller) mariaDBForSecret(s *core.Secret) cache.ExplicitKey {
	ctrl := metav1.GetControllerOf(s)
	ok, err := core_util.IsOwnerOfGroupKind(ctrl, kubedb.GroupName, api.ResourceKindMariaDB)
	if err != nil || !ok {
		return ""
	}

	// Owner ref is set by the enterprise operator
	return cache.ExplicitKey(s.Namespace + "/" + ctrl.Name)
}
