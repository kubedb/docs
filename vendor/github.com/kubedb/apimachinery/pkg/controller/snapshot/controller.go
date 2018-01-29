package snapshot

import (
	"time"

	apiext_util "github.com/appscode/kutil/apiextensions/v1beta1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	jobc "github.com/kubedb/apimachinery/pkg/controller/job"
	"github.com/kubedb/apimachinery/pkg/eventer"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	*amc.Controller
	// Snapshotter interface
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
	// Watch Snapshot with provided ListOption
	go c.watchSnapshot()
	// Watch Job with provided ListOption
	go jobc.NewController(c.Controller, c.snapshotter, c.listOption, c.syncPeriod).Run()
}

func (c *Controller) watchSnapshot() {
	c.initWatcher()

	stop := make(chan struct{})
	defer close(stop)

	c.runWatcher(5, stop)
	select {}
}
