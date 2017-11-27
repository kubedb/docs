package controller

import (
	"reflect"
	"time"

	"github.com/appscode/go/hold"
	"github.com/appscode/go/log"
	pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	api "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1"
	kutildb "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1/util"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
	"github.com/k8sdb/apimachinery/pkg/eventer"
	core "k8s.io/api/core/v1"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type Options struct {
	// Operator namespace
	OperatorNamespace string
	// Exporter tag
	ExporterTag string
	// Governing service
	GoverningService string
	// Address to listen on for web interface and telemetry.
	Address string
	// Enable RBAC for database workloads
	EnableRbac bool
}

type Controller struct {
	*amc.Controller
	// Api Extension Client
	ApiExtKubeClient crd_cs.ApiextensionsV1beta1Interface
	// Prometheus client
	promClient pcm.MonitoringV1Interface
	// Cron Controller
	cronController amc.CronControllerInterface
	// Event Recorder
	recorder record.EventRecorder
	// Flag data
	opt Options
	// sync time to sync the list.
	syncPeriod time.Duration
}

var _ amc.Snapshotter = &Controller{}
var _ amc.Deleter = &Controller{}

func New(
	client kubernetes.Interface,
	apiExtKubeClient crd_cs.ApiextensionsV1beta1Interface,
	extClient cs.KubedbV1alpha1Interface,
	promClient pcm.MonitoringV1Interface,
	cronController amc.CronControllerInterface,
	opt Options,
) *Controller {
	return &Controller{
		Controller: &amc.Controller{
			Client:    client,
			ExtClient: extClient,
		},
		ApiExtKubeClient: apiExtKubeClient,
		promClient:       promClient,
		cronController:   cronController,
		recorder:         eventer.NewEventRecorder(client, "MongoDB operator"),
		opt:              opt,
		syncPeriod:       time.Minute * 2,
	}
}

// Blocks caller. Intended to be called as a Go routine.
func (c *Controller) Run() {
	// Ensure TPR
	c.ensureCustomResourceDefinition()

	// Start Cron
	c.cronController.StartCron()

	// Watch x  TPR objects
	go c.watchMongoDB()
	// Watch DatabaseSnapshot with labelSelector only for MongoDB
	go c.watchDatabaseSnapshot()
	// Watch DeletedDatabase with labelSelector only for MongoDB
	go c.watchDeletedDatabase()
}

// Blocks caller. Intended to be called as a Go routine.
func (c *Controller) RunAndHold() {
	c.Run()

	// Run HTTP server to expose metrics, audit endpoint & debug profiles.
	go c.runHTTPServer()
	// hold
	hold.Hold()
}

func (c *Controller) watchMongoDB() {
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return c.ExtClient.MongoDBs(metav1.NamespaceAll).List(metav1.ListOptions{})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.ExtClient.MongoDBs(metav1.NamespaceAll).Watch(metav1.ListOptions{})
		},
	}

	_, cacheController := cache.NewInformer(
		lw,
		&api.MongoDB{},
		c.syncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				mongodb := obj.(*api.MongoDB)
				kutildb.AssignTypeKind(mongodb)
				if mongodb.Status.CreationTime == nil {
					if err := c.create(mongodb); err != nil {
						log.Errorln(err)
						c.pushFailureEvent(mongodb, err.Error())
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				mongodb := obj.(*api.MongoDB)
				kutildb.AssignTypeKind(mongodb)
				if err := c.pause(mongodb); err != nil {
					log.Errorln(err)
				}
			},
			UpdateFunc: func(old, new interface{}) {
				oldObj, ok := old.(*api.MongoDB)
				if !ok {
					return
				}
				newObj, ok := new.(*api.MongoDB)
				if !ok {
					return
				}
				kutildb.AssignTypeKind(oldObj)
				kutildb.AssignTypeKind(newObj)
				if !reflect.DeepEqual(oldObj.Spec, newObj.Spec) {
					if err := c.update(oldObj, newObj); err != nil {
						log.Errorln(err)
					}
				}
			},
		},
	)
	cacheController.Run(wait.NeverStop)
}

func (c *Controller) watchDatabaseSnapshot() {
	labelMap := map[string]string{
		api.LabelDatabaseKind: api.ResourceKindMongoDB,
	}
	// Watch with label selector
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return c.ExtClient.Snapshots(metav1.NamespaceAll).List(
				metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(labelMap).String(),
				})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.ExtClient.Snapshots(metav1.NamespaceAll).Watch(
				metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(labelMap).String(),
				})
		},
	}

	amc.NewSnapshotController(c.Client, c.ApiExtKubeClient, c.ExtClient, c, lw, c.syncPeriod).Run()
}

func (c *Controller) watchDeletedDatabase() {
	labelMap := map[string]string{
		api.LabelDatabaseKind: api.ResourceKindMongoDB,
	}
	// Watch with label selector
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return c.ExtClient.DormantDatabases(metav1.NamespaceAll).List(
				metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(labelMap).String(),
				})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.ExtClient.DormantDatabases(metav1.NamespaceAll).Watch(
				metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(labelMap).String(),
				})
		},
	}

	amc.NewDormantDbController(c.Client, c.ApiExtKubeClient, c.ExtClient, c, lw, c.syncPeriod).Run()
}

func (c *Controller) ensureCustomResourceDefinition() {
	log.Infoln("Ensuring CustomResourceDefinition...")

	resourceName := api.ResourceTypeMongoDB + "." + api.SchemeGroupVersion.Group
	if _, err := c.ApiExtKubeClient.CustomResourceDefinitions().Get(resourceName, metav1.GetOptions{}); err != nil {
		if !kerr.IsNotFound(err) {
			log.Fatalln(err)
		}
	} else {
		return
	}

	crd := &crd_api.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: resourceName,
			Labels: map[string]string{
				"app": "kubedb",
			},
		},
		Spec: crd_api.CustomResourceDefinitionSpec{
			Group:   api.SchemeGroupVersion.Group,
			Version: api.SchemeGroupVersion.Version,
			Scope:   crd_api.NamespaceScoped,
			Names: crd_api.CustomResourceDefinitionNames{
				Plural:     api.ResourceTypeMongoDB,
				Kind:       api.ResourceKindMongoDB,
				ShortNames: []string{api.ResourceCodeMongoDB},
			},
		},
	}

	if _, err := c.ApiExtKubeClient.CustomResourceDefinitions().Create(crd); err != nil {
		log.Fatalln(err)
	}
}

func (c *Controller) pushFailureEvent(mongodb *api.MongoDB, reason string) {
	c.recorder.Eventf(
		mongodb.ObjectReference(),
		core.EventTypeWarning,
		eventer.EventReasonFailedToStart,
		`Fail to be ready MongoDB: "%v". Reason: %v`,
		mongodb.Name,
		reason,
	)

	_, err := kutildb.TryPatchMongoDB(c.ExtClient, mongodb.ObjectMeta, func(in *api.MongoDB) *api.MongoDB {
		in.Status.Phase = api.DatabasePhaseFailed
		in.Status.Reason = reason
		return in
	})
	if err != nil {
		c.recorder.Eventf(mongodb.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
	}
}
