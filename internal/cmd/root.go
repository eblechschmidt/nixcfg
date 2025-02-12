package cmd

import (
	"fmt"
	"os"

	"github.com/eblechschmidt/nixcfg/internal/options"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var flake string

func init() {
	rootCmd.PersistentFlags().StringVar(&flake, "flake", ".#", "flake to be used to get options")
}

var rootCmd = &cobra.Command{
	Use:   "nixcfg [option]",
	Short: "nixcfg is a tool to inspect nixos options",
	Long: `nixcfg is a command line tool including a tui that makes it really
				easy to inspect nixos configurations`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("only one option as argument supported")
		}
		opt := ""
		if len(args) == 1 {
			opt = args[0]
		}
		r, err := options.List(flake, opt)
		if err != nil {
			return err
		}
		log.Debug().Msg("Reading lines")
		for line := range r {
			fmt.Println(line)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
