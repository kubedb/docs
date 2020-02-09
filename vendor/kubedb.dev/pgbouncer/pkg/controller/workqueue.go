/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"kmodules.xyz/client-go/tools/queue"
)

const (
	systemNamespace = "kube-system"
	publicNamespace = "kube-public"
	namespaceKey    = "namespace"
	nameKey         = "name"
	pbAdminUser     = "kubedb"
	pbAdminDatabase = "pgbouncer"
	pbAdminPassword = "pb-password"
	pbAdminData     = "pb-admin"
	pbUserData      = "pb-user"
)

func (c *Controller) initWatcher() {
	c.pbInformer = c.KubedbInformerFactory.Kubedb().V1alpha1().PgBouncers().Informer()
	c.pbQueue = queue.New("PgBouncer", c.MaxNumRequeues, c.NumThreads, c.managePgBouncerEvent)
	c.pbLister = c.KubedbInformerFactory.Kubedb().V1alpha1().PgBouncers().Lister()
	c.pbInformer.AddEventHandler(queue.NewReconcilableHandler(c.pbQueue.GetQueue()))
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
				pgbouncer, err := c.PgBouncerForSecret(secret)
				if err == nil && pgbouncer != nil {
					queue.Enqueue(c.pbQueue.GetQueue(), pgbouncer)
				}
			}
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			if secret, ok := newObj.(*core.Secret); ok {
				pgbouncer, err := c.PgBouncerForSecret(secret)
				if err == nil && pgbouncer != nil {
					queue.Enqueue(c.pbQueue.GetQueue(), pgbouncer)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if secret, ok := obj.(*core.Secret); ok {
				pgbouncer, err := c.PgBouncerForSecret(secret)
				if err == nil && pgbouncer != nil {
					queue.Enqueue(c.pbQueue.GetQueue(), pgbouncer)
				}
			}
		},
	})
}
