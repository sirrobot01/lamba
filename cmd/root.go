package cmd

import (
	"fmt"
	"github.com/sirrobot01/lamba/pkg/executor"
	"github.com/sirrobot01/lamba/pkg/function"
	"github.com/sirrobot01/lamba/pkg/invoker"
	runtime2 "github.com/sirrobot01/lamba/pkg/runtime"
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

	registry := function.NewRegistry()
	runtimeManager := runtime2.NewManager()
	runtimes := map[string]runtime2.Runtime{
		"python": runtime2.NewPythonRuntime(),
		"nodejs": runtime2.NewNodeJSRuntime(),
	}
	runtimeManager.Register(runtimes)
	ex := executor.NewExecutor(registry, runtimeManager)
	httpInvoker := invoker.NewHTTPInvoker(ex, "8080")

	rootCmd.AddCommand(NewRuntimeCmd(ex))
	rootCmd.AddCommand(NewFunctionCmd(ex))
	rootCmd.AddCommand(NewInvokerCmd(ex))

	// Add a new command to start the server
	rootCmd.AddCommand(NewServeCmd(httpInvoker))

	return rootCmd
}
