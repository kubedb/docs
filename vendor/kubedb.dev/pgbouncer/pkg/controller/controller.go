/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	kutildb "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"
	api_listers "kubedb.dev/apimachinery/client/listers/kubedb/v1alpha2"
	amc "kubedb.dev/apimachinery/pkg/controller"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/appscode/go/log"
	pcm "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	core "k8s.io/api/core/v1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	reg_util "kmodules.xyz/client-go/admissionregistration/v1beta1"
	"kmodules.xyz/client-go/apiextensions"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
	appcat_util "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
	appcat_listers "kmodules.xyz/custom-resources/client/listers/appcatalog/v1alpha1"
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

	// PgBouncer
	pbQueue    *queue.Worker
	pbInformer cache.SharedIndexInformer
	pbLister   api_listers.PgBouncerLister
	// AppBinding
	appBindingQueue    *queue.Worker
	appBindingInformer cache.SharedIndexInformer
	appBindingLister   appcat_listers.AppBindingLister
}

func New(
	clientConfig *rest.Config,
	client kubernetes.Interface,
	crdClient crd_cs.Interface,
	extClient cs.Interface,
	dc dynamic.Interface,
	appCatalogClient appcat_cs.Interface,
	promClient pcm.MonitoringV1Interface,
	opt amc.Config,
	topology *core_util.Topology,
	recorder record.EventRecorder,
) *Controller {
	return &Controller{
		Controller: &amc.Controller{
			ClientConfig:     clientConfig,
			Client:           client,
			DBClient:         extClient,
			CRDClient:        crdClient,
			DynamicClient:    dc,
			AppCatalogClient: appCatalogClient,
			ClusterTopology:  topology,
		},
		Config:     opt,
		promClient: promClient,
		recorder:   recorder,
		selector: labels.SelectorFromSet(map[string]string{
			api.LabelDatabaseKind: api.ResourceKindPgBouncer,
		}),
	}
}

// Ensuring Custom Resource Definitions
func (c *Controller) EnsureCustomResourceDefinitions() error {
	crds := []*apiextensions.CustomResourceDefinition{
		api.PgBouncer{}.CustomResourceDefinition(),
		catalog.PgBouncerVersion{}.CustomResourceDefinition(),
		appcat_util.AppBinding{}.CustomResourceDefinition(),
	}
	return apiextensions.RegisterCRDs(c.CRDClient, crds)
}

// InitInformer initializes PgBouncer, DormantDB amd Snapshot watcher
func (c *Controller) Init() error {
	c.initWatcher()
	c.initSecretWatcher()
	c.initAppBindingWatcher()

	return nil
}

// RunControllers runs queue.worker
func (c *Controller) RunControllers(stopCh <-chan struct{}) {
	c.pbQueue.Run(stopCh)
	//c.secretQueue.Run(stopCh)
	c.appBindingQueue.Run(stopCh)
}

// Blocks caller. Intended to be called as a Go routine.
func (c *Controller) Run(stopCh <-chan struct{}) {
	go c.StartAndRunControllers(stopCh)

	if c.EnableMutatingWebhook {
		cancel1, _ := reg_util.SyncMutatingWebhookCABundle(c.ClientConfig, mutatingWebhookConfig)
		defer cancel1()
	}
	if c.EnableValidatingWebhook {
		cancel2, _ := reg_util.SyncValidatingWebhookCABundle(c.ClientConfig, validatingWebhookConfig)
		defer cancel2()
	}

	<-stopCh
}

// StartAndRunControllers starts InformetFactory and runs queue.worker
func (c *Controller) StartAndRunControllers(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	c.KubeInformerFactory.Start(stopCh)
	c.KubedbInformerFactory.Start(stopCh)
	c.AppCatInformerFactory.Start(stopCh)
	//c.CertManagerInformerFactory.Start(stopCh)
	c.ExternalInformerFactory.Start(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	for t, v := range c.KubeInformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			log.Fatalf("%v timed out waiting for core caches to sync", t)
			return
		}
	}
	for t, v := range c.KubedbInformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			log.Fatalf("%v timed out waiting for kubedb caches to sync", t)
			return
		}
	}
	for t, v := range c.AppCatInformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			log.Fatalf("%v timed out waiting for appCatalog caches to sync", t)
			return
		}
	}
	for t, v := range c.ExternalInformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			log.Fatalf("%v timed out waiting for external caches to sync", t)
			return
		}
	}

	c.RunControllers(stopCh)

	<-stopCh
}

func (c *Controller) pushFailureEvent(pgbouncer *api.PgBouncer, reason string) {
	c.recorder.Eventf(
		pgbouncer,
		core.EventTypeWarning,
		eventer.EventReasonFailedToStart,
		`Fail to be ready PgBouncer: "%v". Reason: %v`,
		pgbouncer.Name,
		reason,
	)

	pg, err := kutildb.UpdatePgBouncerStatus(
		context.TODO(),
		c.DBClient.KubedbV1alpha2(),
		pgbouncer.ObjectMeta,
		func(in *api.PgBouncerStatus) *api.PgBouncerStatus {
			in.Phase = api.DatabasePhaseNotReady
			in.ObservedGeneration = pgbouncer.Generation
			return in
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		c.recorder.Eventf(
			pgbouncer,
			core.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			err.Error(),
		)
	}
	pgbouncer.Status = pg.Status
}
