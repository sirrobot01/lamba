package runtime

import (
	"github.com/sirrobot01/lamba/pkg/core"
	"github.com/spf13/cobra"
)

func NewCmd(executor *core.Executor) *cobra.Command {
	lsCmd := &cobra.Command{
		Use:   "runtime",
		Short: "Manage Lambda runtimes",
		Long:  `Manage Lambda runtimes including listing and adding new runtimes.`,
	}
	lsCmd.AddCommand(newListCommand(executor.RuntimeManager))
	return lsCmd
}
