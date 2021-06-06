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
	"fmt"
	"strings"

	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"github.com/Masterminds/semver/v3"
	core "k8s.io/api/core/v1"
	dynamic_util "kmodules.xyz/client-go/dynamic"
)

func (c *Reconciler) getTLSArgs(db *api.MongoDB, mgVersion *v1alpha1.MongoDBVersion) ([]string, error) {
	var sslArgs []string
	sslMode := string(db.Spec.SSLMode)
	breakingVer, err := semver.NewVersion("4.2")
	if err != nil {
		return nil, err
	}
	currentVer, err := semver.NewVersion(mgVersion.Spec.Version)
	if err != nil {
		return nil, err
	}

	//xREF: https://github.com/docker-library/mongo/issues/367
	if currentVer.Compare(breakingVer) >= 0 {
		var tlsMode = sslMode
		if strings.Contains(sslMode, "SSL") {
			tlsMode = strings.Replace(sslMode, "SSL", "TLS", 1)
		} //ie. requireSSL => requireTLS

		sslArgs = []string{
			fmt.Sprintf("--tlsMode=%v", tlsMode),
		}

		if db.Spec.SSLMode != api.SSLModeDisabled {
			//xREF: https://github.com/docker-library/mongo/issues/367
			sslArgs = append(sslArgs, []string{
				fmt.Sprintf("--tlsCAFile=%v/%v", api.MongoCertDirectory, api.TLSCACertFileName),
				fmt.Sprintf("--tlsCertificateKeyFile=%v/%v", api.MongoCertDirectory, api.MongoPemFileName),
			}...)
		}
	} else {
		sslArgs = []string{
			fmt.Sprintf("--sslMode=%v", sslMode),
		}
		if db.Spec.SSLMode != api.SSLModeDisabled {
			sslArgs = append(sslArgs, []string{
				fmt.Sprintf("--sslCAFile=%v/%v", api.MongoCertDirectory, api.TLSCACertFileName),
				fmt.Sprintf("--sslPEMKeyFile=%v/%v", api.MongoCertDirectory, api.MongoPemFileName),
			}...)
		}
	}

	return sslArgs, nil
}

func (c *Reconciler) IsCertificateSecretsCreated(db *api.MongoDB) (bool, error) {
	// wait for certificates
	if db.Spec.TLS != nil {
		var secrets []string
		if db.Spec.ShardTopology != nil {
			// for config server
			secrets = append(secrets, db.GetCertSecretName(api.MongoDBServerCert, db.ConfigSvrNodeName()))
			// for shards
			for i := 0; i < int(db.Spec.ShardTopology.Shard.Shards); i++ {
				secrets = append(secrets, db.GetCertSecretName(api.MongoDBServerCert, db.ShardNodeName(int32(i))))
			}
			// for mongos
			secrets = append(secrets, db.GetCertSecretName(api.MongoDBServerCert, db.MongosNodeName()))
		} else {
			// ReplicaSet or Standalone
			secrets = append(secrets, db.GetCertSecretName(api.MongoDBServerCert, ""))
		}
		// for stash/user
		secrets = append(secrets, db.GetCertSecretName(api.MongoDBClientCert, ""))
		// for prometheus exporter
		secrets = append(secrets, db.GetCertSecretName(api.MongoDBMetricsExporterCert, ""))

		return dynamic_util.ResourcesExists(
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			db.Namespace,
			secrets...,
		)
	}

	return true, nil
}
