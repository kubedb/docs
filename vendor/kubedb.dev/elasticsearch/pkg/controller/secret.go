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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"

	"github.com/appscode/go/crypto/rand"
	"golang.org/x/crypto/bcrypt"
	"gomodules.xyz/cert"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AdminUser          = "admin"
	ElasticUser        = "elastic"
	KeyAdminUserName   = "ADMIN_USERNAME"
	KeyAdminPassword   = "ADMIN_PASSWORD"
	ReadAllUser        = "readall"
	KeyReadAllUserName = "READALL_USERNAME"
	KeyReadAllPassword = "READALL_PASSWORD"
	ExporterSecretPath = "/var/run/secrets/kubedb.com/"
)

func (c *Controller) ensureCertSecret(elasticsearch *api.Elasticsearch) error {
	if elasticsearch.Spec.DisableSecurity {
		return nil
	}

	certSecretVolumeSource := elasticsearch.Spec.CertificateSecret
	if certSecretVolumeSource == nil {
		var err error
		if certSecretVolumeSource, err = c.createCertSecret(elasticsearch); err != nil {
			return err
		}
		es, _, err := util.PatchElasticsearch(context.TODO(), c.ExtClient.KubedbV1alpha1(), elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
			in.Spec.CertificateSecret = certSecretVolumeSource
			return in
		}, metav1.PatchOptions{})
		if err != nil {
			return err
		}
		elasticsearch.Spec.CertificateSecret = es.Spec.CertificateSecret
	}
	return nil
}

func (c *Controller) ensureDatabaseSecret(elasticsearch *api.Elasticsearch) error {
	databaseSecretVolume := elasticsearch.Spec.DatabaseSecret
	if databaseSecretVolume == nil {
		var err error
		if databaseSecretVolume, err = c.createDatabaseSecret(elasticsearch); err != nil {
			return err
		}
		es, _, err := util.PatchElasticsearch(context.TODO(), c.ExtClient.KubedbV1alpha1(), elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
			in.Spec.DatabaseSecret = databaseSecretVolume
			return in
		}, metav1.PatchOptions{})
		if err != nil {
			return err
		}
		elasticsearch.Spec.DatabaseSecret = es.Spec.DatabaseSecret
		return nil
	}
	return nil
}

func (c *Controller) findCertSecret(elasticsearch *api.Elasticsearch) (*core.Secret, error) {
	name := fmt.Sprintf("%v-cert", elasticsearch.OffshootName())

	secret, err := c.Client.CoreV1().Secrets(elasticsearch.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	if secret.Labels[api.LabelDatabaseKind] != api.ResourceKindElasticsearch ||
		secret.Labels[api.LabelDatabaseName] != elasticsearch.Name {
		return nil, fmt.Errorf(`intended secret "%v/%v" already exists`, elasticsearch.Namespace, name)
	}

	return secret, nil
}

func (c *Controller) createCertSecret(elasticsearch *api.Elasticsearch) (*core.SecretVolumeSource, error) {
	certSecret, err := c.findCertSecret(elasticsearch)
	if err != nil {
		return nil, err
	}
	if certSecret != nil {
		return &core.SecretVolumeSource{
			SecretName: certSecret.Name,
		}, nil
	}

	esVersion, err := c.esVersionLister.Get(string(elasticsearch.Spec.Version))
	if err != nil {
		return nil, err
	}

	certPath := fmt.Sprintf("%v/%v", certsDir, rand.Characters(3))
	if err := os.MkdirAll(certPath, os.ModePerm); err != nil {
		return nil, err
	}

	caKey, caCert, pass, err := createCaCertificate(certPath)
	if err != nil {
		return nil, err
	}
	err = createNodeCertificate(certPath, elasticsearch, caKey, caCert, pass)
	if err != nil {
		return nil, err
	}
	if esVersion.Spec.AuthPlugin == v1alpha1.ElasticsearchAuthPluginSearchGuard {
		err = createAdminCertificate(certPath, caKey, caCert, pass)
		if err != nil {
			return nil, err
		}
	}
	root, err := ioutil.ReadFile(filepath.Join(certPath, rootKeyStore))
	if err != nil {
		return nil, err
	}
	node, err := ioutil.ReadFile(filepath.Join(certPath, nodeKeyStore))
	if err != nil {
		return nil, err
	}

	data := map[string][]byte{
		rootKeyStore: root,
		nodeKeyStore: node,
	}

	if esVersion.Spec.AuthPlugin == v1alpha1.ElasticsearchAuthPluginSearchGuard {
		sgadmin, err := ioutil.ReadFile(filepath.Join(certPath, sgAdminKeyStore))
		if err != nil {
			return nil, err
		}

		data[sgAdminKeyStore] = sgadmin

	}

	if err := createClientCertificate(certPath, elasticsearch, caKey, caCert, pass); err != nil {
		return nil, err
	}

	client, err := ioutil.ReadFile(filepath.Join(certPath, clientKeyStore))
	if err != nil {
		return nil, err
	}

	data[rootCert] = cert.EncodeCertPEM(caCert)
	data[clientKeyStore] = client

	name := fmt.Sprintf("%v-cert", elasticsearch.OffshootName())
	secret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: elasticsearch.OffshootLabels(),
		},
		Type: core.SecretTypeOpaque,
		Data: data,
		StringData: map[string]string{
			"key_pass": pass,
		},
	}
	if _, err := c.Client.CoreV1().Secrets(elasticsearch.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
		return nil, err
	}

	secretVolumeSource := &core.SecretVolumeSource{
		SecretName: secret.Name,
	}

	return secretVolumeSource, nil
}

func (c *Controller) findDatabaseSecret(elasticsearch *api.Elasticsearch) (*core.Secret, error) {
	name := fmt.Sprintf("%v-auth", elasticsearch.OffshootName())

	secret, err := c.Client.CoreV1().Secrets(elasticsearch.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	if secret.Labels[api.LabelDatabaseKind] != api.ResourceKindElasticsearch ||
		secret.Labels[api.LabelDatabaseName] != elasticsearch.Name {
		return nil, fmt.Errorf(`intended secret "%v/%v" already exists`, elasticsearch.Namespace, name)
	}

	return secret, nil
}

var action_group = `
UNLIMITED:
  - "*"

READ:
  - "indices:data/read*"
  - "indices:admin/mappings/fields/get*"

CLUSTER_COMPOSITE_OPS_RO:
  - "indices:data/read/mget"
  - "indices:data/read/msearch"
  - "indices:data/read/mtv"
  - "indices:data/read/coordinate-msearch*"
  - "indices:admin/aliases/exists*"
  - "indices:admin/aliases/get*"

CLUSTER_KUBEDB_SNAPSHOT:
  - "indices:data/read/scroll*"
  - "cluster:monitor/main"

INDICES_KUBEDB_SNAPSHOT:
  - "indices:admin/get"
  - "indices:monitor/settings/get"
  - "indices:admin/mappings/get"
`

var action_group_es7 = `
_sg_meta:
  type: "actiongroups"
  config_version: 2

UNLIMITED:
  allowed_actions:
    - "*"

READ:
  allowed_actions:
    - "indices:data/read*"
    - "indices:admin/mappings/fields/get*"

CLUSTER_COMPOSITE_OPS_RO:
  allowed_actions:
    - "indices:data/read/mget"
    - "indices:data/read/msearch"
    - "indices:data/read/mtv"
    - "indices:data/read/coordinate-msearch*"
    - "indices:admin/aliases/exists*"
    - "indices:admin/aliases/get*"

CLUSTER_KUBEDB_SNAPSHOT:
  allowed_actions:
    - "indices:data/read/scroll*"
    - "cluster:monitor/main"

INDICES_KUBEDB_SNAPSHOT:
  allowed_actions:
    - "indices:admin/get"
    - "indices:monitor/settings/get"
    - "indices:admin/mappings/get"
`

var config = `
searchguard:
  dynamic:
    authc:
      basic_internal_auth_domain:
        enabled: true
        order: 4
        http_authenticator:
          type: basic
          challenge: true
        authentication_backend:
          type: internal
`

var config_es7 = `
_sg_meta:
  type: "config"
  config_version: 2
sg_config:
  dynamic:
    authc:
      basic_internal_auth_domain:
        http_enabled: true
        transport_enabled: true
        order: 4
        http_authenticator:
          type: basic
          challenge: true
        authentication_backend:
          type: internal
`

var internal_user = `
admin:
  hash: %s

readall:
  hash: %s
`

var internal_user_es7 = `
_sg_meta:
  type: "internalusers"
  config_version: 2

admin:
  hash: %s

readall:
  hash: %s
`

var roles = `
sg_all_access:
  cluster:
    - UNLIMITED
  indices:
    '*':
      '*':
        - UNLIMITED
  tenants:
    adm_tenant: RW
    test_tenant_ro: RW

sg_readall:
  cluster:
    - CLUSTER_COMPOSITE_OPS_RO
    - CLUSTER_KUBEDB_SNAPSHOT
  indices:
    '*':
      '*':
        - READ
        - INDICES_KUBEDB_SNAPSHOT
`

var roles_es7 = `
_sg_meta:
  type: "roles"
  config_version: 2
sg_all_access:
  cluster_permissions:
  - UNLIMITED
  index_permissions:
  - index_patterns:
    - "*"
    allowed_actions:
    - "UNLIMITED"
  tenant_permissions:
  - tenant_patterns:
    - adm_tenant
    - test_tenant_ro
    allowed_actions:
    - SGS_KIBANA_ALL_WRITE
sg_readall:
  cluster_permissions:
  - "CLUSTER_COMPOSITE_OPS_RO"
  - "CLUSTER_KUBEDB_SNAPSHOT"
  index_permissions:
  - index_patterns:
    - "*"
    allowed_actions:
    - "READ"
    - "INDICES_KUBEDB_SNAPSHOT"
  tenant_permissions: []
`

var roles_mapping = `
sg_all_access:
  users:
    - admin

sg_readall:
  users:
    - readall
`

var roles_mapping_es7 = `
_sg_meta:
  type: "rolesmapping"
  config_version: 2

sg_all_access:
  users:
    - admin

sg_readall:
  users:
    - readall
`

var tenants = `
_sg_meta:
  type: "tenants"
  config_version: 2
test_tenant_ro:
  reserved: false
  hidden: false
  description: "test_tenant_ro. Migrated from v6"
  static: false
adm_tenant:
  reserved: false
  hidden: false
  description: "adm_tenant. Migrated from v6"
  static: false
`

func (c *Controller) createDatabaseSecret(elasticsearch *api.Elasticsearch) (*core.SecretVolumeSource, error) {
	databaseSecret, err := c.findDatabaseSecret(elasticsearch)
	if err != nil {
		return nil, err
	}
	if databaseSecret != nil {
		return &core.SecretVolumeSource{
			SecretName: databaseSecret.Name,
		}, nil
	}

	esVersion, err := c.esVersionLister.Get(string(elasticsearch.Spec.Version))
	if err != nil {
		return nil, err
	}

	var data map[string][]byte

	if esVersion.Spec.AuthPlugin == v1alpha1.ElasticsearchAuthPluginSearchGuard {
		data, err = getSecretDataForSG(esVersion)
		if err != nil {
			return nil, err
		}
	} else if esVersion.Spec.AuthPlugin == v1alpha1.ElasticsearchAuthPluginXpack {
		data = getSecretDataForXPack()
	}

	name := fmt.Sprintf("%v-auth", elasticsearch.OffshootName())
	secret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: elasticsearch.OffshootLabels(),
		},
		Type: core.SecretTypeOpaque,
		Data: data,
	}
	if _, err := c.Client.CoreV1().Secrets(elasticsearch.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
		return nil, err
	}

	return &core.SecretVolumeSource{
		SecretName: secret.Name,
	}, nil
}

func getSecretDataForSG(esVersion *v1alpha1.ElasticsearchVersion) (map[string][]byte, error) {
	adminPassword := rand.Characters(8)
	hashedAdminPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	readallPassword := rand.Characters(8)
	hashedReadallPassword, err := bcrypt.GenerateFromPassword([]byte(readallPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	data := map[string][]byte{
		KeyAdminUserName:        []byte(AdminUser),
		KeyAdminPassword:        []byte(adminPassword),
		KeyReadAllUserName:      []byte(ReadAllUser),
		KeyReadAllPassword:      []byte(readallPassword),
		"sg_action_groups.yml":  []byte(action_group),
		"sg_config.yml":         []byte(config),
		"sg_internal_users.yml": []byte(fmt.Sprintf(internal_user, hashedAdminPassword, hashedReadallPassword)),
		"sg_roles.yml":          []byte(roles),
		"sg_roles_mapping.yml":  []byte(roles_mapping),
	}
	if strings.HasPrefix(esVersion.Spec.Version, "7.") {
		data["sg_action_groups.yml"] = []byte(action_group_es7)
		data["sg_config.yml"] = []byte(config_es7)
		data["sg_internal_users.yml"] = []byte(fmt.Sprintf(internal_user_es7, hashedAdminPassword, hashedReadallPassword))
		data["sg_roles.yml"] = []byte(roles_es7)
		data["sg_roles_mapping.yml"] = []byte(roles_mapping_es7)
		data["sg_tenants.yml"] = []byte(tenants)
	}

	return data, nil
}

func getSecretDataForXPack() map[string][]byte {
	adminPassword := rand.Characters(8)

	return map[string][]byte{
		KeyAdminUserName: []byte(ElasticUser),
		KeyAdminPassword: []byte(adminPassword),
	}
}
