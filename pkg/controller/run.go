package controller

import (
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

	c.cronController.StopCron()
}

// StartAndRunControllers starts InformetFactory and runs queue.worker
func (c *Controller) StartAndRunControllers(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	log.Infoln("Starting KubeDB controller")
	c.KubeInformerFactory.Start(stopCh)
	c.KubedbInformerFactory.Start(stopCh)

	go func() {
		// start StashInformerFactory only if stash crds (ie, "restoreSession") are available.
		if err := c.BlockOnStashOperator(stopCh); err != nil {
			log.Errorln("error while waiting for restoreSession.", err)
			return
		}

		// start informer factory
		c.StashInformerFactory.Start(stopCh)
		for t, v := range c.StashInformerFactory.WaitForCacheSync(stopCh) {
			if !v {
				log.Fatalf("%v timed out waiting for caches to sync", t)
				return
			}
		}
		// Only postgres, elasticsearch, mongodb and mysql has restoreSession queue initialized.
		// Check RSQueue initialization in ctrl.init() (e.g. c.myCtrl.Init()) to know if it expects RS watcher.
		c.pgCtrl.RSQueue.Run(stopCh)
		c.esCtrl.RSQueue.Run(stopCh)
		c.mgCtrl.RSQueue.Run(stopCh)
		c.myCtrl.RSQueue.Run(stopCh)
	}()

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

	c.pgCtrl.RunControllers(stopCh)
	c.esCtrl.RunControllers(stopCh)
	c.edCtrl.RunControllers(stopCh)
	c.mgCtrl.RunControllers(stopCh)
	c.myCtrl.RunControllers(stopCh)
	c.rdCtrl.RunControllers(stopCh)
	c.mcCtrl.RunControllers(stopCh)

	<-stopCh
	log.Infoln("Stopping KubeDB controller")
}
