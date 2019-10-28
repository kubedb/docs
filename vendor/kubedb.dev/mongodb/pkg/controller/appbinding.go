package controller

import (
	"encoding/json"
	"fmt"

	"kubedb.dev/apimachinery/apis/config/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_util "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/util"
	"stash.appscode.dev/stash/pkg/restic"
)

func (c *Controller) ensureAppBinding(db *api.MongoDB) (kutil.VerbType, error) {
	appmeta := db.AppBindingMeta()

	meta := metav1.ObjectMeta{
		Name:      appmeta.Name(),
		Namespace: db.Namespace,
	}

	ref, err := reference.GetReference(clientsetscheme.Scheme, db)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// jsonBytes contains parameters in json format for appbinding.spec.parameters.raw
	var jsonBytes []byte
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
			ConfigServer: db.ConfigSvrDSN(),
			ReplicaSets:  replicaHosts,
		}
		if jsonBytes, err = json.Marshal(parameter); err != nil {
			return kutil.VerbUnchanged, fmt.Errorf("fail to serialize appbinding spec.Parameters. reason: %v", err)
		}
	}

	var caBundle []byte
	if (db.Spec.SSLMode == api.SSLModeRequireSSL || db.Spec.SSLMode == api.SSLModePreferSSL) &&
		db.Spec.CertificateSecret != nil {
		certSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(db.Spec.CertificateSecret.SecretName, metav1.GetOptions{})
		if err != nil {
			return kutil.VerbUnchanged, errors.Wrapf(err, "failed to read certificate secret for MongoDB %s/%s", db.Namespace, db.Name)
		}
		v, ok := certSecret.Data[api.MongoTLSCertFileName]
		if !ok {
			return kutil.VerbUnchanged, errors.Errorf("ca.cert is missing in certificate secret for MongoDB %s/%s", db.Namespace, db.Name)
		}
		caBundle = v

	}

	secretName := db.Spec.DatabaseSecret.SecretName
	if caBundle != nil {
		secretName = db.Spec.CertificateSecret.SecretName
	}

	mongodbVersion, err := c.ExtClient.CatalogV1alpha1().MongoDBVersions().Get(string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return kutil.VerbUnchanged, fmt.Errorf("failed to get MongoDBVersion %v for %v/%v. Reason: %v", db.Spec.Version, db.Namespace, db.Name, err)
	}

	_, vt, err := appcat_util.CreateOrPatchAppBinding(c.AppCatalogClient.AppcatalogV1alpha1(), meta, func(in *appcat.AppBinding) *appcat.AppBinding {
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = db.OffshootLabels()

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
			Name: secretName,
		}

		if jsonBytes != nil {
			in.Spec.Parameters = &runtime.RawExtension{
				Raw: jsonBytes,
			}
		}

		return in
	})

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
