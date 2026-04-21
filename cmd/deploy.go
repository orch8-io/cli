package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/orch8-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy <file>",
	Short: "Deploy a sequence definition from a JSON file",
	Long:  "Reads a sequence JSON file and creates it via the API. The file should contain a valid sequence definition.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()

		data, err := os.ReadFile(args[0])
		if err != nil {
			return output.Errorf("reading file: %w", err)
		}

		var body map[string]any
		if err := json.Unmarshal(data, &body); err != nil {
			return output.Errorf("parsing JSON: %w", err)
		}

		// Set tenant if not in the file
		if _, ok := body["tenant_id"]; !ok {
			body["tenant_id"] = flagTenantID
		}

		seq, err := client.CreateSequence(ctx, body)
		if err != nil {
			return output.Errorf("deploying sequence: %w", err)
		}

		if flagJSON {
			output.JSON(seq)
			return nil
		}
		fmt.Printf("Deployed: %s (id: %s, version: %d)\n", seq.Name, seq.ID, seq.Version)
		return nil
	},
}
