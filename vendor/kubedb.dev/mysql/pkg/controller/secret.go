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
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"

	"github.com/appscode/go/crypto/rand"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	core_util "kmodules.xyz/client-go/core/v1"
)

const (
	mysqlUser = "root"
)

func (c *Controller) ensureDatabaseSecret(mysql *api.MySQL) error {
	if mysql.Spec.DatabaseSecret == nil {
		secretVolumeSource, err := c.createDatabaseSecret(mysql)
		if err != nil {
			return err
		}

		ms, _, err := util.PatchMySQL(context.TODO(), c.ExtClient.KubedbV1alpha1(), mysql, func(in *api.MySQL) *api.MySQL {
			in.Spec.DatabaseSecret = secretVolumeSource
			return in
		}, metav1.PatchOptions{})
		if err != nil {
			return err
		}
		mysql.Spec.DatabaseSecret = ms.Spec.DatabaseSecret
		return nil
	}
	return c.upgradeDatabaseSecret(mysql)
}

func (c *Controller) createDatabaseSecret(mysql *api.MySQL) (*core.SecretVolumeSource, error) {
	authSecretName := mysql.Name + "-auth"

	sc, err := c.checkSecret(authSecretName, mysql)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		randPassword := ""

		// if the password starts with "-", it will cause error in bash scripts (in mysql-tools)
		for randPassword = rand.GeneratePassword(); randPassword[0] == '-'; {
		}

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   authSecretName,
				Labels: mysql.OffshootLabels(),
			},
			Type: core.SecretTypeOpaque,
			StringData: map[string]string{
				core.BasicAuthUsernameKey: mysqlUser,
				core.BasicAuthPasswordKey: randPassword,
			},
		}
		if _, err := c.Client.CoreV1().Secrets(mysql.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
			return nil, err
		}
	}
	return &core.SecretVolumeSource{
		SecretName: authSecretName,
	}, nil
}

// This is done to fix 0.8.0 -> 0.9.0 upgrade due to
// https://github.com/kubedb/mysql/pull/115/files#diff-10ddaf307bbebafda149db10a28b9c24R17 commit
func (c *Controller) upgradeDatabaseSecret(mysql *api.MySQL) error {
	meta := metav1.ObjectMeta{
		Name:      mysql.Spec.DatabaseSecret.SecretName,
		Namespace: mysql.Namespace,
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

func (c *Controller) checkSecret(secretName string, mysql *api.MySQL) (*core.Secret, error) {
	secret, err := c.Client.CoreV1().Secrets(mysql.Namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	if secret.Labels[api.LabelDatabaseKind] != api.ResourceKindMySQL ||
		secret.Labels[api.LabelDatabaseName] != mysql.Name {
		return nil, fmt.Errorf(`intended secret "%v/%v" already exists`, mysql.Namespace, secretName)
	}
	return secret, nil
}

func (c *Controller) mysqlForSecret(s *core.Secret) cache.ExplicitKey {
	ctrl := metav1.GetControllerOf(s)
	ok, err := core_util.IsOwnerOfGroupKind(ctrl, kubedb.GroupName, api.ResourceKindMySQL)
	if err != nil || !ok {
		return ""
	}

	// Owner ref is set by the enterprise operator
	return cache.ExplicitKey(s.Namespace + "/" + ctrl.Name)
}
