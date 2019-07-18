package controller

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net"

	"github.com/pkg/errors"
	"gomodules.xyz/cert"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
)

const (
	TLSKey         = "tls.key"
	TLSCert        = "tls.crt"
	MongoServerPem = "mongo.pem"
	MongoClientPem = "client.pem"
)

// createCaCertificate returns generated caKey, caCert, err in order.
func createCaCertificate() (*rsa.PrivateKey, *x509.Certificate, error) {
	cfg := cert.Config{
		CommonName:   "KubeDB Com. Root CA",
		Organization: []string{"KubeDB Operator"},
	}

	caKey, err := cert.NewPrivateKey()
	if err != nil {
		return nil, nil, errors.New("failed to generate key for CA certificate")
	}

	caCert, err := cert.NewSelfSignedCACert(cfg, caKey)
	if err != nil {
		return nil, nil, errors.New("failed to generate CA certificate")
	}

	//caKeyByte := cert.EncodePrivateKeyPEM(caKey)
	//caCertByte := cert.EncodeCertPEM(caCert)

	return caKey, caCert, nil
}

// createPEMCertificate returns generated Key, Cert, err in order.
func createPEMCertificate(mongodb *api.MongoDB, caKey *rsa.PrivateKey, caCert *x509.Certificate) ([]byte, error) {
	cfg := cert.Config{
		CommonName:   mongodb.OffshootName(),
		Organization: []string{"MongoDB Operator"},
		AltNames: cert.AltNames{
			DNSNames: []string{
				"localhost",
				fmt.Sprintf("%v.%v.svc", mongodb.OffshootName(), mongodb.Namespace),
			},
			IPs: []net.IP{net.ParseIP("127.0.0.1")},
		},
		Usages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
	}

	clientPrivateKey, err := cert.NewPrivateKey()
	if err != nil {
		return nil, errors.New("failed to generate key for client certificate")
	}

	clientCertificate, err := cert.NewSignedCert(cfg, clientPrivateKey, caCert, caKey)
	if err != nil {
		return nil, errors.New("failed to sign client certificate")
	}

	clientKeyByte := cert.EncodePrivateKeyPEM(clientPrivateKey)
	clientCertByte := cert.EncodeCertPEM(clientCertificate)
	certBytes := append(clientCertByte, clientKeyByte...)

	return certBytes, nil
}
