package cmd

import (
	"fmt"

	"github.com/eblechschmidt/nixcfg/internal/fzf"
	"github.com/eblechschmidt/nixcfg/internal/options"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list [option]",
	Short: "list lists all options that belong to [option]",
	Long: `list all options and their values recursively that are childern of
				[option]`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("only one option as argument supported")
		}
		opt := ""
		if len(args) == 1 {
			opt = args[0]
		}

		log.Info().
			Str("flake", flake).
			Str("option", opt).
			Msgf("List optins")

		r, err := options.List(flake, opt)
		if err != nil {
			return err
		}

		fzf, err := fzf.New(
			fzf.WithQuery(opt),
			fzf.WithPreviewCmd(
				fmt.Sprintf("go run cmd/nixcfg/main.go show {-1} --flake %s", flake)),
		)
		if err != nil {
			return err
		}

		for line := range r {
			err = fzf.Add([]string{line.Path})
			if err != nil {
				log.Err(err).Msg("could not send path to fzf")
			}
		}

		_, err = fzf.Selection()
		if err != nil {
			return err
		}

		return nil
	},
}
