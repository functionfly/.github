package search

// ExecuteToolRequest represents a request to execute a search tool
type ExecuteToolRequest struct {
	ToolName    string                 `json:"tool_name" validate:"required"`
	Parameters  map[string]interface{} `json:"parameters" validate:"required"`
	EnableCache bool                   `json:"enable_cache"`
}

// ToolDefinition represents a search tool definition
type ToolDefinition struct {
	Name              string      `json:"name"`
	Category          string      `json:"category"`
	Description       string      `json:"description"`
	ParametersSchema  interface{} `json:"parameters_schema"`
	CostPerCall       float64     `json:"cost_per_call"`
	CostPerResult     float64     `json:"cost_per_result"`
	TimeoutMs         int         `json:"timeout_ms"`
}

// ExecuteToolResponse represents the response from executing a search tool
type ExecuteToolResponse struct {
	OK              bool        `json:"ok"`
	Cached          bool        `json:"cached"`
	Result          interface{} `json:"result"`
	ExecutionTimeMs int64       `json:"execution_time_ms"`
	CreditsUsed     float64     `json:"credits_used"`
	ResultsCount    int         `json:"results_count"`
	CachedAt        string      `json:"cached_at,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	OK    bool            `json:"ok"`
	Error ErrorDetail     `json:"error"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}