/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"

	"github.com/appscode/go/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

func (c *Controller) initWatcher() {
	c.pgInformer = c.KubedbInformerFactory.Kubedb().V1alpha1().Postgreses().Informer()
	c.pgQueue = queue.New("Postgres", c.MaxNumRequeues, c.NumThreads, c.runPostgres)
	c.pgLister = c.KubedbInformerFactory.Kubedb().V1alpha1().Postgreses().Lister()
	c.pgInformer.AddEventHandler(queue.NewReconcilableHandler(c.pgQueue.GetQueue()))
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
				if err := c.terminate(postgres); err != nil {
					log.Errorln(err)
					return err
				}
				_, _, err = util.PatchPostgres(context.TODO(), c.ExtClient.KubedbV1alpha1(), postgres, func(in *api.Postgres) *api.Postgres {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, api.GenericKey)
					return in
				}, metav1.PatchOptions{})
				return err
			}
		} else {
			postgres, _, err = util.PatchPostgres(context.TODO(), c.ExtClient.KubedbV1alpha1(), postgres, func(in *api.Postgres) *api.Postgres {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, api.GenericKey)
				return in
			}, metav1.PatchOptions{})
			if err != nil {
				return err
			}

			if postgres.Spec.Paused {
				return nil
			}

			if postgres.Spec.Halted {
				if err := c.halt(postgres); err != nil {
					log.Errorln(err)
					c.pushFailureEvent(postgres, err.Error())
					return err
				}
			} else {
				if err := c.create(postgres); err != nil {
					log.Errorln(err)
					c.pushFailureEvent(postgres, err.Error())
					return err
				}
			}
		}
	}
	return nil
}
