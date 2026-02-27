package registry

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// PlaygroundHandler handles interactive playground for registry functions
type PlaygroundHandler struct {
	repo *registry.RegistryRepository
}

// NewPlaygroundHandler creates a new playground handler
func NewPlaygroundHandler(repo *registry.RegistryRepository) *PlaygroundHandler {
	return &PlaygroundHandler{repo: repo}
}

// HandlePlaygroundUI serves the playground UI for a registry function
func (h *PlaygroundHandler) HandlePlaygroundUI(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	fnVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		http.Error(w, "Function version not found", http.StatusNotFound)
		return
	}

	// Get rating for trust information
	rating, _ := h.repo.GetRatingByFunctionID(fn.ID)

	html := h.generateRegistryPlaygroundHTML(fn, fnVersion, rating)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// HandlePlaygroundExecute executes a registry function with custom input
func (h *PlaygroundHandler) HandlePlaygroundExecute(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	fnVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		http.Error(w, "Function version not found", http.StatusNotFound)
		return
	}

	// Parse request body
	var execReq struct {
		Input json.RawMessage `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&execReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Execute the function via the registry execution endpoint
	// For now, we'll use the existing execution logic but enhance it for playground use
	start := time.Now()

	// Create execution request
	execRequest := functionregistry.ExecutionRequest{
		Author:  author,
		Name:    name,
		Version: fnVersion.Version,
		Input:   execReq.Input,
	}

	// Marshal request for internal execution
	reqBytes, err := json.Marshal(execRequest)
	if err != nil {
		http.Error(w, "Failed to marshal request", http.StatusInternalServerError)
		return
	}

	// Create internal HTTP request to the execution endpoint
	// Use the actual server port from the request or environment
	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "8090" // Default to 8090 since that's what we're running on
	}
	targetURL := fmt.Sprintf("http://localhost:%s/v1/fx/%s/%s@%s", serverPort, author, name, fnVersion.Version)

	proxyReq, err := http.NewRequestWithContext(r.Context(), "POST", targetURL, bytes.NewReader(reqBytes))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("X-FunctionFly-Playground", "true")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(proxyReq)
	latencyMs := time.Since(start).Milliseconds()

	if err != nil {
		response := PlaygroundExecuteResponse{
			Success:   false,
			Error:     err.Error(),
			LatencyMs: latencyMs,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var output interface{}
	if err := json.Unmarshal(body, &output); err != nil {
		output = string(body)
	}

	// Parse the execution response
	var execResponse functionregistry.ExecutionResponse
	if err := json.Unmarshal(body, &execResponse); err != nil {
		// Fallback for error responses
		var errorResp functionregistry.ExecutionError
		if err := json.Unmarshal(body, &errorResp); err != nil {
			response := PlaygroundExecuteResponse{
				Success:    resp.StatusCode >= 200 && resp.StatusCode < 300,
				Output:     output,
				LatencyMs:  latencyMs,
				StatusCode: resp.StatusCode,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
		response := PlaygroundExecuteResponse{
			Success:    errorResp.OK,
			Output:     nil,
			Error:      errorResp.Error.Message,
			LatencyMs:  int64(errorResp.DurationMs),
			StatusCode: resp.StatusCode,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := PlaygroundExecuteResponse{
		Success:    execResponse.OK,
		Output:     execResponse.Data,
		LatencyMs:  int64(execResponse.DurationMs),
		StatusCode: resp.StatusCode,
		Version:    execResponse.Version,
	}

	// Record the execution for replay/sharing (synchronously to get the ID)
	var outputJSON json.RawMessage
	if execResponse.Data != nil {
		outputJSON, _ = json.Marshal(execResponse.Data)
	}
	executionID := h.recordPlaygroundExecution(fn.ID.String(), fnVersion.ID.String(), execReq.Input, outputJSON, resp.StatusCode, int(latencyMs))
	response.ExecutionID = executionID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandlePlaygroundShare creates a shareable playground URL
func (h *PlaygroundHandler) HandlePlaygroundShare(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	var req struct {
		Input string `json:"input"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create shareable URL
	shareURL := fmt.Sprintf("/playground/%s/%s", author, name)
	if req.Input != "" {
		// URL encode the input
		shareURL += "?input=" + req.Input
	}

	response := map[string]interface{}{
		"share_url": shareURL,
		"full_url":  fmt.Sprintf("https://functionfly.dev%s", shareURL),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PlaygroundExecuteResponse represents a playground execution response
type PlaygroundExecuteResponse struct {
	Success     bool        `json:"success"`
	Output      interface{} `json:"output,omitempty"`
	Error       string      `json:"error,omitempty"`
	LatencyMs   int64       `json:"latency_ms"`
	StatusCode  int         `json:"status_code"`
	Version     string      `json:"version,omitempty"`
	ExecutionID string      `json:"execution_id,omitempty"` // For replay
}

// recordPlaygroundExecution records a playground execution to the database and returns the public ID
func (h *PlaygroundHandler) recordPlaygroundExecution(functionID, versionID string, input, output json.RawMessage, statusCode, durationMs int) string {
	// Generate a short public ID
	publicID := generatePublicID()

	// Parse function ID
	fnID, err := uuid.Parse(functionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse function ID")
		return ""
	}

	// Create the public execution record
	exec := &registry.RegistryExecutionPublic{
		PublicID:   publicID,
		FunctionID: fnID,
		Version:    versionID,
		InputJSON:  input,
		OutputJSON: output,
		DurationMs: durationMs,
		Cached:     false,
		Shareable:  true,
	}

	// Save to database
	if err := h.repo.CreateExecutionPublic(exec); err != nil {
		logrus.WithError(err).Error("Failed to create public execution")
		return ""
	}

	logrus.WithFields(logrus.Fields{
		"function_id": functionID,
		"public_id":   publicID,
	}).Info("Playground execution recorded")

	return publicID
}

// generatePublicID generates a short, URL-friendly public ID
func generatePublicID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return "exec_" + base64.URLEncoding.EncodeToString(b)[:10]
}

// generateRegistryPlaygroundHTML generates the playground UI HTML for registry functions
func (h *PlaygroundHandler) generateRegistryPlaygroundHTML(fn *registry.RegistryFunction, fnVersion *registry.RegistryFunctionVersion, rating *registry.RegistryFunctionRating) string {
	var manifest functionregistry.FunctionManifest
	json.Unmarshal(fnVersion.Manifest, &manifest)

	title := fn.Title.String
	if title == "" {
		title = fmt.Sprintf("%s/%s", fn.Author, fn.Name)
	}

	description := fn.Description.String
	if description == "" {
		description = "Interactive playground for testing this function"
	}

	// Get example input from manifest
	exampleInput := "{}"
	if manifest.Input != nil && manifest.Input.Example != nil {
		if exampleBytes, err := json.MarshalIndent(manifest.Input.Example, "", "  "); err == nil {
			exampleInput = string(exampleBytes)
		}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - FunctionFly Registry Playground</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: #e2e8f0; min-height: 100vh;
            padding: 20px;
        }
        .container { max-width: 1000px; margin: 0 auto; }
        header {
            text-align: center;
            background: rgba(255,255,255,0.1);
            padding: 40px;
            border-radius: 16px;
            margin-bottom: 30px;
            backdrop-filter: blur(10px);
        }
        h1 { font-size: 2.5em; margin-bottom: 10px; font-weight: 300; }
        .subtitle { font-size: 1.2em; opacity: 0.9; margin-bottom: 20px; }
        .meta { display: flex; gap: 20px; justify-content: center; flex-wrap: wrap; }
        .meta-item {
            background: rgba(255,255,255,0.1);
            padding: 8px 16px;
            border-radius: 20px;
            font-size: 0.9em;
        }
        .playground {
            background: white;
            border-radius: 16px;
            overflow: hidden;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
        }
        .section { padding: 30px; }
        .section:not(:last-child) { border-bottom: 1px solid #e9ecef; }
        .section-title {
            font-size: 1.5em;
            margin-bottom: 20px;
            color: #2c3e50;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .section-title::before {
            content: '';
            width: 4px;
            height: 24px;
            background: #3498db;
            border-radius: 2px;
        }
        textarea {
            width: 100%%;
            min-height: 200px;
            background: #f8f9fa;
            border: 2px solid #e9ecef;
            border-radius: 12px;
            padding: 20px;
            font-family: 'Monaco', 'Consolas', monospace;
            font-size: 14px;
            resize: vertical;
            transition: border-color 0.3s;
        }
        textarea:focus { outline: none; border-color: #3498db; }
        .btn {
            background: #3498db;
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 8px;
            font-size: 1em;
            cursor: pointer;
            transition: all 0.3s;
            font-weight: 500;
        }
        .btn:hover { background: #2980b9; transform: translateY(-1px); }
        .btn:disabled { background: #bdc3c7; cursor: not-allowed; transform: none; }
        .btn.secondary { background: #95a5a6; }
        .btn.secondary:hover { background: #7f8c8d; }
        .btn-row { display: flex; gap: 15px; margin-top: 20px; flex-wrap: wrap; }
        .response {
            background: #f8f9fa;
            border: 2px solid #e9ecef;
            border-radius: 12px;
            padding: 20px;
            min-height: 150px;
            font-family: 'Monaco', 'Consolas', monospace;
            font-size: 14px;
            white-space: pre-wrap;
            word-break: break-all;
            position: relative;
        }
        .response.success { border-color: #27ae60; }
        .response.error { border-color: #e74c3c; }
        .response.loading { border-color: #f39c12; }
        .status-badge {
            display: inline-block;
            padding: 6px 12px;
            border-radius: 20px;
            font-size: 0.8em;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        .status-success { background: #d4edda; color: #155724; }
        .status-error { background: #f8d7da; color: #721c24; }
        .status-loading { background: #fff3cd; color: #856404; }
        .meta-info {
            display: flex;
            gap: 20px;
            font-size: 0.9em;
            color: #6c757d;
            margin-top: 15px;
            flex-wrap: wrap;
        }
        .meta-item { display: flex; align-items: center; gap: 5px; }
        .loading-spinner {
            display: inline-block;
            width: 16px;
            height: 16px;
            border: 2px solid #f3f3f3;
            border-top: 2px solid #3498db;
            border-radius: 50%%;
            animation: spin 1s linear infinite;
        }
        @keyframes spin {
            0%% { transform: rotate(0deg); }
            100%% { transform: rotate(360deg); }
        }
        .toast {
            position: fixed;
            bottom: 20px;
            right: 20px;
            background: #27ae60;
            color: white;
            padding: 15px 25px;
            border-radius: 8px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.3);
            opacity: 0;
            transition: opacity 0.3s;
            z-index: 1000;
        }
        .toast.show { opacity: 1; }
        .examples {
            margin-top: 20px;
            padding: 20px;
            background: #f8f9fa;
            border-radius: 8px;
        }
        .example-btn {
            background: #e9ecef;
            border: none;
            padding: 8px 16px;
            border-radius: 6px;
            cursor: pointer;
            margin: 5px;
            font-size: 0.9em;
            transition: background 0.3s;
        }
        .example-btn:hover { background: #dee2e6; }
        .trust-score {
            background: linear-gradient(45deg, #27ae60, #2ecc71);
            color: white;
            padding: 4px 12px;
            border-radius: 15px;
            font-size: 0.8em;
            font-weight: 600;
        }
        .docs-link {
            color: #3498db;
            text-decoration: none;
            font-weight: 500;
        }
        .docs-link:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>%s</h1>
            <p class="subtitle">%s</p>
            <div class="meta">
                <span class="meta-item">Author: %s</span>
                <span class="meta-item">Version: %s</span>
                <span class="meta-item">Runtime: %s</span>
                <span class="meta-item"><span class="trust-score">Trust: %.1f</span></span>
            </div>
        </header>

        <div class="playground">
            <div class="section">
                <div class="section-title">Input (JSON)</div>
                <textarea id="input" placeholder='%s'>%s</textarea>
                <div class="btn-row">
                    <button class="btn" id="runBtn" onclick="executeFunction()">
                        <span id="runText">Run Function</span>
                    </button>
                    <button class="btn secondary" onclick="sharePlayground()">Share</button>
                    <a href="/docs/%s/%s" class="btn secondary docs-link" style="color: white; text-decoration: none;">View Docs</a>
                </div>
            </div>

            <div class="section">
                <div class="section-title">Output</div>
                <div id="output" class="response">Click "Run Function" to execute...</div>
                <div class="meta-info">
                    <span class="meta-item" id="latency"></span>
                    <span class="meta-item" id="status"></span>
                    <span class="meta-item" id="version"></span>
                </div>
            </div>
        </div>
    </div>

    <div id="toast" class="toast">Link copied to clipboard!</div>

    <script>
        const author = "%s";
        const name = "%s";

        // Parse URL params for pre-filled input
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.has('input')) {
            document.getElementById('input').value = decodeURIComponent(urlParams.get('input'));
        }

        async function executeFunction() {
            const btn = document.getElementById('runBtn');
            const runText = document.getElementById('runText');
            const output = document.getElementById('output');
            const latency = document.getElementById('latency');
            const status = document.getElementById('status');
            const version = document.getElementById('version');

            btn.disabled = true;
            runText.innerHTML = '<div class="loading-spinner"></div> Running...';
            output.className = 'response loading';
            output.textContent = 'Executing function...';
            latency.textContent = '';
            status.textContent = '';
            version.textContent = '';

            try {
                const inputVal = document.getElementById('input').value.trim();
                let input = {};
                if (inputVal) {
                    try {
                        input = JSON.parse(inputVal);
                    } catch(e) {
                        throw new Error('Invalid JSON input: ' + e.message);
                    }
                }

                const start = Date.now();
                const response = await fetch('/v1/playground/' + author + '/' + name + '/execute', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ input: input })
                });
                const result = await response.json();
                const elapsed = Date.now() - start;

                latency.textContent = '⏱️ ' + elapsed + 'ms';
                status.textContent = '📊 Status: ' + result.status_code;
                version.textContent = '🏷️ Version: ' + (result.version || 'latest');

                if (result.success) {
                    output.className = 'response success';
                    output.textContent = '✅ Success\\n\\n' + JSON.stringify(result.output, null, 2);
                } else {
                    output.className = 'response error';
                    let errorMsg = '❌ Error';
                    if (result.error) {
                        errorMsg += ': ' + result.error;
                    }
                    if (result.output) {
                        errorMsg += '\\n\\nOutput: ' + JSON.stringify(result.output, null, 2);
                    }
                    output.textContent = errorMsg;
                }
            } catch (err) {
                output.className = 'response error';
                output.textContent = '❌ Request failed: ' + err.message;
            }

            btn.disabled = false;
            runText.textContent = 'Run Function';
        }

        async function sharePlayground() {
            const input = document.getElementById('input').value.trim();
            try {
                const response = await fetch('/v1/playground/' + author + '/' + name + '/share', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ input: input })
                });
                const result = await response.json();

                // Copy to clipboard
                await navigator.clipboard.writeText(result.full_url);

                const toast = document.getElementById('toast');
                toast.classList.add('show');
                setTimeout(() => toast.classList.remove('show'), 3000);
            } catch (err) {
                alert('Failed to create share link: ' + err.message);
            }
        }

        // Auto-run if input is provided in URL
        if (urlParams.has('input')) {
            setTimeout(executeFunction, 500);
        }
    </script>
</body>
</html>`,
		title, title, description, fn.Author, fnVersion.Version, fnVersion.Runtime, rating.TrustScore,
		"Enter your JSON input here...", exampleInput,
		fn.Author, fn.Name,
		fn.Author, fn.Name)
}

// HandleFunctionPage serves the combined function page (docs + playground)
func (h *PlaygroundHandler) HandleFunctionPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	fnVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		http.Error(w, "Function version not found", http.StatusNotFound)
		return
	}

	// Get rating for trust information
	rating, _ := h.repo.GetRatingByFunctionID(fn.ID)

	// Generate combined function page with docs + playground
	html := h.generateFunctionPageHTML(fn, fnVersion, rating)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// HandleReplay serves a replay page for a past execution
func (h *PlaygroundHandler) HandleReplay(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	executionID := vars["executionId"]

	// Get execution by public ID
	exec, err := h.repo.GetExecutionPublicByID(executionID)
	if err != nil {
		http.Error(w, "Execution not found or not shareable", http.StatusNotFound)
		return
	}

	// Get function info
	fn, err := h.repo.GetFunctionByID(exec.FunctionID)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	// Generate replay page
	html := h.generateReplayHTML(fn, exec)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// HandleCodeExamples returns code examples for a function
func (h *PlaygroundHandler) HandleCodeExamples(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	fnVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		http.Error(w, "Function version not found", http.StatusNotFound)
		return
	}

	// Build API URL
	baseURL := "https://api.functionfly.com"
	if r.TLS == nil {
		baseURL = "http://localhost:8080"
	}
	apiURL := fmt.Sprintf("%s/v1/fx/%s/%s", baseURL, author, name)

	// Get example input from manifest
	var manifest functionregistry.FunctionManifest
	json.Unmarshal(fnVersion.Manifest, &manifest)

	exampleInput := "{}"
	if manifest.Input != nil && manifest.Input.Example != nil {
		if exampleBytes, err := json.Marshal(string(manifest.Input.Example)); err == nil {
			exampleInput = string(exampleBytes)
		}
	}

	// Generate code examples
	codeExamples := map[string]string{
		"curl":       fmt.Sprintf("curl -X POST %s -d '%s'", apiURL, exampleInput),
		"javascript": fmt.Sprintf("await fetch('%s', { method: 'POST', body: %s })", apiURL, exampleInput),
		"python":     fmt.Sprintf("requests.post('%s', data=%s)", apiURL, exampleInput),
		"go":         fmt.Sprintf("resp, err := http.Post(\"%s\", \"application/json\", strings.NewReader(%s))", apiURL, exampleInput),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(codeExamples)
}

// HandleAIToolSchema returns AI tool schema for a function
func (h *PlaygroundHandler) HandleAIToolSchema(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	fnVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		http.Error(w, "Function version not found", http.StatusNotFound)
		return
	}

	// Get manifest
	var manifest functionregistry.FunctionManifest
	json.Unmarshal(fnVersion.Manifest, &manifest)

	// Build OpenAI tool schema
	title := manifest.Title
	if title == "" {
		title = fmt.Sprintf("%s/%s", author, name)
	}

	description := manifest.Description
	if description == "" {
		description = fmt.Sprintf("Execute the %s function", name)
	}

	toolSchema := map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        fmt.Sprintf("%s_%s", author, name),
			"description": description,
			"parameters":  manifest.Input.Schema,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toolSchema)
}

// generateFunctionPageHTML generates the combined function page (docs + playground)
func (h *PlaygroundHandler) generateFunctionPageHTML(fn *registry.RegistryFunction, fnVersion *registry.RegistryFunctionVersion, rating *registry.RegistryFunctionRating) string {
	var manifest functionregistry.FunctionManifest
	json.Unmarshal(fnVersion.Manifest, &manifest)

	title := fn.Title.String
	if title == "" {
		title = fmt.Sprintf("%s/%s", fn.Author, fn.Name)
	}

	description := fn.Description.String
	if description == "" {
		description = manifest.Description
	}

	// Get example input from manifest
	exampleInput := "{}"
	if manifest.Input != nil && manifest.Input.Example != nil {
		if exampleBytes, err := json.MarshalIndent(manifest.Input.Example, "", "  "); err == nil {
			exampleInput = string(exampleBytes)
		}
	}

	// Build API URL
	apiURL := fmt.Sprintf("https://api.functionfly.com/v1/fx/%s/%s", fn.Author, fn.Name)

	trustScore := 0.0
	if rating != nil {
		trustScore = rating.TrustScore
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s | FunctionFly</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: #0f172a; color: #e2e8f0; min-height: 100vh;
        }
        .container { max-width: 1200px; margin: 0 auto; padding: 2rem; }
        header {
            text-align: center; padding: 3rem 0; border-bottom: 1px solid #1e293b;
        }
        h1 { font-size: 2.5rem; margin-bottom: 0.5rem; color: #f8fafc; }
        .subtitle { color: #94a3b8; font-size: 1.1rem; margin-bottom: 1rem; }
        .meta { display: flex; gap: 1rem; justify-content: center; flex-wrap: wrap; margin-top: 1rem; }
        .badge {
            background: #1e293b; padding: 0.5rem 1rem; border-radius: 9999px;
            font-size: 0.875rem; color: #94a3b8;
        }
        .badge.primary { background: #3b82f6; color: white; }
        .badge.success { background: #10b981; color: white; }
        .tabs { display: flex; gap: 0.5rem; margin: 2rem 0; border-bottom: 1px solid #1e293b; }
        .tab {
            padding: 1rem 1.5rem; background: none; border: none; color: #94a3b8;
            cursor: pointer; font-size: 1rem; border-bottom: 2px solid transparent;
        }
        .tab.active { color: #3b82f6; border-bottom-color: #3b82f6; }
        .tab-content { display: none; }
        .tab-content.active { display: block; }
        .playground { background: #1e293b; border-radius: 12px; overflow: hidden; }
        .section { padding: 1.5rem; border-bottom: 1px solid #334155; }
        .section:last-child { border-bottom: none; }
        .section-title {
            font-size: 0.875rem; text-transform: uppercase; letter-spacing: 0.05em;
            color: #64748b; margin-bottom: 1rem;
        }
        textarea {
            width: 100%%; min-height: 150px; background: #0f172a; border: 1px solid #334155;
            border-radius: 8px; padding: 1rem; color: #e2e8f0; font-family: 'Monaco', monospace;
            font-size: 0.9rem; resize: vertical;
        }
        textarea:focus { outline: none; border-color: #3b82f6; }
        .btn {
            background: #3b82f6; color: white; border: none; padding: 0.75rem 1.5rem;
            border-radius: 8px; font-size: 1rem; cursor: pointer; transition: background 0.2s;
        }
        .btn:hover { background: #2563eb; }
        .btn:disabled { background: #475569; cursor: not-allowed; }
        .btn-row { display: flex; gap: 1rem; margin-top: 1rem; }
        .response {
            background: #0f172a; border: 1px solid #334155; border-radius: 8px;
            padding: 1rem; min-height: 100px; font-family: 'Monaco', monospace;
            font-size: 0.9rem; white-space: pre-wrap; word-break: break-all;
        }
        .response.success { border-color: #10b981; }
        .response.error { border-color: #ef4444; }
        .code-examples { background: #0f172a; border-radius: 8px; padding: 1rem; }
        .code-block {
            background: #1e293b; padding: 1rem; border-radius: 6px; margin-bottom: 1rem;
            font-family: 'Monaco', monospace; font-size: 0.85rem; overflow-x: auto;
        }
        .code-block pre { margin: 0; color: #e2e8f0; }
        .code-label { color: #64748b; font-size: 0.75rem; margin-bottom: 0.5rem; }
        .loading { color: #3b82f6; }
        .error { color: #ef4444; }
        .success { color: #10b981; }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>%s</h1>
            <p class="subtitle">%s</p>
            <div class="meta">
                <span class="badge">by %s</span>
                <span class="badge primary">v%s</span>
                <span class="badge">%s</span>
                <span class="badge success">Trust: %.1f</span>
            </div>
        </header>

        <div class="tabs">
            <button class="tab active" data-tab="playground">Playground</button>
            <button class="tab" data-tab="code">Code Examples</button>
            <button class="tab" data-tab="docs">Documentation</button>
        </div>

        <div id="playground" class="tab-content active">
            <div class="playground">
                <div class="section">
                    <div class="section-title">Input (JSON)</div>
                    <textarea id="input">%s</textarea>
                    <div class="btn-row">
                        <button class="btn" id="runBtn" onclick="executeFunction()">Run Function</button>
                        <button class="btn" style="background: #475569;" onclick="shareExecution()">Share Result</button>
                    </div>
                </div>
                <div class="section">
                    <div class="section-title">Output</div>
                    <div id="output" class="response">Click "Run Function" to execute...</div>
                    <div id="meta" style="margin-top: 0.5rem; color: #64748b; font-size: 0.875rem;"></div>
                </div>
            </div>
        </div>

        <div id="code" class="tab-content">
            <div class="code-examples">
                <div class="code-label">cURL</div>
                <div class="code-block"><pre id="curl-example"></pre></div>
                <div class="code-label">JavaScript</div>
                <div class="code-block"><pre id="js-example"></pre></div>
                <div class="code-label">Python</div>
                <div class="code-block"><pre id="python-example"></pre></div>
            </div>
        </div>

        <div id="docs" class="tab-content">
            <div class="section">
                <h3>Description</h3>
                <p style="color: #94a3b8; margin-top: 0.5rem;">%s</p>
            </div>
            <div class="section">
                <h3>API Endpoint</h3>
                <div class="code-block" style="margin-top: 0.5rem;"><pre>%s</pre></div>
            </div>
        </div>
    </div>

    <script>
        const author = "%s";
        const name = "%s";
        const apiURL = "%s";

        // Tab switching
        document.querySelectorAll('.tab').forEach(tab => {
            tab.addEventListener('click', () => {
                document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
                document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
                tab.classList.add('active');
                document.getElementById(tab.dataset.tab).classList.add('active');
            });
        });

        // Load code examples
        fetch('/v1/fx/' + author + '/' + name + '/code')
            .then(r => r.json())
            .then(examples => {
                document.getElementById('curl-example').textContent = examples.curl;
                document.getElementById('js-example').textContent = examples.javascript;
                document.getElementById('python-example').textContent = examples.python;
            });

        // Parse URL params for pre-filled input
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.has('input')) {
            document.getElementById('input').value = decodeURIComponent(urlParams.get('input'));
        }

        async function executeFunction() {
            const btn = document.getElementById('runBtn');
            const output = document.getElementById('output');
            const meta = document.getElementById('meta');

            btn.disabled = true;
            btn.textContent = 'Running...';
            output.className = 'response loading';
            output.textContent = 'Executing function...';
            meta.textContent = '';

            try {
                const inputVal = document.getElementById('input').value || '{}';
                let input;
                try {
                    input = JSON.parse(inputVal);
                } catch(e) {
                    input = { data: inputVal };
                }

                const start = Date.now();
                const response = await fetch('/v1/run/' + author + '/' + name + '/execute', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ input })
                });
                const result = await response.json();
                const elapsed = Date.now() - start;

                meta.textContent = 'Duration: ' + elapsed + 'ms | Version: ' + (result.version || 'N/A');

                if (result.success) {
                    output.className = 'response success';
                    output.textContent = JSON.stringify(result.output, null, 2);
                } else {
                    output.className = 'response error';
                    output.textContent = result.error || 'Unknown error';
                }
            } catch (err) {
                output.className = 'response error';
                output.textContent = 'Request failed: ' + err.message;
            }

            btn.disabled = false;
            btn.textContent = 'Run Function';
        }

        async function shareExecution() {
            const input = document.getElementById('input').value;
            try {
                const response = await fetch('/v1/run/' + author + '/' + name + '/share', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ input })
                });
                const result = await response.json();
                await navigator.clipboard.writeText(result.share_url);
                alert('Share link copied to clipboard!');
            } catch (err) {
                alert('Failed to create share link: ' + err.message);
            }
        }

        // Auto-run if input is provided in URL
        if (urlParams.has('input')) {
            setTimeout(executeFunction, 500);
        }
    </script>
</body>
</html>`,
		title, title, description, fn.Author, fnVersion.Version, fnVersion.Runtime, trustScore,
		exampleInput,
		description, apiURL,
		fn.Author, fn.Name, apiURL)
}

// generateReplayHTML generates the replay page for a past execution
func (h *PlaygroundHandler) generateReplayHTML(fn *registry.RegistryFunction, exec *registry.RegistryExecutionPublic) string {
	title := fn.Title.String
	if title == "" {
		title = fmt.Sprintf("%s/%s execution", fn.Author, fn.Name)
	}

	inputJSON := "{}"
	if exec.InputJSON != nil {
		if inputBytes, err := json.MarshalIndent(exec.InputJSON, "", "  "); err == nil {
			inputJSON = string(inputBytes)
		}
	}

	outputJSON := "null"
	if exec.OutputJSON != nil {
		if outputBytes, err := json.MarshalIndent(exec.OutputJSON, "", "  "); err == nil {
			outputJSON = string(outputBytes)
		}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	   <meta charset="UTF-8">
	   <meta name="viewport" content="width=device-width, initial-scale=1.0">
	   <title>Replay: %s | FunctionFly</title>
	   <style>
	       * { box-sizing: border-box; margin: 0; padding: 0; }
	       body {
	           font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
	           background: #0f172a; color: #e2e8f0; min-height: 100vh;
	       }
	       .container { max-width: 1000px; margin: 0 auto; padding: 2rem; }
	       header { text-align: center; padding: 2rem 0; border-bottom: 1px solid #1e293b; margin-bottom: 2rem; }
	       h1 { font-size: 1.5rem; color: #f8fafc; margin-bottom: 0.5rem; }
	       .meta { color: #64748b; font-size: 0.875rem; }
	       .card { background: #1e293b; border-radius: 12px; overflow: hidden; margin-bottom: 1.5rem; }
	       .card-header { padding: 1rem 1.5rem; background: #0f172a; border-bottom: 1px solid #334155; }
	       .card-title { font-size: 0.875rem; text-transform: uppercase; color: #64748b; }
	       .card-body { padding: 1.5rem; }
	       pre {
	           background: #0f172a; padding: 1rem; border-radius: 8px;
	           font-family: 'Monaco', monospace; font-size: 0.875rem;
	           overflow-x: auto; white-space: pre-wrap; word-break: break-all;
	       }
	       .btn {
	           display: inline-block; background: #3b82f6; color: white; border: none;
	           padding: 0.75rem 1.5rem; border-radius: 8px; font-size: 1rem;
	           cursor: pointer; text-decoration: none; margin-top: 1rem;
	       }
	       .btn:hover { background: #2563eb; }
	       .badge { background: #1e293b; padding: 0.25rem 0.75rem; border-radius: 9999px; font-size: 0.75rem; }
	   </style>
</head>
<body>
	   <div class="container">
	       <header>
	           <h1>%s</h1>
	           <p class="meta">Function: %s/%s | Version: %s | Duration: %dms</p>
	       </header>

	       <div class="card">
	           <div class="card-header"><span class="card-title">Input</span></div>
	           <div class="card-body"><pre>%s</pre></div>
	       </div>

	       <div class="card">
	           <div class="card-header"><span class="card-title">Output</span></div>
	           <div class="card-body"><pre>%s</pre></div>
	       </div>

	       <a href="/run/%s/%s?input=%s" class="btn">Re-run with same input</a>
	   </div>
</body>
</html>`,
		title, title, fn.Author, fn.Name, exec.Version, exec.DurationMs,
		inputJSON, outputJSON,
		fn.Author, fn.Name, url.QueryEscape(inputJSON))
}
