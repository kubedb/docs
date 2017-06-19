package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/appscode/go/runtime"
	"github.com/appscode/log"
	"github.com/appscode/pat"
	pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	tcs "github.com/k8sdb/apimachinery/client/clientset"
	"github.com/k8sdb/apimachinery/pkg/analytics"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
	"github.com/k8sdb/apimachinery/pkg/docker"
	esCtrl "github.com/k8sdb/elasticsearch/pkg/controller"
	pgCtrl "github.com/k8sdb/postgres/pkg/controller"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	clientset "k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
	cgcmd "k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL         string
	kubeconfigPath    string
	governingService  string = "kubedb"
	esOperatorTag     string = "0.1.0"
	elasticDumpTag    string = "2.4.2"
	address           string = ":8080"
	operatorNamespace string = namespace()
	enableAnalytics   bool   = true

	kubeClient clientset.Interface
	dbClient   tcs.ExtensionInterface
)

func NewCmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run kubedb operator in Kubernetes",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			analytics.SendEvent("operator", "stopped", Version)
		},
	}

	// operator flags
	cmd.Flags().StringVar(&masterURL, "master", masterURL, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	cmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", kubeconfigPath, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	cmd.Flags().StringVar(&governingService, "governing-service", governingService, "Governing service for database statefulset")
	cmd.Flags().StringVar(&address, "address", address, "Address to listen on for web interface and telemetry.")

	// elasticsearch flags
	cmd.Flags().StringVar(&esOperatorTag, "elasticsearch.operator-tag", esOperatorTag, "Tag of kubedb/es-operator used for discovery")

	// elasticdump flags
	cmd.Flags().StringVar(&elasticDumpTag, "elasticdump.tag", elasticDumpTag, "Tag of elasticdump")

	// Analytics flags
	cmd.Flags().BoolVar(&enableAnalytics, "analytics", enableAnalytics, "Send analytical event to Google Analytics")

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

	kubeClient := clientset.NewForConfigOrDie(config)
	dbClient := tcs.NewForConfigOrDie(config)

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

	if enableAnalytics {
		analytics.Enable()
	}
	analytics.SendEvent("operator", "started", Version)

	defer runtime.HandleCrash()

	pgCtrl.New(kubeClient, dbClient, promClient, cronController, pgCtrl.Options{
		GoverningService:  governingService,
		OperatorNamespace: operatorNamespace,
		OperatorTag:       esOperatorTag,
		EnableAnalytics:   enableAnalytics,
	}).Run()

	// Need to wait for sometime to run another controller.
	// Or multiple controller will try to create common TPR simultaneously which gives error
	time.Sleep(time.Second * 10)
	esCtrl.New(kubeClient, dbClient, promClient, cronController, esCtrl.Options{
		GoverningService:  governingService,
		ElasticDumpTag:    elasticDumpTag,
		OperatorTag:       esOperatorTag,
		OperatorNamespace: operatorNamespace,
		EnableAnalytics:   enableAnalytics,
	}).Run()

	m := pat.New()
	m.Get("/metrics", promhttp.Handler())
	pattern := fmt.Sprintf("/kubedb.com/v1alpha1/namespaces/%s/%s/%s/metrics", PathParamNamespace, PathParamType, PathParamName)
	log.Infoln("URL pattern:", pattern)
	m.Get(pattern, http.HandlerFunc(ExportMetrics))
	m.Del(pattern, http.HandlerFunc(DeleteRegistry))
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
	return apiv1.NamespaceDefault
}
