//! MicroVM Orchestrator Client for local runtime integration
//!
//! This module provides a client to communicate with the MicroVM orchestrator
//! service for executing functions in Firecracker microVMs.

use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use std::time::Duration;
use tokio::time::timeout;

/// Execution request for MicroVM orchestrator
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MicroVMExecutionRequest {
    /// Function code to execute
    pub code: String,
    /// Input data for the function
    pub input: String,
    /// Function handler name
    pub handler: String,
    /// Python packages to install (optional)
    #[serde(default)]
    pub packages: Vec<String>,
    /// Memory limit in MB
    pub memory_mb: u32,
    /// vCPU count
    pub vcpus: u32,
    /// Timeout in milliseconds
    pub timeout_ms: u64,
    /// Tenant ID for isolation
    pub tenant_id: String,
    /// Allowed outbound hostnames (empty = no network). Enforced by guest agent.
    #[serde(default)]
    pub network_whitelist: Vec<String>,
    /// Reject any host not on the whitelist (as opposed to a soft warning).
    #[serde(default)]
    pub strict_network_whitelist: bool,
    /// Enable per-tenant package caching in the VM.
    #[serde(default)]
    pub package_caching_enabled: bool,
}

/// Execution result from MicroVM orchestrator
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MicroVMExecutionResult {
    /// Output from the function
    pub output: String,
    /// Whether execution was successful
    pub success: bool,
    /// Error message if failed
    pub error: Option<String>,
    /// Execution time in milliseconds
    pub execution_time_ms: u64,
    /// Memory used in MB
    pub memory_used_mb: u32,
    /// Orchestrator statistics
    pub stats: Option<OrchestratorStats>,
}

/// Orchestrator statistics
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrchestratorStats {
    pub active_vms: u32,
    pub warm_vms: u32,
    pub max_vms: u32,
}

/// Client for communicating with MicroVM orchestrator
pub struct OrchestratorClient {
    /// HTTP client
    client: reqwest::Client,
    /// Orchestrator service URL
    orchestrator_url: String,
    /// Request timeout
    timeout: Duration,
    /// Optional Bearer token for the orchestrator's /execute endpoint.
    api_token: Option<String>,
}

impl OrchestratorClient {
    /// Create a new orchestrator client.
    ///
    /// `api_token` is read from `FUNCTIONFLY_MICROVM_API_TOKEN` in the environment;
    /// if the env-var is absent, requests are unauthenticated (dev mode only).
    pub fn new(orchestrator_url: String, timeout_seconds: u64) -> Result<Self> {
        let api_token = std::env::var("FUNCTIONFLY_MICROVM_API_TOKEN")
            .ok()
            .filter(|t| !t.is_empty());

        let client = reqwest::Client::builder()
            .timeout(Duration::from_secs(timeout_seconds))
            .build()
            .context("Failed to create HTTP client")?;

        Ok(Self {
            client,
            orchestrator_url: orchestrator_url.trim_end_matches('/').to_string(),
            timeout: Duration::from_secs(timeout_seconds),
            api_token,
        })
    }

    /// Execute a function in a MicroVM
    pub async fn execute_function(&self, request: MicroVMExecutionRequest) -> Result<MicroVMExecutionResult> {
        let url = format!("{}/execute", self.orchestrator_url);

        let request_json = serde_json::to_string(&request)
            .context("Failed to serialize request")?;

        let mut builder = self.client
            .post(&url)
            .header("Content-Type", "application/json")
            .body(request_json);

        if let Some(token) = &self.api_token {
            builder = builder.header("Authorization", format!("Bearer {token}"));
        }

        let response = timeout(
            self.timeout,
            builder.send()
        ).await??;

        if !response.status().is_success() {
            let status = response.status();
            let error_text = response.text().await.unwrap_or_default();
            return Err(anyhow::anyhow!(
                "Orchestrator request failed with status {}: {}",
                status, error_text
            ));
        }

        let response_text = response.text().await
            .context("Failed to read response text")?;
        let result: MicroVMExecutionResult = serde_json::from_str(&response_text)
            .context("Failed to parse response JSON")?;
        Ok(result)
    }

    /// Get orchestrator health status
    pub async fn health_check(&self) -> Result<bool> {
        let url = format!("{}/health", self.orchestrator_url);

        let response = timeout(
            self.timeout,
            self.client.get(&url).send()
        ).await?;

        match response {
            Ok(resp) => Ok(resp.status().is_success()),
            Err(_) => Ok(false),
        }
    }

    /// Get orchestrator statistics
    pub async fn get_stats(&self) -> Result<OrchestratorStats> {
        let url = format!("{}/stats", self.orchestrator_url);

        let response = timeout(
            self.timeout,
            self.client.get(&url).send()
        ).await??;

        if !response.status().is_success() {
            return Err(anyhow::anyhow!(
                "Failed to get orchestrator stats: {}",
                response.status()
            ));
        }

        let response_text = response.text().await
            .context("Failed to read response text")?;
        let stats: OrchestratorStats = serde_json::from_str(&response_text)
            .context("Failed to parse stats JSON")?;
        Ok(stats)
    }

    /// Test connection to orchestrator
    pub async fn ping(&self) -> bool {
        self.health_check().await.unwrap_or(false)
    }

    /// Get the orchestrator URL
    pub fn url(&self) -> &str {
        &self.orchestrator_url
    }
}

impl Default for OrchestratorClient {
    fn default() -> Self {
        Self::new("http://localhost:9091".to_string(), 30)
            .expect("Failed to create default OrchestratorClient")
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_client_creation() {
        let client = OrchestratorClient::new("http://localhost:8080".to_string(), 30);
        assert_eq!(client.orchestrator_url, "http://localhost:8080");
        assert_eq!(client.timeout, Duration::from_secs(30));
    }

    #[test]
    fn test_default_client() {
        let client = OrchestratorClient::default();
        assert_eq!(client.orchestrator_url, "http://localhost:9091");
        assert_eq!(client.timeout, Duration::from_secs(30));
    }
}