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

package pkcs8

import (
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math"
	"math/big"
	"net"
	"path/filepath"
	"time"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	certlib "kubedb.dev/elasticsearch/pkg/lib/cert"

	"github.com/pkg/errors"
	"gomodules.xyz/cert"
	"gomodules.xyz/x/ioutil"
	"kmodules.xyz/client-go/meta"
)

// Creates pkcs8 encoded certificates in pem format.
// Generated secret contains:
// 	- tls.crt: ca.crt (CA: true)
//	- tls.key: ca.key
func CreateCaCertificate(commonName string, certPath string) (*rsa.PrivateKey, *x509.Certificate, error) {
	cfg := cert.Config{
		CommonName:   commonName,
		Organization: []string{kubedb.GroupName},
	}

	caKey, err := cert.NewPrivateKey()
	if err != nil {
		return nil, nil, errors.New("failed to generate key for CA certificate")
	}

	caCert, err := cert.NewSelfSignedCACert(cfg, caKey)
	if err != nil {
		return nil, nil, errors.New("failed to generate CA certificate")
	}

	caKeyByte, err := cert.EncodePKCS8PrivateKeyPEM(caKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to encode private key")
	}

	if !ioutil.WriteString(filepath.Join(certPath, certlib.CAKey), string(caKeyByte)) {
		return nil, nil, errors.New("failed to write key for CA certificate")
	}

	caCertByte := cert.EncodeCertPEM(caCert)
	if !ioutil.WriteString(filepath.Join(certPath, certlib.CACert), string(caCertByte)) {
		return nil, nil, errors.New("failed to write CA certificate")
	}

	return caKey, caCert, nil
}

// Creates pkcs8 encoded certificates in pem format singed by given CA ( ca.crt, ca.key )
// Generated secret contains:
//	- ca.crt : ca.crt
//	- tls.crt: transport-layer.crt ( signed by ca.crt)
// 	- tls.key: transport-layer.key
func CreateTransportCertificate(certPath string, db *api.Elasticsearch, caKey *rsa.PrivateKey, caCert *x509.Certificate) error {
	cfg := cert.Config{
		CommonName:   db.ClientCertificateCN(api.ElasticsearchTransportCert),
		Organization: []string{kubedb.GroupName},
		Usages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		AltNames: cert.AltNames{
			DNSNames: []string{
				fmt.Sprintf("%v.%v.svc", db.GoverningServiceName(), db.Namespace),
				fmt.Sprintf("*.%v.%v.svc", db.GoverningServiceName(), db.Namespace),
				fmt.Sprintf("*.%v.%v.svc.%v", db.GoverningServiceName(), db.Namespace, meta.ClusterDomain()),
				api.LocalHost,
			},
			IPs: []net.IP{
				net.ParseIP(api.LocalHostIP),
			},
		},
	}

	nodePrivateKey, err := cert.NewPrivateKey()
	if err != nil {
		return errors.New("failed to generate key for node certificate")
	}

	nodeCertificate, err := NewSignedCert(cfg, nodePrivateKey, caCert, caKey)
	if err != nil {
		return errors.New("failed to sign node certificate")
	}

	nodeKeyByte, err := cert.EncodePKCS8PrivateKeyPEM(nodePrivateKey)
	if err != nil {
		return errors.Wrap(err, "failed to encode node private key")
	}

	if !ioutil.WriteString(filepath.Join(certPath, certlib.TLSKey), string(nodeKeyByte)) {
		return errors.New("failed to write key for transport certificate")
	}

	nodeCertByte := cert.EncodeCertPEM(nodeCertificate)
	if !ioutil.WriteString(filepath.Join(certPath, certlib.TLSCert), string(nodeCertByte)) {
		return errors.New("failed to write transport certificate")
	}

	return nil
}

// Creates pkcs8 encoded certificates in pem format singed by given CA ( ca.crt, ca.key )
// Generated secret contains:
//	- ca.crt : ca.crt
//	- tls.crt: http-layer.crt ( signed by ca.crt)
// 	- tls.key: http-layer.key
func CreateHTTPCertificate(certPath string, db *api.Elasticsearch, caKey *rsa.PrivateKey, caCert *x509.Certificate) error {
	cfg := cert.Config{
		CommonName:   db.ClientCertificateCN(api.ElasticsearchHTTPCert),
		Organization: []string{kubedb.GroupName},
		AltNames: cert.AltNames{
			DNSNames: []string{
				fmt.Sprintf("%v.%v.svc", db.OffshootName(), db.Namespace),
				fmt.Sprintf("%v.%v.svc.%v", db.OffshootName(), db.Namespace, meta.ClusterDomain()),
				fmt.Sprintf("%v.%v.svc", db.GoverningServiceName(), db.Namespace),
				fmt.Sprintf("*.%v.%v.svc", db.GoverningServiceName(), db.Namespace),
				fmt.Sprintf("*.%v.%v.svc.%v", db.GoverningServiceName(), db.Namespace, meta.ClusterDomain()),
				api.LocalHost,
			},
			IPs: []net.IP{
				net.ParseIP(api.LocalHostIP),
			},
		},
		Usages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
	}

	clientPrivateKey, err := cert.NewPrivateKey()
	if err != nil {
		return errors.New("failed to generate key for client certificate")
	}

	clientCertificate, err := cert.NewSignedCert(cfg, clientPrivateKey, caCert, caKey)
	if err != nil {
		return errors.New("failed to sign client certificate")
	}

	adminKeyByte, err := cert.EncodePKCS8PrivateKeyPEM(clientPrivateKey)
	if err != nil {
		return errors.Wrap(err, "failed to encode client private key")
	}

	if !ioutil.WriteString(filepath.Join(certPath, certlib.TLSKey), string(adminKeyByte)) {
		return errors.New("failed to write key for client certificate")
	}

	adminCertByte := cert.EncodeCertPEM(clientCertificate)
	if !ioutil.WriteString(filepath.Join(certPath, certlib.TLSCert), string(adminCertByte)) {
		return errors.New("failed to write client certificate")
	}

	return nil
}

// Creates pkcs8 encoded certificates in pem format singed by given CA ( ca.crt, ca.key )
// Generated secret contains:
//	- ca.crt : ca.crt
//	- tls.crt: client.crt ( signed by ca.crt)
// 	- tls.key: client.key
func CreateClientCertificate(alias string, certPath string, db *api.Elasticsearch, caKey *rsa.PrivateKey, caCert *x509.Certificate) error {
	cfg := cert.Config{
		CommonName:   db.ClientCertificateCN(api.ElasticsearchCertificateAlias(alias)),
		Organization: []string{kubedb.GroupName},
		Usages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
		},
	}

	clientPrivateKey, err := cert.NewPrivateKey()
	if err != nil {
		return errors.New(fmt.Sprintf("failed to generate privateKey for: %s", alias))
	}

	clientCertificate, err := cert.NewSignedCert(cfg, clientPrivateKey, caCert, caKey)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to sign certificate for: %s", alias))
	}

	keyByte, err := cert.EncodePKCS8PrivateKeyPEM(clientPrivateKey)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to encode private key for: %s", alias))
	}

	if !ioutil.WriteString(filepath.Join(certPath, certlib.TLSKey), string(keyByte)) {
		return errors.New(fmt.Sprintf("failed to write tls.key for: %s", alias))
	}

	certByte := cert.EncodeCertPEM(clientCertificate)
	if !ioutil.WriteString(filepath.Join(certPath, certlib.TLSCert), string(certByte)) {
		return errors.New(fmt.Sprintf("failed to write tls.crt for: %s", alias))
	}

	return nil
}

// NewSignedCert creates a signed certificate using the given CA certificate and key
func NewSignedCert(cfg cert.Config, key *rsa.PrivateKey, caCert *x509.Certificate, caKey *rsa.PrivateKey) (*x509.Certificate, error) {
	serial, err := cryptorand.Int(cryptorand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, err
	}
	if len(cfg.CommonName) == 0 {
		return nil, errors.New("must specify a CommonName")
	}
	if len(cfg.Usages) == 0 {
		return nil, errors.New("must specify at least one ExtKeyUsage")
	}

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames:     cfg.AltNames.DNSNames,
		IPAddresses:  cfg.AltNames.IPs,
		SerialNumber: serial,
		NotBefore:    caCert.NotBefore,
		NotAfter:     time.Now().Add(certlib.Duration365d).UTC(),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  cfg.Usages,
		ExtraExtensions: []pkix.Extension{
			{
				Id: oidExtensionSubjectAltName,
			},
		},
	}
	certTmpl.ExtraExtensions[0].Value, err = marshalSANs(cfg.AltNames.DNSNames, nil, cfg.AltNames.IPs)
	if err != nil {
		return nil, err
	}

	certDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(certDERBytes)
}

var oidExtensionSubjectAltName = []int{2, 5, 29, 17}

// marshalSANs marshals a list of addresses into a the contents of an X.509
// SubjectAlternativeName extension.
func marshalSANs(dnsNames, emailAddresses []string, ipAddresses []net.IP) (derBytes []byte, err error) {
	var rawValues []asn1.RawValue
	for _, name := range dnsNames {
		rawValues = append(rawValues, asn1.RawValue{Tag: 2, Class: 2, Bytes: []byte(name)})
	}
	for _, email := range emailAddresses {
		rawValues = append(rawValues, asn1.RawValue{Tag: 1, Class: 2, Bytes: []byte(email)})
	}
	for _, rawIP := range ipAddresses {
		// If possible, we always want to encode IPv4 addresses in 4 bytes.
		ip := rawIP.To4()
		if ip == nil {
			ip = rawIP
		}
		rawValues = append(rawValues, asn1.RawValue{Tag: 7, Class: 2, Bytes: ip})
	}
	// https://github.com/floragunncom/search-guard-docs/blob/master/tls_certificates_production.md#using-an-oid-value-as-san-entry
	// https://github.com/floragunncom/search-guard-ssl/blob/a2d1e8e9b25a10ecaf1cb47e48cf04328af7d7fb/example-pki-scripts/gen_node_cert.sh#L55
	// Adds AltName: OID: 1.2.3.4.5.5
	// ref: https://stackoverflow.com/a/47917273/244009
	rawValues = append(rawValues, asn1.RawValue{FullBytes: []byte{0x88, 0x05, 0x2A, 0x03, 0x04, 0x05, 0x05}})
	return asn1.Marshal(rawValues)
}
