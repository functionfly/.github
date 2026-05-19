use std::collections::HashMap;
use std::sync::Arc;
use parking_lot::RwLock;
use axum::{
    http::{Request, StatusCode},
    middleware::Next,
    response::Response,
    body::Body,
};

#[derive(Clone)]
pub struct ApiKeyAuth {
    keys: Arc<RwLock<HashMap<String, ApiKeyInfo>>>,
    header_name: &'static str,
}

#[derive(Clone, Debug)]
pub struct ApiKeyInfo {
    pub name: String,
    pub is_admin: bool,
}

impl ApiKeyAuth {
    pub fn new() -> Self {
        Self {
            keys: Arc::new(RwLock::new(HashMap::new())),
            header_name: "X-API-Key",
        }
    }

    pub fn with_api_key(key: impl Into<String>, name: impl Into<String>) -> Self {
        let auth = Self::new();
        auth.add_key(key.into(), name.into(), false);
        auth
    }

    pub fn from_env() -> Self {
        let auth = Self::new();

        if let Some(key) = std::env::var("SAR_API_KEY").ok() {
            auth.add_key(key, "env-api-key".to_string(), false);
        }
        if let Some(key) = std::env::var("SAR_ADMIN_API_KEY").ok() {
            auth.add_key(key, "env-admin-key".to_string(), true);
        }

        auth
    }

    pub fn add_key(&self, key: String, name: String, is_admin: bool) {
        let mut keys = self.keys.write();
        keys.insert(key, ApiKeyInfo { name, is_admin });
    }

    pub fn validate_key(&self, key: &str) -> Option<ApiKeyInfo> {
        self.keys.read().get(key).cloned()
    }

    pub fn key_count(&self) -> usize {
        self.keys.read().len()
    }
}

impl Default for ApiKeyAuth {
    fn default() -> Self {
        Self::new()
    }
}

pub async fn auth_check(
    auth: Arc<ApiKeyAuth>,
    require_auth: bool,
    request: Request<Body>,
    next: Next,
) -> Result<Response, StatusCode> {
    let path = request.uri().path().to_string();

    if path == "/health" || path == "/metrics" || path == "/api/health" {
        return Ok(next.run(request).await);
    }

    if !require_auth {
        return Ok(next.run(request).await);
    }

    let key = request
        .headers()
        .get("X-API-Key")
        .and_then(|v| v.to_str().ok())
        .map(|s| s.to_string());

    match key {
        Some(k) if auth.validate_key(&k).is_some() => Ok(next.run(request).await),
        _ => Err(StatusCode::UNAUTHORIZED),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_api_key_auth() {
        let auth = ApiKeyAuth::new();
        auth.add_key("test-key".to_string(), "test".to_string(), false);
        assert!(auth.validate_key("test-key").is_some());
        assert!(auth.validate_key("wrong-key").is_none());
    }

    #[test]
    fn test_admin_key() {
        let auth = ApiKeyAuth::new();
        auth.add_key("admin-key".to_string(), "admin".to_string(), true);
        let info = auth.validate_key("admin-key");
        assert!(info.is_some());
        assert!(info.unwrap().is_admin);
    }
}