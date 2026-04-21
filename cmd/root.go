package cmd

import (
	"fmt"
	"os"

	orch8 "github.com/orch8-io/sdk-go"
	"github.com/spf13/cobra"
)

var (
	flagURL      string
	flagTenantID string
	flagAPIKey   string
	flagJSON     bool
	flagVerbose  bool
	version      = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "orch8",
	Short: "CLI for Orch8 workflow engine",
	Long:  "Manage sequences, instances, cron schedules, triggers, and more from the command line.",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagURL, "url", envOr("ORCH8_URL", "http://localhost:8080"), "Engine API URL ($ORCH8_URL)")
	rootCmd.PersistentFlags().StringVar(&flagTenantID, "tenant", envOr("ORCH8_TENANT", "default"), "Tenant ID ($ORCH8_TENANT)")
	rootCmd.PersistentFlags().StringVar(&flagAPIKey, "api-key", os.Getenv("ORCH8_API_KEY"), "API key ($ORCH8_API_KEY)")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "Print request details")

	rootCmd.AddCommand(sequenceCmd)
	rootCmd.AddCommand(instanceCmd)
	rootCmd.AddCommand(cronCmd)
	rootCmd.AddCommand(triggerCmd)
	rootCmd.AddCommand(sessionCmd)
	rootCmd.AddCommand(pluginCmd)
	rootCmd.AddCommand(cbCmd)
	rootCmd.AddCommand(clusterCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(workerCmd)
	rootCmd.AddCommand(poolCmd)
	rootCmd.AddCommand(credentialCmd)
	rootCmd.AddCommand(approvalCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func newClient() *orch8.Client {
	cfg := orch8.ClientConfig{
		BaseURL:  flagURL,
		TenantID: flagTenantID,
	}
	if flagAPIKey != "" {
		cfg.Headers = map[string]string{
			"Authorization": "Bearer " + flagAPIKey,
		}
	}
	return orch8.NewClient(cfg)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print CLI version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "orch8", version)
		return nil
	},
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}
	},
}
