package cmd

import (
	"context"
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
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()

		data, err := os.ReadFile(args[0])
		if err != nil {
			output.Error("reading file: %v", err)
		}

		var body map[string]any
		if err := json.Unmarshal(data, &body); err != nil {
			output.Error("parsing JSON: %v", err)
		}

		// Set tenant if not in the file
		if _, ok := body["tenant_id"]; !ok {
			body["tenant_id"] = flagTenantID
		}

		seq, err := client.CreateSequence(ctx, body)
		if err != nil {
			output.Error("deploying sequence: %v", err)
		}

		if flagJSON {
			output.JSON(seq)
			return
		}
		fmt.Printf("Deployed: %s (id: %s, version: %d)\n", seq.Name, seq.ID, seq.Version)
	},
}
