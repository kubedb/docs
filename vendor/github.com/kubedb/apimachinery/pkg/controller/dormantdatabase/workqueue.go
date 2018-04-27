package dormantdatabase

import (
	"github.com/appscode/go/log"
	meta_util "github.com/appscode/kutil/meta"
	"github.com/appscode/kutil/tools/queue"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *Controller) addEventHandler(selector labels.Selector) {
	c.DrmnQueue = queue.New("DormantDatabase", c.MaxNumRequeues, c.NumThreads, c.runDormantDatabase)
	c.DrmnInformer.AddEventHandler(queue.NewFilteredHandler(queue.NewEventHandler(c.DrmnQueue.GetQueue(), func(old interface{}, new interface{}) bool {
		oldObj := old.(*api.DormantDatabase)
		newObj := new.(*api.DormantDatabase)
		if !dormantDatabaseEqual(oldObj, newObj) {
			return true
		}
		return false
	}), selector))
	c.ddbLister = c.KubedbInformerFactory.Kubedb().V1alpha1().DormantDatabases().Lister()
}

func dormantDatabaseEqual(old, new *api.DormantDatabase) bool {
	if !meta_util.Equal(old.Spec, new.Spec) {
		diff := meta_util.Diff(old.Spec, new.Spec)
		log.Debugf("DormantDatabase %s/%s has changed. Diff: %s\n", new.Namespace, new.Name, diff)
		return false
	}
	return true
}

func (c *Controller) runDormantDatabase(key string) error {
	log.Debugf("started processing, key: %v\n", key)
	obj, exists, err := c.DrmnInformer.GetIndexer().GetByKey(key)
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
		util.AssignTypeKind(dormantDatabase)
		if err := c.create(dormantDatabase); err != nil {
			log.Errorln(err)
			return err
		}
	}
	return nil
}
