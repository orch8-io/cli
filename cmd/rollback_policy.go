package cmd

import (
	"fmt"

	orch8 "github.com/orch8-io/sdk-go"
	"github.com/orch8-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var rollbackPolicyCmd = &cobra.Command{
	Use:     "rollback-policy",
	Aliases: []string{"rbp"},
	Short:   "Manage rollback policies",
}

func init() {
	rollbackPolicyCmd.AddCommand(rbpListCmd)
	rollbackPolicyCmd.AddCommand(rbpGetCmd)
	rollbackPolicyCmd.AddCommand(rbpCreateCmd)
	rollbackPolicyCmd.AddCommand(rbpDeleteCmd)

	rbpCreateCmd.Flags().String("sequence", "", "Sequence name (required)")
	rbpCreateCmd.Flags().Float64("threshold", 0, "Error rate threshold (required)")
	rbpCreateCmd.Flags().Int("time-window", 0, "Time window in seconds (required)")
	rbpCreateCmd.Flags().Int("cooldown", 0, "Cooldown in seconds")
	rbpCreateCmd.Flags().Int("confirmation-window", 0, "Confirmation window in seconds")
	rbpCreateCmd.Flags().String("webhook-url", "", "Webhook URL for notifications")
}

var rbpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List rollback policies",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		policies, err := client.ListRollbackPolicies(ctx, flagTenantID)
		if err != nil {
			return output.Errorf("listing rollback policies: %w", err)
		}
		if flagJSON {
			output.JSON(policies)
			return nil
		}
		headers := []string{"ID", "SEQUENCE", "THRESHOLD", "TIME WINDOW", "ENABLED"}
		var rows [][]string
		for _, p := range policies {
			rows = append(rows, []string{
				fmt.Sprintf("%d", p.ID),
				p.SequenceName,
				fmt.Sprintf("%.2f", p.ErrorRateThreshold),
				fmt.Sprintf("%ds", p.TimeWindowSecs),
				fmt.Sprintf("%v", p.Enabled),
			})
		}
		output.Table(headers, rows)
		return nil
	},
}

var rbpGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get a rollback policy by sequence name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		p, err := client.GetRollbackPolicy(ctx, args[0])
		if err != nil {
			return output.Errorf("getting rollback policy: %w", err)
		}
		if flagJSON {
			output.JSON(p)
			return nil
		}
		fmt.Printf("ID:                  %d\n", p.ID)
		fmt.Printf("Tenant:              %s\n", p.TenantID)
		fmt.Printf("Sequence:            %s\n", p.SequenceName)
		fmt.Printf("Error Threshold:     %.2f\n", p.ErrorRateThreshold)
		fmt.Printf("Time Window:         %ds\n", p.TimeWindowSecs)
		fmt.Printf("Enabled:             %v\n", p.Enabled)
		fmt.Printf("Cooldown:            %ds\n", p.CooldownSecs)
		fmt.Printf("Confirmation Window: %ds\n", p.ConfirmationWindowSecs)
		if p.WebhookURL != nil {
			fmt.Printf("Webhook URL:         %s\n", *p.WebhookURL)
		}
		fmt.Printf("Created:             %s\n", p.CreatedAt)
		fmt.Printf("Updated:             %s\n", p.UpdatedAt)
		return nil
	},
}

var rbpCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a rollback policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()

		sequence, err := cmd.Flags().GetString("sequence")
		if err != nil {
			return output.Errorf("reading --sequence flag: %w", err)
		}
		if sequence == "" {
			return output.Errorf("--sequence is required")
		}

		threshold, err := cmd.Flags().GetFloat64("threshold")
		if err != nil {
			return output.Errorf("reading --threshold flag: %w", err)
		}
		if threshold == 0 {
			return output.Errorf("--threshold is required")
		}

		timeWindow, err := cmd.Flags().GetInt("time-window")
		if err != nil {
			return output.Errorf("reading --time-window flag: %w", err)
		}
		if timeWindow == 0 {
			return output.Errorf("--time-window is required")
		}

		body := &orch8.CreatePolicyRequest{
			TenantID:           flagTenantID,
			SequenceName:       sequence,
			ErrorRateThreshold: threshold,
			TimeWindowSecs:     timeWindow,
		}

		if cmd.Flags().Changed("cooldown") {
			v, _ := cmd.Flags().GetInt("cooldown")
			body.CooldownSecs = &v
		}
		if cmd.Flags().Changed("confirmation-window") {
			v, _ := cmd.Flags().GetInt("confirmation-window")
			body.ConfirmationWindowSecs = &v
		}
		if cmd.Flags().Changed("webhook-url") {
			v, _ := cmd.Flags().GetString("webhook-url")
			body.WebhookURL = &v
		}

		p, err := client.CreateRollbackPolicy(ctx, body)
		if err != nil {
			return output.Errorf("creating rollback policy: %w", err)
		}
		if flagJSON {
			output.JSON(p)
			return nil
		}
		fmt.Printf("Created rollback policy: %d (%s)\n", p.ID, p.SequenceName)
		return nil
	},
}

var rbpDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a rollback policy by sequence name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		if err := client.DeleteRollbackPolicy(ctx, args[0]); err != nil {
			return output.Errorf("deleting rollback policy: %w", err)
		}
		fmt.Println("Deleted:", args[0])
		return nil
	},
}
