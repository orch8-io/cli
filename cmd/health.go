package cmd

import (
	"context"
	"fmt"

	"github.com/orch8-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check engine health",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		resp, err := client.Health(ctx)
		if err != nil {
			output.Error("health check failed: %v", err)
		}
		if flagJSON {
			output.JSON(resp)
			return
		}
		fmt.Printf("Status: %s\n", resp.Status)
	},
}
