use std::collections::HashMap;
use std::sync::Arc;

use async_trait::async_trait;
use parking_lot::RwLock;
use serde::{Deserialize, Serialize};
use tracing::info;
use uuid::Uuid;

use crate::engine::{NodeId, NodeExecutionError};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ActionConfig {
    pub connector: String,
    pub action: String,
    pub params: serde_json::Value,
    pub timeout_ms: u64,
    pub retries: u32,
}

impl ActionConfig {
    pub fn new(connector: String, action: String, params: serde_json::Value) -> Self {
        Self {
            connector,
            action,
            params,
            timeout_ms: 30_000,
            retries: 3,
        }
    }
}

#[async_trait]
pub trait ActionConnector: Send + Sync {
    fn name(&self) -> &str;
    async fn execute(&self, action: &str, params: serde_json::Value) -> Result<serde_json::Value, NodeExecutionError>;
}

pub struct ActionRegistry {
    connectors: Arc<RwLock<HashMap<String, Arc<dyn ActionConnector>>>>,
}

impl ActionRegistry {
    pub fn new() -> Self {
        Self {
            connectors: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub fn register<C: ActionConnector + 'static>(&self, connector: C) {
        let name = connector.name().to_string();
        let mut connectors = self.connectors.write();
        connectors.insert(name.clone(), Arc::new(connector));
        info!(connector = %name, "Action connector registered");
    }

    pub async fn execute(
        &self,
        connector_name: &str,
        action: &str,
        params: serde_json::Value,
    ) -> Result<serde_json::Value, NodeExecutionError> {
        let connectors = self.connectors.read();
        let connector = connectors.get(connector_name)
            .ok_or_else(|| NodeExecutionError::non_retryable(
                NodeId(Uuid::nil()),
                format!("Connector {} not found", connector_name),
            ))?;

        connector.execute(action, params).await
    }

    pub fn list_connectors(&self) -> Vec<String> {
        self.connectors.read().keys().cloned().collect()
    }
}

impl Default for ActionRegistry {
    fn default() -> Self {
        Self::new()
    }
}

pub struct StripeConnector {
    api_key: Option<String>,
}

impl StripeConnector {
    pub fn new(api_key: Option<String>) -> Self {
        Self { api_key }
    }
}

#[async_trait]
impl ActionConnector for StripeConnector {
    fn name(&self) -> &str {
        "stripe"
    }

    async fn execute(&self, action: &str, params: serde_json::Value) -> Result<serde_json::Value, NodeExecutionError> {
        match action {
            "charge" => {
                Ok(serde_json::json!({
                    "charge_id": "ch_stub",
                    "status": "succeeded",
                    "amount": params.get("amount").and_then(|v| v.as_i64()).unwrap_or(0),
                }))
            }
            "create_customer" => {
                Ok(serde_json::json!({
                    "customer_id": "cus_stub",
                    "email": params.get("email").and_then(|v| v.as_str()).unwrap_or(""),
                }))
            }
            _ => Err(NodeExecutionError::non_retryable(
                NodeId(Uuid::nil()),
                format!("Unknown stripe action: {}", action),
            )),
        }
    }
}

pub struct SendGridConnector {
    api_key: Option<String>,
}

impl SendGridConnector {
    pub fn new(api_key: Option<String>) -> Self {
        Self { api_key }
    }
}

#[async_trait]
impl ActionConnector for SendGridConnector {
    fn name(&self) -> &str {
        "sendgrid"
    }

    async fn execute(&self, action: &str, params: serde_json::Value) -> Result<serde_json::Value, NodeExecutionError> {
        match action {
            "send_email" => {
                Ok(serde_json::json!({
                    "message_id": "msg_stub",
                    "status": "sent",
                    "to": params.get("to").and_then(|v| v.as_str()).unwrap_or(""),
                }))
            }
            _ => Err(NodeExecutionError::non_retryable(
                NodeId(Uuid::nil()),
                format!("Unknown sendgrid action: {}", action),
            )),
        }
    }
}

pub struct HttpConnector {
    client: reqwest::Client,
}

impl HttpConnector {
    pub fn new() -> Self {
        Self {
            client: reqwest::Client::new(),
        }
    }
}

#[async_trait]
impl ActionConnector for HttpConnector {
    fn name(&self) -> &str {
        "http"
    }

    async fn execute(&self, action: &str, params: serde_json::Value) -> Result<serde_json::Value, NodeExecutionError> {
        match action {
            "get" => {
                let url = params.get("url").and_then(|v| v.as_str()).unwrap_or("");
                let response = self.client.get(url).send().await
                    .map_err(|e| NodeExecutionError::new(NodeId(Uuid::nil()), format!("HTTP GET failed: {}", e)))?;
                let body = response.text().await
                    .map_err(|e| NodeExecutionError::new(NodeId(Uuid::nil()), format!("Body read failed: {}", e)))?;
                Ok(serde_json::json!({ "body": body }))
            }
            "post" => {
                let url = params.get("url").and_then(|v| v.as_str()).unwrap_or("");
                let body = params.get("body").cloned().unwrap_or(serde_json::Value::Null);
                let response = self.client.post(url).json(&body).send().await
                    .map_err(|e| NodeExecutionError::new(NodeId(Uuid::nil()), format!("HTTP POST failed: {}", e)))?;
                let response_body = response.text().await
                    .map_err(|e| NodeExecutionError::new(NodeId(Uuid::nil()), format!("Body read failed: {}", e)))?;
                Ok(serde_json::json!({ "body": response_body }))
            }
            _ => Err(NodeExecutionError::non_retryable(
                NodeId(Uuid::nil()),
                format!("Unknown http action: {}", action),
            )),
        }
    }
}

impl Default for HttpConnector {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_action_registry() {
        let registry = ActionRegistry::new();
        registry.register(StripeConnector::new(None));
        registry.register(SendGridConnector::new(None));

        let connectors = registry.list_connectors();
        assert!(connectors.contains(&"stripe".to_string()));
        assert!(connectors.contains(&"sendgrid".to_string()));
    }

    #[tokio::test]
    async fn test_stripe_charge() {
        let registry = ActionRegistry::new();
        registry.register(StripeConnector::new(None));

        let result = registry.execute("stripe", "charge", serde_json::json!({
            "amount": 1000,
            "currency": "usd",
        })).await.unwrap();

        assert_eq!(result["charge_id"], "ch_stub");
        assert_eq!(result["amount"], 1000);
    }

    #[tokio::test]
    async fn test_sendgrid_send_email() {
        let registry = ActionRegistry::new();
        registry.register(SendGridConnector::new(None));

        let result = registry.execute("sendgrid", "send_email", serde_json::json!({
            "to": "test@example.com",
            "subject": "Hello",
        })).await.unwrap();

        assert_eq!(result["status"], "sent");
        assert_eq!(result["to"], "test@example.com");
    }

    #[tokio::test]
    async fn test_unknown_connector() {
        let registry = ActionRegistry::new();
        let result = registry.execute("nonexistent", "action", serde_json::json!({})).await;
        assert!(result.is_err());
    }
}
