package cmds

import (
	"log"

	eh "github.com/kubedb/etcd/pkg/etcd-helper"
	"github.com/kubedb/etcd/pkg/etcdmain"
	"github.com/spf13/cobra"
)

func NewCmdEtcdHelper() *cobra.Command {

	etcdConf := etcdmain.NewConfig()
	cmd := &cobra.Command{
		Use:               "etcd-helper",
		Short:             "Run etcd helper",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := etcdConf.ConfigFromCmdLine(); err != nil {
				log.Fatalln(err)
			}
			eh.RunEtcdHelper(etcdConf)
		},
	}
	cmd.Flags().AddGoFlagSet(etcdConf.FlagSet)

	return cmd
}
