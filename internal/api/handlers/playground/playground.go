package playground

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Handler contains playground handlers
type Handler struct {
	repo storage.Repository
}

// NewHandler creates a new playground handler
func NewHandler(repo storage.Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

// PlaygroundInfo represents public function info for the playground
type PlaygroundInfo struct {
	FunctionID        uuid.UUID              `json:"function_id"`
	FunctionName      string                 `json:"function_name"`
	AppSlug           string                 `json:"app_slug"`
	Version           string                 `json:"version"`
	Status            string                 `json:"status"`
	PlaygroundEnabled bool                   `json:"playground_enabled"`
	PlaygroundConfig  map[string]interface{} `json:"playground_config,omitempty"`
	Provider          string                 `json:"provider,omitempty"`
	Region            string                 `json:"region,omitempty"`
	DeployedURL       string                 `json:"deployed_url,omitempty"`
}

// ExecuteRequest represents a playground execution request
type ExecuteRequest struct {
	Input interface{} `json:"input"`
}

// ExecuteResponse represents a playground execution response
type ExecuteResponse struct {
	Success    bool        `json:"success"`
	Output     interface{} `json:"output,omitempty"`
	Error      string      `json:"error,omitempty"`
	LatencyMs  int64       `json:"latency_ms"`
	StatusCode int         `json:"status_code"`
}

// HandlePlaygroundUI serves the playground UI HTML page
func (h *Handler) HandlePlaygroundUI(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appSlug := vars["appSlug"]
	functionName := vars["functionName"]

	// Get app by slug
	app, err := h.repo.GetAppBySlug(appSlug)
	if err != nil {
		http.Error(w, "App not found", http.StatusNotFound)
		return
	}

	// Get function by app ID and name
	function, err := h.repo.GetFunctionByAppIDAndName(r.Context(), app.ID, functionName)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	// Check if playground is enabled
	if !function.PlaygroundEnabled {
		http.Error(w, "Playground is not enabled for this function", http.StatusForbidden)
		return
	}

	// Get active deployment
	deployment, err := h.repo.GetActiveDeploymentForFunction(r.Context(), function.ID)
	if err != nil || deployment == nil || deployment.DeployedURL == nil {
		http.Error(w, "Function is not deployed", http.StatusBadRequest)
		return
	}

	// Serve the playground HTML
	html := h.generatePlaygroundHTML(function, app.Slug, deployment)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// HandleGetFunctionInfo returns public function info for the playground
func (h *Handler) HandleGetFunctionInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appSlug := vars["appSlug"]
	functionName := vars["functionName"]

	// Get app by slug
	app, err := h.repo.GetAppBySlug(appSlug)
	if err != nil {
		http.Error(w, "App not found", http.StatusNotFound)
		return
	}

	// Get function by app ID and name
	function, err := h.repo.GetFunctionByAppIDAndName(r.Context(), app.ID, functionName)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	// Check if playground is enabled
	if !function.PlaygroundEnabled {
		http.Error(w, "Playground is not enabled for this function", http.StatusForbidden)
		return
	}

	// Get active deployment
	deployment, err := h.repo.GetActiveDeploymentForFunction(r.Context(), function.ID)
	if err != nil || deployment == nil {
		// Return info even without deployment
		info := PlaygroundInfo{
			FunctionID:        function.ID,
			FunctionName:      function.Name,
			AppSlug:           app.Slug,
			Version:           function.Version,
			Status:            function.Status,
			PlaygroundEnabled: function.PlaygroundEnabled,
			PlaygroundConfig:  function.PlaygroundConfig,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
		return
	}

	info := PlaygroundInfo{
		FunctionID:        function.ID,
		FunctionName:      function.Name,
		AppSlug:           app.Slug,
		Version:           function.Version,
		Status:            function.Status,
		PlaygroundEnabled: function.PlaygroundEnabled,
		PlaygroundConfig:  function.PlaygroundConfig,
		Provider:          deployment.Provider,
		Region:            deployment.Region,
		DeployedURL:       *deployment.DeployedURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// HandleExecute executes a function with the provided input
func (h *Handler) HandleExecute(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appSlug := vars["appSlug"]
	functionName := vars["functionName"]

	// Get app by slug
	app, err := h.repo.GetAppBySlug(appSlug)
	if err != nil {
		http.Error(w, "App not found", http.StatusNotFound)
		return
	}

	// Get function by app ID and name
	function, err := h.repo.GetFunctionByAppIDAndName(r.Context(), app.ID, functionName)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	// Check if playground is enabled
	if !function.PlaygroundEnabled {
		http.Error(w, "Playground is not enabled for this function", http.StatusForbidden)
		return
	}

	// Get active deployment
	deployment, err := h.repo.GetActiveDeploymentForFunction(r.Context(), function.ID)
	if err != nil || deployment == nil || deployment.DeployedURL == nil {
		http.Error(w, "Function is not deployed", http.StatusBadRequest)
		return
	}

	// Parse request body
	var execReq ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&execReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Marshal input for forwarding
	inputBytes, err := json.Marshal(execReq.Input)
	if err != nil {
		inputBytes, _ = json.Marshal(map[string]interface{}{"data": execReq.Input})
	}

	// Execute function
	start := time.Now()
	targetURL := *deployment.DeployedURL + "/execute"

	proxyReq, err := http.NewRequestWithContext(r.Context(), "POST", targetURL, bytes.NewReader(inputBytes))
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
		response := ExecuteResponse{
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

	response := ExecuteResponse{
		Success:    resp.StatusCode >= 200 && resp.StatusCode < 300,
		Output:     output,
		LatencyMs:  latencyMs,
		StatusCode: resp.StatusCode,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// generatePlaygroundHTML generates the playground UI HTML
func (h *Handler) generatePlaygroundHTML(function *storage.FunctionConfig, appSlug string, deployment *storage.FunctionDeployment) string {
	title := function.Name
	if config, ok := function.PlaygroundConfig["title"].(string); ok {
		title = config
	}

	deployedURL := ""
	if deployment.DeployedURL != nil {
		deployedURL = *deployment.DeployedURL
	}

	fnName := function.Name
	fnStatus := function.Status

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - FunctionFly Playground</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: #0f172a; color: #e2e8f0; min-height: 100vh;
        }
        .container { max-width: 900px; margin: 0 auto; padding: 2rem; }
        header { text-align: center; margin-bottom: 2rem; }
        h1 { font-size: 1.75rem; margin-bottom: 0.5rem; color: #60a5fa; }
        .subtitle { color: #94a3b8; font-size: 0.9rem; }
        .playground { background: #1e293b; border-radius: 12px; overflow: hidden; }
        .section { padding: 1.5rem; border-bottom: 1px solid #334155; }
        .section:last-child { border-bottom: none; }
        .section-title { font-size: 0.85rem; text-transform: uppercase; letter-spacing: 0.05em; color: #64748b; margin-bottom: 1rem; }
        textarea {
            width: 100%%; min-height: 150px; background: #0f172a; border: 1px solid #334155;
            border-radius: 8px; padding: 1rem; color: #e2e8f0; font-family: 'Monaco', 'Consolas', monospace;
            font-size: 0.9rem; resize: vertical;
        }
        textarea:focus { outline: none; border-color: #60a5fa; }
        .btn {
            background: #3b82f6; color: white; border: none; padding: 0.75rem 1.5rem;
            border-radius: 8px; font-size: 1rem; cursor: pointer; transition: background 0.2s;
        }
        .btn:hover { background: #2563eb; }
        .btn:disabled { background: #475569; cursor: not-allowed; }
        .btn-row { display: flex; gap: 1rem; margin-top: 1rem; }
        .response {
            background: #0f172a; border: 1px solid #334155; border-radius: 8px;
            padding: 1rem; min-height: 100px; font-family: 'Monaco', 'Consolas', monospace;
            font-size: 0.9rem; white-space: pre-wrap; word-break: break-all;
        }
        .error { color: #f87171; }
        .success { color: #4ade80; }
        .meta { display: flex; gap: 1rem; font-size: 0.85rem; color: #64748b; margin-top: 0.5rem; }
        .share-btn { background: #475569; }
        .share-btn:hover { background: #64748b; }
        .status-badge {
            display: inline-block; padding: 0.25rem 0.75rem; border-radius: 9999px;
            font-size: 0.75rem; font-weight: 600;
        }
        .status-deployed { background: #166534; color: #4ade80; }
        .status-draft { background: #854d0e; color: #facc15; }
        .toast {
            position: fixed; bottom: 2rem; right: 2rem; background: #22c55e; color: white;
            padding: 0.75rem 1.5rem; border-radius: 8px; opacity: 0; transition: opacity 0.3s;
        }
        .toast.show { opacity: 1; }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>%s</h1>
            <div class="subtitle">Run directly in your browser</div>
            <div style="margin-top: 0.5rem;">
                <span class="status-badge status-%s">%s</span>
            </div>
        </header>

        <div class="playground">
            <div class="section">
                <div class="section-title">Input (JSON)</div>
                <textarea id="input" placeholder='{"text": "Hello, world!"}'></textarea>
                <div class="btn-row">
                    <button class="btn" id="runBtn" onclick="executeFunction()">Run Function</button>
                    <button class="btn share-btn" onclick="sharePlayground()">Copy Share Link</button>
                </div>
            </div>

            <div class="section">
                <div class="section-title">Output</div>
                <div id="output" class="response">Click "Run Function" to execute...</div>
                <div class="meta">
                    <span id="latency"></span>
                    <span id="status"></span>
                </div>
            </div>
        </div>
    </div>

    <div id="toast" class="toast">Link copied to clipboard!</div>

    <script>
        const appSlug = "%s";
        const functionName = "%s";
        const deployedURL = "%s";

        // Parse URL params for pre-filled input
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.has('input')) {
            document.getElementById('input').value = decodeURIComponent(urlParams.get('input'));
        }

        async function executeFunction() {
            const btn = document.getElementById('runBtn');
            const output = document.getElementById('output');
            const latency = document.getElementById('latency');
            const status = document.getElementById('status');

            btn.disabled = true;
            btn.textContent = 'Running...';
            output.innerHTML = '<span style="color: #60a5fa;">Executing function...</span>';
            latency.textContent = '';
            status.textContent = '';

            try {
                const inputVal = document.getElementById('input').value || '{}';
                let input;
                try {
                    input = JSON.parse(inputVal);
                } catch(e) {
                    input = { data: inputVal };
                }

                const start = Date.now();
                const response = await fetch('/v1/run/' + appSlug + '/' + functionName + '/execute', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ input })
                });
                const result = await response.json();
                const elapsed = Date.now() - start;

                latency.textContent = 'Latency: ' + elapsed + 'ms';
                status.textContent = 'Status: ' + result.status_code;

                if (result.success) {
                    output.innerHTML = '<span class="success">Success</span>\n\n' + JSON.stringify(result.output, null, 2);
                } else {
                    output.innerHTML = '<span class="error">Error: ' + (result.error || 'Unknown error') + '</span>';
                    if (result.output) {
                        output.innerHTML += '\n\n' + JSON.stringify(result.output, null, 2);
                    }
                }
            } catch (err) {
                output.innerHTML = '<span class="error">Request failed: ' + err.message + '</span>';
            }

            btn.disabled = false;
            btn.textContent = 'Run Function';
        }

        function sharePlayground() {
            const input = document.getElementById('input').value;
            const url = new URL(window.location.href);
            if (input) {
                url.searchParams.set('input', input);
            }
            navigator.clipboard.writeText(url.toString()).then(() => {
                const toast = document.getElementById('toast');
                toast.classList.add('show');
                setTimeout(() => toast.classList.remove('show'), 2000);
            });
        }
    </script>
</body>
</html>`,
		title, title, fnStatus, fnStatus, appSlug, fnName, deployedURL)
}
