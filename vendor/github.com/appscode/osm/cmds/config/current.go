package config

import (
	"os"

	"github.com/appscode/go/term"
	otx "github.com/appscode/osm/context"
	"github.com/spf13/cobra"
)

func newCmdCurrent() *cobra.Command {
	setCmd := &cobra.Command{
		Use:               "current-context",
		Short:             "Print current context",
		Example:           "osm config current-context",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				cmd.Help()
				os.Exit(1)
			}
			currentContext(otx.GetConfigPath(cmd))
		},
	}
	return setCmd
}

func currentContext(configPath string) {
	config, err := otx.LoadConfig(configPath)
	term.ExitOnError(err)

	term.Infoln(config.CurrentContext)
}
