/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package certholder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type CertHolder struct {
	certDir string
}

var DefaultHolder = New(filepath.Join(os.TempDir(), "certs"))

func New(certDir string) *CertHolder {
	return &CertHolder{
		certDir: certDir,
	}
}

func (h *CertHolder) ForResource(gvr schema.GroupVersionResource, r metav1.ObjectMeta) (*ResourceCerts, bool) {
	oc := &ResourceCerts{
		holder: h,
		gvr:    gvr,
		r:      r,
	}
	_, err := os.Stat(oc.dir())
	return oc, !os.IsNotExist(err)
}

type ResourceCerts struct {
	holder *CertHolder
	gvr    schema.GroupVersionResource
	r      metav1.ObjectMeta
}

type Paths struct {
	CACert string
	Cert   string
	Key    string
	Pem    string
}

func (c *ResourceCerts) Save(secret *core.Secret) (*Paths, error) {
	err := os.MkdirAll(c.dir(), os.ModePerm)
	if err != nil {
		return nil, err
	}

	paths := Paths{
		CACert: filepath.Join(c.dir(), core.ServiceAccountRootCAKey),
		Cert:   filepath.Join(c.dir(), secret.Name+".crt"),
		Key:    filepath.Join(c.dir(), secret.Name+".key"),
		Pem:    filepath.Join(c.dir(), secret.Name+".pem"),
	}

	caCrt, ok := secret.Data[core.ServiceAccountRootCAKey]
	if !ok {
		return nil, fmt.Errorf("missing %s in secret %s/%s", core.ServiceAccountRootCAKey, secret.Namespace, secret.Name)
	}
	err = ioutil.WriteFile(paths.CACert, caCrt, 0644)
	if err != nil {
		return nil, err
	}

	crt, ok := secret.Data[core.TLSCertKey]
	if !ok {
		return nil, fmt.Errorf("missing %s in secret %s/%s", core.TLSCertKey, secret.Namespace, secret.Name)
	}
	err = ioutil.WriteFile(paths.Cert, crt, 0644)
	if err != nil {
		return nil, err
	}

	key, ok := secret.Data[core.TLSPrivateKeyKey]
	if !ok {
		return nil, fmt.Errorf("missing %s in secret %s/%s", core.TLSPrivateKeyKey, secret.Namespace, secret.Name)
	}
	err = ioutil.WriteFile(paths.Key, key, 0600)
	if err != nil {
		return nil, err
	}

	pem := append(crt[:], []byte("\n")...)
	pem = append(pem, key...)
	err = ioutil.WriteFile(paths.Pem, pem, 0600)
	if err != nil {
		return nil, err
	}
	return &paths, nil
}

func (c *ResourceCerts) Get(secretName string) (*Paths, error) {
	paths := Paths{
		CACert: filepath.Join(c.dir(), core.ServiceAccountRootCAKey),
		Cert:   filepath.Join(c.dir(), secretName+".crt"),
		Key:    filepath.Join(c.dir(), secretName+".key"),
		Pem:    filepath.Join(c.dir(), secretName+".pem"),
	}
	if _, err := os.Stat(paths.CACert); os.IsNotExist(err) {
		return nil, err
	}
	if _, err := os.Stat(paths.Cert); os.IsNotExist(err) {
		return nil, err
	}
	if _, err := os.Stat(paths.Key); os.IsNotExist(err) {
		return nil, err
	}
	if _, err := os.Stat(paths.Pem); os.IsNotExist(err) {
		return nil, err
	}
	return &paths, nil
}

func (c *ResourceCerts) dir() string {
	// /tmp/certs/apps/v1/namespaces/$ns/deployments/$name/
	return filepath.Join(c.holder.certDir, c.gvr.Group, c.gvr.Version, "namespaces", c.r.Namespace, c.gvr.Resource, c.r.Name)
}
