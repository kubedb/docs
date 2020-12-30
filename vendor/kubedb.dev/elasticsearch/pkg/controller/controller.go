/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	catalog_lister "kubedb.dev/apimachinery/client/listers/catalog/v1alpha1"
	api_listers "kubedb.dev/apimachinery/client/listers/kubedb/v1alpha2"
	amc "kubedb.dev/apimachinery/pkg/controller"
	"kubedb.dev/apimachinery/pkg/controller/initializer/stash"
	"kubedb.dev/apimachinery/pkg/eventer"

	pcm "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"gomodules.xyz/x/log"
	core "k8s.io/api/core/v1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilRuntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	reg_util "kmodules.xyz/client-go/admissionregistration/v1beta1"
	"kmodules.xyz/client-go/apiextensions"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/client-go/tools/queue"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
)

type Controller struct {
	amc.Config
	*amc.Controller

	// Prometheus client
	promClient pcm.MonitoringV1Interface
	// LabelSelector to filter Stash restore invokers only for this database
	selector metav1.LabelSelector

	// Elasticsearch
	esQueue         *queue.Worker
	esInformer      cache.SharedIndexInformer
	esLister        api_listers.ElasticsearchLister
	esVersionLister catalog_lister.ElasticsearchVersionLister
}

func New(
	restConfig *restclient.Config,
	client kubernetes.Interface,
	crdClient crd_cs.Interface,
	dbClient cs.Interface,
	dc dynamic.Interface,
	appCatalogClient appcat_cs.Interface,
	promClient pcm.MonitoringV1Interface,
	amcConfig amc.Config,
	topology *core_util.Topology,
	recorder record.EventRecorder,
) *Controller {
	return &Controller{
		Controller: &amc.Controller{
			ClientConfig:     restConfig,
			Client:           client,
			DBClient:         dbClient,
			CRDClient:        crdClient,
			DynamicClient:    dc,
			AppCatalogClient: appCatalogClient,
			ClusterTopology:  topology,
			Mapper:           restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(client.Discovery())),
			Recorder:         recorder,
		},
		Config:     amcConfig,
		promClient: promClient,
		selector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				meta_util.NameLabelKey:      api.Elasticsearch{}.ResourceFQN(),
				meta_util.ManagedByLabelKey: kubedb.GroupName,
			},
		},
	}
}

// Ensuring Custom Resources Definitions
func (c *Controller) EnsureCustomResourceDefinitions() error {
	log.Infoln("Ensuring CustomResourceDefinition...")
	crds := []*apiextensions.CustomResourceDefinition{
		api.Elasticsearch{}.CustomResourceDefinition(),
		catalog.ElasticsearchVersion{}.CustomResourceDefinition(),
		appcat.AppBinding{}.CustomResourceDefinition(),
	}
	return apiextensions.RegisterCRDs(c.CRDClient, crds)
}

// InitInformer initializes Elasticsearch, DormantDB amd Snapshot watcher
func (c *Controller) Init() error {
	c.initWatcher()
	c.initSecretWatcher()
	return nil
}

// RunControllers runs queue.worker
func (c *Controller) RunControllers(stopCh <-chan struct{}) {
	// Start Elasticsearch controller
	c.esQueue.Run(stopCh)

	// Start Elasticsearch health checker
	c.RunHealthChecker(stopCh)
}

// Blocks caller. Intended to be called as a Go routine.
func (c *Controller) Run(stopCh <-chan struct{}) {
	go c.StartAndRunControllers(stopCh)

	<-stopCh
}

// StartAndRunControllers starts InformetFactory and runs queue.worker
func (c *Controller) StartAndRunControllers(stopCh <-chan struct{}) {
	defer utilRuntime.HandleCrash()

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

	// Start StatefulSet controller
	c.StsQueue.Run(stopCh)

	// Initialize and start Stash controllers
	go stash.NewController(c.Controller, &c.Config.Initializers.Stash, c.WatchNamespace).StartAfterStashInstalled(c.MaxNumRequeues, c.NumThreads, c.selector, stopCh)

	// Start Elasticsearch controller
	c.RunControllers(stopCh)

	if c.EnableMutatingWebhook {
		cancel1, _ := reg_util.SyncMutatingWebhookCABundle(c.ClientConfig, mutatingWebhookConfig)
		defer cancel1()
	}
	if c.EnableValidatingWebhook {
		cancel2, _ := reg_util.SyncValidatingWebhookCABundle(c.ClientConfig, validatingWebhookConfig)
		defer cancel2()
	}

	<-stopCh
	log.Infoln("Stopping KubeDB controller")
}

func (c *Controller) pushFailureEvent(db *api.Elasticsearch, reason string) {
	c.Recorder.Eventf(
		db,
		core.EventTypeWarning,
		eventer.EventReasonFailedToStart,
		`Fail to be ready Elasticsearch: "%v". Reason: %v`,
		db.Name,
		reason,
	)
}
