package config

import (
	"os"

	"github.com/appscode/go/term"
	otx "github.com/appscode/osm/context"
	"github.com/spf13/cobra"
)

func newCmdUse() *cobra.Command {
	setCmd := &cobra.Command{
		Use:               "use-context <name>",
		Short:             "Use context",
		Example:           "osm config use-context <name>",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				term.Errorln("Provide context name as argument. See examples:")
				cmd.Help()
				os.Exit(1)
			} else if len(args) > 1 {
				cmd.Help()
				os.Exit(1)
			}

			name := args[0]
			useContex(name, otx.GetConfigPath(cmd))
		},
	}
	return setCmd
}

func useContex(name, configPath string) {
	config, err := otx.LoadConfig(configPath)
	term.ExitOnError(err)

	if config.CurrentContext == name {
		return
	}

	found := false
	for i := range config.Contexts {
		if config.Contexts[i].Name == name {
			found = true
			break
		}
	}
	if !found {
		term.Fatalln("Invalid context name")
	}

	config.CurrentContext = name
	err = config.Save(configPath)
	term.ExitOnError(err)
}
