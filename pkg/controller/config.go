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

	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amc "kubedb.dev/apimachinery/pkg/controller"
	sts "kubedb.dev/apimachinery/pkg/controller/statefulset"
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
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	reg_util "kmodules.xyz/client-go/admissionregistration/v1beta1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/discovery"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
)

const (
	mutatingWebhookConfig   = "mutators.kubedb.com"
	validatingWebhookConfig = "validators.kubedb.com"
)

type OperatorConfig struct {
	amc.Config

	LicenseFile      string
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
	)

	ctrl.esCtrl = esc.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, ctrl.Config, topology, c.Recorder)
	ctrl.mcCtrl = mcc.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.AppCatalogClient, c.PromClient, ctrl.Config, topology, c.Recorder)
	ctrl.mgCtrl = mgc.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, ctrl.Config, topology, c.Recorder)
	ctrl.myCtrl = myc.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, ctrl.Config, c.Recorder)
	ctrl.pgbCtrl = pgb.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, ctrl.Config, topology, c.Recorder)
	ctrl.pgCtrl = pgc.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, ctrl.Config, topology, c.Recorder)
	ctrl.prCtrl = prc.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.DynamicClient, c.PromClient, ctrl.Config, c.Recorder)
	ctrl.pxCtrl = pxc.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, ctrl.Config, c.Recorder)
	ctrl.rdCtrl = rdc.New(c.ClientConfig, c.KubeClient, c.CRDClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, ctrl.Config, topology, c.Recorder)

	if err := ctrl.Init(); err != nil {
		return nil, err
	}

	return ctrl, nil
}

// InitInformer initializes MongoDB, DormantDB amd Snapshot watcher
func (c *Controller) Init() error {
	if err := c.EnsureCustomResourceDefinitions(); err != nil {
		return err
	}
	if c.EnableMutatingWebhook {
		if err := reg_util.UpdateMutatingWebhookCABundle(c.ClientConfig, mutatingWebhookConfig); err != nil {
			return err
		}
	}
	if c.EnableValidatingWebhook {
		if err := reg_util.UpdateValidatingWebhookCABundle(c.ClientConfig, validatingWebhookConfig); err != nil {
			return err
		}
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

	if err := c.myCtrl.Init(); err != nil {
		return err
	}

	if err := c.pgbCtrl.Init(); err != nil {
		return err
	}

	if err := c.pgCtrl.Init(); err != nil {
		return err
	}

	if err := c.prCtrl.Init(); err != nil {
		return err
	}

	if err := c.pxCtrl.Init(); err != nil {
		return err
	}

	if err := c.rdCtrl.Init(); err != nil {
		return err
	}

	// Initialize StatefulSet watcher
	sts.NewController(&c.Config, c.Client, c.DBClient, c.DynamicClient).InitStsWatcher()

	return nil
}
