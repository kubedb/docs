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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "kubedb.dev/apimachinery/apis/autoscaling/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// MySQLAutoscalerLister helps list MySQLAutoscalers.
type MySQLAutoscalerLister interface {
	// List lists all MySQLAutoscalers in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.MySQLAutoscaler, err error)
	// MySQLAutoscalers returns an object that can list and get MySQLAutoscalers.
	MySQLAutoscalers(namespace string) MySQLAutoscalerNamespaceLister
	MySQLAutoscalerListerExpansion
}

// mySQLAutoscalerLister implements the MySQLAutoscalerLister interface.
type mySQLAutoscalerLister struct {
	indexer cache.Indexer
}

// NewMySQLAutoscalerLister returns a new MySQLAutoscalerLister.
func NewMySQLAutoscalerLister(indexer cache.Indexer) MySQLAutoscalerLister {
	return &mySQLAutoscalerLister{indexer: indexer}
}

// List lists all MySQLAutoscalers in the indexer.
func (s *mySQLAutoscalerLister) List(selector labels.Selector) (ret []*v1alpha1.MySQLAutoscaler, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MySQLAutoscaler))
	})
	return ret, err
}

// MySQLAutoscalers returns an object that can list and get MySQLAutoscalers.
func (s *mySQLAutoscalerLister) MySQLAutoscalers(namespace string) MySQLAutoscalerNamespaceLister {
	return mySQLAutoscalerNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// MySQLAutoscalerNamespaceLister helps list and get MySQLAutoscalers.
type MySQLAutoscalerNamespaceLister interface {
	// List lists all MySQLAutoscalers in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.MySQLAutoscaler, err error)
	// Get retrieves the MySQLAutoscaler from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.MySQLAutoscaler, error)
	MySQLAutoscalerNamespaceListerExpansion
}

// mySQLAutoscalerNamespaceLister implements the MySQLAutoscalerNamespaceLister
// interface.
type mySQLAutoscalerNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all MySQLAutoscalers in the indexer for a given namespace.
func (s mySQLAutoscalerNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.MySQLAutoscaler, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MySQLAutoscaler))
	})
	return ret, err
}

// Get retrieves the MySQLAutoscaler from the indexer for a given namespace and name.
func (s mySQLAutoscalerNamespaceLister) Get(name string) (*v1alpha1.MySQLAutoscaler, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("mysqlautoscaler"), name)
	}
	return obj.(*v1alpha1.MySQLAutoscaler), nil
}