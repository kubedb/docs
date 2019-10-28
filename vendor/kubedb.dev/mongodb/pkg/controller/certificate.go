package controller

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	"github.com/pkg/errors"
	"gomodules.xyz/cert"
)

// createCaCertificate returns generated caKey, caCert, err in order.
func createCaCertificate() (*rsa.PrivateKey, *x509.Certificate, error) {
	cfg := cert.Config{
		CommonName:   "ca",
		Organization: []string{"kubedb:ca"},
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
func createPEMCertificate(caKey *rsa.PrivateKey, caCert *x509.Certificate, cfg cert.Config) ([]byte, error) {
	privateKey, err := cert.NewPrivateKey()
	if err != nil {
		return nil, errors.New("failed to generate key for client certificate")
	}

	certificate, err := cert.NewSignedCert(cfg, privateKey, caCert, caKey)
	if err != nil {
		return nil, errors.New("failed to sign client certificate")
	}

	keyBytes := cert.EncodePrivateKeyPEM(privateKey)
	certBytes := cert.EncodeCertPEM(certificate)
	pemBytes := append(certBytes, keyBytes...)

	return pemBytes, nil
}

// createServerPEMCertificate returns generated Key, Cert, err in order.
// xref: https://docs.mongodb.com/manual/core/security-x.509/#member-x-509-certificates
func createServerPEMCertificate(mongodb *api.MongoDB, caKey *rsa.PrivateKey, caCert *x509.Certificate) ([]byte, error) {
	cfg := cert.Config{
		CommonName:   mongodb.OffshootName(),
		Organization: []string{"kubedb:server"},
		AltNames: cert.AltNames{
			DNSNames: []string{
				"localhost",
				fmt.Sprintf("%v.%v.svc", mongodb.OffshootName(), mongodb.Namespace),
				mongodb.OffshootName(),
				mongodb.ServiceName(),
			},
			IPs: []net.IP{net.ParseIP("127.0.0.1")},
		},
		Usages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
	}
	return createPEMCertificate(caKey, caCert, cfg)
}

// createPEMCertificate returns generated Key, Cert, err in order.
// xref: https://docs.mongodb.com/manual/tutorial/configure-x509-client-authentication/
func createClientPEMCertificate(mongodb *api.MongoDB, caKey *rsa.PrivateKey, caCert *x509.Certificate) ([]byte, error) {
	cfg := cert.Config{
		CommonName:   "root",
		Organization: []string{"kubedb:client"},
		AltNames: cert.AltNames{
			DNSNames: []string{
				"localhost",
				fmt.Sprintf("%v.%v.svc", mongodb.OffshootName(), mongodb.Namespace),
				mongodb.OffshootName(),
				mongodb.ServiceName(),
			},
			IPs: []net.IP{net.ParseIP("127.0.0.1")},
		},
		Usages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
	}
	return createPEMCertificate(caKey, caCert, cfg)
}
