package commands

import (
	"fmt"

	"github.com/functionfly/functionfly/internal/version"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "ff",
		Short: "FunctionFly CLI — publish functions to the global edge",
		Long: `ff is the FunctionFly developer CLI.

Go from idea → global API in under 60 seconds.

  ff login              Authenticate with FunctionFly
  ff init <name>        Scaffold a new function project
  ff dev                Run function locally
  ff publish            Publish function to the registry
  ff test               Test your deployed function
  ff update <bump>      Bump function version
  ff stats              View usage statistics
  ff logs               Stream live execution logs
  ff rollback           Roll back to a previous version
  ff env                Manage environment variables
  ff secrets            Manage secrets
  ff vault              Manage zero-knowledge vault secrets
  ff whoami             Show current logged-in user
  ff logout             Clear stored credentials
  ff completion         Generate shell completion scripts`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add --version flag (Cobra's built-in version support)
	root.Version = version.Short()
	root.Flags().Bool("version", false, "Show fly CLI version")

	// Add persistent flags for debug/verbose/trace modes
	// These are available to all subcommands
	root.PersistentFlags().BoolVar(&DebugMode, "debug", false, "Enable full debug output")
	root.PersistentFlags().BoolVarP(&VerboseMode, "verbose", "v", false, "Enable verbose API calls")
	root.PersistentFlags().BoolVar(&TraceMode, "trace", false, "Enable HTTP trace with request/response bodies")

	// Set up persistent pre-run to handle debug mode
	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if DebugMode {
			Debug("Debug mode enabled")
			Debug("Version: %s", version.Info())
		}
	}

	// Add version command
	root.AddCommand(NewVersionCmd())

	root.AddCommand(
		NewLoginCmd(),
		NewWhoamiCmd(),
		NewLogoutCmd(),
		NewConfigCmd(),
		NewSelfUpdateCmd(),
		NewInitCmd(),
		NewDevCmd(),
		NewPublishCmd(),
		NewPublishBatchCmd(),
		NewManifestCmd(),
		NewTestCmd(),
		NewUpdateCmd(),
		NewStatsCmd(),
		NewLogsCmd(),
		NewRollbackCmd(),
		NewEnvCmd(),
		NewSecretsCmd(),
		NewVaultCmd(),
		NewScheduleCmd(),
		NewDreCmd(),
		NewCompletionCmd(root),
	)

	return root
}

// NewVersionCmd creates the version command.
func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the fly CLI version",
		Long: `Show version information for the fly CLI.

This displays the semantic version, git commit hash, and build date.
Use this to verify which version of fly you have installed.`,
		Example: `  fly version
  fly version --short  # Show only version number`,
		Run: func(cmd *cobra.Command, args []string) {
			short, _ := cmd.Flags().GetBool("short")
			if short {
				fmt.Println(version.Short())
			} else {
				PrintVersion()
			}
		},
	}

	cmd.Flags().Bool("short", false, "Show only version number")

	return cmd
}

// GetVersion returns the current CLI version string.
