package main

import (
	"context"
	"fmt"

	"github.com/functionfly/functionfly/internal/flypy/backend"
	"github.com/functionfly/functionfly/internal/flypy/ir"
	"github.com/functionfly/functionfly/internal/flypy/parser"
)

func main() {
	pythonCode := `
import csv
import io
import json

def handler(event):
    # Check if input is a list or dict with data field
    if isinstance(event, list):
        data = event
    elif isinstance(event, dict) and "data" in event:
        data = event["data"]
    else:
        return {"error": "Input must be a list or dict with 'data' field"}

    # Check if data is empty
    if len(data) == 0:
        return {"csv": "", "rows": 0}

    # Check if data is a list of dicts
    if not isinstance(data[0], dict):
        return {"error": "Data must be a list of dictionaries"}

    # Get fieldnames from first item
    fieldnames = list(data[0].keys())

    # Create CSV writer
    output = io.StringIO()
    writer = csv.DictWriter(output, fieldnames=fieldnames)
    writer.writeheader()

    # Write data
    for row in data:
        writer.writerow(row)

    return {"csv": output.getvalue(), "rows": len(data)}
`

	ctx := context.Background()
	ast, err := parser.ParsePython(ctx, pythonCode)
	if err != nil {
		fmt.Printf("Error parsing Python: %v\n", err)
		return
	}

	irModule, err := ir.Generate(ast, "complex")
	if err != nil {
		fmt.Printf("Error generating IR: %v\n", err)
		return
	}

	rustCode, err := backend.GenerateRustWithMode(irModule, "complex")
	if err != nil {
		fmt.Printf("Error generating Rust code: %v\n", err)
		return
	}

	fmt.Println("=== Generated Rust Code ===")
	fmt.Println(rustCode)
}
