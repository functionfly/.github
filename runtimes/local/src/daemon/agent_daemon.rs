//! Agent Daemon for Always-On Agent Mode (Phase 2)
//!
//! The daemon enables agents to run continuously, reacting to events
//! from webhooks, database changes, and scheduled triggers.
//!
//! ## Architecture
//!
//! The daemon uses an event-driven architecture:
//! - `EventSource` trait abstracts different event sources
//! - `AgentDaemon` manages running agents and their event subscriptions
//! - Events trigger graph executions via the existing scheduler
//!
//! ## Event Sources
//!
//! - Webhook: Stripe, Shopify, Resend webhooks
//! - Database: PostgreSQL LISTEN/NOTIFY
//! - Scheduled: Cron-like triggers
//! - MessageQueue: NATS/Redis Streams (future)

use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};

use tokio::sync::{RwLock, mpsc};
use tracing::{debug, error, info, warn};
use uuid::Uuid;
use async_trait::async_trait;

use crate::agent_scheduler::agent_scheduler::{AgentScheduler, PriorityLevel};
use crate::agent_scheduler::worker::QueuedGraphExecution;
use crate::engine::graph::{Graph, GraphExecutionInput};
use crate::engine::sar_executor::SarNodeExecutor;

// Import event source types for the enum wrapper
use crate::daemon::event_sources::{WebhookEventSource, DatabaseEventSource, ScheduledEventSource};

// ---------------------------------------------------------------------------
// Event Types
// ---------------------------------------------------------------------------

/// An event that can trigger agent execution
#[derive(Debug, Clone)]
pub struct AgentEvent {
    pub id: Uuid,
    pub source: EventSourceType,
    pub agent_id: String,
    pub payload: serde_json::Value,
    pub timestamp: Instant,
}

/// Types of event sources
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum EventSourceType {
    Webhook { name: String },
    Database { table: String, operation: String },
    Scheduled { schedule_id: String },
    MessageQueue { queue: String },
}

impl EventSourceType {
    pub fn name(&self) -> String {
        match self {
            EventSourceType::Webhook { name } => format!("webhook:{}", name),
            EventSourceType::Database { table, operation } => {
                format!("db:{}:{}", table, operation)
            }
            EventSourceType::Scheduled { schedule_id } => format!("scheduled:{}", schedule_id),
            EventSourceType::MessageQueue { queue } => format!("queue:{}", queue),
        }
    }
}

// ---------------------------------------------------------------------------
// EventSource Trait
// ---------------------------------------------------------------------------

/// Trait for event sources that can trigger agent execution
#[async_trait]
pub trait EventSource: Send + Sync {
    /// Returns the type of this event source
    fn source_type(&self) -> EventSourceType;

    /// Start listening for events
    /// Returns a receiver channel for events
    async fn start(&self) -> anyhow::Result<mpsc::Receiver<AgentEvent>>;

    /// Stop listening for events
    async fn stop(&self) -> anyhow::Result<()>;

    /// Check if this source is currently running
    fn is_running(&self) -> bool;
}

// ---------------------------------------------------------------------------
// EventSource Enum (for dyn compatibility)
// ---------------------------------------------------------------------------

/// Enum wrapper for event sources to enable dyn compatibility.
/// This wraps all concrete event source types to allow storing them
/// in collections without using trait objects.
#[derive(Clone, Debug)]
pub enum EventSourceWrapper {
    Webhook(Arc<WebhookEventSource>),
    Database(Arc<DatabaseEventSource>),
    Scheduled(Arc<ScheduledEventSource>),
}

impl EventSourceWrapper {
    /// Returns the type of this event source
    pub fn source_type(&self) -> EventSourceType {
        match self {
            EventSourceWrapper::Webhook(s) => s.source_type(),
            EventSourceWrapper::Database(s) => s.source_type(),
            EventSourceWrapper::Scheduled(s) => s.source_type(),
        }
    }

    /// Start listening for events
    pub async fn start(&self) -> anyhow::Result<mpsc::Receiver<AgentEvent>> {
        match self {
            EventSourceWrapper::Webhook(s) => s.start().await,
            EventSourceWrapper::Database(s) => s.start().await,
            EventSourceWrapper::Scheduled(s) => s.start().await,
        }
    }

    /// Stop listening for events
    pub async fn stop(&self) -> anyhow::Result<()> {
        match self {
            EventSourceWrapper::Webhook(s) => s.stop().await,
            EventSourceWrapper::Database(s) => s.stop().await,
            EventSourceWrapper::Scheduled(s) => s.stop().await,
        }
    }

    /// Check if this source is currently running
    pub fn is_running(&self) -> bool {
        match self {
            EventSourceWrapper::Webhook(s) => s.is_running(),
            EventSourceWrapper::Database(s) => s.is_running(),
            EventSourceWrapper::Scheduled(s) => s.is_running(),
        }
    }
}

// ---------------------------------------------------------------------------
// RunningAgent State
// ---------------------------------------------------------------------------

/// State for a running agent in daemon mode
#[derive(Debug, Clone)]
pub struct RunningAgent {
    pub agent_id: String,
    pub tenant_id: String,
    pub graph_id: Uuid,
    pub graph: Graph,
    pub started_at: Instant,
    pub last_activity: Instant,
    pub event_sources: Vec<EventSourceWrapper>,
    pub execution_count: u64,
    pub is_shutting_down: bool,
}

impl RunningAgent {
    /// Create new running agent state
    pub fn new(agent_id: String, tenant_id: String, graph_id: Uuid, graph: Graph) -> Self {
        let now = Instant::now();
        Self {
            agent_id,
            tenant_id,
            graph_id,
            graph,
            started_at: now,
            last_activity: now,
            event_sources: Vec::new(),
            execution_count: 0,
            is_shutting_down: false,
        }
    }

    /// Record activity to prevent idle shutdown
    pub fn record_activity(&mut self) {
        self.last_activity = Instant::now();
    }

    /// Increment execution count
    pub fn record_execution(&mut self) {
        self.execution_count += 1;
        self.record_activity();
    }

    /// Check if agent has been idle for too long
    pub fn is_idle(&self, timeout: Duration) -> bool {
        self.last_activity.elapsed() > timeout
    }
}

// ---------------------------------------------------------------------------
// AgentDaemon
// ---------------------------------------------------------------------------

/// Daemon that manages always-on agents
///
/// The daemon maintains a pool of running agents that subscribe to events
/// and execute graphs when events are received.
pub struct AgentDaemon {
    scheduler: Arc<RwLock<AgentScheduler>>,
    executor: Option<Arc<SarNodeExecutor>>,
    running_agents: RwLock<HashMap<String, RunningAgent>>,
    tenant_contexts: RwLock<HashMap<String, TenantAgentState>>,
    event_tx: mpsc::Sender<AgentEvent>,
    event_rx: RwLock<mpsc::Receiver<AgentEvent>>,
    idle_timeout: Duration,
    max_executions_per_day: u64,
}

/// Per-tenant agent state for isolation
#[derive(Debug, Clone)]
pub struct TenantAgentState {
    pub tenant_id: String,
    pub agent_count: usize,
    pub total_executions_today: u64,
    pub quota_remaining: u64,
}

impl AgentDaemon {
    /// Create a new agent daemon
    pub fn new(
        scheduler: Arc<RwLock<AgentScheduler>>,
        executor: Option<Arc<SarNodeExecutor>>,
        idle_timeout: Duration,
    ) -> (Self, mpsc::Sender<AgentEvent>) {
        let (event_tx, event_rx) = mpsc::channel(1000);

        let daemon = Self {
            scheduler,
            executor,
            running_agents: RwLock::new(HashMap::new()),
            tenant_contexts: RwLock::new(HashMap::new()),
            event_tx: event_tx.clone(),
            event_rx: RwLock::new(event_rx),
            idle_timeout,
            max_executions_per_day: 100, // Free tier limit
        };

        (daemon, event_tx)
    }

    /// Start the daemon event loop
    pub async fn start(&self) {
        info!("AgentDaemon starting event loop");

        loop {
            // Check for idle agents periodically
            self.check_idle_agents().await;

            // Process events
            let mut rx = self.event_rx.write().await;
            match tokio::time::timeout(Duration::from_secs(5), rx.recv()).await {
                Ok(Some(event)) => {
                    drop(rx); // Release lock before processing
                    if let Err(e) = self.handle_event(event).await {
                        error!("Failed to handle event: {}", e);
                    }
                }
                Ok(None) => {
                    // Channel closed
                    break;
                }
                Err(_) => {
                    // Timeout - continue to idle check
                }
            }
        }

        info!("AgentDaemon event loop stopped");
    }

    /// Handle an incoming event
    async fn handle_event(&self, event: AgentEvent) -> anyhow::Result<()> {
        debug!(
            agent_id = %event.agent_id,
            source = %event.source.name(),
            "Handling agent event"
        );

        // Get the running agent
        let agent = {
            let agents = self.running_agents.read().await;
            agents.get(&event.agent_id).cloned()
        };

        if let Some(mut agent) = agent {
            // Check tenant quota
            if !self.check_tenant_quota(&agent.tenant_id).await {
                warn!(
                    tenant_id = %agent.tenant_id,
                    "Tenant quota exceeded, dropping event"
                );
                return Ok(());
            }

            // Record activity
            agent.record_activity();

            // Determine priority based on event source type
            let priority = priority_from_event_source(&event.source);

            // Enqueue execution via scheduler with appropriate priority
            self.enqueue_execution_with_priority(&agent, event.payload, priority).await?;

            // Update execution count
            let mut agents = self.running_agents.write().await;
            if let Some(a) = agents.get_mut(&event.agent_id) {
                a.record_execution();
            }
        } else {
            warn!(
                agent_id = %event.agent_id,
                "Received event for non-running agent"
            );
        }

        Ok(())
    }

    /// Start an agent in daemon mode
    pub async fn start_agent(
        &self,
        agent_id: String,
        tenant_id: String,
        graph_id: Uuid,
        graph: Graph,
        event_sources: Vec<EventSourceWrapper>,
    ) -> anyhow::Result<()> {
        info!(
            agent_id = %agent_id,
            tenant_id = %tenant_id,
            graph_id = %graph_id,
            "Starting agent in daemon mode"
        );

        // Check tenant agent limit
        {
            let contexts = self.tenant_contexts.read().await;
            if let Some(ctx) = contexts.get(&tenant_id) {
                if ctx.agent_count >= 10 {
                    // Pro limit
                    return Err(anyhow::anyhow!(
                        "Tenant agent limit reached (10 agents max on Pro plan)"
                    ));
                }
            }
        }

        // Initialize agent state
        let mut agent = RunningAgent::new(agent_id.clone(), tenant_id.clone(), graph_id, graph);
        agent.event_sources = event_sources;

        // Start event sources
        for source in &agent.event_sources {
            let event_tx = self.event_tx.clone();
            let agent_id = agent_id.clone();
            let source = source.clone();

            tokio::spawn(async move {
                match source.start().await {
                    Ok(mut rx) => {
                        while let Some(mut event) = rx.recv().await {
                            event.agent_id = agent_id.clone();
                            if let Err(e) = event_tx.send(event).await {
                                error!("Failed to forward event: {}", e);
                                break;
                            }
                        }
                    }
                    Err(e) => {
                        error!("Failed to start event source: {}", e);
                    }
                }
            });
        }

        // Store agent
        {
            let mut agents = self.running_agents.write().await;
            agents.insert(agent_id, agent);
        }

        // Update tenant context
        {
            let mut contexts = self.tenant_contexts.write().await;
            let ctx = contexts.entry(tenant_id.clone()).or_insert_with(|| {
                TenantAgentState {
                    tenant_id: tenant_id.clone(),
                    agent_count: 0,
                    total_executions_today: 0,
                    quota_remaining: self.max_executions_per_day,
                }
            });
            ctx.agent_count += 1;
        }

        Ok(())
    }

    /// Stop a running agent
    pub async fn stop_agent(&self, agent_id: &str) -> anyhow::Result<()> {
        info!(agent_id = %agent_id, "Stopping agent daemon");

        let agent = {
            let mut agents = self.running_agents.write().await;
            agents.remove(agent_id)
        };

        if let Some(agent) = agent {
            // Stop event sources
            for source in &agent.event_sources {
                if let Err(e) = source.stop().await {
                    error!("Failed to stop event source: {}", e);
                }
            }

            // Update tenant context
            let mut contexts = self.tenant_contexts.write().await;
            if let Some(ctx) = contexts.get_mut(&agent.tenant_id) {
                ctx.agent_count = ctx.agent_count.saturating_sub(1);
            }
        }

        Ok(())
    }

    /// Get daemon status
    pub async fn status(&self) -> DaemonStatus {
        let agents = self.running_agents.read().await;
        let contexts = self.tenant_contexts.read().await;

        DaemonStatus {
            running_agents: agents.len(),
            active_tenants: contexts.len(),
            total_executions_today: contexts.values().map(|c| c.total_executions_today).sum(),
        }
    }

    /// Check for idle agents and shut them down
    async fn check_idle_agents(&self) {
        let idle_ids: Vec<String> = {
            let agents = self.running_agents.read().await;
            agents
                .iter()
                .filter(|(_, agent)| agent.is_idle(self.idle_timeout))
                .map(|(id, _)| id.clone())
                .collect()
        };

        for agent_id in idle_ids {
            info!(
                agent_id = %agent_id,
                "Agent idle, shutting down"
            );
            if let Err(e) = self.stop_agent(&agent_id).await {
                error!("Failed to stop idle agent: {}", e);
            }
        }
    }

    /// Check if tenant has quota remaining
    async fn check_tenant_quota(&self, tenant_id: &str) -> bool {
        let contexts = self.tenant_contexts.read().await;
        if let Some(ctx) = contexts.get(tenant_id) {
            ctx.quota_remaining > 0
        } else {
            true // No record, allow
        }
    }

    /// Enqueue execution for an agent using the scheduler's priority queues.
    ///
    /// Priority is determined by event source type:
    /// - Webhook events: High priority (user-facing, time-sensitive)
    /// - Scheduled events: Normal priority (background processing)
    /// - Database events: Normal priority
    /// - MessageQueue: Low priority (batch processing)
    async fn enqueue_execution(
        &self,
        agent: &RunningAgent,
        payload: serde_json::Value,
    ) -> anyhow::Result<()> {
        debug!(
            agent_id = %agent.agent_id,
            graph_id = %agent.graph_id,
            "Enqueuing agent execution via scheduler"
        );

        // Determine priority based on event source type
        // This would ideally come from the event itself, but we use a sensible default
        // based on typical use cases
        let priority = PriorityLevel::Normal;

        // Build initial input from the event payload
        let mut initial_input = std::collections::HashMap::new();
        initial_input.insert("event".to_string(), payload);
        initial_input.insert("agent_id".to_string(), serde_json::Value::String(agent.agent_id.clone()));

        let input = GraphExecutionInput {
            graph_id: agent.graph_id,
            initial_input,
            tenant_id: Some(agent.tenant_id.clone()),
        };

        // Create the queued execution with the full graph
        let queued_exec = QueuedGraphExecution::new(
            agent.graph.clone(),
            input,
            priority,
            Some(agent.tenant_id.clone()),
        );

        // Enqueue via the scheduler
        let mut scheduler = self.scheduler.write().await;
        match scheduler.enqueue(queued_exec).await {
            Ok(exec_id) => {
                debug!(
                    exec_id = %exec_id,
                    agent_id = %agent.agent_id,
                    priority = ?priority,
                    "Successfully enqueued execution"
                );
                Ok(())
            }
            Err(e) => {
                error!(
                    agent_id = %agent.agent_id,
                    error = %e,
                    "Failed to enqueue execution"
                );
                Err(anyhow::anyhow!("Scheduler enqueue failed: {}", e))
            }
        }
    }

    /// Enqueue execution with a specific priority level.
    /// Used when the event source type is known and priority should be adjusted.
    async fn enqueue_execution_with_priority(
        &self,
        agent: &RunningAgent,
        payload: serde_json::Value,
        priority: PriorityLevel,
    ) -> anyhow::Result<()> {
        debug!(
            agent_id = %agent.agent_id,
            graph_id = %agent.graph_id,
            priority = ?priority,
            "Enqueuing agent execution with specific priority"
        );

        let mut initial_input = std::collections::HashMap::new();
        initial_input.insert("event".to_string(), payload);
        initial_input.insert("agent_id".to_string(), serde_json::Value::String(agent.agent_id.clone()));

        let input = GraphExecutionInput {
            graph_id: agent.graph_id,
            initial_input,
            tenant_id: Some(agent.tenant_id.clone()),
        };

        let queued_exec = QueuedGraphExecution::new(
            agent.graph.clone(),
            input,
            priority,
            Some(agent.tenant_id.clone()),
        );

        let mut scheduler = self.scheduler.write().await;
        match scheduler.enqueue(queued_exec).await {
            Ok(exec_id) => {
                debug!(
                    exec_id = %exec_id,
                    agent_id = %agent.agent_id,
                    priority = ?priority,
                    "Successfully enqueued execution"
                );
                Ok(())
            }
            Err(e) => {
                error!(
                    agent_id = %agent.agent_id,
                    error = %e,
                    "Failed to enqueue execution"
                );
                Err(anyhow::anyhow!("Scheduler enqueue failed: {}", e))
            }
        }
    }

    /// Shutdown all agents and stop the daemon
    pub async fn shutdown(&self) {
        info!("AgentDaemon shutting down");

        let agent_ids: Vec<String> = {
            let agents = self.running_agents.read().await;
            agents.keys().cloned().collect()
        };

        for agent_id in agent_ids {
            if let Err(e) = self.stop_agent(&agent_id).await {
                error!("Failed to stop agent during shutdown: {}", e);
            }
        }
    }
}

/// Daemon status information
#[derive(Debug, Clone)]
pub struct DaemonStatus {
    pub running_agents: usize,
    pub active_tenants: usize,
    pub total_executions_today: u64,
}

// ---------------------------------------------------------------------------
// Priority Mapping
// ---------------------------------------------------------------------------

/// Map event source type to scheduler priority level.
///
/// Priority guidelines:
/// - Critical: System alerts, error conditions requiring immediate attention
/// - High: User-facing webhooks (Stripe, Shopify), time-sensitive operations
/// - Normal: Scheduled tasks, database change events
/// - Low: Background processing, batch message queue operations
fn priority_from_event_source(source: &EventSourceType) -> PriorityLevel {
    match source {
        EventSourceType::Webhook { name } => {
            // Payment webhooks and user-facing operations get high priority
            let lower = name.to_lowercase();
            if lower.contains("stripe") || lower.contains("payment") || lower.contains("checkout") {
                PriorityLevel::High
            } else {
                PriorityLevel::Normal
            }
        }
        EventSourceType::Scheduled { .. } => PriorityLevel::Normal,
        EventSourceType::Database { .. } => PriorityLevel::Normal,
        EventSourceType::MessageQueue { .. } => PriorityLevel::Low,
    }
}

// ---------------------------------------------------------------------------
// Free Tier Limits
// ---------------------------------------------------------------------------

/// Configuration for free tier limits
pub struct FreeTierConfig {
    pub max_agents: usize,
    pub max_executions_per_day: u64,
    pub idle_timeout_minutes: u64,
}

impl FreeTierConfig {
    pub fn default() -> Self {
        Self {
            max_agents: 1,
            max_executions_per_day: 100,
            idle_timeout_minutes: 15,
        }
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use crate::engine::graph::Graph;

    #[test]
    fn test_running_agent_activity() {
        let graph = Graph::new(Uuid::new_v4(), "test-graph".to_string());
        let mut agent = RunningAgent::new(
            "test-agent".to_string(),
            "tenant-1".to_string(),
            graph.id,
            graph,
        );

        assert_eq!(agent.execution_count, 0);
        agent.record_execution();
        assert_eq!(agent.execution_count, 1);
        assert!(!agent.is_idle(Duration::from_secs(60)));
    }

    #[tokio::test]
    async fn test_daemon_status() {
        let scheduler = Arc::new(RwLock::new(AgentScheduler::new(Default::default())));
        let (daemon, _tx) = AgentDaemon::new(scheduler, None, Duration::from_secs(300));

        let status = daemon.status().await;
        assert_eq!(status.running_agents, 0);
        assert_eq!(status.active_tenants, 0);
    }
}
