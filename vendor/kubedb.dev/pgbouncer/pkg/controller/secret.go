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

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	passgen "gomodules.xyz/password-generator"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
)

const UserListKey string = "userlist"

func (c *Controller) GetDefaultSecretSpec(db *api.PgBouncer) *core.Secret {
	return &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db.AuthSecretName(),
			Namespace: db.Namespace,
			Labels:    db.OffshootLabels(),
		},
	}
}

func (c *Controller) ensureAuthSecret(db *api.PgBouncer) (kutil.VerbType, error) {
	userSecret, err := c.GetUserSecret(db)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	//get ca-bundle SecretData from associated PostgresDatabases
	upstreamCAData, err := c.getCABundlesFromAppBindingsInPgBouncerSpec(db)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	objMeta := metav1.ObjectMeta{
		Name:      db.AuthSecretName(),
		Namespace: db.Namespace,
	}
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindPgBouncer))

	_, vt, err := core_util.CreateOrPatchSecret(context.TODO(), c.Client, objMeta, func(in *core.Secret) *core.Secret {
		in.Labels = db.OffshootLabels()
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)

		if in.Data == nil {
			in.Data = map[string][]byte{}
		}

		if _, ok := in.Data[pbAdminPasswordKey]; !ok {
			in.Data[pbAdminPasswordKey] = []byte(passgen.Generate(api.DefaultPasswordLength))
		}

		var pbAdminData = fmt.Sprintf(`"%s" "%s"`, api.PgBouncerAdminUsername, string(in.Data[pbAdminPasswordKey]))
		in.Data[pbAdminDataKey] = []byte(pbAdminData)

		// If user secret is available, add user given userList-data to default secret
		if userSecret != nil {
			data, keyExists := userSecret.Data[UserListKey]
			if keyExists && data != nil {
				in.Data[pbUserDataKey] = []byte(pbAdminData + "\n" + string(userSecret.Data[UserListKey]))
			}
		}

		if upstreamCAData != "" {
			in.Data[api.PgBouncerUpstreamServerCA] = []byte(upstreamCAData)
		}
		return in
	}, metav1.PatchOptions{})
	return vt, err
}

func (c *Controller) GetUserSecret(db *api.PgBouncer) (*core.Secret, error) {
	if db.Spec.UserListSecretRef == nil || db.Spec.UserListSecretRef.Name == "" {
		return nil, nil
	}
	return c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.Spec.UserListSecretRef.Name, metav1.GetOptions{})
}

func (c *Controller) PgBouncerForSecret(s *core.Secret) cache.ExplicitKey {
	ctrl := metav1.GetControllerOf(s)
	ok, err := core_util.IsOwnerOfGroupKind(ctrl, kubedb.GroupName, api.ResourceKindPgBouncer)
	if err != nil || !ok {
		return ""
	}
	// Owner ref is set by the enterprise operator
	return cache.ExplicitKey(s.Namespace + "/" + ctrl.Name)
}
