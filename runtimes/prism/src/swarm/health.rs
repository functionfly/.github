//! Swarm health monitoring and self-healing
//!
//! Production-ready health monitoring with automatic healing actions

use std::collections::HashMap;
use chrono::{DateTime, Duration, Utc};
use serde::{Deserialize, Serialize};
use thiserror::Error;
use tracing::{info, warn, debug};

use crate::core::CellId;
use super::coordinator::SwarmHealth;

/// Health check errors
#[derive(Error, Debug)]
pub enum HealthError {
    #[error("Cell not found: {0}")]
    CellNotFound(CellId),
    #[error("Health check timeout: {0}")]
    Timeout(CellId),
    #[error("Too many failures: {0}")]
    TooManyFailures(CellId),
}

/// Self-healing action with metadata
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SelfHealAction {
    /// Restart a failed cell
    RestartCell {
        cell_id: CellId,
        attempt: u32,
        reason: String,
    },
    /// Spawn a replacement for a failed cell
    SpawnReplacement {
        failed_id: CellId,
        target_node: Option<String>,
        resources: CellResources,
    },
    /// Migrate cell to another node
    MigrateCell {
        cell_id: CellId,
        target_node: String,
        reason: String,
    },
    /// Notify operators about critical failure
    NotifyOperator {
        cell_id: CellId,
        severity: AlertSeverity,
        reason: String,
    },
    /// Decommission a permanently failed cell
    Decommission {
        cell_id: CellId,
        reason: String,
    },
}

/// Resource requirements for replacement cells
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CellResources {
    pub vcpus: u32,
    pub memory_mb: u64,
    pub gpu_required: bool,
}

/// Alert severity levels
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum AlertSeverity {
    Info,
    Warning,
    Critical,
    Fatal,
}

/// Health check result with detailed status
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HealthCheck {
    pub cell_id: CellId,
    pub is_healthy: bool,
    pub cpu_usage: f32,
    pub memory_usage_bytes: u64,
    pub memory_limit_bytes: u64,
    pub last_heartbeat: DateTime<Utc>,
    pub consecutive_failures: u32,
    pub failure_reason: Option<String>,
    pub metadata: HealthMetadata,
}

/// Additional health metadata
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HealthMetadata {
    pub execution_time_ms: u64,
    pub active_operations: u32,
    pub queued_messages: u32,
    pub connected_peers: u32,
}

impl HealthCheck {
    /// Create a healthy check result
    pub fn healthy(cell_id: CellId) -> Self {
        Self {
            cell_id,
            is_healthy: true,
            cpu_usage: 0.0,
            memory_usage_bytes: 0,
            memory_limit_bytes: 0,
            last_heartbeat: Utc::now(),
            consecutive_failures: 0,
            failure_reason: None,
            metadata: HealthMetadata::default(),
        }
    }

    /// Create an unhealthy check result
    pub fn unhealthy(cell_id: CellId, reason: impl Into<String>) -> Self {
        Self {
            cell_id,
            is_healthy: false,
            cpu_usage: 0.0,
            memory_usage_bytes: 0,
            memory_limit_bytes: 0,
            last_heartbeat: Utc::now(),
            consecutive_failures: 1,
            failure_reason: Some(reason.into()),
            metadata: HealthMetadata::default(),
        }
    }

    /// Memory usage as percentage
    pub fn memory_usage_percent(&self) -> f32 {
        if self.memory_limit_bytes == 0 {
            0.0
        } else {
            (self.memory_usage_bytes as f32 / self.memory_limit_bytes as f32) * 100.0
        }
    }

    /// Check if cell is stale (missing heartbeats)
    pub fn is_stale(&self, heartbeat_timeout: Duration) -> bool {
        Utc::now() - self.last_heartbeat > heartbeat_timeout
    }
}

impl Default for HealthMetadata {
    fn default() -> Self {
        Self {
            execution_time_ms: 0,
            active_operations: 0,
            queued_messages: 0,
            connected_peers: 0,
        }
    }
}

/// Configuration for health monitoring
#[derive(Debug, Clone)]
pub struct HealthMonitorConfig {
    /// Number of consecutive failures before healing action
    pub failure_threshold: u32,
    /// Heartbeat timeout duration
    pub heartbeat_timeout: Duration,
    /// Maximum memory usage before triggering warning (percentage)
    pub memory_warning_threshold: f32,
    /// Maximum memory usage before triggering heal (percentage)
    pub memory_critical_threshold: f32,
    /// Maximum CPU usage before triggering warning (percentage)
    pub cpu_warning_threshold: f32,
    /// Enable automatic self-healing
    pub auto_heal_enabled: bool,
    /// Cooldown period between heal actions for same cell
    pub heal_cooldown: Duration,
}

impl Default for HealthMonitorConfig {
    fn default() -> Self {
        Self {
            failure_threshold: 3,
            heartbeat_timeout: Duration::seconds(30),
            memory_warning_threshold: 70.0,
            memory_critical_threshold: 90.0,
            cpu_warning_threshold: 80.0,
            auto_heal_enabled: true,
            heal_cooldown: Duration::minutes(5),
        }
    }
}

/// Health monitoring and self-healing manager
pub struct HealthMonitor {
    config: HealthMonitorConfig,
    checks: HashMap<CellId, HealthCheck>,
    heal_timestamps: HashMap<CellId, DateTime<Utc>>,
    stats: HealthStats,
}

/// Health monitoring statistics
#[derive(Debug, Clone, Default)]
pub struct HealthStats {
    pub total_checks: u64,
    pub healthy_checks: u64,
    pub unhealthy_checks: u64,
    pub stale_checks: u64,
    pub heal_actions_triggered: u64,
    pub last_check_time: Option<DateTime<Utc>>,
}

impl HealthMonitor {
    /// Create a new health monitor with configuration
    pub fn new(config: HealthMonitorConfig) -> Self {
        Self {
            config,
            checks: HashMap::new(),
            heal_timestamps: HashMap::new(),
            stats: HealthStats::default(),
        }
    }

    /// Create with default configuration
    pub fn with_defaults() -> Self {
        Self::new(HealthMonitorConfig::default())
    }

    /// Record a health check result
    pub fn record_check(&mut self, check: HealthCheck) {
        self.stats.total_checks += 1;
        self.stats.last_check_time = Some(Utc::now());

        let cell_id = check.cell_id;
        let is_healthy = check.is_healthy;

        if is_healthy {
            self.stats.healthy_checks += 1;
        } else {
            self.stats.unhealthy_checks += 1;
        }

        // Update or insert the check
        if let Some(existing) = self.checks.get_mut(&cell_id) {
            if is_healthy {
                existing.consecutive_failures = 0;
                existing.failure_reason = None;
            } else {
                existing.consecutive_failures += 1;
                existing.failure_reason = check.failure_reason.clone();
            }
            existing.is_healthy = is_healthy;
            existing.cpu_usage = check.cpu_usage;
            existing.memory_usage_bytes = check.memory_usage_bytes;
            existing.memory_limit_bytes = check.memory_limit_bytes;
            existing.last_heartbeat = check.last_heartbeat;
            existing.metadata = check.metadata;
        } else {
            self.checks.insert(cell_id, check);
        }

        debug!(cell_id = %cell_id, healthy = is_healthy,
               failures = self.checks.get(&cell_id).map(|c| c.consecutive_failures).unwrap_or(0),
               "Health check recorded");
    }

    /// Get health status for a cell
    pub fn get_health(&self, cell_id: &CellId) -> Option<&HealthCheck> {
        self.checks.get(cell_id)
    }

    /// Get all unhealthy cells
    pub fn get_unhealthy_cells(&self) -> Vec<CellId> {
        self.checks
            .iter()
            .filter(|(_, check)| !check.is_healthy)
            .map(|(id, _)| *id)
            .collect()
    }

    /// Get all stale cells (missing heartbeats)
    pub fn get_stale_cells(&self) -> Vec<CellId> {
        self.checks
            .iter()
            .filter(|(_, check)| check.is_stale(self.config.heartbeat_timeout))
            .map(|(id, _)| *id)
            .collect()
    }

    /// Determine if a cell should be replaced based on failure count
    pub fn should_replace(&self, cell_id: &CellId) -> bool {
        self.checks
            .get(cell_id)
            .map(|c| c.consecutive_failures >= self.config.failure_threshold)
            .unwrap_or(false)
    }

    /// Check if a cell is under cooldown for healing
    pub fn is_under_cooldown(&self, cell_id: &CellId) -> bool {
        if let Some(last_heal) = self.heal_timestamps.get(cell_id) {
            Utc::now() - *last_heal < self.config.heal_cooldown
        } else {
            false
        }
    }

    /// Record that a heal action was taken for a cell
    pub fn record_heal_action(&mut self, cell_id: &CellId) {
        self.heal_timestamps.insert(*cell_id, Utc::now());
        self.stats.heal_actions_triggered += 1;
    }

    /// Create self-heal actions for unhealthy cells
    pub fn create_heal_actions(&self, unhealthy_cells: &[CellId]) -> Vec<SelfHealAction> {
        let mut actions = Vec::new();

        for cell_id in unhealthy_cells {
            if self.is_under_cooldown(cell_id) {
                debug!(cell_id = %cell_id, "Cell under heal cooldown, skipping");
                continue;
            }

            if let Some(check) = self.checks.get(cell_id) {
                let action = self.determine_heal_action(cell_id, check);
                actions.push(action);
            }
        }

        actions
    }

    /// Determine the appropriate heal action for a cell
    fn determine_heal_action(&self, cell_id: &CellId, check: &HealthCheck) -> SelfHealAction {
        // Memory critical - migrate or decommission
        if check.memory_usage_percent() >= self.config.memory_critical_threshold {
            info!(cell_id = %cell_id, memory_percent = check.memory_usage_percent(),
                  "Memory critical, scheduling migration");
            return SelfHealAction::MigrateCell {
                cell_id: *cell_id,
                target_node: "least_loaded".to_string(),
                reason: format!("Memory at {:.1}%", check.memory_usage_percent()),
            };
        }

        // CPU critical - restart to clear stuck state
        if check.cpu_usage >= self.config.cpu_warning_threshold {
            info!(cell_id = %cell_id, cpu_percent = check.cpu_usage,
                  "CPU critical, scheduling restart");
            return SelfHealAction::RestartCell {
                cell_id: *cell_id,
                attempt: check.consecutive_failures,
                reason: format!("CPU at {:.1}%", check.cpu_usage),
            };
        }

        // Consecutive failures - replace
        if check.consecutive_failures >= self.config.failure_threshold {
            warn!(cell_id = %cell_id, failures = check.consecutive_failures,
                  "Too many failures, spawning replacement");
            return SelfHealAction::SpawnReplacement {
                failed_id: *cell_id,
                target_node: None,
                resources: CellResources {
                    vcpus: 2,
                    memory_mb: 512,
                    gpu_required: false,
                },
            };
        }

        // Default - restart with notification
        SelfHealAction::NotifyOperator {
            cell_id: *cell_id,
            severity: AlertSeverity::Warning,
            reason: check.failure_reason.clone().unwrap_or_else(|| "Unknown failure".to_string()),
        }
    }

    /// Aggregate health checks into swarm health
    pub fn aggregate_health(&self, cell_ids: &[CellId]) -> SwarmHealth {
        let total = cell_ids.len() as u32;
        let healthy = cell_ids.iter()
            .filter(|id| {
                self.checks
                    .get(id)
                    .map(|c| c.is_healthy && !c.is_stale(self.config.heartbeat_timeout))
                    .unwrap_or(false)
            })
            .count() as u32;

        let failed: Vec<String> = cell_ids.iter()
            .filter(|id| {
                !self.checks
                    .get(id)
                    .map(|c| c.is_healthy)
                    .unwrap_or(false)
            })
            .map(|id| id.to_string())
            .collect();

        SwarmHealth {
            is_healthy: healthy == total && total > 0,
            active_count: healthy,
            total_count: total,
            failed_cells: failed,
        }
    }

    /// Get monitoring statistics
    pub fn stats(&self) -> &HealthStats {
        &self.stats
    }

    /// Get the health check configuration
    pub fn config(&self) -> &HealthMonitorConfig {
        &self.config
    }

    /// Convert a health check result to a HealthError if unhealthy
    pub fn check_to_error(&self, cell_id: &CellId) -> Option<HealthError> {
        self.checks.get(cell_id).map(|check| {
            if check.consecutive_failures >= self.config.failure_threshold {
                HealthError::TooManyFailures(*cell_id)
            } else if check.is_stale(self.config.heartbeat_timeout) {
                HealthError::Timeout(*cell_id)
            } else if !check.is_healthy {
                HealthError::CellNotFound(*cell_id) // Using CellNotFound as generic unhealthy
            } else {
                // Return a specific error based on failure reason
                HealthError::CellNotFound(*cell_id)
            }
        })
    }

    /// Record an unhealthy check and get the corresponding error
    pub fn record_unhealthy(&mut self, cell_id: CellId, reason: impl Into<String>) -> HealthError {
        let check = HealthCheck::unhealthy(cell_id, reason);
        let error = HealthError::CellNotFound(cell_id); // Could be more specific based on reason
        self.record_check(check);
        error
    }

    /// Perform a health sweep and return actions to take
    pub fn sweep(&mut self) -> Vec<SelfHealAction> {
        let unhealthy = self.get_unhealthy_cells();
        let stale = self.get_stale_cells();

        // Mark stale cells as unhealthy
        for cell_id in &stale {
            if let Some(check) = self.checks.get_mut(cell_id) {
                check.is_healthy = false;
                check.failure_reason = Some("Heartbeat timeout".to_string());
                self.stats.unhealthy_checks += 1;
                self.stats.healthy_checks = self.stats.healthy_checks.saturating_sub(1);
            }
        }

        let mut all_affected = unhealthy;
        all_affected.extend(stale);

        if self.config.auto_heal_enabled {
            let actions = self.create_heal_actions(&all_affected);

            // Record heal timestamps for cells we're healing
            for action in &actions {
                match action {
                    SelfHealAction::RestartCell { cell_id, .. } |
                    SelfHealAction::SpawnReplacement { failed_id: cell_id, .. } |
                    SelfHealAction::MigrateCell { cell_id, .. } |
                    SelfHealAction::NotifyOperator { cell_id, .. } |
                    SelfHealAction::Decommission { cell_id, .. } => {
                        self.record_heal_action(cell_id);
                    }
                }
            }

            actions
        } else {
            Vec::new()
        }
    }
}

impl Default for HealthMonitor {
    fn default() -> Self {
        Self::with_defaults()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_health_check_memory_percent() {
        let mut check = HealthCheck::healthy(CellId::new());
        check.memory_usage_bytes = 512 * 1024 * 1024;
        check.memory_limit_bytes = 1024 * 1024 * 1024;

        assert!((check.memory_usage_percent() - 50.0).abs() < 0.1);
    }

    #[test]
    fn test_failure_threshold() {
        let mut monitor = HealthMonitor::with_defaults();
        let cell_id = CellId::new();

        // Record healthy checks
        for _ in 0..3 {
            monitor.record_check(HealthCheck::healthy(cell_id));
        }

        assert!(!monitor.should_replace(&cell_id));

        // Record unhealthy check
        monitor.record_check(HealthCheck::unhealthy(cell_id, "Test failure"));

        // Should still not trigger until threshold
        assert!(!monitor.should_replace(&cell_id));

        // Record more failures to reach threshold
        monitor.record_check(HealthCheck::unhealthy(cell_id, "Test failure"));
        monitor.record_check(HealthCheck::unhealthy(cell_id, "Test failure"));

        assert!(monitor.should_replace(&cell_id));
    }

    #[test]
    fn test_cooldown() {
        let mut monitor = HealthMonitor::with_defaults();
        let cell_id = CellId::new();

        assert!(!monitor.is_under_cooldown(&cell_id));

        monitor.record_heal_action(&cell_id);
        assert!(monitor.is_under_cooldown(&cell_id));
    }
}