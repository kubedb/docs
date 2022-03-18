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

package client

import (
	"context"
	"strings"

	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/tools/clusterid"

	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func NewUncachedClient(cfg *rest.Config, funcs ...func(*runtime.Scheme) error) (client.Client, error) {
	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		return nil, err
	}

	builder := runtime.NewSchemeBuilder(funcs...)
	builder.Register(clientgoscheme.AddToScheme)
	scheme := runtime.NewScheme()
	err = builder.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}

	return client.New(cfg, client.Options{
		Scheme: scheme,
		Mapper: mapper,
		//Opts: client.WarningHandlerOptions{
		//	SuppressWarnings:   false,
		//	AllowDuplicateLogs: false,
		//},
	})
}

type TransformFunc func(obj client.Object, createOp bool) client.Object

func CreateOrPatch(ctx context.Context, c client.Client, obj client.Object, transform TransformFunc, opts ...client.PatchOption) (client.Object, kutil.VerbType, error) {
	key := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
	err := c.Get(ctx, key, obj)
	if kerr.IsNotFound(err) {
		klog.V(3).Infof("Creating %+v %s/%s.", obj.GetObjectKind().GroupVersionKind(), key.Namespace, key.Name)

		createOpts := make([]client.CreateOption, 0, len(opts))
		for i := range opts {
			if opt, ok := opts[i].(client.CreateOption); ok {
				createOpts = append(createOpts, opt)
			}
		}
		obj = transform(obj.DeepCopyObject().(client.Object), true)
		err := c.Create(ctx, obj, createOpts...)
		return obj, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	var patch client.Patch
	if isOfficialTypes(obj.GetObjectKind().GroupVersionKind().Group) {
		patch = client.StrategicMergeFrom(obj)
	} else {
		patch = client.MergeFrom(obj)
	}

	obj = transform(obj.DeepCopyObject().(client.Object), false)
	err = c.Patch(ctx, obj, patch, opts...)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return obj, kutil.VerbPatched, nil
}

func PatchStatus(ctx context.Context, c client.Client, obj client.Object, transform TransformFunc, opts ...client.PatchOption) (client.Object, kutil.VerbType, error) {
	key := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
	err := c.Get(ctx, key, obj)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	var patch client.Patch
	if isOfficialTypes(obj.GetObjectKind().GroupVersionKind().Group) {
		patch = client.StrategicMergeFrom(obj)
	} else {
		patch = client.MergeFrom(obj)
	}

	obj = transform(obj.DeepCopyObject().(client.Object), false)
	err = c.Status().Patch(ctx, obj, patch, opts...)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return obj, kutil.VerbPatched, nil
}

func isOfficialTypes(group string) bool {
	return !strings.ContainsRune(group, '.')
}

func GetForGVR(ctx context.Context, c client.Client, gvr schema.GroupVersionResource, ref types.NamespacedName) (client.Object, error) {
	gvk, err := c.RESTMapper().KindFor(gvr)
	if err != nil {
		return nil, err
	}
	o, err := c.Scheme().New(gvk)
	if err != nil {
		return nil, err
	}
	obj := o.(client.Object)
	err = c.Get(ctx, ref, obj)
	return obj, err
}

func GetForGVK(ctx context.Context, c client.Client, gvk schema.GroupVersionKind, ref types.NamespacedName) (client.Object, error) {
	if gvk.Version == "" {
		mapping, err := c.RESTMapper().RESTMapping(gvk.GroupKind())
		if err != nil {
			return nil, err
		}
		gvk = mapping.GroupVersionKind
	}
	o, err := c.Scheme().New(gvk)
	if err != nil {
		return nil, err
	}
	obj := o.(client.Object)
	err = c.Get(ctx, ref, obj)
	return obj, err
}

func ClusterUID(c client.Reader) (string, error) {
	var ns core.Namespace
	err := c.Get(context.TODO(), client.ObjectKey{Name: metav1.NamespaceSystem}, &ns)
	if err != nil {
		return "", err
	}
	return string(ns.UID), nil
}

func ClusterMetadata(c client.Reader) (*kmapi.ClusterMetadata, error) {
	var ns core.Namespace
	err := c.Get(context.TODO(), client.ObjectKey{Name: metav1.NamespaceSystem}, &ns)
	if err != nil {
		return nil, err
	}
	return clusterid.ClusterMetadataForNamespace(&ns)
}
