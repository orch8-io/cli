package cmd

import (
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
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		triggers, err := client.ListTriggers(ctx, flagTenantID)
		if err != nil {
			return output.Errorf("listing triggers: %w", err)
		}
		if flagJSON {
			output.JSON(triggers)
			return nil
		}
		headers := []string{"SLUG", "TYPE", "SEQUENCE", "ENABLED"}
		var rows [][]string
		for _, t := range triggers {
			rows = append(rows, []string{
				t.Slug, t.TriggerType, t.SequenceName, fmt.Sprintf("%v", t.Enabled),
			})
		}
		output.Table(headers, rows)
		return nil
	},
}

var trigGetCmd = &cobra.Command{
	Use:   "get <slug>",
	Short: "Get a trigger by slug",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		t, err := client.GetTrigger(ctx, args[0])
		if err != nil {
			return output.Errorf("getting trigger: %w", err)
		}
		if flagJSON {
			output.JSON(t)
			return nil
		}
		fmt.Printf("Slug:     %s\n", t.Slug)
		fmt.Printf("Sequence: %s\n", t.SequenceName)
		fmt.Printf("Enabled:  %v\n", t.Enabled)
		return nil
	},
}

var trigCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a trigger from JSON file",
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
		t, err := client.CreateTrigger(ctx, body)
		if err != nil {
			return output.Errorf("creating trigger: %w", err)
		}
		if flagJSON {
			output.JSON(t)
			return nil
		}
		fmt.Printf("Created trigger: %s\n", t.Slug)
		return nil
	},
}

var trigDeleteCmd = &cobra.Command{
	Use:   "delete <slug>",
	Short: "Delete a trigger",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		if err := client.DeleteTrigger(ctx, args[0]); err != nil {
			return output.Errorf("deleting trigger: %w", err)
		}
		fmt.Println("Deleted:", args[0])
		return nil
	},
}

var trigFireCmd = &cobra.Command{
	Use:   "fire <slug>",
	Short: "Fire a trigger manually",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		var payload any
		if p, err := cmd.Flags().GetString("payload"); err != nil {
			return output.Errorf("reading --payload flag: %w", err)
		} else if p != "" {
			if err := json.Unmarshal([]byte(p), &payload); err != nil {
				return output.Errorf("parsing payload: %w", err)
			}
		}
		resp, err := client.FireTrigger(ctx, args[0], payload)
		if err != nil {
			return output.Errorf("firing trigger: %w", err)
		}
		if flagJSON {
			output.JSON(resp)
			return nil
		}
		fmt.Printf("Fired trigger %s -> instance %s\n", args[0], resp.InstanceID)
		return nil
	},
}
