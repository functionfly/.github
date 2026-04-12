package registry

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// writeError writes a structured error response
func (h *Handler) writeError(w http.ResponseWriter, status int, code, message string) {
	errResp := functionregistry.ExecutionError{
		OK: false,
		Error: functionregistry.ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errResp)
}

// executeOnBackend executes a function on a backend service
func executeOnBackend(backendURL, input string, timeoutMs int) (json.RawMessage, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(timeoutMs) * time.Millisecond,
	}

	// Prepare request body
	requestBody := map[string]interface{}{
		"input": input,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create request with context for timeout control
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", backendURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "FunctionFly-Registry/1.0")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute on backend: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("backend returned status %d", resp.StatusCode)
	}

	// Read response body
	var response json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode backend response: %w", err)
	}

	return response, nil
}

// nullStringToPtr converts a null string to a pointer
func nullStringToPtr(s *string) *string {
	if s == nil {
		return nil
	}
	return s
}

// toNullString converts a string pointer to sql.NullString
func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// convertToFunctionInfos converts map representations to FunctionInfo structs
func convertToFunctionInfos(infos []map[string]interface{}) []functionregistry.FunctionInfo {
	result := make([]functionregistry.FunctionInfo, 0, len(infos))
	for _, info := range infos {
		fi := functionregistry.FunctionInfo{
			Author:       info["author"].(string),
			Name:         info["name"].(string),
			Version:      info["version"].(string),
			PricePerCall: info["price_per_call"].(float64),
		}
		// ID (required for gallery)
		if v, ok := info["id"].(string); ok {
			fi.ID = v
		}
		if v, ok := info["title"].(string); ok {
			fi.Title = v
		}
		if v, ok := info["description"].(string); ok {
			fi.Description = v
		}
		if v, ok := info["runtime"].(string); ok {
			fi.Runtime = v
		}
		if v, ok := info["category"].(string); ok {
			fi.Category = v
		}
		if v, ok := info["reliability"].(float64); ok {
			fi.Reliability = v
		}
		// Tags
		if v, ok := info["tags"].([]interface{}); ok {
			var tags []string
			for _, tag := range v {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
			fi.Tags = tags
		}
		// Deterministic
		if v, ok := info["deterministic"].(bool); ok {
			fi.Deterministic = v
		}
		// SideEffects
		if v, ok := info["side_effects"].(string); ok {
			fi.SideEffects = v
		}
		// Idempotent
		if v, ok := info["idempotent"].(bool); ok {
			fi.Idempotent = v
		}
		// CacheTTL
		if v, ok := info["cache_ttl"].(int); ok {
			fi.CacheTTL = v
		}
		// TimeoutMs
		if v, ok := info["timeout_ms"].(int); ok {
			fi.TimeoutMs = v
		}
		// MemoryMB
		if v, ok := info["memory_mb"].(int); ok {
			fi.MemoryMB = v
		}
		// Capabilities
		if v, ok := info["capabilities"].([]interface{}); ok {
			var caps []string
			for _, cap := range v {
				if capStr, ok := cap.(string); ok {
					caps = append(caps, capStr)
				}
			}
			fi.Capabilities = caps
		}
		// BundleSize
		if v, ok := info["bundle_size"].(int); ok {
			fi.BundleSize = v
		} else if v, ok := info["bundle_size"].(int32); ok {
			fi.BundleSize = int(v)
		}
		// SourceHash
		if v, ok := info["source_hash"].(string); ok {
			fi.SourceHash = v
		}
		// DeploymentID
		if v, ok := info["deployment_id"].(string); ok {
			fi.DeploymentID = v
		}
		// BackendID
		if v, ok := info["backend_id"].(string); ok {
			fi.BackendID = v
		}
		// Input/Output types
		if v, ok := info["input_type"].(string); ok {
			fi.InputType = v
		}
		if v, ok := info["output_type"].(string); ok {
			fi.OutputType = v
		}
		// Input/Output examples
		if v, ok := info["input_example"]; ok && v != nil {
			if exampleBytes, err := json.Marshal(v); err == nil {
				fi.InputExample = exampleBytes
			}
		}
		if v, ok := info["output_example"]; ok && v != nil {
			if exampleBytes, err := json.Marshal(v); err == nil {
				fi.OutputExample = exampleBytes
			}
		}
		// Documentation and Playground URLs
		if v, ok := info["documentation_url"].(string); ok {
			fi.DocumentationURL = v
		}
		if v, ok := info["playground_url"].(string); ok {
			fi.PlaygroundURL = v
		}
		// Trust score fields
		if v, ok := info["trust_score"].(float64); ok {
			fi.TrustScore = v
		}
		if v, ok := info["trust_level"].(string); ok {
			fi.TrustLevel = v
		}
		if v, ok := info["success_rate"].(float64); ok {
			fi.SuccessRate = v
		}
		if v, ok := info["p50_latency_ms"].(int); ok {
			fi.P50LatencyMs = v
		}
		if v, ok := info["p95_latency_ms"].(int); ok {
			fi.P95LatencyMs = v
		}
		if v, ok := info["timeout_rate"].(float64); ok {
			fi.TimeoutRate = v
		}
		if v, ok := info["error_rate"].(float64); ok {
			fi.ErrorRate = v
		}
		if v, ok := info["consumer_diversity"].(float64); ok {
			fi.ConsumerDiversity = v
		}
		if v, ok := info["tenant_diversity"].(int); ok {
			fi.TenantDiversity = v
		}
		if v, ok := info["user_diversity"].(int); ok {
			fi.UserDiversity = v
		}
		// Gallery-specific fields
		if v, ok := info["popularity_score"].(int); ok {
			fi.PopularityScore = v
		} else if v, ok := info["popularity_score"].(float64); ok {
			fi.PopularityScore = int(v)
		}
		if v, ok := info["remix_count"].(int); ok {
			fi.RemixCount = v
		}
		if v, ok := info["like_count"].(int); ok {
			fi.LikeCount = v
		}
		if v, ok := info["created_at"].(string); ok {
			fi.CreatedAt = v
		}
		if v, ok := info["updated_at"].(string); ok {
			fi.UpdatedAt = v
		}
		result = append(result, fi)
	}
	return result
}

// executeLocally executes a function locally using the sandbox runtime
func executeLocally(fnVersion *storage.RegistryFunctionVersion, input []byte) (json.RawMessage, error) {
	// Create sandbox executor
	executor, err := NewSandboxExecutor()
	if err != nil {
		logrus.WithError(err).Error("Failed to create sandbox executor")
		return nil, fmt.Errorf("sandbox initialization failed: %w", err)
	}
	defer executor.Close()

	// Execute the function
	output, err := executor.ExecuteFunction(fnVersion, input, fnVersion.TimeoutMs)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"function_id": fnVersion.FunctionID,
			"version":     fnVersion.Version,
			"runtime":     fnVersion.Runtime,
		}).Error("Sandbox execution failed")

		return nil, fmt.Errorf("execution failed: %w", err)
	}

	// Return the raw output from the WASM execution
	// The output should already be JSON from the WASM module
	return output, nil
}
