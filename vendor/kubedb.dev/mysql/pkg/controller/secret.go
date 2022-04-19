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
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/pkg/errors"
	passgen "gomodules.xyz/password-generator"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

const (
	mysqlUser = "root"
)

// ensureAuthSecret create or patch auth secret for mysql instance
// for read replica it validates the allowance of reading from the source object and create secrets to connect with source

func (c *Reconciler) ensureAuthSecret(db *api.MySQL) error {
	if db.IsReadReplica() {
		err := c.ensureReadReplicaAuthSecret(db)
		if err != nil {
			return errors.Wrapf(err, "unable to ensure Read Replica auth Secret %s", db.GetNameSpacedName())
		}
	} else {
		err := c.ensureInstanceAuthSecret(db)
		if err != nil {
			return errors.Wrapf(err, "unable to ensure Auth Secret for %s", db.GetNameSpacedName())
		}
	}
	if db.Spec.AuthSecret != nil {
		return c.upgradeAuthSecret(db)
	}
	return nil
}

func (c *Reconciler) ensureInstanceAuthSecret(db *api.MySQL) error {
	sec, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.GetAuthSecretName(), metav1.GetOptions{})
	if err == nil {
		err = c.validateAndSyncSecret(sec, db)
		if err != nil {
			return err
		}
		err = c.patchAuthSecret(db)
		if err != nil {
			return err
		}
	}

	if err != nil && kerr.IsNotFound(err) {

		_, err := c.createAuthSecret(db)
		if err != nil {
			return errors.Wrapf(err, "err while create atuh Secret %s ", db.GetNameSpacedName())
		}

		err = c.patchAuthSecret(db)
		if err != nil {
			return err
		}

	}

	return err
}

func (c *Reconciler) ensureReadReplicaAuthSecret(db *api.MySQL) error {
	sec, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.GetAuthSecretName(), metav1.GetOptions{})
	if err == nil {
		err = c.validateAndSyncSecret(sec, db)
		if err != nil {
			return err
		}

		err = c.patchAuthSecret(db)
		if err != nil {
			return err
		}
	}

	if err != nil && kerr.IsNotFound(err) {
		err := c.createReadReplicaAuthSecret(db)
		if err != nil {
			return errors.Wrapf(err, "error while creating ReadReplica auth Secret %s/%s", db.Name, db.Namespace)
		}

		err = c.patchAuthSecret(db)
		if err != nil {
			return err
		}

	}
	return err
}

func (c *Reconciler) patchAuthSecret(db *api.MySQL) error {
	ms, _, err := util.PatchMySQL(context.TODO(), c.DBClient.KubedbV1alpha2(), db, func(in *api.MySQL) *api.MySQL {
		in.Spec.AuthSecret = &core.LocalObjectReference{
			Name: db.GetAuthSecretName(),
		}
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	db.Spec.AuthSecret = ms.Spec.AuthSecret
	return nil
}

func (c *Reconciler) validateAndSyncSecret(secret *core.Secret, db *api.MySQL) error {
	data := secret.Data
	if _, exists := data[core.BasicAuthUsernameKey]; !exists {
		return errors.New("user name is missing")
	}
	if _, exists := data[core.BasicAuthPasswordKey]; !exists {
		return errors.New("password is missing")
	}

	if owned, _ := core_util.IsOwnedBy(secret, db); owned {
		if _, _, err := core_util.CreateOrPatchSecret(context.TODO(), c.Client, secret.ObjectMeta, func(in *core.Secret) *core.Secret {
			in.Labels = meta_util.OverwriteKeys(in.Labels, db.OffshootLabels())
			return in
		}, metav1.PatchOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func (c *Reconciler) createAuthSecret(db *api.MySQL) (string, error) {
	authSecretName := db.GetAuthSecretName()
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

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))
	core_util.EnsureOwnerReference(secret, owner)

	secret, err := c.Client.CoreV1().Secrets(db.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		return "", err
	} else {
		c.Recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created database auth secret",
		)
		return secret.Name, err
	}
}

// createReadReplicaAuthSecret Create secret from Source appbinding
// if source has ssl it will create a tls secret for read replica to connect with source
func (c *Reconciler) createReadReplicaAuthSecret(db *api.MySQL) error {
	if !db.IsReadReplica() {
		return nil
	}
	readAppBindingName := db.Spec.Topology.ReadReplica.SourceRef.Name
	readAppBindingNameSpace := db.Spec.Topology.ReadReplica.SourceRef.Namespace

	// get the secret from appbinding
	readAppBinding, err := c.AppCatalogClient.AppcatalogV1alpha1().AppBindings(readAppBindingNameSpace).Get(context.TODO(), readAppBindingName, metav1.GetOptions{})
	if err != nil {
		klog.Error("unable to get read appbinding", err)

		return err
	}

	readSecretName := readAppBinding.Spec.Secret.Name
	secretNamespace := readAppBinding.Namespace
	readSecret, err := c.Client.CoreV1().Secrets(secretNamespace).Get(context.TODO(), readSecretName, metav1.GetOptions{})
	if err != nil {
		klog.Error("unable to get secret form appbinding", err)
		return err
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMySQL))
	meta := metav1.ObjectMeta{
		Name:      db.GetAuthSecretName(),
		Namespace: db.Namespace,
	}

	_, _, err = core_util.CreateOrPatchSecret(context.TODO(), c.Client, meta, func(in *core.Secret) *core.Secret {
		core_util.EnsureOwnerReference(in, owner)
		in.Data = readSecret.Data
		in.Type = core.SecretTypeBasicAuth
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		klog.Error("unable to read-replica source secret the secret")
		return err
	}

	if readAppBinding.Spec.ClientConfig.CABundle != nil {

		secretMeta := metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-source-tls-secret", db.Name),
			Namespace: db.Namespace,
		}
		_, _, err = core_util.CreateOrPatchSecret(context.TODO(), c.Client, secretMeta, func(in *core.Secret) *core.Secret {
			core_util.EnsureOwnerReference(in, owner)
			clientCA := make(map[string][]byte)
			clientCA["ca.crt"] = readAppBinding.Spec.ClientConfig.CABundle
			in.Data = clientCA
			return in
		}, metav1.PatchOptions{})
		if err != nil {
			klog.Error("unable to crate ca secret for ", db.Name)
			return err
		}
	}

	return nil
}

// This is done to fix 0.8.0 -> 0.9.0 upgrade due to
// https://github.com/kubedb/mysql/pull/115/files#diff-10ddaf307bbebafda149db10a28b9c24R17 commit
func (c *Reconciler) upgradeAuthSecret(db *api.MySQL) error {
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

func (c *Controller) mysqlForSecret(s *core.Secret) cache.ExplicitKey {
	ctrl := metav1.GetControllerOf(s)
	ok, err := core_util.IsOwnerOfGroupKind(ctrl, kubedb.GroupName, api.ResourceKindMySQL)
	if err != nil || !ok {
		return ""
	}

	// Owner ref is set by the enterprise operator
	return cache.ExplicitKey(s.Namespace + "/" + ctrl.Name)
}
