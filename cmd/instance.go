package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

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
	instanceCmd.AddCommand(instUpdateContextCmd)
	instanceCmd.AddCommand(instBatchCreateCmd)
	instanceCmd.AddCommand(instCheckpointsCmd)
	instanceCmd.AddCommand(instCheckpointSaveCmd)
	instanceCmd.AddCommand(instCheckpointLatestCmd)
	instanceCmd.AddCommand(instCheckpointPruneCmd)
	instanceCmd.AddCommand(instInjectBlocksCmd)
	instanceCmd.AddCommand(instBulkStateCmd)
	instanceCmd.AddCommand(instBulkRescheduleCmd)
	instanceCmd.AddCommand(instStreamCmd)

	instListCmd.Flags().String("sequence", "", "Filter by sequence ID")
	instListCmd.Flags().String("state", "", "Filter by state")
	instListCmd.Flags().Int("limit", 50, "Max results")

	instCreateCmd.Flags().String("sequence", "", "Sequence ID (required)")
	instCreateCmd.Flags().String("context-file", "", "Path to context JSON file")
	instCreateCmd.Flags().String("context", "", "Inline context JSON")
	instCreateCmd.Flags().String("idempotency-key", "", "Idempotency key")
	instCreateCmd.MarkFlagRequired("sequence")

	instSignalCmd.Flags().String("payload", "", "Signal payload as JSON")

	instUpdateContextCmd.Flags().String("context", "", "Inline context JSON")
	instUpdateContextCmd.Flags().String("context-file", "", "Path to context JSON file")

	instBatchCreateCmd.Flags().String("file", "", "Path to JSON file with array of instances (required)")
	instBatchCreateCmd.MarkFlagRequired("file")

	instCheckpointSaveCmd.Flags().String("data", "", "Checkpoint data as inline JSON")
	instCheckpointSaveCmd.Flags().String("file", "", "Path to checkpoint data JSON file")

	instCheckpointPruneCmd.Flags().Int("keep", 3, "Number of latest checkpoints to keep")

	instInjectBlocksCmd.Flags().String("file", "", "Path to blocks JSON file (required)")
	instInjectBlocksCmd.MarkFlagRequired("file")

	instBulkStateCmd.Flags().String("state", "", "New state to set (required)")
	instBulkStateCmd.Flags().String("filter-tenant", "", "Filter by tenant ID")
	instBulkStateCmd.Flags().String("filter-namespace", "", "Filter by namespace")
	instBulkStateCmd.Flags().String("filter-states", "", "Filter by current states (comma-separated)")
	instBulkStateCmd.MarkFlagRequired("state")

	instBulkRescheduleCmd.Flags().Int("offset", 0, "Offset in seconds (required)")
	instBulkRescheduleCmd.Flags().String("filter-tenant", "", "Filter by tenant ID")
	instBulkRescheduleCmd.Flags().String("filter-namespace", "", "Filter by namespace")
	instBulkRescheduleCmd.Flags().String("filter-states", "", "Filter by current states (comma-separated)")
	instBulkRescheduleCmd.MarkFlagRequired("offset")

	instStreamCmd.Flags().Int("poll-ms", 0, "Polling interval in milliseconds")
}

var instListCmd = &cobra.Command{
	Use:   "list",
	Short: "List instances",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		filters := map[string]string{"tenant_id": flagTenantID}

		if v, err := cmd.Flags().GetString("sequence"); err != nil {
			return output.Errorf("reading --sequence flag: %w", err)
		} else if v != "" {
			filters["sequence_id"] = v
		}

		if v, err := cmd.Flags().GetString("state"); err != nil {
			return output.Errorf("reading --state flag: %w", err)
		} else if v != "" {
			filters["state"] = v
		}

		if limit, err := cmd.Flags().GetInt("limit"); err != nil {
			return output.Errorf("reading --limit flag: %w", err)
		} else if limit > 0 {
			filters["limit"] = strconv.Itoa(limit)
		}

		instances, err := client.ListInstances(ctx, filters)
		if err != nil {
			return output.Errorf("listing instances: %w", err)
		}
		if flagJSON {
			output.JSON(instances)
			return nil
		}
		headers := []string{"ID", "SEQUENCE", "STATE", "CREATED"}
		var rows [][]string
		for _, inst := range instances {
			rows = append(rows, []string{
				inst.ID, inst.SequenceID, inst.State, inst.CreatedAt,
			})
		}
		output.Table(headers, rows)
		return nil
	},
}

var instGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get instance details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		inst, err := client.GetInstance(ctx, args[0])
		if err != nil {
			return output.Errorf("getting instance: %w", err)
		}
		if flagJSON {
			output.JSON(inst)
			return nil
		}
		fmt.Printf("ID:       %s\n", inst.ID)
		fmt.Printf("Sequence: %s\n", inst.SequenceID)
		fmt.Printf("State:    %s\n", inst.State)
		fmt.Printf("Created:  %s\n", inst.CreatedAt)
		return nil
	},
}

var instCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		seqID, err := cmd.Flags().GetString("sequence")
		if err != nil {
			return output.Errorf("reading --sequence flag: %w", err)
		}

		var ctxData any
		if file, err := cmd.Flags().GetString("context-file"); err != nil {
			return output.Errorf("reading --context-file flag: %w", err)
		} else if file != "" {
			data, err := os.ReadFile(file)
			if err != nil {
				return output.Errorf("reading context file: %w", err)
			}
			if err := json.Unmarshal(data, &ctxData); err != nil {
				return output.Errorf("parsing context file: %w", err)
			}
		} else if inline, err := cmd.Flags().GetString("context"); err != nil {
			return output.Errorf("reading --context flag: %w", err)
		} else if inline != "" {
			if err := json.Unmarshal([]byte(inline), &ctxData); err != nil {
				return output.Errorf("parsing context JSON: %w", err)
			}
		}

		body := map[string]any{
			"sequence_id": seqID,
			"tenant_id":   flagTenantID,
		}
		if ctxData != nil {
			body["context"] = ctxData
		}
		if key, err := cmd.Flags().GetString("idempotency-key"); err != nil {
			return output.Errorf("reading --idempotency-key flag: %w", err)
		} else if key != "" {
			body["idempotency_key"] = key
		}

		inst, err := client.CreateInstance(ctx, body)
		if err != nil {
			return output.Errorf("creating instance: %w", err)
		}
		if flagJSON {
			output.JSON(inst)
			return nil
		}
		fmt.Printf("Created instance: %s\n", inst.ID)
		return nil
	},
}

var instSignalCmd = &cobra.Command{
	Use:   "signal <instance-id> <signal-type>",
	Short: "Send a signal to an instance",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()

		body := map[string]any{
			"signal_type": args[1],
		}
		if p, err := cmd.Flags().GetString("payload"); err != nil {
			return output.Errorf("reading --payload flag: %w", err)
		} else if p != "" {
			var payload any
			if err := json.Unmarshal([]byte(p), &payload); err != nil {
				return output.Errorf("parsing payload: %w", err)
			}
			body["payload"] = payload
		}

		resp, err := client.SendSignal(ctx, args[0], body)
		if err != nil {
			return output.Errorf("sending signal: %w", err)
		}
		if flagJSON {
			output.JSON(resp)
			return nil
		}
		fmt.Printf("Signal sent: %s -> %s\n", args[1], args[0])
		return nil
	},
}

var instPauseCmd = &cobra.Command{
	Use:   "pause <id>",
	Short: "Pause an instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		if err := client.UpdateInstanceState(ctx, args[0], map[string]any{"state": "Paused"}); err != nil {
			return output.Errorf("pausing instance: %w", err)
		}
		fmt.Println("Paused:", args[0])
		return nil
	},
}

var instResumeCmd = &cobra.Command{
	Use:   "resume <id>",
	Short: "Resume a paused instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		if err := client.UpdateInstanceState(ctx, args[0], map[string]any{"state": "Pending"}); err != nil {
			return output.Errorf("resuming instance: %w", err)
		}
		fmt.Println("Resumed:", args[0])
		return nil
	},
}

var instCancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Cancel a running instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		if err := client.UpdateInstanceState(ctx, args[0], map[string]any{"state": "Cancelled"}); err != nil {
			return output.Errorf("cancelling instance: %w", err)
		}
		fmt.Println("Cancelled:", args[0])
		return nil
	},
}

var instRetryCmd = &cobra.Command{
	Use:   "retry <id>",
	Short: "Retry a failed instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		inst, err := client.RetryInstance(ctx, args[0])
		if err != nil {
			return output.Errorf("retrying instance: %w", err)
		}
		if flagJSON {
			output.JSON(inst)
			return nil
		}
		fmt.Printf("Retried: %s (state: %s)\n", inst.ID, inst.State)
		return nil
	},
}

var instOutputsCmd = &cobra.Command{
	Use:   "outputs <id>",
	Short: "Get step outputs for an instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		outputs, err := client.GetOutputs(ctx, args[0])
		if err != nil {
			return output.Errorf("getting outputs: %w", err)
		}
		if flagJSON {
			output.JSON(outputs)
			return nil
		}
		headers := []string{"BLOCK", "ATTEMPT", "SIZE", "CREATED"}
		var rows [][]string
		for _, o := range outputs {
			rows = append(rows, []string{
				o.BlockID, fmt.Sprintf("%d", o.Attempt), fmt.Sprintf("%d", o.OutputSize), o.CreatedAt,
			})
		}
		output.Table(headers, rows)
		return nil
	},
}

var instTreeCmd = &cobra.Command{
	Use:   "tree <id>",
	Short: "Get execution tree for an instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		tree, err := client.GetExecutionTree(ctx, args[0])
		if err != nil {
			return output.Errorf("getting execution tree: %w", err)
		}
		if flagJSON {
			output.JSON(tree)
			return nil
		}
		headers := []string{"BLOCK", "TYPE", "STATE", "STARTED", "COMPLETED"}
		var rows [][]string
		for _, n := range tree {
			rows = append(rows, []string{
				n.BlockID, n.BlockType, n.State, n.StartedAt, n.CompletedAt,
			})
		}
		output.Table(headers, rows)
		return nil
	},
}

var instAuditCmd = &cobra.Command{
	Use:   "audit <id>",
	Short: "Show audit log for an instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		entries, err := client.ListAuditLog(ctx, args[0])
		if err != nil {
			return output.Errorf("getting audit log: %w", err)
		}
		if flagJSON {
			output.JSON(entries)
			return nil
		}
		headers := []string{"TIMESTAMP", "EVENT"}
		var rows [][]string
		for _, e := range entries {
			rows = append(rows, []string{e.Timestamp, e.Event})
		}
		output.Table(headers, rows)
		return nil
	},
}

var instDLQCmd = &cobra.Command{
	Use:   "dlq",
	Short: "List dead letter queue instances",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		instances, err := client.ListDLQ(ctx, map[string]string{"tenant_id": flagTenantID})
		if err != nil {
			return output.Errorf("listing DLQ: %w", err)
		}
		if flagJSON {
			output.JSON(instances)
			return nil
		}
		headers := []string{"ID", "SEQUENCE", "STATE", "UPDATED"}
		var rows [][]string
		for _, inst := range instances {
			rows = append(rows, []string{
				inst.ID, inst.SequenceID, inst.State, inst.UpdatedAt,
			})
		}
		output.Table(headers, rows)
		return nil
	},
}

var instUpdateContextCmd = &cobra.Command{
	Use:   "update-context <id>",
	Short: "Update the context of an instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()

		var ctxData any
		if file, err := cmd.Flags().GetString("context-file"); err != nil {
			return output.Errorf("reading --context-file flag: %w", err)
		} else if file != "" {
			data, err := os.ReadFile(file)
			if err != nil {
				return output.Errorf("reading context file: %w", err)
			}
			if err := json.Unmarshal(data, &ctxData); err != nil {
				return output.Errorf("parsing context file: %w", err)
			}
		} else if inline, err := cmd.Flags().GetString("context"); err != nil {
			return output.Errorf("reading --context flag: %w", err)
		} else if inline != "" {
			if err := json.Unmarshal([]byte(inline), &ctxData); err != nil {
				return output.Errorf("parsing context JSON: %w", err)
			}
		}

		if ctxData == nil {
			return output.Errorf("provide --context or --context-file")
		}

		body := map[string]any{"context": ctxData}
		if err := client.UpdateInstanceContext(ctx, args[0], body); err != nil {
			return output.Errorf("updating instance context: %w", err)
		}
		fmt.Println("Context updated:", args[0])
		return nil
	},
}

var instBatchCreateCmd = &cobra.Command{
	Use:   "batch-create",
	Short: "Create multiple instances from a JSON file",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()

		file, err := cmd.Flags().GetString("file")
		if err != nil {
			return output.Errorf("reading --file flag: %w", err)
		}

		data, err := os.ReadFile(file)
		if err != nil {
			return output.Errorf("reading file: %w", err)
		}
		var body any
		if err := json.Unmarshal(data, &body); err != nil {
			return output.Errorf("parsing JSON file: %w", err)
		}

		resp, err := client.BatchCreateInstances(ctx, body)
		if err != nil {
			return output.Errorf("batch creating instances: %w", err)
		}
		if flagJSON {
			output.JSON(resp)
			return nil
		}
		fmt.Printf("Created: %d instances\n", resp.Created)
		return nil
	},
}

var instCheckpointsCmd = &cobra.Command{
	Use:   "checkpoints <id>",
	Short: "List checkpoints for an instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()

		checkpoints, err := client.ListCheckpoints(ctx, args[0])
		if err != nil {
			return output.Errorf("listing checkpoints: %w", err)
		}
		if flagJSON {
			output.JSON(checkpoints)
			return nil
		}
		headers := []string{"ID", "CREATED"}
		var rows [][]string
		for _, cp := range checkpoints {
			rows = append(rows, []string{cp.ID, cp.CreatedAt})
		}
		output.Table(headers, rows)
		return nil
	},
}

var instCheckpointSaveCmd = &cobra.Command{
	Use:   "checkpoint-save <id>",
	Short: "Save a checkpoint for an instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()

		var cpData any
		if file, err := cmd.Flags().GetString("file"); err != nil {
			return output.Errorf("reading --file flag: %w", err)
		} else if file != "" {
			data, err := os.ReadFile(file)
			if err != nil {
				return output.Errorf("reading file: %w", err)
			}
			if err := json.Unmarshal(data, &cpData); err != nil {
				return output.Errorf("parsing file: %w", err)
			}
		} else if inline, err := cmd.Flags().GetString("data"); err != nil {
			return output.Errorf("reading --data flag: %w", err)
		} else if inline != "" {
			if err := json.Unmarshal([]byte(inline), &cpData); err != nil {
				return output.Errorf("parsing data JSON: %w", err)
			}
		}

		if cpData == nil {
			return output.Errorf("provide --data or --file")
		}

		cp, err := client.SaveCheckpoint(ctx, args[0], cpData)
		if err != nil {
			return output.Errorf("saving checkpoint: %w", err)
		}
		if flagJSON {
			output.JSON(cp)
			return nil
		}
		fmt.Println(cp.ID)
		return nil
	},
}

var instCheckpointLatestCmd = &cobra.Command{
	Use:   "checkpoint-latest <id>",
	Short: "Get the latest checkpoint for an instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()

		cp, err := client.GetLatestCheckpoint(ctx, args[0])
		if err != nil {
			return output.Errorf("getting latest checkpoint: %w", err)
		}
		output.JSON(cp)
		return nil
	},
}

var instCheckpointPruneCmd = &cobra.Command{
	Use:   "checkpoint-prune <id>",
	Short: "Prune old checkpoints, keeping the latest N",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()

		keep, err := cmd.Flags().GetInt("keep")
		if err != nil {
			return output.Errorf("reading --keep flag: %w", err)
		}

		if err := client.PruneCheckpoints(ctx, args[0], &keep); err != nil {
			return output.Errorf("pruning checkpoints: %w", err)
		}
		fmt.Printf("Pruned checkpoints, kept last %d\n", keep)
		return nil
	},
}

var instInjectBlocksCmd = &cobra.Command{
	Use:   "inject-blocks <id>",
	Short: "Inject blocks into an instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()

		file, err := cmd.Flags().GetString("file")
		if err != nil {
			return output.Errorf("reading --file flag: %w", err)
		}

		data, err := os.ReadFile(file)
		if err != nil {
			return output.Errorf("reading file: %w", err)
		}
		var body any
		if err := json.Unmarshal(data, &body); err != nil {
			return output.Errorf("parsing JSON file: %w", err)
		}

		if err := client.InjectBlocks(ctx, args[0], body); err != nil {
			return output.Errorf("injecting blocks: %w", err)
		}
		fmt.Println("Blocks injected")
		return nil
	},
}

func parseBulkFilter(cmd *cobra.Command) (map[string]any, error) {
	filter := map[string]any{}

	if v, err := cmd.Flags().GetString("filter-tenant"); err != nil {
		return nil, output.Errorf("reading --filter-tenant flag: %w", err)
	} else if v != "" {
		filter["tenant_id"] = v
	}

	if v, err := cmd.Flags().GetString("filter-namespace"); err != nil {
		return nil, output.Errorf("reading --filter-namespace flag: %w", err)
	} else if v != "" {
		filter["namespace"] = v
	}

	if v, err := cmd.Flags().GetString("filter-states"); err != nil {
		return nil, output.Errorf("reading --filter-states flag: %w", err)
	} else if v != "" {
		filter["states"] = strings.Split(v, ",")
	}

	return filter, nil
}

var instBulkStateCmd = &cobra.Command{
	Use:   "bulk-state",
	Short: "Bulk update instance states matching a filter",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()

		newState, err := cmd.Flags().GetString("state")
		if err != nil {
			return output.Errorf("reading --state flag: %w", err)
		}

		filter, err := parseBulkFilter(cmd)
		if err != nil {
			return err
		}

		resp, err := client.BulkUpdateState(ctx, filter, newState)
		if err != nil {
			return output.Errorf("bulk updating state: %w", err)
		}
		if flagJSON {
			output.JSON(resp)
			return nil
		}
		fmt.Printf("Updated: %d instances\n", resp.Updated)
		return nil
	},
}

var instBulkRescheduleCmd = &cobra.Command{
	Use:   "bulk-reschedule",
	Short: "Bulk reschedule instances matching a filter",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()

		offset, err := cmd.Flags().GetInt("offset")
		if err != nil {
			return output.Errorf("reading --offset flag: %w", err)
		}

		filter, err := parseBulkFilter(cmd)
		if err != nil {
			return err
		}

		resp, err := client.BulkReschedule(ctx, filter, offset)
		if err != nil {
			return output.Errorf("bulk rescheduling: %w", err)
		}
		if flagJSON {
			output.JSON(resp)
			return nil
		}
		fmt.Printf("Rescheduled: %d instances\n", resp.Updated)
		return nil
	},
}

var instStreamCmd = &cobra.Command{
	Use:   "stream <id>",
	Short: "Stream SSE events for an instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()

		pollMs, err := cmd.Flags().GetInt("poll-ms")
		if err != nil {
			return output.Errorf("reading --poll-ms flag: %w", err)
		}

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		events, errs := client.StreamInstance(ctx, args[0], pollMs)
		enc := json.NewEncoder(os.Stdout)
		for {
			select {
			case evt, ok := <-events:
				if !ok {
					return nil
				}
				if err := enc.Encode(evt); err != nil {
					return output.Errorf("encoding event: %w", err)
				}
			case err, ok := <-errs:
				if !ok {
					return nil
				}
				return output.Errorf("stream error: %w", err)
			}
		}
	},
}
