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
	"strings"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"
	"kubedb.dev/apimachinery/pkg/phase"

	"gomodules.xyz/pointer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	v1 "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

func (c *Controller) initWatcher() {
	c.rdInformer = c.KubedbInformerFactory.Kubedb().V1alpha2().Redises().Informer()
	c.rdQueue = queue.New(api.ResourceKindRedis, c.MaxNumRequeues, c.NumThreads, c.runRedis)
	c.rdLister = c.KubedbInformerFactory.Kubedb().V1alpha2().Redises().Lister()
	c.rdInformer.AddEventHandler(queue.NewChangeHandler(c.rdQueue.GetQueue(), c.RestrictToNamespace))
	if c.Auditor != nil {
		c.rdInformer.AddEventHandler(c.Auditor.ForGVK(api.SchemeGroupVersion.WithKind(api.ResourceKindRedis)))
	}
	c.rdStsInformer = c.KubeInformerFactory.Apps().V1().StatefulSets().Informer()
}

func (c *Controller) initSentinelWatcher() {
	c.rsInformer = c.KubedbInformerFactory.Kubedb().V1alpha2().RedisSentinels().Informer()
	c.rsQueue = queue.New(api.ResourceKindRedisSentinel, c.MaxNumRequeues, c.NumThreads, c.runRedisSentinel)
	c.rsLister = c.KubedbInformerFactory.Kubedb().V1alpha2().RedisSentinels().Lister()
	c.rsInformer.AddEventHandler(queue.NewChangeHandler(c.rsQueue.GetQueue(), c.RestrictToNamespace))
	if c.Auditor != nil {
		c.rsInformer.AddEventHandler(c.Auditor.ForGVK(api.SchemeGroupVersion.WithKind(api.ResourceKindRedisSentinel)))
	}
}

func (c *Controller) runRedis(key string) error {
	klog.V(5).Infoln("started processing, key:", key)
	obj, exists, err := c.rdInformer.GetIndexer().GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		klog.V(5).Infof("Redis %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Redis was recreated with the same name
		db := obj.(*api.Redis).DeepCopy()
		if db.DeletionTimestamp != nil {
			if db.Spec.Mode == api.RedisModeSentinel {
				sentinel, err := c.DBClient.KubedbV1alpha2().RedisSentinels(db.Spec.SentinelRef.Namespace).Get(context.TODO(), db.Spec.SentinelRef.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}
				for i := 0; i < int(pointer.Int32(sentinel.Spec.Replicas)); i++ {
					dnsName := fmt.Sprintf("%s-%v.%s.%s.svc", sentinel.Name, i, sentinel.GoverningServiceName(), sentinel.Namespace)
					rdClient, err := c.getRedisSentinelClient(sentinel, dnsName, 26379)
					if err != nil {
						return err
					}
					rdClient.Master(GetRdClusterRegisteredNameInSentinel(db))
					output := rdClient.Remove(GetRdClusterRegisteredNameInSentinel(db))
					err = rdClient.Close()
					if err != nil {
						return err
					}
					if !strings.Contains(output.String(), "OK") && !strings.Contains(output.String(), "No such master") {
						//we need to make sure that the redis sentinel cluster has been remove the sentinel successfully
						//TODO: need to make sure that in every version have the same string type as output
						return fmt.Errorf("failed to remove from sentinel")
					}
				}
			}
			if core_util.HasFinalizer(db.ObjectMeta, kubedb.GroupName) {
				if err := c.terminate(db); err != nil {
					klog.Errorln(err)
					return err
				}
				_, _, err = util.PatchRedis(context.TODO(), c.DBClient.KubedbV1alpha2(), db, func(in *api.Redis) *api.Redis {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, kubedb.GroupName)
					return in
				}, metav1.PatchOptions{})
				return err
			}
		} else {
			db, _, err = util.PatchRedis(context.TODO(), c.DBClient.KubedbV1alpha2(), db, func(in *api.Redis) *api.Redis {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, kubedb.GroupName)
				return in
			}, metav1.PatchOptions{})
			if err != nil {
				return err
			}

			// Get redis phase from condition
			// If new phase is not equal to old phase,
			// update redis phase.
			phase := phase.PhaseFromCondition(db.Status.Conditions)
			if db.Status.Phase != phase {
				_, err := util.UpdateRedisStatus(
					context.TODO(),
					c.DBClient.KubedbV1alpha2(),
					db.ObjectMeta,
					func(in *api.RedisStatus) (types.UID, *api.RedisStatus) {
						in.Phase = phase
						in.ObservedGeneration = db.Generation
						return db.UID, in
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
			if db.Status.Conditions == nil {
				_, err := util.UpdateRedisStatus(
					context.TODO(),
					c.DBClient.KubedbV1alpha2(),
					db.ObjectMeta,
					func(in *api.RedisStatus) (types.UID, *api.RedisStatus) {
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:    api.DatabaseProvisioningStarted,
								Status:  core.ConditionTrue,
								Reason:  api.DatabaseProvisioningStartedSuccessfully,
								Message: fmt.Sprintf("The KubeDB operator has started the provisioning of Redis: %s/%s", db.Namespace, db.Name),
							})
						return db.UID, in
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
					klog.Errorln(err)
					c.pushFailureEvent(db, err.Error())
					return err
				}
			} else {
				// Here, spec.halted=false, remove the halted condition if exists.
				if kmapi.HasCondition(db.Status.Conditions, api.DatabaseHalted) {
					if _, err := util.UpdateRedisStatus(
						context.TODO(),
						c.DBClient.KubedbV1alpha2(),
						db.ObjectMeta,
						func(in *api.RedisStatus) (types.UID, *api.RedisStatus) {
							in.Conditions = kmapi.RemoveCondition(in.Conditions, api.DatabaseHalted)
							return db.UID, in
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
					klog.Errorln(err)
					c.pushFailureEvent(db, err.Error())
					return err
				}
			}
		}
	}
	return nil
}

func (c *Controller) runRedisSentinel(key string) error {
	klog.V(5).Infoln("started processing, key:", key)
	obj, exists, err := c.rsInformer.GetIndexer().GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		klog.V(5).Infof("Redis Sentinel %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Redis was recreated with the same name
		db := obj.(*api.RedisSentinel).DeepCopy()
		if db.DeletionTimestamp != nil {
			if core_util.HasFinalizer(db.ObjectMeta, kubedb.GroupName) {
				if err := c.terminateSentinel(db); err != nil {
					klog.Errorln(err)
					return err
				}
				_, _, err = util.PatchRedisSentinel(context.TODO(), c.DBClient.KubedbV1alpha2(), db, func(in *api.RedisSentinel) *api.RedisSentinel {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, kubedb.GroupName)
					return in
				}, metav1.PatchOptions{})
				return err
			}
		} else {
			db, _, err = util.PatchRedisSentinel(context.TODO(), c.DBClient.KubedbV1alpha2(), db, func(in *api.RedisSentinel) *api.RedisSentinel {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, kubedb.GroupName)
				return in
			}, metav1.PatchOptions{})
			if err != nil {
				return err
			}

			// Get redis phase from condition
			// If new phase is not equal to old phase,
			// update redis phase.
			phase := phase.PhaseFromCondition(db.Status.Conditions)
			if db.Status.Phase != phase {
				_, err := util.UpdateRedisSentinelStatus(
					context.TODO(),
					c.DBClient.KubedbV1alpha2(),
					db.ObjectMeta,
					func(in *api.RedisSentinelStatus) (types.UID, *api.RedisSentinelStatus) {
						in.Phase = phase
						in.ObservedGeneration = db.Generation
						return db.UID, in
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					c.pushSentinelFailureEvent(db, err.Error())
					return err
				}
				// drop the object from queue,
				// the object will be enqueued again from this update event.
				return nil
			}

			// if conditions are empty, set initial condition "ProvisioningStarted" to "true"
			if db.Status.Conditions == nil {
				_, err := util.UpdateRedisSentinelStatus(
					context.TODO(),
					c.DBClient.KubedbV1alpha2(),
					db.ObjectMeta,
					func(in *api.RedisSentinelStatus) (types.UID, *api.RedisSentinelStatus) {
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:    api.DatabaseProvisioningStarted,
								Status:  core.ConditionTrue,
								Reason:  api.DatabaseProvisioningStartedSuccessfully,
								Message: fmt.Sprintf("The KubeDB operator has started the provisioning of Redis: %s/%s", db.Namespace, db.Name),
							})
						return db.UID, in
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
				if err := c.haltSentinel(db); err != nil {
					klog.Errorln(err)
					c.pushSentinelFailureEvent(db, err.Error())
					return err
				}
			} else {
				// Here, spec.halted=false, remove the halted condition if exists.
				if kmapi.HasCondition(db.Status.Conditions, api.DatabaseHalted) {
					if _, err := util.UpdateRedisSentinelStatus(
						context.TODO(),
						c.DBClient.KubedbV1alpha2(),
						db.ObjectMeta,
						func(in *api.RedisSentinelStatus) (types.UID, *api.RedisSentinelStatus) {
							in.Conditions = kmapi.RemoveCondition(in.Conditions, api.DatabaseHalted)
							return db.UID, in
						},
						metav1.UpdateOptions{},
					); err != nil {
						return err
					}
					// return from here, will be enqueued again from the event.
					return nil
				}

				// process db object
				if err := c.createSentinel(db); err != nil {
					klog.Errorln(err)
					c.pushSentinelFailureEvent(db, err.Error())
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
				if key := c.RedisForSecret(secret); key != "" {
					queue.Enqueue(c.rdQueue.GetQueue(), key)
				} else if key := c.RedisSentinelForSecret(secret); key != "" {
					queue.Enqueue(c.rsQueue.GetQueue(), key)
				}
			}
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			if secret, ok := newObj.(*core.Secret); ok {
				if key := c.RedisForSecret(secret); key != "" {
					queue.Enqueue(c.rdQueue.GetQueue(), key)
				} else if key := c.RedisSentinelForSecret(secret); key != "" {
					queue.Enqueue(c.rsQueue.GetQueue(), key)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
		},
	})
}

func (c *Controller) stsWatcher() {
	c.rdStsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if sts, ok := obj.(*apps.StatefulSet); ok {
				owner := metav1.GetControllerOf(sts)
				ok, kind, err := core_util.IsOwnerOfGroup(owner, kubedb.GroupName)
				if err != nil {
					klog.Warningf("failed to enqueue StatefulSet: %s/%s. Reason: %v", sts.Namespace, sts.Name, err)
					return
				}
				if !ok && kind != api.ResourceKindRedis && kind != api.ResourceKindRedisSentinel {
					return
				}

				if v1.IsStatefulSetReady(sts) {
					if kind == api.ResourceKindRedis {
						queue.Enqueue(c.rdQueue.GetQueue(), cache.ExplicitKey(sts.Namespace+"/"+owner.Name))
					} else {
						queue.Enqueue(c.rsQueue.GetQueue(), cache.ExplicitKey(sts.Namespace+"/"+owner.Name))
					}
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if sts, ok := newObj.(*apps.StatefulSet); ok {
				owner := metav1.GetControllerOf(sts)
				ok, kind, err := core_util.IsOwnerOfGroup(owner, kubedb.GroupName)
				if err != nil {
					klog.Warningf("failed to enqueue StatefulSet: %s/%s. Reason: %v", sts.Namespace, sts.Name, err)
					return
				}
				if !ok && kind != api.ResourceKindRedis && kind != api.ResourceKindRedisSentinel {
					return
				}

				if v1.IsStatefulSetReady(sts) && kind == api.ResourceKindRedis {
					if kind == api.ResourceKindRedis {
						queue.Enqueue(c.rdQueue.GetQueue(), cache.ExplicitKey(sts.Namespace+"/"+owner.Name))
					} else {
						queue.Enqueue(c.rsQueue.GetQueue(), cache.ExplicitKey(sts.Namespace+"/"+owner.Name))
					}
				}
			}
		},
		DeleteFunc: nil,
	})
}
