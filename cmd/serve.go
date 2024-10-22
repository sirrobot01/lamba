package cmd

import (
	"context"
	"fmt"
	"github.com/sirrobot01/lamba/pkg/invoker"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

func NewServeCmd(inv invoker.Invoker) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the Lamba server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return serveWithGracefulShutdown(inv)
		},
	}
}

func serveWithGracefulShutdown(inv invoker.Invoker) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := inv.Start(); err != nil {
			fmt.Printf("Error starting listener: %v\n", err)
			stop()
		}
	}()

	<-ctx.Done()
	fmt.Println("Shutting down gracefully...")
	return inv.Stop()
}
