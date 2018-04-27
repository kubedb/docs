package cmds

import (
	"github.com/kubedb/operator/pkg/exporter"
	"github.com/spf13/cobra"
)

func NewCmdExport() *cobra.Command {

	opt := exporter.Options{
		Address: ":8080",
	}

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export Prometheus metrics for HAProxy",
		Run: func(cmd *cobra.Command, args []string) {
			opt.Export()
		},
	}

	cmd.Flags().StringVar(&opt.Address, "address", opt.Address, "Address to listen on for web interface and telemetry.")
	cmd.Flags().StringVar(&opt.MasterURL, "master", opt.MasterURL, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	cmd.Flags().StringVar(&opt.KubeconfigPath, "kubeconfig", opt.KubeconfigPath, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")

	return cmd
}
