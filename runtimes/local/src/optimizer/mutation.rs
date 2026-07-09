//! Graph mutation and optimization application.
//!
//! Implements the actual graph rewriting operations that the optimizer suggests.
//! Unlike the detection phase (which only *suggests* changes), this module applies
//! them to graph structures.
//!
//! ## Mutation Types
//!
//! | Mutation | Description | Confidence Required |
//! |----------|-------------|---------------------|
//! | `AdjustTimeout` | Update node's timeout_ms | Medium (70%) |
//! | `ConsolidateNodes` | Merge sequential nodes | High (90%) |
//! | `EnableCaching` | Add cache node before slow node | Medium (70%) |
//! | `RemoveDeadBranch` | Eliminate unreachable nodes | High (85%) |
//! | `AddRetryPolicy` | Adjust retry for flaky nodes | Medium (70%) |
//! | `ParallelizeFanOut` | Convert sequential to parallel | High (90%) |
//!
//! ## Safety
//!
//! All mutations are recorded in an audit log and can be rolled back.
//! Production mutations require manual approval via the Go backend.

use std::collections::HashMap;

use strsim::{jaro_winkler, normalized_levenshtein};
use tracing::{debug, info, instrument, warn};
use uuid::Uuid;

use crate::engine::graph::{
    ControlKind, Edge, EdgeType, Expr, Graph, Node, NodeId, NodeType, OptStrategy, RetryPolicy,
};
use crate::optimizer::optimizer::{OptimizationAction, OptimizationSuggestion, PatternConfidence};

/// A mutation applied to a graph.
#[derive(Debug, Clone)]
pub struct AppliedMutation {
    /// Unique ID for this mutation.
    pub id: Uuid,
    /// The optimization that was applied.
    pub suggestion_id: Uuid,
    /// Type of mutation performed.
    pub mutation_type: MutationType,
    /// Nodes affected by this mutation.
    pub affected_nodes: Vec<NodeId>,
    /// Human-readable description.
    pub description: String,
    /// Whether the mutation was applied to a canary graph (true) or production (false).
    pub is_canary: bool,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum MutationType {
    AdjustTimeout,
    ConsolidateNodes,
    EnableCaching,
    RemoveDeadBranch,
    AddRetryPolicy,
    ParallelizeFanOut,
    ModelSwitch,
}

/// Result of applying a mutation.
#[derive(Debug, Clone)]
pub struct MutationResult {
    pub success: bool,
    pub mutation_id: Uuid,
    pub applied_changes: Vec<String>,
    pub errors: Vec<String>,
    pub rollback_possible: bool,
    pub rollback_data: Option<GraphBackup>,
}

/// Backup of a graph before mutation (for rollback).
#[derive(Debug, Clone)]
pub struct GraphBackup {
    pub graph_id: Uuid,
    pub nodes: HashMap<NodeId, Node>,
    pub edges: Vec<Edge>,
    pub timestamp: chrono::DateTime<chrono::Utc>,
}

impl GraphBackup {
    /// Create a backup of a graph.
    pub fn new(graph: &Graph) -> Self {
        Self {
            graph_id: graph.id,
            nodes: graph.nodes.clone(),
            edges: graph.edges.clone(),
            timestamp: chrono::Utc::now(),
        }
    }

    /// Restore a graph to this backup state.
    pub fn restore(&self, graph: &mut Graph) {
        graph.nodes = self.nodes.clone();
        graph.edges = self.edges.clone();
    }
}

/// Graph mutator — applies optimization suggestions to graphs.
pub struct GraphMutator {
    /// Minimum confidence required for each mutation type.
    confidence_thresholds: HashMap<MutationType, PatternConfidence>,
    /// Audit log of applied mutations.
    mutation_log: Vec<AppliedMutation>,
}

impl Default for GraphMutator {
    fn default() -> Self {
        Self::new()
    }
}

impl GraphMutator {
    /// Create a new graph mutator with default confidence thresholds.
    pub fn new() -> Self {
        let mut thresholds = HashMap::new();
        thresholds.insert(MutationType::AdjustTimeout, PatternConfidence::Medium);
        thresholds.insert(MutationType::ConsolidateNodes, PatternConfidence::High);
        thresholds.insert(MutationType::EnableCaching, PatternConfidence::Medium);
        thresholds.insert(MutationType::RemoveDeadBranch, PatternConfidence::High);
        thresholds.insert(MutationType::AddRetryPolicy, PatternConfidence::Medium);
        thresholds.insert(MutationType::ParallelizeFanOut, PatternConfidence::High);
        thresholds.insert(MutationType::ModelSwitch, PatternConfidence::High);

        Self {
            confidence_thresholds: thresholds,
            mutation_log: Vec::new(),
        }
    }

    /// Set a custom confidence threshold for a mutation type.
    pub fn set_confidence_threshold(
        &mut self,
        mutation_type: MutationType,
        threshold: PatternConfidence,
    ) {
        self.confidence_thresholds.insert(mutation_type, threshold);
    }

    /// Check if confidence is sufficient for a mutation type.
    fn has_sufficient_confidence(
        &self,
        mutation_type: MutationType,
        confidence: PatternConfidence,
    ) -> bool {
        let required = self
            .confidence_thresholds
            .get(&mutation_type)
            .copied()
            .unwrap_or(PatternConfidence::High);
        confidence >= required
    }

    /// Apply a suggestion to a graph.
    ///
    /// Creates a backup before applying so rollback is possible.
    #[instrument(skip_all, fields(suggestion_id = %suggestion.id, graph_id = %graph.id))]
    pub fn apply_suggestion(
        &mut self,
        graph: &mut Graph,
        suggestion: &OptimizationSuggestion,
        is_canary: bool,
    ) -> MutationResult {
        let backup = GraphBackup::new(graph);
        let mutation_id = Uuid::new_v4();

        let result = match &suggestion.action {
            OptimizationAction::AdjustTimeout { current_ms, new_ms } => self
                .apply_timeout_adjustment(
                    graph,
                    &suggestion.node_id,
                    *current_ms,
                    *new_ms,
                    suggestion.confidence,
                ),
            OptimizationAction::EnableCaching => {
                self.apply_caching_node(graph, &suggestion.node_id, suggestion.confidence)
            }
            OptimizationAction::ModelDowngrade {
                current_model,
                suggested_model,
            } => self.apply_model_switch(
                graph,
                &suggestion.node_id,
                current_model,
                suggested_model,
                suggestion.confidence,
            ),
            OptimizationAction::SimplifyPath { remove_node } => {
                self.apply_path_simplification(graph, remove_node, suggestion.confidence)
            }
            OptimizationAction::AdjustRetry {
                max_attempts,
                backoff_multiplier,
            } => self.apply_retry_adjustment(
                graph,
                &suggestion.node_id,
                *max_attempts,
                *backoff_multiplier,
                suggestion.confidence,
            ),
            OptimizationAction::IncreaseQuota => {
                // Quota changes happen at tenant level, not graph level
                Ok(vec![
                    "Quota increase requested (tenant-level change)".to_string()
                ])
            }
        };

        match result {
            Ok(changes) => {
                let mutation = AppliedMutation {
                    id: mutation_id,
                    suggestion_id: suggestion.id,
                    mutation_type: self.mutation_type_for_action(&suggestion.action),
                    affected_nodes: vec![NodeId(
                        Uuid::parse_str(&suggestion.node_id).unwrap_or_else(|_| Uuid::nil()),
                    )],
                    description: format!("Applied: {}", suggestion.action),
                    is_canary,
                };

                self.mutation_log.push(mutation);

                info!(
                    mutation_id = %mutation_id,
                    changes = changes.len(),
                    "Graph mutation applied successfully"
                );

                MutationResult {
                    success: true,
                    mutation_id,
                    applied_changes: changes,
                    errors: Vec::new(),
                    rollback_possible: true,
                    rollback_data: Some(backup),
                }
            }
            Err(errors) => {
                warn!(
                    mutation_id = %mutation_id,
                    errors = errors.len(),
                    "Graph mutation failed"
                );

                MutationResult {
                    success: false,
                    mutation_id,
                    applied_changes: Vec::new(),
                    errors,
                    rollback_possible: true,
                    rollback_data: Some(backup),
                }
            }
        }
    }

    /// Apply timeout adjustment to a node.
    fn apply_timeout_adjustment(
        &self,
        graph: &mut Graph,
        node_id_str: &str,
        _current_ms: u64,
        new_ms: u64,
        confidence: PatternConfidence,
    ) -> Result<Vec<String>, Vec<String>> {
        if !self.has_sufficient_confidence(MutationType::AdjustTimeout, confidence) {
            return Err(vec![format!(
                "Confidence {:?} below threshold for timeout adjustment",
                confidence
            )]);
        }

        let node_id = NodeId(
            Uuid::parse_str(node_id_str).map_err(|e| vec![format!("Invalid node ID: {}", e)])?,
        );

        if let Some(node) = graph.nodes.get_mut(&node_id) {
            let old_timeout = node.timeout_ms;
            node.timeout_ms = new_ms;

            Ok(vec![format!(
                "Adjusted timeout for node {}: {}ms -> {}ms",
                node.name, old_timeout, new_ms
            )])
        } else {
            Err(vec![format!("Node {} not found in graph", node_id_str)])
        }
    }

    /// Apply caching optimization — adds a cache lookup before expensive nodes.
    fn apply_caching_node(
        &self,
        graph: &mut Graph,
        node_id_str: &str,
        confidence: PatternConfidence,
    ) -> Result<Vec<String>, Vec<String>> {
        if !self.has_sufficient_confidence(MutationType::EnableCaching, confidence) {
            return Err(vec![format!(
                "Confidence {:?} below threshold for caching",
                confidence
            )]);
        }

        let node_id = NodeId(
            Uuid::parse_str(node_id_str).map_err(|e| vec![format!("Invalid node ID: {}", e)])?,
        );

        // Find the target node and clone its name before modifying the graph
        let target_node_name = graph
            .nodes
            .get(&node_id)
            .ok_or_else(|| vec![format!("Node {} not found", node_id_str)])?
            .name
            .clone();

        // Create a cache lookup node
        let cache_node_id = NodeId(Uuid::new_v4());
        let cache_node = Node {
            id: cache_node_id,
            name: format!("cache_lookup_{}", target_node_name),
            node_type: NodeType::Memory {
                operation: crate::engine::graph::MemoryOp::Read,
                key: format!("cache:{}", node_id.0),
            },
            timeout_ms: 1000, // Fast cache timeout
            retry: RetryPolicy::no_retries(),
            input_schema: None,
            output_schema: None,
            metadata: HashMap::from([
                ("auto_generated".to_string(), "true".to_string()),
                ("cache_for".to_string(), node_id_str.to_string()),
            ]),
        };

        // Find upstream nodes of the target
        let upstream_edges: Vec<Edge> = graph
            .edges
            .iter()
            .filter(|e| e.target == node_id)
            .cloned()
            .collect();

        // Add the cache node
        graph.add_node(cache_node);

        // Redirect upstream edges to the cache node
        for mut edge in upstream_edges {
            edge.target = cache_node_id;
            // Find and update the edge in the graph
            if let Some(e) = graph
                .edges
                .iter_mut()
                .find(|e| e.source == edge.source && e.target == node_id)
            {
                e.target = cache_node_id;
            }
        }

        // Add cache -> target edge
        graph.add_edge(Edge::dataflow(cache_node_id, node_id));

        // Add cache miss handler (target node writes to cache)
        let write_cache_node_id = NodeId(Uuid::new_v4());
        let write_cache_node = Node {
            id: write_cache_node_id,
            name: format!("cache_write_{}", target_node_name),
            node_type: NodeType::Memory {
                operation: crate::engine::graph::MemoryOp::Write,
                key: format!("cache:{}", node_id.0),
            },
            timeout_ms: 1000,
            retry: RetryPolicy::no_retries(),
            input_schema: None,
            output_schema: None,
            metadata: HashMap::from([
                ("auto_generated".to_string(), "true".to_string()),
                ("cache_for".to_string(), node_id_str.to_string()),
            ]),
        };

        graph.add_node(write_cache_node);

        // Add target -> write_cache edge
        graph.add_edge(Edge::dataflow(node_id, write_cache_node_id));

        Ok(vec![
            format!(
                "Added cache lookup node {} before {}",
                cache_node_id, node_id_str
            ),
            format!(
                "Added cache write node {} after {}",
                write_cache_node_id, node_id_str
            ),
        ])
    }

    /// Apply model switching for cost/performance optimization.
    fn apply_model_switch(
        &self,
        graph: &mut Graph,
        node_id_str: &str,
        _current_model: &str,
        suggested_model: &str,
        confidence: PatternConfidence,
    ) -> Result<Vec<String>, Vec<String>> {
        if !self.has_sufficient_confidence(MutationType::ModelSwitch, confidence) {
            return Err(vec![format!(
                "Confidence {:?} below threshold for model switch",
                confidence
            )]);
        }

        let node_id = NodeId(
            Uuid::parse_str(node_id_str).map_err(|e| vec![format!("Invalid node ID: {}", e)])?,
        );

        if let Some(node) = graph.nodes.get_mut(&node_id) {
            if let NodeType::LLM { ref mut model, .. } = node.node_type {
                let old_model = model.clone();
                *model = Some(suggested_model.to_string());

                Ok(vec![format!(
                    "Switched model for node {}: {:?} -> {}",
                    node.name, old_model, suggested_model
                )])
            } else {
                Err(vec![format!("Node {} is not an LLM node", node_id_str)])
            }
        } else {
            Err(vec![format!("Node {} not found in graph", node_id_str)])
        }
    }

    /// Apply path simplification — remove unnecessary nodes.
    fn apply_path_simplification(
        &self,
        graph: &mut Graph,
        node_id_str: &str,
        confidence: PatternConfidence,
    ) -> Result<Vec<String>, Vec<String>> {
        if !self.has_sufficient_confidence(MutationType::RemoveDeadBranch, confidence) {
            return Err(vec![format!(
                "Confidence {:?} below threshold for path simplification",
                confidence
            )]);
        }

        let node_id = NodeId(
            Uuid::parse_str(node_id_str).map_err(|e| vec![format!("Invalid node ID: {}", e)])?,
        );

        // Find upstream and downstream edges for this node
        let upstream_sources: Vec<NodeId> = graph
            .edges
            .iter()
            .filter(|e| e.target == node_id)
            .map(|e| e.source)
            .collect();

        let downstream_targets: Vec<NodeId> = graph
            .edges
            .iter()
            .filter(|e| e.source == node_id)
            .map(|e| e.target)
            .collect();

        // Remove the node
        if graph.nodes.remove(&node_id).is_none() {
            return Err(vec![format!("Node {} not found in graph", node_id_str)]);
        }

        // Remove edges connected to this node
        graph
            .edges
            .retain(|e| e.source != node_id && e.target != node_id);

        // Reconnect upstream to downstream (bypass the removed node)
        let mut new_edges = Vec::new();
        for upstream in &upstream_sources {
            for downstream in &downstream_targets {
                if !graph.detect_cycle(*upstream, *downstream) {
                    new_edges.push(Edge::dataflow(*upstream, *downstream));
                }
            }
        }

        for edge in new_edges {
            graph.add_edge(edge);
        }

        Ok(vec![format!(
            "Removed node {} and rewired {} upstream -> {} downstream edges",
            node_id_str,
            upstream_sources.len(),
            downstream_targets.len()
        )])
    }

    /// Apply retry policy adjustment.
    fn apply_retry_adjustment(
        &self,
        graph: &mut Graph,
        node_id_str: &str,
        max_attempts: u32,
        backoff_multiplier: f64,
        confidence: PatternConfidence,
    ) -> Result<Vec<String>, Vec<String>> {
        if !self.has_sufficient_confidence(MutationType::AddRetryPolicy, confidence) {
            return Err(vec![format!(
                "Confidence {:?} below threshold for retry adjustment",
                confidence
            )]);
        }

        let node_id = NodeId(
            Uuid::parse_str(node_id_str).map_err(|e| vec![format!("Invalid node ID: {}", e)])?,
        );

        if let Some(node) = graph.nodes.get_mut(&node_id) {
            let old_retry = node.retry.clone();
            node.retry = RetryPolicy {
                max_attempts,
                initial_delay_ms: old_retry.initial_delay_ms,
                max_delay_ms: old_retry.max_delay_ms,
                backoff_multiplier,
            };

            Ok(vec![format!(
                "Adjusted retry for node {}: attempts={} (was {}), backoff={} (was {})",
                node.name,
                max_attempts,
                old_retry.max_attempts,
                backoff_multiplier,
                old_retry.backoff_multiplier
            )])
        } else {
            Err(vec![format!("Node {} not found in graph", node_id_str)])
        }
    }

    /// Convert an optimization action to its mutation type.
    fn mutation_type_for_action(&self, action: &OptimizationAction) -> MutationType {
        match action {
            OptimizationAction::AdjustTimeout { .. } => MutationType::AdjustTimeout,
            OptimizationAction::EnableCaching => MutationType::EnableCaching,
            OptimizationAction::ModelDowngrade { .. } => MutationType::ModelSwitch,
            OptimizationAction::SimplifyPath { .. } => MutationType::RemoveDeadBranch,
            OptimizationAction::IncreaseQuota => MutationType::AdjustTimeout,
            OptimizationAction::AdjustRetry { .. } => MutationType::AddRetryPolicy,
        }
    }

    /// Rollback a mutation using its backup data.
    pub fn rollback_mutation(
        &mut self,
        graph: &mut Graph,
        result: &MutationResult,
    ) -> Result<(), String> {
        if let Some(backup) = &result.rollback_data {
            backup.restore(graph);
            info!(mutation_id = %result.mutation_id, "Mutation rolled back");
            Ok(())
        } else {
            Err("No rollback data available".to_string())
        }
    }

    /// Get the mutation log.
    pub fn mutation_log(&self) -> &[AppliedMutation] {
        &self.mutation_log
    }

    /// Clear the mutation log.
    pub fn clear_mutation_log(&mut self) {
        self.mutation_log.clear();
    }
}

/// Canary testing for graph mutations.
pub struct CanaryTester {
    /// Maximum executions before comparing results.
    pub max_executions: usize,
    /// Threshold for result similarity (0.0 - 1.0).
    pub similarity_threshold: f64,
}

impl Default for CanaryTester {
    fn default() -> Self {
        Self {
            max_executions: 100,
            similarity_threshold: 0.95,
        }
    }
}

impl CanaryTester {
    /// Compare results from original and mutated graphs.
    ///
    /// Returns true if the mutated graph produces "similar enough" results
    /// to warrant promotion to production.
    pub fn evaluate_results(
        &self,
        original_results: &[serde_json::Value],
        mutated_results: &[serde_json::Value],
    ) -> bool {
        if original_results.len() != mutated_results.len() {
            return false;
        }

        if original_results.is_empty() {
            return false;
        }

        let matches = original_results
            .iter()
            .zip(mutated_results.iter())
            .filter(|(a, b)| self.values_similar(a, b))
            .count();

        let similarity = matches as f64 / original_results.len() as f64;
        similarity >= self.similarity_threshold
    }

const STRING_SIMILARITY_THRESHOLD: f64 = 0.85;

fn strings_similar(a: &str, b: &str) -> bool {
    if a == b {
        return true;
    }
    let jw_score = jaro_winkler(a, b);
    if jw_score >= STRING_SIMILARITY_THRESHOLD {
        return true;
    }
    let lev_score = normalized_levenshtein(a, b);
    lev_score >= STRING_SIMILARITY_THRESHOLD
}

    /// Check if two JSON values are "similar" (for canary evaluation).
    fn values_similar(&self, a: &serde_json::Value, b: &serde_json::Value) -> bool {
        match (a, b) {
            (serde_json::Value::Null, serde_json::Value::Null) => true,
            (serde_json::Value::Bool(a), serde_json::Value::Bool(b)) => a == b,
            (serde_json::Value::Number(a), serde_json::Value::Number(b)) => {
                // Allow small numeric differences (e.g., from model switching)
                if let (Some(a_f), Some(b_f)) = (a.as_f64(), b.as_f64()) {
                    (a_f - b_f).abs() < 0.001
                } else {
                    a == b
                }
            }
            (serde_json::Value::String(a), serde_json::Value::String(b)) => {
                strings_similar(a, b)
            }
            (serde_json::Value::Array(a), serde_json::Value::Array(b)) => {
                if a.len() != b.len() {
                    return false;
                }
                a.iter()
                    .zip(b.iter())
                    .all(|(a, b)| self.values_similar(a, b))
            }
            (serde_json::Value::Object(a), serde_json::Value::Object(b)) => {
                if a.len() != b.len() {
                    return false;
                }
                a.iter().all(|(key, a_val)| {
                    b.get(key)
                        .map_or(false, |b_val| self.values_similar(a_val, b_val))
                })
            }
            _ => false,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_test_graph() -> Graph {
        let a = NodeId(Uuid::new_v4());
        let b = NodeId(Uuid::new_v4());
        let c = NodeId(Uuid::new_v4());

        let mut graph = Graph::new(Uuid::new_v4(), "test".to_string());
        graph.add_node(Node::new(a, "A".to_string(), NodeType::Passthrough));
        graph.add_node(Node::new(b, "B".to_string(), NodeType::Passthrough));
        graph.add_node(Node::new(c, "C".to_string(), NodeType::Passthrough));
        graph.add_edge(Edge::dataflow(a, b));
        graph.add_edge(Edge::dataflow(b, c));

        graph
    }

    #[test]
    fn test_graph_backup_and_restore() {
        let mut graph = make_test_graph();
        let original_node_count = graph.nodes.len();
        let original_edge_count = graph.edges.len();

        let backup = GraphBackup::new(&graph);

        // Modify the graph
        graph.nodes.clear();
        graph.edges.clear();

        assert!(graph.nodes.is_empty());

        // Restore
        backup.restore(&mut graph);

        assert_eq!(graph.nodes.len(), original_node_count);
        assert_eq!(graph.edges.len(), original_edge_count);
    }

    #[test]
    fn test_apply_timeout_adjustment() {
        let mut graph = make_test_graph();
        let node_id = graph.nodes.keys().next().copied().unwrap();
        let node_id_str = node_id.0.to_string();

        let mut mutator = GraphMutator::new();

        let result = mutator.apply_timeout_adjustment(
            &mut graph,
            &node_id_str,
            30_000,
            60_000,
            PatternConfidence::High,
        );

        assert!(result.is_ok());

        let node = graph.nodes.get(&node_id).unwrap();
        assert_eq!(node.timeout_ms, 60_000);
    }

    #[test]
    fn test_apply_timeout_low_confidence() {
        let mut graph = make_test_graph();
        let node_id = graph.nodes.keys().next().copied().unwrap();
        let node_id_str = node_id.0.to_string();

        let mut mutator = GraphMutator::new();

        // Low confidence should fail
        let result = mutator.apply_timeout_adjustment(
            &mut graph,
            &node_id_str,
            30_000,
            60_000,
            PatternConfidence::Low,
        );

        assert!(result.is_err());
    }

    #[test]
    fn test_path_simplification() {
        let mut graph = make_test_graph();
        let nodes: Vec<NodeId> = graph.nodes.keys().copied().collect();
        let b_id = nodes[1]; // Middle node

        let mut mutator = GraphMutator::new();

        let result = mutator.apply_path_simplification(
            &mut graph,
            &b_id.0.to_string(),
            PatternConfidence::High,
        );

        assert!(result.is_ok());
        assert_eq!(graph.nodes.len(), 2); // B removed
    }

    #[test]
    fn test_retry_adjustment() {
        let mut graph = make_test_graph();
        let node_id = graph.nodes.keys().next().copied().unwrap();
        let node_id_str = node_id.0.to_string();

        let mut mutator = GraphMutator::new();

        let result = mutator.apply_retry_adjustment(
            &mut graph,
            &node_id_str,
            5,
            2.5,
            PatternConfidence::High,
        );

        assert!(result.is_ok());

        let node = graph.nodes.get(&node_id).unwrap();
        assert_eq!(node.retry.max_attempts, 5);
        assert_eq!(node.retry.backoff_multiplier, 2.5);
    }

    #[test]
    fn test_canary_similarity() {
        let tester = CanaryTester {
            max_executions: 10,
            similarity_threshold: 0.8,
        };

        let original = vec![
            serde_json::json!({"status": "ok", "data": [1, 2, 3]}),
            serde_json::json!({"status": "ok", "data": [4, 5, 6]}),
        ];

        let mutated = vec![
            serde_json::json!({"status": "ok", "data": [1, 2, 3]}),
            serde_json::json!({"status": "ok", "data": [4, 5, 6]}),
        ];

        assert!(tester.evaluate_results(&original, &mutated));
    }

    #[test]
    fn test_canary_difference() {
        let tester = CanaryTester::default();

        let original = vec![
            serde_json::json!({"status": "ok"}),
            serde_json::json!({"status": "ok"}),
        ];

        let mutated = vec![
            serde_json::json!({"status": "error"}),
            serde_json::json!({"status": "ok"}),
        ];

        assert!(!tester.evaluate_results(&original, &mutated));
    }
}
