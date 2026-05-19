//! Capability types for UCL

use std::collections::HashMap;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

/// Unique identifier for a capability
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct CapabilityId(pub String);

impl CapabilityId {
    pub fn new(id: impl Into<String>) -> Self {
        Self(id.into())
    }

    pub fn as_str(&self) -> &str {
        &self.0
    }
}

impl std::fmt::Display for CapabilityId {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

/// Category of capability
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum CapabilityCategory {
    Ai,        // ML inference, LLM, vision, audio
    Compute,   // General computation
    Storage,   // Database, file, cache
    Network,   // HTTP, websockets, NATS
    Crypto,    // Encryption, signing, hashing
    Sensors,   // IoT, robotics, hardware
    System,    // OS-level operations
}

impl Default for CapabilityCategory {
    fn default() -> Self {
        CapabilityCategory::Compute
    }
}

/// Metadata about a capability
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CapabilityMetadata {
    pub description: String,
    pub version: String,
    pub tags: HashMap<String, String>,
    pub provider: String,
    pub documentation_url: Option<String>,
    pub dependencies: Vec<String>,
    pub created_at: DateTime<Utc>,
}

impl CapabilityMetadata {
    pub fn new(_name: &str, provider: &str) -> Self {
        Self {
            description: String::new(),
            version: "1.0.0".to_string(),
            tags: HashMap::new(),
            provider: provider.to_string(),
            documentation_url: None,
            dependencies: Vec::new(),
            created_at: Utc::now(),
        }
    }
}

/// Trust level for a capability
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrustLevel {
    /// Trust score (0.0 to 1.0)
    pub score: f32,
    /// Remote attestation quote
    pub attestation: Option<String>,
    /// Whether the capability has been verified
    pub is_verified: bool,
    /// Last verification timestamp
    pub last_verified_at: Option<DateTime<Utc>>,
    /// Trust anchors (root certificates)
    pub trust_anchors: Vec<String>,
}

impl Default for TrustLevel {
    fn default() -> Self {
        Self {
            score: 0.9,
            attestation: None,
            is_verified: false,
            last_verified_at: None,
            trust_anchors: Vec::new(),
        }
    }
}

/// A discoverable capability in the mesh
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Capability {
    pub capability_id: CapabilityId,
    pub name: String,
    pub category: CapabilityCategory,
    pub metadata: CapabilityMetadata,
    pub performance: PerformanceProfile,
    pub trust: TrustLevel,
    pub runtimes: Vec<String>,
    pub languages: Vec<String>,
    /// WASM module bytes (if bundled)
    pub wasm_bundle: Option<Vec<u8>>,
    /// Whether this is a remote capability
    pub is_remote: bool,
    /// Endpoint URL if remote
    pub endpoint: Option<String>,
}

impl Capability {
    pub fn new(name: &str, category: CapabilityCategory, provider: &str) -> Self {
        Self {
            capability_id: CapabilityId::new(uuid::Uuid::new_v4().to_string()),
            name: name.to_string(),
            category,
            metadata: CapabilityMetadata::new(name, provider),
            performance: PerformanceProfile::default(),
            trust: TrustLevel::default(),
            runtimes: vec!["wasm".to_string()],
            languages: vec!["rust".to_string()],
            wasm_bundle: None,
            is_remote: false,
            endpoint: None,
        }
    }

    /// Create a remote capability that points to an endpoint
    pub fn remote(name: &str, category: CapabilityCategory, endpoint: &str) -> Self {
        Self {
            capability_id: CapabilityId::new(uuid::Uuid::new_v4().to_string()),
            name: name.to_string(),
            category,
            metadata: CapabilityMetadata::new(name, "remote"),
            performance: PerformanceProfile::default(),
            trust: TrustLevel::default(),
            runtimes: vec!["wasm".to_string()],
            languages: vec!["any".to_string()],
            wasm_bundle: None,
            is_remote: true,
            endpoint: Some(endpoint.to_string()),
        }
    }

    /// Check if this capability requires GPU
    pub fn requires_gpu(&self) -> bool {
        self.metadata.tags.get("gpu") == Some(&"true".to_string())
    }

    /// Check if this capability is free
    pub fn is_free(&self) -> bool {
        self.performance.cost.is_free
    }
}

/// Performance profile for a capability
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PerformanceProfile {
    pub avg_latency_ms: u32,
    pub p99_latency_ms: u32,
    pub throughput_rps: u32,
    pub trust_score: f32,
    pub cost: CostProfile,
}

impl Default for PerformanceProfile {
    fn default() -> Self {
        Self {
            avg_latency_ms: 100,
            p99_latency_ms: 500,
            throughput_rps: 1000,
            trust_score: 0.9,
            cost: CostProfile::default(),
        }
    }
}

/// Cost profile for a capability
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CostProfile {
    pub per_call_usd: f64,
    pub per_mb_usd: f64,
    pub is_free: bool,
}

impl Default for CostProfile {
    fn default() -> Self {
        Self {
            per_call_usd: 0.0,
            per_mb_usd: 0.0,
            is_free: true,
        }
    }
}