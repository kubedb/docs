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
	"strings"

	config_api "kubedb.dev/apimachinery/apis/config/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/appscode/go/types"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_util "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/util"
)

func getGarbdConfig(db *api.PerconaXtraDB) ([]byte, error) {
	if !db.IsCluster() {
		return nil, nil
	}

	var peers []string
	for i := 0; i < int(*db.Spec.Replicas); i += 1 {
		peers = append(peers, db.PeerName(i))
	}

	return json.Marshal(config_api.GaleraArbitratorConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: config_api.SchemeGroupVersion.String(),
			Kind:       config_api.ResourceKindGaleraArbitratorConfiguration,
		},
		Address:   fmt.Sprintf("gcomm://%s", strings.Join(peers, ",")),
		Group:     db.Name,
		SSTMethod: config_api.GarbdXtrabackupSSTMethod,
	})
}

func (c *Controller) ensureAppBinding(db *api.PerconaXtraDB) (kutil.VerbType, error) {
	appmeta := db.AppBindingMeta()

	meta := metav1.ObjectMeta{
		Name:      appmeta.Name(),
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindPerconaXtraDB))

	garbdCnfJson, err := getGarbdConfig(db)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	pxVersion, err := c.ExtClient.CatalogV1alpha1().PerconaXtraDBVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return kutil.VerbUnchanged, fmt.Errorf("failed to get PerconaXtraDBVersion %v for %v/%v. Reason: %v", db.Spec.Version, db.Namespace, db.Name, err)
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
			in.Spec.Version = pxVersion.Spec.Version
			in.Spec.ClientConfig.URL = types.StringP(fmt.Sprintf("tcp(%s:%d)/", db.ServiceName(), defaultDBPort.Port))
			in.Spec.ClientConfig.Service = &appcat.ServiceReference{
				Scheme: "mysql",
				Name:   db.ServiceName(),
				Port:   defaultDBPort.Port,
				Path:   "/",
			}
			in.Spec.ClientConfig.InsecureSkipTLSVerify = false

			in.Spec.Secret = &core.LocalObjectReference{
				Name: db.Spec.DatabaseSecret.SecretName,
			}

			if db.IsCluster() {
				in.Spec.Parameters = &runtime.RawExtension{
					Raw: garbdCnfJson,
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
