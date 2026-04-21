package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/orch8-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Manage cron schedules",
}

func init() {
	cronCmd.AddCommand(cronListCmd)
	cronCmd.AddCommand(cronGetCmd)
	cronCmd.AddCommand(cronCreateCmd)
	cronCmd.AddCommand(cronDeleteCmd)
	cronCmd.AddCommand(cronEnableCmd)
	cronCmd.AddCommand(cronDisableCmd)

	cronCreateCmd.Flags().String("file", "", "Path to cron schedule JSON file")
}

var cronListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cron schedules",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		crons, err := client.ListCron(ctx, flagTenantID)
		if err != nil {
			return output.Errorf("listing cron schedules: %w", err)
		}
		if flagJSON {
			output.JSON(crons)
			return nil
		}
		headers := []string{"ID", "EXPR", "TIMEZONE", "ENABLED", "SEQUENCE", "NEXT FIRE"}
		var rows [][]string
		for _, c := range crons {
			next := ""
			if c.NextFireAt != nil {
				next = *c.NextFireAt
			}
			rows = append(rows, []string{
				c.ID, c.CronExpr, c.Timezone, fmt.Sprintf("%v", c.Enabled), c.SequenceID, next,
			})
		}
		output.Table(headers, rows)
		return nil
	},
}

var cronGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a cron schedule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		cron, err := client.GetCron(ctx, args[0])
		if err != nil {
			return output.Errorf("getting cron: %w", err)
		}
		if flagJSON {
			output.JSON(cron)
			return nil
		}
		fmt.Printf("ID:       %s\n", cron.ID)
		fmt.Printf("Expr:     %s\n", cron.CronExpr)
		fmt.Printf("Enabled:  %v\n", cron.Enabled)
		return nil
	},
}

var cronCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a cron schedule from JSON file",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			return output.Errorf("reading --file flag: %w", err)
		}
		if file == "" {
			return output.Errorf("--file is required")
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return output.Errorf("reading file: %w", err)
		}
		var body map[string]any
		if err := json.Unmarshal(data, &body); err != nil {
			return output.Errorf("parsing JSON: %w", err)
		}
		cron, err := client.CreateCron(ctx, body)
		if err != nil {
			return output.Errorf("creating cron: %w", err)
		}
		if flagJSON {
			output.JSON(cron)
			return nil
		}
		fmt.Printf("Created cron: %s\n", cron.ID)
		return nil
	},
}

var cronDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a cron schedule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		if err := client.DeleteCron(ctx, args[0]); err != nil {
			return output.Errorf("deleting cron: %w", err)
		}
		fmt.Println("Deleted:", args[0])
		return nil
	},
}

var cronEnableCmd = &cobra.Command{
	Use:   "enable <id>",
	Short: "Enable a cron schedule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		if _, err := client.UpdateCron(ctx, args[0], map[string]any{"enabled": true}); err != nil {
			return output.Errorf("enabling cron: %w", err)
		}
		fmt.Println("Enabled:", args[0])
		return nil
	},
}

var cronDisableCmd = &cobra.Command{
	Use:   "disable <id>",
	Short: "Disable a cron schedule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		if _, err := client.UpdateCron(ctx, args[0], map[string]any{"enabled": false}); err != nil {
			return output.Errorf("disabling cron: %w", err)
		}
		fmt.Println("Disabled:", args[0])
		return nil
	},
}
