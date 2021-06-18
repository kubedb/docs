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
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"kmodules.xyz/client-go/tools/queue"
)

const (
	systemNamespace    = "kube-system"
	publicNamespace    = "kube-public"
	namespaceKey       = "namespace"
	nameKey            = "name"
	pbAdminDatabase    = "pgbouncer"
	pbAdminPasswordKey = "pb-password"
	pbAdminDataKey     = "pb-admin"
	pbUserDataKey      = "pb-user"
)

func (c *Controller) initWatcher() {
	c.pbInformer = c.KubedbInformerFactory.Kubedb().V1alpha2().PgBouncers().Informer()
	c.pbQueue = queue.New(api.ResourceKindPgBouncer, c.MaxNumRequeues, c.NumThreads, c.runPgBouncer)
	c.pbLister = c.KubedbInformerFactory.Kubedb().V1alpha2().PgBouncers().Lister()
	c.pbInformer.AddEventHandler(queue.NewChangeHandler(c.pbQueue.GetQueue()))
	if c.Auditor != nil {
		c.pbInformer.AddEventHandler(c.Auditor.ForGVK(api.SchemeGroupVersion.WithKind(api.ResourceKindPgBouncer)))
	}
}

func (c *Controller) initAppBindingWatcher() {
	c.appBindingInformer = c.AppCatInformerFactory.Appcatalog().V1alpha1().AppBindings().Informer()
	c.appBindingQueue = queue.New("AppBinding", c.MaxNumRequeues, c.NumThreads, c.manageAppBindingEvent)
	c.appBindingLister = c.AppCatInformerFactory.Appcatalog().V1alpha1().AppBindings().Lister()
	c.appBindingInformer.AddEventHandler(queue.DefaultEventHandler(c.appBindingQueue.GetQueue()))
}

func (c *Controller) initSecretWatcher() {
	c.SecretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if secret, ok := obj.(*core.Secret); ok {
				if key := c.PgBouncerForSecret(secret); key != "" {
					queue.Enqueue(c.pbQueue.GetQueue(), key)
				}
			}
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			if secret, ok := newObj.(*core.Secret); ok {
				if key := c.PgBouncerForSecret(secret); key != "" {
					queue.Enqueue(c.pbQueue.GetQueue(), key)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
		},
	})
}
