package snapshot

import (
	"github.com/appscode/go/log"
	core_util "github.com/appscode/kutil/core/v1"
	"github.com/appscode/kutil/tools/queue"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *Controller) addEventHandler(selector labels.Selector) {
	c.SnapQueue = queue.New("Snapshot", c.MaxNumRequeues, c.NumThreads, c.runSnapshot)
	c.snLister = c.KubedbInformerFactory.Kubedb().V1alpha1().Snapshots().Lister()
	c.SnapInformer.AddEventHandler(queue.NewFilteredHandler(queue.NewEventHandler(c.SnapQueue.GetQueue(), func(old interface{}, new interface{}) bool {
		snapshot := new.(*api.Snapshot)
		return snapshot.DeletionTimestamp != nil
	}), selector))
}

func (c *Controller) runSnapshot(key string) error {
	log.Debugf("started processing, key: %v", key)
	obj, exists, err := c.SnapInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("Snapshot %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Snapshot was recreated with the same name
		snapshot := obj.(*api.Snapshot).DeepCopy()
		if snapshot.DeletionTimestamp != nil {
			if core_util.HasFinalizer(snapshot.ObjectMeta, api.GenericKey) {
				if err := c.delete(snapshot); err != nil {
					log.Errorln(err)
					return err
				}
				snapshot, _, err = util.PatchSnapshot(c.ExtClient, snapshot, func(in *api.Snapshot) *api.Snapshot {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, api.GenericKey)
					return in
				})
				return err
			}
		} else {
			snapshot, _, err = util.PatchSnapshot(c.ExtClient, snapshot, func(in *api.Snapshot) *api.Snapshot {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, api.GenericKey)
				return in
			})
			if err := c.create(snapshot); err != nil {
				log.Errorln(err)
				return err
			}
		}
	}
	return nil
}
