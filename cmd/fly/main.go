// Package main implements the FunctionFly developer CLI.
package main

import (
	"os"

	"github.com/functionfly/functionfly/cmd/fly/commands"
)

func main() {
	root := commands.NewRootCmd()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
