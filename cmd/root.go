package cmd

import (
	"fmt"
	"github.com/sirrobot01/lamba/cmd/function"
	"github.com/sirrobot01/lamba/cmd/invoke"
	"github.com/sirrobot01/lamba/cmd/runtime"
	"github.com/sirrobot01/lamba/cmd/serve"
	"github.com/sirrobot01/lamba/pkg/core"
	"github.com/spf13/cobra"
)

var (
	verbose bool
)

func Execute() error {
	rootCmd := newRootCmd()
	return rootCmd.Execute()
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "lamba",
		Short: "Lamba is a CLI tool for managing Lamba functions",
		Long:  `Lamba is a comprehensive CLI tool for managing self-hosted "lambda-like" functions".`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set up any global configurations based on flags
			if verbose {
				fmt.Println("Verbose mode enabled")
			}
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	registry := core.NewFunctionRegistry()
	runtimeManager := core.NewRuntimeManager()
	goRuntime := core.NewGoRuntime()
	runtimeManager.Register("go", goRuntime)
	executor := core.NewExecutor(registry, runtimeManager)
	listener := core.NewHTTPListener(executor, "8080")

	rootCmd.AddCommand(runtime.NewCmd(executor))
	rootCmd.AddCommand(function.NewCmd(executor))
	rootCmd.AddCommand(invoke.NewCmd(executor))

	// Add a new command to start the server
	rootCmd.AddCommand(serve.NewCmd(listener))

	return rootCmd
}
