package cmds

import (
	"flag"
	"path/filepath"
	"strings"

	v "github.com/appscode/go/version"
	"github.com/appscode/kutil/tools/analytics"
	cfgCmd "github.com/appscode/osm/cmds/config"
	"github.com/jpillora/go-ogle-analytics"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

const (
	gaTrackingCode = "UA-62096468-20"
)

func NewCmdOsm() *cobra.Command {
	var (
		enableAnalytics = true
	)
	rootCmd := &cobra.Command{
		Use:               "osm [command]",
		Short:             `Object Store Manipulator by AppsCode`,
		DisableAutoGenTag: true,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			if enableAnalytics && gaTrackingCode != "" {
				if client, err := ga.NewClient(gaTrackingCode); err == nil {
					client.ClientID(analytics.ClientID())
					parts := strings.Split(c.CommandPath(), " ")
					client.Send(ga.NewEvent(parts[0], strings.Join(parts[1:], "/")).Label(v.Version.Version))
				}
			}
		},
		Run: func(c *cobra.Command, args []string) {
			c.Help()
		},
	}
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	rootCmd.PersistentFlags().String("osmconfig", filepath.Join(homedir.HomeDir(), ".osm", "config"), "Path to osm config")
	rootCmd.PersistentFlags().BoolVar(&enableAnalytics, "enable-analytics", enableAnalytics, "Send usage events to Google Analytics")

	rootCmd.PersistentFlags().BoolVar(&enableAnalytics, "analytics", enableAnalytics, "Send usage events to Google Analytics")
	rootCmd.PersistentFlags().MarkDeprecated("analytics", "use --enable-analytics")

	rootCmd.AddCommand(cfgCmd.NewCmdConfig())

	rootCmd.AddCommand(NewCmdListContainers())
	rootCmd.AddCommand(NewCmdMakeContainer())
	rootCmd.AddCommand(NewCmdRemoveContainer())

	rootCmd.AddCommand(NewCmdListIetms())
	rootCmd.AddCommand(NewCmdPush())
	rootCmd.AddCommand(NewCmdPull())
	rootCmd.AddCommand(NewCmdStat())
	rootCmd.AddCommand(NewCmdRemove())

	rootCmd.AddCommand(v.NewCmdVersion())

	return rootCmd
}
