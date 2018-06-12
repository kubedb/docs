package controller

import (
	"github.com/appscode/go/log"
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	"github.com/appscode/kutil/tools/queue"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
)

func (c *Controller) initWatcher() {
	c.mcInformer = c.KubedbInformerFactory.Kubedb().V1alpha1().Memcacheds().Informer()
	c.mcQueue = queue.New("Memcached", c.MaxNumRequeues, c.NumThreads, c.runMemcached)
	c.mcLister = c.KubedbInformerFactory.Kubedb().V1alpha1().Memcacheds().Lister()
	c.mcInformer.AddEventHandler(queue.NewEventHandler(c.mcQueue.GetQueue(), func(old interface{}, new interface{}) bool {
		oldObj := old.(*api.Memcached)
		newObj := new.(*api.Memcached)
		return newObj.DeletionTimestamp != nil || !memcachedEqual(oldObj, newObj)
	}))
}

func memcachedEqual(old, new *api.Memcached) bool {
	if !meta_util.Equal(old.Spec, new.Spec) {
		diff := meta_util.Diff(old.Spec, new.Spec)
		log.Infoln("Memcached %s/%s has changed. Diff: %s", new.Namespace, new.Name, diff)
		return false
	}
	return true
}

func (c *Controller) runMemcached(key string) error {
	log.Debugln("started processing, key:", key)
	obj, exists, err := c.mcInformer.GetIndexer().GetByKey(key)
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
			if core_util.HasFinalizer(memcached.ObjectMeta, api.GenericKey) {
				util.AssignTypeKind(memcached)
				if err := c.pause(memcached); err != nil {
					log.Errorln(err)
					return err
				}
				memcached, _, err = util.PatchMemcached(c.ExtClient, memcached, func(in *api.Memcached) *api.Memcached {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, api.GenericKey)
					return in
				})
				return err
			}
		} else {
			memcached, _, err = util.PatchMemcached(c.ExtClient, memcached, func(in *api.Memcached) *api.Memcached {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, api.GenericKey)
				return in
			})
			if err != nil {
				return err
			}
			util.AssignTypeKind(memcached)
			if err := c.create(memcached); err != nil {
				log.Errorln(err)
				c.pushFailureEvent(memcached, err.Error())
				return err
			}
		}
	}
	return nil
}
