package cmds

import (
	"flag"
	"log"
	"os"

	"github.com/appscode/go/log/golog"
	v "github.com/appscode/go/version"
	"github.com/tekliner/apimachinery/client/clientset/versioned/scheme"
	"github.com/kubedb/operator/pkg/controller"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	genericapiserver "k8s.io/apiserver/pkg/server"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
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
			scheme.AddToScheme(clientsetscheme.Scheme)
			controller.LoggerOptions = golog.ParseFlags(c.Flags())
		},
	}
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	// ref: https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	flag.CommandLine.Parse([]string{})

	rootCmd.AddCommand(v.NewCmdVersion())

	stopCh := genericapiserver.SetupSignalHandler()
	rootCmd.AddCommand(NewCmdRun(os.Stdout, os.Stderr, stopCh))

	return rootCmd
}
