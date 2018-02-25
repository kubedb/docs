package controller

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
	"os/exec"
	"time"

	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/ioutil"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/cert"
)

var certsDir = "/tmp/certs/certs"

func createCaCertificate(certPath string) (*rsa.PrivateKey, *x509.Certificate, string, error) {
	cfg := cert.Config{
		CommonName:   "KubeDB Com. Root CA",
		Organization: []string{"Elasticsearch Operator"},
	}

	caKey, err := cert.NewPrivateKey()
	if err != nil {
		return nil, nil, "", errors.New("failed to generate key for CA certificate")
	}

	caCert, err := cert.NewSelfSignedCACert(cfg, caKey)
	if err != nil {
		return nil, nil, "", errors.New("failed to generate CA certificate")
	}

	nodeKeyByte := cert.EncodePrivateKeyPEM(caKey)
	if !ioutil.WriteString(fmt.Sprintf("%s/root-key.pem", certPath), string(nodeKeyByte)) {
		return nil, nil, "", errors.New("failed to write key for CA certificate")
	}
	caCertByte := cert.EncodeCertPEM(caCert)
	if !ioutil.WriteString(fmt.Sprintf("%s/root.pem", certPath), string(caCertByte)) {
		return nil, nil, "", errors.New("failed to write CA certificate")
	}

	pass := rand.Characters(6)

	_, err = exec.Command(
		"keytool",
		"-import",
		"-file", fmt.Sprintf("%s/root.pem", certPath),
		"-alias", "root-ca",
		"-keystore", fmt.Sprintf("%s/root.jks", certPath),
		"-storepass", pass,
		"-srcstoretype", "pkcs12",
		"-noprompt",
	).Output()
	if err != nil {
		return nil, nil, "", errors.New("failed to generate root.pk12")
	}

	return caKey, caCert, pass, nil
}

func createNodeCertificate(certPath string, elasticsearch *api.Elasticsearch, caKey *rsa.PrivateKey, caCert *x509.Certificate, pass string) error {
	cfg := cert.Config{
		CommonName:   elasticsearch.OffshootName(),
		Organization: []string{"Elasticsearch Operator"},
		Usages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
	}

	nodeKey, err := cert.NewPrivateKey()
	if err != nil {
		return errors.New("failed to generate key for node certificate")
	}
	nodeCert, err := NewSignedCert(cfg, nodeKey, caCert, caKey)
	if err != nil {
		return errors.New("failed to sign node certificate")
	}

	nodeKeyByte := cert.EncodePrivateKeyPEM(nodeKey)
	if !ioutil.WriteString(fmt.Sprintf("%s/node-key.pem", certPath), string(nodeKeyByte)) {
		return errors.New("failed to write key for node certificate")
	}
	nodeCertByte := cert.EncodeCertPEM(nodeCert)
	if !ioutil.WriteString(fmt.Sprintf("%s/node.pem", certPath), string(nodeCertByte)) {
		return errors.New("failed to write node certificate")
	}

	_, err = exec.Command(
		"openssl",
		"pkcs12",
		"-export",
		"-certfile", fmt.Sprintf("%s/root.pem", certPath),
		"-inkey", fmt.Sprintf("%s/node-key.pem", certPath),
		"-in", fmt.Sprintf("%s/node.pem", certPath),
		"-password", fmt.Sprintf("pass:%s", pass),
		"-out", fmt.Sprintf("%s/node.pkcs12", certPath),
	).Output()
	if err != nil {
		return errors.New("failed to generate node.pkcs12")
	}

	_, err = exec.Command(
		"keytool",
		"-importkeystore",
		"-srckeystore", fmt.Sprintf("%s/node.pkcs12", certPath),
		"-srcalias", "1",
		"-storepass", pass,
		"-srcstoretype", "pkcs12",
		"-srcstorepass", pass,
		"-destalias", "elasticsearch-node",
		"-destkeystore", fmt.Sprintf("%s/node.jks", certPath),
	).Output()
	if err != nil {
		return errors.New("failed to generate node.pk12")
	}

	return nil
}

func createAdminCertificate(certPath string, caKey *rsa.PrivateKey, caCert *x509.Certificate, pass string) error {
	cfg := cert.Config{
		CommonName:   "sgadmin",
		Organization: []string{"Elasticsearch Operator"},
		AltNames: cert.AltNames{
			DNSNames: []string{
				"localhost",
			},
		},
		Usages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
	}

	sgAdminKey, err := cert.NewPrivateKey()
	if err != nil {
		return errors.New("failed to generate key for sgadmin certificate")
	}
	sgAdminCert, err := cert.NewSignedCert(cfg, sgAdminKey, caCert, caKey)
	if err != nil {
		return errors.New("failed to sign sgadmin certificate")
	}

	sgAdminKeyByte := cert.EncodePrivateKeyPEM(sgAdminKey)
	if !ioutil.WriteString(fmt.Sprintf("%s/sgadmin-key.pem", certPath), string(sgAdminKeyByte)) {
		return errors.New("failed to write key for sgadmin certificate")
	}
	sgAdminCertByte := cert.EncodeCertPEM(sgAdminCert)
	if !ioutil.WriteString(fmt.Sprintf("%s/sgadmin.pem", certPath), string(sgAdminCertByte)) {
		return errors.New("failed to write sgadmin certificate")
	}

	_, err = exec.Command(
		"openssl",
		"pkcs12",
		"-export",
		"-certfile", fmt.Sprintf("%s/root.pem", certPath),
		"-inkey", fmt.Sprintf("%s/sgadmin-key.pem", certPath),
		"-in", fmt.Sprintf("%s/sgadmin.pem", certPath),
		"-password", fmt.Sprintf("pass:%s", pass),
		"-out", fmt.Sprintf("%s/sgadmin.pkcs12", certPath),
	).Output()
	if err != nil {
		return errors.New("failed to generate sgadmin.pkcs12")
	}

	_, err = exec.Command(
		"keytool",
		"-importkeystore",
		"-srckeystore", fmt.Sprintf("%s/sgadmin.pkcs12", certPath),
		"-srcalias", "1",
		"-storepass", pass,
		"-srcstoretype", "pkcs12",
		"-srcstorepass", pass,
		"-destalias", "elasticsearch-sgadmin",
		"-destkeystore", fmt.Sprintf("%s/sgadmin.jks", certPath),
	).Output()

	if err != nil {
		return errors.New("failed to generate sgadmin.jks")
	}

	return nil
}

func createClientCertificate(certPath string, elasticsearch *api.Elasticsearch, caKey *rsa.PrivateKey, caCert *x509.Certificate, pass string) error {
	cfg := cert.Config{
		CommonName:   elasticsearch.OffshootName(),
		Organization: []string{"Elasticsearch Operator"},
		AltNames: cert.AltNames{
			DNSNames: []string{
				"localhost",
				fmt.Sprintf("%v.%v.svc", elasticsearch.OffshootName(), elasticsearch.Namespace),
			},
		},
		Usages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
	}

	clientKey, err := cert.NewPrivateKey()
	if err != nil {
		return errors.New("failed to generate key for client certificate")
	}
	clientCert, err := cert.NewSignedCert(cfg, clientKey, caCert, caKey)
	if err != nil {
		return errors.New("failed to sign client certificate")
	}

	clientKeyByte := cert.EncodePrivateKeyPEM(clientKey)
	if !ioutil.WriteString(fmt.Sprintf("%s/client-key.pem", certPath), string(clientKeyByte)) {
		return errors.New("failed to write key for client certificate")
	}
	clientCertByte := cert.EncodeCertPEM(clientCert)
	if !ioutil.WriteString(fmt.Sprintf("%s/client.pem", certPath), string(clientCertByte)) {
		return errors.New("failed to write client certificate")
	}

	_, err = exec.Command(
		"openssl",
		"pkcs12",
		"-export",
		"-certfile", fmt.Sprintf("%s/root.pem", certPath),
		"-inkey", fmt.Sprintf("%s/client-key.pem", certPath),
		"-in", fmt.Sprintf("%s/client.pem", certPath),
		"-password", fmt.Sprintf("pass:%s", pass),
		"-out", fmt.Sprintf("%s/client.pkcs12", certPath),
	).Output()
	if err != nil {
		return errors.New("failed to generate client.pkcs12")
	}

	_, err = exec.Command(
		"keytool",
		"-importkeystore",
		"-srckeystore", fmt.Sprintf("%s/client.pkcs12", certPath),
		"-srcalias", "1",
		"-storepass", pass,
		"-srcstoretype", "pkcs12",
		"-srcstorepass", pass,
		"-destalias", "elasticsearch-client",
		"-destkeystore", fmt.Sprintf("%s/client.jks", certPath),
	).Output()

	if err != nil {
		return errors.New("failed to generate client.jks")
	}

	return nil
}

const (
	duration365d = time.Hour * 24 * 365
)

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
		NotAfter:     time.Now().Add(duration365d).UTC(),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  cfg.Usages,
		ExtraExtensions: []pkix.Extension{
			{
				Id: oidExtensionSubjectAltName,
			},
		},
	}
	certTmpl.ExtraExtensions[0].Value, err = marshalSANs(cfg.AltNames.DNSNames, nil, cfg.AltNames.IPs)

	certDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(certDERBytes)
}

var (
	oidExtensionSubjectAltName = []int{2, 5, 29, 17}
)

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
