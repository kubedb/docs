package controller

import (
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net"
	"os/exec"
	"time"

	"github.com/appscode/go/io"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"k8s.io/client-go/util/cert"
)

var certsDir = "/tmp/certs/certs"

func createCaCertificate(certPath string) (*rsa.PrivateKey, *x509.Certificate, error) {
	cfg := cert.Config{
		CommonName:   "KubeDB Com. Root CA",
		Organization: []string{"Elasticsearch Operator"},
	}

	caKey, err := cert.NewPrivateKey()
	if err != nil {
		return nil, nil, errors.New("Failed to generate key for CA certificate")
	}

	caCert, err := cert.NewSelfSignedCACert(cfg, caKey)
	if err != nil {
		return nil, nil, errors.New("Failed to generate CA certificate")
	}
	caCertByte := cert.EncodeCertPEM(caCert)
	if !io.WriteString(fmt.Sprintf("%s/ca.pem", certPath), string(caCertByte)) {
		return nil, nil, errors.New("Failed to write CA certificate")
	}

	_, err = exec.Command(
		"keytool",
		"-import",
		"-file", fmt.Sprintf("%s/ca.pem", certPath),
		"-alias", "root-ca",
		"-keystore", fmt.Sprintf("%s/truststore.jks", certPath),
		"-storepass", "changeit",
		"-srcstoretype", "pkcs12",
		"-noprompt",
	).Output()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate truststore.jks")
	}

	return caKey, caCert, nil
}

func createNodeCertificate(certPath string, elasticsearch *api.Elasticsearch, caKey *rsa.PrivateKey, caCert *x509.Certificate) error {
	name := elasticsearch.OffshootName()

	cfg := cert.Config{
		CommonName:   name,
		Organization: []string{"Elasticsearch Operator"},
		AltNames: cert.AltNames{
			DNSNames: []string{
				"localhost",
				name,
				fmt.Sprintf("%v.%v", name, elasticsearch.Namespace),
				fmt.Sprintf("%v.%v.svc.cluster.local", name, elasticsearch.Namespace),
			},
		},
		Usages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
	}

	nodeKey, err := cert.NewPrivateKey()
	if err != nil {
		return errors.New("Failed to generate key for node certificate")
	}
	nodeCert, err := NewSignedCert(cfg, nodeKey, caCert, caKey)
	if err != nil {
		return errors.New("Failed to sign node certificate")
	}

	nodeKeyByte := cert.EncodePrivateKeyPEM(nodeKey)
	if !io.WriteString(fmt.Sprintf("%s/node-key.pem", certPath), string(nodeKeyByte)) {
		return errors.New("Failed to write key for node certificate")
	}
	nodeCertByte := cert.EncodeCertPEM(nodeCert)
	if !io.WriteString(fmt.Sprintf("%s/node.pem", certPath), string(nodeCertByte)) {
		return errors.New("Failed to write node certificate")
	}

	_, err = exec.Command(
		"openssl",
		"pkcs12",
		"-export",
		"-certfile", fmt.Sprintf("%s/ca.pem", certPath),
		"-inkey", fmt.Sprintf("%s/node-key.pem", certPath),
		"-in", fmt.Sprintf("%s/node.pem", certPath),
		"-password", "pass:changeit",
		"-out", fmt.Sprintf("%s/node.pkcs12", certPath),
	).Output()
	if err != nil {
		return errors.New("Failed to generate node.pkcs12")
	}

	_, err = exec.Command(
		"keytool",
		"-importkeystore",
		"-srckeystore", fmt.Sprintf("%s/node.pkcs12", certPath),
		"-srcalias", "1",
		"-storepass", "changeit",
		"-srcstoretype", "pkcs12",
		"-srcstorepass", "changeit",
		"-destalias", "elasticsearch-node",
		"-destkeystore", fmt.Sprintf("%s/keystore.jks", certPath),
	).Output()
	if err != nil {
		return errors.New("Failed to generate keystore.jks")
	}

	return nil
}

func createAdminCertificate(certPath string, caKey *rsa.PrivateKey, caCert *x509.Certificate) error {
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
		return errors.New("Failed to generate key for sgadmin certificate")
	}
	sgAdminCert, err := cert.NewSignedCert(cfg, sgAdminKey, caCert, caKey)
	if err != nil {
		return errors.New("Failed to sign sgadmin certificate")
	}

	sgAdminKeyByte := cert.EncodePrivateKeyPEM(sgAdminKey)
	if !io.WriteString(fmt.Sprintf("%s/sgadmin-key.pem", certPath), string(sgAdminKeyByte)) {
		return errors.New("Failed to write key for sgadmin certificate")
	}
	sgAdminCertByte := cert.EncodeCertPEM(sgAdminCert)
	if !io.WriteString(fmt.Sprintf("%s/sgadmin.pem", certPath), string(sgAdminCertByte)) {
		return errors.New("Failed to write sgadmin certificate")
	}

	_, err = exec.Command(
		"openssl",
		"pkcs12",
		"-export",
		"-certfile", fmt.Sprintf("%s/ca.pem", certPath),
		"-inkey", fmt.Sprintf("%s/sgadmin-key.pem", certPath),
		"-in", fmt.Sprintf("%s/sgadmin.pem", certPath),
		"-password", "pass:changeit",
		"-out", fmt.Sprintf("%s/sgadmin.pkcs12", certPath),
	).Output()
	if err != nil {
		return errors.New("Failed to generate sgadmin.pkcs12")
	}

	_, err = exec.Command(
		"keytool",
		"-importkeystore",
		"-srckeystore", fmt.Sprintf("%s/sgadmin.pkcs12", certPath),
		"-srcalias", "1",
		"-storepass", "changeit",
		"-srcstoretype", "pkcs12",
		"-srcstorepass", "changeit",
		"-destalias", "elasticsearch-sgadmin",
		"-destkeystore", fmt.Sprintf("%s/sgadmin.jks", certPath),
	).Output()

	if err != nil {
		return errors.New("Failed to generate sgadmin-keystore.jks")
	}

	return nil
}

func createClientCertificate(certPath string, caKey *rsa.PrivateKey, caCert *x509.Certificate) error {
	cfg := cert.Config{
		CommonName:   "client",
		Organization: []string{"Elasticsearch Operator"},
		Usages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
	}

	clientKey, err := cert.NewPrivateKey()
	if err != nil {
		return errors.New("Failed to generate key for client certificate")
	}
	clientCert, err := cert.NewSignedCert(cfg, clientKey, caCert, caKey)
	if err != nil {
		return errors.New("Failed to sign client certificate")
	}

	clientKeyByte := cert.EncodePrivateKeyPEM(clientKey)
	if !io.WriteString(fmt.Sprintf("%s/client-key.pem", certPath), string(clientKeyByte)) {
		return errors.New("Failed to write key for client certificate")
	}
	clientCertByte := cert.EncodeCertPEM(clientCert)
	if !io.WriteString(fmt.Sprintf("%s/client.pem", certPath), string(clientCertByte)) {
		return errors.New("Failed to write client certificate")
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
	rawValues = append(rawValues, asn1.RawValue{FullBytes: []byte{0x88, 0x05, 0x2A, 0x03, 0x04, 0x05, 0x05}})
	return asn1.Marshal(rawValues)
}
