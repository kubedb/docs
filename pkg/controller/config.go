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
	"fmt"

	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amc "kubedb.dev/apimachinery/pkg/controller"
	esc "kubedb.dev/elasticsearch/pkg/controller"
	mrc "kubedb.dev/mariadb/pkg/controller"
	mcc "kubedb.dev/memcached/pkg/controller"
	mgc "kubedb.dev/mongodb/pkg/controller"
	myc "kubedb.dev/mysql/pkg/controller"
	pxc "kubedb.dev/percona-xtradb/pkg/controller"
	pgb "kubedb.dev/pgbouncer/pkg/controller"
	pgc "kubedb.dev/postgres/pkg/controller"
	prc "kubedb.dev/proxysql/pkg/controller"
	rdc "kubedb.dev/redis/pkg/controller"

	pcm "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	auditlib "go.bytebuilders.dev/audit/lib"
	"go.bytebuilders.dev/license-verifier/apis/licenses/v1alpha1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/discovery"
	"kmodules.xyz/client-go/tools/clusterid"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
)

type OperatorConfig struct {
	amc.Config

	LicenseFile      string
	License          v1alpha1.License
	ClientConfig     *rest.Config
	KubeClient       kubernetes.Interface
	CRDClient        crd_cs.Interface
	DBClient         cs.Interface
	DynamicClient    dynamic.Interface
	AppCatalogClient appcat_cs.Interface
	PromClient       pcm.MonitoringV1Interface
	Recorder         record.EventRecorder
}

func NewOperatorConfig(clientConfig *rest.Config) *OperatorConfig {
	return &OperatorConfig{
		ClientConfig: clientConfig,
	}
}

func (c *OperatorConfig) New() (*Controller, error) {
	if err := discovery.IsDefaultSupportedVersion(c.KubeClient); err != nil {
		return nil, err
	}

	topology, err := core_util.DetectTopology(context.TODO(), metadata.NewForConfigOrDie(c.ClientConfig))
	if err != nil {
		return nil, err
	}

	mapper, err := discovery.NewDynamicResourceMapper(c.ClientConfig)
	if err != nil {
		return nil, err
	}

	// audit event auditor
	// WARNING: https://stackoverflow.com/a/46275411/244009
	var auditor *auditlib.EventPublisher
	if c.LicenseFile != "" && !c.License.DisableAnalytics() {
		cmeta, err := clusterid.ClusterMetadata(c.KubeClient.CoreV1().Namespaces())
		if err != nil {
			return nil, fmt.Errorf("failed to extract cluster metadata, reason: %v", err)
		}
		fn := auditlib.BillingEventCreator{
			Mapper:          mapper,
			ClusterMetadata: cmeta,
		}
		auditor = auditlib.NewResilientEventPublisher(func() (*auditlib.NatsConfig, error) {
			return auditlib.NewNatsConfig(cmeta.UID, c.LicenseFile)
		}, mapper, fn.CreateEvent)
		err = auditor.SetupSiteInfoPublisher(c.ClientConfig, c.KubeClient, c.KubeInformerFactory)
		if err != nil {
			return nil, fmt.Errorf("failed to setup site info publisher, reason: %v", err)
		}
	}

	// define all the controllers
	ctrl := New(
		c.ClientConfig,
		c.KubeClient,
		c.CRDClient,
		c.DBClient,
		c.DynamicClient,
		c.AppCatalogClient,
		c.PromClient,
		c.Config,
		topology,
		c.Recorder,
		mapper,
		auditor,
	)

	ctrl.esCtrl = esc.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, ctrl.Config, topology, c.Recorder, mapper, auditor)
	ctrl.mcCtrl = mcc.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.AppCatalogClient, c.PromClient, ctrl.Config, topology, c.Recorder, mapper, auditor)
	ctrl.mrCtrl = mrc.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, ctrl.Config, topology, c.Recorder, mapper, auditor)
	ctrl.mgCtrl = mgc.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, ctrl.Config, topology, c.Recorder, mapper, auditor)
	ctrl.myCtrl = myc.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, ctrl.Config, topology, c.Recorder, mapper, auditor)
	ctrl.pgCtrl = pgc.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, ctrl.Config, topology, c.Recorder, mapper, auditor)
	ctrl.pxCtrl = pxc.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, ctrl.Config, c.Recorder, mapper, auditor)
	ctrl.rdCtrl = rdc.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, ctrl.Config, topology, c.Recorder, mapper, auditor)

	if sets.NewString(c.License.Features...).Has("kubedb-enterprise") {
		ctrl.pgbCtrl = pgb.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, ctrl.Config, topology, c.Recorder, mapper, auditor)
		ctrl.prCtrl = prc.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.DynamicClient, c.PromClient, ctrl.Config, c.Recorder, mapper, auditor)
	}

	if err := ctrl.Init(); err != nil {
		return nil, err
	}

	if auditor != nil {
		if err := auditor.SetupSiteInfoPublisher(ctrl.ClientConfig, ctrl.Client, ctrl.KubeInformerFactory); err != nil {
			return nil, err
		}
	}

	return ctrl, nil
}

// InitInformer initializes MongoDB, DormantDB amd Snapshot watcher
func (c *Controller) Init() error {
	if err := c.EnsureCustomResourceDefinitions(); err != nil {
		return err
	}

	if err := c.esCtrl.Init(); err != nil {
		return err
	}

	if err := c.mcCtrl.Init(); err != nil {
		return err
	}

	if err := c.mgCtrl.Init(); err != nil {
		return err
	}

	if err := c.mrCtrl.Init(); err != nil {
		return err
	}

	if err := c.myCtrl.Init(); err != nil {
		return err
	}

	if c.pgbCtrl != nil {
		if err := c.pgbCtrl.Init(); err != nil {
			return err
		}
	}

	if err := c.pgCtrl.Init(); err != nil {
		return err
	}

	if c.prCtrl != nil {
		if err := c.prCtrl.Init(); err != nil {
			return err
		}
	}

	if err := c.pxCtrl.Init(); err != nil {
		return err
	}

	if err := c.rdCtrl.Init(); err != nil {
		return err
	}
	return nil
}
