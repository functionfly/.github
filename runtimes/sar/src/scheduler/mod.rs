use std::collections::{BinaryHeap, HashMap, HashSet};
use std::sync::Arc;
use std::time::Instant;
use std::cmp::Ordering;

use parking_lot::RwLock;
use serde::{Deserialize, Serialize};
use tokio::sync::mpsc;
use tracing::debug;
use uuid::Uuid;

use crate::core::AgentId;
use crate::engine::{
    ExecutionContext, Graph, GraphExecutionInput, GraphExecutionResult,
    GraphExecutor, DefaultNodeExecutor, ExecutionStatus,
};

#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize)]
pub enum ExecutionPriority {
    Low = 1,
    Normal = 2,
    High = 3,
    Critical = 4,
}

impl Default for ExecutionPriority {
    fn default() -> Self {
        ExecutionPriority::Normal
    }
}

#[derive(Debug, Clone)]
pub struct PendingExecution {
    pub execution_id: Uuid,
    pub agent_id: AgentId,
    pub graph: Graph,
    pub input: HashMap<String, serde_json::Value>,
    pub priority: ExecutionPriority,
    pub created_at: Instant,
}

impl PartialEq for PendingExecution {
    fn eq(&self, other: &Self) -> bool {
        self.execution_id == other.execution_id
    }
}

impl Eq for PendingExecution {}

impl PartialOrd for PendingExecution {
    fn partial_cmp(&self, other: &Self) -> Option<Ordering> {
        Some(self.cmp(other))
    }
}

impl Ord for PendingExecution {
    fn cmp(&self, other: &Self) -> Ordering {
        self.priority.cmp(&other.priority)
            .then_with(|| other.created_at.cmp(&self.created_at))
    }
}

#[derive(Debug, Clone)]
pub struct SchedulerConfig {
    pub max_concurrent: usize,
    pub max_queue_size: usize,
    pub rate_limit_per_second: u64,
    pub backpressure_threshold: usize,
}

impl Default for SchedulerConfig {
    fn default() -> Self {
        Self {
            max_concurrent: 10_000,
            max_queue_size: 1_000_000,
            rate_limit_per_second: 100_000,
            backpressure_threshold: 800_000,
        }
    }
}

pub struct AgentScheduler {
    config: SchedulerConfig,
    queue: Arc<RwLock<BinaryHeap<PendingExecution>>>,
    running: Arc<RwLock<HashSet<Uuid>>>,
    execution_tx: mpsc::Sender<PendingExecution>,
    metrics: Arc<RwLock<SchedulerMetrics>>,
}

#[derive(Debug, Clone, Default)]
pub struct SchedulerMetrics {
    pub queued_tasks: u64,
    pub running_tasks: u64,
    pub completed_tasks: u64,
    pub failed_tasks: u64,
    pub average_latency_ms: f64,
}

impl AgentScheduler {
    pub fn new(config: SchedulerConfig) -> Self {
        let (tx, mut rx) = mpsc::channel::<PendingExecution>(config.max_queue_size);
        let queue = Arc::new(RwLock::new(BinaryHeap::new()));
        let running = Arc::new(RwLock::new(HashSet::new()));
        let metrics = Arc::new(RwLock::new(SchedulerMetrics::default()));

        let running_clone = running.clone();
        let metrics_clone = metrics.clone();

        tokio::spawn(async move {
            let executor = GraphExecutor::new(DefaultNodeExecutor);

            while let Some(execution) = rx.recv().await {
                let r = running_clone.clone();
                let m = metrics_clone.clone();

                {
                    let mut running = r.write();
                    running.insert(execution.execution_id);
                    m.write().running_tasks += 1;
                }

                let ctx = Arc::new(ExecutionContext::new(execution.execution_id, None, None));
                let graph_input = GraphExecutionInput {
                    graph_id: execution.execution_id,
                    initial_input: execution.input,
                    tenant_id: None,
                };
                let start = Instant::now();
                let result = executor.execute(&execution.graph, graph_input, ctx).await;
                let elapsed = start.elapsed().as_millis() as f64;

                {
                    let mut running = r.write();
                    running.remove(&execution.execution_id);
                    let mut m = m.write();
                    m.running_tasks -= 1;
                    m.completed_tasks += 1;

                    if result.status == ExecutionStatus::Failed {
                        m.failed_tasks += 1;
                    }

                    m.average_latency_ms = m.average_latency_ms * 0.9 + elapsed * 0.1;
                }

                debug!(execution_id = %execution.execution_id, duration_ms = elapsed as u64, "Execution completed");
            }
        });

        Self {
            config,
            queue,
            running,
            execution_tx: tx,
            metrics,
        }
    }

    pub async fn schedule_and_execute(
        &self,
        _agent_id: AgentId,
        input: HashMap<String, serde_json::Value>,
        ctx: Arc<ExecutionContext>,
    ) -> GraphExecutionResult {
        let graph = Graph::new(Uuid::nil(), "execution".to_string());
        let graph_input = GraphExecutionInput {
            graph_id: Uuid::new_v4(),
            initial_input: input,
            tenant_id: None,
        };

        let executor = GraphExecutor::new(DefaultNodeExecutor);
        executor.execute(&graph, graph_input, ctx).await
    }

    pub async fn submit(&self, execution: PendingExecution) -> anyhow::Result<()> {
        let queue_len = { self.queue.read().len() };

        if queue_len >= self.config.max_queue_size {
            return Err(anyhow::anyhow!("Queue full, backpressure triggered"));
        }

        {
            let mut queue = self.queue.write();
            queue.push(execution.clone());
            self.metrics.write().queued_tasks = queue.len() as u64;
        }

        self.execution_tx.send(execution).await
            .map_err(|_| anyhow::anyhow!("Failed to submit execution"))?;

        Ok(())
    }

    pub fn get_metrics(&self) -> SchedulerMetrics {
        self.metrics.read().clone()
    }

    pub fn queue_depth(&self) -> usize {
        self.queue.read().len()
    }

    pub fn running_count(&self) -> usize {
        self.running.read().len()
    }
}
