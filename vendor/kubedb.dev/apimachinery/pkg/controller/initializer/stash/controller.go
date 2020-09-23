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

package stash

import (
	"time"

	amc "kubedb.dev/apimachinery/pkg/controller"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"kmodules.xyz/client-go/tools/queue"
	"stash.appscode.dev/apimachinery/apis/stash/v1beta1"
	scs "stash.appscode.dev/apimachinery/client/clientset/versioned"
	stashinformer "stash.appscode.dev/apimachinery/client/informers/externalversions"
)

type Controller struct {
	*amc.Controller
	*amc.StashInitializer
	// SnapshotDoer interface
	snapshotter amc.DBHelper
	// Event Recorder
	eventRecorder record.EventRecorder
	// Namespace to watch
	watchNamespace string
}

func NewController(
	ctrl *amc.Controller,
	initializer *amc.StashInitializer,
	snapshotter amc.DBHelper,
	recorder record.EventRecorder,
	watchNamespace string,
) *Controller {
	return &Controller{
		Controller:       ctrl,
		StashInitializer: initializer,
		snapshotter:      snapshotter,
		eventRecorder:    recorder,
		watchNamespace:   watchNamespace,
	}
}

type restoreInfo struct {
	invoker      core.TypedLocalObjectReference
	namespace    string
	target       *v1beta1.RestoreTarget
	phase        v1beta1.RestorePhase
	targetDBKind string
}

func Configure(cfg *rest.Config, s *amc.StashInitializer, resyncPeriod time.Duration) error {
	var err error
	if s.StashClient, err = scs.NewForConfig(cfg); err != nil {
		return err
	}
	s.StashInformerFactory = stashinformer.NewSharedInformerFactory(s.StashClient, resyncPeriod)
	return nil
}

func (c *Controller) InitWatcher(maxNumRequeues, numThreads int, selector labels.Selector) {
	log.Infoln("Initializing stash watchers.....")
	// only watch  the restore invokers that matches the selector
	tweakListOptions := func(options *metav1.ListOptions) {
		options.LabelSelector = selector.String()
	}
	// Initialize RestoreSession Watcher
	c.RSInformer = c.restoreSessionInformer(tweakListOptions)
	c.RSQueue = queue.New(v1beta1.ResourceKindRestoreSession, maxNumRequeues, numThreads, c.processRestoreSession)
	c.RSLister = c.StashInformerFactory.Stash().V1beta1().RestoreSessions().Lister()
	c.RSInformer.AddEventHandler(c.restoreSessionEventHandler(selector))

	// Initialize RestoreBatch Watcher
	c.RBInformer = c.restoreBatchInformer(tweakListOptions)
	c.RBQueue = queue.New(v1beta1.ResourceKindRestoreBatch, maxNumRequeues, numThreads, c.processRestoreBatch)
	c.RBLister = c.StashInformerFactory.Stash().V1beta1().RestoreBatches().Lister()
	c.RBInformer.AddEventHandler(c.restoreBatchEventHandler(selector))
}

func (c *Controller) StartController(stopCh <-chan struct{}) {
	// Start StashInformerFactory only if stash crds (ie, "RestoreSession") are available.
	if err := c.waitUntilStashInstalled(stopCh); err != nil {
		log.Errorln("error during waiting for RestoreSession crd. Reason: ", err)
		return
	}

	log.Infoln("Starting Stash controllers...")
	// start informer factory
	c.StashInformerFactory.Start(stopCh)
	// wait for cache to sync
	for t, v := range c.StashInformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			log.Errorf("%v timed out waiting for caches to sync", t)
			return
		}
	}
	// run the queues
	c.RSQueue.Run(stopCh)
	c.RBQueue.Run(stopCh)
}
