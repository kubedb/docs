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

package elastic_stack

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	certlib "kubedb.dev/elasticsearch/pkg/lib/cert"
	"kubedb.dev/elasticsearch/pkg/lib/cert/pkcs8"

	"github.com/pkg/errors"
	"gomodules.xyz/cert"
	"gomodules.xyz/x/crypto/rand"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

// EnsureCertSecrets creates certificates if they don't exist.
// If the "TLS.IssuerRef" is set, the operator won't create certificates.
func (es *Elasticsearch) EnsureCertSecrets() error {
	if es.db.Spec.DisableSecurity {
		return nil
	}

	if es.db.Spec.TLS == nil {
		return errors.New("tls configuration is missing")
	}

	// Certificates are managed by the enterprise operator.
	// Ignore sync/creation.
	if es.db.Spec.TLS.IssuerRef != nil {
		return nil
	}

	certPath := fmt.Sprintf("%v/%v", certlib.CertsDir, rand.Characters(3))
	if err := os.MkdirAll(certPath, os.ModePerm); err != nil {
		return err
	}

	caKey, caCert, err := es.createCACertSecret(certPath)
	if err != nil {
		return errors.Wrap(err, "failed to create/sync root-cert secret")
	}

	err = es.createTransportCertSecret(caKey, caCert, certPath)
	if err != nil {
		return errors.Wrap(err, "failed to create/sync transport-cert secret")
	}

	if es.db.Spec.EnableSSL {
		// When SSL is enabled, create certificates for HTTP layer
		err = es.createHTTPCertSecret(caKey, caCert, certPath)
		if err != nil {
			return errors.Wrap(err, "failed to create/sync http-cert secret")
		}

		// Create certificates for metrics exporter, if monitoring is enabled.
		if es.db.Spec.Monitor != nil {
			err = es.createExporterCertSecret(caKey, caCert, certPath)
			if err != nil {
				return errors.Wrap(err, "failed to create/sync metrics-exporter-cert secret")
			}
		}

		err = es.createArchiverCertSecret(caKey, caCert, certPath)
		if err != nil {
			return errors.Wrap(err, "failed to create/sync archiver-cert secret")
		}
	}

	return nil
}

func (es *Elasticsearch) createCACertSecret(cPath string) (*rsa.PrivateKey, *x509.Certificate, error) {
	rSecret, err := es.findSecret(es.db.MustCertSecretName(api.ElasticsearchCACert))
	if err != nil {
		return nil, nil, err
	}

	if rSecret == nil {
		// create certs here
		caKey, caCert, err := pkcs8.CreateCaCertificate(es.db.ClientCertificateCN(api.ElasticsearchCACert), cPath)
		if err != nil {
			return nil, nil, err
		}
		rootCa, err := ioutil.ReadFile(filepath.Join(cPath, certlib.CACert))
		if err != nil {
			return nil, nil, err
		}
		rootKey, err := ioutil.ReadFile(filepath.Join(cPath, certlib.CAKey))
		if err != nil {
			return nil, nil, err
		}

		data := map[string][]byte{
			certlib.TLSCert: rootCa,
			certlib.TLSKey:  rootKey,
		}

		owner := metav1.NewControllerRef(es.db, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   es.db.MustCertSecretName(api.ElasticsearchCACert),
				Labels: es.db.OffshootLabels(),
			},
			Type: core.SecretTypeTLS,
			Data: data,
		}
		core_util.EnsureOwnerReference(&secret.ObjectMeta, owner)

		_, err = es.kClient.CoreV1().Secrets(es.db.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
		if err != nil {
			return nil, nil, err
		}

		return caKey, caCert, nil
	}

	data := rSecret.Data
	var caKey *rsa.PrivateKey
	var caCert []*x509.Certificate

	if value, ok := data[certlib.TLSCert]; ok {
		caCert, err = cert.ParseCertsPEM(value)
		if err != nil || len(caCert) == 0 {
			return nil, nil, errors.Wrap(err, "failed to parse tls.crt")
		}
	} else {
		return nil, nil, errors.New("tls.crt is missing")
	}

	if value, ok := data[certlib.TLSKey]; ok {
		key, err := cert.ParsePrivateKeyPEM(value)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to parse tls.key")
		}

		caKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, nil, errors.New("failed to typecast the tls.key")
		}

	} else {
		return nil, nil, errors.New("tls.key is missing")
	}

	return caKey, caCert[0], nil
}

func (es *Elasticsearch) createTransportCertSecret(caKey *rsa.PrivateKey, caCert *x509.Certificate, cPath string) error {
	nSecret, err := es.findSecret(es.db.MustCertSecretName(api.ElasticsearchTransportCert))
	if err != nil {
		return err
	}

	if nSecret == nil {
		// create certs here
		err := pkcs8.CreateTransportCertificate(cPath, es.db, caKey, caCert)
		if err != nil {
			return err
		}

		caCert, err := ioutil.ReadFile(filepath.Join(cPath, certlib.CACert))
		if err != nil {
			return err
		}

		nodeCert, err := ioutil.ReadFile(filepath.Join(cPath, certlib.TLSCert))
		if err != nil {
			return err
		}

		nodeKey, err := ioutil.ReadFile(filepath.Join(cPath, certlib.TLSKey))
		if err != nil {
			return err
		}

		data := map[string][]byte{
			certlib.CACert:  caCert,
			certlib.TLSKey:  nodeKey,
			certlib.TLSCert: nodeCert,
		}

		owner := metav1.NewControllerRef(es.db, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   es.db.MustCertSecretName(api.ElasticsearchTransportCert),
				Labels: es.db.OffshootLabels(),
			},
			Type: core.SecretTypeTLS,
			Data: data,
		}
		core_util.EnsureOwnerReference(&secret.ObjectMeta, owner)

		_, err = es.kClient.CoreV1().Secrets(es.db.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
		if err != nil {
			return err
		}

		return nil
	}

	// If the secret already exists,
	// check whether the keys exist too.
	if value, ok := nSecret.Data[certlib.CACert]; !ok || len(value) == 0 {
		return errors.New("ca.crt is missing")
	}

	if value, ok := nSecret.Data[certlib.TLSKey]; !ok || len(value) == 0 {
		return errors.New("tls.key is missing")
	}

	if value, ok := nSecret.Data[certlib.TLSCert]; !ok || len(value) == 0 {
		return errors.New("tls.crt is missing")
	}

	return nil
}

func (es *Elasticsearch) createHTTPCertSecret(caKey *rsa.PrivateKey, caCert *x509.Certificate, cPath string) error {
	cSecret, err := es.findSecret(es.db.MustCertSecretName(api.ElasticsearchHTTPCert))
	if err != nil {
		return err
	}

	if cSecret == nil {
		// create certs here
		if err := pkcs8.CreateHTTPCertificate(cPath, es.db, caKey, caCert); err != nil {
			return err
		}

		caCert, err := ioutil.ReadFile(filepath.Join(cPath, certlib.CACert))
		if err != nil {
			return err
		}

		clientCert, err := ioutil.ReadFile(filepath.Join(cPath, certlib.TLSCert))
		if err != nil {
			return err
		}

		clientKey, err := ioutil.ReadFile(filepath.Join(cPath, certlib.TLSKey))
		if err != nil {
			return err
		}

		data := map[string][]byte{
			certlib.CACert:  caCert,
			certlib.TLSKey:  clientKey,
			certlib.TLSCert: clientCert,
		}

		owner := metav1.NewControllerRef(es.db, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   es.db.MustCertSecretName(api.ElasticsearchHTTPCert),
				Labels: es.db.OffshootLabels(),
			},
			Type: core.SecretTypeTLS,
			Data: data,
		}
		core_util.EnsureOwnerReference(&secret.ObjectMeta, owner)

		_, err = es.kClient.CoreV1().Secrets(es.db.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
		if err != nil {
			return err
		}

		return nil
	}

	// If the secret already exists,
	// check whether the keys exist too.
	if value, ok := cSecret.Data[certlib.CACert]; !ok || len(value) == 0 {
		return errors.New("ca.crt is missing")
	}

	if value, ok := cSecret.Data[certlib.TLSKey]; !ok || len(value) == 0 {
		return errors.New("tls.key is missing")
	}

	if value, ok := cSecret.Data[certlib.TLSCert]; !ok || len(value) == 0 {
		return errors.New("tls.crt is missing")
	}

	return nil
}

func (es *Elasticsearch) createExporterCertSecret(caKey *rsa.PrivateKey, caCert *x509.Certificate, cPath string) error {
	cSecret, err := es.findSecret(es.db.MustCertSecretName(api.ElasticsearchMetricsExporterCert))
	if err != nil {
		return err
	}

	if cSecret == nil {
		// create certs here
		if err := pkcs8.CreateClientCertificate(string(api.ElasticsearchMetricsExporterCert), cPath, es.db, caKey, caCert); err != nil {
			return err
		}

		caCert, err := ioutil.ReadFile(filepath.Join(cPath, certlib.CACert))
		if err != nil {
			return err
		}

		clientCert, err := ioutil.ReadFile(filepath.Join(cPath, certlib.TLSCert))
		if err != nil {
			return err
		}

		clientKey, err := ioutil.ReadFile(filepath.Join(cPath, certlib.TLSKey))
		if err != nil {
			return err
		}

		data := map[string][]byte{
			certlib.CACert:  caCert,
			certlib.TLSKey:  clientKey,
			certlib.TLSCert: clientCert,
		}

		owner := metav1.NewControllerRef(es.db, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   es.db.MustCertSecretName(api.ElasticsearchMetricsExporterCert),
				Labels: es.db.OffshootLabels(),
			},
			Type: core.SecretTypeTLS,
			Data: data,
		}
		core_util.EnsureOwnerReference(&secret.ObjectMeta, owner)

		_, err = es.kClient.CoreV1().Secrets(es.db.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
		if err != nil {
			return err
		}

		return nil
	}

	// If the secret already exists,
	// check whether the keys exist too.
	if value, ok := cSecret.Data[certlib.CACert]; !ok || len(value) == 0 {
		return errors.New("ca.crt is missing")
	}

	if value, ok := cSecret.Data[certlib.TLSKey]; !ok || len(value) == 0 {
		return errors.New("tls.key is missing")
	}

	if value, ok := cSecret.Data[certlib.TLSCert]; !ok || len(value) == 0 {
		return errors.New("tls.crt is missing")
	}

	return nil
}

func (es *Elasticsearch) createArchiverCertSecret(caKey *rsa.PrivateKey, caCert *x509.Certificate, cPath string) error {
	cSecret, err := es.findSecret(es.db.MustCertSecretName(api.ElasticsearchArchiverCert))
	if err != nil {
		return err
	}

	if cSecret == nil {
		// create certs here
		if err := pkcs8.CreateClientCertificate(string(api.ElasticsearchArchiverCert), cPath, es.db, caKey, caCert); err != nil {
			return err
		}

		caCert, err := ioutil.ReadFile(filepath.Join(cPath, certlib.CACert))
		if err != nil {
			return err
		}

		clientCert, err := ioutil.ReadFile(filepath.Join(cPath, certlib.TLSCert))
		if err != nil {
			return err
		}

		clientKey, err := ioutil.ReadFile(filepath.Join(cPath, certlib.TLSKey))
		if err != nil {
			return err
		}

		data := map[string][]byte{
			certlib.CACert:  caCert,
			certlib.TLSKey:  clientKey,
			certlib.TLSCert: clientCert,
		}

		owner := metav1.NewControllerRef(es.db, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   es.db.MustCertSecretName(api.ElasticsearchArchiverCert),
				Labels: es.db.OffshootLabels(),
			},
			Type: core.SecretTypeTLS,
			Data: data,
		}
		core_util.EnsureOwnerReference(&secret.ObjectMeta, owner)

		_, err = es.kClient.CoreV1().Secrets(es.db.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
		if err != nil {
			return err
		}

		return nil
	}

	// If the secret already exists,
	// check whether the keys exist too.
	if value, ok := cSecret.Data[certlib.CACert]; !ok || len(value) == 0 {
		return errors.New("ca.crt is missing")
	}

	if value, ok := cSecret.Data[certlib.TLSKey]; !ok || len(value) == 0 {
		return errors.New("tls.key is missing")
	}

	if value, ok := cSecret.Data[certlib.TLSCert]; !ok || len(value) == 0 {
		return errors.New("tls.crt is missing")
	}

	return nil
}

func (es *Elasticsearch) findSecret(name string) (*core.Secret, error) {

	secret, err := es.kClient.CoreV1().Secrets(es.db.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return secret, nil
}
