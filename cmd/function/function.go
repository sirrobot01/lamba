package function

import (
	"github.com/sirrobot01/lamba/pkg/core"
	"github.com/spf13/cobra"
)

func NewCmd(executor *core.Executor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "function",
		Short: "Manage Lambda functions",
		Long:  `Manage Lambda functions including creating, editing, and listing functions.`,
	}
	cmd.AddCommand(newCreateCommand(executor))
	cmd.AddCommand(newListCommand(executor))
	return cmd
}
