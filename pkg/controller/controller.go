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
	catalogapi "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amc "kubedb.dev/apimachinery/pkg/controller"
	esc "kubedb.dev/elasticsearch/pkg/controller"
	mcc "kubedb.dev/memcached/pkg/controller"
	mgc "kubedb.dev/mongodb/pkg/controller"
	myc "kubedb.dev/mysql/pkg/controller"
	pxc "kubedb.dev/percona-xtradb/pkg/controller"
	pgb "kubedb.dev/pgbouncer/pkg/controller"
	pgc "kubedb.dev/postgres/pkg/controller"
	prc "kubedb.dev/proxysql/pkg/controller"
	rdc "kubedb.dev/redis/pkg/controller"

	pcm "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"gomodules.xyz/x/log"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"kmodules.xyz/client-go/apiextensions"
	core_util "kmodules.xyz/client-go/core/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
)

type Controller struct {
	amc.Config
	*amc.Controller
	promClient pcm.MonitoringV1Interface

	// LabelSelector to filter Stash restore invokers only for KubeDB
	selector metav1.LabelSelector

	// DB controllers
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
	crdClient crd_cs.Interface,
	dbClient cs.Interface,
	dynamicClient dynamic.Interface,
	appCatalogClient appcat_cs.Interface,
	promClient pcm.MonitoringV1Interface,
	opt amc.Config,
	topology *core_util.Topology,
	recorder record.EventRecorder,
) *Controller {
	return &Controller{
		Controller: &amc.Controller{
			ClientConfig:     clientConfig,
			Client:           client,
			DBClient:         dbClient,
			CRDClient:        crdClient,
			DynamicClient:    dynamicClient,
			AppCatalogClient: appCatalogClient,
			ClusterTopology:  topology,
			Recorder:         recorder,
		},
		Config:     opt,
		promClient: promClient,
		selector: metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      dbapi.LabelDatabaseKind,
					Operator: metav1.LabelSelectorOpExists,
				},
			},
		},
	}
}

// EnsureCustomResourceDefinitions ensures CRD for MySQl, DormantDatabase and Snapshot
func (c *Controller) EnsureCustomResourceDefinitions() error {
	log.Infoln("Ensuring CustomResourceDefinition...")
	crds := []*apiextensions.CustomResourceDefinition{
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
	return apiextensions.RegisterCRDs(c.CRDClient, crds)
}
