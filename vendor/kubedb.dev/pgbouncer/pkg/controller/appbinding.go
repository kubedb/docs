/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

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

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_util "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/util"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

func (r *Reconciler) getCABundlesFromAppBindingsInPgBouncerSpec(db *api.PgBouncer) (string, error) {
	isCAForAppBindingInserted := map[string]bool{}
	var myCAStrings string
	if db.Spec.Databases != nil && len(db.Spec.Databases) > 0 {
		for _, db := range db.Spec.Databases {
			appBinding, err := r.AppCatalogClient.AppcatalogV1alpha1().AppBindings(db.DatabaseRef.Namespace).Get(context.TODO(), db.DatabaseRef.Name, metav1.GetOptions{})
			if err != nil {
				if kerr.IsNotFound(err) {
					klog.Infoln(err)
					continue // because non blocking err
				}
				return "", err
			}
			if !isCAForAppBindingInserted[appBinding.Namespace+"/"+appBinding.Name] && len(appBinding.Spec.ClientConfig.CABundle) > 0 {
				isCAForAppBindingInserted[appBinding.Namespace+"/"+appBinding.Name] = true
				myCAStrings = myCAStrings + fmt.Sprintln(string(appBinding.Spec.ClientConfig.CABundle))
			}
		}
	}

	return myCAStrings, nil
}

func (r *Reconciler) getClientCertFromAppbindingsInPgBouncerSpec(db *api.PgBouncer) (string, error) {
	isTlsCertForAppBindingInserted := map[string]bool{}
	var myCertStrings string
	if db.Spec.Databases != nil && len(db.Spec.Databases) > 0 {
		for _, db := range db.Spec.Databases {
			appBinding, err := r.AppCatalogClient.AppcatalogV1alpha1().AppBindings(db.DatabaseRef.Namespace).Get(context.TODO(), db.DatabaseRef.Name, metav1.GetOptions{})
			if err != nil {
				if kerr.IsNotFound(err) {
					klog.Infoln(err)
					continue // because non blocking err
				}
				return "", err
			}
			if !isTlsCertForAppBindingInserted[appBinding.Namespace+"/"+appBinding.Name] {
				certSecret, err := r.Client.CoreV1().Secrets(db.DatabaseRef.Namespace).Get(context.TODO(), appBinding.Spec.Secret.Name, metav1.GetOptions{})
				if err != nil {
					return "", err
				}

				if data, ok := certSecret.Data[api.PgBouncerTLSCrt]; ok {
					myCertStrings = myCertStrings + fmt.Sprintln(string(data))
				}
				isTlsCertForAppBindingInserted[appBinding.Namespace+"/"+appBinding.Name] = true
			}
		}
	}

	return myCertStrings, nil
}

func (r *Reconciler) getClientKeyFromAppbindingsInPgBouncerSpec(db *api.PgBouncer) (string, error) {
	isTlsKeyForAppBindingInserted := map[string]bool{}
	var myKeyStrings string
	if db.Spec.Databases != nil && len(db.Spec.Databases) > 0 {
		for _, db := range db.Spec.Databases {
			appBinding, err := r.AppCatalogClient.AppcatalogV1alpha1().AppBindings(db.DatabaseRef.Namespace).Get(context.TODO(), db.DatabaseRef.Name, metav1.GetOptions{})
			if err != nil {
				if kerr.IsNotFound(err) {
					klog.Infoln(err)
					continue // because non blocking err
				}
				return "", err
			}
			if !isTlsKeyForAppBindingInserted[appBinding.Namespace+"/"+appBinding.Name] {
				certSecret, err := r.Client.CoreV1().Secrets(db.DatabaseRef.Namespace).Get(context.TODO(), appBinding.Spec.Secret.Name, metav1.GetOptions{})
				if err != nil {
					return "", err
				}

				if data, ok := certSecret.Data[api.PgBouncerTLSKey]; ok {
					myKeyStrings = myKeyStrings + fmt.Sprintln(string(data))
				}
				isTlsKeyForAppBindingInserted[appBinding.Namespace+"/"+appBinding.Name] = true
			}
		}
	}

	return myKeyStrings, nil
}

func (c *Controller) ensureAppBinding(db *api.PgBouncer, pgBouncerVersion *catalog.PgBouncerVersion) (kutil.VerbType, error) {
	port, err := c.GetPrimaryServicePort(db)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	appmeta := db.AppBindingMeta()

	meta := metav1.ObjectMeta{
		Name:      appmeta.Name(),
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindPgBouncer))

	var caBundle []byte
	if db.Spec.TLS != nil {
		certSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.GetCertSecretName(api.PgBouncerClientCert), metav1.GetOptions{})
		if err != nil {
			return kutil.VerbUnchanged, errors.Wrapf(err, "failed to read certificate secret for PgBouncer %s/%s", db.Namespace, db.Name)
		}
		v, ok := certSecret.Data[api.TLSCACertFileName]
		if !ok {
			return kutil.VerbUnchanged, errors.Errorf("ca.cert is missing in certificate secret for PgBouncer %s/%s", db.Namespace, db.Name)
		}
		caBundle = v
	}
	clientPEMSecretName := db.Spec.UserListSecretRef.Name
	if db.Spec.ConnectionPool.AuthType == api.PgBouncerClientAuthModeCert {
		clientPEMSecretName = db.GetCertSecretName(api.PgBouncerClientCert)
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
			in.Spec.Version = pgBouncerVersion.Spec.Version

			in.Spec.ClientConfig.Service = &appcat.ServiceReference{
				Scheme: "pgbouncer",
				Name:   db.ServiceName(),
				Port:   port,
				Path:   "/",
				// this Query field need to have the exact template of sslmode=<your_desire_ssl_mode>
				Query: fmt.Sprintf("sslmode=%s", db.Spec.SSLMode),
			}
			in.Spec.ClientConfig.InsecureSkipTLSVerify = false
			if caBundle != nil {
				in.Spec.ClientConfig.CABundle = caBundle
			}
			in.Spec.Secret = &core.LocalObjectReference{
				Name: clientPEMSecretName,
			}

			return in
		}, metav1.PatchOptions{},
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

func (c *Controller) GetPrimaryServicePort(db *api.PgBouncer) (int32, error) {
	ports := ofst.PatchServicePorts([]core.ServicePort{
		{
			Name:       api.PgBouncerPrimaryServicePortName,
			Port:       api.PgBouncerDatabasePort,
			TargetPort: intstr.FromString(api.PgBouncerDatabasePortName),
		},
	}, api.GetServiceTemplate(db.Spec.ServiceTemplates, api.PrimaryServiceAlias).Spec.Ports)

	for _, p := range ports {
		if p.Name == api.PgBouncerPrimaryServicePortName {
			return p.Port, nil
		}
	}
	return 0, fmt.Errorf("failed to detect primary port for PgBouncer %s/%s", db.Namespace, db.Name)
}
