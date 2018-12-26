package pkcs12

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/appscode/go/ioutil"
	"github.com/kubedb/elasticsearch/third_party/golang/crypto/pkcs12"
	keystorego "github.com/pavel-v-chernykh/keystore-go"
	"github.com/pkg/errors"
)

const (
	contentTypePrivateKey  = "PRIVATE KEY"
	contentTypeCertificate = "CERTIFICATE"
	defaultCertificateType = "X509"
)

func PKCS12ToJKS(sourceFile, destinationFile, pass, alias string) error {

	// read source pkcs12 encoded file
	srcContent, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return errors.Wrapf(err, "failed to read source PKCS12 encoded file %s", sourceFile)
	}

	// decode pkcs12 encoded source file's content
	pvtKeys, certs, err := pkcs12.DecodeAll([]byte(srcContent), pass)
	if err != nil {
		return errors.Wrapf(err, "failed to decode pkcs12 encoded %s", sourceFile)
	}

	var pvtKey interface{}
	for _, key := range pvtKeys {
		pvtKey = key
	}

	keyBytes, err := x509.MarshalPKCS8PrivateKey(pvtKey)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal pvtKeys to PKCS8PrivateKey.")
	}

	certChain := []keystorego.Certificate{}
	for _, cert := range certs {
		certChain = append(certChain, keystorego.Certificate{
			Type:    defaultCertificateType,
			Content: cert.Raw,
		})
	}

	// create keystore
	ks := keystorego.KeyStore{
		alias: &keystorego.PrivateKeyEntry{
			Entry: keystorego.Entry{
				CreationDate: time.Now(),
			},
			PrivKey:   keyBytes,
			CertChain: certChain,
		},
	}

	// write keystore(.jks) file
	err = writeKeyStoreFile(ks, destinationFile, pass)
	if err != nil {
		return errors.Wrapf(err, "failed to create keystore: %s", destinationFile)
	}
	return nil
}

func PEMToJKS(sourceFile, destinationFile, pass, alias string) error {

	// read the source pem encoded file
	srcContent, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return errors.Wrapf(err, "failed to read cert file: %s", sourceFile)
	}

	// decode .pem file content
	decodedContent, _ := pem.Decode([]byte(srcContent))

	var ks keystorego.KeyStore

	// create keystore based on content type
	switch decodedContent.Type {
	case contentTypePrivateKey:
		ks = keystorego.KeyStore{
			alias: &keystorego.PrivateKeyEntry{
				Entry: keystorego.Entry{
					CreationDate: time.Now(),
				},
				PrivKey: decodedContent.Bytes,
			},
		}
	case contentTypeCertificate:
		ks = keystorego.KeyStore{
			alias: &keystorego.TrustedCertificateEntry{
				Entry: keystorego.Entry{
					CreationDate: time.Now(),
				},
				Certificate: keystorego.Certificate{
					Type:    defaultCertificateType,
					Content: decodedContent.Bytes,
				},
			},
		}
	default:
		return errors.Wrap(fmt.Errorf("unknown %s file content type", sourceFile), "failed to create keystore")

	}

	// write keystore(.jks) file
	err = writeKeyStoreFile(ks, destinationFile, pass)
	if err != nil {
		return errors.Wrapf(err, "failed to create keystore: %s", destinationFile)
	}

	return nil
}

func writeKeyStoreFile(keyStore keystorego.KeyStore, filename string, password string) error {
	o, err := os.Create(filename)
	defer o.Close()
	if err != nil {
		return err
	}
	err = keystorego.Encode(o, keyStore, []byte(password))
	if err != nil {
		return err
	}
	return nil
}
