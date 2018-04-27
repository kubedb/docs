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
	c.esInformer = c.KubedbInformerFactory.Kubedb().V1alpha1().Elasticsearches().Informer()
	c.esQueue = queue.New("Elasticsearch", c.MaxNumRequeues, c.NumThreads, c.runElasticsearch)
	c.esLister = c.KubedbInformerFactory.Kubedb().V1alpha1().Elasticsearches().Lister()
	c.esInformer.AddEventHandler(queue.NewEventHandler(c.esQueue.GetQueue(), func(old interface{}, new interface{}) bool {
		oldObj := old.(*api.Elasticsearch)
		newObj := new.(*api.Elasticsearch)
		return newObj.DeletionTimestamp != nil || !elasticsearchEqual(oldObj, newObj)
	}))
}

func elasticsearchEqual(old, new *api.Elasticsearch) bool {
	if !meta_util.Equal(old.Spec, new.Spec) {
		diff := meta_util.Diff(old.Spec, new.Spec)
		log.Infof("Elasticsearch %s/%s has changed. Diff: %s\n", new.Namespace, new.Name, diff)
		return false
	}
	if !meta_util.Equal(old.Annotations, new.Annotations) {
		diff := meta_util.Diff(old.Annotations, new.Annotations)
		log.Infof("Annotations in Elasticsearch %s/%s has changed. Diff: %s\n", new.Namespace, new.Name, diff)
		return false
	}
	return true
}

func (c *Controller) runElasticsearch(key string) error {
	log.Debugf("started processing, key: %v\n", key)
	obj, exists, err := c.esInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v\n", key, err)
		return err
	}

	if !exists {
		log.Debugf("Elasticsearch %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Elasticsearch was recreated with the same name
		elasticsearch := obj.(*api.Elasticsearch).DeepCopy()
		if elasticsearch.DeletionTimestamp != nil {
			if core_util.HasFinalizer(elasticsearch.ObjectMeta, "kubedb.com") {
				util.AssignTypeKind(elasticsearch)
				if err := c.pause(elasticsearch); err != nil {
					log.Errorln(err)
					return err
				}
				elasticsearch, _, err = util.PatchElasticsearch(c.ExtClient, elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, "kubedb.com")
					return in
				})
				return err
			}
		} else {
			elasticsearch, _, err = util.PatchElasticsearch(c.ExtClient, elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, "kubedb.com")
				return in
			})
			util.AssignTypeKind(elasticsearch)
			if err := c.create(elasticsearch); err != nil {
				log.Errorln(err)
				c.pushFailureEvent(elasticsearch, err.Error())
				return err
			}
		}
	}
	return nil
}
