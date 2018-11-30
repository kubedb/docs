package cmds

import (
	"io"
	"strings"
	"time"

	"github.com/jpillora/go-ogle-analytics"

	"github.com/appscode/go/log"
	"github.com/kubedb/operator/pkg/cmds/server"
	"github.com/kubedb/operator/pkg/controller"
	"github.com/spf13/cobra"
)

func NewCmdRun(out, errOut io.Writer, version string, stopCh <-chan struct{}) *cobra.Command {
	o := server.NewKubeDBServerOptions(out, errOut)

	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Run kubedb operator in Kubernetes",
		DisableAutoGenTag: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			if controller.EnableAnalytics && gaTrackingCode != "" {
				ticker := time.NewTicker(24 * time.Hour)
				go func() {
					for range ticker.C {
						if err := sendAnalytics(cmd.CommandPath(), version); err != nil {
							log.Error(err)
						}
					}
				}()
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infoln("Starting kubedb-server...")

			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.Run(stopCh); err != nil {
				return err
			}
			return nil
		},
	}

	o.AddFlags(cmd.Flags())

	return cmd
}

func sendAnalytics(commandPath, version string) error {
	client, err := ga.NewClient(gaTrackingCode)
	if err != nil {
		return err
	}
	client.ClientID(controller.AnalyticsClientID)
	parts := strings.Split(commandPath, " ")
	return client.Send(ga.NewEvent("kubedb-operator", strings.Join(parts[1:], "/")).Label(version))
}
