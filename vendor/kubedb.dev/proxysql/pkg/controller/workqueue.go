/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	"gomodules.xyz/x/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

func (c *Controller) initWatcher() {
	c.proxysqlInformer = c.KubedbInformerFactory.Kubedb().V1alpha2().ProxySQLs().Informer()
	c.proxysqlQueue = queue.New("ProxySQL", c.MaxNumRequeues, c.NumThreads, c.runProxySQL)
	c.proxysqlLister = c.KubedbInformerFactory.Kubedb().V1alpha2().ProxySQLs().Lister()
	c.proxysqlInformer.AddEventHandler(queue.NewChangeHandler(c.proxysqlQueue.GetQueue()))
}

func (c *Controller) runProxySQL(key string) error {
	log.Debugln("started processing, key:", key)
	obj, exists, err := c.proxysqlInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("ProxySQL %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a ProxySQL was recreated with the same name
		proxysql := obj.(*api.ProxySQL).DeepCopy()
		if proxysql.DeletionTimestamp != nil {
			if core_util.HasFinalizer(proxysql.ObjectMeta, kubedb.GroupName) {
				if err := c.terminate(proxysql); err != nil {
					log.Errorln(err)
					return err
				}
				_, _, err = util.PatchProxySQL(context.TODO(), c.DBClient.KubedbV1alpha2(), proxysql, func(in *api.ProxySQL) *api.ProxySQL {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, kubedb.GroupName)
					return in
				}, metav1.PatchOptions{})
				return err
			}
		} else {
			proxysql, _, err = util.PatchProxySQL(context.TODO(), c.DBClient.KubedbV1alpha2(), proxysql, func(in *api.ProxySQL) *api.ProxySQL {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, kubedb.GroupName)
				return in
			}, metav1.PatchOptions{})
			if err != nil {
				return err
			}
			if err := c.create(proxysql); err != nil {
				log.Errorln(err)
				c.pushFailureEvent(proxysql, err.Error())
				return err
			}
		}
	}
	return nil
}
