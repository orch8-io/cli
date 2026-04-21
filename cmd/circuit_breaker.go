package cmd

import (
	"context"
	"fmt"

	"github.com/orch8-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var cbCmd = &cobra.Command{
	Use:     "circuit-breaker",
	Aliases: []string{"cb"},
	Short:   "Manage circuit breakers",
}

func init() {
	cbCmd.AddCommand(cbListCmd)
	cbCmd.AddCommand(cbGetCmd)
	cbCmd.AddCommand(cbResetCmd)
}

var cbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List circuit breaker states",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		breakers, err := client.ListCircuitBreakers(ctx)
		if err != nil {
			output.Error("listing circuit breakers: %v", err)
		}
		if flagJSON {
			output.JSON(breakers)
			return
		}
		headers := []string{"HANDLER", "STATE", "FAILURES", "LAST FAILURE"}
		var rows [][]string
		for _, b := range breakers {
			rows = append(rows, []string{
				b.Handler, b.State, fmt.Sprintf("%d", b.FailureCount), b.LastFailure,
			})
		}
		output.Table(headers, rows)
	},
}

var cbGetCmd = &cobra.Command{
	Use:   "get <handler>",
	Short: "Get circuit breaker state for a handler",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		b, err := client.GetCircuitBreaker(ctx, args[0])
		if err != nil {
			output.Error("getting circuit breaker: %v", err)
		}
		output.JSON(b)
	},
}

var cbResetCmd = &cobra.Command{
	Use:   "reset <handler>",
	Short: "Reset a circuit breaker to closed state",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		if err := client.ResetCircuitBreaker(ctx, args[0]); err != nil {
			output.Error("resetting circuit breaker: %v", err)
		}
		fmt.Println("Reset:", args[0])
	},
}
