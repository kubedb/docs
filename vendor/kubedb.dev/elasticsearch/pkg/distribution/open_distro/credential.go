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

package open_distro

import (
	"context"
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"
	"kubedb.dev/elasticsearch/pkg/lib/user"

	"github.com/pkg/errors"
	"gomodules.xyz/password-generator"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

func (es *Elasticsearch) EnsureAuthSecret() error {

	err := es.setMissingUsersAndRolesMapping()
	if err != nil {
		return errors.Wrap(err, "failed to set missing internal users or roles mapping")
	}

	// For admin user
	authSecret := es.db.Spec.AuthSecret
	if authSecret == nil {
		// create admin credential secret.
		// If the secret already exists in the same name,
		// validate it (ie. it contains username, password as keys).
		var err error
		pass := password.Generate(api.DefaultPasswordLength)
		if authSecret, err = es.createOrSyncUserCredSecret(string(api.ElasticsearchInternalUserAdmin), pass); err != nil {
			return err
		}

		// update the ES object,
		// Add admin credential secret name to Spec.AuthSecret.
		newES, _, err := util.PatchElasticsearch(context.TODO(), es.extClient.KubedbV1alpha2(), es.db, func(in *api.Elasticsearch) *api.Elasticsearch {
			in.Spec.AuthSecret = authSecret
			return in
		}, metav1.PatchOptions{})
		if err != nil {
			return err
		}

		es.db = newES
	} else {
		// Get the secret and validate it.
		dbSecret, err := es.kClient.CoreV1().Secrets(es.db.Namespace).Get(context.TODO(), authSecret.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to get credential secret: %s/%s", es.db.Namespace, authSecret.Name))
		}

		err = es.validateAndSyncLabels(dbSecret)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to validate/sync secret: %s/%s", dbSecret.Namespace, dbSecret.Name))
		}
	}

	// For all internal users
	for username := range es.db.Spec.InternalUsers {
		// secret for admin user is handled separately
		if username == string(api.ElasticsearchInternalUserAdmin) {
			continue
		}

		pass := password.Generate(api.DefaultPasswordLength)
		_, err := es.createOrSyncUserCredSecret(username, pass)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to create credential secret for user: %s", username))
		}
	}
	return nil
}

func (es *Elasticsearch) createOrSyncUserCredSecret(username, password string) (*core.LocalObjectReference, error) {

	dbSecret, err := es.findSecret(es.db.UserCredSecretName(username))
	if err != nil {
		return nil, err
	}

	// if a secret already exist with the given name.
	// Validate it, whether it contains the following keys:
	//	- username
	// 	- password
	// If the secret is owned by this object, sync the labels.
	// Return secretName.
	if dbSecret != nil {
		err = es.validateAndSyncLabels(dbSecret)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to validate/sync secret: %s/%s", dbSecret.Namespace, dbSecret.Name))
		}

		return &core.LocalObjectReference{
			Name: dbSecret.Name,
		}, nil
	}

	// Create the secret
	var data = map[string][]byte{
		core.BasicAuthUsernameKey: []byte(username),
		core.BasicAuthPasswordKey: []byte(password),
	}

	secret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   es.db.UserCredSecretName(username),
			Labels: es.db.OffshootLabels(),
		},
		Type: core.SecretTypeBasicAuth,
		Data: data,
	}

	// add owner reference
	owner := metav1.NewControllerRef(es.db, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))
	core_util.EnsureOwnerReference(&secret.ObjectMeta, owner)

	if _, err := es.kClient.CoreV1().Secrets(es.db.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
		return nil, err
	}

	return &core.LocalObjectReference{
		Name: secret.Name,
	}, nil
}

func (es *Elasticsearch) validateAndSyncLabels(secret *core.Secret) error {
	if secret == nil {
		return errors.New("secret is empty")
	}

	if value, exist := secret.Data[core.BasicAuthUsernameKey]; !exist || len(value) == 0 {
		return errors.New("username is missing")
	}

	if value, exist := secret.Data[core.BasicAuthPasswordKey]; !exist || len(value) == 0 {
		return errors.New("password is missing")
	}

	// If secret is owned by this elasticsearch object,
	// update the labels.
	// Labels hold information like elasticsearch version,
	// should be synced.
	ctrl := metav1.GetControllerOf(secret)
	if ctrl != nil &&
		ctrl.Kind == api.ResourceKindElasticsearch && ctrl.Name == es.db.Name {

		// sync labels
		if _, _, err := core_util.CreateOrPatchSecret(context.TODO(), es.kClient, secret.ObjectMeta, func(in *core.Secret) *core.Secret {
			in.Labels = core_util.UpsertMap(in.Labels, es.db.OffshootLabels())
			return in
		}, metav1.PatchOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func (es *Elasticsearch) setMissingUsersAndRolesMapping() error {

	// Users
	userList := make(map[string]api.ElasticsearchUserSpec)
	if es.db.Spec.InternalUsers != nil {
		userList = es.db.Spec.InternalUsers
	}

	user.SetMissingUser(userList, api.ElasticsearchInternalUserAdmin, api.ElasticsearchUserSpec{Reserved: true})
	user.SetMissingUser(userList, api.ElasticsearchInternalUserKibanaserver, api.ElasticsearchUserSpec{Reserved: true})
	user.SetMissingUser(userList, api.ElasticsearchInternalUserKibanaro, api.ElasticsearchUserSpec{Reserved: false})
	user.SetMissingUser(userList, api.ElasticsearchInternalUserLogstash, api.ElasticsearchUserSpec{Reserved: false})
	user.SetMissingUser(userList, api.ElasticsearchInternalUserReadall, api.ElasticsearchUserSpec{Reserved: false})
	user.SetMissingUser(userList, api.ElasticsearchInternalUserSnapshotrestore, api.ElasticsearchUserSpec{Reserved: false})

	// Set user for metrics-exporter sidecar, if monitoring is enabled.
	if es.db.Spec.Monitor != nil {
		user.SetMissingUser(userList, api.ElasticsearchInternalUserMetricsExporter, api.ElasticsearchUserSpec{
			Reserved: false,
		})
	}

	// RolesMapping
	rolesMapping := make(map[string]api.ElasticsearchRoleMapSpec)
	if es.db.Spec.RolesMapping != nil {
		rolesMapping = es.db.Spec.RolesMapping
	}

	// Add permission for metrics-exporter sidecar, if monitoring is enabled.
	if es.db.Spec.Monitor != nil {
		// The metrics_exporter user will need to have access to
		// readall_and_monitor role.
		// Create rolesMapping if not exists.
		if value, check := rolesMapping[api.ElasticsearchOpendistroReadallMonitorRole]; check {
			value.Users = upsertStringSlice(value.Users, string(api.ElasticsearchInternalUserMetricsExporter))
		} else {
			rolesMapping[api.ElasticsearchOpendistroReadallMonitorRole] = api.ElasticsearchRoleMapSpec{
				Users: []string{string(api.ElasticsearchInternalUserMetricsExporter)},
			}
		}
	}

	newES, _, err := util.PatchElasticsearch(context.TODO(), es.extClient.KubedbV1alpha2(), es.db, func(in *api.Elasticsearch) *api.Elasticsearch {
		in.Spec.InternalUsers = userList
		in.Spec.RolesMapping = rolesMapping
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	es.db = newES
	return nil
}

func upsertStringSlice(inSlice []string, values ...string) []string {
	upsert := func(m string) {
		for _, v := range inSlice {
			if v == m {
				return
			}
		}
		inSlice = append(inSlice, m)
	}

	for _, value := range values {
		upsert(value)
	}
	return inSlice
}
