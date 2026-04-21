package cmd

import (
	"github.com/orch8-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Manage worker tasks",
}

func init() {
	workerCmd.AddCommand(workerListCmd)
	workerCmd.AddCommand(workerStatsCmd)
}

var workerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List worker tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		tasks, err := client.ListWorkerTasks(ctx, map[string]string{"tenant_id": flagTenantID})
		if err != nil {
			return output.Errorf("listing worker tasks: %w", err)
		}
		if flagJSON {
			output.JSON(tasks)
			return nil
		}
		headers := []string{"ID", "HANDLER", "STATE", "INSTANCE"}
		var rows [][]string
		for _, t := range tasks {
			rows = append(rows, []string{
				t.ID, t.HandlerName, t.State, t.InstanceID,
			})
		}
		output.Table(headers, rows)
		return nil
	},
}

var workerStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Get worker task statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		stats, err := client.GetWorkerTaskStats(ctx)
		if err != nil {
			return output.Errorf("getting worker stats: %w", err)
		}
		output.JSON(stats)
		return nil
	},
}
