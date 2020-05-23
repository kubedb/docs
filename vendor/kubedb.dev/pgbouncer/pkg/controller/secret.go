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
	"context"
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
)

const UserListKey string = "userlist"

func (c *Controller) GetDefaultSecretSpec(pgbouncer *api.PgBouncer) *core.Secret {
	return &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pgbouncer.GetName() + AuthSecretSuffix,
			Namespace: pgbouncer.Namespace,
			Labels:    pgbouncer.OffshootLabels(),
		},
	}
}

func (c *Controller) CreateOrPatchDefaultSecret(pgbouncer *api.PgBouncer) (kutil.VerbType, error) {
	var myPgBouncerPass string
	var vt kutil.VerbType

	defaultSecretSpec := c.GetDefaultSecretSpec(pgbouncer)
	secret, err := c.Client.CoreV1().Secrets(defaultSecretSpec.Namespace).Get(context.TODO(), defaultSecretSpec.Name, metav1.GetOptions{})
	//if default secret exists, reuse admin pass. Else create a new pass for admin user
	if err == nil {
		myPgBouncerPass = string(secret.Data[pbAdminPassword])
	} else if kerr.IsNotFound(err) {
		myPgBouncerPass = rand.WithUniqSuffix(pbAdminUser)
	} else {
		return "", err
	}

	var myPgBouncerAdminData = fmt.Sprintf(`"%s" "%s"`, pbAdminUser, myPgBouncerPass)
	mySecretData := map[string]string{
		pbAdminData:     myPgBouncerAdminData,
		pbAdminPassword: myPgBouncerPass,
	}

	//if the referenced secret is available, add user given userList-data to default secret
	userSecretExists, userSecret, err := c.isUserSecretExists(pgbouncer)
	if err != nil && !kerr.IsNotFound(err) {
		return "", err
	}
	if userSecretExists {
		data, keyExists := userSecret.Data[UserListKey]
		if keyExists && data != nil {
			mySecretData[pbUserData] = fmt.Sprintln(myPgBouncerAdminData) + string(userSecret.Data[UserListKey])
		}
	}
	//get ca-bundle SecretData from associated PostgresDatabases
	myUpstreamCAData, err := c.getCABundlesFromAppBindingsInPgBouncerSpec(pgbouncer)
	if err != nil {
		log.Infoln(err)
		return kutil.VerbUnchanged, err
	}
	if myUpstreamCAData != "" {
		mySecretData[api.PgBouncerUpstreamServerCA] = myUpstreamCAData
	}

	defaultSecretSpec.StringData = mySecretData
	ref := metav1.NewControllerRef(pgbouncer, api.SchemeGroupVersion.WithKind(api.ResourceKindPgBouncer))
	core_util.EnsureOwnerReference(&defaultSecretSpec.ObjectMeta, ref)

	_, vt, err = core_util.CreateOrPatchSecret(context.TODO(), c.Client, defaultSecretSpec.ObjectMeta, func(in *core.Secret) *core.Secret {
		return defaultSecretSpec
	}, metav1.PatchOptions{})
	return vt, err
}

func (c *Controller) isUserSecretExists(pgbouncer *api.PgBouncer) (bool, *core.Secret, error) {
	if pgbouncer.Spec.UserListSecretRef == nil || pgbouncer.Spec.UserListSecretRef.Name == "" {
		return false, nil, nil
	}
	secret, err := c.Client.CoreV1().Secrets(pgbouncer.Namespace).Get(context.TODO(), pgbouncer.Spec.UserListSecretRef.Name, metav1.GetOptions{})
	if err == nil {
		return true, secret, nil
	}
	return false, nil, err
}

func (c *Controller) isSecretExists(meta metav1.ObjectMeta) error {
	_, err := c.Client.CoreV1().Secrets(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
	return err
}

func (c *Controller) PgBouncerForSecret(s *core.Secret) (*api.PgBouncer, error) {
	pgbouncers, err := c.pbLister.PgBouncers(s.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, pgbouncer := range pgbouncers {
		if metav1.IsControlledBy(s, pgbouncer) {
			return pgbouncer, nil
		}
	}

	return nil, nil
}
