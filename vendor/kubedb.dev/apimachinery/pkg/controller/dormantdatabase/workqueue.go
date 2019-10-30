package dormantdatabase

import (
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"

	"github.com/appscode/go/log"
	"k8s.io/apimachinery/pkg/labels"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

func (c *Controller) addEventHandler(selector labels.Selector) {
	c.DrmnQueue = queue.New("DormantDatabase", c.MaxNumRequeues, c.NumThreads, c.runDormantDatabase)
	c.DrmnInformer.AddEventHandler(queue.NewFilteredHandler(queue.NewObservableHandler(c.DrmnQueue.GetQueue(), true), selector))
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
				_, _, err = util.PatchDormantDatabase(c.ExtClient.KubedbV1alpha1(), dormantDatabase, func(in *api.DormantDatabase) *api.DormantDatabase {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, api.GenericKey)
					return in
				})
				return err
			}
		} else {
			dormantDatabase, _, err = util.PatchDormantDatabase(c.ExtClient.KubedbV1alpha1(), dormantDatabase, func(in *api.DormantDatabase) *api.DormantDatabase {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, api.GenericKey)
				return in
			})
			if err != nil {
				log.Errorln(err)
				return err
			}
			if err := c.create(dormantDatabase); err != nil {
				log.Errorln(err)
				return err
			}
		}
	}
	return nil
}
