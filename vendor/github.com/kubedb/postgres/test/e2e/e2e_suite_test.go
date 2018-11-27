package e2e_test

import (
	"flag"
	"path/filepath"
	"testing"
	"time"

	"github.com/appscode/go/log"
	logs "github.com/appscode/go/log/golog"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned"
	"github.com/kubedb/apimachinery/client/clientset/versioned/scheme"
	"github.com/kubedb/postgres/pkg/controller"
	"github.com/kubedb/postgres/test/e2e/framework"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/kubernetes"
	clientSetScheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	ka "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
)

var (
	storageClass string
)

func init() {
	scheme.AddToScheme(clientSetScheme.Scheme)

	flag.StringVar(&storageClass, "storageclass", "standard", "Kubernetes StorageClass name")
	flag.StringVar(&framework.DockerRegistry, "docker-registry", "kubedbci", "User provided docker repository")
	flag.StringVar(&framework.DBVersion, "pg-version", "10.2-v1", "Postgres version")
	flag.StringVar(&framework.ExporterTag, "exporter-tag", "v0.4.6", "Tag of postgres_exporter image")
	flag.BoolVar(&framework.EnableRbac, "rbac", true, "Enable RBAC for database workloads")
	flag.BoolVar(&framework.SelfHostedOperator, "selfhosted-operator", false, "Enable this for self-hosted operator")
}

const (
	TIMEOUT = 20 * time.Minute
)

var (
	ctrl *controller.Controller
	root *framework.Framework
)

func TestE2e(t *testing.T) {
	logs.InitLogs()
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(TIMEOUT)

	junitReporter := reporters.NewJUnitReporter("junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "e2e Suite", []Reporter{junitReporter})
}

var _ = BeforeSuite(func() {
	// Kubernetes config
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube/config")
	By("Using kubeconfig from " + kubeconfigPath)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	Expect(err).NotTo(HaveOccurred())
	// raise throttling time. ref: https://github.com/appscode/voyager/issues/640
	config.Burst = 100
	config.QPS = 100

	// Clients
	kubeClient := kubernetes.NewForConfigOrDie(config)
	extClient := cs.NewForConfigOrDie(config)
	kaClient := ka.NewForConfigOrDie(config)
	if err != nil {
		log.Fatalln(err)
	}
	// Framework
	root = framework.New(config, kubeClient, extClient, kaClient, storageClass)

	By("Using namespace " + root.Namespace())

	// Create namespace
	err = root.CreateNamespace()
	Expect(err).NotTo(HaveOccurred())

	if !framework.SelfHostedOperator {
		stopCh := genericapiserver.SetupSignalHandler()
		go root.RunOperatorAndServer(config, kubeconfigPath, stopCh)
	}

	root.EventuallyCRD().Should(Succeed())
	root.EventuallyAPIServiceReady().Should(Succeed())
})

var _ = AfterSuite(func() {
	if !framework.SelfHostedOperator {
		By("Delete Admission Controller Configs")
		root.CleanAdmissionConfigs()
	}
	By("Delete left over Postgres objects")
	root.CleanPostgres()
	By("Delete left over Dormant Database objects")
	root.CleanDormantDatabase()
	By("Delete left over Snapshot objects")
	root.CleanSnapshot()
	By("Delete Namespace")
	err := root.DeleteNamespace()
	Expect(err).NotTo(HaveOccurred())
})
