package function

import (
	"fmt"
	"github.com/sirrobot01/lamba/pkg/core"

	"github.com/spf13/cobra"
)

func newCreateCommand(executor *core.Executor) *cobra.Command {
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
			metadata := core.NewFunctionMetadata(name, runtime, handler, timeout, file)
			if err := executor.CreateFunction(metadata); err != nil {
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
