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
use chrono::Utc;
use chrono_tz::Tz;
use hmac::{Hmac, KeyInit, Mac};
use sha2::Sha256;
use tokio::sync::mpsc;
use tracing::{debug, error, info, warn};
use uuid::Uuid;

use crate::daemon::agent_daemon::{AgentEvent, EventSource, EventSourceType};

type HmacSha256 = Hmac<Sha256>;

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
    hmac_secret: Option<Vec<u8>>,
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
            hmac_secret: None,
        }
    }

    /// Create a new webhook event source with HMAC secret for signature verification
    pub fn with_hmac_secret(mut self, secret: Vec<u8>) -> Self {
        self.hmac_secret = Some(secret);
        self
    }

    /// Validate webhook HMAC-SHA256 signature.
    ///
    /// For Stripe: header is `t=timestamp,v1=signature` — we verify `v1` over `timestamp.body`
    /// For Shopify: header is base64-encoded HMAC-SHA256 of the body
    /// For generic: header is hex-encoded HMAC-SHA256 of the body
    fn validate_signature(&self, headers: &[(String, String)], body: &[u8]) -> bool {
        let secret = match &self.hmac_secret {
            Some(s) => s,
            None => {
                warn!(name = %self.name, "No HMAC secret configured — rejecting webhook");
                return false;
            }
        };

        let header_name = match &self.secret_header {
            Some(h) => h.to_lowercase(),
            None => {
                warn!(name = %self.name, "No signature header configured — rejecting webhook");
                return false;
            }
        };

        let signature_value = headers
            .iter()
            .find(|(k, _)| k.to_lowercase() == header_name)
            .map(|(_, v)| v.clone());

        let sig = match signature_value {
            Some(s) => s,
            None => {
                warn!(name = %self.name, header = %header_name, "Missing signature header");
                return false;
            }
        };

        match self.name.as_str() {
            "stripe" => self.verify_stripe_signature(secret, &sig, body),
            "shopify" => self.verify_shopify_signature(secret, &sig, body),
            _ => self.verify_generic_hmac(secret, &sig, body),
        }
    }

    /// Verify Stripe webhook signature (v1 scheme).
    /// Format: `t=timestamp,v1=hex_signature`
    /// Signed payload: `timestamp.body`
    fn verify_stripe_signature(&self, secret: &[u8], header: &str, body: &[u8]) -> bool {
        let mut timestamp = "";
        let mut sig_hex = "";

        for part in header.split(',') {
            if let Some(t) = part.strip_prefix("t=") {
                timestamp = t;
            } else if let Some(v) = part.strip_prefix("v1=") {
                sig_hex = v;
            }
        }

        if timestamp.is_empty() || sig_hex.is_empty() {
            warn!(name = %self.name, "Invalid Stripe signature format");
            return false;
        }

        // Check timestamp freshness (reject if older than 5 minutes)
        if let Ok(ts) = timestamp.parse::<i64>() {
            let now = chrono::Utc::now().timestamp();
            if (now - ts).abs() > 300 {
                warn!(name = %self.name, ts = ts, now = now, "Stripe webhook timestamp too old or in future");
                return false;
            }
        }

        let signed_payload = format!("{}.{}", timestamp, String::from_utf8_lossy(body));

        let mut mac = match HmacSha256::new_from_slice(secret) {
            Ok(m) => m,
            Err(e) => {
                error!(name = %self.name, error = %e, "Failed to create HMAC");
                return false;
            }
        };
        mac.update(signed_payload.as_bytes());
        let expected = hex::encode(mac.finalize().into_bytes());

        if expected != sig_hex {
            warn!(name = %self.name, "Stripe HMAC signature mismatch");
            return false;
        }

        true
    }

    /// Verify Shopify webhook signature.
    /// Header is base64-encoded HMAC-SHA256 of the raw body.
    fn verify_shopify_signature(&self, secret: &[u8], header: &str, body: &[u8]) -> bool {
        use base64::Engine as _;

        let expected_bytes = match base64::engine::general_purpose::STANDARD.decode(header.trim()) {
            Ok(b) => b,
            Err(e) => {
                warn!(name = %self.name, error = %e, "Invalid base64 in Shopify signature");
                return false;
            }
        };

        let mut mac = match HmacSha256::new_from_slice(secret) {
            Ok(m) => m,
            Err(e) => {
                error!(name = %self.name, error = %e, "Failed to create HMAC");
                return false;
            }
        };
        mac.update(body);
        let computed = mac.finalize().into_bytes();

        // Constant-time comparison
        if computed.len() != expected_bytes.len() {
            warn!(name = %self.name, "Shopify HMAC length mismatch");
            return false;
        }

        let mut result = 0u8;
        for (a, b) in computed.iter().zip(expected_bytes.iter()) {
            result |= a ^ b;
        }

        if result != 0 {
            warn!(name = %self.name, "Shopify HMAC signature mismatch");
            return false;
        }

        true
    }

    /// Verify generic hex-encoded HMAC-SHA256 of the body.
    fn verify_generic_hmac(&self, secret: &[u8], header: &str, body: &[u8]) -> bool {
        let mut mac = match HmacSha256::new_from_slice(secret) {
            Ok(m) => m,
            Err(e) => {
                error!(name = %self.name, error = %e, "Failed to create HMAC");
                return false;
            }
        };
        mac.update(body);
        let expected = hex::encode(mac.finalize().into_bytes());

        if expected != header.trim() {
            warn!(name = %self.name, "Generic HMAC signature mismatch");
            return false;
        }

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
        let (tx, rx) = mpsc::channel(100);
        self.running.store(true, Ordering::SeqCst);

        info!(
            name = %self.name,
            port = self.port,
            path = %self.path,
            "Starting webhook event source"
        );

        let name = self.name.clone();
        let path = self.path.clone();
        let running = self.running.clone();
        let hmac_secret = self.hmac_secret.clone();
        let secret_header = self.secret_header.clone();
        let port = self.port;

        tokio::spawn(async move {
            use axum::{extract::State as AxumState, routing::post, Router};
            use std::net::SocketAddr;

            type WebhookState = Arc<(
                tokio::sync::Mutex<mpsc::Sender<AgentEvent>>,
                String,
                Option<Vec<u8>>,
                Option<String>,
            )>;

            let shared_state: WebhookState = Arc::new((
                tokio::sync::Mutex::new(tx),
                name.clone(),
                hmac_secret.clone(),
                secret_header.clone(),
            ));

            let webhook_handler = move |
                AxumState(state): AxumState<WebhookState>,
                headers: axum::http::HeaderMap,
                body: axum::body::Bytes,
            | {
                let state = state.clone();
                let body_vec = body.to_vec();

                async move {
                    let (ref tx, ref name, ref hmac_secret, ref secret_header) = *state;
                    // Extract headers for HMAC verification
                    let header_pairs: Vec<(String, String)> = headers
                        .iter()
                        .map(|(k, v)| (k.to_string(), v.to_str().unwrap_or("").to_string()))
                        .collect();

                    // Verify HMAC if secret is configured
                    if let Some(ref secret) = hmac_secret {
                        let header_name = secret_header.as_deref().unwrap_or("x-signature-256");
                        let sig = header_pairs
                            .iter()
                            .find(|(k, _)| k.to_lowercase() == header_name.to_lowercase())
                            .map(|(_, v)| v.clone());

                        let sig = match sig {
                            Some(s) => s,
                            None => {
                                warn!(name = %name, "Webhook rejected: missing signature header");
                                return axum::http::StatusCode::UNAUTHORIZED;
                            }
                        };

                        // Generic HMAC-SHA256 verification
                        let mut mac = match HmacSha256::new_from_slice(secret) {
                            Ok(m) => m,
                            Err(_) => return axum::http::StatusCode::INTERNAL_SERVER_ERROR,
                        };
                        mac.update(&body_vec);
                        let expected = hex::encode(mac.finalize().into_bytes());

                        if expected != sig.trim() {
                            warn!(name = %name, "Webhook rejected: HMAC signature mismatch");
                            return axum::http::StatusCode::UNAUTHORIZED;
                        }
                    }

                    // Parse body as JSON
                    let payload = match serde_json::from_slice::<serde_json::Value>(&body_vec) {
                        Ok(v) => v,
                        Err(e) => {
                            warn!(name = %name, error = %e, "Failed to parse webhook body");
                            return axum::http::StatusCode::BAD_REQUEST;
                        }
                    };

                    let event = AgentEvent {
                        id: Uuid::new_v4(),
                        source: EventSourceType::Webhook { name: name.clone() },
                        agent_id: String::new(),
                        payload,
                        timestamp: Instant::now(),
                    };

                    let sender = tx.lock().await;
                    if let Err(e) = sender.send(event).await {
                        error!(name = %name, error = %e, "Failed to send webhook event");
                        return axum::http::StatusCode::INTERNAL_SERVER_ERROR;
                    }

                    info!(name = %name, "Webhook event received and dispatched");
                    axum::http::StatusCode::OK
                }
            };

            let app = Router::new()
                .route(&path, post(webhook_handler))
                .with_state(shared_state);

            let addr = SocketAddr::from(([127, 0, 0, 1], port));
            match tokio::net::TcpListener::bind(addr).await {
                Ok(listener) => {
                    info!(name = %name, port = port, path = %path, "Webhook HTTP server listening");
                    if let Err(e) = axum::serve(listener, app).await {
                        error!(name = %name, error = %e, "Webhook server error");
                    }
                }
                Err(e) => {
                    error!(name = %name, port = port, error = %e, "Failed to bind webhook server");
                }
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
        let (tx, rx) = mpsc::channel(100);
        self.running.store(true, Ordering::SeqCst);

        info!(
            table = %self.table,
            channel = %self.channel,
            "Starting database event source"
        );

        let channel = self.channel.clone();
        let database_url = self.database_url.clone();
        let table = self.table.clone();
        let operations = self.operations.clone();
        let running = self.running.clone();

        tokio::spawn(async move {
            // Connect to PostgreSQL using tokio-postgres
            let (client, connection) = match tokio_postgres::connect(&database_url, tokio_postgres::NoTls).await {
                Ok((c, conn)) => (c, conn),
                Err(e) => {
                    error!(channel = %channel, error = %e, "Failed to connect to PostgreSQL for LISTEN/NOTIFY");
                    return;
                }
            };

            // Spawn connection driver
            tokio::spawn(async move {
                if let Err(e) = connection.await {
                    error!(error = %e, "PostgreSQL connection error in LISTEN task");
                }
            });

            // Execute LISTEN on the channel
            let listen_sql = format!("LISTEN {}", channel);
            if let Err(e) = client.execute(&listen_sql, &[]).await {
                error!(channel = %channel, error = %e, "Failed to execute LISTEN");
                return;
            }

            info!(channel = %channel, table = %table, "PostgreSQL LISTEN active — waiting for notifications");

            // Poll for pending notifications via a periodic check
            // tokio-postgres delivers NOTIFY through the connection future,
            // so we use a polling approach with a check query.
            loop {
                if !running.load(Ordering::SeqCst) {
                    break;
                }

                // Check for pending notifications via a simple query
                // The PostgreSQL NOTIFY/LISTEN mechanism delivers notifications
                // through the connection, but tokio-postgres handles them via
                // the connection future. We use a polling approach with pg_notify.
                let check_sql = format!(
                    "SELECT pg_notify('{}', row_to_json(t)::text) FROM (SELECT * FROM {} WHERE updated_at > NOW() - INTERVAL '5 seconds' LIMIT 10) t",
                    channel, table
                );

                match client.query(&check_sql, &[]).await {
                    Ok(rows) => {
                        for row in &rows {
                            let payload: Option<String> = row.get(0);
                            if let Some(payload_str) = payload {
                                match serde_json::from_str::<serde_json::Value>(&payload_str) {
                                    Ok(data) => {
                                        let operation = data.get("operation")
                                            .and_then(|v| v.as_str())
                                            .unwrap_or("UNKNOWN");

                                        if operations.iter().any(|op| op == operation) || operations.is_empty() {
                                            let event = AgentEvent {
                                                id: Uuid::new_v4(),
                                                source: EventSourceType::Database {
                                                    table: table.clone(),
                                                    operation: operations.join(","),
                                                },
                                                agent_id: String::new(),
                                                payload: data,
                                                timestamp: Instant::now(),
                                            };

                                            if let Err(e) = tx.send(event).await {
                                                error!(error = %e, "Failed to send database event");
                                                break;
                                            }
                                        }
                                    }
                                    Err(e) => {
                                        warn!(error = %e, "Failed to parse notification payload");
                                    }
                                }
                            }
                        }
                    }
                    Err(e) => {
                        debug!(error = %e, "Database event check query failed (may be no matching rows)");
                    }
                }

                tokio::time::sleep(Duration::from_secs(5)).await;
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
        compute_next_trigger(&self.cron_expression, &self.timezone)
    }
}

/// Compute the next trigger instant from a cron expression and timezone
fn compute_next_trigger(cron_expression: &str, timezone: &str) -> Option<Instant> {
    let now = Utc::now();
    let tz: Tz = timezone.parse().unwrap_or(Utc);
    let now_in_tz = now.with_timezone(&tz);

    let next = cron_parser::parse(cron_expression, now_in_tz)
        .map_err(|e| warn!(error = %e, "Failed to parse cron expression"))
        .ok()?;

    Some(Instant::from_secs(next.timestamp() as u64))
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
        let cron_expression = self.cron_expression.clone();
        let timezone = self.timezone.clone();
        let running = self.running.clone();

        tokio::spawn(async move {
            while running.load(Ordering::SeqCst) {
                let sleep_duration = compute_next_trigger(&cron_expression, &timezone)
                    .map(|t| t.saturating_duration_since(Instant::now()))
                    .unwrap_or(Duration::from_secs(60));

                tokio::time::sleep(sleep_duration).await;

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
    let secret = std::env::var("STRIPE_WEBHOOK_SECRET")
        .map(|s| s.into_bytes())
        .unwrap_or_default();
    Arc::new(
        WebhookEventSource::new(
            "stripe".to_string(),
            port,
            "/webhooks/stripe".to_string(),
            Some("Stripe-Signature".to_string()),
        )
        .with_hmac_secret(secret),
    )
}

/// Create a webhook event source for Shopify
pub fn shopify_webhook_source(port: u16) -> Arc<WebhookEventSource> {
    let secret = std::env::var("SHOPIFY_WEBHOOK_SECRET")
        .map(|s| s.into_bytes())
        .unwrap_or_default();
    Arc::new(
        WebhookEventSource::new(
            "shopify".to_string(),
            port,
            "/webhooks/shopify".to_string(),
            Some("X-Shopify-Hmac-SHA256".to_string()),
        )
        .with_hmac_secret(secret),
    )
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
    async fn test_webhook_hmac_rejects_without_secret() {
        let source = WebhookEventSource::new(
            "test".to_string(),
            8080,
            "/webhook".to_string(),
            Some("X-Signature".to_string()),
        );
        // No secret configured — should reject
        let headers = vec![("X-Signature".to_string(), "abc123".to_string())];
        assert!(!source.validate_signature(&headers, b"body"));
    }

    #[tokio::test]
    async fn test_webhook_hmac_rejects_missing_header() {
        let source = WebhookEventSource::new(
            "test".to_string(),
            8080,
            "/webhook".to_string(),
            Some("X-Signature".to_string()),
        )
        .with_hmac_secret(b"test-secret".to_vec());
        // Missing header — should reject
        let headers: Vec<(String, String)> = vec![];
        assert!(!source.validate_signature(&headers, b"body"));
    }

    #[tokio::test]
    async fn test_webhook_generic_hmac_valid() {
        use hmac::{Hmac, KeyInit, Mac};
        use sha2::Sha256;
        type HmacSha256 = Hmac<Sha256>;

        let secret = b"test-secret-key";
        let body = b"hello world";

        let mut mac = HmacSha256::new_from_slice(secret).unwrap();
        mac.update(body);
        let sig = hex::encode(mac.finalize().into_bytes());

        let source = WebhookEventSource::new(
            "generic".to_string(),
            8080,
            "/webhook".to_string(),
            Some("X-Signature-256".to_string()),
        )
        .with_hmac_secret(secret.to_vec());

        let headers = vec![("X-Signature-256".to_string(), sig)];
        assert!(source.validate_signature(&headers, body));
    }

    #[tokio::test]
    async fn test_webhook_generic_hmac_invalid() {
        let source = WebhookEventSource::new(
            "generic".to_string(),
            8080,
            "/webhook".to_string(),
            Some("X-Signature-256".to_string()),
        )
        .with_hmac_secret(b"test-secret".to_vec());

        let headers = vec![("X-Signature-256".to_string(), "invalid_signature".to_string())];
        assert!(!source.validate_signature(&headers, b"body"));
    }

    #[tokio::test]
    async fn test_webhook_stripe_hmac_valid() {
        use hmac::{Hmac, KeyInit, Mac};
        use sha2::Sha256;
        type HmacSha256 = Hmac<Sha256>;

        let secret = b"whsec_test123";
        let body = b"{}";
        let timestamp = chrono::Utc::now().timestamp().to_string();
        let signed_payload = format!("{}.{}", timestamp, String::from_utf8_lossy(body));

        let mut mac = HmacSha256::new_from_slice(secret).unwrap();
        mac.update(signed_payload.as_bytes());
        let sig = hex::encode(mac.finalize().into_bytes());

        let source = WebhookEventSource::new(
            "stripe".to_string(),
            8080,
            "/webhooks/stripe".to_string(),
            Some("Stripe-Signature".to_string()),
        )
        .with_hmac_secret(secret.to_vec());

        let header = format!("t={},v1={}", timestamp, sig);
        let headers = vec![("Stripe-Signature".to_string(), header)];
        assert!(source.validate_signature(&headers, body));
    }

    #[tokio::test]
    async fn test_webhook_stripe_hmac_expired() {
        use hmac::{Hmac, KeyInit, Mac};
        use sha2::Sha256;
        type HmacSha256 = Hmac<Sha256>;

        let secret = b"whsec_test123";
        let body = b"{}";
        // Use a timestamp from 10 minutes ago
        let timestamp = (chrono::Utc::now().timestamp() - 600).to_string();
        let signed_payload = format!("{}.{}", timestamp, String::from_utf8_lossy(body));

        let mut mac = HmacSha256::new_from_slice(secret).unwrap();
        mac.update(signed_payload.as_bytes());
        let sig = hex::encode(mac.finalize().into_bytes());

        let source = WebhookEventSource::new(
            "stripe".to_string(),
            8080,
            "/webhooks/stripe".to_string(),
            Some("Stripe-Signature".to_string()),
        )
        .with_hmac_secret(secret.to_vec());

        let header = format!("t={},v1={}", timestamp, sig);
        let headers = vec![("Stripe-Signature".to_string(), header)];
        // Should reject — timestamp older than 5 minutes
        assert!(!source.validate_signature(&headers, body));
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
