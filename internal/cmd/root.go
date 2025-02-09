package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var flake string

func init() {
	rootCmd.PersistentFlags().StringVar(&flake, "flake", "f", "flake to be used to get options")
}

var rootCmd = &cobra.Command{
	Use:   "nixcfg",
	Short: "nixcfg is a tool to inspect nixos options",
	Long: `nixcfg is a command line tool including a tui that makes it really
				easy to inspect nixos configurations`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
