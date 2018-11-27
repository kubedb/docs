package cmds

import (
	"io"

	"github.com/appscode/go/log"
	"github.com/kubedb/mysql/pkg/cmds/server"
	"github.com/spf13/cobra"
)

func NewCmdRun(out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	o := server.NewMySQLServerOptions(out, errOut)

	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Launch MySQL server",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infoln("Starting mysql-server...")

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
