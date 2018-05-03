package job

import (
	"time"

	"github.com/appscode/kutil/tools/queue"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/apimachinery/pkg/eventer"
	batch "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	batchinformer "k8s.io/client-go/informers/batch/v1"
	"k8s.io/client-go/kubernetes"
	batch_listers "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
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
	// Job
	jobLister batch_listers.JobLister
}

// NewController creates a new Controller
func NewController(
	controller *amc.Controller,
	snapshotter amc.Snapshotter,
	config amc.Config,
	tweakListOptions func(*metav1.ListOptions),
) *Controller {
	// return new DormantDatabase Controller
	return &Controller{
		Controller:       controller,
		snapshotter:      snapshotter,
		Config:           config,
		tweakListOptions: tweakListOptions,
		eventRecorder:    eventer.NewEventRecorder(controller.Client, "Job Controller"),
	}
}

func (c *Controller) InitInformer() cache.SharedIndexInformer {
	return c.KubeInformerFactory.InformerFor(&batch.Job{}, func(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
		return batchinformer.NewFilteredJobInformer(
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
	return c.JobQueue
}
