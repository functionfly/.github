use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};
use parking_lot::RwLock;
use tokio::sync::mpsc;
use tracing::{info, warn};
use chrono::{DateTime, Utc};

use crate::core::{AgentId, AgentStatus};
use crate::persistence::CachedAgentRegistry;

const DEFAULT_HEARTBEAT_TIMEOUT_SECS: u64 = 120;
const DEFAULT_ORPHAN_CHECK_INTERVAL_SECS: u64 = 60;
const DEFAULT_GRACEFUL_SHUTDOWN_TIMEOUT_SECS: u64 = 30;

#[derive(Debug, Clone)]
pub struct AgentLifecycleState {
    pub agent_id: AgentId,
    pub status: AgentStatus,
    pub last_heartbeat: Option<DateTime<Utc>>,
    pub in_flight_executions: u32,
    pub graceful_shutdown_requested: bool,
    pub shutdown_deadline: Option<Instant>,
    pub state_snapshot: HashMap<String, String>,
}

impl AgentLifecycleState {
    pub fn new(agent_id: AgentId) -> Self {
        Self {
            agent_id,
            status: AgentStatus::Idle,
            last_heartbeat: None,
            in_flight_executions: 0,
            graceful_shutdown_requested: false,
            shutdown_deadline: None,
            state_snapshot: HashMap::new(),
        }
    }

    pub fn is_alive(&self) -> bool {
        if self.graceful_shutdown_requested {
            if let Some(deadline) = self.shutdown_deadline {
                return Instant::now() < deadline;
            }
            return false;
        }
        if let Some(last_heartbeat) = self.last_heartbeat {
            let elapsed = Utc::now().signed_duration_since(last_heartbeat);
            elapsed.num_seconds() < DEFAULT_HEARTBEAT_TIMEOUT_SECS as i64
        } else {
            false
        }
    }

    pub fn is_orphaned(&self) -> bool {
        if self.graceful_shutdown_requested {
            return false;
        }
        if let Some(last_heartbeat) = self.last_heartbeat {
            let elapsed = Utc::now().signed_duration_since(last_heartbeat);
            elapsed.num_seconds() >= (DEFAULT_HEARTBEAT_TIMEOUT_SECS * 2) as i64
        } else {
            true
        }
    }
}

pub struct LifecycleManager {
    registry: Arc<CachedAgentRegistry>,
    states: Arc<RwLock<HashMap<AgentId, AgentLifecycleState>>>,
    shutdown_tx: mpsc::Sender<AgentId>,
    stop_ch: Option<mpsc::Sender<()>>,
}

impl LifecycleManager {
    pub fn new(registry: Arc<CachedAgentRegistry>) -> Self {
        let (shutdown_tx, _) = mpsc::channel(100);
        Self {
            registry,
            states: Arc::new(RwLock::new(HashMap::new())),
            shutdown_tx,
            stop_ch: None,
        }
    }

    pub fn register_agent(&self, agent_id: AgentId) {
        let mut states = self.states.write();
        states.insert(agent_id, AgentLifecycleState::new(agent_id));
        info!(agent_id = %agent_id, "Agent registered with lifecycle manager");
    }

    pub fn unregister_agent(&self, agent_id: &AgentId) {
        let mut states = self.states.write();
        states.remove(agent_id);
        info!(agent_id = %agent_id, "Agent unregistered from lifecycle manager");
    }

    pub fn update_heartbeat(&self, agent_id: &AgentId) {
        let mut states = self.states.write();
        if let Some(state) = states.get_mut(agent_id) {
            state.last_heartbeat = Some(Utc::now());
        }
    }

    pub fn request_graceful_shutdown(&self, agent_id: &AgentId, grace_period_secs: u64) -> bool {
        let mut states = self.states.write();
        if let Some(state) = states.get_mut(agent_id) {
            state.graceful_shutdown_requested = true;
            state.shutdown_deadline = Some(Instant::now() + Duration::from_secs(grace_period_secs));
            state.status = AgentStatus::Paused;
            info!(agent_id = %agent_id, grace_period_secs = grace_period_secs, "Graceful shutdown requested");
            return true;
        }
        false
    }

    pub fn check_shutdown_complete(&self, agent_id: &AgentId) -> bool {
        let states = self.states.read();
        if let Some(state) = states.get(agent_id) {
            if state.graceful_shutdown_requested {
                if let Some(deadline) = state.shutdown_deadline {
                    return state.in_flight_executions == 0 || Instant::now() >= deadline;
                }
            }
        }
        false
    }

    pub fn increment_in_flight(&self, agent_id: &AgentId) {
        let mut states = self.states.write();
        if let Some(state) = states.get_mut(agent_id) {
            state.in_flight_executions += 1;
        }
    }

    pub fn decrement_in_flight(&self, agent_id: &AgentId) {
        let mut states = self.states.write();
        if let Some(state) = states.get_mut(agent_id) {
            if state.in_flight_executions > 0 {
                state.in_flight_executions -= 1;
            }
        }
    }

    pub fn get_state(&self, agent_id: &AgentId) -> Option<AgentLifecycleState> {
        let states = self.states.read();
        states.get(agent_id).cloned()
    }

    pub fn list_alive_agents(&self) -> Vec<AgentId> {
        let states = self.states.read();
        states.iter()
            .filter(|(_, state)| state.is_alive())
            .map(|(id, _)| *id)
            .collect()
    }

    pub fn list_orphaned_agents(&self) -> Vec<AgentId> {
        let states = self.states.read();
        states.iter()
            .filter(|(_, state)| state.is_orphaned())
            .map(|(id, _)| *id)
            .collect()
    }

    pub fn save_state_snapshot(&self, agent_id: &AgentId, snapshot: HashMap<String, String>) {
        let mut states = self.states.write();
        if let Some(state) = states.get_mut(agent_id) {
            state.state_snapshot = snapshot;
        }
    }

    pub fn get_state_snapshot(&self, agent_id: &AgentId) -> Option<HashMap<String, String>> {
        let states = self.states.read();
        states.get(agent_id).map(|s| s.state_snapshot.clone())
    }

    pub async fn detect_and_mark_orphans(&self) -> Vec<AgentId> {
        let orphaned: Vec<AgentId> = {
            let states = self.states.read();
            states.iter()
                .filter(|(_, s)| s.is_orphaned())
                .map(|(id, _)| *id)
                .collect()
        };

        for agent_id in &orphaned {
            warn!(agent_id = %agent_id, "Marking agent as orphaned");
            self.registry.update_status(agent_id, AgentStatus::Failed).await.ok();
        }

        orphaned
    }

    pub fn get_stats(&self) -> LifecycleStats {
        let states = self.states.read();
        let mut alive = 0;
        let mut orphaned = 0;
        let mut shutting_down = 0;

        for state in states.values() {
            if state.is_orphaned() {
                orphaned += 1;
            } else if state.graceful_shutdown_requested {
                shutting_down += 1;
            } else if state.is_alive() {
                alive += 1;
            }
        }

        LifecycleStats {
            total_agents: states.len(),
            alive_agents: alive,
            orphaned_agents: orphaned,
            shutting_down_agents: shutting_down,
        }
    }
}

#[derive(Debug, Clone)]
pub struct LifecycleStats {
    pub total_agents: usize,
    pub alive_agents: usize,
    pub orphaned_agents: usize,
    pub shutting_down_agents: usize,
}

pub struct GracefulShutdown {
    agent_id: AgentId,
    in_flight: u32,
    deadline: Instant,
}

impl GracefulShutdown {
    pub fn new(agent_id: AgentId, in_flight: u32, grace_period_secs: u64) -> Self {
        Self {
            agent_id,
            in_flight,
            deadline: Instant::now() + Duration::from_secs(grace_period_secs),
        }
    }

    pub fn is_complete(&self) -> bool {
        self.in_flight == 0 || Instant::now() >= self.deadline
    }

    pub fn time_remaining(&self) -> Option<Duration> {
        if Instant::now() < self.deadline {
            Some(self.deadline - Instant::now())
        } else {
            None
        }
    }
}