package cmds

import (
	"os"

	"github.com/appscode/go/term"
	otx "github.com/appscode/osm/context"
	"github.com/spf13/cobra"
)

type containerMakeRequest struct {
	context   string
	container string
}

func NewCmdMakeContainer() *cobra.Command {
	req := &containerMakeRequest{}
	cmd := &cobra.Command{
		Use:               "mc <name>",
		Short:             "Make container",
		Example:           "osm mc mybucket",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				term.Errorln("Provide container name as argument. See examples:")
				cmd.Help()
				os.Exit(1)
			} else if len(args) > 1 {
				cmd.Help()
				os.Exit(1)
			}

			req.container = args[0]
			makeContainer(req, otx.GetConfigPath(cmd))
		},
	}

	cmd.Flags().StringVarP(&req.context, "context", "", "", "Name of osmconfig context to use")
	return cmd
}

func makeContainer(req *containerMakeRequest, configPath string) {
	cfg, err := otx.LoadConfig(configPath)
	term.ExitOnError(err)

	loc, err := cfg.Dial(req.context)
	term.ExitOnError(err)

	_, err = loc.Container(req.container)
	if err != nil {
		_, err = loc.CreateContainer(req.container)
		term.ExitOnError(err)
		term.Successln("Successfully created container " + req.container)
	} else {
		term.Infoln("Container " + req.container + " already exists!")
	}
}
