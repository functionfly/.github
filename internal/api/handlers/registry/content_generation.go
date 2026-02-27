package registry

import (
	"fmt"
)

// generateJavaScriptSDKFile generates a downloadable JavaScript SDK file
func (h *Handler) generateJavaScriptSDKFile(version string) string {
	baseURL := getBaseURL()
	return fmt.Sprintf(`/**
 * FunctionFly JavaScript SDK v%s
 * https://github.com/functionfly/functionfly
 */

class FunctionFly {
  constructor(options = {}) {
    this.baseURL = options.baseURL || '%s';
    this.apiKey = options.apiKey;
    this.timeout = options.timeout || 30000;
  }

  async execute(author, name, input, options = {}) {
    const url = options.version && options.version !== 'latest'
      ? `+"`"+`${this.baseURL}/${author}/${name}@${options.version}`+"`"+`
      : `+"`"+`${this.baseURL}/${author}/${name}`+"`"+`;

    const headers = {
      'Content-Type': 'application/json',
      'User-Agent': 'FunctionFly-JS-SDK/%s',
    };

    if (this.apiKey) {
      headers['Authorization'] = 'Bearer ' + this.apiKey;
    }
    if (options.apiKey) {
      headers['Authorization'] = 'Bearer ' + options.apiKey;
    }

    const response = await fetch(url, {
      method: 'POST',
      headers,
      body: JSON.stringify(input),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error?.message || 'HTTP ' + response.status + ': ' + response.statusText);
    }

    return await response.json();
  }
}

// Export for different environments
if (typeof module !== 'undefined' && module.exports) {
  module.exports = FunctionFly;
} else if (typeof define === 'function' && define.amd) {
  define([], () => FunctionFly);
} else if (typeof window !== 'undefined') {
  window.FunctionFly = FunctionFly;
}
`, version, baseURL, version)
}

// generatePythonSDKFile generates a downloadable Python SDK file
func (h *Handler) generatePythonSDKFile(version string) string {
	baseURL := getBaseURL()
	return fmt.Sprintf(`"""
FunctionFly Python SDK v%s
https://github.com/functionfly/functionfly
"""

import requests
from typing import Dict, Any, Optional


class FunctionFly:
    def __init__(self, api_key: Optional[str] = None, base_url: str = '%s', timeout: int = 30):
        self.api_key = api_key
        self.base_url = base_url.rstrip('/')
        self.timeout = timeout

    def execute(self, author: str, name: str, input_data: Dict[str, Any],
                version: str = 'latest', api_key: Optional[str] = None) -> Dict[str, Any]:
        url = f"{self.base_url}/{author}/{name}"
        if version and version != 'latest':
            url += f"@{version}"

        headers = {
            "Content-Type": "application/json",
            "User-Agent": f"FunctionFly-Python-SDK/{version}",
        }

        auth_key = api_key or self.api_key
        if auth_key:
            headers["Authorization"] = f"Bearer {auth_key}"

        response = requests.post(
            url,
            json=input_data,
            headers=headers,
            timeout=self.timeout
        )

        response.raise_for_status()
        return response.json()
`, version, baseURL)
}

// generateGoSDKFile generates a downloadable Go SDK file
func (h *Handler) generateGoSDKFile(version string) string {
	baseURL := getBaseURL()
	return fmt.Sprintf(`// FunctionFly Go SDK v%s
// https://github.com/functionfly/functionfly

package functionfly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

func NewClient(options ...Option) *Client {
	client := &Client{
		BaseURL: "%s",
		Timeout: 30 * time.Second,
	}

	for _, option := range options {
		option(client)
	}

	return client
}

type Option func(*Client)

func WithAPIKey(apiKey string) Option {
	return func(c *Client) { c.APIKey = apiKey }
}

func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.BaseURL = baseURL }
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) { c.Timeout = timeout }
}

func (c *Client) Execute(author, name string, input interface{}, options ...ExecuteOption) (map[string]interface{}, error) {
	opts := &executeOptions{
		Version: "latest",
	}
	for _, option := range options {
		option(opts)
	}

	url := fmt.Sprintf("%%s/%%s/%%s", c.BaseURL, author, name)
	if opts.Version != "" && opts.Version != "latest" {
		url += "@" + opts.Version
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %%w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(inputJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %%w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "FunctionFly-Go-SDK/%s")

	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	if opts.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+opts.APIKey)
	}

	httpClient := &http.Client{Timeout: c.Timeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %%w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %%w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %%d: %%s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %%w", err)
	}

	return result, nil
}

type executeOptions struct {
	Version string
	APIKey  string
}

type ExecuteOption func(*executeOptions)

func WithVersion(version string) ExecuteOption {
	return func(o *executeOptions) { o.Version = version }
}

func WithExecuteAPIKey(apiKey string) ExecuteOption {
	return func(o *executeOptions) { o.APIKey = apiKey }
}
`, version, baseURL, version)
}

// generateDocumentation generates documentation content
func (h *Handler) generateDocumentation(docType, version, path string) string {
	switch docType {
	case "api":
		return h.generateAPIReference(version, path)
	case "sdk":
		return h.generateSDKDocs(version, path)
	case "guides":
		return h.generateGuide(version, path)
	default:
		return ""
	}
}

// generateAPIReference generates API reference documentation
func (h *Handler) generateAPIReference(version, path string) string {
	switch path {
	case "index.html":
		return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>FunctionFly API Reference v%s</title>
    <meta charset="utf-8">
</head>
<body>
    <h1>FunctionFly API Reference v%s</h1>
    <p>This is the API reference for FunctionFly v%s.</p>
    <h2>Endpoints</h2>
    <ul>
        <li><code>POST /v1/fx/{author}/{name}</code> - Execute a function</li>
        <li><code>GET /v1/registry/functions</code> - List functions</li>
        <li><code>GET /v1/registry/functions/{author}/{name}</code> - Get function details</li>
    </ul>
</body>
</html>`, version, version, version)
	default:
		return ""
	}
}

// generateSDKDocs generates SDK documentation
func (h *Handler) generateSDKDocs(version, path string) string {
	switch path {
	case "javascript.html":
		return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>JavaScript SDK v%s</title>
    <meta charset="utf-8">
</head>
<body>
    <h1>JavaScript SDK v%s</h1>
    <p>Install: <code>npm install @functionfly/sdk</code></p>
    <pre><code>import { FunctionFly } from '@functionfly/sdk';

const ff = new FunctionFly();
const result = await ff.execute('author', 'function', { input: 'data' });</code></pre>
</body>
</html>`, version, version)
	default:
		return ""
	}
}

// generateGuide generates guide documentation
func (h *Handler) generateGuide(version, path string) string {
	switch path {
	case "getting-started.md":
		return fmt.Sprintf(`# Getting Started with FunctionFly v%s

FunctionFly is a serverless function registry and execution platform.

## Quick Start

1. Sign up for an account
2. Publish your first function
3. Execute it via API or SDK

## SDK Installation

### JavaScript
`+"`"+`bash
npm install @functionfly/sdk
`+"`"+`

### Python
`+"`"+`bash
pip install functionfly
`+"`"+`

### Go
`+"`"+`bash
go get github.com/functionfly/functionfly-go
`+"`"+`
`, version)
	default:
		return ""
	}
}

// generateStaticAsset generates static asset content
func (h *Handler) generateStaticAsset(category, path string) string {
	switch category {
	case "images":
		if path == "logo.svg" {
			return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
  <circle cx="50" cy="50" r="40" fill="#3B82F6"/>
  <text x="50" y="55" text-anchor="middle" fill="white" font-size="20">FF</text>
</svg>`
		}
	case "css":
		if path == "styles.css" {
			return `body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; margin: 0; padding: 20px; }
h1 { color: #3B82F6; }
code { background: #f4f4f4; padding: 2px 4px; border-radius: 4px; }`
		}
	}
	return ""
}
