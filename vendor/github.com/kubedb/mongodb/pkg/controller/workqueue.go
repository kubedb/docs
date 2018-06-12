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
	c.mgInformer = c.KubedbInformerFactory.Kubedb().V1alpha1().MongoDBs().Informer()
	c.mgQueue = queue.New("MongoDB", c.MaxNumRequeues, c.NumThreads, c.runMongoDB)
	c.mgLister = c.KubedbInformerFactory.Kubedb().V1alpha1().MongoDBs().Lister()
	c.mgInformer.AddEventHandler(queue.NewEventHandler(c.mgQueue.GetQueue(), func(old interface{}, new interface{}) bool {
		oldObj := old.(*api.MongoDB)
		newObj := new.(*api.MongoDB)
		return newObj.DeletionTimestamp != nil || !mongodbEqual(oldObj, newObj)
	}))
}

func mongodbEqual(old, new *api.MongoDB) bool {
	if !meta_util.Equal(old.Spec, new.Spec) {
		diff := meta_util.Diff(old.Spec, new.Spec)
		log.Infof("MongoDB %s/%s has changed. Diff: %s", new.Namespace, new.Name, diff)
		return false
	}
	if !meta_util.Equal(old.Annotations, new.Annotations) {
		diff := meta_util.Diff(old.Annotations, new.Annotations)
		log.Infof("Annotations in MongoDB %s/%s has changed. Diff: %s\n", new.Namespace, new.Name, diff)
		return false
	}
	return true
}

func (c *Controller) runMongoDB(key string) error {
	log.Debugln("started processing, key:", key)
	obj, exists, err := c.mgInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("MongoDB %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a MongoDB was recreated with the same name
		mongodb := obj.(*api.MongoDB).DeepCopy()
		if mongodb.DeletionTimestamp != nil {
			if core_util.HasFinalizer(mongodb.ObjectMeta, api.GenericKey) {
				util.AssignTypeKind(mongodb)
				if err := c.pause(mongodb); err != nil {
					log.Errorln(err)
					return err
				}
				mongodb, _, err = util.PatchMongoDB(c.ExtClient, mongodb, func(in *api.MongoDB) *api.MongoDB {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, api.GenericKey)
					return in
				})
				return err
			}
		} else {
			mongodb, _, err = util.PatchMongoDB(c.ExtClient, mongodb, func(in *api.MongoDB) *api.MongoDB {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, api.GenericKey)
				return in
			})
			if err != nil {
				return err
			}
			util.AssignTypeKind(mongodb)
			if err := c.create(mongodb); err != nil {
				log.Errorln(err)
				c.pushFailureEvent(mongodb, err.Error())
				return err
			}
		}
	}
	return nil
}
