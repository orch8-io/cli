package cmd

import (
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
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		breakers, err := client.ListCircuitBreakers(ctx)
		if err != nil {
			return output.Errorf("listing circuit breakers: %w", err)
		}
		if flagJSON {
			output.JSON(breakers)
			return nil
		}
		headers := []string{"HANDLER", "STATE", "FAILURES", "LAST FAILURE"}
		var rows [][]string
		for _, b := range breakers {
			rows = append(rows, []string{
				b.Handler, b.State, fmt.Sprintf("%d", b.FailureCount), b.LastFailure,
			})
		}
		output.Table(headers, rows)
		return nil
	},
}

var cbGetCmd = &cobra.Command{
	Use:   "get <handler>",
	Short: "Get circuit breaker state for a handler",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		b, err := client.GetCircuitBreaker(ctx, args[0])
		if err != nil {
			return output.Errorf("getting circuit breaker: %w", err)
		}
		if flagJSON {
			output.JSON(b)
			return nil
		}
		fmt.Printf("Handler: %s\n", b.Handler)
		fmt.Printf("State:   %s\n", b.State)
		return nil
	},
}

var cbResetCmd = &cobra.Command{
	Use:   "reset <handler>",
	Short: "Reset a circuit breaker to closed state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		if err := client.ResetCircuitBreaker(ctx, args[0]); err != nil {
			return output.Errorf("resetting circuit breaker: %w", err)
		}
		fmt.Println("Reset:", args[0])
		return nil
	},
}
