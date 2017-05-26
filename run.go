package main

import (
	"fmt"
	"time"

	"github.com/appscode/go/hold"
	"github.com/appscode/go/runtime"
	"github.com/appscode/go/strings"
	"github.com/appscode/go/version"
	"github.com/appscode/log"
	pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	tcs "github.com/k8sdb/apimachinery/client/clientset"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
	esCtrl "github.com/k8sdb/elasticsearch/pkg/controller"
	pgCtrl "github.com/k8sdb/postgres/pkg/controller"
	"github.com/spf13/cobra"
	cgcmd "k8s.io/client-go/tools/clientcmd"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
)

type Options struct {
	masterURL        string
	kubeconfigPath   string
	governingService string
	// For elasticsearch operator
	esOperatorTag  string
	elasticDumpTag string
	// For postgres operator
	postgresUtilTag string
}

func NewCmdRun() *cobra.Command {
	opt := Options{
		esOperatorTag:    strings.Val(version.Version.Version, "canary"),
		elasticDumpTag:   "canary",
		postgresUtilTag:  "canary-util",
		governingService: "kubedb",
	}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run kubedb operator in Kubernetes",
		Run: func(cmd *cobra.Command, args []string) {
			run(opt)
		},
	}

	cmd.Flags().StringVar(&opt.masterURL, "master", "", "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	cmd.Flags().StringVar(&opt.kubeconfigPath, "kubeconfig", "", "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	cmd.Flags().StringVar(&opt.esOperatorTag, "es.operator", opt.esOperatorTag, "Tag of elasticsearch opearator")
	cmd.Flags().StringVar(&opt.elasticDumpTag, "es.elasticdump", opt.elasticDumpTag, "Tag of elasticdump")
	cmd.Flags().StringVar(&opt.postgresUtilTag, "pg.postgres-util", opt.postgresUtilTag, "Tag of postgres util")
	cmd.Flags().StringVar(&opt.governingService, "governing-service", opt.governingService, "Governing service for database statefulset")

	return cmd
}

func run(opt Options) {
	// Check elasticsearch operator docker image tag
	if err := amc.CheckDockerImageVersion(esCtrl.ImageOperatorElasticsearch, opt.esOperatorTag); err != nil {
		log.Fatalf(`Image %v:%v not found.`, esCtrl.ImageOperatorElasticsearch, opt.esOperatorTag)
	}

	// Check elasticdump docker image tag
	if err := amc.CheckDockerImageVersion(esCtrl.ImageElasticDump, opt.elasticDumpTag); err != nil {
		log.Fatalf(`Image %v:%v not found.`, esCtrl.ImageElasticDump, opt.elasticDumpTag)
	}

	// Check postgres docker image tag
	if err := amc.CheckDockerImageVersion(pgCtrl.ImagePostgres, opt.postgresUtilTag); err != nil {
		log.Fatalf(`Image %v:%v not found.`, pgCtrl.ImagePostgres, opt.postgresUtilTag)
	}

	config, err := clientcmd.BuildConfigFromFlags(opt.masterURL, opt.kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get kubernetes config: %s", err)
	}

	client := clientset.NewForConfigOrDie(config)
	extClient := tcs.NewExtensionsForConfigOrDie(config)

	cgConfig, err := cgcmd.BuildConfigFromFlags(opt.masterURL, opt.kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get kubernetes config: %s", err)
	}

	promClient, err := pcm.NewForConfig(cgConfig)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Starting operator...")

	defer runtime.HandleCrash()

	go pgCtrl.New(client, extClient, promClient, opt.postgresUtilTag, opt.governingService).RunAndHold()
	// Need to wait for sometime to run another controller.
	// Or multiple controller will try to create common TPR simultaneously which gives error
	time.Sleep(time.Second * 10)
	go esCtrl.New(client, extClient, promClient, opt.esOperatorTag, opt.elasticDumpTag, opt.governingService).RunAndHold()

	hold.Hold()
}
