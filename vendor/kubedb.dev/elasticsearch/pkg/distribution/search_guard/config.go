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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	certlib "kubedb.dev/elasticsearch/pkg/lib/cert"
	"kubedb.dev/elasticsearch/pkg/lib/user"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_util "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

const (
	ConfigFileName       = "elasticsearch.yml"
	InternalUserFileName = "sg_internal_users.yml"
	RolesMappingFileName = "sg_roles_mapping.yml"
)

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

	config, err := es.GetElasticsearchConfig()
	if err != nil {
		return err
	}

	var isEqualInUserConfig bool
	var inUserConfig, rolesMapping string
	if !es.db.Spec.DisableSecurity {
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
	}

	if _, _, err := core_util.CreateOrPatchSecret(context.TODO(), es.kClient, secretMeta, func(in *core.Secret) *core.Secret {
		in.Labels = meta_util.OverwriteKeys(in.Labels, es.db.OffshootLabels())
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

func (es *Elasticsearch) GetElasticsearchConfig() (string, error) {
	dbVersion, err := semver.NewVersion(es.esVersion.Spec.Version)
	if err != nil {
		return "", err
	}

	config := make(map[string]interface{})
	// Basic configs
	if dbVersion.Major() >= 6 {
		config["searchguard.enterprise_modules_enabled"] = false
	} else {
		config["cluster.name"] = `${cluster.name}`
		config["node.name"] = `${node.name}`
		config["network.host"] = `${network.host}`
		config["node.ingest"] = `${node.ingest}`
		config["node.master"] = `${node.master}`
		config["node.data"] = `${node.data}`
		config["discovery.zen.minimum_master_nodes"] = `${discovery.zen.minimum_master_nodes}`
		config["discovery.zen.ping.unicast.hosts"] = `${discovery.zen.ping.unicast.hosts}`
	}
	// For Elasticsearch version >= 7.9.x
	// The legacy node role setting (ie. node.master=true) is deprecated,
	// So, for newer versions, node roles will be managed by elasticsearch.yml file.
	if dbVersion.Major() > 7 || (dbVersion.Major() == 7 && dbVersion.Minor() >= 9) {
		config["node.roles"] = `${NODE_ROLES}`
	}

	// X-Pack
	// For Elasticsearch version >= 7.11.x
	// The searchGuard no longer use the oss version Elasticsearch.
	// It uses the official image which comes with xpack plugin pre-installed.
	// Disable the xpack plugin.
	if dbVersion.Major() > 7 || (dbVersion.Major() == 7 && dbVersion.Minor() >= 11) {
		config["xpack.security.enabled"] = false
	}

	if es.db.Spec.DisableSecurity {
		config["searchguard.disabled"] = true
	} else {
		// If security is enabled, the transport layer must be secured with tls
		if es.db.Spec.TLS == nil {
			return "", errors.New("spec.TLS configuration is empty")
		}

		// Transport layer
		config["searchguard.ssl.transport.enforce_hostname_verification"] = false
		config["searchguard.ssl.transport.pemkey_filepath"] = "certs/transport/tls.key"
		config["searchguard.ssl.transport.pemcert_filepath"] = "certs/transport/tls.crt"
		config["searchguard.ssl.transport.pemtrustedcas_filepath"] = "certs/transport/ca.crt"

		if es.db.Spec.EnableSSL {
			// Rest layer
			config["searchguard.ssl.http.enabled"] = true
			config["searchguard.ssl.http.pemkey_filepath"] = "certs/http/tls.key"
			config["searchguard.ssl.http.pemcert_filepath"] = "certs/http/tls.crt"
			config["searchguard.ssl.http.pemtrustedcas_filepath"] = "certs/http/ca.crt"
		} else {
			config["searchguard.ssl.http.enabled"] = false
		}

		// Admin DN
		// To load configuration changes to the security plugin, you must provide your admin certificate to the tool.
		// So we must create the admin certificate even if the HTTP layer (db.Spec.EnableSSL=false) is disabled.
		// Which will provide us the opportunity to run sgadmin.sh command.
		// Let's calculate the adminDN from the admin certificate.
		// Parse the tls.cert to extract the adminDNs.
		sName, exist := api_util.GetCertificateSecretName(es.db.Spec.TLS.Certificates, string(api.ElasticsearchAdminCert))
		if !exist {
			return "", errors.New(fmt.Sprintf("admin-cert: %s secret is missing", sName))
		}
		cSecret, err := es.kClient.CoreV1().Secrets(es.db.Namespace).Get(context.TODO(), sName, metav1.GetOptions{})
		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("failed to get certificateSecret: %s/%s", es.db.Namespace, sName))
		}
		if value, ok := cSecret.Data[certlib.TLSCert]; ok {
			subj, err := certlib.ExtractSubjectFromCertificate(value)
			if err != nil {
				return "", err
			}
			config["searchguard.authcz.admin_dn"] = []string{subj.String()}
		}

		// Node DN
		// Get transport layer cert secret.
		// Parse the tls.cert to extract the nodeDNs.
		sName, exist = api_util.GetCertificateSecretName(es.db.Spec.TLS.Certificates, string(api.ElasticsearchTransportCert))
		if !exist {
			return "", errors.New(fmt.Sprintf("transport-cert: %s secret is missing", sName))
		}
		cSecret, err = es.kClient.CoreV1().Secrets(es.db.Namespace).Get(context.TODO(), sName, metav1.GetOptions{})
		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("failed to get certificateSecret: %s/%s", es.db.Namespace, sName))
		}
		if value, ok := cSecret.Data[certlib.TLSCert]; ok {
			subj, err := certlib.ExtractSubjectFromCertificate(value)
			if err != nil {
				return "", err
			}
			config["searchguard.nodes_dn"] = []string{subj.String()}
		}

		// Additional configs
		config["searchguard.audit.type"] = "internal_elasticsearch"
		config["searchguard.enable_snapshot_restore_privilege"] = true
		if dbVersion.Major() >= 6 {
			config["searchguard.allow_default_init_sgindex"] = true
			config["searchguard.allow_unsafe_democertificates"] = true
			config["searchguard.check_snapshot_restore_write_privileges"] = true
			config["searchguard.restapi.roles_enabled"] = []string{"SGS_ALL_ACCESS", "sg_all_access"}
		}
	}

	configByt, err := yaml.Marshal(config)
	return string(configByt), err
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
