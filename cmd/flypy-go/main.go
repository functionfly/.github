// Package main implements the FlyPy Python-to-WASM compiler CLI.
// This compiler transforms Python functions into deterministic WebAssembly modules
// that execute in the FunctionFly runtime without requiring a Python interpreter.
package main

import (
	"fmt"
	"os"
)

// Version is the compiler version
const Version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "compile":
		if err := runCompile(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "test-json-to-csv":
		testJSONToCSV()
	case "test-csv-to-json":
		testCSVToJSON()
	case "test-debug":
		testDebug()
	case "version":
		fmt.Printf("FlyPy Compiler v%s\n", Version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("FlyPy Compiler - Compile Python to WebAssembly")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  flypy-go <command> [options]")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  compile    Compile Python function to WebAssembly")
	fmt.Println("  version    Print compiler version")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Run 'flypy-go compile --help' for compile command options.")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  flypy-go compile --input handler.py --metadata meta.json --output ./dist")
	fmt.Println("  flypy-go compile -i handler.py -m meta.json -o ./dist --verbose")
	fmt.Println()
	fmt.Println("For more information, see: https://github.com/functionfly/functionfly")
}
