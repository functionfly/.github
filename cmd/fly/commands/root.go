package commands

import (
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

	root.AddCommand(
		NewLoginCmd(),
		NewWhoamiCmd(),
		NewLogoutCmd(),
		NewInitCmd(),
		NewDevCmd(),
		NewPublishCmd(),
		NewTestCmd(),
		NewUpdateCmd(),
		NewStatsCmd(),
		NewLogsCmd(),
		NewRollbackCmd(),
		NewEnvCmd(),
		NewSecretsCmd(),
		NewCompletionCmd(root),
	)

	return root
}
