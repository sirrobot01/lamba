package serve

import (
	"context"
	"fmt"
	"github.com/sirrobot01/lamba/pkg/core"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

func NewCmd(listener core.Listener) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the Lamba server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return serveWithGracefulShutdown(listener)
		},
	}
}

func serveWithGracefulShutdown(listener core.Listener) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := listener.Start(); err != nil {
			fmt.Printf("Error starting listener: %v\n", err)
			stop()
		}
	}()

	<-ctx.Done()
	fmt.Println("Shutting down gracefully...")
	return listener.Stop()
}
