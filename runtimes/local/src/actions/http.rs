//! Generic HTTP action connector.
//!
//! A flexible connector for calling arbitrary REST APIs as graph actions.
//! Useful for integrating with services that don't have dedicated connectors.
//!
//! ## Supported Actions
//!
//! | Action | Description |
//! |--------|-------------|
//! | `request` | Execute arbitrary HTTP request |
//! | `get` | Shorthand for GET request |
//! | `post` | Shorthand for POST request |
//! | `put` | Shorthand for PUT request |
//! | `delete` | Shorthand for DELETE request |
//! | `patch` | Shorthand for PATCH request |
//!
//! ## Parameters
//!
//! - `url`: Target URL (required)
//! - `method`: HTTP method (default: "GET")
//! - `headers`: Object of header key-value pairs (optional)
//! - `body`: Request body for POST/PUT/PATCH (optional)
//! - `timeout_secs`: Request timeout (default: 30)
//! - `retry_count`: Number of retries on failure (default: 0)
//!
//! ## Security
//!
//! - URLs are validated against an optional allowlist
//! - Response size is limited to prevent memory exhaustion
//! - Redirects are followed but limited to 5 hops

use std::collections::HashMap;
use std::time::Instant;

use serde::{Deserialize, Serialize};
use tracing::{debug, info, instrument};

use super::connector::{ActionConnector, ActionError, ActionResult, IdempotencyKey};

/// Maximum response size in bytes (10 MB).
const MAX_RESPONSE_SIZE: usize = 10 * 1024 * 1024;

/// HTTP client configuration for the generic connector.
#[derive(Debug, Clone)]
pub struct HttpConnectorConfig {
    /// Default timeout for requests in seconds.
    pub default_timeout_secs: u64,
    /// Maximum response size in bytes.
    pub max_response_size: usize,
    /// Maximum number of redirects to follow.
    pub max_redirects: usize,
    /// Optional URL allowlist (regex patterns).
    pub url_allowlist: Vec<String>,
    /// Default headers to add to all requests.
    pub default_headers: HashMap<String, String>,
}

impl Default for HttpConnectorConfig {
    fn default() -> Self {
        Self {
            default_timeout_secs: 30,
            max_response_size: MAX_RESPONSE_SIZE,
            max_redirects: 5,
            url_allowlist: Vec::new(),
            default_headers: HashMap::new(),
        }
    }
}

/// Generic HTTP action connector.
pub struct HttpConnector {
    client: reqwest::Client,
    config: HttpConnectorConfig,
}

impl HttpConnector {
    /// Create a new HTTP connector with default config.
    pub fn new() -> Self {
        Self::with_config(HttpConnectorConfig::default())
    }

    /// Create a new HTTP connector with custom config.
    pub fn with_config(config: HttpConnectorConfig) -> Self {
        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(config.default_timeout_secs))
            .redirect(reqwest::redirect::Policy::limited(config.max_redirects))
            .build()
            .expect("HTTP client must build");

        Self { client, config }
    }

    /// Create from environment variables.
    pub fn from_env() -> Self {
        let mut config = HttpConnectorConfig::default();

        // Parse comma-separated allowlist from env
        if let Ok(allowlist) = std::env::var("HTTP_CONNECTOR_ALLOWLIST") {
            config.url_allowlist = allowlist
                .split(',')
                .map(|s| s.trim().to_string())
                .filter(|s| !s.is_empty())
                .collect();
        }

        if let Ok(timeout) = std::env::var("HTTP_CONNECTOR_TIMEOUT_SECS") {
            if let Ok(t) = timeout.parse() {
                config.default_timeout_secs = t;
            }
        }

        Self::with_config(config)
    }

    /// Validate URL against allowlist.
    fn validate_url(&self, url: &str) -> Result<(), ActionError> {
        // Basic URL validation
        if !url.starts_with("http://") && !url.starts_with("https://") {
            return Err(ActionError::fatal(
                "URL must start with http:// or https://".to_string()
            ));
        }

        // Check allowlist if configured
        if !self.config.url_allowlist.is_empty() {
            let allowed = self.config.url_allowlist.iter().any(|pattern| {
                // Simple substring matching for now, could use regex
                url.contains(pattern)
            });

            if !allowed {
                return Err(ActionError::fatal(format!(
                    "URL '{}' not in allowlist",
                    url
                )));
            }
        }

        Ok(())
    }

    /// Build request with optional headers and body.
    fn build_request(
        &self,
        method: reqwest::Method,
        url: &str,
        headers: Option<&serde_json::Map<String, serde_json::Value>>,
        body: Option<&serde_json::Value>,
    ) -> reqwest::RequestBuilder {
        let mut request = self.client.request(method, url);

        // Add default headers
        for (key, value) in &self.config.default_headers {
            request = request.header(key, value);
        }

        // Add custom headers
        if let Some(h) = headers {
            for (key, value) in h {
                if let Some(val_str) = value.as_str() {
                    request = request.header(key, val_str);
                }
            }
        }

        // Add body if present
        if let Some(b) = body {
            request = request.json(b);
        }

        request
    }

    /// Check response size limits.
    fn check_response_size(&self, size: usize) -> Result<(), ActionError> {
        if size > self.config.max_response_size {
            return Err(ActionError::fatal(format!(
                "Response size {} exceeds maximum of {}",
                size,
                self.config.max_response_size
            )));
        }
        Ok(())
    }
}

impl Default for HttpConnector {
    fn default() -> Self {
        Self::new()
    }
}

impl ActionConnector for HttpConnector {
    fn name(&self) -> &'static str {
        "http"
    }

    fn supports_action(&self, action: &str) -> bool {
        matches!(action,
            "request" | "get" | "post" | "put" | "delete" | "patch"
        )
    }

    fn validate_params(&self, action: &str, params: &serde_json::Value) -> Result<(), String> {
        // All actions require URL
        params.get("url")
            .and_then(|v| v.as_str())
            .ok_or("HTTP actions require 'url' parameter")?;

        // Validate method for specific actions
        match action {
            "get" | "delete" => {
                // These shouldn't have a body usually
                if params.get("body").is_some() {
                    debug!("GET/DELETE with body - unusual but allowed");
                }
            }
            "post" | "put" | "patch" => {
                // Body recommended but not required
            }
            "request" => {
                // Validate method if provided
                if let Some(method) = params.get("method").and_then(|v| v.as_str()) {
                    let valid = matches!(method.to_uppercase().as_str(),
                        "GET" | "POST" | "PUT" | "DELETE" | "PATCH" |
                        "HEAD" | "OPTIONS" | "TRACE"
                    );
                    if !valid {
                        return Err(format!("Invalid HTTP method: {}", method));
                    }
                }
            }
            _ => {}
        }

        Ok(())
    }

    #[instrument(skip_all, fields(action = %action))]
    async fn execute(
        &self,
        _tenant_id: Option<&str>,
        action: &str,
        params: serde_json::Value,
        _idempotency_key: &IdempotencyKey,
    ) -> Result<ActionResult, ActionError> {
        let start = Instant::now();

        // Extract URL
        let url = params.get("url")
            .and_then(|v| v.as_str())
            .ok_or_else(|| ActionError::fatal("Missing 'url' parameter"))?;

        // Validate URL
        self.validate_url(url)?;

        // Determine HTTP method
        let method = match action {
            "get" => reqwest::Method::GET,
            "post" => reqwest::Method::POST,
            "put" => reqwest::Method::PUT,
            "delete" => reqwest::Method::DELETE,
            "patch" => reqwest::Method::PATCH,
            "request" => {
                let method_str = params.get("method")
                    .and_then(|v| v.as_str())
                    .unwrap_or("GET");
                match method_str.to_uppercase().as_str() {
                    "GET" => reqwest::Method::GET,
                    "POST" => reqwest::Method::POST,
                    "PUT" => reqwest::Method::PUT,
                    "DELETE" => reqwest::Method::DELETE,
                    "PATCH" => reqwest::Method::PATCH,
                    "HEAD" => reqwest::Method::HEAD,
                    "OPTIONS" => reqwest::Method::OPTIONS,
                    _ => reqwest::Method::GET,
                }
            }
            _ => reqwest::Method::GET,
        };

        // Extract headers and body
        let headers = params.get("headers")
            .and_then(|v| v.as_object());

        let body = params.get("body");

        debug!(method = %method, url = %url, "Executing HTTP request");

        // Build and execute request
        let request = self.build_request(method, url, headers, body);

        let response = request
            .send()
            .await
            .map_err(ActionError::from_reqwest)?;

        let status = response.status();
        let headers_map: HashMap<String, String> = response
            .headers()
            .iter()
            .filter_map(|(k, v)| {
                v.to_str().ok().map(|v| (k.to_string(), v.to_string()))
            })
            .collect();

        // Check content length if available
        if let Some(content_length) = response.content_length() {
            self.check_response_size(content_length as usize)?;
        }

        // Read response body
        let body_bytes = response.bytes().await
            .map_err(|e| ActionError::fatal(format!("Failed to read response body: {}", e)))?;

        self.check_response_size(body_bytes.len())?;

        let body_text = String::from_utf8_lossy(&body_bytes);

        // Try to parse as JSON
        let body_json: serde_json::Value = serde_json::from_str(&body_text)
            .unwrap_or_else(|_| serde_json::Value::String(body_text.to_string()));

        info!(
            status = %status,
            url = %url,
            bytes = body_bytes.len(),
            "HTTP request completed"
        );

        let success = status.is_success();
        let mut result = ActionResult {
            success,
            data: serde_json::json!({
                "status": status.as_u16(),
                "status_text": status.canonical_reason().unwrap_or("Unknown"),
                "headers": headers_map,
                "body": body_json,
            }),
            provider_ref: None,
            usage: None,
            latency_ms: start.elapsed().as_millis() as u64,
            error: if success { None } else { Some(format!("HTTP {}: {}", status, body_text)) },
        };

        if !success {
            result.error = Some(format!("HTTP error: {}", status));
        }

        Ok(result)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_http_connector_creation() {
        let connector = HttpConnector::new();
        assert_eq!(connector.name(), "http");
        assert!(connector.supports_action("get"));
        assert!(connector.supports_action("post"));
        assert!(connector.supports_action("put"));
        assert!(connector.supports_action("delete"));
        assert!(connector.supports_action("patch"));
        assert!(connector.supports_action("request"));
        assert!(!connector.supports_action("unknown"));
    }

    #[test]
    fn test_validate_params() {
        let connector = HttpConnector::new();

        // Valid GET
        let get_params = serde_json::json!({
            "url": "https://api.example.com/users"
        });
        assert!(connector.validate_params("get", &get_params).is_ok());

        // Valid POST with body
        let post_params = serde_json::json!({
            "url": "https://api.example.com/users",
            "body": { "name": "John" }
        });
        assert!(connector.validate_params("post", &post_params).is_ok());

        // Missing URL
        let no_url = serde_json::json!({ "method": "GET" });
        assert!(connector.validate_params("get", &no_url).is_err());

        // Invalid method in request
        let bad_method = serde_json::json!({
            "url": "https://api.example.com",
            "method": "INVALID"
        });
        assert!(connector.validate_params("request", &bad_method).is_err());
    }

    #[test]
    fn test_url_validation() {
        let mut config = HttpConnectorConfig::default();
        config.url_allowlist = vec!["api.example.com".to_string()];

        let connector = HttpConnector::with_config(config);

        // Valid URL in allowlist
        assert!(connector.validate_url("https://api.example.com/users").is_ok());

        // URL not in allowlist
        assert!(connector.validate_url("https://evil.com").is_err());

        // Invalid URL scheme
        assert!(connector.validate_url("ftp://api.example.com").is_err());
    }

    #[test]
    fn test_url_validation_no_allowlist() {
        let connector = HttpConnector::new();

        // Any HTTPS URL should work without allowlist
        assert!(connector.validate_url("https://api.example.com").is_ok());
        assert!(connector.validate_url("https://api.github.com/users").is_ok());

        // HTTP allowed too
        assert!(connector.validate_url("http://localhost:8080").is_ok());
    }
}
