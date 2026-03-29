// Package main implements the FunctionFly developer CLI.
package main

import (
	"github.com/functionfly/functionfly/cmd/fly/cmd"
	"github.com/functionfly/functionfly/cmd/fly/commands"
)

func main() {
	root := commands.NewRootCmd()
	// Attach backend, flypy, compile from the cmd package (admin excluded from public CLI)
	root.AddCommand(cmd.BackendCmd(), cmd.FlypyCmd(), cmd.CompileCmd())
	if err := root.Execute(); err != nil {
		commands.ExitOnError(err)
	}
}
