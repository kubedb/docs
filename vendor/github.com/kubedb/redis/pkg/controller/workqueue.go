package controller

import (
	"fmt"
	"time"

	"github.com/appscode/go/log"
	core_util "github.com/appscode/kutil/core/v1"
	core_meta "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rt "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

func (c *Controller) initWatcher() {
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (rt.Object, error) {
			return c.ExtClient.Redises(metav1.NamespaceAll).List(metav1.ListOptions{})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.ExtClient.Redises(metav1.NamespaceAll).Watch(metav1.ListOptions{})
		},
	}

	// create the workqueue
	c.queue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "redis")

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the Redis key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Redis than the version which was responsible for triggering the update.
	c.indexer, c.informer = cache.NewIndexerInformer(lw, &api.Redis{}, c.syncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			oldObj, ok := old.(*api.Redis)
			if !ok {
				log.Errorln("Invalid Redis object")
				return
			}
			newObj, ok := new.(*api.Redis)
			if !ok {
				log.Errorln("Invalid Redis object")
				return
			}
			if newObj.DeletionTimestamp != nil || !redisEqual(oldObj, newObj) {
				key, err := cache.MetaNamespaceKeyFunc(new)
				if err == nil {
					c.queue.Add(key)
				}
			}
		},
	}, cache.Indexers{})
}

func redisEqual(old, new *api.Redis) bool {
	var oldSpec, newSpec *api.RedisSpec
	if old != nil {
		oldSpec = &old.Spec
	}
	if new != nil {
		newSpec = &new.Spec
	}
	if !core_meta.Equal(oldSpec, newSpec) {
		diff := core_meta.Diff(oldSpec, newSpec)
		log.Infoln("Redis %s/%s has changed. Diff: %s", new.Namespace, new.Name, diff)
		return false
	}
	return true
}

func (c *Controller) runWatcher(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	log.Infoln("Starting Redis controller")

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	log.Infoln("Stopping Redis controller")

}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two Redises with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.runRedis(key.(string))
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		log.Debugln("Finished Processing key:", key)
		return true
	}
	log.Errorf("Failed to process Redis %v. Reason: %s", key, err)

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < c.opt.MaxNumRequeues {
		log.Infof("Error syncing crd %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return true
	}

	c.queue.Forget(key)
	log.Debugln("Finished Processing key:", key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	log.Infof("Dropping deployment %q out of the queue: %v", key, err)
	return true
}

func (c *Controller) runRedis(key string) error {
	log.Debugln("started processing, key:", key)
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("Redis %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Redis was recreated with the same name
		redis := obj.(*api.Redis).DeepCopy()
		if redis.DeletionTimestamp != nil {
			if core_util.HasFinalizer(redis.ObjectMeta, api.GenericKey) {
				util.AssignTypeKind(redis)
				if err := c.pause(redis); err != nil {
					log.Errorln(err)
					return err
				}
				redis, _, err = util.PatchRedis(c.ExtClient, redis, func(in *api.Redis) *api.Redis {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, api.GenericKey)
					return in
				})
				return err
			}
		} else {
			redis, _, err = util.PatchRedis(c.ExtClient, redis, func(in *api.Redis) *api.Redis {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, api.GenericKey)
				return in
			})
			util.AssignTypeKind(redis)
			if err := c.create(redis); err != nil {
				log.Errorln(err)
				c.pushFailureEvent(redis, err.Error())
				return err
			}
		}
	}
	return nil
}
