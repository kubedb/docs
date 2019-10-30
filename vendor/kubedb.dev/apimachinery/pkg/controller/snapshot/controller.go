package snapshot

import (
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	kubedb_informers "kubedb.dev/apimachinery/client/informers/externalversions/kubedb/v1alpha1"
	api_listers "kubedb.dev/apimachinery/client/listers/kubedb/v1alpha1"
	amc "kubedb.dev/apimachinery/pkg/controller"
	jobc "kubedb.dev/apimachinery/pkg/controller/job"

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
	// Snapshotter interface
	snapshotter amc.Snapshotter
	// tweakListOptions for watcher
	tweakListOptions func(*metav1.ListOptions)
	// Event Recorder
	eventRecorder record.EventRecorder
	// Snapshot
	snLister api_listers.SnapshotLister
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

func (c *Controller) EnsureCustomResourceDefinitions() error {
	crd := []*crd_api.CustomResourceDefinition{
		api.Snapshot{}.CustomResourceDefinition(),
	}
	return crdutils.RegisterCRDs(c.ApiExtKubeClient, crd)
}

// InitInformer ensures snapshot watcher and returns queue.Worker.
// So, it is possible to start queue.run from other package/repositories
// Return type: snapshotInformer, JobInformer
func (c *Controller) InitInformer() (cache.SharedIndexInformer, cache.SharedIndexInformer) {
	c.SnapInformer = c.KubedbInformerFactory.InformerFor(&api.Snapshot{}, func(client cs.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
		return kubedb_informers.NewFilteredSnapshotInformer(
			client,
			c.WatchNamespace,
			resyncPeriod,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
			c.tweakListOptions,
		)
	})
	c.JobInformer = jobc.NewController(c.Controller, c.snapshotter, c.Config, c.tweakListOptions, c.eventRecorder).InitInformer()
	return c.SnapInformer, c.JobInformer
}

// AddEventHandlerFunc adds EventHandler func. Before calling this,
// controller.Informer needs to be initialized
// Return type: Snapshot queue as 1st parameter and Job.Queue as 2nd.
func (c *Controller) AddEventHandlerFunc(selector labels.Selector) (*queue.Worker, *queue.Worker) {
	c.addEventHandler(selector)
	c.JobQueue = jobc.NewController(c.Controller, c.snapshotter, c.Config, c.tweakListOptions, c.eventRecorder).AddEventHandlerFunc(selector)
	return c.SnapQueue, c.JobQueue
}
