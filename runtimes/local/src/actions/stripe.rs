//! Stripe action connector.
//!
//! Implements billing and payment actions via the Stripe API.
//!
//! ## Supported Actions
//!
//! | Action | Description | Idempotency Key |
//! |---------|-------------|-----------------|
//! | `charge` | Create a payment intent and confirm | `stripe:charge:{customer_id}:{amount}:{currency}` |
//! | `create_customer` | Create a new Stripe customer | `stripe:customer:{email}` |
//! | `update_customer` | Update customer metadata | `stripe:customer:{customer_id}` |
//! | `create_subscription` | Create a subscription | `stripe:sub:{customer_id}:{price_id}` |
//! | `cancel_subscription` | Cancel a subscription | `stripe:sub:{subscription_id}` |
//!
//! ## Credentials
//!
//! Reads `STRIPE_SECRET_KEY` from the environment.

use std::collections::HashMap;
use std::sync::Arc;
use std::time::Instant;

use serde::{Deserialize, Serialize};
use tracing::{info, instrument};

use super::connector::{ActionConnector, ActionError, ActionResult, IdempotencyKey};
use crate::actions::connector::IdempotencyCache;

/// Encode form data for Stripe API (application/x-www-form-urlencoded).
fn encode_form(params: &[(String, String)]) -> String {
    serde_urlencoded::to_string(params).unwrap_or_default()
}

/// Stripe connector for billing actions.
pub struct StripeConnector {
    client: reqwest::Client,
    api_key: String,
    idempotency_cache: Arc<IdempotencyCache>,
}

impl StripeConnector {
    /// Create a new Stripe connector.
    pub fn new(api_key: String) -> Self {
        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(30))
            .build()
            .expect("Stripe HTTP client must build");

        Self {
            client,
            api_key,
            idempotency_cache: Arc::new(IdempotencyCache::default()),
        }
    }

    /// Base URL for the Stripe API.
    const BASE_URL: &'static str = "https://api.stripe.com/v1";

    /// Build the Stripe API request headers.
    fn headers(&self) -> reqwest::header::HeaderMap {
        let mut headers = reqwest::header::HeaderMap::new();
        headers.insert(
            reqwest::header::AUTHORIZATION,
            format!("Bearer {}", self.api_key).parse().unwrap(),
        );
        headers.insert(
            reqwest::header::CONTENT_TYPE,
            "application/x-www-form-urlencoded".parse().unwrap(),
        );
        headers
    }
}

impl ActionConnector for StripeConnector {
    fn name(&self) -> &'static str {
        "stripe"
    }

    fn supports_action(&self, action: &str) -> bool {
        matches!(action, "charge" | "create_customer" | "update_customer"
            | "create_subscription" | "cancel_subscription" | "get_customer" | "get_subscription")
    }

    fn validate_params(&self, action: &str, params: &serde_json::Value) -> Result<(), String> {
        match action {
            "charge" => {
                let amount = params.get("amount")
                    .and_then(|v| v.as_i64())
                    .ok_or("charge requires numeric 'amount' (in cents)")?;
                if amount <= 0 {
                    return Err("charge amount must be positive".to_string());
                }
                let currency = params.get("currency")
                    .and_then(|v| v.as_str())
                    .unwrap_or("usd");
                if currency.len() != 3 {
                    return Err("currency must be a 3-letter ISO code".to_string());
                }
                Ok(())
            }
            "create_customer" => {
                params.get("email")
                    .and_then(|v| v.as_str())
                    .ok_or("create_customer requires 'email'")?;
                Ok(())
            }
            "create_subscription" => {
                params.get("customer_id")
                    .and_then(|v| v.as_str())
                    .ok_or("create_subscription requires 'customer_id'")?;
                params.get("price_id")
                    .and_then(|v| v.as_str())
                    .ok_or("create_subscription requires 'price_id'")?;
                Ok(())
            }
            "cancel_subscription" => {
                params.get("subscription_id")
                    .and_then(|v| v.as_str())
                    .ok_or("cancel_subscription requires 'subscription_id'")?;
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
            "charge" => {
                self.execute_charge(tenant_id, &params).await
            }
            "create_customer" => {
                self.execute_create_customer(&params).await
            }
            "get_customer" => {
                self.execute_get_customer(params.get("customer_id")
                    .and_then(|v| v.as_str())
                    .unwrap_or_default()).await
            }
            "create_subscription" => {
                self.execute_create_subscription(&params).await
            }
            "cancel_subscription" => {
                self.execute_cancel_subscription(&params).await
            }
            _ => Err(ActionError::fatal(format!("Unknown action: {}", action))),
        }
        .map(|mut result| {
            result.latency_ms = start.elapsed().as_millis() as u64;
            result
        })
    }
}

// ---------------------------------------------------------------------------
// Stripe action implementations
// ---------------------------------------------------------------------------

impl StripeConnector {
    /// Execute a charge action.
    async fn execute_charge(
        &self,
        tenant_id: Option<&str>,
        params: &serde_json::Value,
    ) -> Result<ActionResult, ActionError> {
        let amount = params.get("amount").and_then(|v| v.as_i64()).unwrap_or(0);
        let currency = params.get("currency").and_then(|v| v.as_str()).unwrap_or("usd");
        let customer_id = params.get("customer_id").and_then(|v| v.as_str());
        let description = params.get("description").and_then(|v| v.as_str());
        let metadata = params.get("metadata").and_then(|v| v.as_object());

        // Build form params
        let mut form = Vec::new();
        form.push(("amount".to_string(), amount.to_string()));
        form.push(("currency".to_string(), currency.to_string()));
        form.push(("description".to_string(), description.unwrap_or("FunctionFly charge").to_string()));

        if let Some(customer) = customer_id {
            form.push(("customer".to_string(), customer.to_string()));
        }

        // Add metadata
        if let Some(meta) = metadata {
            for (key, value) in meta {
                form.push((format!("metadata[{}]", key), value.to_string()));
            }
        }

        // Add idempotency key
        let ik = format!("{}-{}-{}-charge",
            tenant_id.unwrap_or("anon"),
            customer_id.unwrap_or("no-customer"),
            amount
        );
        form.push(("idempotency_key".to_string(), ik));

        let response = self.client
            .post(format!("{}/payment_intents", Self::BASE_URL))
            .headers(self.headers())
            .body(encode_form(&form))
            .send()
            .await
            .map_err(ActionError::from_reqwest)?;

        let status = response.status();
        let body = response.text().await.unwrap_or_default();

        if !status.is_success() {
            let code: Option<String> = serde_json::from_str::<serde_json::Value>(&body)
                .ok()
                .and_then(|v| v.get("error").and_then(|e| e.get("code")).map(|c| c.to_string()));

            return Err(ActionError {
                message: format!("Stripe charge failed: {}", body),
                code,
                retryable: status.as_u16() == 429 || status.is_server_error(),
                status_code: Some(status.as_u16()),
            });
        }

        let parsed: StripePaymentIntent = serde_json::from_str(&body)
            .map_err(|e| ActionError::fatal(format!("Failed to parse Stripe response: {}", e)))?;

        info!(
            payment_intent_id = %parsed.id,
            amount = parsed.amount,
            status = %parsed.status,
            "Stripe charge successful"
        );

        Ok(ActionResult::success(
            serde_json::json!({
                "payment_intent_id": parsed.id,
                "status": parsed.status,
                "amount": parsed.amount,
                "currency": parsed.currency,
            }),
            0,
        ).with_provider_ref(parsed.id))
    }

    /// Execute a create customer action.
    async fn execute_create_customer(&self, params: &serde_json::Value) -> Result<ActionResult, ActionError> {
        let email = params.get("email").and_then(|v| v.as_str()).unwrap_or_default();
        let name = params.get("name").and_then(|v| v.as_str());

        let mut form = vec![
            ("email".to_string(), email.to_string()),
        ];
        if let Some(n) = name {
            form.push(("name".to_string(), n.to_string()));
        }

        let response = self.client
            .post(format!("{}/customers", Self::BASE_URL))
            .headers(self.headers())
            .body(encode_form(&form))
            .send()
            .await
            .map_err(ActionError::from_reqwest)?;

        let status = response.status();
        let body = response.text().await.unwrap_or_default();

        if !status.is_success() {
            return Err(ActionError {
                message: format!("Stripe create_customer failed: {}", body),
                code: None,
                retryable: status.is_server_error(),
                status_code: Some(status.as_u16()),
            });
        }

        let parsed: StripeCustomer = serde_json::from_str(&body)
            .map_err(|e| ActionError::fatal(format!("Failed to parse Stripe response: {}", e)))?;

        Ok(ActionResult::success(
            serde_json::json!({
                "customer_id": parsed.id,
                "email": parsed.email,
            }),
            0,
        ).with_provider_ref(parsed.id))
    }

    async fn execute_get_customer(&self, customer_id: &str) -> Result<ActionResult, ActionError> {
        let response = self.client
            .get(format!("{}/customers/{}", Self::BASE_URL, customer_id))
            .headers(self.headers())
            .send()
            .await
            .map_err(ActionError::from_reqwest)?;

        let status = response.status();
        let body = response.text().await.unwrap_or_default();

        if !status.is_success() {
            return Err(ActionError {
                message: format!("Stripe get_customer failed: {}", body),
                code: None,
                retryable: status.is_server_error(),
                status_code: Some(status.as_u16()),
            });
        }

        let parsed: StripeCustomer = serde_json::from_str(&body)
            .map_err(|e| ActionError::fatal(format!("Failed to parse Stripe response: {}", e)))?;

        Ok(ActionResult::success(
            serde_json::json!({
                "customer_id": parsed.id,
                "email": parsed.email,
                "name": parsed.name,
            }),
            0,
        ))
    }

    async fn execute_create_subscription(&self, params: &serde_json::Value) -> Result<ActionResult, ActionError> {
        let customer_id = params.get("customer_id").and_then(|v| v.as_str()).unwrap_or_default();
        let price_id = params.get("price_id").and_then(|v| v.as_str()).unwrap_or_default();

        let form = vec![
            ("customer".to_string(), customer_id.to_string()),
            ("items[0][price]".to_string(), price_id.to_string()),
        ];

        let response = self.client
            .post(format!("{}/subscriptions", Self::BASE_URL))
            .headers(self.headers())
            .body(encode_form(&form))
            .send()
            .await
            .map_err(ActionError::from_reqwest)?;

        let status = response.status();
        let body = response.text().await.unwrap_or_default();

        if !status.is_success() {
            return Err(ActionError {
                message: format!("Stripe create_subscription failed: {}", body),
                code: None,
                retryable: status.is_server_error(),
                status_code: Some(status.as_u16()),
            });
        }

        let parsed: StripeSubscription = serde_json::from_str(&body)
            .map_err(|e| ActionError::fatal(format!("Failed to parse Stripe response: {}", e)))?;

        Ok(ActionResult::success(
            serde_json::json!({
                "subscription_id": parsed.id,
                "status": parsed.status,
            }),
            0,
        ).with_provider_ref(parsed.id))
    }

    async fn execute_cancel_subscription(&self, params: &serde_json::Value) -> Result<ActionResult, ActionError> {
        let subscription_id = params.get("subscription_id").and_then(|v| v.as_str()).unwrap_or_default();

        let form = vec![("cancel_at_period_end".to_string(), "true".to_string())];

        let response = self.client
            .post(format!("{}/subscriptions/{}", Self::BASE_URL, subscription_id))
            .headers(self.headers())
            .body(encode_form(&form))
            .send()
            .await
            .map_err(ActionError::from_reqwest)?;

        let status = response.status();
        let body = response.text().await.unwrap_or_default();

        if !status.is_success() {
            return Err(ActionError {
                message: format!("Stripe cancel_subscription failed: {}", body),
                code: None,
                retryable: status.is_server_error(),
                status_code: Some(status.as_u16()),
            });
        }

        let parsed: StripeSubscription = serde_json::from_str(&body)
            .map_err(|e| ActionError::fatal(format!("Failed to parse Stripe response: {}", e)))?;

        Ok(ActionResult::success(
            serde_json::json!({
                "subscription_id": parsed.id,
                "cancel_at_period_end": parsed.cancel_at_period_end,
            }),
            0,
        ))
    }
}

// ---------------------------------------------------------------------------
// Stripe API response types
// ---------------------------------------------------------------------------

#[derive(Debug, Deserialize)]
struct StripePaymentIntent {
    id: String,
    amount: i64,
    currency: String,
    status: String,
}

#[derive(Debug, Deserialize)]
struct StripeCustomer {
    id: String,
    email: String,
    name: Option<String>,
}

#[derive(Debug, Deserialize)]
struct StripeSubscription {
    id: String,
    status: String,
    cancel_at_period_end: bool,
}