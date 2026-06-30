//! Host Functions
//! 
//! This module provides host functions that can be called from within
//! the JavaScript execution environment (fetch, console, etc.).

use std::collections::HashMap;
use std::sync::Arc;

use parking_lot::RwLock;
use serde::{Deserialize, Serialize};
use tracing::{debug, info, warn};

/// Host functions context - provides APIs available to executing functions
pub struct HostFunctions {
    /// Console function implementations
    console: ConsoleHandler,
    
    /// Custom environment variables
    env: HashMap<String, String>,
    
    /// Timing functions
    timing: TimingHandler,
    
    /// Fetch implementation (if network enabled)
    fetch: Option<FetchHandler>,

    /// Attestation retrieval handler
    attestation: AttestationHandler,

    /// Function delegation handler
    delegate_handler: DelegateHandler,
}

impl HostFunctions {
    /// Create new host functions
    pub fn new(network_enabled: bool, env: HashMap<String, String>) -> Self {
        Self {
            console: ConsoleHandler::new(),
            env,
            timing: TimingHandler::new(),
            fetch: if network_enabled {
                Some(FetchHandler::new())
            } else {
                None
            },
            attestation: AttestationHandler::new(),
            delegate_handler: DelegateHandler::new(),
        }
    }

    /// Get console handler
    pub fn console(&self) -> &ConsoleHandler {
        &self.console
    }

    /// Get environment variable
    pub fn get_env(&self, key: &str) -> Option<String> {
        self.env.get(key).cloned()
    }

    /// Get timing handler
    pub fn timing(&self) -> &TimingHandler {
        &self.timing
    }

    /// Get fetch handler
    pub fn fetch(&self) -> Option<&FetchHandler> {
        self.fetch.as_ref()
    }

    /// Get attestation handler
    pub fn attestation(&self) -> &AttestationHandler {
        &self.attestation
    }

    /// Get delegate handler
    pub fn delegate_handler(&self) -> &DelegateHandler {
        &self.delegate_handler
    }
}

/// Console handler - provides console.log, console.error, console.warn
pub struct ConsoleHandler {
    logs: RwLock<Vec<ConsoleLog>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConsoleLog {
    pub level: String,
    pub message: String,
    pub timestamp: String,
}

impl ConsoleHandler {
    pub fn new() -> Self {
        Self {
            logs: RwLock::new(Vec::new()),
        }
    }

    /// Log a message
    pub fn log(&self, level: &str, message: String) {
        let log = ConsoleLog {
            level: level.to_string(),
            message,
            timestamp: chrono::Utc::now().to_rfc3339(),
        };
        
        self.logs.write().push(log.clone());
        
        // Also output to stderr/stdout
        match level {
            "error" => eprintln!("[{}] {}", level, log.message),
            "warn" => eprintln!("[{}] {}", level, log.message),
            _ => println!("[{}] {}", level, log.message),
        }
    }

    /// Get all logs
    pub fn get_logs(&self) -> Vec<ConsoleLog> {
        self.logs.read().clone()
    }

    /// Clear logs
    pub fn clear(&self) {
        self.logs.write().clear();
    }
}

impl Default for ConsoleHandler {
    fn default() -> Self {
        Self::new()
    }
}

/// Timing functions handler - provides setTimeout, setInterval, etc.
pub struct TimingHandler {
    timers: RwLock<HashMap<String, TimerEntry>>,
    next_id: RwLock<u64>,
}

struct TimerEntry {
    callback: Box<dyn Fn() + Send + 'static>,
    interval_ms: u64,
    repeating: bool,
}

impl TimingHandler {
    pub fn new() -> Self {
        Self {
            timers: RwLock::new(HashMap::new()),
            next_id: RwLock::new(1),
        }
    }

    /// Set a timer
    pub fn set_timeout<F>(&self, callback: F, _delay_ms: u64) -> String
    where
        F: Fn() + Send + 'static,
    {
        let id = {
            let mut next = self.next_id.write();
            let id = format!("timeout_{}", *next);
            *next += 1;
            id
        };
        
        let mut timers = self.timers.write();
        timers.insert(id.clone(), TimerEntry {
            callback: Box::new(callback),
            interval_ms: 0,
            repeating: false,
        });
        
        debug!("Set timeout: {}", id);
        id
    }

    /// Set an interval
    pub fn set_interval<F>(&self, callback: F, _interval_ms: u64) -> String
    where
        F: Fn() + Send + 'static,
    {
        let id = {
            let mut next = self.next_id.write();
            let id = format!("interval_{}", *next);
            *next += 1;
            id
        };
        
        let mut timers = self.timers.write();
        timers.insert(id.clone(), TimerEntry {
            callback: Box::new(callback),
            interval_ms: _interval_ms,
            repeating: true,
        });
        
        debug!("Set interval: {}", id);
        id
    }

    /// Clear a timer
    pub fn clear_timeout(&self, id: &str) {
        self.timers.write().remove(id);
        debug!("Cleared timer: {}", id);
    }
}

impl Default for TimingHandler {
    fn default() -> Self {
        Self::new()
    }
}

/// Fetch handler - provides fetch API implementation
pub struct FetchHandler {
    client: reqwest::Client,
}

impl FetchHandler {
    pub fn new() -> Self {
        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(30))
            .build()
            .expect("Failed to create HTTP client");
        
        Self { client }
    }

    /// Perform a fetch request
    pub async fn fetch(&self, url: &str, options: FetchOptions) -> Result<FetchResponse, String> {
        info!("Fetch request to: {}", url);
        
        let mut request = match options.method.to_uppercase().as_str() {
            "GET" => self.client.get(url),
            "POST" => self.client.post(url),
            "PUT" => self.client.put(url),
            "DELETE" => self.client.delete(url),
            "PATCH" => self.client.patch(url),
            "HEAD" => self.client.head(url),
            _ => return Err(format!("Unsupported method: {}", options.method)),
        };
        
        // Add headers
        for (key, value) in &options.headers {
            request = request.header(key, value);
        }
        
        // Add body
        if let Some(body) = options.body {
            request = request.body(body);
        }
        
        // Execute request
        let response = request.send().await
            .map_err(|e| format!("Fetch error: {}", e))?;
        
        let status = response.status().as_u16();
        let status_text = response.status().canonical_reason().unwrap_or("").to_string();
        let headers: HashMap<String, String> = response
            .headers()
            .iter()
            .map(|(k, v)| (k.to_string(), v.to_str().unwrap_or("").to_string()))
            .collect();
        
        let body = response.text().await
            .map_err(|e| format!("Failed to read response body: {}", e))?;
        
        Ok(FetchResponse {
            status,
            status_text,
            headers,
            body,
        })
    }
}

impl Default for FetchHandler {
    fn default() -> Self {
        Self::new()
    }
}

/// Fetch options
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FetchOptions {
    pub method: String,
    pub headers: HashMap<String, String>,
    pub body: Option<String>,
}

impl Default for FetchOptions {
    fn default() -> Self {
        Self {
            method: "GET".to_string(),
            headers: HashMap::new(),
            body: None,
        }
    }
}

/// Fetch response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FetchResponse {
    pub status: u16,
    pub status_text: String,
    pub headers: HashMap<String, String>,
    pub body: String,
}

/// Attestation retrieval handler - provides getAttestation() to JS functions
pub struct AttestationHandler {
    /// Base URL for the Trust API
    base_url: String,
    /// API key for authentication
    api_key: String,
}

impl AttestationHandler {
    pub fn new() -> Self {
        Self {
            base_url: std::env::var("FUNCTIONFLY_API_URL")
                .unwrap_or_else(|_| "https://api.functionfly.com".to_string()),
            api_key: std::env::var("FUNCTIONFLY_API_KEY")
                .unwrap_or_default(),
        }
    }

    /// Retrieve an attestation by ID via the Trust API
    pub async fn get_attestation(&self, attestation_id: &str) -> Result<String, String> {
        let url = format!("{}/v1/trust/attestations/{}", self.base_url, attestation_id);
        let client = reqwest::Client::new();
        let response = client
            .get(&url)
            .header("Authorization", format!("Bearer {}", self.api_key))
            .send()
            .await
            .map_err(|e| format!("getAttestation request failed: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("getAttestation failed: HTTP {}", response.status()));
        }

        response.text().await
            .map_err(|e| format!("getAttestation response read failed: {}", e))
    }
}

/// Function delegation handler - provides delegate() to JS functions
pub struct DelegateHandler {
    base_url: String,
    api_key: String,
}

impl DelegateHandler {
    pub fn new() -> Self {
        Self {
            base_url: std::env::var("FUNCTIONFLY_API_URL")
                .unwrap_or_else(|_| "https://api.functionfly.com".to_string()),
            api_key: std::env::var("FUNCTIONFLY_API_KEY")
                .unwrap_or_default(),
        }
    }

    /// Delegate execution to another function via the execution API
    pub async fn delegate(
        &self,
        target_function_id: &str,
        input: &str,
        options: &str,
    ) -> Result<String, String> {
        let url = format!("{}/v1/functions/{}/execute", self.base_url, target_function_id);
        let client = reqwest::Client::new();

        let mut body: serde_json::Value = serde_json::from_str(input)
            .unwrap_or_else(|_| serde_json::json!({ "input": input }));

        if !options.is_empty() {
            if let Ok(opts) = serde_json::from_str::<serde_json::Value>(options) {
                body["options"] = opts;
            }
        }

        let response = client
            .post(&url)
            .header("Authorization", format!("Bearer {}", self.api_key))
            .header("Content-Type", "application/json")
            .body(body.to_string())
            .send()
            .await
            .map_err(|e| format!("delegate request failed: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("delegate failed: HTTP {}", response.status()));
        }

        response.text().await
            .map_err(|e| format!("delegate response read failed: {}", e))
    }
}
