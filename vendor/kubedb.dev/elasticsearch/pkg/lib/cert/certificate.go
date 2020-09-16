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

package cert

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"time"

	"github.com/pkg/errors"
	"gomodules.xyz/cert"
)

const (
	Duration365d = time.Hour * 24 * 365
	CertsDir     = "/tmp/certs/certs"

	KeyEncodingPKCS8  = "pkcs8"
	KeyEncodingPKCS12 = "pkcs12"

	AdminCACert = "admin-ca.crt"
	AdminCAKey  = "admin-ca.key"
	RootCACert  = "root-ca.crt"
	RootCAKey   = "root-ca.key"
	CACert      = "ca.crt"
	TLSCert     = "tls.crt"
	TLSKey      = "tls.key"

	RootKey      = "root-key.pem"
	RootCert     = "root-ca.pem"
	RootKeyStore = "root.jks"
	RootAlias    = "root-ca"

	NodeKey      = "node-key.pem"
	NodeCert     = "node.pem"
	NodePKCS12   = "node.pkcs12"
	NodeKeyStore = "node.jks"
	NodeAlias    = "elasticsearch-node"

	AdminKey  = "admin-key.pem"
	AdminCert = "admin.pem"

	SGAdminKey      = "sgadmin-key.pem"
	SGAdminCert     = "sgadmin.pem"
	SGAdminPKCS12   = "sgadmin.pkcs12"
	SGAdminKeyStore = "sgadmin.jks"
	SGAdminAlias    = "elasticsearch-sgadmin"

	ClientKey      = "client-key.pem"
	ClientCert     = "client.pem"
	ClientPKCS12   = "client.pkcs12"
	ClientKeyStore = "client.jks"
	ClientAlias    = "elasticsearch-client"
)

func ExtractSubjectFromCertificate(crt []byte) (*pkix.Name, error) {
	block, _ := pem.Decode(crt)
	if block == nil || block.Type != cert.CertificateBlockType {
		return nil, errors.New("failed to decode PEM file")
	}
	c, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse the certificate")
	}
	return &c.Subject, nil
}
