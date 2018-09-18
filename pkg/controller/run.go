package controller

import (
	"net/http"

	"github.com/appscode/go/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

// Blocks caller. Intended to be called as a Go routine.
func (c *Controller) Run(stopCh <-chan struct{}) {
	go c.StartAndRunControllers(stopCh)

	<-stopCh
	c.cronController.StopCron()
}

const opsAdress = ":8080"

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
	//c.edCtrl.RunControllers(stopCh)
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
