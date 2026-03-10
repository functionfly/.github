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
	"time"
)

func main() {
	// Generate token using scripts/generate_token.go (run from repo root)
	cmd := exec.Command("go", "run", "./scripts/generate_token.go")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error generating token:", err)
		os.Exit(1)
	}
	token := strings.TrimSpace(string(output))
	fmt.Println("Token obtained:", token[:50], "...")

	// Python source code
	source := `import csv
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

	// Create publish request
	req := map[string]interface{}{
		"author":  "functionfly",
		"name":    "json-to-csv",
		"version": "1.0.5",
		"manifest": map[string]interface{}{
			"runtime": "python3.12",
		},
		"source": map[string]interface{}{
			"code":    source, // Send raw source code, not base64-encoded
			"runtime": "python3.12",
		},
	}

	reqJSON, _ := json.Marshal(req)
	fmt.Println("Request:", string(reqJSON)[:200], "...")

	// Publish
	publishReq, _ := http.NewRequest("POST", "http://localhost:8090/v1/registry/publish", bytes.NewReader(reqJSON))
	publishReq.Header.Set("Content-Type", "application/json")
	publishReq.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 120 * time.Second}
	publishResp, err := client.Do(publishReq)
	if err != nil {
		fmt.Println("Error publishing:", err)
		os.Exit(1)
	}
	defer publishResp.Body.Close()

	publishData, _ := io.ReadAll(publishResp.Body)
	fmt.Println("Publish response (status", publishResp.StatusCode, "):", string(publishData))

	if publishResp.StatusCode != 200 {
		fmt.Println("Publish failed")
		os.Exit(1)
	}

	// Now test execution
	fmt.Println("\n--- Testing execution ---")

	execReq := map[string]interface{}{
		"input": map[string]interface{}{
			"data": []map[string]interface{}{
				{"name": "Alice", "age": 30},
				{"name": "Bob", "age": 25},
			},
		},
	}
	execJSON, _ := json.Marshal(execReq)

	execReqHttp, _ := http.NewRequest("POST", "http://localhost:8090/v1/fx/functionfly/json-to-csv", bytes.NewReader(execJSON))
	execReqHttp.Header.Set("Content-Type", "application/json")
	execReqHttp.Header.Set("Authorization", "Bearer "+token)

	execResp, err := client.Do(execReqHttp)
	if err != nil {
		fmt.Println("Error executing:", err)
		os.Exit(1)
	}
	defer execResp.Body.Close()

	execData, _ := io.ReadAll(execResp.Body)
	fmt.Println("Execute response (status", execResp.StatusCode, "):", string(execData))
}
