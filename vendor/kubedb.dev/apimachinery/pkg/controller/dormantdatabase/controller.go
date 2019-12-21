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

package dormantdatabase

import (
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	kubedb_informers "kubedb.dev/apimachinery/client/informers/externalversions/kubedb/v1alpha1"
	api_listers "kubedb.dev/apimachinery/client/listers/kubedb/v1alpha1"
	amc "kubedb.dev/apimachinery/pkg/controller"

	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	crdutils "kmodules.xyz/client-go/apiextensions/v1beta1"
	"kmodules.xyz/client-go/tools/queue"
)

type Controller struct {
	*amc.Controller
	amc.Config
	// Deleter interface
	deleter amc.Deleter
	// tweakListOptions for watcher
	tweakListOptions func(*metav1.ListOptions)
	// Event Recorder
	recorder record.EventRecorder
	// DormantDatabase
	ddbLister api_listers.DormantDatabaseLister
}

// NewController creates a new DormantDatabase Controller
func NewController(
	controller *amc.Controller,
	deleter amc.Deleter,
	config amc.Config,
	tweakListOptions func(*metav1.ListOptions),
	recorder record.EventRecorder,
) *Controller {
	// return new DormantDatabase Controller
	return &Controller{
		Controller:       controller,
		deleter:          deleter,
		Config:           config,
		tweakListOptions: tweakListOptions,
		recorder:         recorder,
	}
}

func (c *Controller) EnsureCustomResourceDefinitions() error {
	crd := []*crd_api.CustomResourceDefinition{
		api.DormantDatabase{}.CustomResourceDefinition(),
	}
	return crdutils.RegisterCRDs(c.ApiExtKubeClient, crd)
}

func (c *Controller) InitInformer() cache.SharedIndexInformer {
	return c.KubedbInformerFactory.InformerFor(&api.DormantDatabase{}, func(client cs.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
		return kubedb_informers.NewFilteredDormantDatabaseInformer(
			client,
			c.WatchNamespace,
			resyncPeriod,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
			c.tweakListOptions,
		)
	})
}

func (c *Controller) AddEventHandlerFunc(selector labels.Selector) *queue.Worker {
	c.addEventHandler(selector)
	return c.DrmnQueue
}
