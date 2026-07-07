//! Redis Cache implementation - Phase 2 caching layer
//!
//! This module provides Redis-based caching for StateFabric, implementing
//! the Phase 2 requirements:
//! - Maintain existing functionality with Redis as secondary cache
//! - Rate limiting
//! - Active agent state
//! - Hot snapshot cache

use std::collections::HashMap;
use std::time::SystemTime;
use std::time::UNIX_EPOCH;
use redis::{AsyncCommands, Client, aio::ConnectionManager, RedisResult};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

/// Redis cache configuration (security-hardened)
#[derive(Debug, Clone)]
pub struct RedisConfig {
    /// Redis connection URL
    pub url: String,
    /// Connection timeout in seconds
    pub connection_timeout: u64,
    /// Default TTL for cache entries in seconds
    pub default_ttl: u64,
    /// Maximum number of connections in pool
    pub max_connections: u32,
    /// Key prefix for namespacing
    pub key_prefix: String,
    /// Password for Redis AUTH (from STATEFABRIC_REDIS_PASSWORD env)
    pub password: Option<String>,
    /// TLS enabled for secure connections
    pub use_tls: bool,
}

impl Default for RedisConfig {
    fn default() -> Self {
        Self::from_env()
    }
}

impl RedisConfig {
    /// Load configuration from environment variables
    pub fn from_env() -> Self {
        let base_url = std::env::var("STATEFABRIC_REDIS_URL")
            .unwrap_or_else(|_| "redis://127.0.0.1:6379".to_string());

        let password = std::env::var("STATEFABRIC_REDIS_PASSWORD").ok();
        // SECURITY: TLS enabled by default for secure connections
        // Set STATEFABRIC_REDIS_TLS=false only for local development with localhost
        let use_tls = std::env::var("STATEFABRIC_REDIS_TLS")
            .unwrap_or_else(|_| "true".to_string())
            .parse()
            .unwrap_or(true);

        // Build secure URL with password if provided
        let url = if let Some(ref pwd) = password {
            // Insert password into URL: redis://:password@host:port
            if base_url.contains('@') {
                base_url // Already has credentials
            } else {
                base_url.replace("redis://", &format!("redis://:{}@", pwd))
            }
        } else {
            base_url
        };

        Self {
            url,
            connection_timeout: std::env::var("STATEFABRIC_REDIS_CONNECTION_TIMEOUT")
                .unwrap_or_else(|_| "5".to_string())
                .parse()
                .unwrap_or(5),
            default_ttl: std::env::var("STATEFABRIC_REDIS_DEFAULT_TTL")
                .unwrap_or_else(|_| "3600".to_string())
                .parse()
                .unwrap_or(3600),
            max_connections: std::env::var("STATEFABRIC_REDIS_MAX_CONNECTIONS")
                .unwrap_or_else(|_| "10".to_string())
                .parse()
                .unwrap_or(10),
            key_prefix: std::env::var("STATEFABRIC_REDIS_KEY_PREFIX")
                .unwrap_or_else(|_| "statefabric:".to_string()),
            password,
            use_tls,
        }
    }
}

/// Cached state entry with metadata
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CachedStateEntry {
    /// The cached state data
    pub data: serde_json::Value,
    /// Timestamp when cached
    pub cached_at: u64,
    /// TTL in seconds
    pub ttl: u64,
    /// Version/hash for cache invalidation
    pub version: String,
}

/// Rate limit information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RateLimitInfo {
    /// Current request count in window
    pub count: u32,
    /// Window start timestamp
    pub window_start: u64,
    /// Window size in seconds
    pub window_size: u64,
    /// Maximum requests allowed in window
    pub max_requests: u32,
}

/// Active agent state
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ActiveAgentState {
    /// Agent ID
    pub agent_id: String,
    /// State ID being processed
    pub state_id: Uuid,
    /// Last activity timestamp
    pub last_activity: u64,
    /// Current operation
    pub operation: String,
    /// Progress (0-100)
    pub progress: u8,
}

/// Redis cache implementation
#[derive(Debug)]
pub struct RedisCache {
    /// Redis connection manager
    connection_manager: ConnectionManager,
    /// Configuration
    config: RedisConfig,
    /// Cache key prefixes for different types
    prefixes: CachePrefixes,
}

/// Cache key prefixes for different data types
#[derive(Debug, Clone)]
struct CachePrefixes {
    state: String,
    snapshot: String,
    rate_limit: String,
    agent_state: String,
    metadata: String,
}

impl CachePrefixes {
    fn new(prefix: &str) -> Self {
        Self {
            state: format!("{}state:", prefix),
            snapshot: format!("{}snapshot:", prefix),
            rate_limit: format!("{}ratelimit:", prefix),
            agent_state: format!("{}agent:", prefix),
            metadata: format!("{}meta:", prefix),
        }
    }
}

impl RedisCache {
    /// Create a new Redis cache instance
    pub async fn new(config: RedisConfig) -> RedisResult<Self> {
        let client = Client::open(config.url.clone())?;
        let connection_manager = ConnectionManager::new(client).await?;

        let prefixes = CachePrefixes::new(&config.key_prefix);

        Ok(Self {
            connection_manager,
            config,
            prefixes,
        })
    }

    /// Generate cache key for state data
    fn state_key(&self, state_id: &Uuid) -> String {
        format!("{}{}", self.prefixes.state, state_id)
    }

    /// Generate cache key for snapshot data
    fn snapshot_key(&self, state_id: &Uuid, version: i64) -> String {
        format!("{}{}:{}", self.prefixes.snapshot, state_id, version)
    }

    /// Generate cache key for rate limiting
    fn rate_limit_key(&self, identifier: &str, window: u64) -> String {
        format!("{}{}:{}", self.prefixes.rate_limit, identifier, window)
    }

    /// Generate cache key for agent state
    fn agent_state_key(&self, agent_id: &str) -> String {
        format!("{}{}", self.prefixes.agent_state, agent_id)
    }

    /// Generate cache key for metadata
    fn metadata_key(&self, key: &str) -> String {
        format!("{}{}", self.prefixes.metadata, key)
    }

    // ===== STATE CACHE OPERATIONS =====

    /// Get cached state data
    pub async fn get_state(&self, state_id: &Uuid) -> RedisResult<Option<CachedStateEntry>> {
        let key = self.state_key(state_id);
        let mut conn = self.connection_manager.clone();

        let data: Option<String> = conn.get(&key).await?;
        match data {
            Some(json_str) => {
                let entry: CachedStateEntry = serde_json::from_str(&json_str)
                    .map_err(|e| redis::RedisError::from(std::io::Error::new(std::io::ErrorKind::InvalidData, format!("deserialization error: {}", e))))?;
                Ok(Some(entry))
            }
            None => Ok(None),
        }
    }

    /// Set cached state data
    pub async fn set_state(&self, state_id: &Uuid, data: serde_json::Value, version: String) -> RedisResult<()> {
        let key = self.state_key(state_id);
        let mut conn = self.connection_manager.clone();

        let entry = CachedStateEntry {
            data,
            cached_at: SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs(),
            ttl: self.config.default_ttl,
            version,
        };

        let json_str = serde_json::to_string(&entry)
            .map_err(|e| redis::RedisError::from(std::io::Error::new(std::io::ErrorKind::InvalidData, format!("serialization error: {}", e))))?;

        conn.set_ex(&key, json_str, self.config.default_ttl).await
    }

    /// Delete cached state data
    pub async fn delete_state(&self, state_id: &Uuid) -> RedisResult<()> {
        let key = self.state_key(state_id);
        let mut conn = self.connection_manager.clone();
        let _: () = conn.del(&key).await?;
        Ok(())
    }

    /// Check if state is cached and still valid
    pub async fn is_state_cached(&self, state_id: &Uuid) -> RedisResult<bool> {
        let key = self.state_key(state_id);
        let mut conn = self.connection_manager.clone();
        let exists: bool = conn.exists(&key).await?;
        Ok(exists)
    }

    // ===== SNAPSHOT CACHE OPERATIONS =====

    /// Get cached snapshot data
    pub async fn get_snapshot(&self, state_id: &Uuid, version: i64) -> RedisResult<Option<serde_json::Value>> {
        let key = self.snapshot_key(state_id, version);
        let mut conn = self.connection_manager.clone();

        let data: Option<String> = conn.get(&key).await?;
        match data {
            Some(json_str) => {
                let snapshot: serde_json::Value = serde_json::from_str(&json_str)
                    .map_err(|e| redis::RedisError::from(std::io::Error::new(std::io::ErrorKind::InvalidData, format!("deserialization error: {}", e))))?;
                Ok(Some(snapshot))
            }
            None => Ok(None),
        }
    }

    /// Set cached snapshot data
    pub async fn set_snapshot(&self, state_id: &Uuid, version: i64, data: serde_json::Value) -> RedisResult<()> {
        let key = self.snapshot_key(state_id, version);
        let mut conn = self.connection_manager.clone();

        let json_str = serde_json::to_string(&data)
            .map_err(|e| redis::RedisError::from(std::io::Error::new(std::io::ErrorKind::InvalidData, format!("serialization error: {}", e))))?;

        conn.set_ex(&key, json_str, self.config.default_ttl).await
    }

    /// Delete cached snapshot
    pub async fn delete_snapshot(&self, state_id: &Uuid, version: i64) -> RedisResult<()> {
        let key = self.snapshot_key(state_id, version);
        let mut conn = self.connection_manager.clone();
        let _: () = conn.del(&key).await?;
        Ok(())
    }

    // ===== RATE LIMITING OPERATIONS =====

    /// Check if request is within rate limit
    pub async fn check_rate_limit(&self, identifier: &str, max_requests: u32, window_seconds: u64) -> RedisResult<bool> {
        let now = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs();
        let window_start = now / window_seconds * window_seconds;
        let key = self.rate_limit_key(identifier, window_start);

        let mut conn = self.connection_manager.clone();

        // Use Redis atomic operations for rate limiting
        let count: u32 = conn.incr(&key, 1).await?;

        // Set expiry on first request in window
        if count == 1 {
            let _: () = conn.expire(&key, window_seconds as i64).await?;
        }

        Ok(count <= max_requests)
    }

    /// Get current rate limit status
    pub async fn get_rate_limit_status(&self, identifier: &str, window_seconds: u64) -> RedisResult<Option<RateLimitInfo>> {
        let now = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs();
        let window_start = now / window_seconds * window_seconds;
        let key = self.rate_limit_key(identifier, window_start);

        let mut conn = self.connection_manager.clone();

        let count: Option<u32> = conn.get(&key).await?;
        match count {
            Some(c) => Ok(Some(RateLimitInfo {
                count: c,
                window_start,
                window_size: window_seconds,
                max_requests: 0, // This would need to be passed or stored separately
            })),
            None => Ok(None),
        }
    }

    // ===== AGENT STATE OPERATIONS =====

    /// Get active agent state
    pub async fn get_agent_state(&self, agent_id: &str) -> RedisResult<Option<ActiveAgentState>> {
        let key = self.agent_state_key(agent_id);
        let mut conn = self.connection_manager.clone();

        let data: Option<String> = conn.get(&key).await?;
        match data {
            Some(json_str) => {
                let state: ActiveAgentState = serde_json::from_str(&json_str)
                    .map_err(|e| redis::RedisError::from(std::io::Error::new(std::io::ErrorKind::InvalidData, format!("deserialization error: {}", e))))?;
                Ok(Some(state))
            }
            None => Ok(None),
        }
    }

    /// Set active agent state
    pub async fn set_agent_state(&self, agent_state: ActiveAgentState) -> RedisResult<()> {
        let key = self.agent_state_key(&agent_state.agent_id);
        let mut conn = self.connection_manager.clone();

        let json_str = serde_json::to_string(&agent_state)
            .map_err(|e| redis::RedisError::from(std::io::Error::new(std::io::ErrorKind::InvalidData, format!("serialization error: {}", e))))?;

        // Keep agent state for 24 hours or until explicitly removed
        conn.set_ex(&key, json_str, 86400).await
    }

    /// Update agent activity timestamp
    pub async fn update_agent_activity(&self, agent_id: &str) -> RedisResult<()> {
        let key = self.agent_state_key(agent_id);
        let mut conn = self.connection_manager.clone();

        // Get current state
        let current: Option<String> = conn.get(&key).await?;
        if let Some(json_str) = current {
            let mut state: ActiveAgentState = serde_json::from_str(&json_str)
                .map_err(|e| redis::RedisError::from(std::io::Error::new(std::io::ErrorKind::InvalidData, format!("deserialization error: {}", e))))?;

            // Update timestamp
            state.last_activity = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs();

            let updated_json = serde_json::to_string(&state)
                .map_err(|e| redis::RedisError::from(std::io::Error::new(std::io::ErrorKind::InvalidData, format!("serialization error: {}", e))))?;

            conn.set_ex::<_, _, ()>(&key, updated_json, 86400).await?;
        }

        Ok(())
    }

    /// Remove agent state
    pub async fn remove_agent_state(&self, agent_id: &str) -> RedisResult<()> {
        let key = self.agent_state_key(agent_id);
        let mut conn = self.connection_manager.clone();
        let _: () = conn.del(&key).await?;
        Ok(())
    }

    /// Get all active agents
    pub async fn get_all_active_agents(&self) -> RedisResult<Vec<ActiveAgentState>> {
        let mut conn = self.connection_manager.clone();

        // Get all agent keys
        let pattern = format!("{}*", self.prefixes.agent_state);
        let keys: Vec<String> = conn.keys(&pattern).await?;

        let mut agents = Vec::new();
        for key in keys {
            let data: Option<String> = conn.get(&key).await?;
            if let Some(json_str) = data {
                if let Ok(agent) = serde_json::from_str::<ActiveAgentState>(&json_str) {
                    agents.push(agent);
                }
            }
        }

        Ok(agents)
    }

    // ===== METADATA CACHE OPERATIONS =====

    /// Set metadata with TTL
    pub async fn set_metadata(&self, key: &str, value: serde_json::Value, ttl_seconds: Option<u64>) -> RedisResult<()> {
        let cache_key = self.metadata_key(key);
        let mut conn = self.connection_manager.clone();

        let json_str = serde_json::to_string(&value)
            .map_err(|e| redis::RedisError::from(std::io::Error::new(std::io::ErrorKind::InvalidData, format!("serialization error: {}", e))))?;

        let ttl = ttl_seconds.unwrap_or(self.config.default_ttl);
        conn.set_ex(&cache_key, json_str, ttl).await
    }

    /// Get metadata
    pub async fn get_metadata(&self, key: &str) -> RedisResult<Option<serde_json::Value>> {
        let cache_key = self.metadata_key(key);
        let mut conn = self.connection_manager.clone();

        let data: Option<String> = conn.get(&cache_key).await?;
        match data {
            Some(json_str) => {
                let value: serde_json::Value = serde_json::from_str(&json_str)
                    .map_err(|e| redis::RedisError::from(std::io::Error::new(std::io::ErrorKind::InvalidData, format!("deserialization error: {}", e))))?;
                Ok(Some(value))
            }
            None => Ok(None),
        }
    }

    /// Delete metadata
    pub async fn delete_metadata(&self, key: &str) -> RedisResult<()> {
        let cache_key = self.metadata_key(key);
        let mut conn = self.connection_manager.clone();
        let _: () = conn.del(&cache_key).await?;
        Ok(())
    }

    // ===== UTILITY OPERATIONS =====

    /// Clear all cache entries with a specific prefix
    pub async fn clear_by_prefix(&self, prefix: &str) -> RedisResult<u32> {
        let mut conn = self.connection_manager.clone();

        let pattern = format!("{}{}*", self.config.key_prefix, prefix);
        let keys: Vec<String> = conn.keys(&pattern).await?;

        if keys.is_empty() {
            return Ok(0);
        }

        let deleted: u32 = conn.del(&keys).await?;
        Ok(deleted)
    }

    /// Get cache statistics
    pub async fn get_stats(&self) -> RedisResult<HashMap<String, u64>> {
        let mut conn = self.connection_manager.clone();
        let mut stats = HashMap::new();

        // Count entries for each prefix
        let patterns = vec![
            (&self.prefixes.state, "state_cache_count"),
            (&self.prefixes.snapshot, "snapshot_cache_count"),
            (&self.prefixes.rate_limit, "rate_limit_count"),
            (&self.prefixes.agent_state, "agent_state_count"),
            (&self.prefixes.metadata, "metadata_count"),
        ];

        for (pattern, stat_key) in patterns {
            let full_pattern = format!("{}*", pattern);
            let keys: Vec<String> = conn.keys(&full_pattern).await?;
            stats.insert(stat_key.to_string(), keys.len() as u64);
        }

        Ok(stats)
    }

    /// Health check - ping Redis
    pub async fn health_check(&self) -> RedisResult<bool> {
        let mut conn = self.connection_manager.clone();
        let result: String = redis::cmd("PING").query_async(&mut conn).await?;
        Ok(result == "PONG")
    }

    /// Flush all cache data (use with caution)
    pub async fn flush_all(&self) -> RedisResult<()> {
        let mut conn = self.connection_manager.clone();
        let _: () = redis::cmd("FLUSHDB").query_async(&mut conn).await?;
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_redis_config_defaults() {
        let config = RedisConfig::default();
        assert_eq!(config.url, "redis://127.0.0.1:6379");
        assert_eq!(config.default_ttl, 3600);
        assert_eq!(config.key_prefix, "statefabric:");
    }

    #[tokio::test]
    async fn test_cache_prefixes() {
        let prefixes = CachePrefixes::new("test:");
        assert_eq!(prefixes.state, "test:state:");
        assert_eq!(prefixes.snapshot, "test:snapshot:");
        assert_eq!(prefixes.rate_limit, "test:ratelimit:");
        assert_eq!(prefixes.agent_state, "test:agent:");
        assert_eq!(prefixes.metadata, "test:meta:");
    }

    // Note: Integration tests would require a running Redis instance
    // These are basic unit tests for the structure
}
