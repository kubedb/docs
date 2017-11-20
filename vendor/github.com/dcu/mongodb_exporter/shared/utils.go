package shared

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"regexp"
	"strings"
)

var (
	snakeRegexp        = regexp.MustCompile("\\B[A-Z]+[^_$]")
	parameterizeRegexp = regexp.MustCompile("[^A-Za-z0-9_]+")
)

// SnakeCase converts the given text to snakecase/underscore syntax.
func SnakeCase(text string) string {
	result := snakeRegexp.ReplaceAllStringFunc(text, func(match string) string {
		return "_" + match
	})

	return ParameterizeString(result)
}

// ParameterizeString parameterizes the given string.
func ParameterizeString(text string) string {
	result := parameterizeRegexp.ReplaceAllString(text, "_")
	return strings.ToLower(result)
}

// LoadCertificatesFrom returns certificates for a given pem file
func LoadCertificatesFrom(pemFile string) (*x509.CertPool, error) {
	caCert, err := ioutil.ReadFile(pemFile)
	if err != nil {
		return nil, err
	}
	certificates := x509.NewCertPool()
	certificates.AppendCertsFromPEM(caCert)
	return certificates, nil
}

// LoadKeyPairFrom returns a configured TLS certificate
func LoadKeyPairFrom(pemFile string, privateKeyPemFile string) (tls.Certificate, error) {
	targetPrivateKeyPemFile := privateKeyPemFile
	if len(targetPrivateKeyPemFile) <= 0 {
		targetPrivateKeyPemFile = pemFile
	}
	return tls.LoadX509KeyPair(pemFile, targetPrivateKeyPemFile)
}
