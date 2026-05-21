package cmd

import (
	"encoding/json"
	"fmt"
	"os"

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
	sessionCmd.AddCommand(sessUpdateDataCmd)

	sessCreateCmd.Flags().String("key", "", "Session key (required)")
	sessCreateCmd.Flags().String("data", "", "Session data as JSON")
	sessCreateCmd.MarkFlagRequired("key")

	sessUpdateDataCmd.Flags().String("data", "", "Session data as JSON string")
	sessUpdateDataCmd.Flags().String("data-file", "", "Path to JSON file with session data")
}

var sessGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a session by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		sess, err := client.GetSession(ctx, args[0])
		if err != nil {
			return output.Errorf("getting session: %w", err)
		}
		if flagJSON {
			output.JSON(sess)
			return nil
		}
		fmt.Printf("ID:    %s\n", sess.ID)
		fmt.Printf("Key:   %s\n", sess.SessionKey)
		fmt.Printf("State: %s\n", sess.State)
		return nil
	},
}

var sessGetByKeyCmd = &cobra.Command{
	Use:   "get-by-key <key>",
	Short: "Look up a session by key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		sess, err := client.GetSessionByKey(ctx, flagTenantID, args[0])
		if err != nil {
			return output.Errorf("getting session by key: %w", err)
		}
		if flagJSON {
			output.JSON(sess)
			return nil
		}
		fmt.Printf("ID:    %s\n", sess.ID)
		fmt.Printf("Key:   %s\n", sess.SessionKey)
		fmt.Printf("State: %s\n", sess.State)
		return nil
	},
}

var sessCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a session",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		key, err := cmd.Flags().GetString("key")
		if err != nil {
			return output.Errorf("reading --key flag: %w", err)
		}
		body := map[string]any{
			"tenant_id":   flagTenantID,
			"session_key": key,
		}
		if d, err := cmd.Flags().GetString("data"); err != nil {
			return output.Errorf("reading --data flag: %w", err)
		} else if d != "" {
			var data any
			if err := json.Unmarshal([]byte(d), &data); err != nil {
				return output.Errorf("parsing data: %w", err)
			}
			body["data"] = data
		}
		sess, err := client.CreateSession(ctx, body)
		if err != nil {
			return output.Errorf("creating session: %w", err)
		}
		if flagJSON {
			output.JSON(sess)
			return nil
		}
		fmt.Printf("Created session: %s (key: %s)\n", sess.ID, sess.SessionKey)
		return nil
	},
}

var sessInstancesCmd = &cobra.Command{
	Use:   "instances <session-id>",
	Short: "List instances in a session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		instances, err := client.ListSessionInstances(ctx, args[0])
		if err != nil {
			return output.Errorf("listing session instances: %w", err)
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

var sessCloseCmd = &cobra.Command{
	Use:   "close <id>",
	Short: "Close a session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		if _, err := client.UpdateSessionState(ctx, args[0], map[string]any{"state": "closed"}); err != nil {
			return output.Errorf("closing session: %w", err)
		}
		fmt.Println("Closed:", args[0])
		return nil
	},
}

var sessUpdateDataCmd = &cobra.Command{
	Use:   "update-data <id>",
	Short: "Update session data",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()

		dataStr, _ := cmd.Flags().GetString("data")
		dataFile, _ := cmd.Flags().GetString("data-file")

		if dataStr == "" && dataFile == "" {
			return output.Errorf("either --data or --data-file is required")
		}
		if dataStr != "" && dataFile != "" {
			return output.Errorf("specify only one of --data or --data-file")
		}

		var raw []byte
		if dataFile != "" {
			var err error
			raw, err = os.ReadFile(dataFile)
			if err != nil {
				return output.Errorf("reading data file: %w", err)
			}
		} else {
			raw = []byte(dataStr)
		}

		var body any
		if err := json.Unmarshal(raw, &body); err != nil {
			return output.Errorf("parsing JSON: %w", err)
		}

		sess, err := client.UpdateSessionData(ctx, args[0], body)
		if err != nil {
			return output.Errorf("updating session data: %w", err)
		}
		if flagJSON {
			output.JSON(sess)
			return nil
		}
		fmt.Printf("Updated session: %s\n", sess.ID)
		return nil
	},
}
