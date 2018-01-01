package dormant_database

import (
	"fmt"
	"time"

	"github.com/appscode/go/log"
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

func (c *Controller) initWatcher() {

	// create the workqueue
	c.queue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "dormant_database")

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the DormantDatabase key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the DormantDatabase than the version which was responsible for triggering the update.
	c.indexer, c.informer = cache.NewIndexerInformer(c.lw, &api.DormantDatabase{}, c.syncPeriod, cache.ResourceEventHandlerFuncs{
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
			oldObj, ok := old.(*api.DormantDatabase)
			if !ok {
				log.Errorln("Invalid DormantDatabase object")
				return
			}
			newObj, ok := new.(*api.DormantDatabase)
			if !ok {
				log.Errorln("Invalid DormantDatabase object")
				return
			}

			if newObj.DeletionTimestamp != nil || !dormantDatabaseEqual(oldObj, newObj) {
				key, err := cache.MetaNamespaceKeyFunc(new)
				if err == nil {
					c.queue.Add(key)
				}
			}
		},
	}, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
}

func dormantDatabaseEqual(old, new *api.DormantDatabase) bool {
	if !meta_util.Equal(old.Spec, new.Spec) {
		diff := meta_util.Diff(old.Spec, new.Spec)
		log.Debugf("DormantDatabase %s/%s has changed. Diff: %s\n", new.Namespace, new.Name, diff)
		return false
	}
	return true
}

func (c *Controller) runWatcher(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	log.Infoln("Starting DormantDatabase Controller")

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	log.Infoln("Stopping DormantDatabase Controller")
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
	// This allows safe parallel processing because two DormantDatabases with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.runDormantDatabase(key.(string))
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		log.Debugf("Finished Processing key: %v\n", key)
		return true
	}
	log.Errorf("Failed to process DormantDatabase %v. Reason: %s\n", key, err)

	// This Controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < c.maxNumRequests {
		log.Infof("Error syncing crd %v: %v\n", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return true
	}

	c.queue.Forget(key)
	log.Debugf("Finished Processing key: %v\n", key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	log.Infof("Dropping DormantDatabase %q out of the queue: %v\n", key, err)
	return true
}

func (c *Controller) runDormantDatabase(key string) error {
	log.Debugf("started processing, key: %v\n", key)
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v\n", key, err)
		return err
	}

	if !exists {
		log.Debugf("DormantDatabase %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a DormantDatabase was recreated with the same name
		dormantDatabase := obj.(*api.DormantDatabase).DeepCopy()
		if dormantDatabase.DeletionTimestamp != nil {
			if core_util.HasFinalizer(dormantDatabase.ObjectMeta, "kubedb.com") {
				util.AssignTypeKind(dormantDatabase)
				if err := c.delete(dormantDatabase); err != nil {
					log.Errorln(err)
					return err
				}
				dormantDatabase, _, err = util.PatchDormantDatabase(c.ExtClient, dormantDatabase, func(in *api.DormantDatabase) *api.DormantDatabase {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, "kubedb.com")
					return in
				})
				return err
			}
		} else {
			dormantDatabase, _, err = util.PatchDormantDatabase(c.ExtClient, dormantDatabase, func(in *api.DormantDatabase) *api.DormantDatabase {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, "kubedb.com")
				return in
			})
			util.AssignTypeKind(dormantDatabase)
			if err := c.create(dormantDatabase); err != nil {
				log.Errorln(err)
				return err
			}
		}
	}
	return nil
}
