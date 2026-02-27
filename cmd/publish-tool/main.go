package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Generate token using generate_token.go
	cmd := exec.Command("go", "run", "generate_token.go")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error generating token: %v\n", err)
		os.Exit(1)
	}
	token := strings.TrimSpace(string(output))
	fmt.Printf("Token obtained: %s...\n", token[:50])

	// Read the Python source
	source := `import csv
import io
import json

def handler(event):
    # Handle direct list input
    if isinstance(event, list):
        data = event
    # Handle dict input with data key
    elif isinstance(event, dict):
        data = event.get("data", [])
    else:
        return {"error": "Input must be a list or dict with data key"}

    if not data:
        return {"csv": "", "rows": 0}

    # Get field names from first object
    if isinstance(data[0], dict):
        fieldnames = list(data[0].keys())
    else:
        return {"error": "Array items must be objects"}

    # Create CSV
    output = io.StringIO()
    writer = csv.DictWriter(output, fieldnames=fieldnames)
    writer.writeheader()
    for row in data:
        writer.writerow(row)

    return {"csv": output.getvalue(), "rows": len(data)}`

	// Create manifest - with properties field
	inputProps, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"type":        "array",
			"description": "Array of objects to convert to CSV",
		},
	})
	outputProps, _ := json.Marshal(map[string]interface{}{
		"csv": map[string]interface{}{
			"type": "string",
		},
		"rows": map[string]interface{}{
			"type": "number",
		},
	})

	manifest := map[string]interface{}{
		"name":        "json-to-csv",
		"version":     "1.0.0",
		"runtime":     "python3.12",
		"title":       "JSON to CSV Converter",
		"description": "Convert a JSON array of objects to CSV format",
		"category":    "data-formatting",
		"tags":        []string{"json", "csv", "convert"},
		"input": map[string]interface{}{
			"type":       "object",
			"properties": json.RawMessage(inputProps),
		},
		"output": map[string]interface{}{
			"type":       "object",
			"properties": json.RawMessage(outputProps),
		},
		"timeout_ms": 5000,
		"memory_mb":  128,
	}

	manifestJSON, _ := json.Marshal(manifest)

	// Create publish request
	req := map[string]interface{}{
		"author":   "functionfly",
		"name":     "json-to-csv",
		"version":  "1.0.0",
		"manifest": json.RawMessage(manifestJSON),
		"source": map[string]interface{}{
			"code":    source,
			"runtime": "python3.12",
		},
	}

	reqJSON, _ := json.Marshal(req)

	// Publish
	httpReq, _ := http.NewRequest("POST", "http://localhost:8090/v1/registry/publish", bytes.NewReader(reqJSON))
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		fmt.Printf("Error publishing: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))
}
