package controller

import (
	"github.com/appscode/go/log/golog"
	pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/apimachinery/pkg/controller/dormantdatabase"
	snapc "github.com/kubedb/apimachinery/pkg/controller/snapshot"
	esc "github.com/kubedb/elasticsearch/pkg/controller"
	edc "github.com/kubedb/etcd/pkg/controller"
	mcc "github.com/kubedb/memcached/pkg/controller"
	mgc "github.com/kubedb/mongodb/pkg/controller"
	myc "github.com/kubedb/mysql/pkg/controller"
	pgc "github.com/kubedb/postgres/pkg/controller"
	rdc "github.com/kubedb/redis/pkg/controller"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	AnalyticsClientID string
	EnableAnalytics   = true
	LoggerOptions     golog.Options
)

type OperatorConfig struct {
	amc.Config

	ClientConfig     *rest.Config
	KubeClient       kubernetes.Interface
	APIExtKubeClient crd_cs.ApiextensionsV1beta1Interface
	DBClient         cs.Interface
	PromClient       pcm.MonitoringV1Interface
	CronController   snapc.CronControllerInterface
}

func NewOperatorConfig(clientConfig *rest.Config) *OperatorConfig {
	return &OperatorConfig{
		ClientConfig: clientConfig,
	}
}

func (c *OperatorConfig) New() (*Controller, error) {
	// define all the controllers
	ctrl := New(
		c.KubeClient,
		c.APIExtKubeClient,
		c.DBClient.KubedbV1alpha1(),
		c.PromClient,
		c.CronController,
		c.Config,
	)

	ctrl.DrmnInformer = dormantdatabase.NewController(ctrl.Controller, nil, ctrl.Config, nil).InitInformer()
	ctrl.SnapInformer, ctrl.JobInformer = snapc.NewController(ctrl.Controller, nil, ctrl.Config, nil).InitInformer()

	ctrl.pgCtrl = pgc.New(c.KubeClient, c.APIExtKubeClient, c.DBClient.KubedbV1alpha1(), c.PromClient, c.CronController, ctrl.Config)
	ctrl.esCtrl = esc.New(c.ClientConfig, c.KubeClient, c.APIExtKubeClient, c.DBClient.KubedbV1alpha1(), c.PromClient, c.CronController, ctrl.Config)
	ctrl.edCtrl = edc.New(c.KubeClient, c.APIExtKubeClient, c.DBClient.KubedbV1alpha1(), c.PromClient, c.CronController, ctrl.Config)
	ctrl.mgCtrl = mgc.New(c.KubeClient, c.APIExtKubeClient, c.DBClient.KubedbV1alpha1(), c.PromClient, c.CronController, ctrl.Config)
	ctrl.myCtrl = myc.New(c.KubeClient, c.APIExtKubeClient, c.DBClient.KubedbV1alpha1(), c.PromClient, c.CronController, ctrl.Config)
	ctrl.rdCtrl = rdc.New(c.KubeClient, c.APIExtKubeClient, c.DBClient.KubedbV1alpha1(), c.PromClient, ctrl.Config)
	ctrl.mcCtrl = mcc.New(c.KubeClient, c.APIExtKubeClient, c.DBClient.KubedbV1alpha1(), c.PromClient, ctrl.Config)

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
	if err := c.pgCtrl.Init(); err != nil {
		return err
	}

	if err := c.esCtrl.Init(); err != nil {
		return err
	}

	if err := c.edCtrl.Init(); err != nil {
		return err
	}

	if err := c.mgCtrl.Init(); err != nil {
		return err
	}

	if err := c.myCtrl.Init(); err != nil {
		return err
	}

	if err := c.rdCtrl.Init(); err != nil {
		return err
	}

	if err := c.mcCtrl.Init(); err != nil {
		return err
	}
	return nil
}
