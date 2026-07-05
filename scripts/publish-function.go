//go:build ignore

// Publish a single function from ./functions/<author>/<name>/ to the local
// orchestrator as the authenticated user. Uses the active /v1/fx/publish
// endpoint (the legacy /v1/functions/publish route used by
// scripts/publish-all-functions.go no longer exists).
//
// Usage:
//   TOKEN=... [API=http://localhost:8080] go run scripts/publish-function.go <author> <name>
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const Author = "functionfly"

func main() {
	apiURL := os.Getenv("API")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}
	token := os.Getenv("TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "TOKEN env var is required")
		os.Exit(2)
	}
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: TOKEN=... [API=...] go run scripts/publish-function.go <author> <name>")
		os.Exit(2)
	}
	author := os.Args[1]
	name := os.Args[2]
	dir := filepath.Join("functions", author, name)

	manifestPath := filepath.Join(dir, "functionfly.jsonc")
	mainPath := filepath.Join(dir, "main.py")
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		die("read manifest %s: %v", manifestPath, err)
	}
	sourceBytes, err := os.ReadFile(mainPath)
	if err != nil {
		die("read source %s: %v", mainPath, err)
	}

	stripped := stripJSONC(manifestBytes)
	var m struct {
		Version string `json:"version"`
		Runtime string `json:"runtime"`
	}
	if err := json.Unmarshal(stripped, &m); err != nil {
		die("parse manifest: %v", err)
	}
	if m.Version == "" || m.Runtime == "" {
		die("manifest must declare version + runtime")
	}

	req := map[string]any{
		"author":      author,
		"name":        name,
		"version":     m.Version,
		"manifest":    json.RawMessage(stripped),
		"trust_level": "high",
		"source": map[string]any{
			"code":    string(sourceBytes),
			"runtime": m.Runtime,
		},
	}
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", apiURL+"/v1/registry/publish", bytes.NewReader(body))
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 60 * time.Second}).Do(httpReq)
	if err != nil {
		die("http: %v", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("[%s] %s/%s@%s -> %d %s\n", author, author, name, m.Version, resp.StatusCode, http.StatusText(resp.StatusCode))
	if len(respBody) > 0 && resp.StatusCode >= 400 {
		fmt.Println(string(respBody))
		os.Exit(1)
	}
	if len(respBody) > 0 {
		fmt.Println(string(respBody))
	}
}

// stripJSONC removes // and /* */ comments while respecting string literals.
func stripJSONC(data []byte) []byte {
	var out []byte
	inStr, esc, inLine, inBlock := false, false, false, false
	for i := 0; i < len(data); i++ {
		c := data[i]
		switch {
		case inLine:
			if c == '\n' {
				inLine = false
				out = append(out, c)
			}
			continue
		case inBlock:
			if c == '*' && i+1 < len(data) && data[i+1] == '/' {
				inBlock = false
				i++
			}
			continue
		case inStr:
			out = append(out, c)
			if esc {
				esc = false
			} else if c == '\\' {
				esc = true
			} else if c == '"' {
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
			out = append(out, c)
		case '/':
			if i+1 < len(data) {
				if data[i+1] == '/' {
					inLine = true
					i++
					continue
				}
				if data[i+1] == '*' {
					inBlock = true
					i++
					continue
				}
			}
			out = append(out, c)
		default:
			out = append(out, c)
		}
	}
	// Trailing newline so many editors don't choke.
	if len(out) > 0 && !strings.HasSuffix(string(out), "\n") {
		out = append(out, '\n')
	}
	return out
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}