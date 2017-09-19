package controller

import (
	"reflect"
	"time"

	"github.com/appscode/go/hold"
	"github.com/appscode/go/log"
	kutildb "github.com/appscode/kutil/kubedb/v1alpha1"
	pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	tcs "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
	"github.com/k8sdb/apimachinery/pkg/eventer"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	clientset "k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
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
	ApiExtKubeClient apiextensionsclient.Interface
	// Prometheus client
	promClient pcm.MonitoringV1alpha1Interface
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
	client clientset.Interface,
	apiExtKubeClient apiextensionsclient.Interface,
	extClient tcs.KubedbV1alpha1Interface,
	promClient pcm.MonitoringV1alpha1Interface,
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
		recorder:         eventer.NewEventRecorder(client, "Postgres operator"),
		opt:              opt,
		syncPeriod:       time.Minute * 2,
	}
}

func (c *Controller) Run() {
	// Ensure Postgres CRD
	c.ensureCustomResourceDefinition()

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
						log.Errorln(err)
						c.pushFailureEvent(postgres, err.Error())
					}
				}

			},
			DeleteFunc: func(obj interface{}) {
				if err := c.pause(obj.(*tapi.Postgres)); err != nil {
					log.Errorln(err)
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
		tapi.LabelDatabaseKind: tapi.ResourceKindPostgres,
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

func (c *Controller) watchDormantDatabase() {
	labelMap := map[string]string{
		tapi.LabelDatabaseKind: tapi.ResourceKindPostgres,
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

	resourceName := tapi.ResourceTypePostgres + "." + tapi.SchemeGroupVersion.Group
	if _, err := c.ApiExtKubeClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(resourceName, metav1.GetOptions{}); err != nil {
		if !kerr.IsNotFound(err) {
			log.Fatalln(err)
		}
	} else {
		return
	}

	crd := &extensionsobj.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: resourceName,
			Labels: map[string]string{
				"app": "kubedb",
			},
		},
		Spec: extensionsobj.CustomResourceDefinitionSpec{
			Group:   tapi.SchemeGroupVersion.Group,
			Version: tapi.SchemeGroupVersion.Version,
			Scope:   extensionsobj.NamespaceScoped,
			Names: extensionsobj.CustomResourceDefinitionNames{
				Plural:     tapi.ResourceTypePostgres,
				Kind:       tapi.ResourceKindPostgres,
				ShortNames: []string{tapi.ResourceCodePostgres},
			},
		},
	}

	if _, err := c.ApiExtKubeClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd); err != nil {
		log.Fatalln(err)
	}
}

func (c *Controller) pushFailureEvent(postgres *tapi.Postgres, reason string) {
	c.recorder.Eventf(
		postgres.ObjectReference(),
		apiv1.EventTypeWarning,
		eventer.EventReasonFailedToStart,
		`Fail to be ready Postgres: "%v". Reason: %v`,
		postgres.Name,
		reason,
	)

	_, err := kutildb.TryPatchPostgres(c.ExtClient, postgres.ObjectMeta, func(in *tapi.Postgres) *tapi.Postgres {
		in.Status.Phase = tapi.DatabasePhaseFailed
		in.Status.Reason = reason
		return in
	})
	if err != nil {
		c.recorder.Eventf(postgres.ObjectReference(), apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
	}
}
