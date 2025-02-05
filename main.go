package main

import (
	"github.com/rs/zerolog/log"
	"os"

	"github.com/sirrobot01/lamba/cmd"
)

func main() {
	if err := cmd.Start(); err != nil {
		log.Info().Msgf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
