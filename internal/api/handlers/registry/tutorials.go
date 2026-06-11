package registry

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// TutorialsHandler handles tutorial and getting-started content
type TutorialsHandler struct{}

// NewTutorialsHandler creates a new tutorials handler
func NewTutorialsHandler() *TutorialsHandler {
	return &TutorialsHandler{}
}

// HandleIndex serves the tutorials index page
func (h *TutorialsHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	html := h.generateTutorialsIndexHTML()
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// HandleGettingStarted serves the getting started tutorial
func (h *TutorialsHandler) HandleGettingStarted(w http.ResponseWriter, r *http.Request) {
	html := h.generateGettingStartedHTML()
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// HandleAPIUsage serves the API usage tutorial
func (h *TutorialsHandler) HandleAPIUsage(w http.ResponseWriter, r *http.Request) {
	html := h.generateAPIUsageHTML()
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// HandleFunctionDevelopment serves the function development tutorial
func (h *TutorialsHandler) HandleFunctionDevelopment(w http.ResponseWriter, r *http.Request) {
	html := h.generateFunctionDevelopmentHTML()
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// HandleInteractiveExample serves an interactive example page
func (h *TutorialsHandler) HandleInteractiveExample(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exampleID := vars["example"]

	var html string
	switch exampleID {
	case "hello-world":
		html = h.generateHelloWorldExample()
	case "api-integration":
		html = h.generateAPIIntegrationExample()
	case "data-processing":
		html = h.generateDataProcessingExample()
	default:
		html = h.generateExampleNotFound(exampleID)
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// generateTutorialsIndexHTML creates the tutorials index page
func (h *TutorialsHandler) generateTutorialsIndexHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>FunctionFly Tutorials - Learn & Build</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background: #f8f9fa;
        }
        .header {
            text-align: center;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 60px 40px;
            border-radius: 12px;
            margin-bottom: 40px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .title {
            font-size: 3em;
            margin: 0 0 20px 0;
            font-weight: 300;
        }
        .subtitle {
            font-size: 1.3em;
            opacity: 0.9;
            margin: 0;
        }
        .tutorials-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
            gap: 30px;
            margin-bottom: 40px;
        }
        .tutorial-card {
            background: white;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 4px 6px rgba(0,0,0,0.05);
            transition: transform 0.3s, box-shadow 0.3s;
            text-decoration: none;
            color: inherit;
        }
        .tutorial-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 8px 25px rgba(0,0,0,0.15);
            text-decoration: none;
            color: inherit;
        }
        .tutorial-header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
        }
        .tutorial-title {
            font-size: 1.5em;
            margin: 0 0 10px 0;
            font-weight: 600;
        }
        .tutorial-description {
            margin: 0;
            opacity: 0.9;
            font-size: 1em;
        }
        .tutorial-content {
            padding: 30px;
        }
        .tutorial-meta {
            display: flex;
            gap: 15px;
            font-size: 0.9em;
            color: #666;
            margin-bottom: 20px;
        }
        .tutorial-meta-item {
            display: flex;
            align-items: center;
            gap: 5px;
        }
        .tutorial-steps {
            list-style: none;
            padding: 0;
        }
        .tutorial-step {
            display: flex;
            align-items: flex-start;
            gap: 15px;
            margin-bottom: 20px;
            padding: 20px;
            background: #f8f9fa;
            border-radius: 8px;
        }
        .step-number {
            background: #667eea;
            color: white;
            width: 30px;
            height: 30px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-weight: bold;
            flex-shrink: 0;
        }
        .step-content h4 {
            margin: 0 0 10px 0;
            color: #2c3e50;
        }
        .step-content p {
            margin: 0;
            color: #555;
        }
        .cta-button {
            display: inline-block;
            background: #667eea;
            color: white;
            padding: 15px 30px;
            border-radius: 8px;
            text-decoration: none;
            font-weight: 600;
            transition: background 0.3s;
            margin-top: 10px;
        }
        .cta-button:hover {
            background: #5a67d8;
            color: white;
            text-decoration: none;
        }
        .examples-section {
            background: white;
            padding: 40px;
            border-radius: 12px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.05);
        }
        .examples-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-top: 30px;
        }
        .example-card {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            text-decoration: none;
            color: inherit;
            transition: background 0.3s;
        }
        .example-card:hover {
            background: #e9ecef;
            text-decoration: none;
            color: inherit;
        }
        .example-title {
            font-weight: 600;
            margin: 0 0 10px 0;
            color: #2c3e50;
        }
        .example-description {
            margin: 0;
            color: #666;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1 class="title">FunctionFly Tutorials</h1>
        <p class="subtitle">Learn to build and integrate serverless functions</p>
    </div>

    <div class="tutorials-grid">
        <a href="/tutorials/getting-started" class="tutorial-card">
            <div class="tutorial-header">
                <h3 class="tutorial-title">🚀 Getting Started</h3>
                <p class="tutorial-description">Your first steps with FunctionFly</p>
            </div>
            <div class="tutorial-content">
                <div class="tutorial-meta">
                    <span class="tutorial-meta-item">⏱️ 10 min</span>
                    <span class="tutorial-meta-item">🎯 Beginner</span>
                </div>
                <p>Learn the basics of FunctionFly, from account setup to your first function execution.</p>
                <div class="cta-button">Start Tutorial →</div>
            </div>
        </a>

        <a href="/tutorials/api-usage" class="tutorial-card">
            <div class="tutorial-header">
                <h3 class="tutorial-title">🔌 API Integration</h3>
                <p class="tutorial-description">Integrate functions into your applications</p>
            </div>
            <div class="tutorial-content">
                <div class="tutorial-meta">
                    <span class="tutorial-meta-item">⏱️ 15 min</span>
                    <span class="tutorial-meta-item">🎯 Intermediate</span>
                </div>
                <p>Discover how to call functions from your web apps, APIs, and backend services.</p>
                <div class="cta-button">Start Tutorial →</div>
            </div>
        </a>

        <a href="/tutorials/function-development" class="tutorial-card">
            <div class="tutorial-header">
                <h3 class="tutorial-title">⚡ Function Development</h3>
                <p class="tutorial-description">Build and publish your own functions</p>
            </div>
            <div class="tutorial-content">
                <div class="tutorial-meta">
                    <span class="tutorial-meta-item">⏱️ 20 min</span>
                    <span class="tutorial-meta-item">🎯 Advanced</span>
                </div>
                <p>Master function development, testing, and publishing to the FunctionFly registry.</p>
                <div class="cta-button">Start Tutorial →</div>
            </div>
        </a>
    </div>

    <div class="examples-section">
        <h2 style="text-align: center; margin-bottom: 10px; color: #2c3e50;">Interactive Examples</h2>
        <p style="text-align: center; color: #666; margin-bottom: 30px;">Try these hands-on examples to see FunctionFly in action</p>

        <div class="examples-grid">
            <a href="/tutorials/examples/hello-world" class="example-card">
                <h4 class="example-title">Hello World</h4>
                <p class="example-description">A simple function that returns a greeting</p>
            </a>
            <a href="/tutorials/examples/api-integration" class="example-card">
                <h4 class="example-title">API Integration</h4>
                <p class="example-description">Call external APIs and process responses</p>
            </a>
            <a href="/tutorials/examples/data-processing" class="example-card">
                <h4 class="example-title">Data Processing</h4>
                <p class="example-description">Transform and analyze data</p>
            </a>
        </div>
    </div>
</body>
</html>`
}

// generateGettingStartedHTML creates the getting started tutorial
func (h *TutorialsHandler) generateGettingStartedHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Getting Started - FunctionFly Tutorials</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background: #f8f9fa;
        }
        .header {
            text-align: center;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px;
            border-radius: 12px;
            margin-bottom: 30px;
        }
        .title { font-size: 2.5em; margin: 0 0 10px 0; font-weight: 300; }
        .subtitle { font-size: 1.2em; opacity: 0.9; margin: 0; }
        .content { background: white; padding: 40px; border-radius: 12px; box-shadow: 0 2px 4px rgba(0,0,0,0.05); }
        .step { margin-bottom: 40px; }
        .step-title {
            font-size: 1.5em;
            color: #2c3e50;
            margin-bottom: 15px;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .step-number {
            background: #667eea;
            color: white;
            width: 30px;
            height: 30px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-weight: bold;
        }
        .code-block {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 8px;
            padding: 20px;
            margin: 20px 0;
            overflow-x: auto;
            font-family: 'Monaco', 'Consolas', monospace;
            font-size: 0.9em;
        }
        .note {
            background: #e3f2fd;
            border-left: 4px solid #2196f3;
            padding: 15px;
            margin: 20px 0;
            border-radius: 0 8px 8px 0;
        }
        .warning {
            background: #fff3e0;
            border-left: 4px solid #ff9800;
            padding: 15px;
            margin: 20px 0;
            border-radius: 0 8px 8px 0;
        }
        .cta-button {
            display: inline-block;
            background: #667eea;
            color: white;
            padding: 12px 24px;
            border-radius: 6px;
            text-decoration: none;
            font-weight: 500;
            transition: background 0.3s;
        }
        .cta-button:hover { background: #5a67d8; color: white; text-decoration: none; }
        .navigation {
            display: flex;
            justify-content: space-between;
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #e9ecef;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1 class="title">Getting Started</h1>
        <p class="subtitle">Your first steps with FunctionFly</p>
    </div>

    <div class="content">
        <div class="step">
            <h2 class="step-title">
                <span class="step-number">1</span>
                What is FunctionFly?
            </h2>
            <p>FunctionFly is a serverless function registry and execution platform. It allows you to:</p>
            <ul>
                <li>Discover and execute serverless functions from a public registry</li>
                <li>Build and publish your own functions</li>
                <li>Integrate functions into your applications with simple API calls</li>
                <li>Test functions interactively through our playground</li>
            </ul>
        </div>

        <div class="step">
            <h2 class="step-title">
                <span class="step-number">2</span>
                Your First Function Execution
            </h2>
            <p>Let's start by executing a simple function. Visit the <a href="/docs">Function Registry</a> and find a function to try.</p>

            <div class="note">
                <strong>💡 Tip:</strong> Look for functions with high trust scores and good documentation.
            </div>

            <p>Once you've found a function, you can:</p>
            <ol>
                <li>View its documentation and examples</li>
                <li>Test it in the interactive playground</li>
                <li>Execute it via API calls</li>
            </ol>
        </div>

        <div class="step">
            <h2 class="step-title">
                <span class="step-number">3</span>
                API Integration
            </h2>
            <p>To call a function from your application, use this simple API pattern:</p>

            <div class="code-block">
// Example API call (JavaScript)
const response = await fetch('/fx/author/function-name', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    input: { /* your input data */ }
  })
});

const result = await response.json();
console.log(result.data);
            </div>

            <div class="code-block">
# Example API call (curl)
curl -X POST /fx/author/function-name \\
  -H "Content-Type: application/json" \\
  -d '{"input": {"key": "value"}}'
            </div>
        </div>

        <div class="step">
            <h2 class="step-title">
                <span class="step-number">4</span>
                Understanding Responses
            </h2>
            <p>All function responses follow this standard format:</p>

            <div class="code-block">
{
  "ok": true,           // Success indicator
  "data": { /* output */ }, // Function result
  "cached": false,      // Whether result came from cache
  "duration_ms": 150,   // Execution time
  "version": "1.0.0"    // Function version used
}
            </div>

            <div class="note">
                <strong>🔍 Response Fields:</strong>
                <ul>
                    <li><code>ok</code>: true for success, false for errors</li>
                    <li><code>data</code>: The function's output (only present on success)</li>
                    <li><code>cached</code>: Whether the result was served from cache</li>
                    <li><code>duration_ms</code>: How long the function took to execute</li>
                    <li><code>version</code>: Which version of the function was executed</li>
                </ul>
            </div>
        </div>

        <div class="step">
            <h2 class="step-title">
                <span class="step-number">5</span>
                Error Handling
            </h2>
            <p>When functions fail, you'll receive an error response:</p>

            <div class="code-block">
{
  "ok": false,
  "error": {
    "code": "INVALID_INPUT",
    "message": "The input parameter 'name' is required"
  },
  "duration_ms": 50
}
            </div>

            <div class="warning">
                <strong>⚠️ Important:</strong> Always check the <code>ok</code> field before using the <code>data</code>.
            </div>
        </div>

        <div class="step">
            <h2 class="step-title">
                <span class="step-number">6</span>
                Next Steps
            </h2>
            <p>Now that you understand the basics:</p>
            <ul>
                <li><a href="/tutorials/api-usage">Learn about advanced API integration</a></li>
                <li><a href="/tutorials/function-development">Build your own functions</a></li>
                <li><a href="/docs">Explore the function registry</a></li>
                <li><a href="/playground/example/hello-world">Try an interactive example</a></li>
            </ul>
        </div>

        <div class="navigation">
            <a href="/tutorials" class="cta-button">← Back to Tutorials</a>
            <a href="/tutorials/api-usage" class="cta-button">API Integration →</a>
        </div>
    </div>
</body>
</html>`
}

// generateAPIUsageHTML creates the API usage tutorial
func (h *TutorialsHandler) generateAPIUsageHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API Integration - FunctionFly Tutorials</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 800px; margin: 0 auto; padding: 20px; background: #f8f9fa; }
        .header { text-align: center; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 40px; border-radius: 12px; margin-bottom: 30px; }
        .title { font-size: 2.5em; margin: 0 0 10px 0; font-weight: 300; }
        .content { background: white; padding: 40px; border-radius: 12px; box-shadow: 0 2px 4px rgba(0,0,0,0.05); }
        .code-block { background: #f8f9fa; border: 1px solid #e9ecef; border-radius: 8px; padding: 20px; margin: 20px 0; overflow-x: auto; font-family: 'Monaco', 'Consolas', monospace; font-size: 0.9em; }
        .language-tabs { display: flex; margin-bottom: 10px; border-bottom: 1px solid #e9ecef; }
        .tab { padding: 10px 20px; cursor: pointer; border-bottom: 2px solid transparent; }
        .tab.active { border-bottom-color: #667eea; color: #667eea; font-weight: 500; }
        .tab-content { display: none; }
        .tab-content.active { display: block; }
        .note { background: #e3f2fd; border-left: 4px solid #2196f3; padding: 15px; margin: 20px 0; border-radius: 0 8px 8px 0; }
        .navigation { display: flex; justify-content: space-between; margin-top: 40px; padding-top: 20px; border-top: 1px solid #e9ecef; }
        .cta-button { display: inline-block; background: #667eea; color: white; padding: 12px 24px; border-radius: 6px; text-decoration: none; font-weight: 500; transition: background 0.3s; }
        .cta-button:hover { background: #5a67d8; color: white; text-decoration: none; }
    </style>
</head>
<body>
    <div class="header">
        <h1 class="title">API Integration</h1>
        <p class="subtitle">Integrate FunctionFly into your applications</p>
    </div>

    <div class="content">
        <h2>Authentication</h2>
        <p>Most functions are public and don't require authentication. For private functions or higher rate limits, you'll need to authenticate:</p>

        <div class="code-block">
// Get authentication token
const loginResponse = await fetch('/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'your-email@example.com',
    password: 'your-password'
  })
});

const { token } = await loginResponse.json();

// Use token in subsequent requests
const response = await fetch('/fx/author/function', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer ' + token
  },
  body: JSON.stringify({ input: {} })
});
        </div>

        <h2>Language Examples</h2>

        <div class="language-tabs">
            <div class="tab active" onclick="showTab('javascript')">JavaScript</div>
            <div class="tab" onclick="showTab('python')">Python</div>
            <div class="tab" onclick="showTab('go')">Go</div>
            <div class="tab" onclick="showTab('curl')">cURL</div>
        </div>

        <div id="javascript" class="tab-content active">
            <div class="code-block">
class FunctionFly {
  constructor(baseURL = '') {
    this.baseURL = baseURL;
  }

  async execute(author, name, input = {}, version = null, token = null) {
    const url = version
      ? this.baseURL + '/fx/' + author + '/' + name + '@' + version
      : this.baseURL + '/fx/' + author + '/' + name;

    const headers = { 'Content-Type': 'application/json' };
    if (token) headers['Authorization'] = 'Bearer ' + token;

    const response = await fetch(url, {
      method: 'POST',
      headers,
      body: JSON.stringify({ input })
    });

    return await response.json();
  }

  // Usage
  async example() {
    const result = await this.execute('example', 'hello-world', {
      name: 'Alice'
    });

    if (result.ok) {
      console.log('Success:', result.data);
    } else {
      console.error('Error:', result.error);
    }
  }
}
            </div>
        </div>

        <div id="python" class="tab-content">
            <div class="code-block">
import requests
import json

class FunctionFly:
    def __init__(self, base_url=""):
        self.base_url = base_url

    def execute(self, author, name, input_data=None, version=None, token=None):
        if input_data is None:
            input_data = {}

        url = f"{self.base_url}/fx/{author}/{name}"
        if version:
            url += f"@{version}"

        headers = {"Content-Type": "application/json"}
        if token:
            headers["Authorization"] = f"Bearer {token}"

        response = requests.post(url, json={"input": input_data}, headers=headers)
        return response.json()

# Usage
ff = FunctionFly()
result = ff.execute("example", "hello-world", {"name": "Alice"})

if result.get("ok"):
    print("Success:", result["data"])
else:
    print("Error:", result["error"])
            </div>
        </div>

        <div id="go" class="tab-content">
            <div class="code-block">
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type FunctionFly struct {
    BaseURL string
}

type ExecuteRequest struct {
    Input interface{} ` + "`" + `json:"input"` + "`" + `
}

type ExecuteResponse struct {
    OK        bool        ` + "`" + `json:"ok"` + "`" + `
    Data      interface{} ` + "`" + `json:"data,omitempty"` + "`" + `
    Error     *ErrorDetail ` + "`" + `json:"error,omitempty"` + "`" + `
    Duration  int         ` + "`" + `json:"duration_ms"` + "`" + `
    Version   string      ` + "`" + `json:"version,omitempty"` + "`" + `
}

type ErrorDetail struct {
    Code    string ` + "`" + `json:"code"` + "`" + `
    Message string ` + "`" + `json:"message"` + "`" + `
}

func (ff *FunctionFly) Execute(author, name string, input interface{}, version, token string) (*ExecuteResponse, error) {
    url := fmt.Sprintf("%s/fx/%s/%s", ff.BaseURL, author, name)
    if version != "" {
        url += "@" + version
    }

    reqBody := ExecuteRequest{Input: input}
    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")
    if token != "" {
        req.Header.Set("Authorization", "Bearer "+token)
    }

	client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var result ExecuteResponse
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, err
    }

    return &result, nil
}

func main() {
    ff := &FunctionFly{BaseURL: "https://api.functionfly.dev"}

    result, err := ff.Execute("example", "hello-world", map[string]string{
        "name": "Alice",
    }, "", "")

    if err != nil {
        fmt.Println("Request error:", err)
        return
    }

    if result.OK {
        fmt.Println("Success:", result.Data)
    } else {
        fmt.Println("Error:", result.Error.Message)
    }
}
            </div>
        </div>

        <div id="curl" class="tab-content">
            <div class="code-block">
# Basic execution
curl -X POST https://api.functionfly.dev/fx/example/hello-world \\
  -H "Content-Type: application/json" \\
  -d '{"input": {"name": "Alice"}}'

# With specific version
curl -X POST https://api.functionfly.dev/fx/example/hello-world@1.2.0 \\
  -H "Content-Type: application/json" \\
  -d '{"input": {"name": "Alice"}}'

# With authentication
curl -X POST https://api.functionfly.dev/fx/example/private-function \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_TOKEN" \\
  -d '{"input": {"data": "secret"}}'
            </div>
        </div>

        <h2>Error Handling & Retries</h2>
        <div class="code-block">
// Robust error handling with retries
async function executeWithRetry(author, name, input, maxRetries = 3) {
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      const result = await executeFunction(author, name, input);

      if (result.ok) {
        return result;
      }

      // Check if error is retryable
      const retryableErrors = ['TIMEOUT', 'INTERNAL_ERROR'];
      if (!retryableErrors.includes(result.error?.code)) {
        throw new Error(result.error?.message || 'Function execution failed');
      }

      if (attempt === maxRetries) {
        throw new Error('Failed after ' + maxRetries + ' attempts: ' + (result.error?.message || 'Unknown error'));
      }

      // Exponential backoff
      await new Promise(resolve => setTimeout(resolve, Math.pow(2, attempt) * 1000));

    } catch (error) {
      if (attempt === maxRetries) {
        throw error;
      }
      await new Promise(resolve => setTimeout(resolve, Math.pow(2, attempt) * 1000));
    }
  }
}
        </div>

        <div class="note">
            <strong>🔄 Retry Strategy:</strong>
            <ul>
                <li>Only retry on transient errors (timeouts, internal errors)</li>
                <li>Use exponential backoff (1s, 2s, 4s, etc.)</li>
                <li>Limit total retry attempts (3-5 is usually sufficient)</li>
                <li>Consider circuit breaker patterns for high-volume calls</li>
            </ul>
        </div>

        <div class="navigation">
            <a href="/tutorials/getting-started" class="cta-button">← Getting Started</a>
            <a href="/tutorials/function-development" class="cta-button">Function Development →</a>
        </div>
    </div>

    <script>
        function showTab(tabName) {
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));

            document.querySelector('.tab[onclick*=\"showTab(\'' + tabName + '\')\"]').classList.add('active');
            document.getElementById(tabName).classList.add('active');
        }
    </script>
</body>
</html>`
}

// generateFunctionDevelopmentHTML creates the function development tutorial
func (h *TutorialsHandler) generateFunctionDevelopmentHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Function Development - FunctionFly Tutorials</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 800px; margin: 0 auto; padding: 20px; background: #f8f9fa; }
        .header { text-align: center; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 40px; border-radius: 12px; margin-bottom: 30px; }
        .title { font-size: 2.5em; margin: 0 0 10px 0; font-weight: 300; }
        .content { background: white; padding: 40px; border-radius: 12px; box-shadow: 0 2px 4px rgba(0,0,0,0.05); }
        .code-block { background: #f8f9fa; border: 1px solid #e9ecef; border-radius: 8px; padding: 20px; margin: 20px 0; overflow-x: auto; font-family: 'Monaco', 'Consolas', monospace; font-size: 0.9em; }
        .file-structure { background: #2d3748; color: #e2e8f0; padding: 20px; border-radius: 8px; margin: 20px 0; font-family: 'Monaco', 'Consolas', monospace; }
        .navigation { display: flex; justify-content: space-between; margin-top: 40px; padding-top: 20px; border-top: 1px solid #e9ecef; }
        .cta-button { display: inline-block; background: #667eea; color: white; padding: 12px 24px; border-radius: 6px; text-decoration: none; font-weight: 500; transition: background 0.3s; }
        .cta-button:hover { background: #5a67d8; color: white; text-decoration: none; }
        .highlight { background: #fff3cd; padding: 2px 4px; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="header">
        <h1 class="title">Function Development</h1>
        <p class="subtitle">Build and publish your own serverless functions</p>
    </div>

    <div class="content">
        <h2>Project Structure</h2>
        <p>Start with a clean project structure:</p>

        <div class="file-structure">
my-function/
├── functionfly.json      # Function manifest
├── index.js             # Main function code
├── package.json         # Node.js dependencies
└── README.md            # Documentation
        </div>

        <h2>Function Manifest</h2>
        <p>The <code>functionfly.json</code> file defines your function:</p>

        <div class="code-block">
{
  "name": "text-analyzer",
  "version": "1.0.0",
  "runtime": "node18",
  "title": "Text Analyzer",
  "description": "Analyzes text for sentiment, keywords, and readability",
  "input": {
    "type": "object",
    "schema": {
      "type": "object",
      "properties": {
        "text": { "type": "string", "description": "The text to analyze" },
        "analysis": {
          "type": "array",
          "items": { "type": "string", "enum": ["sentiment", "keywords", "readability"] },
          "default": ["sentiment"]
        }
      },
      "required": ["text"]
    },
    "example": {
      "text": "This is a great product! I love using it.",
      "analysis": ["sentiment", "keywords"]
    }
  },
  "output": {
    "type": "object",
    "schema": {
      "type": "object",
      "properties": {
        "sentiment": { "type": "string", "enum": ["positive", "negative", "neutral"] },
        "keywords": { "type": "array", "items": { "type": "string" } },
        "readability": { "type": "number", "description": "Flesch reading ease score" }
      }
    },
    "example": {
      "sentiment": "positive",
      "keywords": ["great", "product", "love"],
      "readability": 78.5
    }
  },
  "timeout_ms": 10000,
  "memory_mb": 256,
  "deterministic": false,
  "side_effects": "none",
  "idempotent": true,
  "cache_ttl": 3600,
  "category": "text-processing",
  "tags": ["nlp", "analysis", "text"],
  "capabilities": ["ai"],
  "price_per_call": 0.01
}
        </div>

        <h2>Function Implementation</h2>
        <p>Implement your function logic:</p>

        <div class="code-block">
// index.js
const natural = require('natural');
const sentiment = require('sentiment');

module.exports = async function(input) {
  const { text, analysis = ['sentiment'] } = input;
  const result = {};

  if (analysis.includes('sentiment')) {
    const sentimentAnalysis = sentiment(text);
    result.sentiment = sentimentAnalysis.score > 0 ? 'positive' :
                       sentimentAnalysis.score < 0 ? 'negative' : 'neutral';
  }

  if (analysis.includes('keywords')) {
    const TfIdf = natural.TfIdf;
    const tfidf = new TfIdf();
    tfidf.addDocument(text);

    const keywords = [];
    tfidf.listTerms(0).slice(0, 5).forEach(item => {
      keywords.push(item.term);
    });
    result.keywords = keywords;
  }

  if (analysis.includes('readability')) {
    // Simple readability calculation
    const words = text.split(/\s+/).length;
    const sentences = text.split(/[.!?]+/).length;
    const syllables = text.split(/\s+/).reduce((count, word) => {
      return count + (word.match(/[aeiou]/gi) || []).length;
    }, 0);

    // Flesch Reading Ease formula
    const readability = 206.835 - 1.015 * (words / sentences) - 84.6 * (syllables / words);
    result.readability = Math.max(0, Math.min(100, readability));
  }

  return result;
};
        </div>

        <h2>Testing Your Function</h2>
        <p>Test locally before publishing:</p>

        <div class="code-block">
# Install dependencies
npm install

# Test with sample input
node -e "
const fn = require('./index.js');
fn({
  text: 'This is amazing! I love this product so much.',
  analysis: ['sentiment', 'keywords', 'readability']
}).then(result => console.log(JSON.stringify(result, null, 2)));
"
        </div>

        <h2>Publishing to Registry</h2>
        <p>Publish your function to make it available to others:</p>

        <div class="code-block">
# Login to FunctionFly (if required)
curl -X POST https://api.functionfly.dev/v1/auth/login \\
  -H "Content-Type: application/json" \\
  -d '{"email": "your-email@example.com", "password": "your-password"}'

# Publish your function
curl -X POST https://api.functionfly.dev/v1/registry/publish \\
  -H "Authorization: Bearer YOUR_TOKEN" \\
  -F "author=your-username" \\
  -F "name=text-analyzer" \\
  -F "version=1.0.0" \\
  -F "manifest=@functionfly.json" \\
  -F "source=@index.js" \\
  -F "source=@package.json"
        </div>

        <h2>Best Practices</h2>

        <h3>Performance</h3>
        <ul>
            <li>Use appropriate <code>timeout_ms</code> and <code>memory_mb</code> limits</li>
            <li>Enable caching with <code>cache_ttl</code> for deterministic functions</li>
            <li>Mark functions as <code>idempotent</code> when safe to retry</li>
            <li>Use efficient algorithms and data structures</li>
        </ul>

        <h3>Security</h3>
        <ul>
            <li>Validate all input parameters</li>
            <li>Use parameterized queries for database operations</li>
            <li>Limit external API calls and validate responses</li>
            <li>Declare required <code>capabilities</code> explicitly</li>
        </ul>

        <h3>Reliability</h3>
        <ul>
            <li>Implement proper error handling</li>
            <li>Return meaningful error messages</li>
            <li>Use appropriate HTTP status codes</li>
            <li>Test edge cases and failure scenarios</li>
        </ul>

        <h3>Documentation</h3>
        <ul>
            <li>Provide comprehensive examples in the manifest</li>
            <li>Document all input/output parameters</li>
            <li>Include usage examples and common patterns</li>
            <li>Specify function capabilities and limitations</li>
        </ul>

        <div class="navigation">
            <a href="/tutorials/api-usage" class="cta-button">← API Integration</a>
            <a href="/docs" class="cta-button">Explore Registry →</a>
        </div>
    </div>
</body>
</html>`
}

// generateHelloWorldExample creates an interactive hello world example
func (h *TutorialsHandler) generateHelloWorldExample() string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Hello World Example - FunctionFly</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: #e2e8f0;
            margin: 0;
            padding: 20px;
            min-height: 100vh;
        }
        .container {
            max-width: 900px;
            margin: 0 auto;
            background: white;
            border-radius: 16px;
            overflow: hidden;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            padding: 40px;
            text-align: center;
        }
        .title { font-size: 2em; margin: 0 0 10px 0; font-weight: 300; }
        .subtitle { margin: 0; opacity: 0.9; }
        .content { padding: 40px; }
        .playground {
            background: #f8f9fa;
            border-radius: 12px;
            padding: 30px;
            margin: 20px 0;
        }
        .section-title {
            font-size: 1.3em;
            color: #2c3e50;
            margin-bottom: 15px;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .code-block {
            background: #2d3748;
            color: #e2e8f0;
            border-radius: 8px;
            padding: 20px;
            margin: 15px 0;
            font-family: 'Monaco', 'Consolas', monospace;
            font-size: 14px;
            overflow-x: auto;
        }
        .input-group { margin-bottom: 20px; }
        .input-group label {
            display: block;
            margin-bottom: 5px;
            font-weight: 500;
            color: #2c3e50;
        }
        .input-field {
            width: 100%%;
            padding: 12px;
            border: 2px solid #e9ecef;
            border-radius: 8px;
            font-size: 16px;
            transition: border-color 0.3s;
        }
        .input-field:focus { outline: none; border-color: #667eea; }
        .btn {
            background: #667eea;
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 8px;
            font-size: 16px;
            cursor: pointer;
            transition: background 0.3s;
            font-weight: 500;
        }
        .btn:hover { background: #5a67d8; }
        .btn:disabled { background: #bdc3c7; cursor: not-allowed; }
        .output {
            background: #f8f9fa;
            border: 2px solid #e9ecef;
            border-radius: 8px;
            padding: 20px;
            min-height: 100px;
            margin-top: 20px;
            font-family: 'Monaco', 'Consolas', monospace;
            white-space: pre-wrap;
        }
        .success { border-color: #27ae60; background: #d4edda; color: #155724; }
        .error { border-color: #e74c3c; background: #f8d7da; color: #721c24; }
        .meta { display: flex; gap: 20px; margin-top: 10px; font-size: 14px; color: #666; }
        .explanation {
            background: #e3f2fd;
            border-left: 4px solid #2196f3;
            padding: 20px;
            margin: 20px 0;
            border-radius: 0 8px 8px 0;
        }
        .navigation { text-align: center; margin-top: 40px; padding-top: 20px; border-top: 1px solid #e9ecef; }
        .nav-btn { background: #95a5a6; margin: 0 10px; }
        .nav-btn:hover { background: #7f8c8d; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="title">Hello World Example</h1>
            <p class="subtitle">Your first interactive function execution</p>
        </div>

        <div class="content">
            <div class="explanation">
                <strong>🎯 What this example does:</strong>
                <p>This simple function takes your name as input and returns a personalized greeting. It's a great way to understand how FunctionFly functions work!</p>
                <p><strong>Function:</strong> <code>example/hello-world</code></p>
                <p><strong>Input:</strong> A JSON object with a "name" field</p>
                <p><strong>Output:</strong> A greeting message</p>
            </div>

            <div class="playground">
                <div class="section-title">✏️ Try It Yourself</div>

                <div class="input-group">
                    <label for="name">Your Name:</label>
                    <input type="text" id="name" class="input-field" placeholder="Enter your name" value="Alice">
                </div>

                <button class="btn" id="runBtn" onclick="executeFunction()">
                    <span id="btnText">🚀 Run Function</span>
                </button>

                <div id="output" class="output">Click "Run Function" to see the result...</div>

                <div class="meta" id="meta" style="display: none;">
                    <span id="latency"></span>
                    <span id="status"></span>
                </div>
            </div>

            <div class="section-title">📚 How It Works</div>

            <p>The function code looks like this:</p>

            <div class="code-block">
// example/hello-world function
module.exports = async function(input) {
  const { name = 'World' } = input;

  // Validate input
  if (typeof name !== 'string' || name.trim().length === 0) {
    throw new Error('Name must be a non-empty string');
  }

  // Return greeting
  return {
    greeting: 'Hello, ' + name.trim() + '! Welcome to FunctionFly!',
    timestamp: new Date().toISOString(),
    function: 'hello-world'
  };
};
            </div>

            <p>When you call this function with <code>{"name": "Alice"}</code>, it returns:</p>

            <div class="code-block">
{
  "greeting": "Hello, Alice! Welcome to FunctionFly!",
  "timestamp": "2024-01-15T10:30:00.000Z",
  "function": "hello-world"
}
            </div>

            <div class="section-title">🔧 API Call</div>

            <p>You can also call this function directly via API:</p>

            <div class="code-block">
curl -X POST /fx/example/hello-world \\
  -H "Content-Type: application/json" \\
  -d '{"input": {"name": "Alice"}}'
            </div>

            <div class="navigation">
                <a href="/tutorials" class="btn nav-btn">← Back to Tutorials</a>
                <a href="/playground/example/hello-world" class="btn">Open Full Playground</a>
                <a href="/tutorials/examples/api-integration" class="btn nav-btn">Next Example →</a>
            </div>
        </div>
    </div>

    <script>
        async function executeFunction() {
            const btn = document.getElementById('runBtn');
            const btnText = document.getElementById('btnText');
            const output = document.getElementById('output');
            const meta = document.getElementById('meta');
            const latency = document.getElementById('latency');
            const status = document.getElementById('status');
            const nameInput = document.getElementById('name');

            btn.disabled = true;
            btnText.textContent = '⏳ Running...';
            output.className = 'output';
            output.textContent = 'Executing function...';
            meta.style.display = 'none';

            try {
                const name = nameInput.value.trim() || 'World';
                const start = Date.now();

                const response = await fetch('/fx/example/hello-world', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ input: { name } })
                });

                const result = await response.json();
                const elapsed = Date.now() - start;

                latency.textContent = '⚡ ' + elapsed + 'ms';
                status.textContent = '📊 Status: ' + response.status;
                meta.style.display = 'flex';

                if (result.ok) {
                    output.className = 'output success';
                    output.textContent = '✅ Success!\\n\\n' + JSON.stringify(result.data, null, 2);
                } else {
                    output.className = 'output error';
                    output.textContent = '❌ Error: ' + (result.error?.message || 'Unknown error');
                }

            } catch (err) {
                output.className = 'output error';
                output.textContent = '❌ Request failed: ' + err.message;
            }

            btn.disabled = false;
            btnText.textContent = '🚀 Run Function';
        }

        // Auto-run on page load
        setTimeout(executeFunction, 1000);
    </script>
</body>
</html>`)
}

// generateAPIIntegrationExample creates an API integration example
func (h *TutorialsHandler) generateAPIIntegrationExample() string {
	return `<p>API Integration Example - Coming Soon!</p>`
}

// generateDataProcessingExample creates a data processing example
func (h *TutorialsHandler) generateDataProcessingExample() string {
	return `<p>Data Processing Example - Coming Soon!</p>`
}

// generateExampleNotFound creates a 404 page for unknown examples
func (h *TutorialsHandler) generateExampleNotFound(exampleID string) string {
	return fmt.Sprintf(`<h1>Example Not Found</h1><p>The example "%s" doesn't exist.</p><a href="/tutorials">Back to Tutorials</a>`, exampleID)
}
