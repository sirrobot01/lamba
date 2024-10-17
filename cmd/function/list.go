package function

import (
	"fmt"
	"github.com/sirrobot01/lamba/pkg/core"

	"github.com/spf13/cobra"
)

func newListCommand(executor *core.Executor) *cobra.Command {
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
