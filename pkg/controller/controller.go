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
	catalogapi "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amc "kubedb.dev/apimachinery/pkg/controller"
	snapc "kubedb.dev/apimachinery/pkg/controller/snapshot"
	esc "kubedb.dev/elasticsearch/pkg/controller"
	edc "kubedb.dev/etcd/pkg/controller"
	mcc "kubedb.dev/memcached/pkg/controller"
	mgc "kubedb.dev/mongodb/pkg/controller"
	myc "kubedb.dev/mysql/pkg/controller"
	pxc "kubedb.dev/percona-xtradb/pkg/controller"
	pgb "kubedb.dev/pgbouncer/pkg/controller"
	pgc "kubedb.dev/postgres/pkg/controller"
	prc "kubedb.dev/proxysql/pkg/controller"
	rdc "kubedb.dev/redis/pkg/controller"

	"github.com/appscode/go/log"
	pcm "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	apiext_util "kmodules.xyz/client-go/apiextensions/v1beta1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
	scs "stash.appscode.dev/stash/client/clientset/versioned"
)

type Controller struct {
	amc.Config
	*amc.Controller
	promClient     pcm.MonitoringV1Interface
	cronController snapc.CronControllerInterface

	// DB controllers
	edCtrl  *edc.Controller
	esCtrl  *esc.Controller
	mcCtrl  *mcc.Controller
	mgCtrl  *mgc.Controller
	myCtrl  *myc.Controller
	pgbCtrl *pgb.Controller
	pgCtrl  *pgc.Controller
	prCtrl  *prc.Controller
	pxCtrl  *pxc.Controller
	rdCtrl  *rdc.Controller
}

func New(
	clientConfig *rest.Config,
	client kubernetes.Interface,
	apiExtKubeClient crd_cs.ApiextensionsV1beta1Interface,
	dbClient cs.Interface,
	stashClient scs.Interface,
	dynamicClient dynamic.Interface,
	appCatalogClient appcat_cs.Interface,
	promClient pcm.MonitoringV1Interface,
	cronController snapc.CronControllerInterface,
	opt amc.Config,
) *Controller {
	return &Controller{
		Controller: &amc.Controller{
			ClientConfig:     clientConfig,
			Client:           client,
			ExtClient:        dbClient,
			ApiExtKubeClient: apiExtKubeClient,
			StashClient:      stashClient,
			DynamicClient:    dynamicClient,
			AppCatalogClient: appCatalogClient,
		},
		Config:         opt,
		promClient:     promClient,
		cronController: cronController,
	}
}

// EnsureCustomResourceDefinitions ensures CRD for MySQl, DormantDatabase and Snapshot
func (c *Controller) EnsureCustomResourceDefinitions() error {
	log.Infoln("Ensuring CustomResourceDefinition...")
	crds := []*crd_api.CustomResourceDefinition{
		dbapi.DormantDatabase{}.CustomResourceDefinition(),
		dbapi.Elasticsearch{}.CustomResourceDefinition(),
		dbapi.Etcd{}.CustomResourceDefinition(),
		dbapi.Memcached{}.CustomResourceDefinition(),
		dbapi.MongoDB{}.CustomResourceDefinition(),
		dbapi.MySQL{}.CustomResourceDefinition(),
		dbapi.PerconaXtraDB{}.CustomResourceDefinition(),
		dbapi.PgBouncer{}.CustomResourceDefinition(),
		dbapi.Postgres{}.CustomResourceDefinition(),
		dbapi.ProxySQL{}.CustomResourceDefinition(),
		dbapi.Redis{}.CustomResourceDefinition(),
		dbapi.Snapshot{}.CustomResourceDefinition(),

		catalogapi.ElasticsearchVersion{}.CustomResourceDefinition(),
		catalogapi.EtcdVersion{}.CustomResourceDefinition(),
		catalogapi.MemcachedVersion{}.CustomResourceDefinition(),
		catalogapi.MongoDBVersion{}.CustomResourceDefinition(),
		catalogapi.MySQLVersion{}.CustomResourceDefinition(),
		catalogapi.PerconaXtraDBVersion{}.CustomResourceDefinition(),
		catalogapi.PgBouncerVersion{}.CustomResourceDefinition(),
		catalogapi.PostgresVersion{}.CustomResourceDefinition(),
		catalogapi.ProxySQLVersion{}.CustomResourceDefinition(),
		catalogapi.RedisVersion{}.CustomResourceDefinition(),

		appcat.AppBinding{}.CustomResourceDefinition(),
	}
	return apiext_util.RegisterCRDs(c.ApiExtKubeClient, crds)
}
