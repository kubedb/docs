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
package controller

import (
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"

	"github.com/appscode/go/log"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

func (c *Controller) initWatcher() {
	c.esInformer = c.KubedbInformerFactory.Kubedb().V1alpha1().Elasticsearches().Informer()
	c.esQueue = queue.New("Elasticsearch", c.MaxNumRequeues, c.NumThreads, c.runElasticsearch)
	c.esLister = c.KubedbInformerFactory.Kubedb().V1alpha1().Elasticsearches().Lister()
	c.esVersionLister = c.KubedbInformerFactory.Catalog().V1alpha1().ElasticsearchVersions().Lister()
	c.esInformer.AddEventHandler(queue.NewReconcilableHandler(c.esQueue.GetQueue()))
}

func (c *Controller) runElasticsearch(key string) error {
	log.Debugf("started processing, key: %v", key)
	obj, exists, err := c.esInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("Elasticsearch %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Elasticsearch was recreated with the same name
		elasticsearch := obj.(*api.Elasticsearch).DeepCopy()
		if elasticsearch.DeletionTimestamp != nil {
			if core_util.HasFinalizer(elasticsearch.ObjectMeta, "kubedb.com") {
				if err := c.terminate(elasticsearch); err != nil {
					log.Errorln(err)
					return err
				}
				_, _, err = util.PatchElasticsearch(c.ExtClient.KubedbV1alpha1(), elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, "kubedb.com")
					return in
				})
				return err
			}
		} else {
			elasticsearch, _, err = util.PatchElasticsearch(c.ExtClient.KubedbV1alpha1(), elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, "kubedb.com")
				return in
			})
			if err != nil {
				return err
			}

			if elasticsearch.Spec.Paused {
				return nil
			}

			if elasticsearch.Spec.Halted {
				if err := c.halt(elasticsearch); err != nil {
					log.Errorln(err)
					c.pushFailureEvent(elasticsearch, err.Error())
					return err
				}
			} else {
				if err := c.create(elasticsearch); err != nil {
					log.Errorln(err)
					c.pushFailureEvent(elasticsearch, err.Error())
					return err
				}
			}
		}
	}
	return nil
}
