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
	"encoding/json"
	"fmt"

	"kubedb.dev/apimachinery/apis/config/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_util "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/util"
	"stash.appscode.dev/apimachinery/pkg/restic"
)

func (c *Controller) ensureAppBinding(db *api.MongoDB) (kutil.VerbType, error) {
	appmeta := db.AppBindingMeta()

	meta := metav1.ObjectMeta{
		Name:      appmeta.Name(),
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMongoDB))

	// jsonBytes contains parameters in json format for appbinding.spec.parameters.raw
	var jsonBytes []byte
	var err error
	if db.Spec.ShardTopology != nil || db.Spec.ReplicaSet != nil {
		replicaHosts := make(map[string]string)
		if db.Spec.ShardTopology != nil {
			for i := int32(0); i < db.Spec.ShardTopology.Shard.Shards; i++ {
				replicaHosts[fmt.Sprintf("host-%v", i)] = db.ShardDSN(i)
			}
		} else if db.Spec.ReplicaSet != nil {
			replicaHosts[restic.DefaultHost] = db.HostAddress()
		}

		parameter := v1alpha1.MongoDBConfiguration{
			TypeMeta: metav1.TypeMeta{
				APIVersion: v1alpha1.SchemeGroupVersion.String(),
				Kind:       v1alpha1.ResourceKindMongoConfiguration,
			},
			ConfigServer: db.ConfigSvrDSN(),
			ReplicaSets:  replicaHosts,
		}
		if jsonBytes, err = json.Marshal(parameter); err != nil {
			return kutil.VerbUnchanged, fmt.Errorf("fail to serialize appbinding spec.Parameters. reason: %v", err)
		}
	}

	var caBundle []byte
	if (db.Spec.SSLMode == api.SSLModeRequireSSL || db.Spec.SSLMode == api.SSLModePreferSSL) &&
		db.Spec.TLS != nil {

		certSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.MustCertSecretName(api.MongoDBClientCert, ""), metav1.GetOptions{})
		if err != nil {
			return kutil.VerbUnchanged, errors.Wrapf(err, "failed to read certificate secret for MongoDB %s/%s", db.Namespace, db.Name)
		}
		v, ok := certSecret.Data[api.TLSCACertFileName]
		if !ok {
			return kutil.VerbUnchanged, errors.Errorf("ca.cert is missing in certificate secret for MongoDB %s/%s", db.Namespace, db.Name)
		}
		caBundle = v
	}

	clientPEMSecretName := db.Spec.DatabaseSecret.SecretName
	if caBundle != nil {
		clientPEMSecretName = db.MustCertSecretName(api.MongoDBClientCert, "")
	}

	mongodbVersion, err := c.ExtClient.CatalogV1alpha1().MongoDBVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return kutil.VerbUnchanged, fmt.Errorf("failed to get MongoDBVersion %v for %v/%v. Reason: %v", db.Spec.Version, db.Namespace, db.Name, err)
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
			in.Spec.Version = mongodbVersion.Spec.Version
			in.Spec.ClientConfig.Service = &appcat.ServiceReference{
				Scheme: "mongodb",
				Name:   db.ServiceName(),
				Port:   defaultDBPort.Port,
			}
			in.Spec.ClientConfig.CABundle = caBundle
			in.Spec.ClientConfig.InsecureSkipTLSVerify = false

			in.Spec.Secret = &core.LocalObjectReference{
				Name: clientPEMSecretName,
			}

			if jsonBytes != nil {
				in.Spec.Parameters = &runtime.RawExtension{
					Raw: jsonBytes,
				}
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
