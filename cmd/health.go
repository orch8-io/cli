package cmd

import (
	"fmt"

	"github.com/orch8-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check engine health",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		resp, err := client.Health(ctx)
		if err != nil {
			return output.Errorf("health check failed: %w", err)
		}
		if flagJSON {
			output.JSON(resp)
			return nil
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", resp.Status)
		return nil
	},
}
