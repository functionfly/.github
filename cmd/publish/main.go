package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type FunctionSource struct {
	Code       string            `json:"code"`
	Files      map[string]string `json:"files,omitempty"`
	Runtime    string            `json:"runtime"`
	WasmBinary string            `json:"wasm_binary,omitempty"`
}

type PublishRequest struct {
	Author     string          `json:"author"`
	Name       string          `json:"name"`
	Version    string          `json:"version"`
	Manifest   json.RawMessage `json:"manifest"`
	Source     *FunctionSource `json:"source"`
	TrustLevel string          `json:"trust_level,omitempty"`
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run cmd/publish/main.go <wasm_file> <function_name> <jwt_token>")
		os.Exit(1)
	}

	wasmFile := os.Args[1]
	funcName := os.Args[2]
	token := os.Args[3]

	// Read WASM file
	wasmData, err := os.ReadFile(wasmFile)
	if err != nil {
		fmt.Printf("Error reading WASM file: %v\n", err)
		os.Exit(1)
	}

	// Read Python source
	pythonSource, err := os.ReadFile("functions/functionfly/" + funcName + "/main.py")
	if err != nil {
		fmt.Printf("Error reading Python source: %v\n", err)
		os.Exit(1)
	}

	// Create manifest
	manifest := map[string]interface{}{
		"name":          funcName,
		"version":       "1.0.1",
		"title":         "JSON to CSV Converter",
		"description":   "Convert a JSON array of objects to CSV format.",
		"author":        "functionfly",
		"runtime":       "wasm",
		"category":      "data-formatting",
		"tags":          []string{"json", "csv", "convert", "data"},
		"verified":      true,
		"system_owned":  true,
		"pricing_model": "free",
		"input_schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"data":      map[string]interface{}{"type": "array"},
				"columns":   map[string]interface{}{"type": "array"},
				"delimiter": map[string]interface{}{"type": "string", "default": ","},
			},
			"required": []string{"data"},
		},
		"output_schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"ok":        map[string]interface{}{"type": "boolean"},
				"csv":       map[string]interface{}{"type": "string"},
				"row_count": map[string]interface{}{"type": "number"},
			},
		},
		"example_input":  map[string]interface{}{"data": []interface{}{map[string]interface{}{"name": "John", "age": 30}}},
		"example_output": map[string]interface{}{"ok": true, "csv": "name,age\nJohn,30\n", "row_count": 1},
	}

	manifestJSON, _ := json.Marshal(manifest)

	// Create publish request with pre-compiled WASM
	req := PublishRequest{
		Author:   "functionfly",
		Name:     funcName,
		Version:  "1.0.1",
		Manifest: manifestJSON,
		Source: &FunctionSource{
			Code:       string(pythonSource),
			Runtime:    "wasm",
			WasmBinary: base64.StdEncoding.EncodeToString(wasmData),
		},
		TrustLevel: "high",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	// Make HTTP request
	url := "http://localhost:8090/v1/registry/publish"
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Pretty print response
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, body, "", "  ")
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response:\n%s\n", prettyJSON.String())
}
