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

	"kubedb.dev/apimachinery/apis/config/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_util "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/util"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	"stash.appscode.dev/apimachinery/pkg/restic"
)

func (c *Controller) ensureAppBinding(db *api.MongoDB) (kutil.VerbType, error) {
	port, err := c.GetPrimaryServicePort(db)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	appmeta := db.AppBindingMeta()

	meta := metav1.ObjectMeta{
		Name:      appmeta.Name(),
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindMongoDB))

	mongodbVersion, err := c.DBClient.CatalogV1alpha1().MongoDBVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return kutil.VerbUnchanged, fmt.Errorf("failed to get MongoDBVersion %v for %v/%v. Reason: %v", db.Spec.Version, db.Namespace, db.Name, err)
	}

	params := v1alpha1.MongoDBConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       v1alpha1.ResourceKindMongoConfiguration,
		},
		Stash: mongodbVersion.Spec.Stash,
	}

	if db.Spec.ShardTopology != nil || db.Spec.ReplicaSet != nil {
		replicaHosts := make(map[string]string)
		if db.Spec.ShardTopology != nil {
			for i := int32(0); i < db.Spec.ShardTopology.Shard.Shards; i++ {
				replicaHosts[fmt.Sprintf("host-%v", i)] = db.ShardDSN(i)
			}
		} else if db.Spec.ReplicaSet != nil {
			replicaHosts[restic.DefaultHost] = db.HostAddress()
		}

		params.ConfigServer = db.ConfigSvrDSN()
		params.ReplicaSets = replicaHosts
	}

	var caBundle []byte
	if (db.Spec.SSLMode == api.SSLModeRequireSSL || db.Spec.SSLMode == api.SSLModePreferSSL) &&
		db.Spec.TLS != nil {

		certSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.GetCertSecretName(api.MongoDBClientCert, ""), metav1.GetOptions{})
		if err != nil {
			return kutil.VerbUnchanged, errors.Wrapf(err, "failed to read certificate secret for MongoDB %s/%s", db.Namespace, db.Name)
		}
		v, ok := certSecret.Data[api.TLSCACertFileName]
		if !ok {
			return kutil.VerbUnchanged, errors.Errorf("ca.cert is missing in certificate secret for MongoDB %s/%s", db.Namespace, db.Name)
		}
		caBundle = v
	}

	clientPEMSecretName := db.Spec.AuthSecret.Name
	if caBundle != nil {
		clientPEMSecretName = db.GetCertSecretName(api.MongoDBClientCert, "")
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
			in.Spec.Version = mongodbVersion.Spec.Version
			in.Spec.ClientConfig.Service = &appcat.ServiceReference{
				Scheme: "mongodb",
				Name:   db.ServiceName(),
				Port:   port,
			}
			in.Spec.ClientConfig.CABundle = caBundle
			in.Spec.ClientConfig.InsecureSkipTLSVerify = false
			in.Spec.Parameters = &runtime.RawExtension{
				Object: &params,
			}

			in.Spec.Secret = &core.LocalObjectReference{
				Name: clientPEMSecretName,
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

func (c *Controller) GetPrimaryServicePort(db *api.MongoDB) (int32, error) {
	ports := ofst.PatchServicePorts([]core.ServicePort{
		{
			Name:       api.MongoDBPrimaryServicePortName,
			Port:       api.MongoDBDatabasePort,
			TargetPort: intstr.FromString(api.MongoDBDatabasePortName),
		},
	}, api.GetServiceTemplate(db.Spec.ServiceTemplates, api.PrimaryServiceAlias).Spec.Ports)

	for _, p := range ports {
		if p.Name == api.MongoDBPrimaryServicePortName {
			return p.Port, nil
		}
	}
	return 0, fmt.Errorf("failed to detect primary port for MongoDB %s/%s", db.Namespace, db.Name)
}
