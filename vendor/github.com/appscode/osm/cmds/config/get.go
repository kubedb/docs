package config

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/appscode/go/term"
	otx "github.com/appscode/osm/context"
	"github.com/spf13/cobra"
)

func newCmdGet() *cobra.Command {
	setCmd := &cobra.Command{
		Use:               "get-contexts",
		Short:             "List available contexts",
		Example:           "osm config get-contexts",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				cmd.Help()
				os.Exit(1)
			}
			getContexts(otx.GetConfigPath(cmd))
		},
	}
	return setCmd
}

func getContexts(configPath string) {
	config, err := otx.LoadConfig(configPath)
	term.ExitOnError(err)

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintln(w, "CURRENT\tNAME\tPROVIDER")
	ctx := config.CurrentContext
	for _, osmCtx := range config.Contexts {
		if osmCtx.Name == ctx {
			fmt.Fprintf(w, "*\t%s\t%s\n", osmCtx.Name, osmCtx.Provider)
		} else {
			fmt.Fprintf(w, "\t%s\t%s\n", osmCtx.Name, osmCtx.Provider)
		}
	}
	w.Flush()
}
