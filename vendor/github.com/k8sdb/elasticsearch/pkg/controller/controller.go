package controller

import (
	"reflect"
	"time"

	"github.com/appscode/go/hold"
	"github.com/appscode/log"
	pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	tapi "github.com/k8sdb/apimachinery/api"
	tcs "github.com/k8sdb/apimachinery/client/clientset"
	"github.com/k8sdb/apimachinery/pkg/analytics"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
	"github.com/k8sdb/apimachinery/pkg/eventer"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	clientset "k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type Options struct {
	// Tag of elasticdump
	ElasticDumpTag string
	// Tag of elasticsearch operator
	DiscoveryTag string
	// Exporter namespace
	OperatorNamespace      string
	OperatorServiceAccount string
	ExporterTag            string
	// Governing service
	GoverningService string
	// Address to listen on for web interface and telemetry.
	Address string
	// Enable analytics
	EnableAnalytics bool
}

type Controller struct {
	*amc.Controller
	// Prometheus client
	promClient *pcm.MonitoringV1alpha1Client
	// Cron Controller
	cronController amc.CronControllerInterface
	// Event Recorder
	eventRecorder record.EventRecorder
	// Flag data
	opt Options
	// sync time to sync the list.
	syncPeriod time.Duration
}

var _ amc.Snapshotter = &Controller{}
var _ amc.Deleter = &Controller{}

func New(
	client clientset.Interface,
	extClient tcs.ExtensionInterface,
	promClient *pcm.MonitoringV1alpha1Client,
	cronController amc.CronControllerInterface,
	opt Options,
) *Controller {
	return &Controller{
		Controller: &amc.Controller{
			Client:    client,
			ExtClient: extClient,
		},
		promClient:     promClient,
		cronController: cronController,
		eventRecorder:  eventer.NewEventRecorder(client, "Elastic operator"),
		opt:            opt,
		syncPeriod:     time.Minute * 2,
	}
}

func (c *Controller) Run() {
	// Ensure Elastic TPR
	c.ensureThirdPartyResource()

	// Start Cron
	c.cronController.StartCron()

	// Watch Elastic TPR objects
	go c.watchElastic()
	// Watch Snapshot with labelSelector only for Elastic
	go c.watchSnapshot()
	// Watch DormantDatabase with labelSelector only for Elastic
	go c.watchDormantDatabase()
}

// Blocks caller. Intended to be called as a Go routine.
func (c *Controller) RunAndHold() {
	// Enable analytics
	if c.opt.EnableAnalytics {
		analytics.Enable()
	}

	c.Run()

	// Run HTTP server to expose metrics, audit endpoint & debug profiles.
	go c.runHTTPServer()
	// hold
	hold.Hold()
}

func (c *Controller) watchElastic() {
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return c.ExtClient.Elastics(apiv1.NamespaceAll).List(metav1.ListOptions{})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.ExtClient.Elastics(apiv1.NamespaceAll).Watch(metav1.ListOptions{})
		},
	}

	_, cacheController := cache.NewInformer(
		lw,
		&tapi.Elastic{},
		c.syncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				elastic := obj.(*tapi.Elastic)
				if elastic.Status.CreationTime == nil {
					if err := c.create(elastic); err != nil {
						elasticFailedToCreate()
						log.Errorln(err)
						c.pushFailureEvent(elastic, err.Error())
					} else {
						elasticSuccessfullyCreated()
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				if err := c.pause(obj.(*tapi.Elastic)); err != nil {
					elasticFailedToDelete()
					log.Errorln(err)
				} else {
					elasticSuccessfullyDeleted()
				}
			},
			UpdateFunc: func(old, new interface{}) {
				oldObj, ok := old.(*tapi.Elastic)
				if !ok {
					return
				}
				newObj, ok := new.(*tapi.Elastic)
				if !ok {
					return
				}
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
		amc.LabelDatabaseKind: tapi.ResourceKindElastic,
	}
	// Watch with label selector
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return c.ExtClient.Snapshots(apiv1.NamespaceAll).List(
				metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(labelMap).String(),
				})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.ExtClient.Snapshots(apiv1.NamespaceAll).Watch(
				metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(labelMap).String(),
				})
		},
	}

	amc.NewSnapshotController(c.Client, c.ExtClient, c, lw, c.syncPeriod).Run()
}

func (c *Controller) watchDormantDatabase() {
	labelMap := map[string]string{
		amc.LabelDatabaseKind: tapi.ResourceKindElastic,
	}
	// Watch with label selector
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return c.ExtClient.DormantDatabases(apiv1.NamespaceAll).List(
				metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(labelMap).String(),
				})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.ExtClient.DormantDatabases(apiv1.NamespaceAll).Watch(
				metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(labelMap).String(),
				})
		},
	}

	amc.NewDormantDbController(c.Client, c.ExtClient, c, lw, c.syncPeriod).Run()
}

func (c *Controller) ensureThirdPartyResource() {
	log.Infoln("Ensuring ThirdPartyResource...")

	resourceName := tapi.ResourceNameElastic + "." + tapi.V1alpha1SchemeGroupVersion.Group

	if _, err := c.Client.ExtensionsV1beta1().ThirdPartyResources().Get(resourceName, metav1.GetOptions{}); err != nil {
		if !kerr.IsNotFound(err) {
			log.Fatalln(err)
		}
	} else {
		return
	}

	thirdPartyResource := &extensions.ThirdPartyResource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "ThirdPartyResource",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: resourceName,
		},
		Description: "Elasticsearch Database in Kubernetes by appscode.com",
		Versions: []extensions.APIVersion{
			{
				Name: tapi.V1alpha1SchemeGroupVersion.Version,
			},
		},
	}

	if _, err := c.Client.Extensions().ThirdPartyResources().Create(thirdPartyResource); err != nil {
		log.Fatalln(err)
	}
}

func (c *Controller) pushFailureEvent(elastic *tapi.Elastic, reason string) {
	c.eventRecorder.Eventf(
		elastic,
		apiv1.EventTypeWarning,
		eventer.EventReasonFailedToStart,
		`Fail to be ready Elastic: "%v". Reason: %v`,
		elastic.Name,
		reason,
	)

	var err error
	if elastic, err = c.ExtClient.Elastics(elastic.Namespace).Get(elastic.Name); err != nil {
		log.Errorln(err)
		return
	}

	elastic.Status.Phase = tapi.DatabasePhaseFailed
	elastic.Status.Reason = reason
	if _, err := c.ExtClient.Elastics(elastic.Namespace).Update(elastic); err != nil {
		c.eventRecorder.Eventf(
			elastic,
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			`Fail to update Postgres: "%v". Reason: %v`,
			elastic.Name,
			err,
		)
		log.Errorln(err)
	}
}

func elasticSuccessfullyCreated() {
	analytics.SendEvent(tapi.ResourceNameElastic, "created", "success")
}

func elasticFailedToCreate() {
	analytics.SendEvent(tapi.ResourceNameElastic, "created", "failure")
}

func elasticSuccessfullyDeleted() {
	analytics.SendEvent(tapi.ResourceNameElastic, "deleted", "success")
}

func elasticFailedToDelete() {
	analytics.SendEvent(tapi.ResourceNameElastic, "deleted", "failure")
}
