package cmds

import (
	"os"

	"github.com/appscode/go/term"
	otx "github.com/appscode/osm/context"
	"github.com/graymeta/stow"
	"github.com/spf13/cobra"
)

type itemRemoveRequest struct {
	context   string
	container string
	itemID    string
}

func NewCmdRemove() *cobra.Command {
	req := &itemRemoveRequest{}
	cmd := &cobra.Command{
		Use:               "rm <id>",
		Short:             "Remove item from container",
		Example:           "osm rm -c mybucket f1.txt",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				term.Errorln("Provide item id as argument. See examples:")
				cmd.Help()
				os.Exit(1)
			} else if len(args) > 1 {
				cmd.Help()
				os.Exit(1)
			}

			req.itemID = args[0]
			removeItem(req, otx.GetConfigPath(cmd))
		},
	}

	cmd.Flags().StringVar(&req.context, "context", "", "Name of osmconfig context to use")
	cmd.Flags().StringVarP(&req.container, "container", "c", "", "Name of container")
	return cmd
}

func removeItem(req *itemRemoveRequest, configPath string) {
	cfg, err := otx.LoadConfig(configPath)
	term.ExitOnError(err)

	loc, err := cfg.Dial(req.context)
	term.ExitOnError(err)

	c, err := loc.Container(req.container)
	term.ExitOnError(err)

	cursor := stow.CursorStart
	for {
		items, next, err := c.Items(req.itemID, cursor, 50)
		term.ExitOnError(err)
		for _, item := range items {
			err = c.RemoveItem(item.ID())
			term.ExitOnError(err)
			term.Successln("Successfully removed item " + item.ID())
		}
		cursor = next
		if stow.IsCursorEnd(cursor) {
			break
		}
	}
}
