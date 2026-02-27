package main

import (
	"context"
	"fmt"
	"os"

	"github.com/functionfly/functionfly/internal/flypy/parser"
)

func testDebug() {
	// Test simpler Python code to understand IR generation
	testCases := []string{
		`def handler(event):
    if "data" in event:
        return event["data"]
`,
		`def handler(data):
    if len(data) == 0:
        return "empty"
`,
		`def handler(data):
    if not isinstance(data, list):
        return "not list"
`,
		`def handler():
    import io
    output = io.StringIO()
    return output.getvalue()
`,
	}

	for i, code := range testCases {
		fmt.Printf("\n=== Test Case %d ===\n", i+1)
		fmt.Println("Python Code:")
		fmt.Println(code)

		ctx := context.Background()
		ast, err := parser.ParsePython(ctx, code)
		if err != nil {
			fmt.Printf("Error parsing: %v\n", err)
			continue
		}

		functions := parser.GetFunctions(ast)
		if len(functions) > 0 {
			fn := functions[0]
			body := parser.GetFunctionBody(fn)
			fmt.Println("\nAST Body:")
			for j, stmt := range body {
				fmt.Printf("Statement %d: %#v\n", j, stmt)
			}
		}
	}
	os.Exit(0)
}
