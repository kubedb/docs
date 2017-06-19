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
	// Operator namespace
	OperatorNamespace string
	// Operator tag
	OperatorTag string
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
	promClient pcm.MonitoringV1alpha1Interface
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
	promClient pcm.MonitoringV1alpha1Interface,
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
		eventRecorder:  eventer.NewEventRecorder(client, "Postgres operator"),
		opt:            opt,
		syncPeriod:     time.Minute * 2,
	}
}

func (c *Controller) Run() {
	// Ensure Postgres TPR
	c.ensureThirdPartyResource()

	// Start Cron
	c.cronController.StartCron()

	// Watch Postgres TPR objects
	go c.watchPostgres()
	// Watch Snapshot with labelSelector only for Postgres
	go c.watchSnapshot()
	// Watch DormantDatabase with labelSelector only for Postgres
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

func (c *Controller) watchPostgres() {
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return c.ExtClient.Postgreses(metav1.NamespaceAll).List(metav1.ListOptions{})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.ExtClient.Postgreses(metav1.NamespaceAll).Watch(metav1.ListOptions{})
		},
	}

	_, cacheController := cache.NewInformer(
		lw,
		&tapi.Postgres{},
		c.syncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				postgres := obj.(*tapi.Postgres)
				if postgres.Status.CreationTime == nil {
					if err := c.create(postgres); err != nil {
						postgresFailedToCreate()
						log.Errorln(err)
						c.pushFailureEvent(postgres, err.Error())
					} else {
						postgresSuccessfullyCreated()
					}
				}

			},
			DeleteFunc: func(obj interface{}) {
				if err := c.pause(obj.(*tapi.Postgres)); err != nil {
					postgresFailedToDelete()
					log.Errorln(err)
				} else {
					postgresSuccessfullyDeleted()
				}
			},
			UpdateFunc: func(old, new interface{}) {
				oldObj, ok := old.(*tapi.Postgres)
				if !ok {
					return
				}
				newObj, ok := new.(*tapi.Postgres)
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
		amc.LabelDatabaseKind: tapi.ResourceKindPostgres,
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

	amc.NewSnapshotController(c.Client, c.ExtClient, c, lw, c.syncPeriod).Run()
}

func (c *Controller) watchDormantDatabase() {
	labelMap := map[string]string{
		amc.LabelDatabaseKind: tapi.ResourceKindPostgres,
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

	amc.NewDormantDbController(c.Client, c.ExtClient, c, lw, c.syncPeriod).Run()
}

func (c *Controller) ensureThirdPartyResource() {
	log.Infoln("Ensuring ThirdPartyResource...")

	resourceName := tapi.ResourceNamePostgres + "." + tapi.V1alpha1SchemeGroupVersion.Group
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
		Description: "Postgres Database in Kubernetes by appscode.com",
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

func (c *Controller) pushFailureEvent(postgres *tapi.Postgres, reason string) {
	c.eventRecorder.Eventf(
		postgres,
		apiv1.EventTypeWarning,
		eventer.EventReasonFailedToStart,
		`Fail to be ready Postgres: "%v". Reason: %v`,
		postgres.Name,
		reason,
	)

	var err error
	if postgres, err = c.ExtClient.Postgreses(postgres.Namespace).Get(postgres.Name); err != nil {
		log.Errorln(err)
		return
	}

	postgres.Status.Phase = tapi.DatabasePhaseFailed
	postgres.Status.Reason = reason
	if _, err := c.ExtClient.Postgreses(postgres.Namespace).Update(postgres); err != nil {
		c.eventRecorder.Eventf(
			postgres,
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			`Fail to update Postgres: "%v". Reason: %v`,
			postgres.Name,
			err,
		)
		log.Errorln(err)
	}
}

func postgresSuccessfullyCreated() {
	analytics.SendEvent(tapi.ResourceNamePostgres, "created", "success")
}

func postgresFailedToCreate() {
	analytics.SendEvent(tapi.ResourceNamePostgres, "created", "failure")
}

func postgresSuccessfullyDeleted() {
	analytics.SendEvent(tapi.ResourceNamePostgres, "deleted", "success")
}

func postgresFailedToDelete() {
	analytics.SendEvent(tapi.ResourceNamePostgres, "deleted", "failure")
}
