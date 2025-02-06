package main

import (
	"github.com/rs/zerolog/log"
	"github.com/sirrobot01/lamba/cmd/lamba"
	"os"
)

func main() {
	if err := lamba.Start(); err != nil {
		log.Info().Msgf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
