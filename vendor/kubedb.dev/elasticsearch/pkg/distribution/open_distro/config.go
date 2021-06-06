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
	certlib "kubedb.dev/elasticsearch/pkg/lib/cert"
	"kubedb.dev/elasticsearch/pkg/lib/user"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_util "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

const (
	ConfigFileName       = "elasticsearch.yml"
	InternalUserFileName = "internal_users.yml"
	RolesMappingFileName = "roles_mapping.yml"
)

var adminDNTemplate = `
opendistro_security.authcz.admin_dn:
- "%s"
`

var nodesDNTemplate = `
opendistro_security.nodes_dn:
- "%s"
`

var opendistro_security_enabled = `
opendistro_security.ssl.transport.pemcert_filepath: certs/transport/tls.crt
opendistro_security.ssl.transport.pemkey_filepath: certs/transport/tls.key
opendistro_security.ssl.transport.pemtrustedcas_filepath: certs/transport/ca.crt
opendistro_security.ssl.transport.enforce_hostname_verification: false

# opendistro_security.nodes_dn:
%s

opendistro_security.allow_default_init_securityindex: true
opendistro_security.audit.type: internal_elasticsearch
opendistro_security.enable_snapshot_restore_privilege: true
opendistro_security.check_snapshot_restore_write_privileges: true
opendistro_security.restapi.roles_enabled: ["all_access", "security_rest_api_access"]
cluster.routing.allocation.disk.threshold_enabled: false
node.max_local_storage_nodes: 3
`

var opendistro_security_disabled = `
opendistro_security.disabled: true

cluster.routing.allocation.disk.threshold_enabled: false
node.max_local_storage_nodes: 3
`

var https_enabled = `
opendistro_security.ssl.http.enabled: true
opendistro_security.ssl.http.pemcert_filepath: certs/http/tls.crt
opendistro_security.ssl.http.pemkey_filepath: certs/http/tls.key
opendistro_security.ssl.http.pemtrustedcas_filepath: certs/http/ca.crt

# opendistro_security.authcz.admin_dn:
%s
`

var https_disabled = `
opendistro_security.ssl.http.enabled: false
`

func (es *Elasticsearch) EnsureDefaultConfig() error {
	secret, err := es.findSecret(es.db.ConfigSecretName())
	if err != nil {
		return err
	}

	// If the secret already exists
	// and not controlled by the ES object.
	// i.e. user provided custom secret file.
	if secret != nil && !metav1.IsControlledBy(secret, es.db) {
		// check whether it contains "elasticsearch.yml" file or not.
		if value, ok := secret.Data[ConfigFileName]; !ok || len(value) == 0 {
			return errors.New("elasticsearch.yml is missing")
		}
		return nil
	}

	// config secret isn't created yet.
	// let's create it.
	owner := metav1.NewControllerRef(es.db, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))
	secretMeta := metav1.ObjectMeta{
		Name:      es.db.ConfigSecretName(),
		Namespace: es.db.Namespace,
	}

	var isEqualInUserConfig bool
	var config, inUserConfig, rolesMapping string

	if !es.db.Spec.DisableSecurity {
		if es.db.Spec.TLS == nil {
			return errors.New("spec.TLS configuration is empty")
		}

		if secret != nil {
			if value, ok := secret.Data[InternalUserFileName]; ok {
				oldInUsers, err := user.ParseInUserConfig(string(value))
				if err != nil {
					return err
				}
				// Every time we create internalUser
				// It varies, even if the userSpec is same.
				// It is because of the bcrypt hash generator.
				// Let's check, whether the userSpec is changed or not.
				isEqualInUserConfig, err = user.InUserConfigCompareEqual(es.db.Spec.InternalUsers, oldInUsers)
				if err != nil {
					return errors.Wrap(err, "failed to compare internal user config file")
				}
			}
		}

		// If not internal user spec has changed,
		// generate internal user config file for the default users: admin, kibanaserver, etc.
		if !isEqualInUserConfig {
			inUserConfig, err = es.getInternalUserConfig()
			if err != nil {
				return errors.Wrap(err, "failed to generate default internal_users.yml")
			}
		}

		rolesMapping, err = es.getRolesMapping()
		if err != nil {
			return errors.Wrap(err, "failed to generate default roles_mapping.yml")
		}

		// Get transport layer cert secret.
		// Transport layer is always secured with certificate.
		// Parse the tls.cert to extract the nodeDNs.
		sName, exist := api_util.GetCertificateSecretName(es.db.Spec.TLS.Certificates, string(api.ElasticsearchTransportCert))
		if !exist {
			return errors.New("transport-cert secret is missing")
		}
		cSecret, err := es.kClient.CoreV1().Secrets(es.db.Namespace).Get(context.TODO(), sName, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to get certificateSecret: %s/%s", es.db.Namespace, sName))
		}
		nodesDN := ""
		if value, ok := cSecret.Data[certlib.TLSCert]; ok {
			subj, err := certlib.ExtractSubjectFromCertificate(value)
			if err != nil {
				return err
			}
			nodesDN = fmt.Sprintf(nodesDNTemplate, subj.String())
		}

		config = fmt.Sprintf(opendistro_security_enabled, nodesDN)

		// If rest layer is secured with certs
		// Calculate the adminDN
		if es.db.Spec.EnableSSL {
			// Get openDistro admin cert secret.
			// Parse the tls.cert to extract the adminDNs.
			sName, exist = api_util.GetCertificateSecretName(es.db.Spec.TLS.Certificates, string(api.ElasticsearchAdminCert))
			if !exist {
				return errors.New("admin-cert secret is missing")
			}

			cSecret, err = es.kClient.CoreV1().Secrets(es.db.Namespace).Get(context.TODO(), sName, metav1.GetOptions{})
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to get certificateSecret: %s/%s", es.db.Namespace, sName))
			}

			adminDN := ""
			if value, ok := cSecret.Data[certlib.TLSCert]; ok {
				subj, err := certlib.ExtractSubjectFromCertificate(value)
				if err != nil {
					return err
				}
				adminDN = fmt.Sprintf(adminDNTemplate, subj.String())
			}

			config += fmt.Sprintf(https_enabled, adminDN)
		} else {
			config += https_disabled
		}

	} else {
		config = opendistro_security_disabled
	}

	if _, _, err := core_util.CreateOrPatchSecret(context.TODO(), es.kClient, secretMeta, func(in *core.Secret) *core.Secret {
		in.Labels = core_util.UpsertMap(in.Labels, es.db.OffshootLabels())
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		if in.Data == nil {
			in.Data = make(map[string][]byte)
		}
		in.Data[ConfigFileName] = []byte(config)
		in.Data[RolesMappingFileName] = []byte(rolesMapping)
		if !isEqualInUserConfig {
			in.Data[InternalUserFileName] = []byte(inUserConfig)
		}
		return in
	}, metav1.PatchOptions{}); err != nil {
		return err
	}

	return nil
}

func (es *Elasticsearch) getInternalUserConfig() (string, error) {
	userList := es.db.Spec.InternalUsers
	if userList == nil {
		return "", errors.New("spec.internalUsers is empty")
	}

	for username := range userList {
		var pass string
		var err error

		secretName, err := es.db.GetUserCredSecretName(username)
		if err != nil {
			return "", err
		}
		pass, err = es.getPasswordFromSecret(secretName)
		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("failed to get password from secret: %s/%s", es.db.Namespace, secretName))
		}

		err = user.SetPasswordHashForUser(userList, username, pass)
		if err != nil {
			return "", errors.Wrap(err, "failed to generate the password hash")
		}
	}

	byt, err := yaml.Marshal(userList)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal the internal user list")
	}

	return string(byt), nil
}

func (es *Elasticsearch) getRolesMapping() (string, error) {
	rolesMapping := es.db.Spec.RolesMapping
	// if rolesMapping is nil, return empty string
	// no need to  perform yaml.Marshal().
	// coz it will generate ( `{}` ).
	if rolesMapping == nil {
		return "", nil
	}

	byt, err := yaml.Marshal(rolesMapping)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal the roles mapping")
	}

	return string(byt), nil
}

func (es *Elasticsearch) getPasswordFromSecret(sName string) (string, error) {
	secret, err := es.kClient.CoreV1().Secrets(es.db.Namespace).Get(context.TODO(), sName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if value, exist := secret.Data[core.BasicAuthPasswordKey]; exist && len(value) != 0 {
		return string(value), nil
	}

	return "", errors.New(fmt.Sprintf("password is missing in secret: %s/%s", es.db.Namespace, sName))
}
