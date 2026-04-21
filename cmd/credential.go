package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/orch8-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var credentialCmd = &cobra.Command{
	Use:   "credential",
	Short: "Manage credentials",
}

func init() {
	credentialCmd.AddCommand(credListCmd)
	credentialCmd.AddCommand(credGetCmd)
	credentialCmd.AddCommand(credCreateCmd)
	credentialCmd.AddCommand(credDeleteCmd)

	credCreateCmd.Flags().String("file", "", "Path to credential JSON file")
}

var credListCmd = &cobra.Command{
	Use:   "list",
	Short: "List credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		creds, err := client.ListCredentials(ctx, flagTenantID)
		if err != nil {
			return output.Errorf("listing credentials: %w", err)
		}
		if flagJSON {
			output.JSON(creds)
			return nil
		}
		headers := []string{"ID", "NAME", "TYPE"}
		var rows [][]string
		for _, c := range creds {
			rows = append(rows, []string{
				c.ID, c.Name, c.CredentialType,
			})
		}
		output.Table(headers, rows)
		return nil
	},
}

var credGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a credential by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		c, err := client.GetCredential(ctx, args[0])
		if err != nil {
			return output.Errorf("getting credential: %w", err)
		}
		if flagJSON {
			output.JSON(c)
			return nil
		}
		fmt.Printf("ID:   %s\n", c.ID)
		fmt.Printf("Name: %s\n", c.Name)
		fmt.Printf("Type: %s\n", c.CredentialType)
		return nil
	},
}

var credCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a credential from JSON file",
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
		c, err := client.CreateCredential(ctx, body)
		if err != nil {
			return output.Errorf("creating credential: %w", err)
		}
		if flagJSON {
			output.JSON(c)
			return nil
		}
		fmt.Printf("Created credential: %s\n", c.ID)
		return nil
	},
}

var credDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a credential",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		if err := client.DeleteCredential(ctx, args[0]); err != nil {
			return output.Errorf("deleting credential: %w", err)
		}
		fmt.Println("Deleted:", args[0])
		return nil
	},
}
