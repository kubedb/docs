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

package restoresession

import (
	"github.com/appscode/go/log"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"kmodules.xyz/client-go/tools/queue"
	"stash.appscode.dev/apimachinery/apis/stash/v1beta1"
)

func (c *Controller) addEventHandler(selector labels.Selector) {
	c.RSQueue = queue.New("RestoreSession", c.MaxNumRequeues, c.NumThreads, c.runRestoreSession)
	c.rsLister = c.StashInformerFactory.Stash().V1beta1().RestoreSessions().Lister()
	c.RSInformer.AddEventHandler(queue.NewFilteredHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			rs := obj.(*v1beta1.RestoreSession)
			if rs.Status.Phase == v1beta1.RestoreSessionSucceeded || rs.Status.Phase == v1beta1.RestoreSessionFailed {
				queue.Enqueue(c.RSQueue.GetQueue(), obj)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			oldObj := old.(*v1beta1.RestoreSession)
			newObj := new.(*v1beta1.RestoreSession)
			if newObj.Status.Phase != oldObj.Status.Phase && (newObj.Status.Phase == v1beta1.RestoreSessionSucceeded || newObj.Status.Phase == v1beta1.RestoreSessionFailed) {
				queue.Enqueue(c.RSQueue.GetQueue(), newObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
		},
	}, selector))
}

func (c *Controller) runRestoreSession(key string) error {
	log.Debugf("started processing, key: %v", key)
	obj, exists, err := c.RSInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("RestoreSession %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Job was recreated with the same name
		rs := obj.(*v1beta1.RestoreSession).DeepCopy()
		if err := c.handleRestoreSession(rs); err != nil {
			log.Errorln(err)
			return err
		}
	}
	return nil
}
