package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/eblechschmidt/nixcfg/internal/parser"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func main() {
	start := time.Now()

	p, err := parser.NewWithFlake("/home/eike/repos/nixos")
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create repl")
	}

	// str := "nixosConfigurations.nixserve.config.stylix"
	// str := "nixosConfigurations.nixserve.config.programs.ssh.knownHosts.nixserve"
	// str := "nixosConfigurations.nixserve.config.assertions"
	// str := "nixosConfigurations.nixserve.config"
	// str := "nixosConfigurations.nixserve.options"
	str := "(builtins.toJSON nixosConfigurations.nixserve.options)"
	res, err := p.Parse(str, true)
	if err != nil {
		log.Fatal().Err(err).Msgf("Could not evaluate %s", str)
	}
	log.Debug().Any("Result", res).Msg("Returned result")

	b, err := json.Marshal(res)
	if err != nil {
		log.Fatal().Err(err).Msg("Error marshalling json")
	}
	fmt.Println(string(b))

	if err := p.Close(); err != nil {
		log.Fatal().Err(err).Msg("Error closing repl")
	}
	elapsed := time.Since(start)
	log.Debug().Msgf("Evaluation done after %s", elapsed)
}
