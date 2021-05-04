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

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	dmcond "kmodules.xyz/client-go/dynamic/conditions"
	"kmodules.xyz/client-go/tools/queue"
	"stash.appscode.dev/apimachinery/apis/stash/v1beta1"
	scs "stash.appscode.dev/apimachinery/client/clientset/versioned"
	stashinformer "stash.appscode.dev/apimachinery/client/informers/externalversions"
)

type Controller struct {
	*amc.Controller
	*amc.StashInitializer
	// Namespace to watch
	watchNamespace string
}

func NewController(
	ctrl *amc.Controller,
	initializer *amc.StashInitializer,
	watchNamespace string,
) *Controller {
	return &Controller{
		Controller:       ctrl,
		StashInitializer: initializer,
		watchNamespace:   watchNamespace,
	}
}

type restoreInfo struct {
	invoker core.TypedLocalObjectReference
	target  *v1beta1.RestoreTarget
	phase   v1beta1.RestorePhase
	do      dmcond.DynamicOptions
}

func Configure(cfg *rest.Config, s *amc.StashInitializer, resyncPeriod time.Duration) error {
	var err error
	if s.StashClient, err = scs.NewForConfig(cfg); err != nil {
		return err
	}
	s.StashInformerFactory = stashinformer.NewSharedInformerFactory(s.StashClient, resyncPeriod)
	return nil
}

func (c *Controller) StartAfterStashInstalled(maxNumRequeues, numThreads int, selector metav1.LabelSelector, stopCh <-chan struct{}) {
	// Wait until Stash operator installed
	if err := c.waitUntilStashInstalled(stopCh); err != nil {
		klog.Errorln("error during waiting for RestoreSession crd. Reason: ", err)
		return
	}

	// Initialize the watchers
	err := c.initWatcher(maxNumRequeues, numThreads, selector)
	if err != nil {
		klog.Errorln("Failed to initialize Stash controllers. Reason: ", err)
		return
	}

	// Run the Stash controllers
	c.startController(stopCh)
}

func (c *Controller) initWatcher(maxNumRequeues, numThreads int, selector metav1.LabelSelector) error {
	klog.Infoln("Initializing stash watchers.....")
	// only watch  the restore invokers that matches the selector
	ls, err := metav1.LabelSelectorAsSelector(&selector)
	if err != nil {
		return err
	}
	tweakListOptions := func(options *metav1.ListOptions) {
		options.LabelSelector = ls.String()
	}
	// Initialize RestoreSession Watcher
	c.RSInformer = c.restoreSessionInformer(tweakListOptions)
	c.RSQueue = queue.New(v1beta1.ResourceKindRestoreSession, maxNumRequeues, numThreads, c.processRestoreSession)
	c.RSLister = c.StashInformerFactory.Stash().V1beta1().RestoreSessions().Lister()
	c.RSInformer.AddEventHandler(queue.NewFilteredHandler(queue.NewChangeHandler(c.RSQueue.GetQueue()), ls))

	// Initialize RestoreBatch Watcher
	c.RBInformer = c.restoreBatchInformer(tweakListOptions)
	c.RBQueue = queue.New(v1beta1.ResourceKindRestoreBatch, maxNumRequeues, numThreads, c.processRestoreBatch)
	c.RBLister = c.StashInformerFactory.Stash().V1beta1().RestoreBatches().Lister()
	c.RBInformer.AddEventHandler(queue.NewFilteredHandler(queue.NewChangeHandler(c.RBQueue.GetQueue()), ls))
	return nil
}

func (c *Controller) startController(stopCh <-chan struct{}) {
	klog.Infoln("Starting Stash controllers...")
	// start informer factory
	c.StashInformerFactory.Start(stopCh)
	// wait for cache to sync
	for t, v := range c.StashInformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			klog.Errorf("%v timed out waiting for caches to sync", t)
			return
		}
	}
	// run the queues
	c.RSQueue.Run(stopCh)
	c.RBQueue.Run(stopCh)
}
