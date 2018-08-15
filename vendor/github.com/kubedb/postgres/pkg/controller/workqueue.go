package controller

import (
	"github.com/appscode/go/log"
	core_util "github.com/appscode/kutil/core/v1"
	"github.com/appscode/kutil/tools/queue"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
)

func (c *Controller) initWatcher() {
	c.pgInformer = c.KubedbInformerFactory.Kubedb().V1alpha1().Postgreses().Informer()
	c.pgQueue = queue.New("Postgres", c.MaxNumRequeues, c.NumThreads, c.runPostgres)
	c.pgLister = c.KubedbInformerFactory.Kubedb().V1alpha1().Postgreses().Lister()
	c.pgInformer.AddEventHandler(queue.NewEventHandler(c.pgQueue.GetQueue(), func(old interface{}, new interface{}) bool {
		oldObj := old.(*api.Postgres)
		newObj := new.(*api.Postgres)
		return newObj.DeletionTimestamp != nil || !newObj.AlreadyObserved(oldObj)
	}))
}

func (c *Controller) runPostgres(key string) error {
	log.Debugln("started processing, key:", key)
	obj, exists, err := c.pgInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("Postgres %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Postgres was recreated with the same name
		postgres := obj.(*api.Postgres).DeepCopy()
		if postgres.DeletionTimestamp != nil {
			if core_util.HasFinalizer(postgres.ObjectMeta, api.GenericKey) {
				if err := c.pause(postgres); err != nil {
					log.Errorln(err)
					return err
				}
				postgres, _, err = util.PatchPostgres(c.ExtClient, postgres, func(in *api.Postgres) *api.Postgres {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, api.GenericKey)
					return in
				})
				return err
			}
		} else {
			postgres, _, err = util.PatchPostgres(c.ExtClient, postgres, func(in *api.Postgres) *api.Postgres {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, api.GenericKey)
				return in
			})
			if err != nil {
				return err
			}
			if err := c.create(postgres); err != nil {
				log.Errorln(err)
				c.pushFailureEvent(postgres, err.Error())
				return err
			}
		}
	}
	return nil
}
