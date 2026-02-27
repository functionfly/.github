package main

import (
	"context"
	"fmt"
	"os"

	"github.com/functionfly/functionfly/internal/flypy/backend"
	"github.com/functionfly/functionfly/internal/flypy/compiler"
	"github.com/functionfly/functionfly/internal/flypy/ir"
	"github.com/functionfly/functionfly/internal/flypy/parser"
)

func testCSVToJSON() {
	// Test Python code - handles CSV string input and converts to JSON
	pythonCode := `
import csv
import io
import json

def handler(event):
    # Handle direct string input
    if isinstance(event, str):
        csv_data = event
    # Handle dict input with csv key
    elif isinstance(event, dict):
        csv_data = event.get("csv", "")
    else:
        return {"error": "Input must be a string or dict with csv key"}

    if not csv_data.strip():
        return {"json": [], "rows": 0}

    # Parse CSV
    try:
        reader = csv.DictReader(io.StringIO(csv_data))
        rows = []
        for row in reader:
            rows.append(row)
        return {"json": rows, "rows": len(rows)}
    except Exception as e:
        return {"error": f"CSV parsing error: {str(e)}"}
`

	fmt.Println("=== Python Source Code ===")
	fmt.Println(pythonCode)
	fmt.Println()

	// Step 1: Parse Python to AST
	ctx := context.Background()
	ast, err := parser.ParsePython(ctx, pythonCode)
	if err != nil {
		fmt.Printf("Error parsing Python: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== AST ===")
	fmt.Printf("%#v\n", ast)
	fmt.Println()

	// Step 2: Convert AST to IR
	functions := parser.GetFunctions(ast)
	if len(functions) == 0 {
		fmt.Println("No functions found in AST")
		os.Exit(1)
	}

	irModule, err := ir.Generate(ast, "complex")
	if err != nil {
		fmt.Printf("Error generating IR: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== IR Module ===")
	fmt.Printf("Functions: %d\n", len(irModule.Functions))
	if len(irModule.Functions) > 0 {
		fn := irModule.Functions[0]
		fmt.Printf("Function Name: %s\n", fn.Name)
		fmt.Printf("Parameters: %#v\n", fn.Parameters)
		fmt.Printf("Body Operations: %d\n", len(fn.Body))
		for i, op := range fn.Body {
			fmt.Printf("Op %d: Type=%s, Result=%s, Value=%#v\n", i, op.Type, op.Result, op.Value)
		}
	}
	fmt.Println()

	// Step 3: Generate Rust code
	rustCode, err := backend.GenerateRustWithMode(irModule, "complex")
	if err != nil {
		fmt.Printf("Error generating Rust code: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Generated Rust Code ===")
	fmt.Println(rustCode)
	fmt.Println()

	// Step 4: Try to compile Rust to WASM
	fmt.Println("=== Compiling to WASM ===")
	wasmBytes, err := compiler.CompileRustWithMode(rustCode, "wasm32-wasip1", "complex")
	if err != nil {
		fmt.Printf("Error compiling to WASM: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("WASM compiled successfully (size: %d bytes)\n", len(wasmBytes))

	// Save WASM to file for testing
	wasmPath := "/tmp/flypy-test/csv-to-json.wasm"
	if err := os.WriteFile(wasmPath, wasmBytes, 0644); err != nil {
		fmt.Printf("Error saving WASM: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("WASM saved to %s\n", wasmPath)
}
