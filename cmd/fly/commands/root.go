package commands

import (
	"fmt"
	"os"

	"github.com/functionfly/functionfly/internal/version"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "fly",
		Short: "FunctionFly CLI — publish functions to the global edge",
		Long: `fly is the FunctionFly developer CLI.

Go from idea → global API in under 60 seconds.

  fly login              Authenticate with FunctionFly
  fly init <name>        Scaffold a new function project
  fly dev                Run function locally
  fly publish            Publish function to the registry
  fly test               Test your deployed function
  fly update <bump>      Bump function version
  fly stats              View usage statistics
  fly logs               Stream live execution logs
  fly rollback           Roll back to a previous version
  fly env                Manage environment variables
  fly secrets            Manage secrets
  fly whoami             Show current logged-in user
  fly logout             Clear stored credentials
  fly completion         Generate shell completion scripts`,
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

// handleError handles errors from command execution.
// It prints the error message and exits with the appropriate code.
func handleError(err error) {
	if err == nil {
		return
	}

	// In debug mode, also print stack trace or additional context
	if DebugMode {
		fmt.Fprintf(os.Stderr, "[DEBUG] Error: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	}

	// Get the appropriate exit code
	code := GetExitCode(err)
	os.Exit(code)
}
