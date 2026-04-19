//! Shopify e-commerce action connector.
//!
//! Implements e-commerce operations via the Shopify Admin REST API.
//!
//! ## Supported Actions
//!
//! | Action | Description |
//! |--------|-------------|
//! | `create_order` | Create a new order |
//! | `get_order` | Retrieve order details |
//! | `update_order` | Update order status/notes |
//! | `create_customer` | Create a new customer |
//! | `get_customer` | Retrieve customer details |
//! | `update_inventory` | Update product inventory |
//! | `get_product` | Retrieve product details |
//! | `search_products` | Search products by keyword |
//!
//! ## Credentials
//!
//! Requires:
//! - `SHOPIFY_SHOP_DOMAIN` (e.g., "my-shop.myshopify.com")
//! - `SHOPIFY_ACCESS_TOKEN` (Admin API access token)
//! - `SHOPIFY_API_VERSION` (default: "2024-01")

use std::collections::HashMap;
use std::time::Instant;

use serde::Deserialize;
use tracing::{info, instrument};

use super::connector::{ActionConnector, ActionError, ActionResult, IdempotencyKey};

/// Shopify connector for e-commerce actions.
pub struct ShopifyConnector {
    client: reqwest::Client,
    shop_domain: String,
    access_token: String,
    api_version: String,
}

impl ShopifyConnector {
    /// Create a new Shopify connector.
    pub fn new(shop_domain: String, access_token: String, api_version: String) -> Self {
        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(30))
            .build()
            .expect("Shopify HTTP client must build");

        // Ensure shop_domain has protocol
        let shop = if shop_domain.starts_with("http") {
            shop_domain
        } else {
            format!("https://{}", shop_domain)
        };

        Self {
            client,
            shop_domain: shop,
            access_token,
            api_version,
        }
    }

    /// Create from environment variables.
    pub fn from_env() -> Option<Self> {
        let shop_domain = std::env::var("SHOPIFY_SHOP_DOMAIN").ok()?;
        let access_token = std::env::var("SHOPIFY_ACCESS_TOKEN").ok()?;
        let api_version = std::env::var("SHOPIFY_API_VERSION")
            .unwrap_or_else(|_| "2024-01".to_string());

        Some(Self::new(shop_domain, access_token, api_version))
    }

    /// Build the base API URL.
    fn base_url(&self) -> String {
        format!("{}/admin/api/{}", self.shop_domain, self.api_version)
    }

    /// Build request headers.
    fn headers(&self) -> reqwest::header::HeaderMap {
        let mut headers = reqwest::header::HeaderMap::new();
        headers.insert(
            reqwest::header::AUTHORIZATION,
            format!("X-Shopify-Access-Token {}", self.access_token).parse().unwrap(),
        );
        headers.insert(
            reqwest::header::CONTENT_TYPE,
            "application/json".parse().unwrap(),
        );
        headers
    }

    /// Handle Shopify API errors.
    fn handle_error(&self, status: reqwest::StatusCode, body: &str) -> ActionError {
        #[derive(Deserialize)]
        struct ShopifyError {
            errors: Option<serde_json::Value>,
        }

        let error_detail = serde_json::from_str::<ShopifyError>(body)
            .ok()
            .and_then(|e| e.errors)
            .map(|e| e.to_string())
            .unwrap_or_else(|| body.to_string());

        let retryable = status.is_server_error()
            || status.as_u16() == 429
            || status.as_u16() == 423; // Shopify rate limit

        ActionError {
            message: format!("Shopify API error: {}", error_detail),
            code: Some(status.to_string()),
            retryable,
            status_code: Some(status.as_u16()),
        }
    }
}

impl ActionConnector for ShopifyConnector {
    fn name(&self) -> &'static str {
        "shopify"
    }

    fn supports_action(&self, action: &str) -> bool {
        matches!(action,
            "create_order" | "get_order" | "update_order" |
            "create_customer" | "get_customer" | "update_customer" |
            "get_product" | "search_products" | "update_inventory"
        )
    }

    fn validate_params(&self, action: &str, params: &serde_json::Value) -> Result<(), String> {
        match action {
            "create_order" => {
                params.get("line_items")
                    .and_then(|v| v.as_array())
                    .filter(|arr| !arr.is_empty())
                    .ok_or("create_order requires non-empty 'line_items' array")?;
                Ok(())
            }
            "get_order" | "update_order" => {
                params.get("order_id")
                    .and_then(|v| v.as_str())
                    .ok_or("get_order/update_order requires 'order_id'")?;
                Ok(())
            }
            "create_customer" => {
                params.get("email")
                    .and_then(|v| v.as_str())
                    .ok_or("create_customer requires 'email'")?;
                Ok(())
            }
            "get_customer" | "update_customer" => {
                params.get("customer_id")
                    .and_then(|v| v.as_str())
                    .ok_or("get_customer/update_customer requires 'customer_id'")?;
                Ok(())
            }
            "get_product" | "update_inventory" => {
                params.get("product_id")
                    .and_then(|v| v.as_str())
                    .or_else(|| params.get("variant_id").and_then(|v| v.as_str()))
                    .ok_or("requires 'product_id' or 'variant_id'")?;
                Ok(())
            }
            "search_products" => {
                params.get("query")
                    .and_then(|v| v.as_str())
                    .ok_or("search_products requires 'query' string")?;
                Ok(())
            }
            _ => Ok(()),
        }
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

        let result = match action {
            "create_order" => self.create_order(&params).await,
            "get_order" => self.get_order(&params).await,
            "update_order" => self.update_order(&params).await,
            "create_customer" => self.create_customer(&params).await,
            "get_customer" => self.get_customer(&params).await,
            "update_customer" => self.update_customer(&params).await,
            "get_product" => self.get_product(&params).await,
            "search_products" => self.search_products(&params).await,
            "update_inventory" => self.update_inventory(&params).await,
            _ => Err(ActionError::fatal(format!("Unknown action: {}", action))),
        };

        result.map(|mut r| {
            r.latency_ms = start.elapsed().as_millis() as u64;
            r
        })
    }
}

// ---------------------------------------------------------------------------
// Shopify action implementations
// ---------------------------------------------------------------------------

impl ShopifyConnector {
    async fn create_order(&self, params: &serde_json::Value) -> Result<ActionResult, ActionError> {
        let url = format!("{}/orders.json", self.base_url());

        // Build order payload
        let mut order_payload = serde_json::json!({
            "order": {
                "line_items": params.get("line_items").cloned().unwrap_or_default(),
            }
        });

        // Add optional fields
        if let Some(obj) = order_payload["order"].as_object_mut() {
            if let Some(customer_id) = params.get("customer_id").and_then(|v| v.as_i64()) {
                obj.insert("customer".to_string(), serde_json::json!({ "id": customer_id }));
            }
            if let Some(email) = params.get("email").and_then(|v| v.as_str()) {
                obj.insert("email".to_string(), serde_json::Value::String(email.to_string()));
            }
            if let Some(note) = params.get("note").and_then(|v| v.as_str()) {
                obj.insert("note".to_string(), serde_json::Value::String(note.to_string()));
            }
            if let Some(tags) = params.get("tags").and_then(|v| v.as_str()) {
                obj.insert("tags".to_string(), serde_json::Value::String(tags.to_string()));
            }
            if let Some(financial_status) = params.get("financial_status").and_then(|v| v.as_str()) {
                obj.insert("financial_status".to_string(), serde_json::Value::String(financial_status.to_string()));
            }
            if let Some(fulfillment_status) = params.get("fulfillment_status").and_then(|v| v.as_str()) {
                obj.insert("fulfillment_status".to_string(), serde_json::Value::String(fulfillment_status.to_string()));
            }
            if let Some(shipping_address) = params.get("shipping_address") {
                obj.insert("shipping_address".to_string(), shipping_address.clone());
            }
            if let Some(billing_address) = params.get("billing_address") {
                obj.insert("billing_address".to_string(), billing_address.clone());
            }
        }

        let response = self.client
            .post(&url)
            .headers(self.headers())
            .json(&order_payload)
            .send()
            .await
            .map_err(ActionError::from_reqwest)?;

        let status = response.status();
        let body = response.text().await.unwrap_or_default();

        if !status.is_success() {
            return Err(self.handle_error(status, &body));
        }

        #[derive(Deserialize)]
        struct OrderResponse { order: serde_json::Value }

        let parsed: OrderResponse = serde_json::from_str(&body)
            .map_err(|e| ActionError::fatal(format!("Failed to parse Shopify response: {}", e)))?;

        let order_id = parsed.order.get("id")
            .and_then(|v| v.as_i64())
            .map(|id| id.to_string())
            .unwrap_or_default();

        let total_price = parsed.order.get("total_price")
            .and_then(|v| v.as_str())
            .map(|s| s.to_string());

        info!(order_id = %order_id, "Shopify order created");

        let mut result = ActionResult::success(
            serde_json::json!({
                "order": parsed.order,
                "order_id": order_id,
                "total_price": total_price,
            }),
            0,
        );
        result.provider_ref = Some(order_id);
        Ok(result)
    }

    async fn get_order(&self, params: &serde_json::Value) -> Result<ActionResult, ActionError> {
        let order_id = params.get("order_id").and_then(|v| v.as_str()).unwrap_or_default();
        let url = format!("{}/orders/{}.json", self.base_url(), order_id);

        let response = self.client
            .get(&url)
            .headers(self.headers())
            .send()
            .await
            .map_err(ActionError::from_reqwest)?;

        let status = response.status();
        let body = response.text().await.unwrap_or_default();

        if !status.is_success() {
            return Err(self.handle_error(status, &body));
        }

        #[derive(Deserialize)]
        struct OrderResponse { order: serde_json::Value }

        let parsed: OrderResponse = serde_json::from_str(&body)
            .map_err(|e| ActionError::fatal(format!("Failed to parse Shopify response: {}", e)))?;

        Ok(ActionResult::success(
            serde_json::json!({ "order": parsed.order }),
            0,
        ))
    }

    async fn update_order(&self, params: &serde_json::Value) -> Result<ActionResult, ActionError> {
        let order_id = params.get("order_id").and_then(|v| v.as_str()).unwrap_or_default();
        let url = format!("{}/orders/{}.json", self.base_url(), order_id);

        let mut order_update = serde_json::Map::new();

        if let Some(note) = params.get("note").and_then(|v| v.as_str()) {
            order_update.insert("note".to_string(), serde_json::Value::String(note.to_string()));
        }
        if let Some(tags) = params.get("tags").and_then(|v| v.as_str()) {
            order_update.insert("tags".to_string(), serde_json::Value::String(tags.to_string()));
        }
        if let Some(financial_status) = params.get("financial_status").and_then(|v| v.as_str()) {
            order_update.insert("financial_status".to_string(), serde_json::Value::String(financial_status.to_string()));
        }
        if let Some(fulfillment_status) = params.get("fulfillment_status").and_then(|v| v.as_str()) {
            order_update.insert("fulfillment_status".to_string(), serde_json::Value::String(fulfillment_status.to_string()));
        }

        let payload = serde_json::json!({ "order": order_update });

        let response = self.client
            .put(&url)
            .headers(self.headers())
            .json(&payload)
            .send()
            .await
            .map_err(ActionError::from_reqwest)?;

        let status = response.status();
        let body = response.text().await.unwrap_or_default();

        if !status.is_success() {
            return Err(self.handle_error(status, &body));
        }

        #[derive(Deserialize)]
        struct OrderResponse { order: serde_json::Value }

        let parsed: OrderResponse = serde_json::from_str(&body)
            .map_err(|e| ActionError::fatal(format!("Failed to parse Shopify response: {}", e)))?;

        Ok(ActionResult::success(
            serde_json::json!({ "order": parsed.order }),
            0,
        ))
    }

    async fn create_customer(&self, params: &serde_json::Value) -> Result<ActionResult, ActionError> {
        let url = format!("{}/customers.json", self.base_url());

        let mut customer = serde_json::json!({
            "customer": {
                "email": params.get("email").and_then(|v| v.as_str()).unwrap_or_default(),
            }
        });

        if let Some(obj) = customer["customer"].as_object_mut() {
            if let Some(first_name) = params.get("first_name").and_then(|v| v.as_str()) {
                obj.insert("first_name".to_string(), serde_json::Value::String(first_name.to_string()));
            }
            if let Some(last_name) = params.get("last_name").and_then(|v| v.as_str()) {
                obj.insert("last_name".to_string(), serde_json::Value::String(last_name.to_string()));
            }
            if let Some(phone) = params.get("phone").and_then(|v| v.as_str()) {
                obj.insert("phone".to_string(), serde_json::Value::String(phone.to_string()));
            }
            if let Some(tags) = params.get("tags").and_then(|v| v.as_str()) {
                obj.insert("tags".to_string(), serde_json::Value::String(tags.to_string()));
            }
            if let Some(note) = params.get("note").and_then(|v| v.as_str()) {
                obj.insert("note".to_string(), serde_json::Value::String(note.to_string()));
            }
            if let Some(verified) = params.get("verified_email").and_then(|v| v.as_bool()) {
                obj.insert("verified_email".to_string(), serde_json::Value::Bool(verified));
            }
        }

        let response = self.client
            .post(&url)
            .headers(self.headers())
            .json(&customer)
            .send()
            .await
            .map_err(ActionError::from_reqwest)?;

        let status = response.status();
        let body = response.text().await.unwrap_or_default();

        if !status.is_success() {
            return Err(self.handle_error(status, &body));
        }

        #[derive(Deserialize)]
        struct CustomerResponse { customer: serde_json::Value }

        let parsed: CustomerResponse = serde_json::from_str(&body)
            .map_err(|e| ActionError::fatal(format!("Failed to parse Shopify response: {}", e)))?;

        let customer_id = parsed.customer.get("id")
            .and_then(|v| v.as_i64())
            .map(|id| id.to_string())
            .unwrap_or_default();

        info!(customer_id = %customer_id, "Shopify customer created");

        let mut result = ActionResult::success(
            serde_json::json!({
                "customer": parsed.customer,
                "customer_id": customer_id,
            }),
            0,
        );
        result.provider_ref = Some(customer_id);
        Ok(result)
    }

    async fn get_customer(&self, params: &serde_json::Value) -> Result<ActionResult, ActionError> {
        let customer_id = params.get("customer_id").and_then(|v| v.as_str()).unwrap_or_default();
        let url = format!("{}/customers/{}.json", self.base_url(), customer_id);

        let response = self.client
            .get(&url)
            .headers(self.headers())
            .send()
            .await
            .map_err(ActionError::from_reqwest)?;

        let status = response.status();
        let body = response.text().await.unwrap_or_default();

        if !status.is_success() {
            return Err(self.handle_error(status, &body));
        }

        #[derive(Deserialize)]
        struct CustomerResponse { customer: serde_json::Value }

        let parsed: CustomerResponse = serde_json::from_str(&body)
            .map_err(|e| ActionError::fatal(format!("Failed to parse Shopify response: {}", e)))?;

        Ok(ActionResult::success(
            serde_json::json!({ "customer": parsed.customer }),
            0,
        ))
    }

    async fn update_customer(&self, params: &serde_json::Value) -> Result<ActionResult, ActionError> {
        let customer_id = params.get("customer_id").and_then(|v| v.as_str()).unwrap_or_default();
        let url = format!("{}/customers/{}.json", self.base_url(), customer_id);

        let mut customer_update = serde_json::Map::new();

        if let Some(email) = params.get("email").and_then(|v| v.as_str()) {
            customer_update.insert("email".to_string(), serde_json::Value::String(email.to_string()));
        }
        if let Some(first_name) = params.get("first_name").and_then(|v| v.as_str()) {
            customer_update.insert("first_name".to_string(), serde_json::Value::String(first_name.to_string()));
        }
        if let Some(last_name) = params.get("last_name").and_then(|v| v.as_str()) {
            customer_update.insert("last_name".to_string(), serde_json::Value::String(last_name.to_string()));
        }
        if let Some(phone) = params.get("phone").and_then(|v| v.as_str()) {
            customer_update.insert("phone".to_string(), serde_json::Value::String(phone.to_string()));
        }
        if let Some(tags) = params.get("tags").and_then(|v| v.as_str()) {
            customer_update.insert("tags".to_string(), serde_json::Value::String(tags.to_string()));
        }
        if let Some(note) = params.get("note").and_then(|v| v.as_str()) {
            customer_update.insert("note".to_string(), serde_json::Value::String(note.to_string()));
        }

        let payload = serde_json::json!({ "customer": customer_update });

        let response = self.client
            .put(&url)
            .headers(self.headers())
            .json(&payload)
            .send()
            .await
            .map_err(ActionError::from_reqwest)?;

        let status = response.status();
        let body = response.text().await.unwrap_or_default();

        if !status.is_success() {
            return Err(self.handle_error(status, &body));
        }

        #[derive(Deserialize)]
        struct CustomerResponse { customer: serde_json::Value }

        let parsed: CustomerResponse = serde_json::from_str(&body)
            .map_err(|e| ActionError::fatal(format!("Failed to parse Shopify response: {}", e)))?;

        Ok(ActionResult::success(
            serde_json::json!({ "customer": parsed.customer }),
            0,
        ))
    }

    async fn get_product(&self, params: &serde_json::Value) -> Result<ActionResult, ActionError> {
        let product_id = params.get("product_id").and_then(|v| v.as_str()).unwrap_or_default();
        let url = format!("{}/products/{}.json", self.base_url(), product_id);

        let response = self.client
            .get(&url)
            .headers(self.headers())
            .send()
            .await
            .map_err(ActionError::from_reqwest)?;

        let status = response.status();
        let body = response.text().await.unwrap_or_default();

        if !status.is_success() {
            return Err(self.handle_error(status, &body));
        }

        #[derive(Deserialize)]
        struct ProductResponse { product: serde_json::Value }

        let parsed: ProductResponse = serde_json::from_str(&body)
            .map_err(|e| ActionError::fatal(format!("Failed to parse Shopify response: {}", e)))?;

        Ok(ActionResult::success(
            serde_json::json!({ "product": parsed.product }),
            0,
        ))
    }

    async fn search_products(&self, params: &serde_json::Value) -> Result<ActionResult, ActionError> {
        let query = params.get("query").and_then(|v| v.as_str()).unwrap_or_default();
        let limit = params.get("limit").and_then(|v| v.as_i64()).unwrap_or(20);
        let fields = params.get("fields").and_then(|v| v.as_str());

        let mut url = format!("{}/products.json?limit={}", self.base_url(), limit);

        // Use published_status or handle search if available
        // For now, we fetch all and filter (not ideal for large catalogs)
        if let Some(f) = fields {
            url.push_str(&format!("&fields={}", f));
        }

        let response = self.client
            .get(&url)
            .headers(self.headers())
            .send()
            .await
            .map_err(ActionError::from_reqwest)?;

        let status = response.status();
        let body = response.text().await.unwrap_or_default();

        if !status.is_success() {
            return Err(self.handle_error(status, &body));
        }

        #[derive(Deserialize)]
        struct ProductsResponse {
            products: Vec<serde_json::Value>,
        }

        let parsed: ProductsResponse = serde_json::from_str(&body)
            .map_err(|e| ActionError::fatal(format!("Failed to parse Shopify response: {}", e)))?;

        // Filter by query in title or handle
        let filtered: Vec<_> = parsed.products
            .into_iter()
            .filter(|p| {
                let title = p.get("title").and_then(|v| v.as_str()).unwrap_or("").to_lowercase();
                let handle = p.get("handle").and_then(|v| v.as_str()).unwrap_or("").to_lowercase();
                let q = query.to_lowercase();
                title.contains(&q) || handle.contains(&q)
            })
            .collect();

        info!(query = %query, found = filtered.len(), "Shopify product search");

        Ok(ActionResult::success(
            serde_json::json!({
                "products": filtered,
                "query": query,
                "count": filtered.len(),
            }),
            0,
        ))
    }

    async fn update_inventory(&self, params: &serde_json::Value) -> Result<ActionResult, ActionError> {
        // Shopify inventory is updated via inventory levels
        let _inventory_item_id = params.get("inventory_item_id").and_then(|v| v.as_str())
            .or_else(|| params.get("product_id").and_then(|v| v.as_str()))
            .unwrap_or_default();

        let _location_id = params.get("location_id").and_then(|v| v.as_str())
            .unwrap_or("default"); // Will use primary location if not specified

        let available = params.get("available").and_then(|v| v.as_i64());
        let _adjust = params.get("adjust").and_then(|v| v.as_i64());

        // For now, we update the variant directly which is simpler
        if let Some(variant_id) = params.get("variant_id").and_then(|v| v.as_str()) {
            let url = format!("{}/variants/{}.json", self.base_url(), variant_id);

            let mut variant_update = serde_json::Map::new();

            if let Some(inv) = available {
                variant_update.insert("inventory_quantity".to_string(), serde_json::json!(inv));
            }
            if let Some(ref track_qty) = params.get("inventory_management").and_then(|v| v.as_str()) {
                variant_update.insert("inventory_management".to_string(), serde_json::Value::String(track_qty.to_string()));
            }

            let payload = serde_json::json!({ "variant": variant_update });

            let response = self.client
                .put(&url)
                .headers(self.headers())
                .json(&payload)
                .send()
                .await
                .map_err(ActionError::from_reqwest)?;

            let status = response.status();
            let body = response.text().await.unwrap_or_default();

            if !status.is_success() {
                return Err(self.handle_error(status, &body));
            }

            #[derive(Deserialize)]
            struct VariantResponse { variant: serde_json::Value }

            let parsed: VariantResponse = serde_json::from_str(&body)
                .map_err(|e| ActionError::fatal(format!("Failed to parse Shopify response: {}", e)))?;

            return Ok(ActionResult::success(
                serde_json::json!({
                    "variant": parsed.variant,
                    "variant_id": variant_id,
                    "updated": true,
                }),
                0,
            ));
        }

        Err(ActionError::fatal("update_inventory requires 'variant_id'".to_string()))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_shopify_connector_creation() {
        let connector = ShopifyConnector::new(
            "my-shop.myshopify.com".to_string(),
            "test_token".to_string(),
            "2024-01".to_string(),
        );

        assert_eq!(connector.name(), "shopify");
        assert!(connector.supports_action("create_order"));
        assert!(connector.supports_action("get_customer"));
        assert!(!connector.supports_action("unknown_action"));
    }

    #[test]
    fn test_validate_params_create_order() {
        let connector = ShopifyConnector::new(
            "my-shop.myshopify.com".to_string(),
            "test_token".to_string(),
            "2024-01".to_string(),
        );

        // Valid params
        let valid = serde_json::json!({
            "line_items": [{ "variant_id": "123", "quantity": 1 }]
        });
        assert!(connector.validate_params("create_order", &valid).is_ok());

        // Missing line_items
        let invalid = serde_json::json!({ "email": "test@example.com" });
        assert!(connector.validate_params("create_order", &invalid).is_err());

        // Empty line_items
        let empty = serde_json::json!({ "line_items": [] });
        assert!(connector.validate_params("create_order", &empty).is_err());
    }

    #[test]
    fn test_validate_params_search_products() {
        let connector = ShopifyConnector::new(
            "my-shop.myshopify.com".to_string(),
            "test_token".to_string(),
            "2024-01".to_string(),
        );

        let valid = serde_json::json!({ "query": "t-shirt" });
        assert!(connector.validate_params("search_products", &valid).is_ok());

        let invalid = serde_json::json!({ "limit": 10 });
        assert!(connector.validate_params("search_products", &invalid).is_err());
    }
}
