package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/orch8-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage plugins",
}

func init() {
	pluginCmd.AddCommand(plugListCmd)
	pluginCmd.AddCommand(plugGetCmd)
	pluginCmd.AddCommand(plugCreateCmd)
	pluginCmd.AddCommand(plugDeleteCmd)
	pluginCmd.AddCommand(plugUpdateCmd)

	plugCreateCmd.Flags().String("file", "", "Path to plugin JSON file")
	plugUpdateCmd.Flags().String("file", "", "Path to JSON file with update fields (required)")
}

var plugListCmd = &cobra.Command{
	Use:   "list",
	Short: "List plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		plugins, err := client.ListPlugins(ctx, flagTenantID)
		if err != nil {
			return output.Errorf("listing plugins: %w", err)
		}
		if flagJSON {
			output.JSON(plugins)
			return nil
		}
		headers := []string{"NAME", "TYPE", "ENABLED"}
		var rows [][]string
		for _, p := range plugins {
			rows = append(rows, []string{
				p.Name, p.PluginType, fmt.Sprintf("%v", p.Enabled),
			})
		}
		output.Table(headers, rows)
		return nil
	},
}

var plugGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get a plugin by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		p, err := client.GetPlugin(ctx, args[0])
		if err != nil {
			return output.Errorf("getting plugin: %w", err)
		}
		if flagJSON {
			output.JSON(p)
			return nil
		}
		fmt.Printf("Name: %s\n", p.Name)
		fmt.Printf("Type: %s\n", p.PluginType)
		return nil
	},
}

var plugCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Register a plugin from JSON file",
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
		p, err := client.CreatePlugin(ctx, body)
		if err != nil {
			return output.Errorf("creating plugin: %w", err)
		}
		if flagJSON {
			output.JSON(p)
			return nil
		}
		fmt.Printf("Created plugin: %s\n", p.Name)
		return nil
	},
}

var plugDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		if err := client.DeletePlugin(ctx, args[0]); err != nil {
			return output.Errorf("deleting plugin: %w", err)
		}
		fmt.Println("Deleted:", args[0])
		return nil
	},
}

var plugUpdateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Update a plugin",
	Args:  cobra.ExactArgs(1),
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
		p, err := client.UpdatePlugin(ctx, args[0], body)
		if err != nil {
			return output.Errorf("updating plugin: %w", err)
		}
		if flagJSON {
			output.JSON(p)
			return nil
		}
		fmt.Printf("Updated: %s\n", p.Name)
		return nil
	},
}
