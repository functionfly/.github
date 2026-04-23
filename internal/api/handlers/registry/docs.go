package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DocumentationHandler handles API documentation generation and serving
type DocumentationHandler struct {
	repo *registry.RegistryRepository
}

// NewDocumentationHandler creates a new documentation handler
func NewDocumentationHandler(repo *registry.RegistryRepository) *DocumentationHandler {
	return &DocumentationHandler{repo: repo}
}

// OpenAPISpec represents an OpenAPI 3.0 specification
type OpenAPISpec struct {
	OpenAPI    string                                 `json:"openapi"`
	Info       OpenAPIInfo                            `json:"info"`
	Servers    []OpenAPIServer                        `json:"servers,omitempty"`
	Paths      map[string]map[string]OpenAPIOperation `json:"paths"`
	Components OpenAPIComponents                      `json:"components,omitempty"`
}

type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

type OpenAPIOperation struct {
	Summary     string                     `json:"summary,omitempty"`
	Description string                     `json:"description,omitempty"`
	OperationID string                     `json:"operationId,omitempty"`
	Tags        []string                   `json:"tags,omitempty"`
	Parameters  []OpenAPIParameter         `json:"parameters,omitempty"`
	RequestBody *OpenAPIRequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]OpenAPIResponse `json:"responses"`
}

type OpenAPIParameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"`
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Schema      interface{} `json:"schema"`
}

type OpenAPIRequestBody struct {
	Description string                      `json:"description,omitempty"`
	Required    bool                        `json:"required,omitempty"`
	Content     map[string]OpenAPIMediaType `json:"content"`
}

type OpenAPIMediaType struct {
	Schema interface{} `json:"schema"`
}

type OpenAPIResponse struct {
	Description string                      `json:"description"`
	Content     map[string]OpenAPIMediaType `json:"content,omitempty"`
}

type OpenAPIComponents struct {
	Schemas map[string]interface{} `json:"schemas,omitempty"`
}

// HandleOpenAPISpec generates and serves OpenAPI specification
func (h *DocumentationHandler) HandleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	spec, err := h.generateOpenAPISpec(r)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate OpenAPI spec")
		http.Error(w, "Failed to generate OpenAPI specification", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	middleware.SetCORSHeaders(w, r)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	json.NewEncoder(w).Encode(spec)
}

// HandleFunctionDocs serves documentation for a specific function
func (h *DocumentationHandler) HandleFunctionDocs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			http.Error(w, "Function not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to get function")
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	fnVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		http.Error(w, "Function version not found", http.StatusNotFound)
		return
	}

	// Get rating for trust information
	rating, _ := h.repo.GetRatingByFunctionID(fn.ID)

	// Get recent executions for examples
	executions, err := h.repo.GetRecentPublicExecutions(fn.ID, 5)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get recent executions")
	}

	docs := h.generateFunctionDocs(fn, fnVersion, rating, executions)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

// HandleFunctionHTMLDocs serves HTML documentation for a specific function
func (h *DocumentationHandler) HandleFunctionHTMLDocs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			http.Error(w, "Function not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to get function")
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	fnVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		http.Error(w, "Function version not found", http.StatusNotFound)
		return
	}

	// Get rating for trust information
	rating, _ := h.repo.GetRatingByFunctionID(fn.ID)

	// Get recent executions for examples
	executions, err := h.repo.GetRecentPublicExecutions(fn.ID, 5)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get recent executions")
	}

	html := h.generateFunctionHTMLDocs(fn, fnVersion, rating, executions)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// HandleIndex serves the main documentation index
func (h *DocumentationHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	functions, _, err := h.repo.ListFunctions("", "", nil, "public", 100, 0)
	if err != nil {
		logrus.WithError(err).Error("Failed to list functions")
		http.Error(w, "Failed to list functions", http.StatusInternalServerError)
		return
	}

	// Group functions by category
	categories := make(map[string][]FunctionDocSummary)
	for _, fn := range functions {
		fnVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
		if err != nil {
			continue
		}

		rating, _ := h.repo.GetRatingByFunctionID(fn.ID)

		summary := FunctionDocSummary{
			Author:      fn.Author,
			Name:        fn.Name,
			Title:       fn.Title.String,
			Description: fn.Description.String,
			Category:    fn.Category.String,
			Version:     fnVersion.Version,
			TrustScore:  rating.TrustScore,
		}

		cat := fn.Category.String
		if cat == "" {
			cat = "Uncategorized"
		}
		categories[cat] = append(categories[cat], summary)
	}

	html := h.generateIndexHTML(categories)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// generateOpenAPISpec creates an OpenAPI specification from all public functions
func (h *DocumentationHandler) generateOpenAPISpec(r *http.Request) (*OpenAPISpec, error) {
	functions, _, err := h.repo.ListFunctions("", "", nil, "public", 1000, 0)
	if err != nil {
		return nil, err
	}

	spec := &OpenAPISpec{
		OpenAPI: "3.0.3",
		Info: OpenAPIInfo{
			Title:       "FunctionFly Registry API",
			Description: "Execute serverless functions with FunctionFly",
			Version:     "1.0.0",
		},
		Servers: []OpenAPIServer{
			{
				URL:         fmt.Sprintf("%s://%s", getScheme(r), r.Host),
				Description: "FunctionFly Registry",
			},
		},
		Paths:      make(map[string]map[string]OpenAPIOperation),
		Components: OpenAPIComponents{Schemas: make(map[string]interface{})},
	}

	for _, fn := range functions {
		fnVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
		if err != nil {
			continue
		}

		path := fmt.Sprintf("/execute/%s/%s", fn.Author, fn.Name)
		if spec.Paths[path] == nil {
			spec.Paths[path] = make(map[string]OpenAPIOperation)
		}

		operation := h.generateFunctionOperation(&fn, fnVersion)
		spec.Paths[path]["post"] = operation

		// Add schemas to components
		h.addFunctionSchemas(spec, &fn, fnVersion)
	}

	return spec, nil
}

// generateFunctionOperation creates an OpenAPI operation for a function
func (h *DocumentationHandler) generateFunctionOperation(fn *registry.RegistryFunction, fnVersion *registry.RegistryFunctionVersion) OpenAPIOperation {
	var manifest functionregistry.FunctionManifest
	if err := json.Unmarshal(fnVersion.Manifest, &manifest); err != nil {
		return OpenAPIOperation{}
	}

	operation := OpenAPIOperation{
		Summary:     fmt.Sprintf("Execute %s", fn.Title.String),
		Description: fn.Description.String,
		OperationID: fmt.Sprintf("execute%s%s", strings.Title(fn.Author), strings.Title(fn.Name)),
		Tags:        []string{fn.Category.String},
		Responses: map[string]OpenAPIResponse{
			"200": {
				Description: "Successful execution",
				Content: map[string]OpenAPIMediaType{
					"application/json": {
						Schema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"ok": map[string]interface{}{
									"type":    "boolean",
									"example": true,
								},
								"data": map[string]interface{}{
									"description": "Function output",
								},
								"cached": map[string]interface{}{
									"type":    "boolean",
									"example": false,
								},
								"duration_ms": map[string]interface{}{
									"type":    "integer",
									"example": 150,
								},
								"version": map[string]interface{}{
									"type":    "string",
									"example": fnVersion.Version,
								},
							},
							"required": []string{"ok", "data", "cached", "duration_ms", "version"},
						},
					},
				},
			},
			"400": {
				Description: "Bad Request",
				Content: map[string]OpenAPIMediaType{
					"application/json": {
						Schema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"ok": map[string]interface{}{
									"type":    "boolean",
									"example": false,
								},
								"error": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"code": map[string]interface{}{
											"type":    "string",
											"example": "INVALID_INPUT",
										},
										"message": map[string]interface{}{
											"type":    "string",
											"example": "Invalid input parameters",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Add request body if input is defined
	if manifest.Input != nil {
		requestBody := &OpenAPIRequestBody{
			Description: "Function input parameters",
			Required:    manifest.Input.Required.IsRequired(),
			Content: map[string]OpenAPIMediaType{
				"application/json": {
					Schema: h.generateSchemaFromIOType(manifest.Input),
				},
			},
		}
		operation.RequestBody = requestBody
	}

	return operation
}

// addFunctionSchemas adds function input/output schemas to OpenAPI components
func (h *DocumentationHandler) addFunctionSchemas(spec *OpenAPISpec, fn *registry.RegistryFunction, fnVersion *registry.RegistryFunctionVersion) {
	var manifest functionregistry.FunctionManifest
	if err := json.Unmarshal(fnVersion.Manifest, &manifest); err != nil {
		return
	}

	schemaName := fmt.Sprintf("%s%s", strings.Title(fn.Author), strings.Title(fn.Name))

	if manifest.Input != nil {
		inputSchema := h.generateSchemaFromIOType(manifest.Input)
		spec.Components.Schemas[fmt.Sprintf("%sInput", schemaName)] = inputSchema
	}

	if manifest.Output != nil {
		outputSchema := h.generateSchemaFromIOType(manifest.Output)
		spec.Components.Schemas[fmt.Sprintf("%sOutput", schemaName)] = outputSchema
	}
}

// generateSchemaFromIOType converts IOType to OpenAPI schema
func (h *DocumentationHandler) generateSchemaFromIOType(ioType *functionregistry.IOType) interface{} {
	if ioType == nil {
		return map[string]interface{}{
			"type": "object",
		}
	}

	schema := make(map[string]interface{})

	// Parse schema if provided
	if ioType.Schema != nil {
		var schemaData interface{}
		if err := json.Unmarshal(ioType.Schema, &schemaData); err == nil {
			return schemaData
		}
	}

	// Fallback to basic type
	if ioType.Type != "" {
		schema["type"] = ioType.Type
	} else {
		schema["type"] = "object"
	}

	// Add example if provided
	if ioType.Example != nil {
		schema["example"] = ioType.Example
	}

	return schema
}

// FunctionDocSummary represents a summary of function documentation
type FunctionDocSummary struct {
	Author      string  `json:"author"`
	Name        string  `json:"name"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Version     string  `json:"version"`
	TrustScore  float64 `json:"trust_score"`
}

// FunctionDocs represents detailed function documentation
type FunctionDocs struct {
	Function     FunctionDocSummary                `json:"function"`
	Manifest     functionregistry.FunctionManifest `json:"manifest"`
	Runtime      string                            `json:"runtime"`
	TrustScore   float64                           `json:"trust_score"`
	SuccessRate  float64                           `json:"success_rate"`
	AvgLatency   int                               `json:"avg_latency_ms"`
	Examples     []ExecutionExample                `json:"examples,omitempty"`
	Capabilities []string                          `json:"capabilities,omitempty"`
}

// ExecutionExample represents an execution example
type ExecutionExample struct {
	Input      interface{} `json:"input"`
	Output     interface{} `json:"output"`
	Cached     bool        `json:"cached"`
	DurationMs int         `json:"duration_ms"`
}

// generateFunctionDocs creates detailed documentation for a function
func (h *DocumentationHandler) generateFunctionDocs(fn *registry.RegistryFunction, fnVersion *registry.RegistryFunctionVersion, rating *registry.RegistryFunctionRating, executions []registry.RegistryExecutionPublic) FunctionDocs {
	var manifest functionregistry.FunctionManifest
	json.Unmarshal(fnVersion.Manifest, &manifest)

	// Extract capabilities
	var capabilities []string
	json.Unmarshal(fnVersion.Capabilities, &capabilities)

	// Create examples from executions
	var examples []ExecutionExample
	for _, exec := range executions {
		if exec.Shareable {
			var input, output interface{}
			json.Unmarshal(exec.InputJSON, &input)
			json.Unmarshal(exec.OutputJSON, &output)

			examples = append(examples, ExecutionExample{
				Input:      input,
				Output:     output,
				Cached:     exec.Cached,
				DurationMs: exec.DurationMs,
			})
		}
	}

	return FunctionDocs{
		Function: FunctionDocSummary{
			Author:      fn.Author,
			Name:        fn.Name,
			Title:       fn.Title.String,
			Description: fn.Description.String,
			Category:    fn.Category.String,
			Version:     fnVersion.Version,
			TrustScore:  rating.TrustScore,
		},
		Manifest:     manifest,
		Runtime:      fnVersion.Runtime,
		TrustScore:   rating.TrustScore,
		SuccessRate:  rating.SuccessRate,
		AvgLatency:   rating.AvgLatencyMs,
		Examples:     examples,
		Capabilities: capabilities,
	}
}

// generateFunctionHTMLDocs creates HTML documentation for a function
func (h *DocumentationHandler) generateFunctionHTMLDocs(fn *registry.RegistryFunction, fnVersion *registry.RegistryFunctionVersion, rating *registry.RegistryFunctionRating, executions []registry.RegistryExecutionPublic) string {
	docs := h.generateFunctionDocs(fn, fnVersion, rating, executions)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - FunctionFly Documentation</title>
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
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            padding: 40px;
            border-radius: 12px;
            margin-bottom: 30px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .title {
            font-size: 2.5em;
            margin: 0 0 10px 0;
            font-weight: 300;
        }
        .subtitle {
            font-size: 1.2em;
            opacity: 0.9;
            margin: 0;
        }
        .meta {
            display: flex;
            gap: 20px;
            margin-top: 20px;
            flex-wrap: wrap;
        }
        .meta-item {
            background: rgba(255,255,255,0.1);
            padding: 8px 16px;
            border-radius: 20px;
            font-size: 0.9em;
        }
        .section {
            background: white;
            padding: 30px;
            border-radius: 12px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.05);
        }
        .section-title {
            font-size: 1.8em;
            margin-bottom: 20px;
            color: #2c3e50;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
        }
        .code-block {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 8px;
            padding: 20px;
            margin: 20px 0;
            overflow-x: auto;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.9em;
        }
        .example {
            background: #f8f9fa;
            border-left: 4px solid #3498db;
            padding: 20px;
            margin: 20px 0;
            border-radius: 0 8px 8px 0;
        }
        .example-title {
            font-weight: bold;
            margin-bottom: 10px;
            color: #2c3e50;
        }
        .json-output {
            white-space: pre-wrap;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.85em;
            background: #2d3748;
            color: #e2e8f0;
            padding: 15px;
            border-radius: 6px;
            overflow-x: auto;
        }
        .badge {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 0.8em;
            font-weight: 500;
            margin: 2px;
        }
        .badge-trust { background: #48bb78; color: white; }
        .badge-runtime { background: #4299e1; color: white; }
        .badge-capability { background: #ed8936; color: white; }
        .try-it {
            background: #3498db;
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 1em;
            margin: 20px 0;
            transition: background 0.3s;
        }
        .try-it:hover {
            background: #2980b9;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1 class="title">%s</h1>
        <p class="subtitle">%s</p>
        <div class="meta">
            <span class="meta-item">Author: %s</span>
            <span class="meta-item">Version: %s</span>
            <span class="meta-item">Runtime: %s</span>
            <span class="meta-item">Trust Score: %.1f</span>
        </div>
    </div>

    <div class="section">
        <h2 class="section-title">Overview</h2>
        <p>%s</p>
        <div>
            <span class="badge badge-runtime">%s</span>
            <span class="badge badge-trust">Trust Score: %.1f</span>
        </div>
    </div>

    <div class="section">
        <h2 class="section-title">API Usage</h2>
        <p>Execute this function using the following API endpoint:</p>
        <div class="code-block">
curl -X POST %s/v1/fx/%s/%s \\
  -H "Content-Type: application/json" \\
  -d '{"input": "your-input-here"}'
        </div>
    </div>

    <div class="section">
        <h2 class="section-title">Examples</h2>`,
		docs.Function.Title,
		docs.Function.Title,
		docs.Function.Description,
		docs.Function.Author,
		docs.Function.Version,
		docs.Runtime,
		docs.TrustScore,
		docs.Function.Description,
		docs.Runtime,
		docs.TrustScore,
		getAPIBase(), docs.Function.Author, docs.Function.Name)

	if len(docs.Examples) > 0 {
		for i, example := range docs.Examples {
			inputJSON, _ := json.MarshalIndent(example.Input, "", "  ")
			outputJSON, _ := json.MarshalIndent(example.Output, "", "  ")
			html += fmt.Sprintf(`
        <div class="example">
            <div class="example-title">Example %d (%dms)</div>
            <p><strong>Input:</strong></p>
            <div class="json-output">%s</div>
            <p><strong>Output:</strong></p>
            <div class="json-output">%s</div>
        </div>`, i+1, example.DurationMs, string(inputJSON), string(outputJSON))
		}
	} else {
		html += `<p>No execution examples available yet.</p>`
	}

	html += `
    </div>

    <div class="section">
        <h2 class="section-title">Technical Details</h2>
        <ul>
            <li><strong>Runtime:</strong> ` + docs.Runtime + `</li>
            <li><strong>Success Rate:</strong> ` + fmt.Sprintf("%.1f%%", docs.SuccessRate*100) + `</li>
            <li><strong>Average Latency:</strong> ` + fmt.Sprintf("%dms", docs.AvgLatency) + `</li>
            <li><strong>Timeout:</strong> ` + fmt.Sprintf("%dms", docs.Manifest.TimeoutMs) + `</li>
            <li><strong>Memory Limit:</strong> ` + fmt.Sprintf("%dMB", docs.Manifest.MemoryMB) + `</li>
        </ul>
    </div>`

	if len(docs.Capabilities) > 0 {
		html += `
    <div class="section">
        <h2 class="section-title">Capabilities</h2>
        <p>This function has the following capabilities:</p>
        <div>`
		for _, cap := range docs.Capabilities {
			html += fmt.Sprintf(`<span class="badge badge-capability">%s</span>`, cap)
		}
		html += `
        </div>
    </div>`
	}

	html += `
    <div class="section">
        <h2 class="section-title">Try It Out</h2>
        <p>Test this function interactively:</p>
        <button class="try-it" onclick="window.open('/playground/` + docs.Function.Author + `/` + docs.Function.Name + `', '_blank')">Open in Playground</button>
        <br><br>
        <p>Get the SDK for programmatic access:</p>
        <button class="try-it" onclick="window.open('/sdk/js/latest/functionfly.js', '_blank')" style="background: #27ae60;">Download JavaScript SDK</button>
    </div>
</body>
</html>`

	return html
}

// generateIndexHTML creates the main documentation index page
func (h *DocumentationHandler) generateIndexHTML(categories map[string][]FunctionDocSummary) string {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>FunctionFly Registry Documentation</title>
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
            margin-bottom: 30px;
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
        .search-bar {
            width: 100%;
            max-width: 600px;
            padding: 15px;
            font-size: 1.1em;
            border: none;
            border-radius: 8px;
            margin: 20px auto;
            display: block;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .category {
            background: white;
            padding: 30px;
            border-radius: 12px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.05);
        }
        .category-title {
            font-size: 1.8em;
            margin-bottom: 20px;
            color: #2c3e50;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
        }
        .function-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
            gap: 20px;
        }
        .function-card {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            border: 1px solid #e9ecef;
            transition: transform 0.2s, box-shadow 0.2s;
            text-decoration: none;
            color: inherit;
            display: block;
        }
        .function-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 8px rgba(0,0,0,0.1);
            text-decoration: none;
            color: inherit;
        }
        .function-title {
            font-size: 1.2em;
            font-weight: 600;
            margin: 0 0 10px 0;
            color: #2c3e50;
        }
        .function-author {
            color: #666;
            font-size: 0.9em;
            margin: 0 0 10px 0;
        }
        .function-description {
            color: #555;
            margin: 0 0 15px 0;
            display: -webkit-box;
            -webkit-line-clamp: 3;
            -webkit-box-orient: vertical;
            overflow: hidden;
        }
        .function-meta {
            display: flex;
            justify-content: space-between;
            align-items: center;
            font-size: 0.85em;
        }
        .trust-score {
            color: #48bb78;
            font-weight: 500;
        }
        .version {
            color: #666;
        }
        .footer {
            text-align: center;
            padding: 40px 20px;
            color: #666;
        }
        .nav-links {
            display: flex;
            justify-content: center;
            gap: 20px;
            margin: 20px 0;
        }
        .nav-link {
            color: #3498db;
            text-decoration: none;
            padding: 8px 16px;
            border: 1px solid #3498db;
            border-radius: 6px;
            transition: all 0.3s;
        }
        .nav-link:hover {
            background: #3498db;
            color: white;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1 class="title">FunctionFly Registry</h1>
        <p class="subtitle">Discover and execute serverless functions</p>
        <input type="text" class="search-bar" placeholder="Search functions..." onkeyup="filterFunctions(this.value)">
    </div>

    <div class="nav-links">
        <a href="/docs/openapi.json" class="nav-link">OpenAPI Spec</a>
        <a href="/playground" class="nav-link">Playground</a>
        <a href="/tutorials" class="nav-link">Tutorials</a>
        <a href="/sdk/js/latest/functionfly.js" class="nav-link">JavaScript SDK</a>
    </div>`

	// Sort categories
	var sortedCats []string
	for cat := range categories {
		sortedCats = append(sortedCats, cat)
	}
	sort.Strings(sortedCats)

	for _, cat := range sortedCats {
		functions := categories[cat]
		if len(functions) == 0 {
			continue
		}

		html += fmt.Sprintf(`
    <div class="category" data-category="%s">
        <h2 class="category-title">%s</h2>
        <div class="function-grid">`, cat, cat)

		for _, fn := range functions {
			html += fmt.Sprintf(`
            <a href="/docs/%s/%s" class="function-card">
                <h3 class="function-title">%s</h3>
                <p class="function-author">by %s</p>
                <p class="function-description">%s</p>
                <div class="function-meta">
                    <span class="trust-score">Trust: %.1f</span>
                    <span class="version">v%s</span>
                </div>
            </a>`, fn.Author, fn.Name, fn.Title, fn.Author, fn.Description, fn.TrustScore, fn.Version)
		}

		html += `
        </div>
    </div>`
	}

	html += `
    <div class="footer">
        <p>Built with FunctionFly • <a href="/docs/openapi.json">API Documentation</a> • <a href="/playground">Interactive Playground</a></p>
    </div>

    <script>
        function filterFunctions(query) {
            const cards = document.querySelectorAll('.function-card');
            const categories = document.querySelectorAll('.category');

            query = query.toLowerCase();

            cards.forEach(card => {
                const title = card.querySelector('.function-title').textContent.toLowerCase();
                const author = card.querySelector('.function-author').textContent.toLowerCase();
                const desc = card.querySelector('.function-description').textContent.toLowerCase();

                if (title.includes(query) || author.includes(query) || desc.includes(query)) {
                    card.style.display = 'block';
                } else {
                    card.style.display = 'none';
                }
            });

            // Hide empty categories
            categories.forEach(cat => {
                const visibleCards = cat.querySelectorAll('.function-card[style*="block"]');
                if (visibleCards.length === 0) {
                    cat.style.display = 'none';
                } else {
                    cat.style.display = 'block';
                }
            });
        }
    </script>
</body>
</html>`

	return html
}

// getScheme returns the scheme (http/https) from the request
func getScheme(r *http.Request) string {
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		return "https"
	}
	return "http"
}

// FunctionVersionResponse represents a function version for API response
type FunctionVersionResponse struct {
	Version     string          `json:"version"`
	Runtime     string          `json:"runtime"`
	PublishedAt string          `json:"published_at"`
	Manifest    json.RawMessage `json:"manifest,omitempty"`
}

// HandleFunctionVersions serves all versions of a function
func (h *DocumentationHandler) HandleFunctionVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			http.Error(w, "Function not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to get function")
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	versions, err := h.repo.ListFunctionVersions(fn.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list function versions")
		http.Error(w, "Failed to list versions", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	response := make([]FunctionVersionResponse, len(versions))
	for i, v := range versions {
		response[i] = FunctionVersionResponse{
			Version:     v.Version,
			Runtime:     v.Runtime,
			PublishedAt: v.PublishedAt.Format("2006-01-02T15:04:05Z07:00"),
			Manifest:    v.Manifest,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
