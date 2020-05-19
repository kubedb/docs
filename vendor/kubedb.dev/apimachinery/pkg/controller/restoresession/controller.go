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

package restoresession

import (
	"time"

	amc "kubedb.dev/apimachinery/pkg/controller"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"kmodules.xyz/client-go/tools/queue"
	"stash.appscode.dev/apimachinery/apis/stash/v1beta1"
	scs "stash.appscode.dev/apimachinery/client/clientset/versioned"
	stashinformers "stash.appscode.dev/apimachinery/client/informers/externalversions/stash/v1beta1"
	stashLister "stash.appscode.dev/apimachinery/client/listers/stash/v1beta1"
)

type Controller struct {
	*amc.Controller
	amc.Config
	// SnapshotDoer interface
	snapshotter amc.DBHelper
	// tweakListOptions for watcher
	tweakListOptions func(*metav1.ListOptions)
	// Event Recorder
	eventRecorder record.EventRecorder
	// restoreSession Lister
	rsLister stashLister.RestoreSessionLister
}

// NewController creates a new Controller
func NewController(
	controller *amc.Controller,
	snapshotter amc.DBHelper,
	config amc.Config,
	tweakListOptions func(*metav1.ListOptions),
	eventRecorder record.EventRecorder,
) *Controller {
	return &Controller{
		Controller:       controller,
		snapshotter:      snapshotter,
		Config:           config,
		tweakListOptions: tweakListOptions,
		eventRecorder:    eventRecorder,
	}
}

func (c *Controller) InitInformer() cache.SharedIndexInformer {
	return c.StashInformerFactory.InformerFor(&v1beta1.RestoreSession{}, func(client scs.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
		return stashinformers.NewFilteredRestoreSessionInformer(
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
	return c.RSQueue
}
