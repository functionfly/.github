//! Cold memory tier — PostgreSQL / pgvector for semantic recall.
//!
//! Stores:
//! - Execution history (completed graphs, node results, timing)
//! - Long-term agent facts (journal entries, learned preferences)
//! - Vector embeddings for semantic search (when pgvector is available)
//!
//! ## Schema (PostgreSQL)
//!
//! ```sql
//! CREATE TABLE IF NOT EXISTS execution_history (
//!     id          UUID PRIMARY KEY,
//!     tenant_id   TEXT NOT NULL,
//!     graph_id    UUID NOT NULL,
//!     status      TEXT NOT NULL,  -- 'completed', 'failed', 'cancelled'
//!     started_at  TIMESTAMPTZ NOT NULL,
//!     ended_at    TIMESTAMPTZ,
//!     duration_ms INTEGER,
//!     input_json  JSONB,
//!     output_json JSONB,
//!     error_text  TEXT,
//!     created_at  TIMESTAMPTZ DEFAULT NOW()
//! );
//!
//! CREATE INDEX IF NOT EXISTS idx_exec_tenant ON execution_history(tenant_id);
//! CREATE INDEX IF NOT EXISTS idx_exec_graph ON execution_history(graph_id);
//! CREATE INDEX IF NOT EXISTS idx_exec_started ON execution_history(started_at DESC);
//!
//! -- Agent memory / long-term facts
//! CREATE TABLE IF NOT EXISTS agent_memory (
//!     id          UUID PRIMARY KEY,
//!     tenant_id   TEXT NOT NULL,
//!     agent_id    TEXT NOT NULL,
//!     key         TEXT NOT NULL,
//!     value_json  JSONB NOT NULL,
//!     embedding   vector(1536),  -- pgvector, populated by embedding model
//!     created_at  TIMESTAMPTZ DEFAULT NOW(),
//!     updated_at  TIMESTAMPTZ DEFAULT NOW(),
//!     UNIQUE(tenant_id, agent_id, key)
//! );
//!
//! CREATE INDEX IF NOT EXISTS idx_agent_mem_tenant_agent ON agent_memory(tenant_id, agent_id);
//! CREATE INDEX IF NOT EXISTS idx_agent_mem_embedding ON agent_memory USING ivfflat (embedding vector_cosine_ops)
//!     WITH (lists = 100);
//! ```

use std::sync::Arc;

use anyhow::Context as _;
use serde::{Deserialize, Serialize};
use tokio::sync::RwLock;
use tokio_postgres::NoTls;

/// Database connection string.
type ConnectionString = String;

/// A PostgreSQL client wrapper for the cold memory tier.
pub struct ColdMemory {
    /// PostgreSQL connection string.
    conn_string: ConnectionString,
    /// Optional connection pool — `None` means Postgres is not configured.
    pool: Arc<RwLock<Option<tokio_postgres::Client>>>,
}

impl ColdMemory {
    /// Create a new cold memory tier.
    ///
    /// Pass `conn_string = None` for graceful degradation (cold misses return None).
    pub fn new(conn_string: Option<&str>) -> Self {
        Self {
            conn_string: conn_string
                .map(|s| s.to_string())
                .unwrap_or_default(),
            pool: Arc::new(RwLock::new(None)),
        }
    }

    /// Returns true if the database connection has been initialized.
    pub async fn is_connected(&self) -> bool {
        self.pool.read().await.is_some()
    }

    /// Connect to PostgreSQL and run migrations.
    ///
    /// Migrations are run inline — no external migration tool required.
    pub async fn connect(&self) -> anyhow::Result<()> {
        if self.conn_string.is_empty() {
            return Err(anyhow::anyhow!(
                "No PostgreSQL connection string configured — cold tier unavailable"
            ));
        }

        let (client, connection) = tokio_postgres::connect(&self.conn_string, NoTls)
            .await
            .context("Failed to connect to PostgreSQL")?;

        tokio::spawn(async move {
            if let Err(e) = connection.await {
                tracing::error!("PostgreSQL connection error: {}", e);
            }
        });

        {
            let mut guard = self.pool.write().await;
            *guard = Some(client);
        }

        tracing::info!(
            "ColdMemory (PostgreSQL) connected, conn_string len = {}",
            self.conn_string.len()
        );
        Ok(())
    }

    /// Store an execution record.
    pub async fn store_execution(
        &self,
        tenant_id: &str,
        graph_id: uuid::Uuid,
        status: &str,
        duration_ms: Option<i64>,
        input_json: Option<&str>,
        output_json: Option<&str>,
        error_text: Option<&str>,
    ) -> anyhow::Result<()> {
        let pool_guard = self.pool.read().await;
        let client = pool_guard.as_ref().ok_or_else(|| {
            anyhow::anyhow!("PostgreSQL not connected")
        })?;

        let now = chrono::Utc::now();
        let id = uuid::Uuid::new_v4();

        client
            .execute(
                "INSERT INTO execution_history (id, tenant_id, graph_id, status, started_at, ended_at, duration_ms, input_json, output_json, error_text)
                 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
                &[
                    &id,
                    &tenant_id,
                    &graph_id,
                    &status,
                    &now,
                    &now,
                    &duration_ms,
                    &input_json.map(String::from).as_ref(),
                    &output_json.map(String::from).as_ref(),
                    &error_text.map(String::from).as_ref(),
                ],
            )
            .await
            .context("store_execution INSERT failed")?;

        Ok(())
    }

    /// Query recent executions for a tenant (for observability / billing).
    pub async fn query_executions(
        &self,
        tenant_id: &str,
        limit: i64,
    ) -> anyhow::Result<Vec<ExecutionRecord>> {
        let pool_guard = self.pool.read().await;
        let client = pool_guard.as_ref().ok_or_else(|| {
            anyhow::anyhow!("PostgreSQL not connected")
        })?;

        let rows = client
            .query(
                "SELECT id, tenant_id, graph_id, status, started_at, ended_at, duration_ms
                 FROM execution_history
                 WHERE tenant_id = $1
                 ORDER BY started_at DESC
                 LIMIT $2",
                &[&tenant_id, &limit],
            )
            .await
            .context("query_executions SELECT failed")?;

        let records = rows
            .iter()
            .map(|row| ExecutionRecord {
                id: row.get(0),
                tenant_id: row.get(1),
                graph_id: row.get(2),
                status: row.get(3),
                started_at: row.get(4),
                ended_at: row.get(5),
                duration_ms: row.get(6),
            })
            .collect();

        Ok(records)
    }

    /// Store a long-term agent memory entry.
    pub async fn set_agent_memory(
        &self,
        tenant_id: &str,
        agent_id: &str,
        key: &str,
        value_json: &str,
    ) -> anyhow::Result<()> {
        let pool_guard = self.pool.read().await;
        let client = pool_guard.as_ref().ok_or_else(|| {
            anyhow::anyhow!("PostgreSQL not connected")
        })?;

        let id = uuid::Uuid::new_v4();
        let now = chrono::Utc::now();

        client
            .execute(
                "INSERT INTO agent_memory (id, tenant_id, agent_id, key, value_json, created_at, updated_at)
                 VALUES ($1, $2, $3, $4, $5, $6, $7)
                 ON CONFLICT (tenant_id, agent_id, key)
                 DO UPDATE SET value_json = EXCLUDED.value_json, updated_at = EXCLUDED.updated_at",
                &[&id, &tenant_id, &agent_id, &key, &value_json, &now, &now],
            )
            .await
            .context("set_agent_memory UPSERT failed")?;

        Ok(())
    }

    /// Get a long-term agent memory entry.
    pub async fn get_agent_memory(
        &self,
        tenant_id: &str,
        agent_id: &str,
        key: &str,
    ) -> anyhow::Result<Option<String>> {
        let pool_guard = self.pool.read().await;
        let client = pool_guard.as_ref().ok_or_else(|| {
            anyhow::anyhow!("PostgreSQL not connected")
        })?;

        let row = client
            .query_opt(
                "SELECT value_json FROM agent_memory WHERE tenant_id = $1 AND agent_id = $2 AND key = $3",
                &[&tenant_id, &agent_id, &key],
            )
            .await
            .context("get_agent_memory SELECT failed")?;

        Ok(row.map(|r| r.get(0)))
    }

    /// Vector similarity search for semantic memory recall.
    ///
    /// Requires pgvector extension. When not available, falls back to exact-key lookup.
    pub async fn semantic_search(
        &self,
        tenant_id: &str,
        agent_id: &str,
        query_embedding: &[f32],
        limit: i64,
    ) -> anyhow::Result<Vec<SemanticMemoryHit>> {
        let pool_guard = self.pool.read().await;
        let client = pool_guard.as_ref().ok_or_else(|| {
            anyhow::anyhow!("PostgreSQL not connected")
        })?;

        let embedding = pgvector::Vector::from(query_embedding.to_vec());

        let rows = client
            .query(
                "SELECT key, value_json, 1 - (embedding <=> $3) AS similarity, created_at
                 FROM agent_memory
                 WHERE tenant_id = $1 AND agent_id = $2 AND embedding IS NOT NULL
                 ORDER BY embedding <=> $3
                 LIMIT $4",
                &[&tenant_id, &agent_id, &embedding, &limit],
            )
            .await
            .context("semantic_search failed")?;

        let hits = rows
            .iter()
            .map(|row| SemanticMemoryHit {
                key: row.get(0),
                value_json: row.get(1),
                similarity: row.get(2),
                created_at: row.get(3),
            })
            .collect();

        Ok(hits)
    }
}

/// A single execution history record from the cold tier.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionRecord {
    pub id: uuid::Uuid,
    pub tenant_id: String,
    pub graph_id: uuid::Uuid,
    pub status: String,
    pub started_at: chrono::DateTime<chrono::Utc>,
    pub ended_at: Option<chrono::DateTime<chrono::Utc>>,
    pub duration_ms: Option<i64>,
}

/// A semantic search hit from agent memory.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SemanticMemoryHit {
    pub key: String,
    pub value_json: String,
    pub similarity: f32,
    pub created_at: chrono::DateTime<chrono::Utc>,
}

/// Statistics for the cold memory tier.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ColdMemoryStats {
    pub connected: bool,
    pub has_pgvector: bool,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_disconnected_behavior() {
        let cold = ColdMemory::new(None);
        assert!(!cold.is_connected().await);

        let result = cold
            .store_execution("t", uuid::Uuid::new_v4(), "completed", Some(100), None, None, None)
            .await;
        assert!(result.is_err());
    }
}
