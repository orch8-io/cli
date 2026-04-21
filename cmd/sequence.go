package cmd

import (
	"context"
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
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		seq, err := client.GetSequence(ctx, args[0])
		if err != nil {
			output.Error("getting sequence: %v", err)
		}
		output.JSON(seq)
	},
}

var seqGetByNameCmd = &cobra.Command{
	Use:   "get-by-name <namespace> <name>",
	Short: "Look up a sequence by namespace and name",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		seq, err := client.GetSequenceByName(ctx, flagTenantID, args[0], args[1], nil)
		if err != nil {
			output.Error("getting sequence by name: %v", err)
		}
		output.JSON(seq)
	},
}

var seqDeprecateCmd = &cobra.Command{
	Use:   "deprecate <id>",
	Short: "Mark a sequence as deprecated",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		if err := client.DeprecateSequence(ctx, args[0]); err != nil {
			output.Error("deprecating sequence: %v", err)
		}
		fmt.Println("Deprecated:", args[0])
	},
}

var seqVersionsCmd = &cobra.Command{
	Use:   "versions <namespace> <name>",
	Short: "List all versions of a sequence",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		ctx := context.Background()
		versions, err := client.ListSequenceVersions(ctx, flagTenantID, args[0], args[1])
		if err != nil {
			output.Error("listing versions: %v", err)
		}
		if flagJSON {
			output.JSON(versions)
			return
		}
		headers := []string{"ID", "VERSION", "DEPRECATED", "CREATED"}
		var rows [][]string
		for _, v := range versions {
			rows = append(rows, []string{
				v.ID, fmt.Sprintf("%d", v.Version), fmt.Sprintf("%v", v.Deprecated), v.CreatedAt,
			})
		}
		output.Table(headers, rows)
	},
}
