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
	"fmt"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	db_cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amc "kubedb.dev/apimachinery/pkg/controller"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

type Controller struct {
	*amc.Config

	// Kubernetes client
	Client kubernetes.Interface
	// KubeDB client
	DBClient db_cs.Interface
	// Dynamic client
	DynamicClient dynamic.Interface
}

func NewController(
	config *amc.Config,
	client kubernetes.Interface,
	dbClient db_cs.Interface,
	dmClient dynamic.Interface,
) *Controller {
	return &Controller{
		Config:        config,
		Client:        client,
		DBClient:      dbClient,
		DynamicClient: dmClient,
	}
}

func (c *Controller) InitStsWatcher() {
	klog.Infoln("Initializing StatefulSet watcher.....")
	// Initialize RestoreSession Watcher
	c.StsInformer = c.KubeInformerFactory.Apps().V1().StatefulSets().Informer()
	c.StsQueue = queue.New(api.ResourceKindStatefulSet, c.MaxNumRequeues, c.NumThreads, c.processStatefulSet)
	c.StsLister = c.KubeInformerFactory.Apps().V1().StatefulSets().Lister()
	c.StsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if sts, ok := obj.(*apps.StatefulSet); ok {
				if c.RestrictToNamespace != core.NamespaceAll {
					if sts.GetNamespace() != c.RestrictToNamespace {
						klog.Info("Skipping StatefulSet %s/%s. Only %s namespace is supported for Community Edition. Please upgrade to Enterprise to use any namespace.", sts.GetNamespace(), sts.GetName(), c.RestrictToNamespace)
						return
					}
				}

				c.enqueueOnlyKubeDBSts(sts)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if sts, ok := newObj.(*apps.StatefulSet); ok {
				if c.RestrictToNamespace != core.NamespaceAll {
					if sts.GetNamespace() != c.RestrictToNamespace {
						klog.Info("Skipping StatefulSet %s/%s. Only %s namespace is supported for Community Edition. Please upgrade to Enterprise to use any namespace.", sts.GetNamespace(), sts.GetName(), c.RestrictToNamespace)
						return
					}
				}

				c.enqueueOnlyKubeDBSts(sts)
			}
		},
		DeleteFunc: func(obj interface{}) {
			var sts *apps.StatefulSet
			var ok bool
			if sts, ok = obj.(*apps.StatefulSet); !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					klog.V(5).Info("error decoding object, invalid type")
					return
				}
				sts, ok = tombstone.Obj.(*apps.StatefulSet)
				if !ok {
					utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
					return
				}
				klog.V(5).Infof("Recovered deleted object '%v' from tombstone", tombstone.Obj.(metav1.Object).GetName())
			}

			if c.RestrictToNamespace != core.NamespaceAll {
				if sts.GetNamespace() != c.RestrictToNamespace {
					klog.Info("Skipping StatefulSet %s/%s. Only %s namespace is supported for Community Edition. Please upgrade to Enterprise to use any namespace.", sts.GetNamespace(), sts.GetName(), c.RestrictToNamespace)
					return
				}
			}

			ok, _, err := core_util.IsOwnerOfGroup(metav1.GetControllerOf(sts), kubedb.GroupName)
			if err != nil || !ok {
				klog.Warningln(err)
				return
			}
			dbInfo, err := c.extractDatabaseInfo(sts)
			if err != nil {
				if !kerr.IsNotFound(err) {
					klog.Warningf("failed to extract database info from StatefulSet: %s/%s. Reason: %v", sts.Namespace, sts.Name, err)
				}
				return
			}
			err = c.ensureReadyReplicasCond(dbInfo)
			if err != nil {
				klog.Warningf("failed to update ReadyReplicas condition. Reason: %v", err)
				return
			}
		},
	})
}

func (c *Controller) enqueueOnlyKubeDBSts(sts *apps.StatefulSet) {
	// only enqueue if the controlling owner is a KubeDB resource
	ok, _, err := core_util.IsOwnerOfGroup(metav1.GetControllerOf(sts), kubedb.GroupName)
	if err != nil {
		klog.Warningf("failed to enqueue StatefulSet: %s/%s. Reason: %v", sts.Namespace, sts.Name, err)
		return
	}
	if ok {
		queue.Enqueue(c.StsQueue.GetQueue(), cache.ExplicitKey(sts.Namespace+"/"+sts.Name))
	}
}

func (c *Controller) processStatefulSet(key string) error {
	klog.Infof("Started processing, key: %v", key)
	obj, exists, err := c.StsInformer.GetIndexer().GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		klog.V(5).Infof("StatefulSet %s does not exist anymore", key)
	} else {
		sts := obj.(*apps.StatefulSet).DeepCopy()
		dbInfo, err := c.extractDatabaseInfo(sts)
		if err != nil {
			return fmt.Errorf("failed to extract database info from StatefulSet: %s/%s. Reason: %v", sts.Namespace, sts.Name, err)
		}
		return c.ensureReadyReplicasCond(dbInfo)
	}
	return nil
}
