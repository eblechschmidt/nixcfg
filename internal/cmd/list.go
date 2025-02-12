package cmd

import (
	"fmt"

	"github.com/eblechschmidt/nixcfg/internal/options"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var jsonOut bool

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "output as json")
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
		if jsonOut {
			return jsonList(r, opt)
		}

		return tui(r, opt)
	},
}

func jsonList(r <-chan options.Option, option string) error {
	tree := options.New()
	for o := range r {
		if err := tree.Add(o.Path, o.Value); err != nil {
			log.Err(err)
		}
	}
	out, err := tree.JSON()
	if err != nil {
		return err
	}

	fmt.Print(out)

	return nil
}

func tui(r <-chan options.Option, option string) error {
	for o := range r {
		fmt.Printf("%s = %s\n", o.Path, o.Value)
	}

	// fzf, err := fzf.New(
	// 	fzf.WithQuery(option),
	// 	fzf.WithPreviewCmd(
	// 		fmt.Sprintf("go run cmd/nixcfg/main.go show {-1} --flake %s", flake)),
	// )
	// if err != nil {
	// 	return err
	// }

	// for line := range r {
	// 	err = fzf.Add([]string{line.Path})
	// 	if err != nil {
	// 		log.Err(err).Msg("could not send path to fzf")
	// 	}
	// }

	// _, err = fzf.Selection()
	// if err != nil {
	// 	return err
	// }

	return nil
}
