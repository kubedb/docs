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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	passgen "gomodules.xyz/password-generator"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

func (c *Controller) ensureSentinelAuthSecret(db *api.RedisSentinel) error {
	authSecret := db.Spec.AuthSecret
	if authSecret == nil {
		var err error
		if authSecret, err = c.createSentinelAuthSecret(db); err != nil {
			return err
		}
		pg, _, err := util.PatchRedisSentinel(context.TODO(), c.DBClient.KubedbV1alpha2(), db, func(in *api.RedisSentinel) *api.RedisSentinel {
			in.Spec.AuthSecret = authSecret
			return in
		}, metav1.PatchOptions{})
		if err != nil {
			return err
		}
		db.Spec.AuthSecret = pg.Spec.AuthSecret
	} else {
		err := c.upgradeSentinelAuthSecret(db)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) findSentinelAuthSecret(db *api.RedisSentinel) (*core.Secret, error) {
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

func (c *Controller) createSentinelAuthSecret(db *api.RedisSentinel) (*core.LocalObjectReference, error) {
	databaseSecret, err := c.findSentinelAuthSecret(db)
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
			core.BasicAuthUsernameKey: []byte("root"),
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

func (c *Controller) upgradeSentinelAuthSecret(db *api.RedisSentinel) error {
	meta := metav1.ObjectMeta{
		Name:      db.Spec.AuthSecret.Name,
		Namespace: db.Namespace,
	}

	_, _, err := core_util.CreateOrPatchSecret(context.TODO(), c.Client, meta, func(in *core.Secret) *core.Secret {
		if in.Data == nil {
			in.Data = map[string][]byte{}
		}
		if _, ok := in.Data[core.BasicAuthUsernameKey]; !ok {
			in.Data[core.BasicAuthUsernameKey] = []byte("root")
		}
		if _, ok := in.Data[core.BasicAuthPasswordKey]; !ok {
			in.Data[core.BasicAuthPasswordKey] = in.Data[api.EnvRedisPassword]
		}
		return in
	}, metav1.PatchOptions{})
	return err
}
