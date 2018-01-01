package cmds

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/appscode/go/log"
	"github.com/appscode/go/runtime"
	apiext_util "github.com/appscode/kutil/apiextensions/v1beta1"
	"github.com/appscode/pat"
	pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	tcs "github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/apimachinery/pkg/docker"
	"github.com/kubedb/apimachinery/pkg/migrator"
	esCtrl "github.com/kubedb/elasticsearch/pkg/controller"
	memCtrl "github.com/kubedb/memcached/pkg/controller"
	mgoCtrl "github.com/kubedb/mongodb/pkg/controller"
	msCtrl "github.com/kubedb/mysql/pkg/controller"
	pgCtrl "github.com/kubedb/postgres/pkg/controller"
	rdCtrl "github.com/kubedb/redis/pkg/controller"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	ecs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
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
	// elasticdump flags
	cmd.Flags().StringVar(&elasticDumpTag, "elasticdump.tag", elasticDumpTag, "Tag of elasticdump")

	return cmd
}

func run() {
	// Check elasticdump docker image tag
	if err := docker.CheckDockerImageVersion(docker.ImageElasticdump, elasticDumpTag); err != nil {
		log.Fatalf(`Image %v:%v not found.`, docker.ImageElasticdump, elasticDumpTag)
	}

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	kubeClient = kubernetes.NewForConfigOrDie(config)
	apiExtKubeClient := ecs.NewForConfigOrDie(config)
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
		&api.Postgres{},
		&api.Elasticsearch{},
		&api.Snapshot{},
		&api.DormantDatabase{},
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer runtime.HandleCrash()

	// Register CRDs
	err = setup(apiExtKubeClient)
	if err != nil {
		log.Fatalln(err)
	}

	// Postgres controller
	pgCtrl.New(kubeClient, apiExtKubeClient, dbClient, promClient, cronController, pgCtrl.Options{
		GoverningService:  governingService,
		OperatorNamespace: operatorNamespace,
		ExporterTag:       exporterTag,
		EnableRbac:        enableRbac,
	}).Run()

	// Elasticsearch controller
	esCtrl.New(kubeClient, apiExtKubeClient, dbClient, promClient, cronController, esCtrl.Options{
		GoverningService:  governingService,
		ExporterTag:       exporterTag,
		ElasticDumpTag:    elasticDumpTag,
		OperatorNamespace: operatorNamespace,
		EnableRbac:        enableRbac,
	}).Run()

	// MySQL controller
	msCtrl.New(kubeClient, apiExtKubeClient, dbClient, promClient, cronController, msCtrl.Options{
		GoverningService:  governingService,
		OperatorNamespace: operatorNamespace,
		ExporterTag:       exporterTag,
		EnableRbac:        enableRbac,
	}).Run()

	// MongoDB controller
	mgoCtrl.New(kubeClient, apiExtKubeClient, dbClient, promClient, cronController, mgoCtrl.Options{
		GoverningService:  governingService,
		OperatorNamespace: operatorNamespace,
		ExporterTag:       exporterTag,
		EnableRbac:        enableRbac,
	}).Run()

	// Redis controller
	rdCtrl.New(kubeClient, apiExtKubeClient, dbClient, promClient, rdCtrl.Options{
		GoverningService:  governingService,
		OperatorNamespace: operatorNamespace,
		ExporterTag:       exporterTag,
		EnableRbac:        enableRbac,
	}).Run()

	// Memcached controller
	memCtrl.New(kubeClient, apiExtKubeClient, dbClient, promClient, memCtrl.Options{
		GoverningService:  governingService,
		OperatorNamespace: operatorNamespace,
		ExporterTag:       exporterTag,
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

// Ensure Custom Resource definitions
func setup(client ecs.ApiextensionsV1beta1Interface) error {
	log.Infoln("Ensuring CustomResourceDefinition...")
	crds := []*crd_api.CustomResourceDefinition{
		api.Elasticsearch{}.CustomResourceDefinition(),
		api.Postgres{}.CustomResourceDefinition(),
		api.MySQL{}.CustomResourceDefinition(),
		api.MongoDB{}.CustomResourceDefinition(),
		api.Redis{}.CustomResourceDefinition(),
		api.Memcached{}.CustomResourceDefinition(),
		api.DormantDatabase{}.CustomResourceDefinition(),
		api.Snapshot{}.CustomResourceDefinition(),
	}
	return apiext_util.RegisterCRDs(client, crds)
}
