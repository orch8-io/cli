package cmd

import (
	"context"
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

	plugCreateCmd.Flags().String("file", "", "Path to plugin JSON file")
}

var plugListCmd = &cobra.Command{
	Use:   "list",
	Short: "List plugins",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		plugins, err := client.ListPlugins(ctx, flagTenantID)
		if err != nil {
			output.Error("listing plugins: %v", err)
		}
		if flagJSON {
			output.JSON(plugins)
			return
		}
		headers := []string{"NAME", "TYPE", "ENABLED"}
		var rows [][]string
		for _, p := range plugins {
			rows = append(rows, []string{
				p.Name, p.PluginType, fmt.Sprintf("%v", p.Enabled),
			})
		}
		output.Table(headers, rows)
	},
}

var plugGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get a plugin by name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		p, err := client.GetPlugin(ctx, args[0])
		if err != nil {
			output.Error("getting plugin: %v", err)
		}
		output.JSON(p)
	},
}

var plugCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Register a plugin from JSON file",
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
		p, err := client.CreatePlugin(ctx, body)
		if err != nil {
			output.Error("creating plugin: %v", err)
		}
		if flagJSON {
			output.JSON(p)
			return
		}
		fmt.Printf("Created plugin: %s\n", p.Name)
	},
}

var plugDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a plugin",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		if err := client.DeletePlugin(ctx, args[0]); err != nil {
			output.Error("deleting plugin: %v", err)
		}
		fmt.Println("Deleted:", args[0])
	},
}
