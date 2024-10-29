package cmd

import (
	"flag"
	"fmt"
	"github.com/sirrobot01/lamba/pkg/event"
	"github.com/sirrobot01/lamba/pkg/executor"
	"github.com/sirrobot01/lamba/pkg/function"
	"github.com/sirrobot01/lamba/pkg/runtime"
	"github.com/sirrobot01/lamba/server"
	"os"
)

func Start() error {
	var (
		port   string
		config string
		help   bool
	)
	flag.StringVar(&port, "port", "8080", "Port to run the server on")
	flag.StringVar(&config, "config", ".", "Path to the config folder")
	flag.BoolVar(&help, "help", false, "Show help")
	flag.Parse()
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		_, _ = fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	if help {
		flag.Usage()
		return nil
	}

	registry := function.NewRegistry(config)
	runtimeManager, err := runtime.NewManager()
	if err != nil {
		return err
	}
	eventManager := event.NewManager(config)
	memory := runtime.NewMemoryManager("128MB")
	runtimes := map[string]runtime.Runtime{
		"python": runtime.NewPythonRuntime(),
		"nodejs": runtime.NewNodeJSRuntime(),
	}
	if err := runtimeManager.Register(runtimes); err != nil {
		return err
	}
	ex := executor.NewExecutor(registry, runtimeManager, eventManager, memory)
	s := server.NewServer(ex, port)
	showWelcome()
	if err := s.Start(); err != nil {
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
