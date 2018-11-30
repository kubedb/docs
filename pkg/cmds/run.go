package cmds

import (
	"io"
	"strings"
	"time"

	"github.com/appscode/go/log"
	"github.com/jpillora/go-ogle-analytics"
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
					// ref: https://stackoverflow.com/a/17799161/4628962
					for {
						select {
						case <-ticker.C:
							if client, err := ga.NewClient(gaTrackingCode); err == nil {
								client.ClientID(controller.AnalyticsClientID)
								parts := strings.Split(cmd.CommandPath(), " ")
								client.Send(ga.NewEvent("kubedb-operator", strings.Join(parts[1:], "/")).Label(version))
							}
						case <-stopCh:
							return
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
