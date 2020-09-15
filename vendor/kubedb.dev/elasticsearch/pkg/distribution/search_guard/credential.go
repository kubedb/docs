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

package search_guard

import (
	"context"
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/elasticsearch/pkg/lib/user"

	"github.com/appscode/go/crypto/rand"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

func (es *Elasticsearch) EnsureDatabaseSecret() error {

	err := es.setMissingUsersAndRolesMapping()
	if err != nil {
		return errors.Wrap(err, "failed to set missing internal users or role mapping")
	}

	// For admin user
	dbSecretVolume := es.elasticsearch.Spec.DatabaseSecret
	if dbSecretVolume == nil {
		// create admin credential secret.
		// If the secret already exists in the same name,
		// validate it (ie. it contains username, password as keys).
		var err error
		pass := rand.Characters(8)
		if dbSecretVolume, err = es.createOrSyncUserCredSecret(string(api.ElasticsearchInternalUserAdmin), pass); err != nil {
			return err
		}

		// update the ES object,
		// Add admin credential secret name to Spec.DatabaseSecret.
		newES, _, err := util.PatchElasticsearch(context.TODO(), es.extClient.KubedbV1alpha1(), es.elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
			in.Spec.DatabaseSecret = dbSecretVolume
			return in
		}, metav1.PatchOptions{})
		if err != nil {
			return err
		}

		es.elasticsearch = newES
		return nil
	} else {
		// Get the secret and validate it.
		dbSecret, err := es.kClient.CoreV1().Secrets(es.elasticsearch.Namespace).Get(context.TODO(), dbSecretVolume.SecretName, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to get credential secret: %s/%s", es.elasticsearch.Namespace, dbSecretVolume.SecretName))
		}

		err = es.validateAndSyncLabels(dbSecret)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to validate/sync secret: %s/%s", dbSecret.Namespace, dbSecret.Name))
		}
	}

	// For all internal users
	for username := range es.elasticsearch.Spec.InternalUsers {
		// secret for admin user is handled separately
		if username == string(api.ElasticsearchInternalUserAdmin) {
			continue
		}

		pass := rand.Characters(8)
		_, err := es.createOrSyncUserCredSecret(username, pass)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to create credential secret for user: %s", username))
		}
	}
	return nil
}

func (es *Elasticsearch) createOrSyncUserCredSecret(username, password string) (*corev1.SecretVolumeSource, error) {

	dbSecret, err := es.findSecret(es.elasticsearch.UserCredSecretName(username))
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

		return &corev1.SecretVolumeSource{
			SecretName: dbSecret.Name,
		}, nil
	}

	// Create the secret
	var data = map[string][]byte{
		corev1.BasicAuthUsernameKey: []byte(username),
		corev1.BasicAuthPasswordKey: []byte(password),
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   es.elasticsearch.UserCredSecretName(username),
			Labels: es.elasticsearch.OffshootLabels(),
		},
		Type: corev1.SecretTypeBasicAuth,
		Data: data,
	}

	// add owner reference
	owner := metav1.NewControllerRef(es.elasticsearch, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))
	core_util.EnsureOwnerReference(&secret.ObjectMeta, owner)

	if _, err := es.kClient.CoreV1().Secrets(es.elasticsearch.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
		return nil, err
	}

	return &corev1.SecretVolumeSource{
		SecretName: secret.Name,
	}, nil
}

func (es *Elasticsearch) validateAndSyncLabels(secret *corev1.Secret) error {
	if secret == nil {
		return errors.New("secret is empty")
	}

	if value, exist := secret.Data[corev1.BasicAuthUsernameKey]; !exist || len(value) == 0 {
		return errors.New("username is missing")
	}

	if value, exist := secret.Data[corev1.BasicAuthPasswordKey]; !exist || len(value) == 0 {
		return errors.New("password is missing")
	}

	// If secret is owned by this elasticsearch object,
	// update the labels.
	// Labels hold information like elasticsearch version,
	// should be synced.
	ctrl := metav1.GetControllerOf(secret)
	if ctrl != nil &&
		ctrl.Kind == api.ResourceKindElasticsearch && ctrl.Name == es.elasticsearch.Name {

		// sync labels
		if _, _, err := core_util.CreateOrPatchSecret(context.TODO(), es.kClient, secret.ObjectMeta, func(in *corev1.Secret) *corev1.Secret {
			in.Labels = core_util.UpsertMap(in.Labels, es.elasticsearch.OffshootLabels())
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
	if es.elasticsearch.Spec.InternalUsers != nil {
		userList = es.elasticsearch.Spec.InternalUsers
	}

	user.SetMissingUser(userList, api.ElasticsearchInternalUserAdmin, api.ElasticsearchUserSpec{Reserved: true})
	user.SetMissingUser(userList, api.ElasticsearchInternalUserKibanaserver, api.ElasticsearchUserSpec{Reserved: true})
	user.SetMissingUser(userList, api.ElasticsearchInternalUserKibanaro, api.ElasticsearchUserSpec{Reserved: false})
	user.SetMissingUser(userList, api.ElasticsearchInternalUserLogstash, api.ElasticsearchUserSpec{Reserved: false})
	user.SetMissingUser(userList, api.ElasticsearchInternalUserReadall, api.ElasticsearchUserSpec{Reserved: false})
	user.SetMissingUser(userList, api.ElasticsearchInternalUserSnapshotrestore, api.ElasticsearchUserSpec{Reserved: false})

	// Ref:
	// - https://docs.search-guard.com/latest/upgrading-6-7
	// Set user for metrics-exporter sidecar, if monitoring is enabled.
	if es.elasticsearch.Spec.Monitor != nil {
		user.SetMissingUser(userList, api.ElasticsearchInternalUserMetricsExporter, api.ElasticsearchUserSpec{
			Reserved: false,
		})
	}

	// RolesMapping
	rolesMapping := make(map[string]api.ElasticsearchRoleMapSpec)
	if es.elasticsearch.Spec.RolesMapping != nil {
		rolesMapping = es.elasticsearch.Spec.RolesMapping
	}

	// Add permission for metrics-exporter sidecar, if monitoring is enabled.
	if es.elasticsearch.Spec.Monitor != nil {
		// The metrics_exporter user will need to have access to
		// readall_and_monitor role.

		// readall_and_monitor role name varies in ES version
		// 	V7        = "SGS_READALL_AND_MONITOR"
		//	V6        = "sg_readall_and_monitor"
		var readallMonitor string
		if string(es.esVersion.Spec.Version[0]) == "6" {
			readallMonitor = ReadallMonitorRoleV6
		} else {
			readallMonitor = ReadallMonitorRoleV7
		}

		// Create rolesMapping if not exists.
		if value, exist := rolesMapping[readallMonitor]; exist {
			value.Users = upsertStringSlice(value.Users, string(api.ElasticsearchInternalUserMetricsExporter))
		} else {
			rolesMapping[readallMonitor] = api.ElasticsearchRoleMapSpec{
				Users: []string{string(api.ElasticsearchInternalUserMetricsExporter)},
			}
		}
	}

	newES, _, err := util.PatchElasticsearch(context.TODO(), es.extClient.KubedbV1alpha1(), es.elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
		in.Spec.InternalUsers = userList
		in.Spec.RolesMapping = rolesMapping
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	es.elasticsearch = newES
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
