package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/orch8-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var triggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "Manage triggers",
}

func init() {
	triggerCmd.AddCommand(trigListCmd)
	triggerCmd.AddCommand(trigGetCmd)
	triggerCmd.AddCommand(trigCreateCmd)
	triggerCmd.AddCommand(trigDeleteCmd)
	triggerCmd.AddCommand(trigFireCmd)

	trigCreateCmd.Flags().String("file", "", "Path to trigger JSON file")
	trigFireCmd.Flags().String("payload", "", "Payload JSON")
}

var trigListCmd = &cobra.Command{
	Use:   "list",
	Short: "List triggers",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		triggers, err := client.ListTriggers(ctx, flagTenantID)
		if err != nil {
			output.Error("listing triggers: %v", err)
		}
		if flagJSON {
			output.JSON(triggers)
			return
		}
		headers := []string{"SLUG", "TYPE", "SEQUENCE", "ENABLED"}
		var rows [][]string
		for _, t := range triggers {
			rows = append(rows, []string{
				t.Slug, t.TriggerType, t.SequenceName, fmt.Sprintf("%v", t.Enabled),
			})
		}
		output.Table(headers, rows)
	},
}

var trigGetCmd = &cobra.Command{
	Use:   "get <slug>",
	Short: "Get a trigger by slug",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		t, err := client.GetTrigger(ctx, args[0])
		if err != nil {
			output.Error("getting trigger: %v", err)
		}
		output.JSON(t)
	},
}

var trigCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a trigger from JSON file",
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
		t, err := client.CreateTrigger(ctx, body)
		if err != nil {
			output.Error("creating trigger: %v", err)
		}
		if flagJSON {
			output.JSON(t)
			return
		}
		fmt.Printf("Created trigger: %s\n", t.Slug)
	},
}

var trigDeleteCmd = &cobra.Command{
	Use:   "delete <slug>",
	Short: "Delete a trigger",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		if err := client.DeleteTrigger(ctx, args[0]); err != nil {
			output.Error("deleting trigger: %v", err)
		}
		fmt.Println("Deleted:", args[0])
	},
}

var trigFireCmd = &cobra.Command{
	Use:   "fire <slug>",
	Short: "Fire a trigger manually",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		var payload any
		if p, _ := cmd.Flags().GetString("payload"); p != "" {
			if err := json.Unmarshal([]byte(p), &payload); err != nil {
				output.Error("parsing payload: %v", err)
			}
		}
		resp, err := client.FireTrigger(ctx, args[0], payload)
		if err != nil {
			output.Error("firing trigger: %v", err)
		}
		if flagJSON {
			output.JSON(resp)
			return
		}
		fmt.Printf("Fired trigger %s -> instance %s\n", args[0], resp.InstanceID)
	},
}
