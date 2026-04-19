//! Scheduler worker — background task that dequeues and executes graphs.
//!
//! Connects the `AgentScheduler` priority queues to actual graph execution
//! via `SarNodeExecutor`. Each worker processes one execution at a time.

use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};

use tokio::sync::RwLock;
use tracing::{debug, info, instrument, warn};
use uuid::Uuid;

use crate::agent_scheduler::agent_scheduler::AgentScheduler;
use crate::engine::graph::{DefaultNodeExecutor, ExecutionContext, GraphExecutionResult, ExecutionStatus, Graph, GraphExecutionInput, GraphExecutor};
use crate::engine::sar_executor::SarNodeExecutor;
use crate::observability::cost::CostAttributor;

/// Status of an async execution job.
#[derive(Debug, Clone, serde::Serialize)]
pub enum JobStatus {
    /// Job is waiting in queue.
    Queued,
    /// Job is being processed.
    Processing,
    /// Job completed successfully.
    Completed,
    /// Job failed.
    Failed,
    /// Job was cancelled.
    Cancelled,
}

/// Tracked execution job for async processing.
#[derive(Debug, Clone)]
pub struct TrackedJob {
    pub id: Uuid,
    pub graph_id: Uuid,
    pub status: JobStatus,
    pub queued_at: Instant,
    pub started_at: Option<Instant>,
    pub completed_at: Option<Instant>,
    pub result: Option<GraphExecutionResult>,
    pub error: Option<String>,
}

/// In-memory job tracker for scheduled executions.
#[derive(Debug)]
pub struct JobTracker {
    jobs: Arc<RwLock<HashMap<Uuid, TrackedJob>>>,
    /// Max jobs to retain (LRU eviction)
    max_jobs: usize,
}

impl JobTracker {
    pub fn new(max_jobs: usize) -> Self {
        Self {
            jobs: Arc::new(RwLock::new(HashMap::with_capacity(max_jobs))),
            max_jobs: max_jobs.max(1000),
        }
    }

    /// Track a new job.
    pub async fn track(&self, job_id: Uuid, graph_id: Uuid) {
        let mut jobs = self.jobs.write().await;
        
        // Evict oldest if at capacity
        if jobs.len() >= self.max_jobs {
            let to_remove: Vec<Uuid> = jobs
                .iter()
                .filter(|(_, j)| matches!(j.status, JobStatus::Completed | JobStatus::Failed | JobStatus::Cancelled))
                .take(jobs.len() - self.max_jobs + 1)
                .map(|(id, _)| *id)
                .collect();
            for id in to_remove {
                jobs.remove(&id);
            }
        }

        jobs.insert(job_id, TrackedJob {
            id: job_id,
            graph_id,
            status: JobStatus::Queued,
            queued_at: Instant::now(),
            started_at: None,
            completed_at: None,
            result: None,
            error: None,
        });
    }

    /// Mark job as started.
    pub async fn mark_started(&self, job_id: Uuid) {
        let mut jobs = self.jobs.write().await;
        if let Some(job) = jobs.get_mut(&job_id) {
            job.status = JobStatus::Processing;
            job.started_at = Some(Instant::now());
        }
    }

    /// Mark job as completed.
    pub async fn mark_completed(&self, job_id: Uuid, result: GraphExecutionResult) {
        let mut jobs = self.jobs.write().await;
        if let Some(job) = jobs.get_mut(&job_id) {
            job.status = match result.status {
                ExecutionStatus::Completed => JobStatus::Completed,
                _ => JobStatus::Failed,
            };
            job.completed_at = Some(Instant::now());
            job.result = Some(result);
        }
    }

    /// Mark job as failed with error.
    pub async fn mark_failed(&self, job_id: Uuid, error: String) {
        let mut jobs = self.jobs.write().await;
        if let Some(job) = jobs.get_mut(&job_id) {
            job.status = JobStatus::Failed;
            job.completed_at = Some(Instant::now());
            job.error = Some(error);
        }
    }

    /// Get job status.
    pub async fn get_job(&self, job_id: Uuid) -> Option<TrackedJob> {
        self.jobs.read().await.get(&job_id).cloned()
    }

    /// List recent jobs.
    pub async fn list_jobs(&self, limit: usize) -> Vec<TrackedJob> {
        let jobs = self.jobs.read().await;
        jobs.values()
            .cloned()
            .take(limit)
            .collect()
    }

    /// Clean up old completed jobs.
    pub async fn cleanup_old(&self, max_age: Duration) -> usize {
        let mut jobs = self.jobs.write().await;
        let before = jobs.len();
        jobs.retain(|_, job| {
            if let Some(completed) = job.completed_at {
                completed.elapsed() < max_age
            } else {
                true
            }
        });
        before - jobs.len()
    }
}

/// Spawn scheduler workers that actually dequeue and execute graphs.
///
/// Creates `worker_count` workers that continuously:
/// 1. Dequeue the next execution from the scheduler (with full graph definition)
/// 2. Execute the graph using SarNodeExecutor
/// 3. Track results in JobTracker
///
/// Returns the JobTracker for status polling.
pub fn spawn_scheduler_workers(
    scheduler: Arc<RwLock<AgentScheduler>>,
    sar_executor: Option<Arc<SarNodeExecutor>>,
    job_tracker: Arc<JobTracker>,
    cost_attributor: Option<Arc<CostAttributor>>,
    worker_count: usize,
) {
    for worker_id in 0..worker_count {
        let scheduler = scheduler.clone();
        let sar = sar_executor.clone();
        let tracker = job_tracker.clone();
        let cost_attr = cost_attributor.clone();

        tokio::spawn(async move {
            info!(worker_id, "Scheduler worker started");

            loop {
                // Dequeue next execution (now includes full graph)
                let queued = {
                    let mut sched = scheduler.write().await;
                    sched.dequeue().await
                };

                let Some(exec) = queued else {
                    // Queue empty, wait before checking again
                    tokio::time::sleep(Duration::from_millis(100)).await;
                    continue;
                };

                // Execute the job with full graph definition
                execute_job(worker_id, exec, sar.clone(), tracker.clone(), cost_attr.clone()).await;
            }
        });
    }
}

#[instrument(skip(exec, sar, tracker, cost_attr), fields(job_id = %exec.id, graph_id = %exec.graph.id, worker_id))]
async fn execute_job(
    worker_id: usize,
    exec: QueuedGraphExecution,
    sar: Option<Arc<SarNodeExecutor>>,
    tracker: Arc<JobTracker>,
    cost_attr: Option<Arc<CostAttributor>>,
) {
    let job_id = exec.id;
    let graph_id = exec.graph.id;

    debug!(worker_id, job_id = %job_id, graph_id = %graph_id, "Processing scheduled job");
    tracker.mark_started(job_id).await;

    // Execute the graph using the integrated executor
    // The graph is now included in the QueuedGraphExecution, so no separate
    // storage lookup is needed — the "graph builder" is the scheduler itself.
    let result = execute_queued_graph(exec, sar, cost_attr).await;

    // Track the result
    match result.status {
        ExecutionStatus::Completed => {
            tracker.mark_completed(job_id, result).await;
        }
        _ => {
            let error_msg = result.error.clone()
                .unwrap_or_else(|| "Graph execution failed".to_string());
            tracker.mark_failed(job_id, error_msg).await;
        }
    }
}

/// Extended queued execution that includes the full graph definition.
/// This replaces QueuedExecution in the scheduler for full async support.
#[derive(Debug, Clone)]
pub struct QueuedGraphExecution {
    pub id: Uuid,
    pub graph: Graph,
    pub input: GraphExecutionInput,
    pub priority: crate::agent_scheduler::agent_scheduler::PriorityLevel,
    pub enqueued_at: Instant,
    pub tenant_id: Option<String>,
}

impl QueuedGraphExecution {
    pub fn new(
        graph: Graph,
        input: GraphExecutionInput,
        priority: crate::agent_scheduler::agent_scheduler::PriorityLevel,
        tenant_id: Option<String>,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            graph,
            input,
            priority,
            enqueued_at: Instant::now(),
            tenant_id,
        }
    }
}

/// Execute a queued graph execution to completion.
#[instrument(skip(exec, sar, cost_attr), fields(job_id = %exec.id, graph_id = %exec.graph.id))]
pub async fn execute_queued_graph(
    exec: QueuedGraphExecution,
    sar: Option<Arc<SarNodeExecutor>>,
    cost_attr: Option<Arc<CostAttributor>>,
) -> GraphExecutionResult {
    let ctx = Arc::new(ExecutionContext::with_graph_id(
        exec.id,
        exec.graph.id,
        exec.tenant_id.clone(),
    ));

    // Use the SAR executor if available, otherwise use default stubs
    // Enable cost attribution when CostAttributor is provided (Phase 6: Observability)
    if let Some(ref sar) = sar {
        let executor = if let Some(ref cost) = cost_attr {
            GraphExecutor::with_cost_attributor(sar.as_ref(), cost.clone())
        } else {
            GraphExecutor::new(sar.as_ref())
        };
        executor.execute(&exec.graph, exec.input, ctx).await
    } else {
        let executor = GraphExecutor::new(DefaultNodeExecutor::new());
        executor.execute(&exec.graph, exec.input, ctx).await
    }
}
