/*
Copyright © 2026 FunctionFly
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// compileCmd represents the compile command
var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile functions to various formats",
	Long: `Compile functions to different output formats.

Supported compilers:
  python    Compile Python functions to WebAssembly (WASM)`,
	SilenceUsage: true,
}

func init() {
	// Add compile command to root
	rootCmd.AddCommand(compileCmd)

	// Add python subcommand (wraps flypy-go)
	compileCmd.AddCommand(newCompilePythonCmd())
}
