/*
Copyright The KubeDB Authors.

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
package controller

import (
	"kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	"github.com/appscode/go/encoding/json/types"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

func (c *Controller) MigrateObservedGeneration() error {
	var errs []error
	for _, gvr := range []schema.GroupVersionResource{
		v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralDormantDatabase),
		v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralElasticsearch),
		v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralEtcd),
		// v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralMariaDB),
		v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralMemcached),
		v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralMongoDB),
		v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralMySQL),
		// v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralPerconaXtraDB),
		v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralPgBouncer),
		v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralPostgres),
		// v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralProxySQL),
		v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralRedis),
	} {
		client := c.DynamicClient.Resource(gvr)
		objects, err := client.Namespace(core.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			errs = append(errs, err)
			continue
			// return err
		}
		for _, obj := range objects.Items {
			changed1, e1 := convertObservedGenerationToInt64(&obj)
			if e1 != nil {
				errs = append(errs, e1)
			} else if changed1 {
				_, e2 := client.Namespace(obj.GetNamespace()).UpdateStatus(&obj, metav1.UpdateOptions{})
				errs = append(errs, e2)
			}
		}
	}
	return utilerrors.NewAggregate(errs)
}

func convertObservedGenerationToInt64(u *unstructured.Unstructured) (bool, error) {
	val, found, err := unstructured.NestedFieldNoCopy(u.Object, "status", "observedGeneration")
	if err != nil {
		return false, err
	}
	if found {
		if _, ok := val.(string); ok {
			observed, err := types.ParseIntHash(val)
			if err != nil {
				return false, err
			}
			err = unstructured.SetNestedField(u.Object, observed.Generation(), "status", "observedGeneration")
			if err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}
