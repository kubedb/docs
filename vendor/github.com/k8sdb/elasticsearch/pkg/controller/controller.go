package controller

import (
	"reflect"
	"time"

	"github.com/appscode/go/hold"
	"github.com/appscode/log"
	pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	tapi "github.com/k8sdb/apimachinery/api"
	tcs "github.com/k8sdb/apimachinery/client/clientset"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
	"github.com/k8sdb/apimachinery/pkg/eventer"
	kapi "k8s.io/kubernetes/pkg/api"
	k8serr "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/cache"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/record"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/watch"
)

type Controller struct {
	*amc.Controller
	// Cron Controller
	cronController amc.CronControllerInterface
	// Prometheus client
	promClient *pcm.MonitoringV1alpha1Client
	// Event Recorder
	eventRecorder record.EventRecorder
	// Tag of elasticsearch opearator
	operatorTag string
	// Tag of elasticdump
	elasticDumpTag string
	// Governing service
	governingService string
	// sync time to sync the list.
	syncPeriod time.Duration
}

var _ amc.Snapshotter = &Controller{}
var _ amc.Deleter = &Controller{}

func New(
	client clientset.Interface,
	extClient tcs.ExtensionInterface,
	promClient *pcm.MonitoringV1alpha1Client,
	operatorTag string,
	elasticDumpTag string,
	governingService string,
) *Controller {
	return &Controller{
		Controller: &amc.Controller{
			Client:    client,
			ExtClient: extClient,
		},
		cronController:   amc.NewCronController(client, extClient),
		promClient:       promClient,
		eventRecorder:    eventer.NewEventRecorder(client, "Elastic Controller"),
		operatorTag:      operatorTag,
		elasticDumpTag:   elasticDumpTag,
		governingService: governingService,
		syncPeriod:       time.Minute * 2,
	}
}

// Blocks caller. Intended to be called as a Go routine.
func (c *Controller) RunAndHold() {
	// Ensure Elastic TPR
	c.ensureThirdPartyResource()

	// Start Cron
	c.cronController.StartCron()
	// Stop Cron
	defer c.cronController.StopCron()

	// Watch Elastic TPR objects
	go c.watchElastic()
	// Watch Snapshot with labelSelector only for Elastic
	go c.watchSnapshot()
	// Watch DormantDatabase with labelSelector only for Elastic
	go c.watchDormantDatabase()
	// hold
	hold.Hold()
}

func (c *Controller) watchElastic() {
	lw := &cache.ListWatch{
		ListFunc: func(opts kapi.ListOptions) (runtime.Object, error) {
			return c.ExtClient.Elastics(kapi.NamespaceAll).List(kapi.ListOptions{})
		},
		WatchFunc: func(options kapi.ListOptions) (watch.Interface, error) {
			return c.ExtClient.Elastics(kapi.NamespaceAll).Watch(kapi.ListOptions{})
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
						log.Errorln(err)
						c.pushFailureEvent(elastic, err.Error())
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				if err := c.pause(obj.(*tapi.Elastic)); err != nil {
					log.Errorln(err)
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
		ListFunc: func(opts kapi.ListOptions) (runtime.Object, error) {
			return c.ExtClient.Snapshots(kapi.NamespaceAll).List(
				kapi.ListOptions{
					LabelSelector: labels.SelectorFromSet(labels.Set(labelMap)),
				})
		},
		WatchFunc: func(options kapi.ListOptions) (watch.Interface, error) {
			return c.ExtClient.Snapshots(kapi.NamespaceAll).Watch(
				kapi.ListOptions{
					LabelSelector: labels.SelectorFromSet(labels.Set(labelMap)),
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
		ListFunc: func(opts kapi.ListOptions) (runtime.Object, error) {
			return c.ExtClient.DormantDatabases(kapi.NamespaceAll).List(
				kapi.ListOptions{
					LabelSelector: labels.SelectorFromSet(labels.Set(labelMap)),
				})
		},
		WatchFunc: func(options kapi.ListOptions) (watch.Interface, error) {
			return c.ExtClient.DormantDatabases(kapi.NamespaceAll).Watch(
				kapi.ListOptions{
					LabelSelector: labels.SelectorFromSet(labels.Set(labelMap)),
				})
		},
	}

	amc.NewDormantDbController(c.Client, c.ExtClient, c, lw, c.syncPeriod).Run()
}

func (c *Controller) ensureThirdPartyResource() {
	log.Infoln("Ensuring ThirdPartyResource...")

	resourceName := tapi.ResourceNameElastic + "." + tapi.V1beta1SchemeGroupVersion.Group

	if _, err := c.Client.Extensions().ThirdPartyResources().Get(resourceName); err != nil {
		if !k8serr.IsNotFound(err) {
			log.Fatalln(err)
		}
	} else {
		return
	}

	thirdPartyResource := &extensions.ThirdPartyResource{
		TypeMeta: unversioned.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "ThirdPartyResource",
		},
		ObjectMeta: kapi.ObjectMeta{
			Name: resourceName,
		},
		Versions: []extensions.APIVersion{
			{
				Name: tapi.V1beta1SchemeGroupVersion.Version,
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
		kapi.EventTypeWarning,
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
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			`Fail to update Postgres: "%v". Reason: %v`,
			elastic.Name,
			err,
		)
		log.Errorln(err)
	}
}
