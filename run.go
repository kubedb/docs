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
	"github.com/k8sdb/apimachinery/pkg/docker"
	esCtrl "github.com/k8sdb/elasticsearch/pkg/controller"
	pgCtrl "github.com/k8sdb/postgres/pkg/controller"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	cgcmd "k8s.io/client-go/tools/clientcmd"
	kapi "k8s.io/kubernetes/pkg/api"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
)

var (
	masterURL         string
	kubeconfigPath    string
	governingService  string = "kubedb"
	esOperatorTag     string = "0.1.0"
	elasticDumpTag    string = "2.4.2"
	address           string = ":8080"
	exporterNamespace string = namespace()
	exporterTag       string = "0.1.0"
	enableAnalytics   bool   = true
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
	cmd.Flags().StringVar(&address, "address", address, "Address to listen on for web interface and telemetry.")

	// elasticsearch flags
	cmd.Flags().StringVar(&esOperatorTag, "elasticsearch.operator-tag", esOperatorTag, "Tag of kubedb/es-operator used for discovery")

	// elasticdump flags
	cmd.Flags().StringVar(&elasticDumpTag, "elasticdump.tag", elasticDumpTag, "Tag of elasticdump")

	// exporter tags
	cmd.Flags().StringVar(&exporterNamespace, "exporter.namespace", exporterNamespace, "Namespace for monitoring exporter")
	cmd.Flags().StringVar(&exporterTag, "exporter.tag", exporterTag, "Tag of monitoring exporter")

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

	client := clientset.NewForConfigOrDie(config)
	extClient := tcs.NewForConfigOrDie(config)

	cgConfig, err := cgcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	promClient, err := pcm.NewForConfig(cgConfig)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Starting operator...")

	if enableAnalytics {
		analytics.Enable()
	}
	analytics.SendEvent(docker.ImageOperator, "started", Version)

	defer runtime.HandleCrash()

	pgCtrl.New(client, extClient, promClient, pgCtrl.Options{
		GoverningService:  governingService,
		ExporterNamespace: exporterNamespace,
		ExporterTag:       exporterTag,
		EnableAnalytics:   enableAnalytics,
	}).Run()

	// Need to wait for sometime to run another controller.
	// Or multiple controller will try to create common TPR simultaneously which gives error
	time.Sleep(time.Second * 10)
	esCtrl.New(client, extClient, promClient, esCtrl.Options{
		GoverningService:  governingService,
		ElasticDumpTag:    elasticDumpTag,
		OperatorTag:       esOperatorTag,
		ExporterNamespace: exporterNamespace,
		ExporterTag:       exporterTag,
		EnableAnalytics:   enableAnalytics,
	}).Run()

	m := pat.New()
	m.Get("/metrics", promhttp.Handler())
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
	return kapi.NamespaceDefault
}
