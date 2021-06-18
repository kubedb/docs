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
	c.pxInformer = c.KubedbInformerFactory.Kubedb().V1alpha2().PerconaXtraDBs().Informer()
	c.pxQueue = queue.New(api.ResourceKindPerconaXtraDB, c.MaxNumRequeues, c.NumThreads, c.runPerconaXtraDB)
	c.pxLister = c.KubedbInformerFactory.Kubedb().V1alpha2().PerconaXtraDBs().Lister()
	c.pxInformer.AddEventHandler(queue.NewChangeHandler(c.pxQueue.GetQueue()))
	if c.Auditor != nil {
		c.pxInformer.AddEventHandler(c.Auditor.ForGVK(api.SchemeGroupVersion.WithKind(api.ResourceKindPerconaXtraDB)))
	}
}

func (c *Controller) runPerconaXtraDB(key string) error {
	klog.V(5).Infoln("started processing, key:", key)
	obj, exists, err := c.pxInformer.GetIndexer().GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		klog.V(5).Infof("PerconaXtraDB %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a PerconaXtraDB was recreated with the same name
		px := obj.(*api.PerconaXtraDB).DeepCopy()
		if px.DeletionTimestamp != nil {
			if core_util.HasFinalizer(px.ObjectMeta, kubedb.GroupName) {
				if err := c.terminate(px); err != nil {
					klog.Errorln(err)
					return err
				}
				_, _, err = util.PatchPerconaXtraDB(context.TODO(), c.DBClient.KubedbV1alpha2(), px, func(in *api.PerconaXtraDB) *api.PerconaXtraDB {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, kubedb.GroupName)
					return in
				}, metav1.PatchOptions{})
				return err
			}
		} else {
			px, _, err = util.PatchPerconaXtraDB(context.TODO(), c.DBClient.KubedbV1alpha2(), px, func(in *api.PerconaXtraDB) *api.PerconaXtraDB {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, kubedb.GroupName)
				return in
			}, metav1.PatchOptions{})
			if err != nil {
				return err
			}

			if kmapi.IsConditionTrue(px.Status.Conditions, api.DatabasePaused) {
				return nil
			}

			if px.Spec.Halted {
				if err := c.halt(px); err != nil {
					klog.Errorln(err)
					c.pushFailureEvent(px, err.Error())
					return err
				}
			} else {
				if err := c.create(px); err != nil {
					klog.Errorln(err)
					c.pushFailureEvent(px, err.Error())
					return err
				}
			}
		}
	}
	return nil
}
