//! Warm memory tier — Redis.
//!
//! Reuses the Go backend Redis key schema from `internal/cache/redis.go`:
//!
//! - Execution state: `exec:{tenant}:{exec_id}`
//! - Agent memory: `mem:{tenant}:{agent_id}:{key}`
//! - Vector search: `vec:{tenant}:{index}:{id}`
//!
//! All keys are namespaced per tenant for isolation.

use std::sync::Arc;
use std::time::Duration;

use serde::{Deserialize, Serialize};
use tokio::sync::RwLock;

/// A Redis connection pool wrapper for the warm memory tier.
pub struct WarmMemory {
    /// Redis URL (e.g. "redis://localhost:6379").
    redis_url: String,
    /// Optional Redis client — `None` means Redis is not configured
    /// and all operations will return `Err`.
    client: Arc<RwLock<Option<redis::Client>>>,
    /// Default TTL for keys written to Redis.
    default_ttl: Duration,
}

impl WarmMemory {
    /// Create a new warm memory tier.
    ///
    /// Pass `redis_url = None` to create a tier that is permanently unavailable
    /// (graceful degradation — hot misses fall through to warm which returns
    /// None rather than crashing).
    pub fn new(redis_url: Option<&str>, _pool_max_size: usize) -> Self {
        Self {
            redis_url: redis_url
                .map(|s| s.to_string())
                .unwrap_or_else(|| {
                    eprintln!("Warning: REDIS_URL not set for warm memory tier");
                    String::new() // Empty means unavailable
                }),
            client: Arc::new(RwLock::new(None)),
            default_ttl: Duration::from_secs(3600), // 1 hour default
        }
    }

    /// Initialize the connection pool.
    ///
    /// Call this once during startup after the Redis URL is confirmed reachable.
    pub async fn connect(&self) -> anyhow::Result<()> {
        let client = redis::Client::open(self.redis_url.as_str())
            .map_err(|e| anyhow::anyhow!("Failed to open Redis client: {}", e))?;

        // Ping to verify connectivity
        let mut conn = client
            .get_multiplexed_async_connection()
            .await
            .map_err(|e| anyhow::anyhow!("Failed to acquire Redis connection: {}", e))?;

        redis::cmd("PING")
            .query_async::<String>(&mut conn)
            .await
            .map_err(|e| anyhow::anyhow!("Redis PING failed: {}", e))?;

        let mut guard = self.client.write().await;
        *guard = Some(client);

        tracing::info!("WarmMemory (Redis) connected to {}", self.redis_url);
        Ok(())
    }

    /// Returns true if the Redis client has been initialized.
    pub async fn is_connected(&self) -> bool {
        self.client.read().await.is_some()
    }

    /// Namespaced execution state key.
    fn exec_key(tenant: Option<&str>, exec_id: &str) -> String {
        match tenant {
            Some(t) => format!("exec:{}:{}", t, exec_id),
            None => format!("exec:{}", exec_id),
        }
    }

    /// Namespaced agent memory key.
    fn mem_key(tenant: Option<&str>, agent_id: &str, key: &str) -> String {
        match tenant {
            Some(t) => format!("mem:{}:{}:{}", t, agent_id, key),
            None => format!("mem:{}:{}", agent_id, key),
        }
    }

    /// Generic namespaced key.
    fn key(tenant: Option<&str>, key: &str) -> String {
        match tenant {
            Some(t) => format!("mem:{}:{}", t, key),
            None => format!("mem:{}", key),
        }
    }

    /// Get a value from Redis.
    pub async fn get(&self, tenant_id: Option<&str>, key: &str) -> anyhow::Result<Option<String>> {
        let client_guard = self.client.read().await;
        let client = client_guard.as_ref().ok_or_else(|| {
            anyhow::anyhow!("Redis not connected — set REDIS_URL to enable warm tier")
        })?;

        let mut conn = client
            .get_multiplexed_async_connection()
            .await
            .map_err(|e| anyhow::anyhow!("Failed to get Redis connection: {}", e))?;

        let ns_key = Self::key(tenant_id, key);
        let result: Option<String> = redis::cmd("GET")
            .arg(&ns_key)
            .query_async(&mut conn)
            .await
            .map_err(|e| anyhow::anyhow!("Redis GET failed: {}", e))?;

        Ok(result)
    }

    /// Set a value in Redis with an optional TTL.
    pub async fn set(
        &self,
        tenant_id: Option<&str>,
        key: &str,
        value: String,
        ttl_secs: Option<u64>,
    ) -> anyhow::Result<()> {
        let client_guard = self.client.read().await;
        let client = client_guard.as_ref().ok_or_else(|| {
            anyhow::anyhow!("Redis not connected")
        })?;

        let mut conn = client
            .get_multiplexed_async_connection()
            .await
            .map_err(|e| anyhow::anyhow!("Failed to get Redis connection: {}", e))?;

        let ns_key = Self::key(tenant_id, key);
        let ttl = ttl_secs.unwrap_or(self.default_ttl.as_secs() as u64);

        let _: () = redis::cmd("SET")
            .arg(&ns_key)
            .arg(&value)
            .arg("EX")
            .arg(ttl)
            .query_async(&mut conn)
            .await
            .map_err(|e| anyhow::anyhow!("Redis SET failed: {}", e))?;

        Ok(())
    }

    /// Delete a key from Redis.
    pub async fn delete(&self, tenant_id: Option<&str>, key: &str) -> anyhow::Result<bool> {
        let client_guard = self.client.read().await;
        let client = client_guard.as_ref().ok_or_else(|| anyhow::anyhow!("Redis not connected"))?;

        let mut conn = client
            .get_multiplexed_async_connection()
            .await
            .map_err(|e| anyhow::anyhow!("Failed to get Redis connection: {}", e))?;

        let ns_key = Self::key(tenant_id, key);
        let deleted: i64 = redis::cmd("DEL")
            .arg(&ns_key)
            .query_async(&mut conn)
            .await
            .map_err(|e| anyhow::anyhow!("Redis DEL failed: {}", e))?;

        Ok(deleted > 0)
    }

    /// Store execution state in Redis.
    pub async fn set_execution_state(
        &self,
        tenant_id: Option<&str>,
        exec_id: &str,
        state: &str,
        ttl_secs: Option<u64>,
    ) -> anyhow::Result<()> {
        let client_guard = self.client.read().await;
        let client = client_guard.as_ref().ok_or_else(|| anyhow::anyhow!("Redis not connected"))?;

        let mut conn = client
            .get_multiplexed_async_connection()
            .await
            .map_err(|e| anyhow::anyhow!("Failed to get Redis connection: {}", e))?;

        let ns_key = Self::exec_key(tenant_id, exec_id);
        let ttl = ttl_secs.unwrap_or(self.default_ttl.as_secs() as u64);

        let _: () = redis::cmd("SET")
            .arg(&ns_key)
            .arg(state)
            .arg("EX")
            .arg(ttl)
            .query_async(&mut conn)
            .await
            .map_err(|e| anyhow::anyhow!("Redis SET failed: {}", e))?;

        Ok(())
    }

    /// Get execution state from Redis.
    pub async fn get_execution_state(
        &self,
        tenant_id: Option<&str>,
        exec_id: &str,
    ) -> anyhow::Result<Option<String>> {
        let client_guard = self.client.read().await;
        let client = client_guard.as_ref().ok_or_else(|| anyhow::anyhow!("Redis not connected"))?;

        let mut conn = client
            .get_multiplexed_async_connection()
            .await
            .map_err(|e| anyhow::anyhow!("Failed to get Redis connection: {}", e))?;

        let ns_key = Self::exec_key(tenant_id, exec_id);
        let result: Option<String> = redis::cmd("GET")
            .arg(&ns_key)
            .query_async(&mut conn)
            .await
            .map_err(|e| anyhow::anyhow!("Redis GET failed: {}", e))?;

        Ok(result)
    }

    /// Increment a counter (useful for rate limiting, call counting).
    pub async fn incr(&self, tenant_id: Option<&str>, key: &str, ttl_secs: Option<u64>) -> anyhow::Result<u64> {
        let client_guard = self.client.read().await;
        let client = client_guard.as_ref().ok_or_else(|| anyhow::anyhow!("Redis not connected"))?;

        let mut conn = client
            .get_multiplexed_async_connection()
            .await
            .map_err(|e| anyhow::anyhow!("Failed to get Redis connection: {}", e))?;

        let ns_key = Self::key(tenant_id, key);
        let result: u64 = redis::cmd("INCR")
            .arg(&ns_key)
            .query_async(&mut conn)
            .await
            .map_err(|e| anyhow::anyhow!("Redis INCR failed: {}", e))?;

        // Set TTL if requested
        if let Some(ttl) = ttl_secs {
            let _: () = redis::cmd("EXPIRE")
                .arg(&ns_key)
                .arg(ttl as i64)
                .query_async(&mut conn)
                .await
                .map_err(|e| anyhow::anyhow!("Redis EXPIRE failed: {}", e))?;
        }

        Ok(result)
    }
}

/// Statistics for the warm memory tier.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WarmMemoryStats {
    pub connected: bool,
    pub redis_url: String,
    pub default_ttl_secs: u64,
}

#[cfg(test)]
mod tests {
    // Redis integration tests would require a running Redis instance.
    // These are placeholder tests that verify the struct compiles.
    use super::*;

    #[test]
    fn test_key_schema() {
        // Verify key schemas match Go backend patterns
        assert_eq!(
            WarmMemory::key(Some("tenant-a"), "mykey"),
            "mem:tenant-a:mykey"
        );
        assert_eq!(
            WarmMemory::exec_key(Some("tenant-a"), "exec-123"),
            "exec:tenant-a:exec-123"
        );
        assert_eq!(
            WarmMemory::mem_key(Some("tenant-a"), "agent-x", "memory"),
            "mem:tenant-a:agent-x:memory"
        );
        assert_eq!(
            WarmMemory::key(None, "naked"),
            "mem:naked"
        );
    }

    #[tokio::test]
    async fn test_disconnected_returns_error() {
        let warm = WarmMemory::new(None, 4);
        assert!(!warm.is_connected().await);

        let result = warm.get(None, "key").await;
        assert!(result.is_err());
    }
}
