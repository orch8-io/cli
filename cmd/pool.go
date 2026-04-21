package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/orch8-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var poolCmd = &cobra.Command{
	Use:   "pool",
	Short: "Manage resource pools",
}

func init() {
	poolCmd.AddCommand(poolListCmd)
	poolCmd.AddCommand(poolGetCmd)
	poolCmd.AddCommand(poolCreateCmd)
	poolCmd.AddCommand(poolDeleteCmd)
	poolCmd.AddCommand(poolResourcesCmd)

	poolCreateCmd.Flags().String("file", "", "Path to pool JSON file")
}

var poolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List resource pools",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		pools, err := client.ListPools(ctx, flagTenantID)
		if err != nil {
			return output.Errorf("listing pools: %w", err)
		}
		if flagJSON {
			output.JSON(pools)
			return nil
		}
		headers := []string{"ID", "NAME", "MAX", "CURRENT"}
		var rows [][]string
		for _, p := range pools {
			rows = append(rows, []string{
				p.ID, p.Name, fmt.Sprintf("%d", p.MaxSize), fmt.Sprintf("%d", p.CurrentSize),
			})
		}
		output.Table(headers, rows)
		return nil
	},
}

var poolGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a resource pool by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		p, err := client.GetPool(ctx, args[0])
		if err != nil {
			return output.Errorf("getting pool: %w", err)
		}
		if flagJSON {
			output.JSON(p)
			return nil
		}
		fmt.Printf("ID:      %s\n", p.ID)
		fmt.Printf("Name:    %s\n", p.Name)
		fmt.Printf("Max:     %d\n", p.MaxSize)
		fmt.Printf("Current: %d\n", p.CurrentSize)
		return nil
	},
}

var poolCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a resource pool from JSON file",
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
		p, err := client.CreatePool(ctx, body)
		if err != nil {
			return output.Errorf("creating pool: %w", err)
		}
		if flagJSON {
			output.JSON(p)
			return nil
		}
		fmt.Printf("Created pool: %s\n", p.ID)
		return nil
	},
}

var poolDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a resource pool",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		if err := client.DeletePool(ctx, args[0]); err != nil {
			return output.Errorf("deleting pool: %w", err)
		}
		fmt.Println("Deleted:", args[0])
		return nil
	},
}

var poolResourcesCmd = &cobra.Command{
	Use:   "resources <pool-id>",
	Short: "List resources in a pool",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		resources, err := client.ListPoolResources(ctx, args[0])
		if err != nil {
			return output.Errorf("listing pool resources: %w", err)
		}
		if flagJSON {
			output.JSON(resources)
			return nil
		}
		headers := []string{"ID", "KEY", "STATE", "LOCKED BY"}
		var rows [][]string
		for _, r := range resources {
			locked := r.LockedBy
			rows = append(rows, []string{
				r.ID, r.ResourceKey, r.State, locked,
			})
		}
		output.Table(headers, rows)
		return nil
	},
}
