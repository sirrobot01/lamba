package cmd

import (
	"fmt"
	"github.com/sirrobot01/lamba/pkg/executor"
	"github.com/spf13/cobra"
)

func newRuntimeListCommand(ex *executor.Executor) *cobra.Command {
	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List available runtimes",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Listing available runtimes...")
			_runtimes := ex.RuntimeManager.List()
			for _, rtn := range _runtimes {
				fmt.Printf("Name: %s\n", rtn)
			}
		},
	}
	return lsCmd
}

func NewRuntimeCmd(ex *executor.Executor) *cobra.Command {
	lsCmd := &cobra.Command{
		Use:   "runtime",
		Short: "Manage Lambda runtimes",
		Long:  `Manage Lambda runtimes including listing and adding new runtimes.`,
	}
	lsCmd.AddCommand(newRuntimeListCommand(ex))
	return lsCmd
}
