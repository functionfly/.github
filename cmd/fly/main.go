// Package main implements the FunctionFly developer CLI.
package main

import (
	"github.com/functionfly/functionfly/cmd/fly/cmd"
	"github.com/functionfly/functionfly/cmd/fly/commands"
)

func main() {
	root := commands.NewRootCmd()
	// Attach backend, admin, flypy from the cmd package (no import cycle: main imports both)
	root.AddCommand(cmd.BackendCmd(), cmd.AdminCmd(), cmd.FlypyCmd(), cmd.CompileCmd())
	if err := root.Execute(); err != nil {
		commands.ExitOnError(err)
	}
}
