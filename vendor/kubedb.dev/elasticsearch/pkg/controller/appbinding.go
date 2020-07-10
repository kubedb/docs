/*
Copyright AppsCode Inc. and Contributors

Licensed under the PolyForm Noncommercial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/PolyForm-Noncommercial-1.0.0.md

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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_util "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/util"
)

func (c *Controller) ensureAppBinding(db *api.Elasticsearch) (kutil.VerbType, error) {
	appmeta := db.AppBindingMeta()

	meta := metav1.ObjectMeta{
		Name:      appmeta.Name(),
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindElasticsearch))

	var caBundle []byte
	if db.Spec.EnableSSL {
		certSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.Spec.CertificateSecret.SecretName, metav1.GetOptions{})
		if err != nil {
			return kutil.VerbUnchanged, errors.Wrapf(err, "failed to read certificate secret for Elasticsearch %s/%s", db.Namespace, db.Name)
		}
		if v, ok := certSecret.Data["root.pem"]; !ok {
			return kutil.VerbUnchanged, errors.Errorf("root.pem is missing in certificate secret for Elasticsearch %s/%s", db.Namespace, db.Name)
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
			in.Annotations = meta_util.FilterKeys(api.GenericKey, in.Annotations, db.Annotations)

			in.Spec.Type = appmeta.Type()
			in.Spec.Version = elasticsearchVersion.Spec.Version
			in.Spec.ClientConfig.Service = &appcat.ServiceReference{
				Scheme: db.GetConnectionScheme(),
				Name:   db.ServiceName(),
				Port:   defaultClientPort.Port,
			}
			in.Spec.ClientConfig.CABundle = caBundle
			in.Spec.ClientConfig.InsecureSkipTLSVerify = false

			in.Spec.Secret = &core.LocalObjectReference{
				Name: db.Spec.DatabaseSecret.SecretName,
			}
			in.Spec.SecretTransforms = []appcat.SecretTransform{
				{
					RenameKey: &appcat.RenameKeyTransform{
						From: KeyAdminUserName,
						To:   appcat.KeyUsername,
					},
				},
				{
					RenameKey: &appcat.RenameKeyTransform{
						From: KeyAdminPassword,
						To:   appcat.KeyPassword,
					},
				},
			}

			return in
		},
		metav1.PatchOptions{},
	)

	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s appbinding",
			vt,
		)
	}
	return vt, nil
}
