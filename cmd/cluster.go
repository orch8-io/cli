package cmd

import (
	"context"
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
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		nodes, err := client.ListClusterNodes(ctx)
		if err != nil {
			output.Error("listing nodes: %v", err)
		}
		if flagJSON {
			output.JSON(nodes)
			return
		}
		headers := []string{"ID", "ADDRESS", "STATE", "LAST HEARTBEAT"}
		var rows [][]string
		for _, n := range nodes {
			rows = append(rows, []string{
				n.ID, n.Address, n.State, n.LastHeartbeat,
			})
		}
		output.Table(headers, rows)
	},
}

var clusterDrainCmd = &cobra.Command{
	Use:   "drain <node-id>",
	Short: "Drain a node for graceful removal",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		if err := client.DrainNode(ctx, args[0]); err != nil {
			output.Error("draining node: %v", err)
		}
		fmt.Println("Draining:", args[0])
	},
}
