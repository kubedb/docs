/*
Copyright The KubeDB Authors.

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

package snapshot

import (
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"

	"github.com/appscode/go/log"
	"k8s.io/apimachinery/pkg/labels"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

func (c *Controller) addEventHandler(selector labels.Selector) {
	c.SnapQueue = queue.New("Snapshot", c.MaxNumRequeues, c.NumThreads, c.runSnapshot)
	c.snLister = c.KubedbInformerFactory.Kubedb().V1alpha1().Snapshots().Lister()
	c.SnapInformer.AddEventHandler(queue.NewFilteredHandler(queue.NewEventHandler(c.SnapQueue.GetQueue(), func(old interface{}, new interface{}) bool {
		snapshot := new.(*api.Snapshot)
		return snapshot.DeletionTimestamp != nil
	}), selector))
}

func (c *Controller) runSnapshot(key string) error {
	log.Debugf("started processing, key: %v", key)
	obj, exists, err := c.SnapInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("Snapshot %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Snapshot was recreated with the same name
		snapshot := obj.(*api.Snapshot).DeepCopy()
		if snapshot.DeletionTimestamp != nil {
			if core_util.HasFinalizer(snapshot.ObjectMeta, api.GenericKey) {
				if err := c.delete(snapshot); kutil.IsRequestRetryable(err) {
					log.Errorln(err)
					return err
				}
				_, _, err = util.PatchSnapshot(c.ExtClient.KubedbV1alpha1(), snapshot, func(in *api.Snapshot) *api.Snapshot {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, api.GenericKey)
					return in
				})
				return err
			}
		} else {
			snapshot, _, err = util.PatchSnapshot(c.ExtClient.KubedbV1alpha1(), snapshot, func(in *api.Snapshot) *api.Snapshot {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, api.GenericKey)
				return in
			})
			if err != nil {
				log.Errorln(err)
				return err
			}
			if err := c.create(snapshot); kutil.IsRequestRetryable(err) {
				log.Errorln(err)
				return err
			}
		}
	}
	return nil
}
