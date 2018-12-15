package controller

import (
	reg_util "github.com/appscode/kutil/admissionregistration/v1beta1"
	"github.com/appscode/kutil/discovery"
	pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/apimachinery/pkg/controller/dormantdatabase"
	snapc "github.com/kubedb/apimachinery/pkg/controller/snapshot"
	"github.com/kubedb/apimachinery/pkg/eventer"
	esc "github.com/kubedb/elasticsearch/pkg/controller"
	edc "github.com/kubedb/etcd/pkg/controller"
	mcc "github.com/kubedb/memcached/pkg/controller"
	mgc "github.com/kubedb/mongodb/pkg/controller"
	myc "github.com/kubedb/mysql/pkg/controller"
	pgc "github.com/kubedb/postgres/pkg/controller"
	rdc "github.com/kubedb/redis/pkg/controller"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

const (
	mutatingWebhookConfig   = "mutators.kubedb.com"
	validatingWebhookConfig = "validators.kubedb.com"
)

type OperatorConfig struct {
	amc.Config

	ClientConfig     *rest.Config
	KubeClient       kubernetes.Interface
	APIExtKubeClient crd_cs.ApiextensionsV1beta1Interface
	DBClient         cs.Interface
	DynamicClient    dynamic.Interface
	AppCatalogClient appcat_cs.AppcatalogV1alpha1Interface
	PromClient       pcm.MonitoringV1Interface
	CronController   snapc.CronControllerInterface
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

	recorder := eventer.NewEventRecorder(c.KubeClient, "KubeDB operator")

	// define all the controllers
	ctrl := New(
		c.ClientConfig,
		c.KubeClient,
		c.APIExtKubeClient,
		c.DBClient,
		c.DynamicClient,
		c.AppCatalogClient,
		c.PromClient,
		c.CronController,
		c.Config,
	)

	ctrl.DrmnInformer = dormantdatabase.NewController(ctrl.Controller, nil, ctrl.Config, nil, recorder).InitInformer()
	ctrl.SnapInformer, ctrl.JobInformer = snapc.NewController(ctrl.Controller, nil, ctrl.Config, nil, recorder).InitInformer()

	ctrl.pgCtrl = pgc.New(c.ClientConfig, c.KubeClient, c.APIExtKubeClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, c.CronController, ctrl.Config, recorder)
	ctrl.esCtrl = esc.New(c.ClientConfig, c.KubeClient, c.APIExtKubeClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, c.CronController, ctrl.Config, recorder)
	ctrl.edCtrl = edc.New(c.ClientConfig, c.KubeClient, c.APIExtKubeClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, c.CronController, ctrl.Config, recorder)
	ctrl.mgCtrl = mgc.New(c.ClientConfig, c.KubeClient, c.APIExtKubeClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, c.CronController, ctrl.Config, recorder)
	ctrl.myCtrl = myc.New(c.ClientConfig, c.KubeClient, c.APIExtKubeClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, c.CronController, ctrl.Config, recorder)
	ctrl.rdCtrl = rdc.New(c.ClientConfig, c.KubeClient, c.APIExtKubeClient, c.DBClient, c.DynamicClient, c.AppCatalogClient, c.PromClient, ctrl.Config, recorder)
	ctrl.mcCtrl = mcc.New(c.ClientConfig, c.KubeClient, c.APIExtKubeClient, c.DBClient, c.AppCatalogClient, c.PromClient, ctrl.Config, recorder)

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
