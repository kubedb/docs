package controller

import (
	"net/http"

	"github.com/appscode/go/log"
	reg_util "github.com/appscode/kutil/admissionregistration/v1beta1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

const (
	opsAdress         = ":8080"
	mutatingWebhook   = "mutators.kubedb.com"
	validatingWebhook = "validators.kubedb.com"
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

	http.Handle("/metrics", promhttp.Handler())
	log.Infof("Starting Server: %s", opsAdress)
	log.Fatal(http.ListenAndServe(opsAdress, nil))

	<-stopCh
	log.Infoln("Stopping KubeDB controller")
}
