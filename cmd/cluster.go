package cmd

import (
	"fmt"

	"github.com/orch8-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage cluster nodes",
}

func init() {
	clusterCmd.AddCommand(clusterListCmd)
	clusterCmd.AddCommand(clusterDrainCmd)
}

var clusterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cluster nodes",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		nodes, err := client.ListClusterNodes(ctx)
		if err != nil {
			return output.Errorf("listing nodes: %w", err)
		}
		if flagJSON {
			output.JSON(nodes)
			return nil
		}
		headers := []string{"ID", "ADDRESS", "STATE", "LAST HEARTBEAT"}
		var rows [][]string
		for _, n := range nodes {
			rows = append(rows, []string{
				n.ID, n.Address, n.State, n.LastHeartbeat,
			})
		}
		output.Table(headers, rows)
		return nil
	},
}

var clusterDrainCmd = &cobra.Command{
	Use:   "drain <node-id>",
	Short: "Drain a node for graceful removal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		if err := client.DrainNode(ctx, args[0]); err != nil {
			return output.Errorf("draining node: %w", err)
		}
		fmt.Println("Draining:", args[0])
		return nil
	},
}
