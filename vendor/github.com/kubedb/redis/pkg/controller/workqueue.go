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
	c.rdInformer = c.KubedbInformerFactory.Kubedb().V1alpha1().Redises().Informer()
	c.rdQueue = queue.New("Redis", c.MaxNumRequeues, c.NumThreads, c.runRedis)
	c.rdLister = c.KubedbInformerFactory.Kubedb().V1alpha1().Redises().Lister()
	c.rdInformer.AddEventHandler(queue.NewEventHandler(c.rdQueue.GetQueue(), func(old interface{}, new interface{}) bool {
		oldObj := old.(*api.Redis)
		newObj := new.(*api.Redis)
		return newObj.DeletionTimestamp != nil || !redisEqual(oldObj, newObj)
	}))
}

func redisEqual(old, new *api.Redis) bool {
	if !meta_util.Equal(old.Spec, new.Spec) {
		diff := meta_util.Diff(old.Spec, new.Spec)
		log.Infof("Redis %s/%s has changed. Diff: %s", new.Namespace, new.Name, diff)
		return false
	}
	if !meta_util.Equal(old.Annotations, new.Annotations) {
		diff := meta_util.Diff(old.Annotations, new.Annotations)
		log.Infof("Annotations in Redis %s/%s has changed. Diff: %s\n", new.Namespace, new.Name, diff)
		return false
	}
	return true
}

func (c *Controller) runRedis(key string) error {
	log.Debugln("started processing, key:", key)
	obj, exists, err := c.rdInformer.GetIndexer().GetByKey(key)
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
