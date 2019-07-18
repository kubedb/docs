package controller

import (
	"encoding/json"
	"fmt"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_util "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/util"
	"kubedb.dev/apimachinery/apis/config/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"
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
	replicaHosts := make(map[string]string)
	if db.Spec.ShardTopology != nil || db.Spec.ReplicaSet != nil {
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

	_, vt, err := appcat_util.CreateOrPatchAppBinding(c.AppCatalogClient, meta, func(in *appcat.AppBinding) *appcat.AppBinding {
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = db.OffshootLabels()
		in.Annotations = db.Spec.ServiceTemplate.Annotations

		in.Spec.Type = appmeta.Type()
		in.Spec.ClientConfig.Service = &appcat.ServiceReference{
			Scheme: "mongodb",
			Name:   db.ServiceName(),
			Port:   defaultDBPort.Port,
		}
		in.Spec.ClientConfig.InsecureSkipTLSVerify = false

		in.Spec.Secret = &core.LocalObjectReference{
			Name: db.Spec.DatabaseSecret.SecretName,
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
