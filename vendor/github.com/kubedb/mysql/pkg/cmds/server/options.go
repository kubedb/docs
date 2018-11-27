package server

import (
	"flag"
	"time"

	"github.com/appscode/kutil/meta"
	prom "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/kubedb/apimachinery/apis"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned"
	kubedbinformers "github.com/kubedb/apimachinery/client/informers/externalversions"
	snapc "github.com/kubedb/apimachinery/pkg/controller/snapshot"
	"github.com/kubedb/mysql/pkg/controller"
	"github.com/spf13/pflag"
	core "k8s.io/api/core/v1"
	kext_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

type ExtraOptions struct {
	EnableRBAC                  bool
	OperatorNamespace           string
	RestrictToOperatorNamespace bool
	GoverningService            string
	QPS                         float64
	Burst                       int
	ResyncPeriod                time.Duration
	MaxNumRequeues              int
	NumThreads                  int

	PrometheusCrdGroup string
	PrometheusCrdKinds prom.CrdKinds

	EnableMutatingWebhook   bool
	EnableValidatingWebhook bool
}

func (s ExtraOptions) WatchNamespace() string {
	if s.RestrictToOperatorNamespace {
		return s.OperatorNamespace
	}
	return core.NamespaceAll
}

func NewExtraOptions() *ExtraOptions {
	return &ExtraOptions{
		EnableRBAC:        true,
		OperatorNamespace: meta.Namespace(),
		GoverningService:  "kubedb",
		ResyncPeriod:      10 * time.Minute,
		MaxNumRequeues:    5,
		NumThreads:        2,
		// ref: https://github.com/kubernetes/ingress-nginx/blob/e4d53786e771cc6bdd55f180674b79f5b692e552/pkg/ingress/controller/launch.go#L252-L259
		// High enough QPS to fit all expected use cases. QPS=0 is not set here, because client code is overriding it.
		QPS: 1e6,
		// High enough Burst to fit all expected use cases. Burst=0 is not set here, because client code is overriding it.
		Burst:              1e6,
		PrometheusCrdGroup: prom.Group,
		PrometheusCrdKinds: prom.DefaultCrdKinds,
	}
}

func (s *ExtraOptions) AddGoFlags(fs *flag.FlagSet) {
	fs.StringVar(&s.GoverningService, "governing-service", s.GoverningService, "Governing service for database statefulset")
	fs.BoolVar(&s.EnableRBAC, "rbac", s.EnableRBAC, "Enable RBAC for operator & offshoot Kubernetes objects")

	fs.Float64Var(&s.QPS, "qps", s.QPS, "The maximum QPS to the master from this client")
	fs.IntVar(&s.Burst, "burst", s.Burst, "The maximum burst for throttle")
	fs.DurationVar(&s.ResyncPeriod, "resync-period", s.ResyncPeriod, "If non-zero, will re-list this often. Otherwise, re-list will be delayed aslong as possible (until the upstream source closes the watch or times out.")

	fs.BoolVar(&s.RestrictToOperatorNamespace, "restrict-to-operator-namespace", s.RestrictToOperatorNamespace, "If true, KubeDB operator will only handle Kubernetes objects in its own namespace.")

	fs.StringVar(&s.PrometheusCrdGroup, "prometheus-crd-apigroup", s.PrometheusCrdGroup, "prometheus CRD  API group name")
	fs.Var(&s.PrometheusCrdKinds, "prometheus-crd-kinds", " - EXPERIMENTAL (could be removed in future releases) - customize CRD kind names")

	fs.BoolVar(&s.EnableMutatingWebhook, "enable-mutating-webhook", s.EnableMutatingWebhook, "If true, enables mutating webhooks for KubeDB CRDs.")
	fs.BoolVar(&s.EnableValidatingWebhook, "enable-validating-webhook", s.EnableValidatingWebhook, "If true, enables validating webhooks for KubeDB CRDs.")
	fs.BoolVar(&apis.EnableStatusSubresource, "enable-status-subresource", apis.EnableStatusSubresource, "If true, uses sub resource for KubeDB crds.")
}

func (s *ExtraOptions) AddFlags(fs *pflag.FlagSet) {
	pfs := flag.NewFlagSet("mysql-server", flag.ExitOnError)
	s.AddGoFlags(pfs)
	fs.AddGoFlagSet(pfs)
}

func (s *ExtraOptions) ApplyTo(cfg *controller.OperatorConfig) error {
	var err error

	cfg.EnableRBAC = s.EnableRBAC
	cfg.OperatorNamespace = s.OperatorNamespace
	cfg.GoverningService = s.GoverningService

	cfg.EnableAnalytics = controller.EnableAnalytics
	cfg.AnalyticsClientID = controller.AnalyticsClientID
	cfg.LoggerOptions = controller.LoggerOptions

	cfg.ClientConfig.QPS = float32(s.QPS)
	cfg.ClientConfig.Burst = s.Burst
	cfg.ResyncPeriod = s.ResyncPeriod
	cfg.MaxNumRequeues = s.MaxNumRequeues
	cfg.NumThreads = s.NumThreads
	cfg.WatchNamespace = s.WatchNamespace()
	cfg.EnableMutatingWebhook = s.EnableMutatingWebhook
	cfg.EnableValidatingWebhook = s.EnableValidatingWebhook

	if cfg.KubeClient, err = kubernetes.NewForConfig(cfg.ClientConfig); err != nil {
		return err
	}
	if cfg.APIExtKubeClient, err = kext_cs.NewForConfig(cfg.ClientConfig); err != nil {
		return err
	}
	if cfg.DBClient, err = cs.NewForConfig(cfg.ClientConfig); err != nil {
		return err
	}
	if cfg.PromClient, err = prom.NewForConfig(&s.PrometheusCrdKinds, s.PrometheusCrdGroup, cfg.ClientConfig); err != nil {
		return err
	}
	if cfg.DynamicClient, err = dynamic.NewForConfig(cfg.ClientConfig); err != nil {
		return err
	}
	cfg.KubeInformerFactory = informers.NewSharedInformerFactory(cfg.KubeClient, cfg.ResyncPeriod)
	cfg.KubedbInformerFactory = kubedbinformers.NewSharedInformerFactory(cfg.DBClient, cfg.ResyncPeriod)

	cfg.CronController = snapc.NewCronController(cfg.KubeClient, cfg.DBClient, cfg.DynamicClient)

	return nil
}
