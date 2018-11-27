package cmds

import (
	le "github.com/kubedb/postgres/pkg/leader_election"
	"github.com/spf13/cobra"
)

func NewCmdLeaderElection() *cobra.Command {

	cmd := &cobra.Command{
		Use:               "leader_election",
		Short:             "Run leader election for postgres",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			le.RunLeaderElection()
		},
	}

	return cmd
}
