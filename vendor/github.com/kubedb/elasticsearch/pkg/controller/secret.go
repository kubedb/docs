package controller

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/appscode/go/crypto/rand"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"golang.org/x/crypto/bcrypt"
	"gomodules.xyz/cert"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

const (
	AdminUser          = "admin"
	KeyAdminUserName   = "ADMIN_USERNAME"
	KeyAdminPassword   = "ADMIN_PASSWORD"
	ReadAllUser        = "readall"
	KeyReadAllUserName = "READALL_USERNAME"
	KeyReadAllPassword = "READALL_PASSWORD"
	ExporterSecretPath = "/var/run/secrets/kubedb.com/"
)

func (c *Controller) ensureCertSecret(elasticsearch *api.Elasticsearch) error {
	certSecretVolumeSource := elasticsearch.Spec.CertificateSecret
	if certSecretVolumeSource == nil {
		var err error
		if certSecretVolumeSource, err = c.createCertSecret(elasticsearch); err != nil {
			return err
		}
		es, _, err := util.PatchElasticsearch(c.ExtClient.KubedbV1alpha1(), elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
			in.Spec.CertificateSecret = certSecretVolumeSource
			return in
		})
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
		es, _, err := util.PatchElasticsearch(c.ExtClient.KubedbV1alpha1(), elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
			in.Spec.DatabaseSecret = databaseSecretVolume
			return in
		})
		if err != nil {
			return err
		}
		elasticsearch.Spec.DatabaseSecret = es.Spec.DatabaseSecret
		return nil
	}
	return c.upgradeDatabaseSecret(elasticsearch)
}

func (c *Controller) findCertSecret(elasticsearch *api.Elasticsearch) (*core.Secret, error) {
	name := fmt.Sprintf("%v-cert", elasticsearch.OffshootName())

	secret, err := c.Client.CoreV1().Secrets(elasticsearch.Namespace).Get(name, metav1.GetOptions{})
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

	certPath := fmt.Sprintf("%v/%v", certsDir, rand.Characters(3))
	os.Mkdir(certPath, os.ModePerm)

	caKey, caCert, pass, err := createCaCertificate(certPath)
	if err != nil {
		return nil, err
	}
	err = createNodeCertificate(certPath, elasticsearch, caKey, caCert, pass)
	if err != nil {
		return nil, err
	}
	err = createAdminCertificate(certPath, caKey, caCert, pass)
	if err != nil {
		return nil, err
	}
	root, err := ioutil.ReadFile(filepath.Join(certPath, rootKeyStore))
	if err != nil {
		return nil, err
	}
	node, err := ioutil.ReadFile(filepath.Join(certPath, nodeKeyStore))
	if err != nil {
		return nil, err
	}
	sgadmin, err := ioutil.ReadFile(filepath.Join(certPath, sgAdminKeyStore))
	if err != nil {
		return nil, err
	}

	data := map[string][]byte{
		rootKeyStore:    root,
		nodeKeyStore:    node,
		sgAdminKeyStore: sgadmin,
	}

	if elasticsearch.Spec.EnableSSL {
		if err := createClientCertificate(certPath, elasticsearch, caKey, caCert, pass); err != nil {
			return nil, err
		}

		client, err := ioutil.ReadFile(filepath.Join(certPath, clientKeyStore))
		if err != nil {
			return nil, err
		}

		data[rootCert] = cert.EncodeCertPEM(caCert)
		data[clientKeyStore] = client
	}

	name := fmt.Sprintf("%v-cert", elasticsearch.OffshootName())
	secret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindElasticsearch,
				api.LabelDatabaseName: elasticsearch.Name,
			},
		},
		Type: core.SecretTypeOpaque,
		Data: data,
		StringData: map[string]string{
			"key_pass": pass,
		},
	}
	if _, err := c.Client.CoreV1().Secrets(elasticsearch.Namespace).Create(secret); err != nil {
		return nil, err
	}

	secretVolumeSource := &core.SecretVolumeSource{
		SecretName: secret.Name,
	}

	return secretVolumeSource, nil
}

func (c *Controller) findDatabaseSecret(elasticsearch *api.Elasticsearch) (*core.Secret, error) {
	name := fmt.Sprintf("%v-auth", elasticsearch.OffshootName())

	secret, err := c.Client.CoreV1().Secrets(elasticsearch.Namespace).Get(name, metav1.GetOptions{})
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

var internal_user = `
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

var roles_mapping = `
sg_all_access:
  users:
    - admin

sg_readall:
  users:
    - readall
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

	name := fmt.Sprintf("%v-auth", elasticsearch.OffshootName())
	secret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindElasticsearch,
				api.LabelDatabaseName: elasticsearch.Name,
			},
		},
		Type: core.SecretTypeOpaque,
		Data: data,
	}
	if _, err := c.Client.CoreV1().Secrets(elasticsearch.Namespace).Create(secret); err != nil {
		return nil, err
	}

	return &core.SecretVolumeSource{
		SecretName: secret.Name,
	}, nil
}

// This is done to fix 0.8.0 -> 0.9.0 upgrade due to
// https://github.com/kubedb/elasticsearch/pull/181/files#diff-10ddaf307bbebafda149db10a28b9c24R23 commit
func (c *Controller) upgradeDatabaseSecret(elasticsearch *api.Elasticsearch) error {
	meta := metav1.ObjectMeta{
		Name:      elasticsearch.Spec.DatabaseSecret.SecretName,
		Namespace: elasticsearch.Namespace,
	}

	_, _, err := core_util.CreateOrPatchSecret(c.Client, meta, func(in *core.Secret) *core.Secret {
		in.StringData = make(map[string]string)
		if _, ok := in.Data[KeyAdminUserName]; !ok {
			in.StringData[KeyAdminUserName] = AdminUser
		}
		if _, ok := in.Data[KeyReadAllUserName]; !ok {
			in.StringData[KeyReadAllUserName] = ReadAllUser
		}
		return in
	})
	return err
}
