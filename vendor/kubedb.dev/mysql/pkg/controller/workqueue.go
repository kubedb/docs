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
	"kubedb.dev/apimachinery/pkg/phase"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

func (c *Controller) initWatcher() {
	c.myInformer = c.KubedbInformerFactory.Kubedb().V1alpha2().MySQLs().Informer()
	c.myQueue = queue.New("MySQL", c.MaxNumRequeues, c.NumThreads, c.runMySQL)
	c.myLister = c.KubedbInformerFactory.Kubedb().V1alpha2().MySQLs().Lister()
	c.myInformer.AddEventHandler(queue.NewChangeHandler(c.myQueue.GetQueue()))
}

func (c *Controller) runMySQL(key string) error {
	log.Debugln("started processing, key:", key)
	obj, exists, err := c.myInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("MySQL %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a MySQL was recreated with the same name
		db := obj.(*api.MySQL).DeepCopy()
		if db.DeletionTimestamp != nil {
			if core_util.HasFinalizer(db.ObjectMeta, kubedb.GroupName) {
				if err := c.terminate(db); err != nil {
					log.Errorln(err)
					return err
				}
				_, _, err = util.PatchMySQL(context.TODO(), c.DBClient.KubedbV1alpha2(), db, func(in *api.MySQL) *api.MySQL {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, kubedb.GroupName)
					return in
				}, metav1.PatchOptions{})
				return err
			}
		} else {
			db, _, err = util.PatchMySQL(context.TODO(), c.DBClient.KubedbV1alpha2(), db, func(in *api.MySQL) *api.MySQL {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, kubedb.GroupName)
				return in
			}, metav1.PatchOptions{})
			if err != nil {
				return err
			}

			// Get MySQL phase from condition
			// If new phase is not equal to old phase,
			// update MySQL phase.
			phase := phase.PhaseFromCondition(db.Status.Conditions)
			if db.Status.Phase != phase {
				_, err := util.UpdateMySQLStatus(
					context.TODO(),
					c.DBClient.KubedbV1alpha2(),
					db.ObjectMeta,
					func(in *api.MySQLStatus) *api.MySQLStatus {
						in.Phase = phase
						in.ObservedGeneration = db.Generation
						return in
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					c.pushFailureEvent(db, err.Error())
					return err
				}
				// drop the object from queue,
				// the object will be enqueued again from this update event.
				return nil
			}

			// if conditions are empty, set initial condition "ProvisioningStarted" to "true"
			if !kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseProvisioningStarted) {
				_, err := util.UpdateMySQLStatus(
					context.TODO(),
					c.DBClient.KubedbV1alpha2(),
					db.ObjectMeta,
					func(in *api.MySQLStatus) *api.MySQLStatus {
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:    api.DatabaseProvisioningStarted,
								Status:  core.ConditionTrue,
								Reason:  api.DatabaseProvisioningStartedSuccessfully,
								Message: fmt.Sprintf("The KubeDB operator has started the provisioning of MySQL: %s/%s", db.Namespace, db.Name),
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
			if kmapi.IsConditionTrue(db.Status.Conditions, api.DatabasePaused) {
				return nil
			}

			if db.Spec.Halted {
				if err := c.halt(db); err != nil {
					log.Errorln(err)
					c.pushFailureEvent(db, err.Error())
					return err
				}
			} else {
				// Here, spec.halted=false, remove the halted condition if exists.
				if kmapi.HasCondition(db.Status.Conditions, api.DatabaseHalted) {
					if _, err := util.UpdateMySQLStatus(
						context.TODO(),
						c.DBClient.KubedbV1alpha2(),
						db.ObjectMeta,
						func(in *api.MySQLStatus) *api.MySQLStatus {
							in.Conditions = kmapi.RemoveCondition(in.Conditions, api.DatabaseHalted)
							return in
						},
						metav1.UpdateOptions{},
					); err != nil {
						return err
					}
					// return from here, will be enqueued again from the event.
					return nil
				}
				// process db object
				if err := c.create(db); err != nil {
					log.Errorln(err)
					c.pushFailureEvent(db, err.Error())
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
				if key := c.mysqlForSecret(secret); key != "" {
					queue.Enqueue(c.myQueue.GetQueue(), key)
				}
			}
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			if secret, ok := newObj.(*core.Secret); ok {
				if key := c.mysqlForSecret(secret); key != "" {
					queue.Enqueue(c.myQueue.GetQueue(), key)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
		},
	})
}
