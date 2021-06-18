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
	"context"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

func (c *Controller) initWatcher() {
	c.mcInformer = c.KubedbInformerFactory.Kubedb().V1alpha2().Memcacheds().Informer()
	c.mcQueue = queue.New(api.ResourceKindMemcached, c.MaxNumRequeues, c.NumThreads, c.runMemcached)
	c.mcLister = c.KubedbInformerFactory.Kubedb().V1alpha2().Memcacheds().Lister()
	c.mcInformer.AddEventHandler(queue.NewChangeHandler(c.mcQueue.GetQueue()))
	if c.Auditor != nil {
		c.mcInformer.AddEventHandler(c.Auditor.ForGVK(api.SchemeGroupVersion.WithKind(api.ResourceKindMemcached)))
	}
}

func (c *Controller) runMemcached(key string) error {
	klog.V(5).Infoln("started processing, key:", key)
	obj, exists, err := c.mcInformer.GetIndexer().GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		klog.V(5).Infof("Memcached %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Memcached was recreated with the same name
		memcached := obj.(*api.Memcached).DeepCopy()
		if memcached.DeletionTimestamp != nil {
			if core_util.HasFinalizer(memcached.ObjectMeta, kubedb.GroupName) {
				if err := c.terminate(memcached); err != nil {
					klog.Errorln(err)
					return err
				}
				_, _, err = util.PatchMemcached(context.TODO(), c.DBClient.KubedbV1alpha2(), memcached, func(in *api.Memcached) *api.Memcached {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, kubedb.GroupName)
					return in
				}, metav1.PatchOptions{})
				return err
			}
		} else {
			memcached, _, err = util.PatchMemcached(context.TODO(), c.DBClient.KubedbV1alpha2(), memcached, func(in *api.Memcached) *api.Memcached {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, kubedb.GroupName)
				return in
			}, metav1.PatchOptions{})
			if err != nil {
				return err
			}

			if kmapi.IsConditionTrue(memcached.Status.Conditions, api.DatabasePaused) {
				return nil
			}

			if memcached.Spec.Halted {
				if err := c.halt(memcached); err != nil {
					klog.Errorln(err)
					c.pushFailureEvent(memcached, err.Error())
					return err
				}
			} else {
				if err := c.create(memcached); err != nil {
					klog.Errorln(err)
					c.pushFailureEvent(memcached, err.Error())
					return err
				}
			}
		}
	}
	return nil
}
