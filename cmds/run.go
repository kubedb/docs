package cmds

import (
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/go/runtime"
	"github.com/appscode/pat"
	pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	tcs "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
	"github.com/k8sdb/apimachinery/pkg/docker"
	"github.com/k8sdb/apimachinery/pkg/migrator"
	esCtrl "github.com/k8sdb/elasticsearch/pkg/controller"
	pgCtrl "github.com/k8sdb/postgres/pkg/controller"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	cgcmd "k8s.io/client-go/tools/clientcmd"
)

func NewCmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run kubedb operator in Kubernetes",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}

	// operator flags
	cmd.Flags().StringVar(&masterURL, "master", masterURL, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	cmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", kubeconfigPath, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	cmd.Flags().StringVar(&governingService, "governing-service", governingService, "Governing service for database statefulset")
	cmd.Flags().StringVar(&exporterTag, "exporter-tag", exporterTag, "Tag of kubedb/operator used as exporter")
	cmd.Flags().StringVar(&address, "address", address, "Address to listen on for web interface and telemetry.")
	cmd.Flags().BoolVar(&enableRbac, "rbac", enableRbac, "Enable RBAC for database workloads")
	// elasticsearch flags
	cmd.Flags().StringVar(&esOperatorTag, "elasticsearch.operator-tag", esOperatorTag, "Tag of kubedb/es-operator used for discovery")

	// elasticdump flags
	cmd.Flags().StringVar(&elasticDumpTag, "elasticdump.tag", elasticDumpTag, "Tag of elasticdump")

	return cmd
}

func run() {
	// Check elasticsearch operator docker image tag
	if err := docker.CheckDockerImageVersion(docker.ImageElasticOperator, esOperatorTag); err != nil {
		log.Fatalf(`Image %v:%v not found.`, docker.ImageElasticOperator, esOperatorTag)
	}

	// Check elasticdump docker image tag
	if err := docker.CheckDockerImageVersion(docker.ImageElasticdump, elasticDumpTag); err != nil {
		log.Fatalf(`Image %v:%v not found.`, docker.ImageElasticdump, elasticDumpTag)
	}

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	kubeClient = clientset.NewForConfigOrDie(config)
	apiExtKubeClient := apiextensionsclient.NewForConfigOrDie(config)
	dbClient = tcs.NewForConfigOrDie(config)

	cgConfig, err := cgcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	promClient, err := pcm.NewForConfig(cgConfig)
	if err != nil {
		log.Fatalln(err)
	}

	cronController := amc.NewCronController(kubeClient, dbClient)
	// Start Cron
	cronController.StartCron()
	// Stop Cron
	defer cronController.StopCron()

	fmt.Println("Starting operator...")

	tprMigrator := migrator.NewMigrator(kubeClient, apiExtKubeClient, dbClient)
	err = tprMigrator.RunMigration(
		&tapi.Postgres{},
		&tapi.Elasticsearch{},
		&tapi.Snapshot{},
		&tapi.DormantDatabase{},
	)
	if err != nil {
		log.Fatalln(err)
	}

	defer runtime.HandleCrash()

	pgCtrl.New(kubeClient, apiExtKubeClient, dbClient, promClient, cronController, pgCtrl.Options{
		GoverningService:  governingService,
		OperatorNamespace: operatorNamespace,
		ExporterTag:       exporterTag,
		EnableRbac:        enableRbac,
	}).Run()

	// Need to wait for sometime to run another controller.
	// Or multiple controller will try to create common TPR simultaneously which gives error
	time.Sleep(time.Second * 10)
	esCtrl.New(kubeClient, apiExtKubeClient, dbClient, promClient, cronController, esCtrl.Options{
		GoverningService:  governingService,
		ExporterTag:       exporterTag,
		ElasticDumpTag:    elasticDumpTag,
		DiscoveryTag:      esOperatorTag,
		OperatorNamespace: operatorNamespace,
		EnableRbac:        enableRbac,
	}).Run()

	m := pat.New()
	// For go metrics
	m.Get("/metrics", promhttp.Handler())
	metricsPattern := fmt.Sprintf("/kubedb.com/v1alpha1/namespaces/%s/%s/%s/metrics", PathParamNamespace, PathParamType, PathParamName)
	log.Infoln("Metrics URL pattern:", metricsPattern)
	m.Get(metricsPattern, http.HandlerFunc(ExportMetrics))
	m.Del(metricsPattern, http.HandlerFunc(DeleteRegistry))

	// For database summary report
	auditPattern := fmt.Sprintf("/kubedb.com/v1alpha1/namespaces/%s/%s/%s/report", PathParamNamespace, PathParamType, PathParamName)
	log.Infoln("Report URL pattern:", auditPattern)
	m.Get(auditPattern, http.HandlerFunc(ExportSummaryReport))

	http.Handle("/", m)
	log.Infof("Starting Server: %s", address)
	log.Fatal(http.ListenAndServe(address, nil))
}

func namespace() string {
	if ns := os.Getenv("OPERATOR_NAMESPACE"); ns != "" {
		return ns
	}
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}
	return core.NamespaceDefault
}
