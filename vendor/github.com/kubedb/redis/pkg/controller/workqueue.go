package controller

import (
	"github.com/appscode/go/log"
	core_util "github.com/appscode/kutil/core/v1"
	"github.com/appscode/kutil/tools/queue"
	"github.com/kubedb/apimachinery/apis"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
)

func (c *Controller) initWatcher() {
	c.rdInformer = c.KubedbInformerFactory.Kubedb().V1alpha1().Redises().Informer()
	c.rdQueue = queue.New("Redis", c.MaxNumRequeues, c.NumThreads, c.runRedis)
	c.rdLister = c.KubedbInformerFactory.Kubedb().V1alpha1().Redises().Lister()
	c.rdInformer.AddEventHandler(queue.NewObservableUpdateHandler(c.rdQueue.GetQueue(), apis.EnableStatusSubresource))
}

func (c *Controller) runRedis(key string) error {
	log.Debugln("started processing, key:", key)
	obj, exists, err := c.rdInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("Redis %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Redis was recreated with the same name
		redis := obj.(*api.Redis).DeepCopy()
		if redis.DeletionTimestamp != nil {
			if core_util.HasFinalizer(redis.ObjectMeta, api.GenericKey) {
				if err := c.terminate(redis); err != nil {
					log.Errorln(err)
					return err
				}
				redis, _, err = util.PatchRedis(c.ExtClient.KubedbV1alpha1(), redis, func(in *api.Redis) *api.Redis {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, api.GenericKey)
					return in
				})
				return err
			}
		} else {
			redis, _, err = util.PatchRedis(c.ExtClient.KubedbV1alpha1(), redis, func(in *api.Redis) *api.Redis {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, api.GenericKey)
				return in
			})
			if err != nil {
				return err
			}
			if err := c.create(redis); err != nil {
				log.Errorln(err)
				c.pushFailureEvent(redis, err.Error())
				return err
			}
		}
	}
	return nil
}
