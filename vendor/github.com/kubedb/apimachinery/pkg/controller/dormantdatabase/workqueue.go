package dormantdatabase

import (
	"github.com/appscode/go/log"
	core_util "github.com/appscode/kutil/core/v1"
	"github.com/appscode/kutil/tools/queue"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *Controller) addEventHandler(selector labels.Selector) {
	c.DrmnQueue = queue.New("DormantDatabase", c.MaxNumRequeues, c.NumThreads, c.runDormantDatabase)
	c.DrmnInformer.AddEventHandler(queue.NewFilteredHandler(queue.NewObservableHandler(c.DrmnQueue.GetQueue(), api.EnableStatusSubresource), selector))
	c.ddbLister = c.KubedbInformerFactory.Kubedb().V1alpha1().DormantDatabases().Lister()
}

func (c *Controller) runDormantDatabase(key string) error {
	log.Debugf("started processing, key: %v", key)
	obj, exists, err := c.DrmnInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("DormantDatabase %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a DormantDatabase was recreated with the same name
		dormantDatabase := obj.(*api.DormantDatabase).DeepCopy()
		if dormantDatabase.DeletionTimestamp != nil {
			if core_util.HasFinalizer(dormantDatabase.ObjectMeta, api.GenericKey) {
				if err := c.delete(dormantDatabase); err != nil {
					log.Errorln(err)
					return err
				}
				dormantDatabase, _, err = util.PatchDormantDatabase(c.ExtClient, dormantDatabase, func(in *api.DormantDatabase) *api.DormantDatabase {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, api.GenericKey)
					return in
				})
				return err
			}
		} else {
			dormantDatabase, _, err = util.PatchDormantDatabase(c.ExtClient, dormantDatabase, func(in *api.DormantDatabase) *api.DormantDatabase {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, api.GenericKey)
				return in
			})
			if err := c.create(dormantDatabase); err != nil {
				log.Errorln(err)
				return err
			}
		}
	}
	return nil
}
