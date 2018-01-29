package job

import (
	"time"

	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/apimachinery/pkg/eventer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	*amc.Controller
	// SnapshotDoer interface
	snapshotter amc.Snapshotter
	// ListOptions for watcher
	listOption metav1.ListOptions
	// Event Recorder
	eventRecorder record.EventRecorder
	// sync time to sync the list.
	syncPeriod time.Duration
	// Workqueue
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
	//Max number requests for retries
	maxNumRequests int
}

// NewController creates a new Controller
func NewController(
	controller *amc.Controller,
	snapshotter amc.Snapshotter,
	listOption metav1.ListOptions,
	syncPeriod time.Duration,
) *Controller {

	// return new DormantDatabase Controller
	return &Controller{
		Controller:     controller,
		snapshotter:    snapshotter,
		listOption:     listOption,
		eventRecorder:  eventer.NewEventRecorder(controller.Client, "Job Controller"),
		syncPeriod:     syncPeriod,
		maxNumRequests: 5,
	}
}

func (c *Controller) Run() {
	// Watch DormantDatabase with provided ListerWatcher
	c.watchJob()
}

func (c *Controller) watchJob() {

	c.initWatcher()

	stop := make(chan struct{})
	defer close(stop)

	c.runWatcher(5, stop)
	select {}
}
