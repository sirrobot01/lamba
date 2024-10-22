package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirrobot01/lamba/pkg/event"
	"github.com/sirrobot01/lamba/pkg/executor"
	"github.com/sirrobot01/lamba/pkg/function"
	"github.com/sirrobot01/lamba/pkg/runtime"
	"github.com/sirrobot01/lamba/server"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func serveWithGracefulShutdown(server *http.Server) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("Starting server on %s\n", server.Addr)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Error starting server: %v\n", err)
			stop()
		}
	}()

	<-ctx.Done()
	fmt.Println("Shutting down gracefully...")
	return server.Shutdown(context.Background())
}

func Start() error {
	showWelcome()
	registry := function.NewRegistry()
	runtimeManager := runtime.NewManager()
	eventManager := event.NewManager()
	runtimes := map[string]runtime.Runtime{
		"python": runtime.NewPythonRuntime(),
		"nodejs": runtime.NewNodeJSRuntime(),
	}
	if err := runtimeManager.Register(runtimes); err != nil {
		return err
	}
	ex := executor.NewExecutor(registry, runtimeManager, eventManager)
	s := server.NewServer(ex)
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
