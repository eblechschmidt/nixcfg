package main

import (
	"os"

	"github.com/eblechschmidt/nixcfg/internal/cmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func main() {
	cmd.Execute()
}
