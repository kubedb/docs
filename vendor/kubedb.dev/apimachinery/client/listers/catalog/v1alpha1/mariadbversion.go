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
	v1alpha1 "kubedb.dev/apimachinery/apis/catalog/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// MariaDBVersionLister helps list MariaDBVersions.
// All objects returned here must be treated as read-only.
type MariaDBVersionLister interface {
	// List lists all MariaDBVersions in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.MariaDBVersion, err error)
	// Get retrieves the MariaDBVersion from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.MariaDBVersion, error)
	MariaDBVersionListerExpansion
}

// mariaDBVersionLister implements the MariaDBVersionLister interface.
type mariaDBVersionLister struct {
	indexer cache.Indexer
}

// NewMariaDBVersionLister returns a new MariaDBVersionLister.
func NewMariaDBVersionLister(indexer cache.Indexer) MariaDBVersionLister {
	return &mariaDBVersionLister{indexer: indexer}
}

// List lists all MariaDBVersions in the indexer.
func (s *mariaDBVersionLister) List(selector labels.Selector) (ret []*v1alpha1.MariaDBVersion, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MariaDBVersion))
	})
	return ret, err
}

// Get retrieves the MariaDBVersion from the index for a given name.
func (s *mariaDBVersionLister) Get(name string) (*v1alpha1.MariaDBVersion, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("mariadbversion"), name)
	}
	return obj.(*v1alpha1.MariaDBVersion), nil
}