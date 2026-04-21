package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/orch8-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage sessions",
}

func init() {
	sessionCmd.AddCommand(sessGetCmd)
	sessionCmd.AddCommand(sessGetByKeyCmd)
	sessionCmd.AddCommand(sessCreateCmd)
	sessionCmd.AddCommand(sessInstancesCmd)
	sessionCmd.AddCommand(sessCloseCmd)

	sessCreateCmd.Flags().String("key", "", "Session key (required)")
	sessCreateCmd.Flags().String("data", "", "Session data as JSON")
	sessCreateCmd.MarkFlagRequired("key")
}

var sessGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a session by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		sess, err := client.GetSession(ctx, args[0])
		if err != nil {
			output.Error("getting session: %v", err)
		}
		output.JSON(sess)
	},
}

var sessGetByKeyCmd = &cobra.Command{
	Use:   "get-by-key <key>",
	Short: "Look up a session by key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		sess, err := client.GetSessionByKey(ctx, flagTenantID, args[0])
		if err != nil {
			output.Error("getting session by key: %v", err)
		}
		output.JSON(sess)
	},
}

var sessCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a session",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		key, _ := cmd.Flags().GetString("key")
		body := map[string]any{
			"tenant_id":   flagTenantID,
			"session_key": key,
		}
		if d, _ := cmd.Flags().GetString("data"); d != "" {
			var data any
			if err := json.Unmarshal([]byte(d), &data); err != nil {
				output.Error("parsing data: %v", err)
			}
			body["data"] = data
		}
		sess, err := client.CreateSession(ctx, body)
		if err != nil {
			output.Error("creating session: %v", err)
		}
		if flagJSON {
			output.JSON(sess)
			return
		}
		fmt.Printf("Created session: %s (key: %s)\n", sess.ID, sess.SessionKey)
	},
}

var sessInstancesCmd = &cobra.Command{
	Use:   "instances <session-id>",
	Short: "List instances in a session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		instances, err := client.ListSessionInstances(ctx, args[0])
		if err != nil {
			output.Error("listing session instances: %v", err)
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

var sessCloseCmd = &cobra.Command{
	Use:   "close <id>",
	Short: "Close a session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		if _, err := client.UpdateSessionState(ctx, args[0], map[string]any{"state": "closed"}); err != nil {
			output.Error("closing session: %v", err)
		}
		fmt.Println("Closed:", args[0])
	},
}
