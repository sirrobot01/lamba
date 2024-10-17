package invoke

import (
	"github.com/sirrobot01/lamba/pkg/core"
	"github.com/spf13/cobra"
)

func NewCmd(executor *core.Executor) *cobra.Command {
	var (
		functionName string
		payload      string
	)
	cmd := &cobra.Command{
		Use:   "invoke",
		Short: "Invoke a Lambda function",
		Run: func(cmd *cobra.Command, args []string) {
			// Convert the payload to a byte array
			payloadBytes := []byte(payload)
			// Invoke the function
			result, err := executor.Execute(functionName, payloadBytes)
			if err != nil {
				cmd.PrintErrf("Error invoking function: %s\n", err)
			} else {
				cmd.Println(string(result))
			}
		},
	}
	cmd.Flags().StringVarP(&functionName, "function", "f", "", "Name of the function to invoke")
	cmd.Flags().StringVarP(&payload, "payload", "p", "", "Payload to pass to the function")

	_ = cmd.MarkFlagRequired("function")
	return cmd
}
