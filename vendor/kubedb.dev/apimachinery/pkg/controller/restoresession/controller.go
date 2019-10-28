package restoresession

import (
	"time"

	amc "kubedb.dev/apimachinery/pkg/controller"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"kmodules.xyz/client-go/tools/queue"
	"stash.appscode.dev/stash/apis/stash/v1beta1"
	scs "stash.appscode.dev/stash/client/clientset/versioned"
	stashinformers "stash.appscode.dev/stash/client/informers/externalversions/stash/v1beta1"
	stashLister "stash.appscode.dev/stash/client/listers/stash/v1beta1"
)

type Controller struct {
	*amc.Controller
	amc.Config
	// SnapshotDoer interface
	snapshotter amc.Snapshotter
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
	snapshotter amc.Snapshotter,
	config amc.Config,
	tweakListOptions func(*metav1.ListOptions),
	eventRecorder record.EventRecorder,
) *Controller {
	// return new DormantDatabase Controller
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
