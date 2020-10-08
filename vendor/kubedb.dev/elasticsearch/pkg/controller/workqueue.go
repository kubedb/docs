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
	"fmt"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

func (c *Controller) initWatcher() {
	c.esInformer = c.KubedbInformerFactory.Kubedb().V1alpha2().Elasticsearches().Informer()
	c.esQueue = queue.New("Elasticsearch", c.MaxNumRequeues, c.NumThreads, c.runElasticsearch)
	c.esLister = c.KubedbInformerFactory.Kubedb().V1alpha2().Elasticsearches().Lister()
	c.esVersionLister = c.KubedbInformerFactory.Catalog().V1alpha1().ElasticsearchVersions().Lister()
	c.esInformer.AddEventHandler(queue.NewChangeHandler(c.esQueue.GetQueue()))
}

func (c *Controller) runElasticsearch(key string) error {
	log.Debugf("Processing, key: %v", key)
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
			if core_util.HasFinalizer(elasticsearch.ObjectMeta, kubedb.GroupName) {
				if err := c.terminate(elasticsearch); err != nil {
					log.Errorln(err)
					return err
				}
				_, _, err = util.PatchElasticsearch(context.TODO(), c.DBClient.KubedbV1alpha2(), elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, kubedb.GroupName)
					return in
				}, metav1.PatchOptions{})
				return err
			}
		} else {
			elasticsearch, _, err = util.PatchElasticsearch(context.TODO(), c.DBClient.KubedbV1alpha2(), elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, kubedb.GroupName)
				return in
			}, metav1.PatchOptions{})
			if err != nil {
				return err
			}

			// if conditions are empty, set initial condition "ProvisioningStarted" to "true"
			if elasticsearch.Status.Conditions == nil {
				_, err := util.UpdateElasticsearchStatus(
					context.TODO(),
					c.DBClient.KubedbV1alpha2(),
					elasticsearch.ObjectMeta,
					func(in *api.ElasticsearchStatus) *api.ElasticsearchStatus {
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:               api.DatabaseProvisioningStarted,
								Status:             kmapi.ConditionTrue,
								ObservedGeneration: elasticsearch.ObjectMeta.Generation,
								LastTransitionTime: metav1.Now(),
								Reason:             "",
								Message:            fmt.Sprintf("The KubeDB operator has started the provisioning of Elasticsearch: %s/%s", elasticsearch.Namespace, elasticsearch.Name),
							})
						return in
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					return err
				}
				// drop the object from queue,
				// the object will be enqueued again from this update event.
				return nil
			}

			// If the DB object is Paused, the operator will ignore the change events from
			// the DB object.
			if kmapi.IsConditionTrue(elasticsearch.Status.Conditions, api.DatabasePaused) {
				return nil
			}

			if elasticsearch.Spec.Halted {
				// If the DB object is Halted, but the status.phase is yet to be changed.
				if elasticsearch.Status.Phase != api.DatabasePhaseHalted {
					_, err := util.UpdateElasticsearchStatus(
						context.TODO(),
						c.DBClient.KubedbV1alpha2(),
						elasticsearch.ObjectMeta,
						func(in *api.ElasticsearchStatus) *api.ElasticsearchStatus {
							in.Phase = api.DatabasePhaseHalted
							in.ObservedGeneration = elasticsearch.Generation
							return in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						return err
					}
					// drop the object from queue,
					// the object will be enqueued again from this update event.
					return nil
				}

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

func (c *Controller) initSecretWatcher() {
	c.SecretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if secret, ok := obj.(*core.Secret); ok {
				if key := c.elasticsearchForSecret(secret); key != "" {
					queue.Enqueue(c.esQueue.GetQueue(), key)
				}
			}
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			if secret, ok := newObj.(*core.Secret); ok {
				if key := c.elasticsearchForSecret(secret); key != "" {
					queue.Enqueue(c.esQueue.GetQueue(), key)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
		},
	})
}
