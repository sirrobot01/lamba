package runtime

import (
	"fmt"
	"github.com/sirrobot01/lamba/pkg/core"

	"github.com/spf13/cobra"
)

func newListCommand(runtimes *core.RuntimeManager) *cobra.Command {
	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List available runtimes",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Listing available runtimes...")
			_runtimes := runtimes.List()
			for _, runtime := range _runtimes {
				fmt.Printf("Name: %s\n", runtime)
			}
		},
	}
	return lsCmd
}
