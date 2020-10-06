/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package statefulset

import (
	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	db_cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amc "kubedb.dev/apimachinery/pkg/controller"

	"github.com/appscode/go/log"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	core_util "kmodules.xyz/client-go/core/v1"
	dmcond "kmodules.xyz/client-go/dynamic/conditions"
	"kmodules.xyz/client-go/tools/queue"
)

type Controller struct {
	*amc.Controller
	*amc.Config
}

func NewController(
	config *amc.Config,
	client kubernetes.Interface,
	dbClient db_cs.Interface,
	dmClient dynamic.Interface,
) *Controller {
	return &Controller{
		Controller: &amc.Controller{
			Client:        client,
			DBClient:      dbClient,
			DynamicClient: dmClient,
		},
		Config: config,
	}
}

type databaseInfo struct {
	do            dmcond.DynamicOptions
	replicasReady bool
	msg           string
}

func (c *Controller) InitStsWatcher() {
	log.Infoln("Initializing StatefulSet watcher.....")
	// Initialize RestoreSession Watcher
	c.StsInformer = c.KubeInformerFactory.Apps().V1().StatefulSets().Informer()
	c.StsQueue = queue.New(api.ResourceKindStatefulSet, c.MaxNumRequeues, c.NumThreads, c.processStatefulSet)
	c.StsLister = c.KubeInformerFactory.Apps().V1().StatefulSets().Lister()
	c.StsInformer.AddEventHandler(c.newStsEventHandlerFuncs())
}

func (c *Controller) enqueueOnlyKubeDBSts(sts *appsv1.StatefulSet) {
	// only enqueue if the controlling owner is a KubeDB resource
	ok, _, err := core_util.IsOwnerOfGroup(metav1.GetControllerOf(sts), kubedb.GroupName)
	if err != nil {
		log.Warningln(err)
		return
	}
	if key, err := cache.MetaNamespaceKeyFunc(sts); ok && err == nil {
		queue.Enqueue(c.StsQueue.GetQueue(), key)
	}
}
