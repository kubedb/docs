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
	"kubedb.dev/apimachinery/pkg/controller/initializer/stash"

	"github.com/appscode/go/log"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	reg_util "kmodules.xyz/client-go/admissionregistration/v1beta1"
)

// Blocks caller. Intended to be called as a Go routine.
func (c *Controller) Run(stopCh <-chan struct{}) {
	go c.StartAndRunControllers(stopCh)

	if c.EnableMutatingWebhook {
		cancel1, _ := reg_util.SyncMutatingWebhookCABundle(c.ClientConfig, mutatingWebhookConfig)
		defer cancel1()
	}
	if c.EnableValidatingWebhook {
		cancel2, _ := reg_util.SyncValidatingWebhookCABundle(c.ClientConfig, validatingWebhookConfig)
		defer cancel2()
	}

	<-stopCh
}

// StartAndRunControllers starts InformetFactory and runs queue.worker
func (c *Controller) StartAndRunControllers(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	log.Infoln("Starting KubeDB controller")
	c.KubeInformerFactory.Start(stopCh)
	c.KubedbInformerFactory.Start(stopCh)

	// Run Stash initializer controllers
	go stash.NewController(c.esCtrl.Controller, &c.esCtrl.Config.Initializers.Stash, c.esCtrl, c.Recorder, c.WatchNamespace).StartController(stopCh)
	go stash.NewController(c.mgCtrl.Controller, &c.mgCtrl.Config.Initializers.Stash, c.mgCtrl, c.Recorder, c.WatchNamespace).StartController(stopCh)
	go stash.NewController(c.myCtrl.Controller, &c.myCtrl.Config.Initializers.Stash, c.myCtrl, c.Recorder, c.WatchNamespace).StartController(stopCh)
	go stash.NewController(c.pxCtrl.Controller, &c.pxCtrl.Config.Initializers.Stash, c.pxCtrl, c.Recorder, c.WatchNamespace).StartController(stopCh)
	go stash.NewController(c.pgCtrl.Controller, &c.pgCtrl.Config.Initializers.Stash, c.pgCtrl, c.Recorder, c.WatchNamespace).StartController(stopCh)
	go stash.NewController(c.rdCtrl.Controller, &c.rdCtrl.Config.Initializers.Stash, c.rdCtrl, c.Recorder, c.WatchNamespace).StartController(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	for t, v := range c.KubeInformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			log.Fatalf("%v timed out waiting for caches to sync\n", t)
			return
		}
	}
	for t, v := range c.KubedbInformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			log.Fatalf("%v timed out waiting for caches to sync\n", t)
			return
		}
	}

	c.esCtrl.RunControllers(stopCh)
	c.mcCtrl.RunControllers(stopCh)
	c.mgCtrl.RunControllers(stopCh)
	c.myCtrl.RunControllers(stopCh)
	c.pgCtrl.RunControllers(stopCh)
	c.prCtrl.RunControllers(stopCh)
	c.pxCtrl.RunControllers(stopCh)
	c.rdCtrl.RunControllers(stopCh)

	<-stopCh
	log.Infoln("Stopping KubeDB controller")
}
