//! Resend email action connector.
//!
//! Implements email sending actions via the Resend API.
//!
//! ## Supported Actions
//!
//! | Action | Description |
//! |---------|-------------|
//! | `send_email` | Send a transactional email |
//! | `send_verification` | Send email verification link |
//! | `send_password_reset` | Send password reset email |
//!
//! ## Credentials
//!
//! Reads `RESEND_API_KEY` from the environment.

use std::sync::Arc;
use std::time::Instant;

use serde::{Deserialize, Serialize};
use tracing::{debug, info, instrument};

use super::connector::{ActionConnector, ActionError, ActionResult, IdempotencyKey};
use crate::actions::connector::IdempotencyCache;

/// Resend email connector.
pub struct ResendConnector {
    client: reqwest::Client,
    api_key: String,
    from_email: String,
    from_name: String,
    idempotency_cache: Arc<IdempotencyCache>,
}

impl ResendConnector {
    /// Create a new Resend connector.
    pub fn new(api_key: String, from_email: String, from_name: String) -> Self {
        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(30))
            .build()
            .expect("Resend HTTP client must build");

        Self {
            client,
            api_key,
            from_email,
            from_name,
            idempotency_cache: Arc::new(IdempotencyCache::default()),
        }
    }

    const BASE_URL: &'static str = "https://api.resend.com";
}

impl ActionConnector for ResendConnector {
    fn name(&self) -> &'static str {
        "resend"
    }

    fn supports_action(&self, action: &str) -> bool {
        matches!(action, "send_email" | "send_verification" | "send_password_reset" | "send_magic_link")
    }

    fn validate_params(&self, action: &str, params: &serde_json::Value) -> Result<(), String> {
        match action {
            "send_email" => {
                params.get("to")
                    .and_then(|v| v.as_str())
                    .ok_or("send_email requires 'to' (email)")?;
                params.get("subject")
                    .and_then(|v| v.as_str())
                    .ok_or("send_email requires 'subject'")?;
                params.get("html")
                    .and_then(|v| v.as_str())
                    .ok_or("send_email requires 'html' body")?;
                Ok(())
            }
            "send_verification" | "send_password_reset" | "send_magic_link" => {
                params.get("to")
                    .and_then(|v| v.as_str())
                    .ok_or("verification email requires 'to' (email)")?;
                params.get("link")
                    .and_then(|v| v.as_str())
                    .ok_or("verification email requires 'link'")?;
                Ok(())
            }
            _ => Ok(()),
        }
    }

    #[instrument(skip_all, fields(action = %action))]
    async fn execute(
        &self,
        tenant_id: Option<&str>,
        action: &str,
        params: serde_json::Value,
        _idempotency_key: &IdempotencyKey,
    ) -> Result<ActionResult, ActionError> {
        let start = Instant::now();

        match action {
            "send_email" => self.execute_send_email(&params).await,
            "send_verification" => self.execute_send_verification(tenant_id, &params).await,
            "send_password_reset" => self.execute_send_password_reset(tenant_id, &params).await,
            "send_magic_link" => self.execute_send_magic_link(tenant_id, &params).await,
            _ => Err(ActionError::fatal(format!("Unknown action: {}", action))),
        }
        .map(|mut result| {
            result.latency_ms = start.elapsed().as_millis() as u64;
            result
        })
    }
}

// ---------------------------------------------------------------------------
// Resend action implementations
// ---------------------------------------------------------------------------

impl ResendConnector {
    async fn execute_send_email(&self, params: &serde_json::Value) -> Result<ActionResult, ActionError> {
        let to = params.get("to").and_then(|v| v.as_str()).unwrap_or_default();
        let subject = params.get("subject").and_then(|v| v.as_str()).unwrap_or_default();
        let html = params.get("html").and_then(|v| v.as_str()).unwrap_or_default();
        let text = params.get("text").and_then(|v| v.as_str());

        #[derive(Serialize)]
        struct EmailRequest<'a> {
            from: &'a str,
            to: &'a str,
            subject: &'a str,
            html: &'a str,
            text: Option<&'a str>,
        }

        let email_body = EmailRequest {
            from: &format!("{} <{}>", self.from_name, self.from_email),
            to,
            subject,
            html,
            text,
        };

        let response = self.client
            .post(format!("{}/emails", Self::BASE_URL))
            .header("Authorization", format!("Bearer {}", self.api_key))
            .header("Content-Type", "application/json")
            .json(&email_body)
            .send()
            .await
            .map_err(ActionError::from_reqwest)?;

        let status = response.status();
        let body = response.text().await.unwrap_or_default();

        if !status.is_success() {
            return Err(ActionError {
                message: format!("Resend send_email failed: {}", body),
                code: None,
                retryable: status.is_server_error() || status.as_u16() == 429,
                status_code: Some(status.as_u16()),
            });
        }

        #[derive(Deserialize)]
        struct ResendResponse { id: String }

        let parsed: ResendResponse = serde_json::from_str(&body)
            .map_err(|e| ActionError::fatal(format!("Failed to parse Resend response: {}", e)))?;

        info!(email_id = %parsed.id, to = %to, "Resend email sent successfully");

        Ok(ActionResult::success(
            serde_json::json!({
                "email_id": parsed.id,
                "to": to,
            }),
            0,
        ).with_provider_ref(parsed.id))
    }

    async fn execute_send_verification(
        &self,
        tenant_id: Option<&str>,
        params: &serde_json::Value,
    ) -> Result<ActionResult, ActionError> {
        let to = params.get("to").and_then(|v| v.as_str()).unwrap_or_default();
        let link = params.get("link").and_then(|v| v.as_str()).unwrap_or_default();

        let html = format!(
            r#"<!DOCTYPE html><html><body style="font-family: sans-serif;">
            <h2>Verify your email</h2>
            <p>Click the link below to verify your email address:</p>
            <a href="{}">Verify Email</a>
            <p style="color: #666; font-size: 12px;">This link expires in 24 hours.</p>
            </body></html>"#,
            link
        );

        let tenant_label = tenant_id.unwrap_or("functionfly");
        let subject = format!("[{}] Verify your email", tenant_label);

        let email_params = serde_json::json!({
            "to": to,
            "subject": subject,
            "html": html,
        });

        self.execute_send_email(&email_params).await
    }

    async fn execute_send_password_reset(
        &self,
        tenant_id: Option<&str>,
        params: &serde_json::Value,
    ) -> Result<ActionResult, ActionError> {
        let to = params.get("to").and_then(|v| v.as_str()).unwrap_or_default();
        let link = params.get("link").and_then(|v| v.as_str()).unwrap_or_default();

        let html = format!(
            r#"<!DOCTYPE html><html><body style="font-family: sans-serif;">
            <h2>Reset your password</h2>
            <p>Click the link below to reset your password:</p>
            <a href="{}">Reset Password</a>
            <p style="color: #666; font-size: 12px;">This link expires in 1 hour. If you didn't request this, ignore this email.</p>
            </body></html>"#,
            link
        );

        let tenant_label = tenant_id.unwrap_or("functionfly");
        let subject = format!("[{}] Reset your password", tenant_label);

        let email_params = serde_json::json!({
            "to": to,
            "subject": subject,
            "html": html,
        });

        self.execute_send_email(&email_params).await
    }

    async fn execute_send_magic_link(
        &self,
        tenant_id: Option<&str>,
        params: &serde_json::Value,
    ) -> Result<ActionResult, ActionError> {
        let to = params.get("to").and_then(|v| v.as_str()).unwrap_or_default();
        let link = params.get("link").and_then(|v| v.as_str()).unwrap_or_default();

        let html = format!(
            r#"<!DOCTYPE html><html><body style="font-family: sans-serif;">
            <h2>Your magic link</h2>
            <p>Click the link below to sign in:</p>
            <a href="{}">Sign In</a>
            <p style="color: #666; font-size: 12px;">This link expires in 15 minutes.</p>
            </body></html>"#,
            link
        );

        let tenant_label = tenant_id.unwrap_or("functionfly");
        let subject = format!("[{}] Your magic link", tenant_label);

        let email_params = serde_json::json!({
            "to": to,
            "subject": subject,
            "html": html,
        });

        self.execute_send_email(&email_params).await
    }
}