package cmd

import (
	"fmt"

	"github.com/eblechschmidt/nixcfg/internal/options"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use:   "show option",
	Short: "Print option details of the given option",
	Long:  `Print option details of the given option`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("Option needs to be specified as an argument")
		}
		res, err := options.Show(flake, args[0])
		if err != nil {
			return err
		}
		fmt.Print(res)
		return nil
	},
}
