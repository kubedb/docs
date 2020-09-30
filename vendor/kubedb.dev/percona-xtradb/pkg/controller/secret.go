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
	mysqlUser = "root"
)

func (c *Controller) ensureDatabaseSecret(px *api.PerconaXtraDB) error {
	if px.Spec.DatabaseSecret == nil {
		secretVolumeSource, err := c.createDatabaseSecret(px)
		if err != nil {
			return err
		}

		per, _, err := util.PatchPerconaXtraDB(context.TODO(), c.ExtClient.KubedbV1alpha1(), px, func(in *api.PerconaXtraDB) *api.PerconaXtraDB {
			in.Spec.DatabaseSecret = secretVolumeSource
			return in
		}, metav1.PatchOptions{})
		if err != nil {
			return err
		}
		px.Spec.DatabaseSecret = per.Spec.DatabaseSecret
		return nil
	}
	return c.upgradeDatabaseSecret(px)
}

func (c *Controller) createDatabaseSecret(px *api.PerconaXtraDB) (*core.SecretVolumeSource, error) {
	authSecretName := px.Name + "-auth"

	sc, err := c.checkSecret(authSecretName, px)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   authSecretName,
				Labels: px.OffshootLabels(),
			},
			Type: core.SecretTypeOpaque,
			StringData: map[string]string{
				core.BasicAuthUsernameKey: mysqlUser,
				core.BasicAuthPasswordKey: passgen.Generate(api.DefaultPasswordLength),
			},
		}

		if _, err := c.Client.CoreV1().Secrets(px.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
			return nil, err
		}
	}
	return &core.SecretVolumeSource{
		SecretName: authSecretName,
	}, nil
}

// This is done to fix 0.8.0 -> 0.9.0 upgrade due to
// https://github.com/kubedb/percona-xtradb/pull/115/files#diff-10ddaf307bbebafda149db10a28b9c24R17 commit
func (c *Controller) upgradeDatabaseSecret(px *api.PerconaXtraDB) error {
	meta := metav1.ObjectMeta{
		Name:      px.Spec.DatabaseSecret.SecretName,
		Namespace: px.Namespace,
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

func (c *Controller) checkSecret(secretName string, px *api.PerconaXtraDB) (*core.Secret, error) {
	secret, err := c.Client.CoreV1().Secrets(px.Namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	if secret.Labels[api.LabelDatabaseKind] != api.ResourceKindPerconaXtraDB ||
		secret.Labels[api.LabelDatabaseName] != px.Name {
		return nil, fmt.Errorf(`intended secret "%v/%v" already exists`, px.Namespace, secretName)
	}
	return secret, nil
}
