package main

import (
	"fmt"
	"os"
	"time"

	"github.com/eblechschmidt/nixcfg/internal/options"
	"github.com/eblechschmidt/nixcfg/internal/repl"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func main() {
	start := time.Now()

	r, err := repl.NewWithFlake("/home/eike/repos/nixos")
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create repl")
	}

	t := options.NewTree(r, "nixserve")

	opts, err := t.List("boot")
	if err != nil {
		log.Fatal().Err(err).Msg("Could not generate option list")
	}

	for _, o := range opts {
		fmt.Println(o.Path())
	}

	// o := "bullerbyn.traefik.dataDir"
	// o := "boot.initrd.systemd.contents"
	// val, err := t.Get(o)
	// if err != nil {
	// 	log.Fatal().Err(err)
	// }

	// fmt.Printf("Value:\n  %s\n\n", val.(*options.Option).ValueStr())
	// fmt.Printf("Default:\n  %s\n\n", val.(*options.Option).Default())
	// fmt.Printf("Type:\n  %s\n\n", val.(*options.Option).Type())
	// fmt.Printf("Description:\n  %s\n\n", val.(*options.Option).Description())
	// fmt.Printf("Declared by:\n  %+v\n\n", val.(*options.Option).DeclaredBy())
	// fmt.Printf("Defined by:\n  %+v\n\n", val.(*options.Option).DefinedBy())

	elapsed := time.Since(start)
	fmt.Printf("Evaluation done after %s", elapsed)
}
