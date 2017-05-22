package main

import (
	"flag"
	"log"

	v "github.com/appscode/go/version"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "operator [command]",
		Short: `KubeDB operator by AppsCode`,
		Run: func(c *cobra.Command, args []string) {
			c.Help()
		},
	}
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	rootCmd.AddCommand(NewCmdRun())
	rootCmd.AddCommand(v.NewCmdVersion())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
