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

const (
	APIURL   = "http://localhost:8080"
	Email    = "admin@functionfly.local"
	Password = "admin123"
	Author   = "functionfly"
)

type Manifest struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Runtime      string   `json:"runtime"`
	Title        string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Category     string   `json:"category,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Input        any      `json:"input,omitempty"`
	Output       any      `json:"output,omitempty"`
	TimeoutMs    int      `json:"timeout_ms,omitempty"`
	MemoryMB     int      `json:"memory_mb,omitempty"`
	Tier         string   `json:"tier,omitempty"`
	PricePerCall float64  `json:"price_per_call,omitempty"`
}

func readFunction(dir string) (string, string, string, error) {
	manifestPath := filepath.Join(dir, "functionfly.jsonc")
	mainPath := filepath.Join(dir, "main.py")

	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", "", "", fmt.Errorf("read manifest: %w", err)
	}

	cleanJSON := stripJSONCComments(manifestData)

	var m Manifest
	if err := json.Unmarshal(cleanJSON, &m); err != nil {
		return "", "", "", fmt.Errorf("unmarshal manifest: %w", err)
	}

	if m.Tier == "" {
		return "", "", "", fmt.Errorf("skipping non-paid function (no tier field)")
	}

	sourceCode, err := os.ReadFile(mainPath)
	if err != nil {
		return "", "", "", fmt.Errorf("read source: %w", err)
	}

	return string(cleanJSON), string(sourceCode), m.Runtime, nil
}

func login() (string, error) {
	body, _ := json.Marshal(map[string]string{"email": Email, "password": Password})
	req, _ := http.NewRequest("POST", APIURL+"/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	result, _ := io.ReadAll(resp.Body)
	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(result, &loginResp); err != nil {
		return "", fmt.Errorf("parse error: %s", string(result))
	}
	if loginResp.Token == "" {
		return "", fmt.Errorf("login failed: %s", string(result))
	}
	return loginResp.Token, nil
}

func publish(token string, name string, manifestJSON string, sourceCode string, runtime string) error {
	reqBody := map[string]interface{}{
		"author":   Author,
		"name":     name,
		"version":  "1.0.0",
		"manifest": json.RawMessage(manifestJSON),
		"source": map[string]interface{}{
			"code":    sourceCode,
			"runtime": runtime,
		},
		"trust_level": "high",
	}
	reqJSON, _ := json.Marshal(reqBody)
	httpReq, _ := http.NewRequest("POST", APIURL+"/v1/functions/publish", bytes.NewReader(reqJSON))
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		OK      bool   `json:"ok"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parse error: %s", string(body))
	}
	if !result.OK {
		return fmt.Errorf("publish failed: %s", result.Message)
	}
	return nil
}

func main() {
	start := time.Now()

	token, err := login()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Logged in successfully (%.1fs)\n", time.Since(start).Seconds())

	baseDir := "./functions/functionfly"

	var dirs []string
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Read dir: %v\n", err)
		os.Exit(1)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") || name == "__pycache__" {
			continue
		}
		dirs = append(dirs, filepath.Join(baseDir, name))
	}

	fmt.Printf("Publishing %d functions...\n", len(dirs))

	success := 0
	failed := 0

	for _, dir := range dirs {
		_, name := filepath.Split(dir)
		manifestJSON, sourceCode, runtime, err := readFunction(dir)
		if err != nil {
			fmt.Printf("❌ %s: %v\n", name, err)
			failed++
			continue
		}

		var m Manifest
		json.Unmarshal([]byte(manifestJSON), &m)

		if err := publish(token, m.Name, manifestJSON, sourceCode, runtime); err != nil {
			fmt.Printf("❌ %s@%s: %v\n", name, m.Version, err)
			failed++
			continue
		}
		fmt.Printf("✅ %s@%s\n", name, m.Version)
		success++
	}

	elapsed := time.Since(start).Seconds()
	fmt.Printf("\n✅ Published: %d, ❌ Failed: %d (%.1fs total)\n", success, failed, elapsed)
}

func stripJSONCComments(data []byte) []byte {
	var result []byte
	inComment := false
	inBlockComment := false
	for i := 0; i < len(data); i++ {
		c := data[i]
		if inBlockComment {
			if i+1 < len(data) && data[i] == '*' && data[i+1] == '/' {
				inBlockComment = false
				i++
				continue
			}
			continue
		}
		if inComment {
			if c == '\n' {
				inComment = false
				result = append(result, c)
			}
			continue
		}
		if i+1 < len(data) && data[i] == '/' && data[i+1] == '/' {
			inComment = true
			i++
			continue
		}
		if i+1 < len(data) && data[i] == '/' && data[i+1] == '*' {
			inBlockComment = true
			i++
			continue
		}
		result = append(result, c)
	}
	return result
}
