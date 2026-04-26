// Package functionfly provides the Go SDK for building FunctionFly functions.
// Compile to WASM: GOOS=wasip1 GOARCH=wasm go build -o function.wasm .
package functionfly

import (
	"fmt"
	"os"
)

// Function is the interface that all FunctionFly functions must implement.
type Function interface {
	// Handle processes the input and returns the output.
	Handle(input string) (string, error)
}

// Context provides access to host functions during execution.
type Context struct {
	FunctionName string
	Version      string
}

// Run executes the given function. Call this from main().
func Run(fn Function) {
	input := readStdin()
	output, err := fn.Handle(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}

// readStdin reads all input from stdin.
func readStdin() string {
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 1024)
	for {
		n, err := os.Stdin.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
	}
	return string(buf)
}

// Log sends a message to the FunctionFly logging system.
func Log(msg string) {
	fmt.Fprintf(os.Stderr, "[functionfly] %s\n", msg)
}

// GetEnv retrieves an environment variable.
func GetEnv(key string) string {
	return os.Getenv(key)
}

// Error returns an error response in JSON format.
func Error(code, message string) string {
	return fmt.Sprintf(`{"error": {"code": "%s", "message": "%s"}}`, code, message)
}

// JSON returns a successful JSON response.
func JSON(data string) string {
	return fmt.Sprintf(`{"ok": true, "data": %s}`, data)
}
