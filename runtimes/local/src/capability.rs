//! Capability-based permission system for function execution.
//!
//! This module implements a deny-by-default capability model where functions
//! must explicitly declare their required capabilities in the manifest. The runtime
//! only exposes bindings for declared capabilities.

use std::collections::HashSet;

/// Represents a function's declared capabilities
#[derive(Debug, Clone, Default)]
pub struct Capabilities {
    /// Set of declared capabilities
    capabilities: HashSet<String>,
}

impl Capabilities {
    /// Create new capabilities from a comma-separated string
    pub fn from_string(cap_str: &str) -> Self {
        let capabilities: HashSet<String> = cap_str
            .split(',')
            .map(|s| s.trim().to_string())
            .filter(|s| !s.is_empty())
            .collect();

        Self { capabilities }
    }

    /// Create capabilities from a vector of strings
    pub fn from_vec(caps: Vec<String>) -> Self {
        let capabilities: HashSet<String> = caps
            .into_iter()
            .map(|s| s.trim().to_string())
            .filter(|s| !s.is_empty())
            .collect();

        Self { capabilities }
    }

    /// Check if a specific capability is granted
    pub fn has(&self, capability: &str) -> bool {
        self.capabilities.contains(capability)
    }

    /// Check if fetch:read capability is granted (HTTP GET)
    pub fn can_fetch_read(&self) -> bool {
        self.has("fetch:read")
    }

    /// Check if fetch:write capability is granted (HTTP POST/PUT/PATCH/DELETE)
    pub fn can_fetch_write(&self) -> bool {
        self.has("fetch:write")
    }

    /// Check if any fetch capability is granted
    pub fn can_fetch(&self) -> bool {
        self.can_fetch_read() || self.can_fetch_write()
    }

    /// Check if crypto capability is granted
    pub fn can_crypto(&self) -> bool {
        self.has("crypto")
    }

    /// Check if cache:read capability is granted
    pub fn can_cache_read(&self) -> bool {
        self.has("cache:read")
    }

    /// Check if cache:write capability is granted
    pub fn can_cache_write(&self) -> bool {
        self.has("cache:write")
    }

    /// Check if any cache capability is granted
    pub fn can_cache(&self) -> bool {
        self.can_cache_read() || self.can_cache_write()
    }

    /// Check if kv capability is granted
    pub fn can_kv(&self) -> bool {
        self.has("kv")
    }

    /// Check if webhook capability is granted
    pub fn can_webhook(&self) -> bool {
        self.has("webhook")
    }

    /// Check if email capability is granted
    pub fn can_email(&self) -> bool {
        self.has("email")
    }

    /// Check if storage capability is granted
    pub fn can_storage(&self) -> bool {
        self.has("storage")
    }

    /// Check if ai capability is granted
    pub fn can_ai(&self) -> bool {
        self.has("ai")
    }

    /// Check if external_api capability is granted
    pub fn can_external_api(&self) -> bool {
        self.has("external_api")
    }

    /// Get all granted capabilities
    pub fn all(&self) -> &HashSet<String> {
        &self.capabilities
    }

    /// Check if capabilities is empty (deny by default)
    pub fn is_empty(&self) -> bool {
        self.capabilities.is_empty()
    }
}

/// Allowed capabilities that can be declared in a function manifest
pub const ALLOWED_CAPABILITIES: &[&str] = &[
    "fetch:read",   // HTTP GET requests
    "fetch:write",  // HTTP POST/PUT/PATCH/DELETE
    "crypto",       // Cryptographic operations
    "cache:read",   // Read from cache
    "cache:write",  // Write to cache
    "kv",          // Key-value store
    "webhook",     // Webhook triggers
    "email",       // Email sending
    "storage",     // File storage
    "ai",          // AI/ML inference
    "external_api", // External API access
];

/// Validate that all capabilities are in the allowed list
pub fn validate_capabilities(capabilities: &Capabilities) -> Result<(), String> {
    for cap in capabilities.all() {
        if !ALLOWED_CAPABILITIES.contains(&cap.as_str()) {
            return Err(format!(
                "Invalid capability '{}'. Allowed: {:?}",
                cap, ALLOWED_CAPABILITIES
            ));
        }
    }
    Ok(())
}

/// Get a human-readable description of a capability
pub fn describe_capability(cap: &str) -> &'static str {
    match cap {
        "fetch:read" => "Make HTTP GET requests",
        "fetch:write" => "Make HTTP POST/PUT/PATCH/DELETE requests",
        "crypto" => "Perform cryptographic operations",
        "cache:read" => "Read from cache",
        "cache:write" => "Write to cache",
        "kv" => "Access key-value store",
        "webhook" => "Trigger webhooks",
        "email" => "Send emails",
        "storage" => "Access file storage",
        "ai" => "Use AI/ML inference",
        "external_api" => "Access external APIs",
        _ => "Unknown capability",
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_capabilities_from_string() {
        let caps = Capabilities::from_string("fetch:read,crypto,cache:write");
        assert!(caps.can_fetch_read());
        assert!(caps.can_crypto());
        assert!(caps.can_cache_write());
        assert!(!caps.can_kv());
    }

    #[test]
    fn test_empty_capabilities() {
        let caps = Capabilities::from_string("");
        assert!(caps.is_empty());
        assert!(!caps.can_fetch());
    }

    #[test]
    fn test_validate_capabilities() {
        let caps = Capabilities::from_string("fetch:read,crypto");
        assert!(validate_capabilities(&caps).is_ok());

        let invalid_caps = Capabilities::from_string("invalid_capability");
        assert!(validate_capabilities(&invalid_caps).is_err());
    }

    #[test]
    fn test_new_capabilities() {
        let caps = Capabilities::from_string("email,storage,ai,external_api");
        assert!(caps.can_email());
        assert!(caps.can_storage());
        assert!(caps.can_ai());
        assert!(caps.can_external_api());
        assert!(!caps.can_fetch());
    }

    #[test]
    fn test_capability_descriptions() {
        assert_eq!(describe_capability("email"), "Send emails");
        assert_eq!(describe_capability("storage"), "Access file storage");
        assert_eq!(describe_capability("ai"), "Use AI/ML inference");
        assert_eq!(describe_capability("external_api"), "Access external APIs");
        assert_eq!(describe_capability("unknown"), "Unknown capability");
    }

    #[test]
    fn test_allowed_capabilities_contains_new_ones() {
        assert!(ALLOWED_CAPABILITIES.contains(&"email"));
        assert!(ALLOWED_CAPABILITIES.contains(&"storage"));
        assert!(ALLOWED_CAPABILITIES.contains(&"ai"));
        assert!(ALLOWED_CAPABILITIES.contains(&"external_api"));
    }
}
