use std::sync::Arc;
use chrono::{DateTime, Utc};
use parking_lot::RwLock;
use serde::{Deserialize, Serialize};
use sqlx::postgres::{PgPool, PgPoolOptions};
use thiserror::Error;
use tracing::info;
use uuid::Uuid;

use crate::core::{AgentConfig, AgentId, AgentMetrics, AgentStatus};
use crate::engine::Graph;

#[derive(Error, Debug)]
pub enum PersistenceError {
    #[error("Database error: {0}")]
    Database(#[from] sqlx::Error),
    #[error("Agent not found: {0}")]
    NotFound(AgentId),
    #[error("Serialization error: {0}")]
    Serialization(String),
}

pub struct AgentRepository {
    pool: Option<PgPool>,
    memory_store: Arc<RwLock<Vec<AgentPersistence>>>,
}

impl AgentRepository {
    pub async fn new(database_url: &str) -> Result<Self, PersistenceError> {
        let pool = PgPoolOptions::new()
            .max_connections(10)
            .connect(database_url)
            .await?;

        Ok(Self { pool: Some(pool), memory_store: Arc::new(RwLock::new(Vec::new())) })
    }

    pub fn new_in_memory() -> Self {
        Self { pool: None, memory_store: Arc::new(RwLock::new(Vec::new())) }
    }

    pub async fn init_schema(&self) -> Result<(), PersistenceError> {
        if let Some(ref pool) = self.pool {
            sqlx::query(
                r#"
                CREATE TABLE IF NOT EXISTS agents (
                    id TEXT PRIMARY KEY,
                    name TEXT NOT NULL,
                    graph_json TEXT NOT NULL,
                    priority INTEGER NOT NULL DEFAULT 2,
                    max_concurrent_cells INTEGER NOT NULL DEFAULT 100,
                    isolation_enabled BOOLEAN NOT NULL DEFAULT true,
                    event_subscriptions TEXT[] NOT NULL DEFAULT '{}',
                    metadata_json TEXT NOT NULL DEFAULT '{}',
                    status TEXT NOT NULL DEFAULT 'Idle',
                    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                    last_heartbeat TIMESTAMPTZ,
                    metrics_json TEXT NOT NULL DEFAULT '{"total_executions":0,"successful_executions":0,"failed_executions":0,"average_latency_ms":0,"total_cost_usd":0}'
                );

                CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
                CREATE INDEX IF NOT EXISTS idx_agents_name ON agents(name);
                "#,
            )
            .execute(pool)
            .await?;
            info!("Agent persistence schema initialized");
        }
        Ok(())
    }

    #[tracing::instrument(skip(self))]
    pub async fn save(&self, agent: &AgentPersistence) -> Result<(), PersistenceError> {
        if let Some(ref pool) = self.pool {
            let graph_json = serde_json::to_string(&agent.graph)
                .map_err(|e| PersistenceError::Serialization(e.to_string()))?;
            let metadata_json = serde_json::to_string(&agent.metadata)
                .map_err(|e| PersistenceError::Serialization(e.to_string()))?;
            let metrics_json = serde_json::to_string(&agent.metrics)
                .map_err(|e| PersistenceError::Serialization(e.to_string()))?;

            sqlx::query(
                r#"
                INSERT INTO agents (id, name, graph_json, priority, max_concurrent_cells, isolation_enabled, event_subscriptions, metadata_json, status, registered_at, last_heartbeat, metrics_json)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
                ON CONFLICT (id) DO UPDATE SET
                    name = EXCLUDED.name,
                    graph_json = EXCLUDED.graph_json,
                    priority = EXCLUDED.priority,
                    max_concurrent_cells = EXCLUDED.max_concurrent_cells,
                    isolation_enabled = EXCLUDED.isolation_enabled,
                    event_subscriptions = EXCLUDED.event_subscriptions,
                    metadata_json = EXCLUDED.metadata_json,
                    status = EXCLUDED.status,
                    last_heartbeat = EXCLUDED.last_heartbeat,
                    metrics_json = EXCLUDED.metrics_json
                "#,
            )
            .bind(&agent.id.to_string())
            .bind(&agent.name)
            .bind(&graph_json)
            .bind(agent.priority as i32)
            .bind(agent.max_concurrent_cells as i32)
            .bind(agent.isolation_enabled)
            .bind(&agent.event_subscriptions)
            .bind(&metadata_json)
            .bind(agent.status.as_str())
            .bind(agent.registered_at)
            .bind(agent.last_heartbeat)
            .bind(&metrics_json)
            .execute(pool)
            .await?;
        } else {
            let mut store = self.memory_store.write();
            if let Some(idx) = store.iter().position(|a| a.id == agent.id) {
                store[idx] = agent.clone();
            } else {
                store.push(agent.clone());
            }
        }
        Ok(())
    }

    pub async fn find_by_id(&self, agent_id: &AgentId) -> Result<Option<AgentPersistence>, PersistenceError> {
        if let Some(ref pool) = self.pool {
            let row = sqlx::query_scalar::<_, (String, String, String, String, Option<DateTime<Utc>>)>(
                "SELECT id, name, graph_json, status, last_heartbeat FROM agents WHERE id = $1",
            )
            .bind(agent_id.to_string())
            .fetch_optional(pool)
            .await?;

            match row {
                Some((id, name, graph_json, status, last_heartbeat)) => {
                    let graph: Graph = serde_json::from_str(&graph_json)
                        .map_err(|e| PersistenceError::Serialization(e.to_string()))?;
                    let status = AgentStatus::from_str(&status);

                    Ok(Some(AgentPersistence {
                        id: AgentId(Uuid::parse_str(&id).unwrap()),
                        name,
                        graph,
                        priority: 2,
                        max_concurrent_cells: 100,
                        isolation_enabled: true,
                        event_subscriptions: vec![],
                        metadata: Default::default(),
                        status,
                        registered_at: Utc::now(),
                        last_heartbeat,
                        metrics: AgentMetrics::default(),
                    }))
                }
                None => Ok(None),
            }
        } else {
            let store = self.memory_store.read();
            Ok(store.iter().find(|a| a.id == *agent_id).cloned())
        }
    }

    pub async fn list_all(&self) -> Result<Vec<AgentPersistence>, PersistenceError> {
        if let Some(ref pool) = self.pool {
            let rows = sqlx::query_as::<_, (String, String, String, i32, i32, bool, Vec<String>, String, String, DateTime<Utc>, Option<DateTime<Utc>>, String)>(
                "SELECT id, name, graph_json, priority, max_concurrent_cells, isolation_enabled, event_subscriptions, metadata_json, status, registered_at, last_heartbeat, metrics_json FROM agents",
            )
            .fetch_all(pool)
            .await?;

            let agents = rows
                .into_iter()
                .map(|row| {
                    let (id, name, graph_json, priority, max_concurrent_cells, isolation_enabled, event_subscriptions, metadata_json, status, registered_at, last_heartbeat, metrics_json) = row;
                    let graph: Graph = serde_json::from_str(&graph_json).unwrap_or_else(|_| Graph::new(Uuid::new_v4(), "unknown".to_string()));
                    let metadata: std::collections::HashMap<String, String> = serde_json::from_str(&metadata_json).unwrap_or_default();
                    let metrics: AgentMetrics = serde_json::from_str(&metrics_json).unwrap_or_default();

                    AgentPersistence {
                        id: AgentId(Uuid::parse_str(&id).unwrap_or_else(|_| Uuid::new_v4())),
                        name,
                        graph,
                        priority: priority as u8,
                        max_concurrent_cells: max_concurrent_cells as usize,
                        isolation_enabled,
                        event_subscriptions,
                        metadata,
                        status: AgentStatus::from_str(&status),
                        registered_at,
                        last_heartbeat,
                        metrics,
                    }
                })
                .collect();

            Ok(agents)
        } else {
            Ok(self.memory_store.read().clone())
        }
    }

    pub async fn delete(&self, agent_id: &AgentId) -> Result<bool, PersistenceError> {
        if let Some(ref pool) = self.pool {
            let result = sqlx::query("DELETE FROM agents WHERE id = $1")
                .bind(agent_id.to_string())
                .execute(pool)
                .await?;
            Ok(result.rows_affected() > 0)
        } else {
            let mut store = self.memory_store.write();
            let len_before = store.len();
            store.retain(|a| a.id != *agent_id);
            Ok(store.len() < len_before)
        }
    }

    pub async fn update_heartbeat(&self, agent_id: &AgentId) -> Result<(), PersistenceError> {
        if let Some(ref pool) = self.pool {
            sqlx::query("UPDATE agents SET last_heartbeat = NOW() WHERE id = $1")
                .bind(agent_id.to_string())
                .execute(pool)
                .await?;
        }
        Ok(())
    }

    pub async fn update_status(&self, agent_id: &AgentId, status: AgentStatus) -> Result<(), PersistenceError> {
        if let Some(ref pool) = self.pool {
            sqlx::query("UPDATE agents SET status = $1 WHERE id = $2")
                .bind(status.as_str())
                .bind(agent_id.to_string())
                .execute(pool)
                .await?;
        }
        Ok(())
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentPersistence {
    pub id: AgentId,
    pub name: String,
    pub graph: Graph,
    pub priority: u8,
    pub max_concurrent_cells: usize,
    pub isolation_enabled: bool,
    pub event_subscriptions: Vec<String>,
    pub metadata: std::collections::HashMap<String, String>,
    pub status: AgentStatus,
    pub registered_at: DateTime<Utc>,
    pub last_heartbeat: Option<DateTime<Utc>>,
    pub metrics: AgentMetrics,
}

impl From<AgentPersistence> for AgentConfig {
    fn from(p: AgentPersistence) -> Self {
        AgentConfig {
            id: p.id,
            name: p.name,
            graph: p.graph,
            priority: p.priority,
            max_concurrent_cells: p.max_concurrent_cells,
            isolation_enabled: p.isolation_enabled,
            event_subscriptions: p.event_subscriptions,
        }
    }
}

impl AgentStatus {
    pub fn as_str(&self) -> &'static str {
        match self {
            AgentStatus::Idle => "Idle",
            AgentStatus::Running => "Running",
            AgentStatus::WaitingForEvent => "WaitingForEvent",
            AgentStatus::Paused => "Paused",
            AgentStatus::Failed => "Failed",
        }
    }

    pub fn from_str(s: &str) -> Self {
        match s {
            "Running" => AgentStatus::Running,
            "WaitingForEvent" => AgentStatus::WaitingForEvent,
            "Paused" => AgentStatus::Paused,
            "Failed" => AgentStatus::Failed,
            _ => AgentStatus::Idle,
        }
    }
}

pub struct CachedAgentRegistry {
    repository: Arc<AgentRepository>,
    cache: Arc<RwLock<std::collections::HashMap<AgentId, AgentPersistence>>>,
}

impl CachedAgentRegistry {
    pub fn new(repository: Arc<AgentRepository>) -> Self {
        Self {
            repository,
            cache: Arc::new(RwLock::new(std::collections::HashMap::new())),
        }
    }

    pub async fn load_all(&self) -> Result<(), PersistenceError> {
        let agents = self.repository.list_all().await?;
        let mut cache = self.cache.write();
        for agent in agents {
            cache.insert(agent.id, agent);
        }
        info!(count = cache.len(), "Loaded agents from persistence");
        Ok(())
    }

    pub fn get_cached(&self, agent_id: &AgentId) -> Option<AgentPersistence> {
        self.cache.read().get(agent_id).cloned()
    }

    pub fn list_cached(&self) -> Vec<AgentPersistence> {
        self.cache.read().values().cloned().collect()
    }

    pub async fn register(&self, config: AgentConfig) -> Result<AgentPersistence, PersistenceError> {
        let persistence = AgentPersistence {
            id: config.id,
            name: config.name,
            graph: config.graph,
            priority: config.priority,
            max_concurrent_cells: config.max_concurrent_cells,
            isolation_enabled: config.isolation_enabled,
            event_subscriptions: config.event_subscriptions,
            metadata: Default::default(),
            status: AgentStatus::Idle,
            registered_at: Utc::now(),
            last_heartbeat: None,
            metrics: AgentMetrics::default(),
        };

        self.repository.save(&persistence).await?;
        {
            let mut cache = self.cache.write();
            cache.insert(persistence.id, persistence.clone());
        }

        Ok(persistence)
    }

    pub async fn unregister(&self, agent_id: &AgentId) -> Result<bool, PersistenceError> {
        let deleted = self.repository.delete(agent_id).await?;
        if deleted {
            let mut cache = self.cache.write();
            cache.remove(agent_id);
        }
        Ok(deleted)
    }

    pub async fn update(&self, agent_id: &AgentId, update: AgentUpdate) -> Result<Option<AgentPersistence>, PersistenceError> {
        let agent_clone = {
            let mut cache = self.cache.write();
            if let Some(mut agent) = cache.get_mut(agent_id) {
                if let Some(name) = update.name {
                    agent.name = name;
                }
                if let Some(graph) = update.graph {
                    agent.graph = graph;
                }
                if let Some(priority) = update.priority {
                    agent.priority = priority;
                }
                if let Some(max_concurrent_cells) = update.max_concurrent_cells {
                    agent.max_concurrent_cells = max_concurrent_cells;
                }
                if let Some(isolation_enabled) = update.isolation_enabled {
                    agent.isolation_enabled = isolation_enabled;
                }
                if let Some(event_subscriptions) = update.event_subscriptions {
                    agent.event_subscriptions = event_subscriptions;
                }
                Some(agent.clone())
            } else {
                None
            }
        };

        if let Some(agent) = agent_clone {
            self.repository.save(&agent).await?;
            let mut cache = self.cache.write();
            cache.insert(*agent_id, agent);
            Ok(cache.get(agent_id).cloned())
        } else {
            Ok(None)
        }
    }

    pub async fn update_status(&self, agent_id: &AgentId, status: AgentStatus) -> Result<(), PersistenceError> {
        self.repository.update_status(agent_id, status).await?;
        if let Some(mut cache) = self.cache.write().get_mut(agent_id) {
            cache.status = status;
        }
        Ok(())
    }

    pub async fn update_heartbeat(&self, agent_id: &AgentId) -> Result<(), PersistenceError> {
        self.repository.update_heartbeat(agent_id).await?;
        if let Some(mut cache) = self.cache.write().get_mut(agent_id) {
            cache.last_heartbeat = Some(Utc::now());
        }
        Ok(())
    }
}

pub struct AgentUpdate {
    pub name: Option<String>,
    pub graph: Option<Graph>,
    pub priority: Option<u8>,
    pub max_concurrent_cells: Option<usize>,
    pub isolation_enabled: Option<bool>,
    pub event_subscriptions: Option<Vec<String>>,
}

impl AgentUpdate {
    pub fn new() -> Self {
        Self {
            name: None,
            graph: None,
            priority: None,
            max_concurrent_cells: None,
            isolation_enabled: None,
            event_subscriptions: None,
        }
    }
}

impl Default for AgentUpdate {
    fn default() -> Self {
        Self::new()
    }
}