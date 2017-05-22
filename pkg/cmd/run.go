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

type esOptions struct {
	operatorTag    string
	elasticDumpTag string
}

type pgOptions struct {
	postgresUtilTag string
}

type commonOptions struct {
	masterURL        string
	kubeconfigPath   string
	governingService string
}

func NewCmdRun() *cobra.Command {

	es := &esOptions{}
	pg := &pgOptions{}
	common := &commonOptions{}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run kubedb operator in Kubernetes",
		Run: func(cmd *cobra.Command, args []string) {
			run(common, es, pg)
		},
	}

	operatorVersion := version.Version.Version
	if operatorVersion == "" {
		operatorVersion = canary
	}

	cmd.Flags().StringVar(&common.masterURL, "master", "", "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	cmd.Flags().StringVar(&common.kubeconfigPath, "kubeconfig", "", "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	cmd.Flags().StringVar(&es.operatorTag, "es.operator", operatorVersion, "Tag of elasticsearch opearator")
	cmd.Flags().StringVar(&es.elasticDumpTag, "es.elasticdump", canary, "Tag of elasticdump")
	cmd.Flags().StringVar(&pg.postgresUtilTag, "pg.postgres-util", canaryUtil, "Tag of postgres util")
	cmd.Flags().StringVar(&common.governingService, "governing-service", "k8sdb", "Governing service for database statefulset")

	return cmd
}

func run(common *commonOptions, es *esOptions, pg *pgOptions) {
	config, err := clientcmd.BuildConfigFromFlags(common.masterURL, common.kubeconfigPath)
	if err != nil {
		fmt.Printf("Could not get kubernetes config: %s", err)
		panic(err)
	}
	defer runtime.HandleCrash()

	fmt.Println("Starting operator...")

	go pgCtrl.New(config, pg.postgresUtilTag, common.governingService).RunAndHold()
	// Need to wait for sometime to run another controller.
	// Or multiple controller will try to create common TPR simultaneously which gives error
	time.Sleep(time.Second * 30)
	go esCtrl.New(config, es.operatorTag, es.elasticDumpTag, common.governingService).RunAndHold()

	hold.Hold()
}
