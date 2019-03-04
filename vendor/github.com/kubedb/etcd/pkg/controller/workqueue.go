package controller

import (
	"github.com/appscode/go/log"
	"github.com/kubedb/apimachinery/apis"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	kwatch "k8s.io/apimachinery/pkg/watch"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

func (c *Controller) initWatcher() {
	c.etcdInformer = c.KubedbInformerFactory.Kubedb().V1alpha1().Etcds().Informer()
	c.etcdQueue = queue.New("Etcd", c.MaxNumRequeues, c.NumThreads, c.runEtcd)
	c.etcdLister = c.KubedbInformerFactory.Kubedb().V1alpha1().Etcds().Lister()
	c.etcdInformer.AddEventHandler(queue.NewObservableUpdateHandler(c.etcdQueue.GetQueue(), apis.EnableStatusSubresource))
}

func (c *Controller) runEtcd(key string) error {
	log.Debugln("started processing, key:", key)
	obj, exists, err := c.etcdInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("Etcd %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Etcd was recreated with the same name
		etcd := obj.(*api.Etcd).DeepCopy()
		if etcd.DeletionTimestamp != nil {
			ev := &Event{
				Type:   kwatch.Deleted,
				Object: etcd,
			}
			err = c.handleEtcdEvent(ev)
			if core_util.HasFinalizer(etcd.ObjectMeta, api.GenericKey) {
				if err := c.terminate(etcd); err != nil {
					log.Errorln(err)
					return err
				}
				etcd, _, err = util.PatchEtcd(c.ExtClient.KubedbV1alpha1(), etcd, func(in *api.Etcd) *api.Etcd {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, api.GenericKey)
					return in
				})
				return err
			}
		} else {
			etcd, _, err = util.PatchEtcd(c.ExtClient.KubedbV1alpha1(), etcd, func(in *api.Etcd) *api.Etcd {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, api.GenericKey)
				return in
			})
			if err := c.syncEtcd(etcd); err != nil {
				log.Errorln(err)
				c.pushFailureEvent(etcd, err.Error())
				return err
			}
		}
	}
	return nil
}
