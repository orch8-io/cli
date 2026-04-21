package cmd

import (
	"fmt"

	"github.com/orch8-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var sequenceCmd = &cobra.Command{
	Use:     "sequence",
	Aliases: []string{"seq"},
	Short:   "Manage sequence definitions",
}

func init() {
	sequenceCmd.AddCommand(seqGetCmd)
	sequenceCmd.AddCommand(seqGetByNameCmd)
	sequenceCmd.AddCommand(seqDeprecateCmd)
	sequenceCmd.AddCommand(seqVersionsCmd)
}

var seqGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a sequence by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		seq, err := client.GetSequence(ctx, args[0])
		if err != nil {
			return output.Errorf("getting sequence: %w", err)
		}
		if flagJSON {
			output.JSON(seq)
			return nil
		}
		fmt.Printf("ID:        %s\n", seq.ID)
		fmt.Printf("Name:      %s\n", seq.Name)
		fmt.Printf("Namespace: %s\n", seq.Namespace)
		fmt.Printf("Version:   %d\n", seq.Version)
		return nil
	},
}

var seqGetByNameCmd = &cobra.Command{
	Use:   "get-by-name <namespace> <name>",
	Short: "Look up a sequence by namespace and name",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		seq, err := client.GetSequenceByName(ctx, flagTenantID, args[0], args[1], nil)
		if err != nil {
			return output.Errorf("getting sequence by name: %w", err)
		}
		if flagJSON {
			output.JSON(seq)
			return nil
		}
		fmt.Printf("ID:        %s\n", seq.ID)
		fmt.Printf("Name:      %s\n", seq.Name)
		fmt.Printf("Namespace: %s\n", seq.Namespace)
		fmt.Printf("Version:   %d\n", seq.Version)
		return nil
	},
}

var seqDeprecateCmd = &cobra.Command{
	Use:   "deprecate <id>",
	Short: "Mark a sequence as deprecated",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		if err := client.DeprecateSequence(ctx, args[0]); err != nil {
			return output.Errorf("deprecating sequence: %w", err)
		}
		fmt.Println("Deprecated:", args[0])
		return nil
	},
}

var seqVersionsCmd = &cobra.Command{
	Use:   "versions <namespace> <name>",
	Short: "List all versions of a sequence",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx := cmd.Context()
		versions, err := client.ListSequenceVersions(ctx, flagTenantID, args[0], args[1])
		if err != nil {
			return output.Errorf("listing versions: %w", err)
		}
		if flagJSON {
			output.JSON(versions)
			return nil
		}
		headers := []string{"ID", "VERSION", "DEPRECATED", "CREATED"}
		var rows [][]string
		for _, v := range versions {
			rows = append(rows, []string{
				v.ID, fmt.Sprintf("%d", v.Version), fmt.Sprintf("%v", v.Deprecated), v.CreatedAt,
			})
		}
		output.Table(headers, rows)
		return nil
	},
}
