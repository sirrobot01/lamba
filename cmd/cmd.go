package cmd

import (
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sirrobot01/lamba/pkg/event"
	"github.com/sirrobot01/lamba/pkg/executor"
	"github.com/sirrobot01/lamba/pkg/function"
	"github.com/sirrobot01/lamba/pkg/runtime"
	"github.com/sirrobot01/lamba/pkg/runtime/engines"
	"github.com/sirrobot01/lamba/server"
	"os"
	"strings"
)

func Start() error {

	var (
		port       string
		config     string
		help       bool
		debug      bool
		engineType engines.Type
	)
	flag.StringVar(&port, "port", "8080", "Port to run the server on")
	flag.Var(&engineType, "engine", "Runtime engine to use. Options: [docker, containerd]. Default: docker")
	flag.StringVar(&config, "config", ".", "Path to the config folder")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.BoolVar(&help, "help", false, "Show help")
	flag.Parse()

	// Set log level
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    false,
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
		},
		FormatCaller: func(i interface{}) string {
			if i == nil {
				return ""
			}
			return fmt.Sprintf("%s |", i)
		},
	}

	// Enable caller tracking
	log.Logger = log.Logger.With().Logger().Output(writer)
	if debug {
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	} else {
		log.Logger = log.Logger.Level(zerolog.InfoLevel)
	}

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		_, _ = fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	if help {
		flag.Usage()
		return nil
	}

	showWelcome()
	log.Info().Msgf("Using %s engine", engineType)
	registry := function.NewRegistry(config)
	runtimeManager := runtime.NewManager(engineType)
	eventManager := event.NewManager(config)
	memory := runtime.NewMemoryManager("128MB")
	runtimes := runtime.GetRuntimes(engineType)
	if err := runtimeManager.Register(runtimes); err != nil {
		log.Info().Err(err).Msg("Failed to register runtimes")
		return err
	}
	s := server.NewServer(executor.NewExecutor(registry, runtimeManager, eventManager, memory), port)
	if err := s.Start(); err != nil {
		log.Info().Err(err).Msgf("Failed to start server")
		return err
	}
	return nil
}

func showWelcome() {
	welcome := `
    ██╗      █████╗ ███╗   ███╗██████╗  █████╗ 
    ██║     ██╔══██╗████╗ ████║██╔══██╗██╔══██╗
    ██║     ███████║██╔████╔██║██████╔╝███████║
    ██║     ██╔══██║██║╚██╔╝██║██╔══██╗██╔══██║
    ███████╗██║  ██║██║ ╚═╝ ██║██████╔╝██║  ██║
    ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚═════╝ ╚═╝  ╚═╝
                                                
    Welcome to Lamba! Your serverless platform.
    `
	fmt.Println(welcome)
}
