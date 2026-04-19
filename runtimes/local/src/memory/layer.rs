//! Unified memory layer — fan-out across hot, warm, and cold tiers.
//!
//! `MemoryLayer` is the single entry point graph nodes use for memory operations.
//! It implements a read-through / write-through cascade:
//!
//! ## Read path
//! ```
//!   Read(key)
//!     │
//!     ▼
//!   HOT.get() ──found──► return value
//!     │
//!     ▼ miss
//!   WARM.get() ──found──► cache in HOT ──► return value
//!     │
//!     ▼ miss
//!   COLD.get() ──found──► cache in HOT + WARM ──► return value
//!     │
//!     ▼ miss
//!   return None
//! ```
//!
//! ## Write path
//! ```
//!   Write(key, value)
//!     │
//!     ▼
//!   HOT.set()  (always write hot first)
//!   WARM.set() (write-through to warm)
//!   COLD.set() (async write to cold for archival / vector index)
//! ```

use std::sync::Arc;
use std::time::Duration;

use serde::{Deserialize, Serialize};
use tokio::sync::RwLock;
use tracing::{debug, instrument, warn};

use crate::engine::graph::{MemoryOp, NodeResult};
use crate::memory::cold::ColdMemory;
use crate::memory::hot::HotMemory;
use crate::memory::state::StateGraphMemory;
use crate::memory::warm::WarmMemory;

/// A unified memory handle — passed to graph nodes via `ExecutionContext`.
pub struct MemoryLayer {
    /// Hot (in-process LRU).
    hot: HotMemory,
    /// Warm (Redis).
    warm: Arc<WarmMemory>,
    /// Cold (PostgreSQL).
    cold: Arc<ColdMemory>,
    /// State graph memory (per-node metrics, decisions).
    state: Arc<StateGraphMemory>,
}

impl MemoryLayer {
    /// Create a new memory layer.
    pub fn new(hot_capacity: usize) -> Self {
        Self {
            hot: HotMemory::new(hot_capacity),
            warm: Arc::new(WarmMemory::new(None, 8)),
            cold: Arc::new(ColdMemory::new(None)),
            state: Arc::new(StateGraphMemory::new()),
        }
    }

    /// Create a memory layer with explicit tier connections.
    pub fn with_tiers(
        redis_url: Option<&str>,
        postgres_url: Option<&str>,
        hot_capacity: usize,
    ) -> Self {
        Self {
            hot: HotMemory::new(hot_capacity),
            warm: Arc::new(WarmMemory::new(redis_url, 8)),
            cold: Arc::new(ColdMemory::new(postgres_url)),
            state: Arc::new(StateGraphMemory::new()),
        }
    }

    /// Initialize Redis and PostgreSQL connections.
    ///
    /// Call this during startup. Failures are logged but do not prevent startup —
    /// the tier simply degrades gracefully (misses return None).
    pub async fn connect_tiers(&self) {
        if let Err(e) = self.warm.connect().await {
            warn!("WarmMemory (Redis) connection failed: {}", e);
        }
        if let Err(e) = self.cold.connect().await {
            warn!("ColdMemory (PostgreSQL) connection failed: {}", e);
        }
    }

    /// Get the state graph memory (for optimizer access).
    pub fn state(&self) -> Arc<StateGraphMemory> {
        Arc::clone(&self.state)
    }

    /// Read a value, cascading through tiers.
    #[instrument(skip(self), fields(tenant_id = ?tenant_id, key = %key))]
    pub async fn read(
        &self,
        tenant_id: Option<&str>,
        key: &str,
    ) -> anyhow::Result<Option<(String, &'static str)>> {
        // 1. Try hot first
        if let Some(v) = self.hot.get(tenant_id, key).await? {
            debug!(tier = "hot", "memory hit");
            return Ok(Some((v, "hot")));
        }

        // 2. Try warm (Redis)
        match self.warm.get(tenant_id, key).await {
            Ok(Some(v)) => {
                debug!(tier = "warm", "memory hit, caching in hot");
                // Write-through cache in hot
                self.hot.set(tenant_id, key, v.clone()).await?;
                return Ok(Some((v, "warm")));
            }
            Ok(None) => {}
            Err(e) => {
                debug!(tier = "warm", error = %e, "warm miss");
            }
        }

        // 3. Try cold (PostgreSQL)
        match self
            .cold
            .get_agent_memory(
                tenant_id.unwrap_or("default"),
                "default-agent",
                key,
            )
            .await
        {
            Ok(Some(v)) => {
                debug!(tier = "cold", "memory hit, caching in hot + warm");
                // Cache in hot and warm for future reads
                self.hot.set(tenant_id, key, v.clone()).await?;
                let _ = self.warm.set(tenant_id, key, v.clone(), None).await;
                return Ok(Some((v, "cold")));
            }
            Ok(None) => {}
            Err(e) => {
                debug!(tier = "cold", error = %e, "cold miss");
            }
        }

        Ok(None)
    }

    /// Write a value, cascading through tiers (write-through).
    #[instrument(skip(self, value), fields(tenant_id = ?tenant_id, key = %key))]
    pub async fn write(
        &self,
        tenant_id: Option<&str>,
        key: &str,
        value: String,
    ) -> anyhow::Result<()> {
        // Always write hot first (synchronous, < 1ms)
        self.hot.set(tenant_id, key, value.clone()).await?;

        // Write-through to warm (fire and forget, don't block on Redis)
        if let Err(e) = self.warm.set(tenant_id, key, value.clone(), None).await {
            debug!(tier = "warm", error = %e, "warm write-through failed, continuing");
        }

        // Async write to cold for archival / vector index
        let tenant = tenant_id.map(String::from).unwrap_or_else(|| "default".to_string());
        let cold = Arc::clone(&self.cold);
        let key_str = key.to_string();
        let value_clone = value.clone();
        tokio::spawn(async move {
            if let Err(e) = cold
                .set_agent_memory(&tenant, "default-agent", &key_str, &value_clone)
                .await
            {
                debug!(tier = "cold", error = %e, "cold write failed");
            }
        });

        Ok(())
    }

    /// Delete a value across all tiers.
    #[instrument(skip(self), fields(tenant_id = ?tenant_id, key = %key))]
    pub async fn delete(
        &self,
        tenant_id: Option<&str>,
        key: &str,
    ) -> anyhow::Result<bool> {
        let hot_deleted = self.hot.delete(tenant_id, key).await?;

        let warm_deleted = match self.warm.delete(tenant_id, key).await {
            Ok(b) => b,
            Err(e) => {
                debug!(tier = "warm", error = %e, "warm delete failed");
                false
            }
        };

        // Cold delete is async
        let cold = Arc::clone(&self.cold);
        let tenant = tenant_id.map(String::from).unwrap_or_else(|| "default".to_string());
        let key_str = key.to_string();
        tokio::spawn(async move {
            let _ = cold
                .set_agent_memory(&tenant, "default-agent", &key_str, "")
                .await;
        });

        Ok(hot_deleted || warm_deleted)
    }

    /// Execute a memory operation (Read / Write / Delete / List).
    ///
    /// This is what `NodeExecutor` calls when a graph node has `NodeType::Memory`.
    #[instrument(skip(self), fields(operation = ?operation, key = %key))]
    pub async fn execute_memory_op(
        &self,
        tenant_id: Option<&str>,
        operation: MemoryOp,
        key: &str,
        value: Option<String>,
    ) -> anyhow::Result<NodeResult> {
        let start = std::time::Instant::now();

        match operation {
            MemoryOp::Read => {
                match self.read(tenant_id, key).await {
                    Ok(Some((v, tier))) => {
                        let elapsed = start.elapsed().as_millis() as u64;
                        Ok(NodeResult::success(
                            crate::engine::graph::NodeId(uuid::Uuid::nil()),
                            serde_json::json!({
                                "key": key,
                                "value": v,
                                "tier": tier,
                            }),
                            elapsed,
                        ))
                    }
                    Ok(None) => {
                        let elapsed = start.elapsed().as_millis() as u64;
                        Ok(NodeResult::failure(
                            crate::engine::graph::NodeId(uuid::Uuid::nil()),
                            format!("key not found: {}", key),
                            elapsed,
                        ))
                    }
                    Err(e) => {
                        let elapsed = start.elapsed().as_millis() as u64;
                        Ok(NodeResult::failure(
                            crate::engine::graph::NodeId(uuid::Uuid::nil()),
                            format!("memory read error: {}", e),
                            elapsed,
                        ))
                    }
                }
            }
            MemoryOp::Write => {
                let value = value.ok_or_else(|| {
                    anyhow::anyhow!("MemoryOp::Write requires a value")
                })?;
                self.write(tenant_id, key, value.clone()).await?;
                let elapsed = start.elapsed().as_millis() as u64;
                Ok(NodeResult::success(
                    crate::engine::graph::NodeId(uuid::Uuid::nil()),
                    serde_json::json!({
                        "key": key,
                        "written": true,
                    }),
                    elapsed,
                ))
            }
            MemoryOp::Delete => {
                let deleted = self.delete(tenant_id, key).await?;
                let elapsed = start.elapsed().as_millis() as u64;
                Ok(NodeResult::success(
                    crate::engine::graph::NodeId(uuid::Uuid::nil()),
                    serde_json::json!({
                        "key": key,
                        "deleted": deleted,
                    }),
                    elapsed,
                ))
            }
            MemoryOp::List => {
                // List is a warm-tier scan (not implemented in MVP)
                let elapsed = start.elapsed().as_millis() as u64;
                Ok(NodeResult::success(
                    crate::engine::graph::NodeId(uuid::Uuid::nil()),
                    serde_json::json!({
                        "key": key,
                        "entries": [],
                        "note": "List not yet implemented — returns empty",
                    }),
                    elapsed,
                ))
            }
        }
    }

    /// Get statistics for all tiers.
    pub async fn stats(&self) -> MemoryStats {
        let hot_stats = self.hot.stats().await;
        let warm_connected = self.warm.is_connected().await;
        let cold_connected = self.cold.is_connected().await;

        MemoryStats {
            hot: hot_stats,
            warm: WarmTierStats {
                connected: warm_connected,
            },
            cold: ColdTierStats {
                connected: cold_connected,
            },
        }
    }
}

/// Statistics for all memory tiers.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryStats {
    pub hot: crate::memory::hot::HotMemoryStats,
    pub warm: WarmTierStats,
    pub cold: ColdTierStats,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WarmTierStats {
    pub connected: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ColdTierStats {
    pub connected: bool,
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_layer() -> MemoryLayer {
        MemoryLayer::new(100)
    }

    #[tokio::test]
    async fn test_write_and_read() {
        let layer = make_layer();
        layer
            .write(None, "test-key", "test-value".to_string())
            .await
            .unwrap();

        let result = layer.read(None, "test-key").await.unwrap();
        assert_eq!(result, Some(("test-value".to_string(), "hot")));
    }

    #[tokio::test]
    async fn test_delete() {
        let layer = make_layer();
        layer
            .write(None, "key-to-delete", "value".to_string())
            .await
            .unwrap();

        let deleted = layer.delete(None, "key-to-delete").await.unwrap();
        assert!(deleted);

        let v = layer.read(None, "key-to-delete").await.unwrap();
        assert_eq!(v, None);
    }

    #[tokio::test]
    async fn test_read_tier_cascade() {
        let layer = make_layer();

        // Write a value (goes to hot)
        layer
            .write(None, "tier-test-key", "tier-test-value".to_string())
            .await
            .unwrap();

        // Read should return hot tier
        let result = layer.read(None, "tier-test-key").await.unwrap();
        assert!(result.is_some());
        let (value, tier) = result.unwrap();
        assert_eq!(value, "tier-test-value");
        assert_eq!(tier, "hot"); // Value is in hot, so should report hot
    }

    #[tokio::test]
    async fn test_memory_op_read_hit() {
        let layer = make_layer();
        layer
            .write(None, "my-key", "my-value".to_string())
            .await
            .unwrap();

        let result = layer
            .execute_memory_op(None, MemoryOp::Read, "my-key", None)
            .await
            .unwrap();

        assert_eq!(result.status, crate::engine::graph::ExecutionStatus::Completed);
    }

    #[tokio::test]
    async fn test_memory_op_write() {
        let layer = make_layer();
        let result = layer
            .execute_memory_op(None, MemoryOp::Write, "new-key", Some("new-value".to_string()))
            .await
            .unwrap();

        assert_eq!(result.status, crate::engine::graph::ExecutionStatus::Completed);
    }

    #[tokio::test]
    async fn test_memory_op_delete() {
        let layer = make_layer();
        layer
            .write(None, "del-key", "del-value".to_string())
            .await
            .unwrap();

        let result = layer
            .execute_memory_op(None, MemoryOp::Delete, "del-key", None)
            .await
            .unwrap();

        assert_eq!(result.status, crate::engine::graph::ExecutionStatus::Completed);
    }

    #[tokio::test]
    async fn test_stats() {
        let layer = make_layer();
        let stats = layer.stats().await;
        assert_eq!(stats.warm.connected, false); // No Redis URL set
        assert_eq!(stats.cold.connected, false); // No Postgres URL set
        assert_eq!(stats.hot.total_entries, 0);
    }
}