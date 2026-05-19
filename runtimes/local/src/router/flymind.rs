//! FlyMind HTTP client for LLM inference.
//!
//! Calls the Python FlyMind service (port 8081) for provider selection and inference.
//! FlyMind owns the 9-provider routing logic; this client just wraps the HTTP API.

use std::collections::HashMap;
use std::time::Duration;

use anyhow::{anyhow, Result};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use tracing::{debug, instrument, warn};

// Re-export LlmTrafficType from engine::graph so that router/mod.rs can re-export it
pub use crate::engine::graph::LlmTrafficType;

/// Configuration for the FlyMind client.
#[derive(Debug, Clone)]
pub struct FlyMindConfig {
    /// Base URL of the FlyMind service (e.g. "http://localhost:8081")
    pub base_url: String,
    /// Request timeout for LLM calls
    pub timeout: Duration,
    /// Optional API key for FlyMind auth
    pub api_key: Option<String>,
}

impl Default for FlyMindConfig {
    fn default() -> Self {
        Self {
            base_url: std::env::var("FLYMIND_URL").unwrap_or_default(),
            timeout: Duration::from_secs(120),
            api_key: std::env::var("FLYMIND_API_KEY").ok(),
        }
    }
}

/// Result of routing a request to FlyMind.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RouteResult {
    /// The content of the LLM response.
    pub content: String,
    /// Provider that fulfilled the request.
    pub provider: String,
    /// Model that was used.
    pub model: String,
    /// Token usage breakdown.
    pub usage: Usage,
    /// Latency in milliseconds.
    pub latency_ms: f64,
    /// Finish reason (stop, length, etc.)
    #[serde(default)]
    pub finish_reason: Option<String>,
}

/// Token usage from a completion.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Usage {
    #[serde(rename = "prompt_tokens")]
    pub prompt_tokens: u32,
    #[serde(rename = "completion_tokens")]
    pub completion_tokens: u32,
    #[serde(rename = "total_tokens")]
    pub total_tokens: u32,
}

impl Default for Usage {
    fn default() -> Self {
        Self {
            prompt_tokens: 0,
            completion_tokens: 0,
            total_tokens: 0,
        }
    }
}

/// A chat message in a completion request.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatMessage {
    pub role: String,
    pub content: String,
}

/// Request body for the `/api/complete` endpoint.
#[derive(Debug, Serialize)]
struct CompletionRequest<'a> {
    provider: Option<&'a str>,
    model: Option<&'a str>,
    messages: Vec<ChatMessage>,
    temperature: f32,
    #[serde(rename = "max_tokens")]
    max_tokens: Option<u32>,
    stream: bool,
    #[serde(rename = "top_p")]
    top_p: Option<f32>,
    stop: Option<Vec<&'a str>>,
}

/// Response body from `/api/complete`.
#[derive(Debug, Deserialize)]
struct CompletionResponse {
    content: String,
    provider: String,
    model: String,
    usage: Usage,
    #[serde(rename = "finish_reason")]
    finish_reason: Option<String>,
    #[serde(rename = "latency_ms")]
    latency_ms: f64,
}

/// Maps our `LlmTrafficType` to a FlyMind provider hint string.
/// FlyMind uses these provider hints to override auto-classification.
fn traffic_type_to_provider_hint(traffic_type: LlmTrafficType) -> Option<&'static str> {
    match traffic_type {
        LlmTrafficType::Realtime => Some("groq"),
        LlmTrafficType::Structured => Some("fireworks"),
        LlmTrafficType::FunctionCalling => Some("fireworks"),
        LlmTrafficType::Background => Some("deepinfra"),
        LlmTrafficType::General => None,
    }
}

/// FlyMind HTTP client.
#[derive(Debug, Clone)]
pub struct FlyMindClient {
    client: Client,
    config: FlyMindConfig,
}

impl FlyMindClient {
    /// Create a new FlyMind client from config.
    pub fn new(config: FlyMindConfig) -> Self {
        let client = Client::builder()
            .timeout(config.timeout)
            .build()
            .expect("FlyMind HTTP client must build");
        Self { client, config }
    }

    /// Create a client with default configuration.
    pub fn default_client() -> Self {
        Self::new(FlyMindConfig::default())
    }

    /// Build the completion URL.
    fn complete_url(&self) -> String {
        format!("{}/api/complete", self.config.base_url)
    }

    /// Execute an LLM completion via FlyMind.
    ///
    /// `messages` is a map of role → content (e.g. "system" → "You are helpful").
    /// `traffic_type` is the hint passed to FlyMind's router.
    #[instrument(skip_all, fields(traffic_type = ?traffic_type))]
    pub async fn complete(
        &self,
        messages: &HashMap<String, String>,
        traffic_type: LlmTrafficType,
        model: Option<&str>,
        temperature: f32,
        max_tokens: Option<u32>,
    ) -> Result<RouteResult> {
        let provider_hint = traffic_type_to_provider_hint(traffic_type);

        let chat_messages: Vec<ChatMessage> = messages
            .iter()
            .map(|(role, content)| ChatMessage {
                role: role.clone(),
                content: content.clone(),
            })
            .collect();

        let req_body = CompletionRequest {
            provider: provider_hint,
            model: model.unwrap_or("default").into(),
            messages: chat_messages,
            temperature,
            max_tokens,
            stream: false,
            top_p: None,
            stop: None,
        };

        debug!(
            url = %self.complete_url(),
            provider_hint = ?provider_hint,
            message_count = req_body.messages.len(),
            "Calling FlyMind /api/complete"
        );

        let mut req = self.client.post(self.complete_url()).json(&req_body);

        if let Some(ref key) = self.config.api_key {
            req = req.header("Authorization", format!("Bearer {}", key));
        }

        let response = req.send().await?;

        if !response.status().is_success() {
            let status = response.status();
            let body = response.text().await.unwrap_or_default();
            warn!(
                status = %status,
                body = %body,
                "FlyMind returned error"
            );
            return Err(anyhow!("FlyMind returned {}: {}", status, body));
        }

        let completion: CompletionResponse = response.json().await?;

        debug!(
            provider = %completion.provider,
            model = %completion.model,
            latency_ms = completion.latency_ms,
            tokens = ?completion.usage,
            "FlyMind completion successful"
        );

        Ok(RouteResult {
            content: completion.content,
            provider: completion.provider,
            model: completion.model,
            usage: completion.usage,
            latency_ms: completion.latency_ms,
            finish_reason: completion.finish_reason,
        })
    }

    /// Check if FlyMind service is healthy.
    pub async fn health_check(&self) -> bool {
        let url = format!("{}/health", self.config.base_url);
        match self.client.get(&url).send().await {
            Ok(resp) => resp.status().is_success(),
            Err(_) => false,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_traffic_type_to_provider_hint() {
        assert_eq!(traffic_type_to_provider_hint(LlmTrafficType::Realtime), Some("groq"));
        assert_eq!(traffic_type_to_provider_hint(LlmTrafficType::Structured), Some("fireworks"));
        assert_eq!(traffic_type_to_provider_hint(LlmTrafficType::FunctionCalling), Some("fireworks"));
        assert_eq!(traffic_type_to_provider_hint(LlmTrafficType::Background), Some("deepinfra"));
        assert_eq!(traffic_type_to_provider_hint(LlmTrafficType::General), None);
    }

    #[test]
    fn test_flymind_client_default_url() {
        let client = FlyMindClient::default_client();
        assert_eq!(client.config.base_url, "http://localhost:8081");
    }
}
