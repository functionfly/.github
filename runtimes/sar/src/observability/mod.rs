use std::collections::HashMap;
use std::sync::Arc;
use std::time::Instant;

use parking_lot::RwLock;
use prometheus::{Encoder, IntCounter, IntGauge, Registry, TextEncoder, Opts};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::core::AgentId;
use crate::engine::{ExecutionStatus, NodeId};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TraceEvent {
    pub execution_id: Uuid,
    pub agent_id: AgentId,
    pub node_id: NodeId,
    pub event_type: TraceEventType,
    pub timestamp: chrono::DateTime<chrono::Utc>,
    pub metadata: HashMap<String, String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum TraceEventType {
    NodeStarted,
    NodeCompleted,
    NodeFailed,
    AgentStarted,
    AgentCompleted,
    AgentFailed,
    EventReceived,
    MemoryRead,
    MemoryWrite,
}

#[derive(Debug, Clone)]
pub struct ExecutionTrace {
    pub execution_id: Uuid,
    pub agent_id: AgentId,
    pub events: Vec<TraceEvent>,
    pub start_time: Instant,
    pub end_time: Option<Instant>,
}

impl ExecutionTrace {
    pub fn new(execution_id: Uuid, agent_id: AgentId) -> Self {
        Self {
            execution_id,
            agent_id,
            events: Vec::new(),
            start_time: Instant::now(),
            end_time: None,
        }
    }

    pub fn add_event(&mut self, event: TraceEvent) {
        self.events.push(event);
    }

    pub fn finish(&mut self) {
        self.end_time = Some(Instant::now());
    }

    pub fn duration_ms(&self) -> Option<u64> {
        self.end_time.map(|e| e.duration_since(self.start_time).as_millis() as u64)
    }
}

pub struct CostAttributor {
    agent_costs: Arc<RwLock<HashMap<AgentId, AgentCost>>>,
}

#[derive(Debug, Clone, Default)]
pub struct AgentCost {
    pub total_cost_usd: f64,
    pub llm_cost_usd: f64,
    pub tool_cost_usd: f64,
    pub compute_cost_usd: f64,
    pub last_updated: chrono::DateTime<chrono::Utc>,
}

impl CostAttributor {
    pub fn new() -> Self {
        Self {
            agent_costs: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub fn record_node(
        &self,
        agent_id: AgentId,
        node_type: &str,
        prompt_tokens: u32,
        completion_tokens: u32,
        duration_ms: u64,
    ) {
        let mut costs = self.agent_costs.write();
        let cost = costs.entry(agent_id).or_insert_with(AgentCost::default);

        let node_cost = match node_type {
            "llm" => {
                let c = (prompt_tokens as f64 * 0.0000015) + (completion_tokens as f64 * 0.000002);
                cost.llm_cost_usd += c;
                c
            }
            "tool" => {
                let c = duration_ms as f64 * 0.000001;
                cost.tool_cost_usd += c;
                c
            }
            _ => 0.0,
        };

        cost.compute_cost_usd += duration_ms as f64 * 0.0000001;
        cost.total_cost_usd += node_cost;
        cost.last_updated = chrono::Utc::now();
    }

    pub fn get_cost(&self, agent_id: AgentId) -> AgentCost {
        self.agent_costs.read().get(&agent_id).cloned().unwrap_or_default()
    }

    pub fn get_all_costs(&self) -> HashMap<AgentId, AgentCost> {
        self.agent_costs.read().clone()
    }
}

impl Default for CostAttributor {
    fn default() -> Self {
        Self::new()
    }
}

pub struct MetricsCollector {
    registry: Registry,
    agents_running: IntGauge,
    agents_total: IntGauge,
    executions_completed: IntCounter,
    executions_failed: IntCounter,
    queue_depth: IntGauge,
}

impl MetricsCollector {
    pub fn new() -> Self {
        let registry = Registry::new();

        let agents_running = IntGauge::with_opts(Opts::new("sar_agents_running", "Number of running agents")).unwrap();
        let agents_total = IntGauge::with_opts(Opts::new("sar_agents_total", "Total registered agents")).unwrap();
        let executions_completed = IntCounter::with_opts(Opts::new("sar_executions_completed", "Total completed executions")).unwrap();
        let executions_failed = IntCounter::with_opts(Opts::new("sar_executions_failed", "Total failed executions")).unwrap();
        let queue_depth = IntGauge::with_opts(Opts::new("sar_queue_depth", "Current queue depth")).unwrap();

        registry.register(Box::new(agents_running.clone())).ok();
        registry.register(Box::new(agents_total.clone())).ok();
        registry.register(Box::new(executions_completed.clone())).ok();
        registry.register(Box::new(executions_failed.clone())).ok();
        registry.register(Box::new(queue_depth.clone())).ok();

        Self {
            registry,
            agents_running,
            agents_total,
            executions_completed,
            executions_failed,
            queue_depth,
        }
    }

    pub fn set_agents_running(&self, count: i64) {
        self.agents_running.set(count);
    }

    pub fn set_agents_total(&self, count: i64) {
        self.agents_total.set(count);
    }

    pub fn record_execution(&self, status: ExecutionStatus) {
        match status {
            ExecutionStatus::Completed => self.executions_completed.inc(),
            ExecutionStatus::Failed => self.executions_failed.inc(),
            _ => {}
        }
    }

    pub fn set_queue_depth(&self, depth: i64) {
        self.queue_depth.set(depth);
    }

    pub fn render(&self) -> String {
        let encoder = TextEncoder::new();
        let mut buffer = Vec::new();
        encoder.encode(&self.registry.gather(), &mut buffer).ok();
        String::from_utf8(buffer).unwrap_or_default()
    }
}

impl Default for MetricsCollector {
    fn default() -> Self {
        Self::new()
    }
}

pub struct SelfOptimizationEngine {
    success_rates: Arc<RwLock<HashMap<AgentId, f64>>>,
    cost_history: Arc<RwLock<HashMap<AgentId, Vec<f64>>>>,
    latency_history: Arc<RwLock<HashMap<AgentId, Vec<u64>>>>,
}

impl SelfOptimizationEngine {
    pub fn new() -> Self {
        Self {
            success_rates: Arc::new(RwLock::new(HashMap::new())),
            cost_history: Arc::new(RwLock::new(HashMap::new())),
            latency_history: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub fn record_execution(&self, agent_id: AgentId, success: bool, cost_usd: f64, latency_ms: u64) {
        let mut rates = self.success_rates.write();
        let current = rates.get(&agent_id).copied().unwrap_or(1.0);
        let n = rates.get(&agent_id).map(|_| 100u64).unwrap_or(1);
        let new_rate = if success {
            current + (1.0 - current) / n as f64
        } else {
            current - current / n as f64
        };
        rates.insert(agent_id, new_rate);

        self.cost_history.write().entry(agent_id).or_default().push(cost_usd);
        self.latency_history.write().entry(agent_id).or_default().push(latency_ms);
    }

    pub fn suggest_optimizations(&self, agent_id: AgentId) -> Vec<OptimizationSuggestion> {
        let mut suggestions = Vec::new();

        let rate = self.success_rates.read().get(&agent_id).copied().unwrap_or(1.0);
        if rate < 0.95 {
            suggestions.push(OptimizationSuggestion {
                agent_id,
                suggestion_type: SuggestionType::IncreaseRetries,
                reason: format!("Success rate {:.2}% below 95%", rate * 100.0),
                estimated_impact: "high".to_string(),
            });
        }

        let avg_cost = {
            let costs = self.cost_history.read();
            costs.get(&agent_id)
                .map(|v| if v.is_empty() { 0.0 } else { v.iter().sum::<f64>() / v.len() as f64 })
                .unwrap_or(0.0)
        };

        if avg_cost > 0.10 {
            suggestions.push(OptimizationSuggestion {
                agent_id,
                suggestion_type: SuggestionType::SwitchToCheaperModel,
                reason: format!("Average cost ${:.4} per execution above threshold", avg_cost),
                estimated_impact: "medium".to_string(),
            });
        }

        suggestions
    }
}

impl Default for SelfOptimizationEngine {
    fn default() -> Self {
        Self::new()
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OptimizationSuggestion {
    pub agent_id: AgentId,
    pub suggestion_type: SuggestionType,
    pub reason: String,
    pub estimated_impact: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SuggestionType {
    IncreaseRetries,
    SwitchToCheaperModel,
    SimplifyGraph,
    EnableCaching,
    ReduceTimeout,
}

#[cfg(test)]
mod tests {
    use super::*;
    use uuid::Uuid;

    #[test]
    fn test_cost_attributor_record_and_get() {
        let ca = CostAttributor::new();
        let agent_id = AgentId(Uuid::new_v4());

        ca.record_node(agent_id, "llm", 1000, 500, 200);
        let cost = ca.get_cost(agent_id);
        assert!(cost.llm_cost_usd > 0.0);
        assert!(cost.total_cost_usd > 0.0);
    }

    #[test]
    fn test_cost_attributor_tool_cost() {
        let ca = CostAttributor::new();
        let agent_id = AgentId(Uuid::new_v4());

        ca.record_node(agent_id, "tool", 0, 0, 500);
        let cost = ca.get_cost(agent_id);
        assert!(cost.tool_cost_usd > 0.0);
    }

    #[test]
    fn test_metrics_collector_render() {
        let mc = MetricsCollector::new();
        mc.set_agents_running(5);
        mc.set_agents_total(10);
        mc.record_execution(ExecutionStatus::Completed);
        mc.record_execution(ExecutionStatus::Failed);

        let output = mc.render();
        assert!(output.contains("sar_agents_running"));
        assert!(output.contains("sar_agents_total"));
        assert!(output.contains("sar_executions_completed"));
    }

    #[test]
    fn test_self_optimization_engine() {
        let engine = SelfOptimizationEngine::new();
        let agent_id = AgentId(Uuid::new_v4());

        // Record some successful executions
        for _ in 0..10 {
            engine.record_execution(agent_id, true, 0.01, 50);
        }

        let suggestions = engine.suggest_optimizations(agent_id);
        assert!(suggestions.is_empty()); // all successful, no suggestions

        // Record failures to drop success rate
        for _ in 0..10 {
            engine.record_execution(agent_id, false, 0.01, 50);
        }

        let suggestions = engine.suggest_optimizations(agent_id);
        assert!(!suggestions.is_empty());
    }

    #[test]
    fn test_self_optimization_high_cost() {
        let engine = SelfOptimizationEngine::new();
        let agent_id = AgentId(Uuid::new_v4());

        engine.record_execution(agent_id, true, 0.50, 50);

        let suggestions = engine.suggest_optimizations(agent_id);
        let has_cost_suggestion = suggestions.iter().any(|s| {
            matches!(s.suggestion_type, SuggestionType::SwitchToCheaperModel)
        });
        assert!(has_cost_suggestion);
    }

    #[test]
    fn test_execution_trace() {
        let agent_id = AgentId(Uuid::new_v4());
        let exec_id = Uuid::new_v4();
        let mut trace = ExecutionTrace::new(exec_id, agent_id);

        trace.add_event(TraceEvent {
            execution_id: exec_id,
            agent_id,
            node_id: crate::engine::NodeId(Uuid::new_v4()),
            event_type: TraceEventType::NodeStarted,
            timestamp: chrono::Utc::now(),
            metadata: HashMap::new(),
        });

        assert_eq!(trace.events.len(), 1);

        trace.finish();
        assert!(trace.duration_ms().is_some());
    }
}
