package cmd

import (
	"context"
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
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		crons, err := client.ListCron(ctx, flagTenantID)
		if err != nil {
			output.Error("listing cron schedules: %v", err)
		}
		if flagJSON {
			output.JSON(crons)
			return
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
	},
}

var cronGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a cron schedule",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		cron, err := client.GetCron(ctx, args[0])
		if err != nil {
			output.Error("getting cron: %v", err)
		}
		output.JSON(cron)
	},
}

var cronCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a cron schedule from JSON file",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		file, _ := cmd.Flags().GetString("file")
		if file == "" {
			output.Error("--file is required")
		}
		data, err := os.ReadFile(file)
		if err != nil {
			output.Error("reading file: %v", err)
		}
		var body map[string]any
		if err := json.Unmarshal(data, &body); err != nil {
			output.Error("parsing JSON: %v", err)
		}
		cron, err := client.CreateCron(ctx, body)
		if err != nil {
			output.Error("creating cron: %v", err)
		}
		if flagJSON {
			output.JSON(cron)
			return
		}
		fmt.Printf("Created cron: %s\n", cron.ID)
	},
}

var cronDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a cron schedule",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		if err := client.DeleteCron(ctx, args[0]); err != nil {
			output.Error("deleting cron: %v", err)
		}
		fmt.Println("Deleted:", args[0])
	},
}

var cronEnableCmd = &cobra.Command{
	Use:   "enable <id>",
	Short: "Enable a cron schedule",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		if _, err := client.UpdateCron(ctx, args[0], map[string]any{"enabled": true}); err != nil {
			output.Error("enabling cron: %v", err)
		}
		fmt.Println("Enabled:", args[0])
	},
}

var cronDisableCmd = &cobra.Command{
	Use:   "disable <id>",
	Short: "Disable a cron schedule",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		if _, err := client.UpdateCron(ctx, args[0], map[string]any{"enabled": false}); err != nil {
			output.Error("disabling cron: %v", err)
		}
		fmt.Println("Disabled:", args[0])
	},
}
