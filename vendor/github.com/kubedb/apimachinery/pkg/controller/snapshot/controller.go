package snapshot

import (
	"time"

	apiext_util "github.com/appscode/kutil/apiextensions/v1beta1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/apimachinery/pkg/eventer"
	batch "k8s.io/api/batch/v1"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type Snapshotter interface {
	ValidateSnapshot(*api.Snapshot) error
	GetDatabase(*api.Snapshot) (runtime.Object, error)
	GetSnapshotter(*api.Snapshot) (*batch.Job, error)
	WipeOutSnapshot(*api.Snapshot) error
}

type Controller struct {
	*amc.Controller
	// Snapshotter interface
	snapshotter Snapshotter
	// ListerWatcher
	lw *cache.ListWatch
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
	snapshotter Snapshotter,
	lw *cache.ListWatch,
	syncPeriod time.Duration,
) *Controller {

	// return new DormantDatabase Controller
	return &Controller{
		Controller:     controller,
		snapshotter:    snapshotter,
		lw:             lw,
		eventRecorder:  eventer.NewEventRecorder(controller.Client, "Snapshot Controller"),
		syncPeriod:     syncPeriod,
		maxNumRequests: 5,
	}
}

func (c *Controller) Setup() error {
	crd := []*crd_api.CustomResourceDefinition{
		api.Snapshot{}.CustomResourceDefinition(),
	}
	return apiext_util.RegisterCRDs(c.ApiExtKubeClient, crd)
}

func (c *Controller) Run() {
	// Watch DormantDatabase with provided ListerWatcher
	c.watchSnapshot()
}

func (c *Controller) watchSnapshot() {

	c.initWatcher()

	stop := make(chan struct{})
	defer close(stop)

	c.runWatcher(1, stop)
	select {}
}
