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

	"gomodules.xyz/version"
)

func (c *Controller) getTLSArgs(mongoDB *api.MongoDB, mgVersion *v1alpha1.MongoDBVersion) ([]string, error) {
	var sslArgs []string
	sslMode := string(mongoDB.Spec.SSLMode)
	breakingVer, err := version.NewVersion("4.2")
	if err != nil {
		return nil, err
	}
	currentVer, err := version.NewVersion(mgVersion.Spec.Version)
	if err != nil {
		return nil, err
	}

	//xREF: https://github.com/docker-library/mongo/issues/367
	if currentVer.GreaterThanOrEqual(breakingVer) {
		var tlsMode = sslMode
		if strings.Contains(sslMode, "SSL") {
			tlsMode = strings.Replace(sslMode, "SSL", "TLS", 1)
		} //ie. requireSSL => requireTLS

		sslArgs = []string{
			fmt.Sprintf("--tlsMode=%v", tlsMode),
		}

		if mongoDB.Spec.SSLMode != api.SSLModeDisabled {
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
		if mongoDB.Spec.SSLMode != api.SSLModeDisabled {
			sslArgs = append(sslArgs, []string{
				fmt.Sprintf("--sslCAFile=%v/%v", api.MongoCertDirectory, api.TLSCACertFileName),
				fmt.Sprintf("--sslPEMKeyFile=%v/%v", api.MongoCertDirectory, api.MongoPemFileName),
			}...)
		}
	}

	return sslArgs, nil
}
