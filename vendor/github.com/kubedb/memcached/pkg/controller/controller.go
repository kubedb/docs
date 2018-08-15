package controller

import (
	"github.com/appscode/go/log"
	apiext_util "github.com/appscode/kutil/apiextensions/v1beta1"
	"github.com/appscode/kutil/tools/queue"
	pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1"
	kutildb "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	api_listers "github.com/kubedb/apimachinery/client/listers/kubedb/v1alpha1"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/apimachinery/pkg/controller/dormantdatabase"
	"github.com/kubedb/apimachinery/pkg/eventer"
	core "k8s.io/api/core/v1"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
)

type Controller struct {
	amc.Config
	*amc.Controller

	// Prometheus client
	promClient pcm.MonitoringV1Interface
	// Event Recorder
	recorder record.EventRecorder
	// labelselector for event-handler of Snapshot, Dormant and Job
	selector labels.Selector

	// Memcached
	mcQueue    *queue.Worker
	mcInformer cache.SharedIndexInformer
	mcLister   api_listers.MemcachedLister
}

var _ amc.Deleter = &Controller{}

func New(
	client kubernetes.Interface,
	apiExtKubeClient crd_cs.ApiextensionsV1beta1Interface,
	extClient cs.KubedbV1alpha1Interface,
	promClient pcm.MonitoringV1Interface,
	opt amc.Config,
) *Controller {
	return &Controller{
		Controller: &amc.Controller{
			Client:           client,
			ExtClient:        extClient,
			ApiExtKubeClient: apiExtKubeClient,
		},
		Config:     opt,
		promClient: promClient,
		recorder:   eventer.NewEventRecorder(client, "Memcached operator"),
		selector: labels.SelectorFromSet(map[string]string{
			api.LabelDatabaseKind: api.ResourceKindMemcached,
		}),
	}
}

// EnsureCustomResourceDefinitions ensures CRD for MySQl, DormantDatabase
func (c *Controller) EnsureCustomResourceDefinitions() error {
	log.Infoln("Ensuring CustomResourceDefinition...")
	crds := []*crd_api.CustomResourceDefinition{
		api.Memcached{}.CustomResourceDefinition(),
		api.MemcachedVersion{}.CustomResourceDefinition(),
		api.DormantDatabase{}.CustomResourceDefinition(),
	}
	return apiext_util.RegisterCRDs(c.ApiExtKubeClient, crds)
}

// InitInformer initializes Memcached, DormantDB amd Snapshot watcher
func (c *Controller) Init() error {
	c.initWatcher()
	c.DrmnQueue = dormantdatabase.NewController(c.Controller, c, c.Config, nil).AddEventHandlerFunc(c.selector)

	return nil
}

// RunControllers runs queue.worker
func (c *Controller) RunControllers(stopCh <-chan struct{}) {
	// Watch x  TPR objects
	c.mcQueue.Run(stopCh)
	c.DrmnQueue.Run(stopCh)
}

// Blocks caller. Intended to be called as a Go routine.
func (c *Controller) Run(stopCh <-chan struct{}) {
	go c.StartAndRunControllers(stopCh)
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
			log.Fatalf("%v timed out waiting for caches to sync", t)
			return
		}
	}
	for t, v := range c.KubedbInformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			log.Fatalf("%v timed out waiting for caches to sync", t)
			return
		}
	}

	c.RunControllers(stopCh)

	<-stopCh
	log.Infoln("Stopping KubeDB controller")
}

func (c *Controller) pushFailureEvent(memcached *api.Memcached, reason string) {
	if ref, rerr := reference.GetReference(clientsetscheme.Scheme, memcached); rerr == nil {
		c.recorder.Eventf(
			ref,
			core.EventTypeWarning,
			eventer.EventReasonFailedToStart,
			`Fail to be ready Memcached: "%v". Reason: %v`,
			memcached.Name,
			reason,
		)
	}

	mc, err := kutildb.UpdateMemcachedStatus(c.ExtClient, memcached, func(in *api.MemcachedStatus) *api.MemcachedStatus {
		in.Phase = api.DatabasePhaseFailed
		in.Reason = reason
		return in
	}, api.EnableStatusSubresource)
	if err != nil {
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, memcached); rerr == nil {
			c.recorder.Eventf(
				ref,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
		}
	}
	memcached.Status = mc.Status
}
