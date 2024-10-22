package cmd

import (
	"fmt"
	"github.com/sirrobot01/lamba/pkg/executor"

	"github.com/spf13/cobra"
)

func newFunctionCreateCommand(ex *executor.Executor) *cobra.Command {
	var (
		name    string
		runtime string
		timeout int
		handler string
		file    string
	)
	createCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new function",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Adding a new function...")
			if err := ex.CreateFunction(name, runtime, handler, timeout, file, ""); err != nil {
				return err
			}
			fmt.Printf("Function '%s' created successfully\n", name)
			return nil
		},
	}
	createCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the function to create")
	createCmd.Flags().StringVarP(&runtime, "runtime", "r", "", "Runtime for the function")
	createCmd.Flags().StringVarP(&handler, "handler", "", "", "Handler for the function")
	createCmd.Flags().IntVarP(&timeout, "timeout", "t", 10, "Function timeout in seconds")
	createCmd.Flags().StringVarP(&file, "file", "f", "", "Path to the function code")

	_ = createCmd.MarkFlagRequired("name")
	_ = createCmd.MarkFlagRequired("runtime")
	_ = createCmd.MarkFlagRequired("handler")
	_ = createCmd.MarkFlagRequired("file")

	return createCmd
}

func NewFunctionCmd(executor *executor.Executor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "function",
		Short: "Manage Lambda functions",
		Long:  `Manage Lambda functions including creating, editing, and listing functions.`,
	}
	cmd.AddCommand(newFunctionCreateCommand(executor))
	cmd.AddCommand(newFunctionListCommand(executor))
	return cmd
}

func newFunctionListCommand(executor *executor.Executor) *cobra.Command {
	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List all available functions",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Listing available functions...")
			functions := executor.FunctionRegistry.List()
			for _, f := range functions {
				fmt.Printf("Name: %s\n", f.Name)
				fmt.Printf("Runtime: %s\n", f.Runtime)
				fmt.Printf("Handler: %s\n", f.Handler)
				fmt.Printf("Timeout: %d\n", f.Timeout)
				fmt.Println("--------------------")
			}
		},
	}
	return lsCmd
}
