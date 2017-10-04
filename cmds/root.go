package cmds

import (
	"flag"
	"log"
	"strings"

	stringz "github.com/appscode/go/strings"
	v "github.com/appscode/go/version"
	"github.com/jpillora/go-ogle-analytics"
	tcs "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	clientset "k8s.io/client-go/kubernetes"
)

const (
	gaTrackingCode = "UA-62096468-20"
)

var (
	masterURL         string
	kubeconfigPath    string
	governingService  string = "kubedb"
	exporterTag       string
	esOperatorTag     string = "0.7.1"
	elasticDumpTag    string = "2.4.2"
	address           string = ":8080"
	operatorNamespace string = namespace()
	enableRbac        bool   = false

	kubeClient clientset.Interface
	dbClient   tcs.KubedbV1alpha1Interface
)

func NewRootCmd(version string) *cobra.Command {
	enableAnalytics := true
	exporterTag = stringz.Val(version, "canary")

	var rootCmd = &cobra.Command{
		Use:               "operator [command]",
		Short:             `KubeDB operator by AppsCode`,
		DisableAutoGenTag: true,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			c.Flags().VisitAll(func(flag *pflag.Flag) {
				log.Printf("FLAG: --%s=%q", flag.Name, flag.Value)
			})
			if enableAnalytics && gaTrackingCode != "" {
				if client, err := ga.NewClient(gaTrackingCode); err == nil {
					parts := strings.Split(c.CommandPath(), " ")
					client.Send(ga.NewEvent("kubedb-operator", strings.Join(parts[1:], "/")).Label(version))
				}
			}
		},
	}
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	// ref: https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	flag.CommandLine.Parse([]string{})
	rootCmd.PersistentFlags().BoolVar(&enableAnalytics, "analytics", enableAnalytics, "Send analytical events to Google Analytics")

	rootCmd.AddCommand(NewCmdRun())
	rootCmd.AddCommand(NewCmdExport())
	rootCmd.AddCommand(v.NewCmdVersion())

	return rootCmd
}
