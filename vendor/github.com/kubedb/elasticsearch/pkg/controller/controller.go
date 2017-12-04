package controller

import (
	"reflect"
	"time"

	"github.com/appscode/go/hold"
	"github.com/appscode/go/log"
	pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1"
	kutildb "github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1/util"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/apimachinery/pkg/eventer"
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
	// Tag of elasticdump
	ElasticDumpTag string
	// Exporter namespace
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
		recorder:         eventer.NewEventRecorder(client, "Elasticsearch operator"),
		opt:              opt,
		syncPeriod:       time.Minute * 2,
	}
}

func (c *Controller) Run() {
	// Ensure Elasticsearch CRD
	c.ensureCustomResourceDefinition()

	// Start Cron
	c.cronController.StartCron()

	// Watch Elasticsearch TPR objects
	go c.watchElastic()
	// Watch Snapshot with labelSelector only for Elasticsearch
	go c.watchSnapshot()
	// Watch DormantDatabase with labelSelector only for Elasticsearch
	go c.watchDormantDatabase()
}

// Blocks caller. Intended to be called as a Go routine.
func (c *Controller) RunAndHold() {
	c.Run()

	// Run HTTP server to expose metrics, audit endpoint & debug profiles.
	go c.runHTTPServer()
	// hold
	hold.Hold()
}

func (c *Controller) watchElastic() {
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return c.ExtClient.Elasticsearchs(core.NamespaceAll).List(metav1.ListOptions{})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.ExtClient.Elasticsearchs(core.NamespaceAll).Watch(metav1.ListOptions{})
		},
	}

	_, cacheController := cache.NewInformer(
		lw,
		&api.Elasticsearch{},
		c.syncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				elastic := obj.(*api.Elasticsearch)
				kutildb.AssignTypeKind(elastic)
				if elastic.Status.CreationTime == nil {
					if err := c.create(elastic); err != nil {
						log.Errorln(err)
						c.pushFailureEvent(elastic, err.Error())
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				elastic := obj.(*api.Elasticsearch)
				kutildb.AssignTypeKind(elastic)
				if err := c.pause(elastic); err != nil {
					log.Errorln(err)
				}
			},
			UpdateFunc: func(old, new interface{}) {
				oldObj, ok := old.(*api.Elasticsearch)
				if !ok {
					return
				}
				newObj, ok := new.(*api.Elasticsearch)
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

func (c *Controller) watchSnapshot() {
	labelMap := map[string]string{
		api.LabelDatabaseKind: api.ResourceKindElasticsearch,
	}
	// Watch with label selector
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return c.ExtClient.Snapshots(core.NamespaceAll).List(
				metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(labelMap).String(),
				})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.ExtClient.Snapshots(core.NamespaceAll).Watch(
				metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(labelMap).String(),
				})
		},
	}

	amc.NewSnapshotController(c.Client, c.ApiExtKubeClient, c.ExtClient, c, lw, c.syncPeriod).Run()
}

func (c *Controller) watchDormantDatabase() {
	labelMap := map[string]string{
		api.LabelDatabaseKind: api.ResourceKindElasticsearch,
	}
	// Watch with label selector
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return c.ExtClient.DormantDatabases(core.NamespaceAll).List(
				metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(labelMap).String(),
				})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.ExtClient.DormantDatabases(core.NamespaceAll).Watch(
				metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(labelMap).String(),
				})
		},
	}

	amc.NewDormantDbController(c.Client, c.ApiExtKubeClient, c.ExtClient, c, lw, c.syncPeriod).Run()
}

func (c *Controller) ensureCustomResourceDefinition() {
	log.Infoln("Ensuring CustomResourceDefinition...")

	resourceName := api.ResourceTypeElasticsearch + "." + api.SchemeGroupVersion.Group
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
				Plural:     api.ResourceTypeElasticsearch,
				Kind:       api.ResourceKindElasticsearch,
				ShortNames: []string{api.ResourceCodeElasticsearch},
			},
		},
	}

	if _, err := c.ApiExtKubeClient.CustomResourceDefinitions().Create(crd); err != nil {
		log.Fatalln(err)
	}
}

func (c *Controller) pushFailureEvent(elasticsearch *api.Elasticsearch, reason string) {
	c.recorder.Eventf(
		elasticsearch.ObjectReference(),
		core.EventTypeWarning,
		eventer.EventReasonFailedToStart,
		`Fail to be ready Elasticsearch: "%v". Reason: %v`,
		elasticsearch.Name,
		reason,
	)

	var err error
	if elasticsearch, err = c.ExtClient.Elasticsearchs(elasticsearch.Namespace).Get(elasticsearch.Name, metav1.GetOptions{}); err != nil {
		log.Errorln(err)
		return
	}

	es, err := kutildb.PatchElasticsearch(c.ExtClient, elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
		in.Status.Phase = api.DatabasePhaseFailed
		in.Status.Reason = reason
		return in
	})
	if err != nil {
		c.recorder.Eventf(elasticsearch.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
	}
	*elasticsearch = *es
}
