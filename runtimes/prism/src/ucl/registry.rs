//! Capability registry for managing discoverable capabilities

use std::collections::{HashMap, HashSet};
use std::sync::Arc;
use tokio::sync::RwLock;

use super::{Capability, CapabilityId, CapabilityCategory};

/// Configuration for the capability registry
#[derive(Debug, Clone)]
pub struct RegistryConfig {
    /// Whether to enable automatic capability discovery
    pub auto_discovery: bool,
    /// Discovery interval in seconds
    pub discovery_interval_secs: u64,
    /// Cache TTL in seconds
    pub cache_ttl_secs: u64,
    /// Maximum capabilities to cache
    pub max_cached_capabilities: usize,
}

impl Default for RegistryConfig {
    fn default() -> Self {
        Self {
            auto_discovery: true,
            discovery_interval_secs: 60,
            cache_ttl_secs: 300,
            max_cached_capabilities: 10_000,
        }
    }
}

/// The Universal Capability Registry - stores and manages all capabilities
pub struct CapabilityRegistry {
    config: RegistryConfig,
    /// Local capabilities (implemented by this node)
    local_capabilities: Arc<RwLock<HashMap<CapabilityId, Capability>>>,
    /// Remote capabilities (discovered from the mesh)
    remote_capabilities: Arc<RwLock<HashMap<CapabilityId, Capability>>>,
    /// Capabilities by category for fast lookup
    by_category: Arc<RwLock<HashMap<CapabilityCategory, HashSet<CapabilityId>>>>,
    /// Capabilities by tag for fast lookup
    by_tag: Arc<RwLock<HashMap<String, HashSet<CapabilityId>>>>,
}

impl CapabilityRegistry {
    pub fn new(config: RegistryConfig) -> Self {
        Self {
            config,
            local_capabilities: Arc::new(RwLock::new(HashMap::new())),
            remote_capabilities: Arc::new(RwLock::new(HashMap::new())),
            by_category: Arc::new(RwLock::new(HashMap::new())),
            by_tag: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    /// Register a local capability
    pub async fn register_local(&self, capability: Capability) -> PrismResult<CapabilityId> {
        let id = capability.capability_id.clone();
        self.register_capability(&self.local_capabilities, capability).await?;
        Ok(id)
    }

    /// Register a remote capability (discovered from mesh)
    pub async fn register_remote(&self, capability: Capability) -> PrismResult<CapabilityId> {
        let id = capability.capability_id.clone();
        self.register_capability(&self.remote_capabilities, capability).await?;
        Ok(id)
    }

    /// Internal method to register a capability in a store
    async fn register_capability(
        &self,
        store: &Arc<RwLock<HashMap<CapabilityId, Capability>>>,
        capability: Capability,
    ) -> PrismResult<()> {
        let mut capabilities = store.write().await;
        let id = capability.capability_id.clone();

        // Insert into main store
        capabilities.insert(id.clone(), capability.clone());

        // Index by category
        let mut by_category = self.by_category.write().await;
        by_category
            .entry(capability.category)
            .or_default()
            .insert(id.clone());

        // Index by tags
        let mut by_tag = self.by_tag.write().await;
        for tag in capability.metadata.tags.values() {
            by_tag
                .entry(tag.clone())
                .or_default()
                .insert(id.clone());
        }

        Ok(())
    }

    /// Get a capability by ID
    pub async fn get(&self, id: &CapabilityId) -> Option<Capability> {
        // Check local first
        {
            let local = self.local_capabilities.read().await;
            if let Some(cap) = local.get(id) {
                return Some(cap.clone());
            }
        }

        // Then remote
        let remote = self.remote_capabilities.read().await;
        remote.get(id).cloned()
    }

    /// Find capabilities by category
    pub async fn by_category(&self, category: CapabilityCategory) -> Vec<Capability> {
        let by_category = self.by_category.read().await;
        let capabilities = self.local_capabilities.read().await;

        by_category
            .get(&category)
            .map(|ids| {
                ids.iter()
                    .filter_map(|id| capabilities.get(id).cloned())
                    .collect()
            })
            .unwrap_or_default()
    }

    /// Find capabilities by tag
    pub async fn by_tag(&self, tag: &str) -> Vec<Capability> {
        let by_tag = self.by_tag.read().await;
        let capabilities = self.local_capabilities.read().await;

        by_tag
            .get(tag)
            .map(|ids| {
                ids.iter()
                    .filter_map(|id| capabilities.get(id).cloned())
                    .collect()
            })
            .unwrap_or_default()
    }

    /// Search capabilities by query string
    pub async fn search(&self, query: &str) -> Vec<Capability> {
        let query_lower = query.to_lowercase();
        let all_capabilities = self.list_all().await;

        all_capabilities
            .into_iter()
            .filter(|cap| {
                cap.name.to_lowercase().contains(&query_lower)
                    || cap.metadata.description.to_lowercase().contains(&query_lower)
                    || cap.metadata.tags.values().any(|t| t.to_lowercase().contains(&query_lower))
            })
            .collect()
    }

    /// List all available capabilities
    pub async fn list_all(&self) -> Vec<Capability> {
        let mut result = Vec::new();

        let local = self.local_capabilities.read().await;
        result.extend(local.values().cloned());

        let remote = self.remote_capabilities.read().await;
        result.extend(remote.values().cloned());

        result
    }

    /// List only local capabilities
    pub async fn list_local(&self) -> Vec<Capability> {
        let local = self.local_capabilities.read().await;
        local.values().cloned().collect()
    }

    /// Remove a capability
    pub async fn unregister(&self, id: &CapabilityId) -> bool {
        let mut local = self.local_capabilities.write().await;
        local.remove(id).is_some()
    }

    /// Get the registry configuration
    pub fn config(&self) -> &RegistryConfig {
        &self.config
    }

    /// Get the total count of capabilities
    pub async fn count(&self) -> usize {
        let local = self.local_capabilities.read().await.len();
        let remote = self.remote_capabilities.read().await.len();
        local + remote
    }

    /// Get capability count by category
    pub async fn count_by_category(&self, category: &CapabilityCategory) -> usize {
        let by_category = self.by_category.read().await;
        by_category.get(category).map(|s| s.len()).unwrap_or(0)
    }

    /// Check if auto-discovery is enabled
    pub fn is_auto_discovery_enabled(&self) -> bool {
        self.config.auto_discovery
    }

    /// Get discovery interval in seconds
    pub fn discovery_interval(&self) -> u64 {
        self.config.discovery_interval_secs
    }

    /// Check if a capability with the given ID is registered locally
    pub async fn is_registered_locally(&self, id: &CapabilityId) -> bool {
        let local = self.local_capabilities.read().await;
        local.contains_key(id)
    }
}

use crate::core::PrismResult;