package controller

import (
	"github.com/appscode/go/log/golog"
	reg_util "github.com/appscode/kutil/admissionregistration/v1beta1"
	"github.com/appscode/kutil/discovery"
	pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/apimachinery/pkg/controller/dormantdatabase"
	snapc "github.com/kubedb/apimachinery/pkg/controller/snapshot"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	mutatingWebhookConfig   = "mutators.kubedb.com"
	validatingWebhookConfig = "validators.kubedb.com"
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
	if err := discovery.IsDefaultSupportedVersion(c.KubeClient); err != nil {
		return nil, err
	}

	ctrl := New(
		c.ClientConfig,
		c.KubeClient,
		c.APIExtKubeClient,
		c.DBClient,
		c.PromClient,
		c.Config,
	)

	tweakListOptions := func(options *metav1.ListOptions) {
		options.LabelSelector = ctrl.selector.String()
	}

	// Initialize Job and Snapshot Informer. Later EventHandler will be added to these informers.
	ctrl.DrmnInformer = dormantdatabase.NewController(ctrl.Controller, ctrl, ctrl.Config, tweakListOptions).InitInformer()

	if err := ctrl.EnsureCustomResourceDefinitions(); err != nil {
		return nil, err
	}
	if c.EnableMutatingWebhook {
		if err := reg_util.UpdateMutatingWebhookCABundle(c.ClientConfig, mutatingWebhookConfig); err != nil {
			return nil, err
		}
	}
	if c.EnableValidatingWebhook {
		if err := reg_util.UpdateValidatingWebhookCABundle(c.ClientConfig, validatingWebhookConfig); err != nil {
			return nil, err
		}
	}

	if err := ctrl.Init(); err != nil {
		return nil, err
	}

	return ctrl, nil
}