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

package controller

import (
	"context"
	"fmt"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/pkg/eventer"
	certlib "kubedb.dev/elasticsearch/pkg/lib/cert"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	kutil "kmodules.xyz/client-go"
	api_util "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_util "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/util"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

func (c *Controller) ensureAppBinding(db *api.Elasticsearch) (kutil.VerbType, error) {
	port, err := c.GetPrimaryServicePort(db)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	appmeta := db.AppBindingMeta()

	meta := metav1.ObjectMeta{
		Name:      appmeta.Name(),
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))

	var caBundle []byte
	if db.Spec.EnableSSL {
		if db.Spec.TLS == nil {
			return kutil.VerbUnchanged, errors.New("missing TLS configuration")
		}

		sName, exist := api_util.GetCertificateSecretName(db.Spec.TLS.Certificates, string(api.ElasticsearchHTTPCert))
		if !exist {
			return kutil.VerbUnchanged, errors.New("http-cert secret is missing")
		}

		certSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), sName, metav1.GetOptions{})
		if err != nil {
			return kutil.VerbUnchanged, errors.Wrapf(err, "failed to read http-cert secret for Elasticsearch %s/%s", db.Namespace, db.Name)
		}

		if v, ok := certSecret.Data[certlib.CACert]; !ok {
			return kutil.VerbUnchanged, errors.Errorf("ca.crt is missing in http-cert secret for Elasticsearch %s/%s", db.Namespace, db.Name)
		} else {
			caBundle = v
		}
	}

	elasticsearchVersion, err := c.esVersionLister.Get(string(db.Spec.Version))
	if err != nil {
		return kutil.VerbUnchanged, fmt.Errorf("failed to get ElasticsearchVersion %v for %v/%v. Reason: %v", db.Spec.Version, db.Namespace, db.Name, err)
	}

	_, vt, err := appcat_util.CreateOrPatchAppBinding(
		context.TODO(),
		c.AppCatalogClient.AppcatalogV1alpha1(),
		meta,
		func(in *appcat.AppBinding) *appcat.AppBinding {
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
			in.Labels = db.OffshootLabels()
			in.Annotations = meta_util.FilterKeys(kubedb.GroupName, nil, db.Annotations)

			in.Spec.Type = appmeta.Type()
			in.Spec.Version = elasticsearchVersion.Spec.Version
			in.Spec.ClientConfig.Service = &appcat.ServiceReference{
				Scheme: db.GetConnectionScheme(),
				Name:   db.ServiceName(),
				Port:   port,
			}
			in.Spec.ClientConfig.CABundle = caBundle
			in.Spec.ClientConfig.InsecureSkipTLSVerify = false
			in.Spec.Parameters = &runtime.RawExtension{
				Object: &appcat.StashAddon{
					TypeMeta: metav1.TypeMeta{
						APIVersion: appcat.SchemeGroupVersion.String(),
						Kind:       "StashAddon",
					},
					Stash: elasticsearchVersion.Spec.Stash,
				},
			}

			in.Spec.Secret = &core.LocalObjectReference{
				Name: db.Spec.AuthSecret.Name,
			}

			return in
		},
		metav1.PatchOptions{},
	)

	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.Recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s appbinding",
			vt,
		)
	}
	return vt, nil
}

func (c *Controller) GetPrimaryServicePort(db *api.Elasticsearch) (int32, error) {
	ports := ofst.PatchServicePorts([]core.ServicePort{
		{
			Name:       api.ElasticsearchRestPortName,
			Port:       api.ElasticsearchRestPort,
			TargetPort: intstr.FromString(api.ElasticsearchRestPortName),
		},
	}, api.GetServiceTemplate(db.Spec.ServiceTemplates, api.PrimaryServiceAlias).Spec.Ports)

	for _, p := range ports {
		if p.Name == api.ElasticsearchRestPortName {
			return p.Port, nil
		}
	}
	return 0, fmt.Errorf("failed to detect primary port for Elasticsearch %s/%s", db.Namespace, db.Name)
}
