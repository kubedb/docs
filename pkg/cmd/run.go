package cmd

import (
	"fmt"

	"github.com/appscode/go/hold"
	"github.com/appscode/go/runtime"
	"github.com/appscode/go/version"
	esCtrl "github.com/k8sdb/elasticsearch/pkg/controller"
	pgCtrl "github.com/k8sdb/postgres/pkg/controller"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"time"
)

const (
	// Default tag
	canary     = "canary"
	canaryUtil = "canary-util"
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
	opt := &Options{}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run kubedb operator in Kubernetes",
		Run: func(cmd *cobra.Command, args []string) {
			run(opt)
		},
	}

	operatorVersion := version.Version.Version
	if operatorVersion == "" {
		operatorVersion = canary
	}

	cmd.Flags().StringVar(&opt.masterURL, "master", "", "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	cmd.Flags().StringVar(&opt.kubeconfigPath, "kubeconfig", "", "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	cmd.Flags().StringVar(&opt.esOperatorTag, "es.operator", operatorVersion, "Tag of elasticsearch opearator")
	cmd.Flags().StringVar(&opt.elasticDumpTag, "es.elasticdump", canary, "Tag of elasticdump")
	cmd.Flags().StringVar(&opt.postgresUtilTag, "pg.postgres-util", canaryUtil, "Tag of postgres util")
	cmd.Flags().StringVar(&opt.governingService, "governing-service", "k8sdb", "Governing service for database statefulset")

	return cmd
}

func run(opt *Options) {
	config, err := clientcmd.BuildConfigFromFlags(opt.masterURL, opt.kubeconfigPath)
	if err != nil {
		fmt.Printf("Could not get kubernetes config: %s", err)
		panic(err)
	}
	defer runtime.HandleCrash()

	fmt.Println("Starting operator...")

	go pgCtrl.New(config, opt.postgresUtilTag, opt.governingService).RunAndHold()
	// Need to wait for sometime to run another controller.
	// Or multiple controller will try to create common TPR simultaneously which gives error
	time.Sleep(time.Second * 30)
	go esCtrl.New(config, opt.esOperatorTag, opt.elasticDumpTag, opt.governingService).RunAndHold()

	hold.Hold()
}
