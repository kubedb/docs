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

package server

import (
	"flag"
	"fmt"
	"time"

	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	kubedbinformers "kubedb.dev/apimachinery/client/informers/externalversions"
	"kubedb.dev/apimachinery/pkg/controller/initializer/stash"
	sts "kubedb.dev/apimachinery/pkg/controller/statefulset"
	"kubedb.dev/apimachinery/pkg/eventer"
	"kubedb.dev/operator/pkg/controller"

	prom "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"github.com/spf13/pflag"
	licenseapi "go.bytebuilders.dev/license-verifier/apis/licenses/v1alpha1"
	license "go.bytebuilders.dev/license-verifier/kubernetes"
	core "k8s.io/api/core/v1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/tools/clientcmd"
	"kmodules.xyz/client-go/tools/queue"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
	appcatinformers "kmodules.xyz/custom-resources/client/informers/externalversions"
)

type OperatorOptions struct {
	MasterURL              string
	KubeconfigPath         string
	LicenseFile            string
	QPS                    float64
	Burst                  int
	ResyncPeriod           time.Duration
	ReadinessProbeInterval time.Duration
	MaxNumRequeues         int
	NumThreads             int
}

func NewOperatorOptions() *OperatorOptions {
	return &OperatorOptions{
		ResyncPeriod:           10 * time.Minute,
		ReadinessProbeInterval: 10 * time.Second,
		MaxNumRequeues:         5,
		NumThreads:             2,
		// ref: https://github.com/kubernetes/ingress-nginx/blob/e4d53786e771cc6bdd55f180674b79f5b692e552/pkg/ingress/controller/launch.go#L252-L259
		// High enough QPS to fit all expected use cases. QPS=0 is not set here, because client code is overriding it.
		QPS: 1e6,
		// High enough Burst to fit all expected use cases. Burst=0 is not set here, because client code is overriding it.
		Burst: 1e6,
	}
}

func (s *OperatorOptions) AddGoFlags(fs *flag.FlagSet) {
	fs.StringVar(&s.MasterURL, "master", s.MasterURL, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	fs.StringVar(&s.KubeconfigPath, "kubeconfig", s.KubeconfigPath, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")

	fs.StringVar(&s.LicenseFile, "license-file", s.LicenseFile, "Path to license file")

	fs.Float64Var(&s.QPS, "qps", s.QPS, "The maximum QPS to the master from this client")
	fs.IntVar(&s.Burst, "burst", s.Burst, "The maximum burst for throttle")
	fs.DurationVar(&s.ResyncPeriod, "resync-period", s.ResyncPeriod, "If non-zero, will re-list this often. Otherwise, re-list will be delayed aslong as possible (until the upstream source closes the watch or times out.")
	fs.DurationVar(&s.ReadinessProbeInterval, "readiness-probe-interval", s.ReadinessProbeInterval, "The time between two consecutive health checks that the operator performs to the database.")
}

func (s *OperatorOptions) AddFlags(fs *pflag.FlagSet) {
	pfs := flag.NewFlagSet("extra-flags", flag.ExitOnError)
	s.AddGoFlags(pfs)
	fs.AddGoFlagSet(pfs)
}

func (s *OperatorOptions) ApplyTo(cfg *controller.OperatorConfig) error {
	var err error

	cfg.LicenseFile = s.LicenseFile

	cfg.ClientConfig.QPS = float32(s.QPS)
	cfg.ClientConfig.Burst = s.Burst
	cfg.ResyncPeriod = s.ResyncPeriod
	cfg.ReadinessProbeInterval = s.ReadinessProbeInterval
	cfg.MaxNumRequeues = s.MaxNumRequeues
	cfg.NumThreads = s.NumThreads

	cfg.RestrictToNamespace = queue.NamespaceDemo
	if cfg.LicenseFile != "" {
		info := license.NewLicenseEnforcer(cfg.ClientConfig, cfg.LicenseFile).LoadLicense()
		if info.Status != licenseapi.LicenseActive {
			return fmt.Errorf("license status %s, reason: %s", info.Status, info.Reason)
		}
		if sets.NewString(info.Features...).Has("kubedb-enterprise") {
			cfg.RestrictToNamespace = core.NamespaceAll
		} else if !sets.NewString(info.Features...).Has("kubedb-community") {
			return fmt.Errorf("not a valid license for this product")
		}
		cfg.License = info
	}

	if cfg.KubeClient, err = kubernetes.NewForConfig(cfg.ClientConfig); err != nil {
		return err
	}
	if cfg.CRDClient, err = crd_cs.NewForConfig(cfg.ClientConfig); err != nil {
		return err
	}
	if cfg.DBClient, err = cs.NewForConfig(cfg.ClientConfig); err != nil {
		return err
	}
	if cfg.DynamicClient, err = dynamic.NewForConfig(cfg.ClientConfig); err != nil {
		return err
	}
	if cfg.AppCatalogClient, err = appcat_cs.NewForConfig(cfg.ClientConfig); err != nil {
		return err
	}
	if cfg.PromClient, err = prom.NewForConfig(cfg.ClientConfig); err != nil {
		return err
	}
	cfg.KubeInformerFactory = informers.NewSharedInformerFactory(cfg.KubeClient, cfg.ResyncPeriod)
	cfg.KubedbInformerFactory = kubedbinformers.NewSharedInformerFactory(cfg.DBClient, cfg.ResyncPeriod)
	cfg.AppCatInformerFactory = appcatinformers.NewSharedInformerFactory(cfg.AppCatalogClient, cfg.ResyncPeriod)

	cfg.SecretInformer = cfg.KubeInformerFactory.InformerFor(&core.Secret{}, func(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
		return coreinformers.NewSecretInformer(
			client,
			cfg.RestrictToNamespace,
			resyncPeriod,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
		)
	})
	cfg.SecretLister = corelisters.NewSecretLister(cfg.SecretInformer.GetIndexer())
	// Create event recorder
	cfg.Recorder = eventer.NewEventRecorder(cfg.KubeClient, "KubeDB Operator")
	// Initialize StatefulSet watcher
	sts.NewController(&cfg.Config, cfg.KubeClient, cfg.DBClient, cfg.DynamicClient).InitStsWatcher()
	// Configure Stash initializer
	return stash.Configure(cfg.ClientConfig, &cfg.Initializers.Stash, cfg.ResyncPeriod)
}

func (s *OperatorOptions) Validate() []error {
	return nil
}

func (s *OperatorOptions) Complete() error {
	return nil
}

func (s OperatorOptions) Config() (*controller.OperatorConfig, error) {
	clientConfig, err := clientcmd.BuildConfigFromFlags(s.MasterURL, s.KubeconfigPath)
	if err != nil {
		return nil, err
	}

	// Fixes https://github.com/Azure/AKS/issues/522
	clientcmd.Fix(clientConfig)

	cfg := controller.NewOperatorConfig(clientConfig)
	if err := s.ApplyTo(cfg); err != nil {
		return nil, err
	}
	if cfg.RestrictToNamespace != core.NamespaceAll {
		klog.Infof("Operator restricted to %s namespace", cfg.RestrictToNamespace)
	}

	return cfg, nil
}

func (s OperatorOptions) Run(stopCh <-chan struct{}) error {
	cfg, err := s.Config()
	if err != nil {
		return err
	}

	ctrl, err := cfg.New()
	if err != nil {
		return err
	}

	// Start periodic license verification
	//nolint:errcheck
	go license.VerifyLicensePeriodically(cfg.ClientConfig, s.LicenseFile, stopCh)

	ctrl.Run(stopCh)
	return nil
}
