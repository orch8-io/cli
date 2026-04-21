package cmd

import (
	"github.com/orch8-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var approvalCmd = &cobra.Command{
	Use:   "approval",
	Short: "Manage approvals",
}

func init() {
	approvalCmd.AddCommand(approvalListCmd)
}

var approvalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List instances awaiting approval",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		approvals, err := client.ListApprovals(ctx, map[string]string{"tenant_id": flagTenantID})
		if err != nil {
			return output.Errorf("listing approvals: %w", err)
		}
		if flagJSON {
			output.JSON(approvals)
			return nil
		}
		headers := []string{"ID", "SEQUENCE", "STATE", "CREATED"}
		var rows [][]string
		for _, a := range approvals {
			rows = append(rows, []string{
				a.ID, a.SequenceID, a.State, a.CreatedAt,
			})
		}
		output.Table(headers, rows)
		return nil
	},
}
