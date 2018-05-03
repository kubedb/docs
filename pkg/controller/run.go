package controller

import (
	"fmt"
	"net/http"

	"github.com/appscode/go/log"
	"github.com/appscode/pat"
	"github.com/kubedb/operator/pkg/exporter"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

// Blocks caller. Intended to be called as a Go routine.
func (c *Controller) Run(stopCh <-chan struct{}) {
	go c.StartAndRunControllers(stopCh)

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
	c.mgCtrl.RunControllers(stopCh)
	c.myCtrl.RunControllers(stopCh)
	c.rdCtrl.RunControllers(stopCh)
	c.mcCtrl.RunControllers(stopCh)

	// For database summary report
	ex := exporter.New("", "", ":8080", c.Client, c.ExtClient)
	m := pat.New()
	auditPattern := fmt.Sprintf("/kubedb.com/v1alpha1/namespaces/%s/%s/%s/report", exporter.PathParamNamespace, exporter.PathParamType, exporter.PathParamName)
	log.Infoln("Report URL pattern:", auditPattern)
	m.Get(auditPattern, http.HandlerFunc(ex.ExportSummaryReport))

	http.Handle("/", m)
	log.Infof("Starting Server: %s", ex.Address)
	log.Fatal(http.ListenAndServe(ex.Address, nil))

	<-stopCh
	log.Infoln("Stopping KubeDB controller")
}
