package cmds

import (
	"flag"
	"log"
	"os"

	log2 "github.com/appscode/go/log"
	"github.com/appscode/go/log/golog"
	v "github.com/appscode/go/version"
	"github.com/kubedb/apimachinery/client/clientset/versioned/scheme"
	"github.com/kubedb/operator/pkg/controller"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	genericapiserver "k8s.io/apiserver/pkg/server"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	appcatscheme "kmodules.xyz/custom-resources/client/clientset/versioned/scheme"
)

const (
	gaTrackingCode = "UA-62096468-20"
)

func NewRootCmd(version string) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:               "operator [command]",
		Short:             `KubeDB operator by AppsCode`,
		DisableAutoGenTag: true,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			c.Flags().VisitAll(func(flag *pflag.Flag) {
				log.Printf("FLAG: --%s=%q", flag.Name, flag.Value)
			})
			if controller.EnableAnalytics && gaTrackingCode != "" {
				if err := sendAnalytics(c.CommandPath(), version); err != nil {
					log2.Error(err)
				}
			}
			scheme.AddToScheme(clientsetscheme.Scheme)
			appcatscheme.AddToScheme(clientsetscheme.Scheme)
			controller.LoggerOptions = golog.ParseFlags(c.Flags())
		},
	}
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	// ref: https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	flag.CommandLine.Parse([]string{})
	rootCmd.PersistentFlags().BoolVar(&controller.EnableAnalytics, "enable-analytics", controller.EnableAnalytics, "Send analytical events to Google Analytics")

	rootCmd.AddCommand(v.NewCmdVersion())

	stopCh := genericapiserver.SetupSignalHandler()
	rootCmd.AddCommand(NewCmdRun(os.Stdout, os.Stderr, version, stopCh))

	return rootCmd
}
