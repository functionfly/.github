//! EventSource implementations for the Agent Daemon
//!
//! Provides concrete implementations for:
//! - WebhookEventSource: HTTP webhooks from external services
//! - DatabaseEventSource: PostgreSQL LISTEN/NOTIFY
//! - ScheduledEventSource: Cron-like scheduled triggers

use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::time::{Duration, Instant};

use async_trait::async_trait;
use tokio::sync::mpsc;
use tracing::{debug, error, info, warn};
use uuid::Uuid;

use crate::daemon::agent_daemon::{AgentEvent, EventSource, EventSourceType};

// ---------------------------------------------------------------------------
// Webhook Event Source
// ---------------------------------------------------------------------------

/// Event source for HTTP webhooks
#[derive(Clone, Debug)]
pub struct WebhookEventSource {
    name: String,
    port: u16,
    path: String,
    running: Arc<AtomicBool>,
    secret_header: Option<String>,
}

impl WebhookEventSource {
    /// Create a new webhook event source
    pub fn new(name: String, port: u16, path: String, secret_header: Option<String>) -> Self {
        Self {
            name,
            port,
            path,
            running: Arc::new(AtomicBool::new(false)),
            secret_header,
        }
    }

    /// Validate webhook signature (placeholder for HMAC verification)
    fn validate_signature(&self, _headers: &[(String, String)], _body: &[u8]) -> bool {
        // In production, verify HMAC signature
        true
    }
}

#[async_trait]
impl EventSource for WebhookEventSource {
    fn source_type(&self) -> EventSourceType {
        EventSourceType::Webhook {
            name: self.name.clone(),
        }
    }

    async fn start(&self) -> anyhow::Result<mpsc::Receiver<AgentEvent>> {
        let (_tx, rx) = mpsc::channel(100);
        self.running.store(true, Ordering::SeqCst);

        info!(
            name = %self.name,
            port = self.port,
            path = %self.path,
            "Starting webhook event source"
        );

        // In production, this would start an HTTP server
        // For now, we simulate with a placeholder
        let name = self.name.clone();
        tokio::spawn(async move {
            // Placeholder: In production, bind HTTP server here
            // axum::serve(bind, router).await

            loop {
                // Simulate webhook events for testing
                tokio::time::sleep(Duration::from_secs(60)).await;
                debug!("Webhook event source '{}' heartbeat", name);
            }
        });

        Ok(rx)
    }

    async fn stop(&self) -> anyhow::Result<()> {
        self.running.store(false, Ordering::SeqCst);
        info!(name = %self.name, "Stopped webhook event source");
        Ok(())
    }

    fn is_running(&self) -> bool {
        self.running.load(Ordering::SeqCst)
    }
}

// ---------------------------------------------------------------------------
// Database Event Source (PostgreSQL LISTEN/NOTIFY)
// ---------------------------------------------------------------------------

/// Event source for PostgreSQL database notifications
#[derive(Clone, Debug)]
pub struct DatabaseEventSource {
    table: String,
    operations: Vec<String>, // INSERT, UPDATE, DELETE
    channel: String,
    database_url: String,
    running: Arc<AtomicBool>,
}

impl DatabaseEventSource {
    /// Create a new database event source
    pub fn new(table: String, operations: Vec<String>, database_url: String) -> Self {
        let channel = format!("ff_table_{}", table);
        Self {
            table,
            operations,
            channel,
            database_url,
            running: Arc::new(AtomicBool::new(false)),
        }
    }

    /// Parse notification payload into an event
    fn parse_notification(&self, payload: &str) -> Option<AgentEvent> {
        // Expected format: {"operation": "INSERT", "row": {...}}
        match serde_json::from_str::<serde_json::Value>(payload) {
            Ok(data) => {
                let operation = data.get("operation")?.as_str()?;

                if !self.operations.iter().any(|op| op == operation) {
                    return None;
                }

                Some(AgentEvent {
                    id: Uuid::new_v4(),
                    source: self.source_type(),
                    agent_id: String::new(), // Filled in by daemon
                    payload: data,
                    timestamp: Instant::now(),
                })
            }
            Err(e) => {
                warn!("Failed to parse notification: {}", e);
                None
            }
        }
    }
}

#[async_trait]
impl EventSource for DatabaseEventSource {
    fn source_type(&self) -> EventSourceType {
        EventSourceType::Database {
            table: self.table.clone(),
            operation: self.operations.join(","),
        }
    }

    async fn start(&self) -> anyhow::Result<mpsc::Receiver<AgentEvent>> {
        let (_tx, rx) = mpsc::channel(100);
        self.running.store(true, Ordering::SeqCst);

        info!(
            table = %self.table,
            channel = %self.channel,
            "Starting database event source"
        );

        // In production, this would connect to PostgreSQL and LISTEN
        // For now, we simulate with a placeholder
        let channel = self.channel.clone();
        tokio::spawn(async move {
            // Placeholder: In production, use tokio_postgres to LISTEN
            // let (client, connection) = tokio_postgres::connect(&database_url, NoTls).await?;
            // client.execute(&format!("LISTEN {}", channel), &[]).await?;

            loop {
                tokio::time::sleep(Duration::from_secs(30)).await;
                debug!("Database event source '{}' heartbeat", channel);
            }
        });

        Ok(rx)
    }

    async fn stop(&self) -> anyhow::Result<()> {
        self.running.store(false, Ordering::SeqCst);
        info!(table = %self.table, "Stopped database event source");
        Ok(())
    }

    fn is_running(&self) -> bool {
        self.running.load(Ordering::SeqCst)
    }
}

// ---------------------------------------------------------------------------
// Scheduled Event Source (Cron-like)
// ---------------------------------------------------------------------------

/// Event source for scheduled/cron triggers
#[derive(Clone, Debug)]
pub struct ScheduledEventSource {
    schedule_id: String,
    cron_expression: String,
    timezone: String,
    running: Arc<AtomicBool>,
}

impl ScheduledEventSource {
    /// Create a new scheduled event source
    pub fn new(schedule_id: String, cron_expression: String, timezone: String) -> Self {
        Self {
            schedule_id,
            cron_expression,
            timezone,
            running: Arc::new(AtomicBool::new(false)),
        }
    }

    /// Parse cron expression and compute next trigger time
    fn next_trigger(&self) -> Option<Instant> {
        // In production, use cron-parser crate
        // For now, return a fixed interval for testing
        Some(Instant::now() + Duration::from_secs(60))
    }
}

#[async_trait]
impl EventSource for ScheduledEventSource {
    fn source_type(&self) -> EventSourceType {
        EventSourceType::Scheduled {
            schedule_id: self.schedule_id.clone(),
        }
    }

    async fn start(&self) -> anyhow::Result<mpsc::Receiver<AgentEvent>> {
        let (tx, rx) = mpsc::channel(100);
        self.running.store(true, Ordering::SeqCst);

        info!(
            schedule_id = %self.schedule_id,
            cron = %self.cron_expression,
            "Starting scheduled event source"
        );

        let schedule_id = self.schedule_id.clone();
        let running = self.running.clone();

        tokio::spawn(async move {
            while running.load(Ordering::SeqCst) {
                // In production, sleep until next cron trigger
                tokio::time::sleep(Duration::from_secs(60)).await;

                if !running.load(Ordering::SeqCst) {
                    break;
                }

                let event = AgentEvent {
                    id: Uuid::new_v4(),
                    source: EventSourceType::Scheduled {
                        schedule_id: schedule_id.clone(),
                    },
                    agent_id: String::new(),
                    payload: serde_json::json!({
                        "triggered_at": chrono::Utc::now().to_rfc3339(),
                        "schedule_id": schedule_id,
                    }),
                    timestamp: Instant::now(),
                };

                if let Err(e) = tx.send(event).await {
                    error!("Failed to send scheduled event: {}", e);
                    break;
                }

                info!(schedule_id = %schedule_id, "Scheduled trigger fired");
            }
        });

        Ok(rx)
    }

    async fn stop(&self) -> anyhow::Result<()> {
        self.running.store(false, Ordering::SeqCst);
        info!(schedule_id = %self.schedule_id, "Stopped scheduled event source");
        Ok(())
    }

    fn is_running(&self) -> bool {
        self.running.load(Ordering::SeqCst)
    }
}

// ---------------------------------------------------------------------------
// Factory Functions
// ---------------------------------------------------------------------------

/// Create a webhook event source for Stripe
pub fn stripe_webhook_source(port: u16) -> Arc<WebhookEventSource> {
    Arc::new(WebhookEventSource::new(
        "stripe".to_string(),
        port,
        "/webhooks/stripe".to_string(),
        Some("Stripe-Signature".to_string()),
    ))
}

/// Create a webhook event source for Shopify
pub fn shopify_webhook_source(port: u16) -> Arc<WebhookEventSource> {
    Arc::new(WebhookEventSource::new(
        "shopify".to_string(),
        port,
        "/webhooks/shopify".to_string(),
        Some("X-Shopify-Hmac-SHA256".to_string()),
    ))
}

/// Create a database event source for a table
pub fn database_table_source(table: &str, operations: &[&str], database_url: &str) -> Arc<DatabaseEventSource> {
    Arc::new(DatabaseEventSource::new(
        table.to_string(),
        operations.iter().map(|&s| s.to_string()).collect(),
        database_url.to_string(),
    ))
}

/// Create a scheduled event source
pub fn scheduled_source(schedule_id: &str, cron: &str, timezone: &str) -> Arc<ScheduledEventSource> {
    Arc::new(ScheduledEventSource::new(
        schedule_id.to_string(),
        cron.to_string(),
        timezone.to_string(),
    ))
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_webhook_event_source() {
        let source = WebhookEventSource::new(
            "test".to_string(),
            8080,
            "/webhook".to_string(),
            None,
        );

        assert!(!source.is_running());
        assert_eq!(
            source.source_type().name(),
            "webhook:test"
        );
    }

    #[tokio::test]
    async fn test_database_event_source() {
        let source = DatabaseEventSource::new(
            "orders".to_string(),
            vec!["INSERT".to_string(), "UPDATE".to_string()],
            "postgres://localhost/test".to_string(),
        );

        assert!(!source.is_running());
        assert_eq!(
            source.source_type().name(),
            "db:orders:INSERT,UPDATE"
        );
    }

    #[tokio::test]
    async fn test_scheduled_event_source() {
        let source = ScheduledEventSource::new(
            "daily-report".to_string(),
            "0 9 * * *".to_string(),
            "UTC".to_string(),
        );

        assert!(!source.is_running());
        assert_eq!(
            source.source_type().name(),
            "scheduled:daily-report"
        );
    }
}
