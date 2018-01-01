package controller

import (
	"fmt"
	"time"

	"github.com/appscode/go/log"
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1/util"
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
			return c.ExtClient.Memcacheds(metav1.NamespaceAll).List(metav1.ListOptions{})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.ExtClient.Memcacheds(metav1.NamespaceAll).Watch(metav1.ListOptions{})
		},
	}

	// create the workqueue
	c.queue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "memcached")

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the Memcached key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Memcached than the version which was responsible for triggering the update.
	c.indexer, c.informer = cache.NewIndexerInformer(lw, &api.Memcached{}, c.syncPeriod, cache.ResourceEventHandlerFuncs{
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
			oldObj, ok := old.(*api.Memcached)
			if !ok {
				log.Errorln("Invalid Memcached object")
				return
			}
			newObj, ok := new.(*api.Memcached)
			if !ok {
				log.Errorln("Invalid Memcached object")
				return
			}
			if newObj.DeletionTimestamp != nil || !memcachedEqual(oldObj, newObj) {
				key, err := cache.MetaNamespaceKeyFunc(new)
				if err == nil {
					c.queue.Add(key)
				}
			}
		},
	}, cache.Indexers{})
}

func memcachedEqual(old, new *api.Memcached) bool {
	if !meta_util.Equal(old.Spec, new.Spec) {
		diff := meta_util.Diff(old.Spec, new.Spec)
		log.Infoln("Memcached %s/%s has changed. Diff: %s", new.Namespace, new.Name, diff)
		return false
	}
	return true
}

func (c *Controller) runWatcher(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	log.Infof("Starting Memcached controller")

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
	log.Infof("Stopping Memcached controller")

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
	// This allows safe parallel processing because two Memcacheds with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.runMemcached(key.(string))
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		log.Debugln("Finished Processing key: %v", key)
		return true
	}
	log.Errorf("Failed to process Memcached %v. Reason: %s", key, err)

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < c.opt.MaxNumRequeues {
		log.Infof("Error syncing crd %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return true
	}

	c.queue.Forget(key)
	log.Debugln("Finished Processing key: %v", key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	log.Infof("Dropping deployment %q out of the queue: %v", key, err)
	return true
}

func (c *Controller) runMemcached(key string) error {
	log.Debugln("started processing, key: %v", key)
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("Memcached %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Memcached was recreated with the same name
		memcached := obj.(*api.Memcached).DeepCopy()
		if memcached.DeletionTimestamp != nil {
			if core_util.HasFinalizer(memcached.ObjectMeta, "kubedb.com") {
				if err := c.pause(memcached); err != nil {
					log.Errorln(err)
					return err
				}
				memcached, _, err = util.PatchMemcached(c.ExtClient, memcached, func(in *api.Memcached) *api.Memcached {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, "kubedb.com")
					return in
				})
				return err
			}
		} else {
			memcached, _, err = util.PatchMemcached(c.ExtClient, memcached, func(in *api.Memcached) *api.Memcached {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, "kubedb.com")
				return in
			})
			if err := c.create(memcached); err != nil {
				log.Errorln(err)
				c.pushFailureEvent(memcached, err.Error())
				return err
			}
		}
	}
	return nil
}
