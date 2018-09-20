package controller

import (
	"github.com/appscode/go/log"
	core_util "github.com/appscode/kutil/core/v1"
	"github.com/appscode/kutil/tools/queue"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
)

func (c *Controller) initWatcher() {
	c.myInformer = c.KubedbInformerFactory.Kubedb().V1alpha1().MySQLs().Informer()
	c.myQueue = queue.New("MySQL", c.MaxNumRequeues, c.NumThreads, c.runMySQL)
	c.myLister = c.KubedbInformerFactory.Kubedb().V1alpha1().MySQLs().Lister()
	c.myInformer.AddEventHandler(queue.NewObservableHandler(c.myQueue.GetQueue(), api.EnableStatusSubresource))
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
		mysql := obj.(*api.MySQL).DeepCopy()
		if mysql.DeletionTimestamp != nil {
			if core_util.HasFinalizer(mysql.ObjectMeta, api.GenericKey) {
				if err := c.terminate(mysql); err != nil {
					log.Errorln(err)
					return err
				}
				mysql, _, err = util.PatchMySQL(c.ExtClient, mysql, func(in *api.MySQL) *api.MySQL {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, api.GenericKey)
					return in
				})
				return err
			}
		} else {
			mysql, _, err = util.PatchMySQL(c.ExtClient, mysql, func(in *api.MySQL) *api.MySQL {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, api.GenericKey)
				return in
			})
			if err != nil {
				return err
			}
			if err := c.create(mysql); err != nil {
				log.Errorln(err)
				c.pushFailureEvent(mysql, err.Error())
				return err
			}
		}
	}
	return nil
}
