package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/orch8-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var instanceCmd = &cobra.Command{
	Use:     "instance",
	Aliases: []string{"inst", "i"},
	Short:   "Manage workflow instances",
}

func init() {
	instanceCmd.AddCommand(instListCmd)
	instanceCmd.AddCommand(instGetCmd)
	instanceCmd.AddCommand(instCreateCmd)
	instanceCmd.AddCommand(instSignalCmd)
	instanceCmd.AddCommand(instPauseCmd)
	instanceCmd.AddCommand(instResumeCmd)
	instanceCmd.AddCommand(instCancelCmd)
	instanceCmd.AddCommand(instRetryCmd)
	instanceCmd.AddCommand(instOutputsCmd)
	instanceCmd.AddCommand(instTreeCmd)
	instanceCmd.AddCommand(instAuditCmd)
	instanceCmd.AddCommand(instDLQCmd)

	instListCmd.Flags().String("sequence", "", "Filter by sequence ID")
	instListCmd.Flags().String("state", "", "Filter by state")
	instListCmd.Flags().Int("limit", 50, "Max results")

	instCreateCmd.Flags().String("sequence", "", "Sequence ID (required)")
	instCreateCmd.Flags().String("context-file", "", "Path to context JSON file")
	instCreateCmd.Flags().String("context", "", "Inline context JSON")
	instCreateCmd.Flags().String("idempotency-key", "", "Idempotency key")
	instCreateCmd.MarkFlagRequired("sequence")

	instSignalCmd.Flags().String("payload", "", "Signal payload as JSON")
}

var instListCmd = &cobra.Command{
	Use:   "list",
	Short: "List instances",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		filters := map[string]string{"tenant_id": flagTenantID}
		if v, _ := cmd.Flags().GetString("sequence"); v != "" {
			filters["sequence_id"] = v
		}
		if v, _ := cmd.Flags().GetString("state"); v != "" {
			filters["state"] = v
		}
		instances, err := client.ListInstances(ctx, filters)
		if err != nil {
			output.Error("listing instances: %v", err)
		}
		if flagJSON {
			output.JSON(instances)
			return
		}
		headers := []string{"ID", "SEQUENCE", "STATE", "CREATED"}
		var rows [][]string
		for _, inst := range instances {
			rows = append(rows, []string{
				inst.ID, inst.SequenceID, inst.State, inst.CreatedAt,
			})
		}
		output.Table(headers, rows)
	},
}

var instGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get instance details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		inst, err := client.GetInstance(ctx, args[0])
		if err != nil {
			output.Error("getting instance: %v", err)
		}
		output.JSON(inst)
	},
}

var instCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new instance",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		seqID, _ := cmd.Flags().GetString("sequence")

		var ctxData any
		if file, _ := cmd.Flags().GetString("context-file"); file != "" {
			data, err := os.ReadFile(file)
			if err != nil {
				output.Error("reading context file: %v", err)
			}
			if err := json.Unmarshal(data, &ctxData); err != nil {
				output.Error("parsing context file: %v", err)
			}
		} else if inline, _ := cmd.Flags().GetString("context"); inline != "" {
			if err := json.Unmarshal([]byte(inline), &ctxData); err != nil {
				output.Error("parsing context JSON: %v", err)
			}
		}

		body := map[string]any{
			"sequence_id": seqID,
			"tenant_id":   flagTenantID,
		}
		if ctxData != nil {
			body["context"] = ctxData
		}
		if key, _ := cmd.Flags().GetString("idempotency-key"); key != "" {
			body["idempotency_key"] = key
		}

		inst, err := client.CreateInstance(ctx, body)
		if err != nil {
			output.Error("creating instance: %v", err)
		}
		if flagJSON {
			output.JSON(inst)
			return
		}
		fmt.Printf("Created instance: %s\n", inst.ID)
	},
}

var instSignalCmd = &cobra.Command{
	Use:   "signal <instance-id> <signal-type>",
	Short: "Send a signal to an instance",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()

		body := map[string]any{
			"signal_type": args[1],
		}
		if p, _ := cmd.Flags().GetString("payload"); p != "" {
			var payload any
			if err := json.Unmarshal([]byte(p), &payload); err != nil {
				output.Error("parsing payload: %v", err)
			}
			body["payload"] = payload
		}

		resp, err := client.SendSignal(ctx, args[0], body)
		if err != nil {
			output.Error("sending signal: %v", err)
		}
		if flagJSON {
			output.JSON(resp)
			return
		}
		fmt.Printf("Signal sent: %s -> %s\n", args[1], args[0])
	},
}

var instPauseCmd = &cobra.Command{
	Use:   "pause <id>",
	Short: "Pause an instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		if err := client.UpdateInstanceState(ctx, args[0], map[string]any{"state": "Paused"}); err != nil {
			output.Error("pausing instance: %v", err)
		}
		fmt.Println("Paused:", args[0])
	},
}

var instResumeCmd = &cobra.Command{
	Use:   "resume <id>",
	Short: "Resume a paused instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		if err := client.UpdateInstanceState(ctx, args[0], map[string]any{"state": "Pending"}); err != nil {
			output.Error("resuming instance: %v", err)
		}
		fmt.Println("Resumed:", args[0])
	},
}

var instCancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Cancel a running instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		if err := client.UpdateInstanceState(ctx, args[0], map[string]any{"state": "Cancelled"}); err != nil {
			output.Error("cancelling instance: %v", err)
		}
		fmt.Println("Cancelled:", args[0])
	},
}

var instRetryCmd = &cobra.Command{
	Use:   "retry <id>",
	Short: "Retry a failed instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		inst, err := client.RetryInstance(ctx, args[0])
		if err != nil {
			output.Error("retrying instance: %v", err)
		}
		if flagJSON {
			output.JSON(inst)
			return
		}
		fmt.Printf("Retried: %s (state: %s)\n", inst.ID, inst.State)
	},
}

var instOutputsCmd = &cobra.Command{
	Use:   "outputs <id>",
	Short: "Get step outputs for an instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		outputs, err := client.GetOutputs(ctx, args[0])
		if err != nil {
			output.Error("getting outputs: %v", err)
		}
		if flagJSON {
			output.JSON(outputs)
			return
		}
		headers := []string{"BLOCK", "ATTEMPT", "SIZE", "CREATED"}
		var rows [][]string
		for _, o := range outputs {
			rows = append(rows, []string{
				o.BlockID, fmt.Sprintf("%d", o.Attempt), fmt.Sprintf("%d", o.OutputSize), o.CreatedAt,
			})
		}
		output.Table(headers, rows)
	},
}

var instTreeCmd = &cobra.Command{
	Use:   "tree <id>",
	Short: "Get execution tree for an instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		tree, err := client.GetExecutionTree(ctx, args[0])
		if err != nil {
			output.Error("getting execution tree: %v", err)
		}
		if flagJSON {
			output.JSON(tree)
			return
		}
		headers := []string{"BLOCK", "TYPE", "STATE", "STARTED", "COMPLETED"}
		var rows [][]string
		for _, n := range tree {
			rows = append(rows, []string{
				n.BlockID, n.BlockType, n.State, n.StartedAt, n.CompletedAt,
			})
		}
		output.Table(headers, rows)
	},
}

var instAuditCmd = &cobra.Command{
	Use:   "audit <id>",
	Short: "Show audit log for an instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		entries, err := client.ListAuditLog(ctx, args[0])
		if err != nil {
			output.Error("getting audit log: %v", err)
		}
		if flagJSON {
			output.JSON(entries)
			return
		}
		headers := []string{"TIMESTAMP", "EVENT"}
		var rows [][]string
		for _, e := range entries {
			rows = append(rows, []string{e.Timestamp, e.Event})
		}
		output.Table(headers, rows)
	},
}

var instDLQCmd = &cobra.Command{
	Use:   "dlq",
	Short: "List dead letter queue instances",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		instances, err := client.ListDLQ(ctx, map[string]string{"tenant_id": flagTenantID})
		if err != nil {
			output.Error("listing DLQ: %v", err)
		}
		if flagJSON {
			output.JSON(instances)
			return
		}
		headers := []string{"ID", "SEQUENCE", "STATE", "UPDATED"}
		var rows [][]string
		for _, inst := range instances {
			rows = append(rows, []string{
				inst.ID, inst.SequenceID, inst.State, inst.UpdatedAt,
			})
		}
		output.Table(headers, rows)
	},
}
