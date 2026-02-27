package registry

import (
	"fmt"
	"os"
	"strings"
)

// getBaseURL returns the base URL for API calls, configurable via environment
func getBaseURL() string {
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		// Remove trailing slash if present
		baseURL = strings.TrimSuffix(baseURL, "/")
		return baseURL + "/v1/fx"
	}
	// Default to production URL
	return "https://api.functionfly.com/v1/fx"
}

// SDK generation functions (production-ready implementations)
func generateJavaScriptSDK(author, name, version, title, description string) string {
	baseURL := getBaseURL()
	return fmt.Sprintf(`/**
 * FunctionFly SDK for %s/%s v%s
 * %s
 *
 *
 * Generated SDK for executing serverless functions on FunctionFly.
 * Supports async/await and proper error handling.
 */

class FunctionFlyError extends Error {
  constructor(message, code, statusCode) {
    super(message);
    this.name = 'FunctionFlyError';
    this.code = code;
    this.statusCode = statusCode;
  }
}

/**
 * Execute %s function
 * @param {Object} input - Input parameters for the function
 * @param {Object} options - Execution options
 * @param {string} options.version - Specific version to execute (optional)
 * @param {string} options.apiKey - API key for authentication (optional)
 * @returns {Promise<Object>} Execution result with ok, data, cached, duration_ms, version
 */
const execute%s%s = async (input, options = {}) => {
  try {
    const url = options.version && options.version !== 'latest'
      ? '%s/%s/%s@' + options.version
      : '%s/%s/%s';

    const headers = {
      'Content-Type': 'application/json',
      'User-Agent': 'FunctionFly-JS-SDK/1.0',
    };

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
      throw new FunctionFlyError(
        errorData.error?.message || 'HTTP ' + response.status + ': ' + response.statusText,
        errorData.error?.code || 'EXECUTION_FAILED',
        response.status
      );
    }

    const result = await response.json();

    if (!result.ok) {
      throw new FunctionFlyError(
        result.error?.message || 'Function execution failed',
        result.error?.code || 'EXECUTION_FAILED',
        500
      );
    }

    return result;
  } catch (error) {
    if (error instanceof FunctionFlyError) {
      throw error;
    }
    throw new FunctionFlyError(
      error.message || 'Network or parsing error',
      'NETWORK_ERROR',
      0
    );
  }
};

/**
 * Execute function synchronously (convenience method)
 * @param {Object} input - Input parameters
 * @param {Object} options - Execution options
 * @returns {Promise<Object>} Execution data
 */
execute%s%s.sync = async (input, options = {}) => {
  const result = await execute%s%s(input, options);
  return result.data;
};

export { execute%s%s, FunctionFlyError };`,
		author, name, version, description, // 4
		name,                                       // 5
		strings.Title(author), strings.Title(name), // 6-7
		baseURL, author, name, // 8-10
		baseURL, author, name, // 11-13
		strings.Title(author), strings.Title(name), // 14-15
		strings.Title(author), strings.Title(name), // 16-17
		strings.Title(author), strings.Title(name)) // 18
}

func generatePythonSDK(author, name, version, title, description string) string {
	baseURL := getBaseURL()
	return fmt.Sprintf(`"""
FunctionFly SDK for %s/%s v%s
%s

Generated SDK for executing serverless functions on FunctionFly.
Supports proper error handling and response parsing.
"""

import requests
from typing import Dict, Any, Optional, Union
import json


class FunctionFlyError(Exception):
    """Exception raised for FunctionFly execution errors"""

    def __init__(self, message: str, code: str = None, status_code: int = None):
        super().__init__(message)
        self.code = code or "EXECUTION_FAILED"
        self.status_code = status_code


def execute_%s_%s(input_data: Dict[str, Any], version: str = "latest", api_key: str = None) -> Dict[str, Any]:
    """
    Execute %s function

    Args:
        input_data (dict): Input parameters for the function
        version (str): Function version to execute (default: "latest")
        api_key (str): API key for authentication (optional)

    Returns:
        dict: Execution result with keys: ok, data, cached, duration_ms, version, execution_id

    Raises:
        FunctionFlyError: If execution fails
        requests.RequestException: If network request fails
    """
    try:
        url = f"%s/%s/%s"
        if version and version != "latest":
            url += f"@{version}"

        headers = {
            "Content-Type": "application/json",
            "User-Agent": "FunctionFly-Python-SDK/1.0",
        }

        if api_key:
            headers["Authorization"] = f"Bearer {api_key}"

        response = requests.post(
            url,
            json=input_data,
            headers=headers,
            timeout=30  # 30 second timeout
        )

        response.raise_for_status()
        result = response.json()

        if not result.get("ok", False):
            error_info = result.get("error", {})
            raise FunctionFlyError(
                error_info.get("message", "Function execution failed"),
                error_info.get("code", "EXECUTION_FAILED"),
                response.status_code
            )

        return result

    except requests.RequestException as e:
        raise FunctionFlyError(f"Network error: {str(e)}", "NETWORK_ERROR") from e
    except json.JSONDecodeError as e:
        raise FunctionFlyError(f"Invalid JSON response: {str(e)}", "PARSE_ERROR") from e


def execute_%s_%s_sync(input_data: Dict[str, Any], version: str = "latest", api_key: str = None) -> Any:
    """
    Execute function and return only the data (convenience method)

    Args:
        input_data (dict): Input parameters for the function
        version (str): Function version to execute (default: "latest")
        api_key (str): API key for authentication (optional)

    Returns:
        Any: The function execution data (result.data)
    """
    result = execute_%s_%s(input_data, version, api_key)
    return result.get("data")`,
		author, name, version, description,
		author, name, name,
		baseURL, author, name,
		author, name,
		author, name)
}

func generateGoSDK(author, name, version, title, description string) string {
	baseURL := getBaseURL()

	// Build the struct names
	inputType := strings.Title(author) + strings.Title(name) + "Input"
	outputType := strings.Title(author) + strings.Title(name) + "Output"
	executeFunc := "Execute" + strings.Title(author) + strings.Title(name)
	executeSyncFunc := "Execute" + strings.Title(author) + strings.Title(name) + "Sync"

	return fmt.Sprintf(`// FunctionFly SDK for %s/%s v%s
// %s
//
// Generated SDK for executing serverless functions on FunctionFly.
// Supports proper error handling, timeouts, and response parsing.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// FunctionFlyError represents an error from FunctionFly execution
type FunctionFlyError struct {
	Message    string
	Code       string
	StatusCode int
}

func (e *FunctionFlyError) Error() string {
	return fmt.Sprintf("FunctionFly error [%%s]: %%s", e.Code, e.Message)
}

// %s represents input parameters for the function
type %s struct {
	// Function-specific input fields should be added here
	// based on the function's manifest/schema
	Data interface{} `+"`json:\"data,omitempty\"`"+`
}

// %s represents the execution response
type %s struct {
	OK          bool            `+"`json:\"ok\"`"+`
	Data        json.RawMessage `+"`json:\"data,omitempty\"`"+`
	Cached      bool            `+"`json:\"cached\"`"+`
	DurationMs  int             `+"`json:\"duration_ms\"`"+`
	Version     string          `+"`json:\"version\"`"+`
	ExecutionID *string         `+"`json:\"execution_id,omitempty\"`"+`
	Error       *ErrorDetail    `+"`json:\"error,omitempty\"`"+`
}

// ErrorDetail represents error information
type ErrorDetail struct {
	Code    string `+"`json:\"code\"`"+`
	Message string `+"`json:\"message\"`"+`
}

// ExecutionOptions contains options for function execution
type ExecutionOptions struct {
	Version string
	APIKey  string
	Timeout time.Duration
}

// %s executes the %s function
func %s(input %s, options *ExecutionOptions) (*%s, error) {
	if options == nil {
		options = &ExecutionOptions{
			Version: "latest",
			Timeout: 30 * time.Second,
		}
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %%w", err)
	}

	url := "%s/%s/%s"
	if options.Version != "" && options.Version != "latest" {
		url += "@" + options.Version
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(inputJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %%w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "FunctionFly-Go-SDK/1.0")

	if options.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+options.APIKey)
	}

	client := &http.Client{Timeout: options.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &FunctionFlyError{
			Message: fmt.Sprintf("request failed: %%v", err),
			Code:    "NETWORK_ERROR",
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %%w", err)
	}

	var result %s
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, &FunctionFlyError{
			Message:    fmt.Sprintf("failed to parse response: %%v", err),
			Code:       "PARSE_ERROR",
			StatusCode: resp.StatusCode,
		}
	}

	if !result.OK {
		if result.Error != nil {
			return nil, &FunctionFlyError{
				Message:    result.Error.Message,
				Code:       result.Error.Code,
				StatusCode: resp.StatusCode,
			}
		}
		return nil, &FunctionFlyError{
			Message:    "function execution failed",
			Code:       "EXECUTION_FAILED",
			StatusCode: resp.StatusCode,
		}
	}

	return &result, nil
}

// %s is a convenience method that returns only the data
func %s(input %s, options *ExecutionOptions) (interface{}, error) {
	result, err := %s(input, options)
	if err != nil {
		return nil, err
	}

	var data interface{}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result data: %%w", err)
	}

	return data, nil
}`,
		author, name, version, description,
		inputType, inputType,
		outputType, outputType,
		executeFunc, name, executeFunc, inputType, outputType,
		baseURL, author, name,
		outputType,
		executeSyncFunc, executeSyncFunc, inputType, executeFunc)
}
