package cmds

import (
	"flag"
	"log"
	"strings"

	stringz "github.com/appscode/go/strings"
	v "github.com/appscode/go/version"
	"github.com/appscode/kutil/meta"
	"github.com/appscode/kutil/tools/analytics"
	"github.com/jpillora/go-ogle-analytics"
	"github.com/kubedb/apimachinery/client/scheme"
	tcs "github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
)

const (
	gaTrackingCode = "UA-62096468-20"
)

var (
	masterURL         string
	kubeconfigPath    string
	governingService  string = "kubedb"
	exporterTag       string
	registry          string = "kubedb"
	address           string = ":8080"
	operatorNamespace string = meta.Namespace()
	enableRbac        bool   = false
	analyticsClientID string = analytics.ClientID()

	kubeClient kubernetes.Interface
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
					client.ClientID(analyticsClientID)
					parts := strings.Split(c.CommandPath(), " ")
					client.Send(ga.NewEvent("kubedb-operator", strings.Join(parts[1:], "/")).Label(version))
				}
			}
			scheme.AddToScheme(clientsetscheme.Scheme)
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
