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
)

func (c *Controller) ensureAppBinding(db *api.Redis) (kutil.VerbType, error) {
	port, err := c.GetPrimaryServicePort(db)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	appmeta := db.AppBindingMeta()

	meta := metav1.ObjectMeta{
		Name:      appmeta.Name(),
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindRedis))

	redisVersion, err := c.DBClient.CatalogV1alpha1().RedisVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	params := v1alpha1.RedisConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       v1alpha1.ResourceKindRedisConfiguration,
		},
		Stash: redisVersion.Spec.Stash,
	}
	var caBundle []byte
	if db.Spec.TLS != nil {

		certSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.GetCertSecretName(api.RedisClientCert), metav1.GetOptions{})
		if err != nil {
			return kutil.VerbUnchanged, errors.Wrapf(err, "failed to read certificate secret for Redis %s/%s", db.Namespace, db.Name)
		}
		v, ok := certSecret.Data[api.TLSCACertFileName]
		if !ok {
			return kutil.VerbUnchanged, errors.Errorf("ca.cert is missing in certificate secret for Redis %s/%s", db.Namespace, db.Name)
		}
		caBundle = v
		params.ClientCertSecret = &core.LocalObjectReference{
			Name: certSecret.Name,
		}
	}

	_, vt, err := appcat_util.CreateOrPatchAppBinding(context.TODO(), c.AppCatalogClient.AppcatalogV1alpha1(), meta, func(in *appcat.AppBinding) *appcat.AppBinding {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = db.OffshootLabels()
		in.Annotations = meta_util.FilterKeys(kubedb.GroupName, nil, db.Annotations)

		in.Spec.Type = appmeta.Type()
		in.Spec.Version = redisVersion.Spec.Version
		in.Spec.Secret = db.Spec.AuthSecret

		scheme := "redis"
		if db.Spec.TLS != nil {
			scheme = "rediss"
		}
		in.Spec.ClientConfig.Service = &appcat.ServiceReference{
			Scheme: scheme,
			Name:   db.ServiceName(),
			Port:   port,
		}
		in.Spec.ClientConfig.CABundle = caBundle
		in.Spec.ClientConfig.InsecureSkipTLSVerify = false
		in.Spec.Parameters = &runtime.RawExtension{
			Object: &params,
		}

		return in
	}, metav1.PatchOptions{})

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

func (c *Controller) GetPrimaryServicePort(db *api.Redis) (int32, error) {
	ports := ofst.PatchServicePorts([]core.ServicePort{
		{
			Name:       api.RedisPrimaryServicePortName,
			Port:       api.RedisDatabasePort,
			TargetPort: intstr.FromString(api.RedisDatabasePortName),
		},
	}, api.GetServiceTemplate(db.Spec.ServiceTemplates, api.PrimaryServiceAlias).Spec.Ports)

	for _, p := range ports {
		if p.Name == api.RedisPrimaryServicePortName {
			return p.Port, nil
		}
	}
	return 0, fmt.Errorf("failed to detect primary port for Redis %s/%s", db.Namespace, db.Name)
}

func (c *Controller) ensureSentinelAppBinding(db *api.RedisSentinel) (kutil.VerbType, error) {
	port := api.RedisSentinelPort

	appmeta := db.AppBindingMeta()

	meta := metav1.ObjectMeta{
		Name:      appmeta.Name(),
		Namespace: db.Namespace,
	}

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindRedisSentinel))

	redisVersion, err := c.DBClient.CatalogV1alpha1().RedisVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	params := v1alpha1.RedisConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       v1alpha1.ResourceKindRedisConfiguration,
		},
		Stash: redisVersion.Spec.Stash,
	}
	var caBundle []byte
	if db.Spec.TLS != nil {

		certSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.GetCertSecretName(api.RedisClientCert), metav1.GetOptions{})
		if err != nil {
			return kutil.VerbUnchanged, errors.Wrapf(err, "failed to read certificate secret for Redis Seninel %s/%s", db.Namespace, db.Name)
		}
		v, ok := certSecret.Data[api.TLSCACertFileName]
		if !ok {
			return kutil.VerbUnchanged, errors.Errorf("ca.cert is missing in certificate secret for Redis Sentinel %s/%s", db.Namespace, db.Name)
		}
		caBundle = v
		params.ClientCertSecret = &core.LocalObjectReference{
			Name: certSecret.Name,
		}
	}

	_, vt, err := appcat_util.CreateOrPatchAppBinding(context.TODO(), c.AppCatalogClient.AppcatalogV1alpha1(), meta, func(in *appcat.AppBinding) *appcat.AppBinding {
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Labels = db.OffshootLabels()
		in.Annotations = meta_util.FilterKeys(kubedb.GroupName, nil, db.Annotations)

		in.Spec.Type = appmeta.Type()
		in.Spec.Version = redisVersion.Spec.Version
		in.Spec.Secret = db.Spec.AuthSecret

		scheme := "redis"
		if db.Spec.TLS != nil {
			scheme = "rediss"
		}
		in.Spec.ClientConfig.Service = &appcat.ServiceReference{
			Scheme: scheme,
			Name:   db.GoverningServiceName(),
			Port:   int32(port),
		}
		in.Spec.ClientConfig.CABundle = caBundle
		in.Spec.ClientConfig.InsecureSkipTLSVerify = false
		in.Spec.Parameters = &runtime.RawExtension{
			Object: &params,
		}
		return in
	}, metav1.PatchOptions{})

	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.Recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s Sentinel appbinding",
			vt,
		)
	}
	return vt, nil
}
