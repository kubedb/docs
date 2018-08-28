package dormantdatabase

import (
	"github.com/appscode/go/log"
	"github.com/appscode/kutil/tools/queue"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
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
		if err := c.create(dormantDatabase); err != nil {
			log.Errorln(err)
			return err
		}
	}
	return nil
}
