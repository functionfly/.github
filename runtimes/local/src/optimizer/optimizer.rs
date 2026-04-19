//! Self-optimization engine for SAR graph execution.
//!
//! Analyzes execution history from `StateGraphMemory` and emits `OptimizationSuggestion`
//! records when patterns cross confidence thresholds. Never auto-mutates production graphs.
//!
//! ## Architecture
//!
//! The optimizer uses the existing `StateGraphMemory::detect_patterns()` method
//! which returns pre-computed patterns. This module wraps those patterns with:
//! - `PatternConfidence` calculation based on sample size and metric value
//! - `OptimizationAction` generation from detected patterns
//! - HTTP emission to the Go backend for review
//!
//! ## Safety
//!
//! This module NEVER automatically mutates a production graph. All optimizations
//! are emitted as `OptimizationSuggestion` records that the Go backend can review
//! and approve before applying.

use std::sync::Arc;
use std::time::{Duration, Instant};

use tokio::sync::RwLock;
use tracing::{debug, info, instrument};

use crate::memory::state::{self, StateGraphMemory};
use crate::optimizer::config::ThresholdConfig;

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

/// Confidence level of a detected pattern.
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
pub enum PatternConfidence {
    Low = 1,
    Medium = 2,
    High = 3,
}

impl PatternConfidence {
    pub fn label(&self) -> &'static str {
        match self {
            PatternConfidence::Low => "low",
            PatternConfidence::Medium => "medium",
            PatternConfidence::High => "high",
        }
    }
}

impl From<f64> for PatternConfidence {
    fn from(confidence: f64) -> Self {
        if confidence >= 0.9 {
            PatternConfidence::High
        } else if confidence >= 0.7 {
            PatternConfidence::Medium
        } else {
            PatternConfidence::Low
        }
    }
}

/// An optimization suggestion emitted to the Go backend for review.
#[derive(Debug, Clone)]
pub struct OptimizationSuggestion {
    pub id: uuid::Uuid,
    pub node_id: String,
    pub node_name: String,
    /// What kind of optimization to apply.
    pub action: OptimizationAction,
    /// Expected improvement as a fraction (e.g., 0.15 = 15% improvement).
    pub expected_impact: Option<f64>,
    /// Implementation difficulty: low, medium, high.
    pub difficulty: &'static str,
    /// Human-readable description of the suggestion.
    pub description: String,
    /// Original graph ID (for audit trail).
    pub graph_id: uuid::Uuid,
    /// The confidence score from pattern detection.
    pub confidence: PatternConfidence,
}

#[derive(Debug, Clone)]
pub enum OptimizationAction {
    /// Increase the node's timeout by `delta_ms`.
    AdjustTimeout { current_ms: u64, new_ms: u64 },
    /// Enable result caching for this node.
    EnableCaching,
    /// Switch to a cheaper model for this node's traffic type.
    ModelDowngrade { current_model: String, suggested_model: String },
    /// Simplify the graph path by removing an unnecessary node.
    SimplifyPath { remove_node: String },
    /// Increase the quota for this tenant.
    IncreaseQuota,
    /// Adjust retry policy (more/fewer retries, different backoff).
    AdjustRetry { max_attempts: u32, backoff_multiplier: f64 },
}

impl std::fmt::Display for OptimizationAction {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            OptimizationAction::AdjustTimeout { current_ms, new_ms } => {
                write!(f, "adjust_timeout: {}ms -> {}ms", current_ms, new_ms)
            }
            OptimizationAction::EnableCaching => write!(f, "enable_caching"),
            OptimizationAction::ModelDowngrade { current_model, suggested_model } => {
                write!(f, "model_downgrade: {} -> {}", current_model, suggested_model)
            }
            OptimizationAction::SimplifyPath { remove_node } => {
                write!(f, "simplify_path: remove {}", remove_node)
            }
            OptimizationAction::IncreaseQuota => write!(f, "increase_quota"),
            OptimizationAction::AdjustRetry { max_attempts, backoff_multiplier } => {
                write!(f, "adjust_retry: attempts={}, backoff={}", max_attempts, backoff_multiplier)
            }
        }
    }
}

// ---------------------------------------------------------------------------
// Graph Optimizer
// ---------------------------------------------------------------------------

/// Self-optimization engine for SAR graphs.
///
/// Wraps `StateGraphMemory::detect_patterns()` with action generation and
/// Go-backend emission.
pub struct GraphOptimizer {
    state: Arc<StateGraphMemory>,
    config: ThresholdConfig,
    last_run: Arc<RwLock<Instant>>,
}

impl GraphOptimizer {
    /// Create a new optimizer with state graph memory and config.
    pub fn new(state: Arc<StateGraphMemory>, config: ThresholdConfig) -> Self {
        Self {
            state,
            config,
            last_run: Arc::new(RwLock::new(Instant::now())),
        }
    }

    /// Analyze execution history for a graph and emit optimization suggestions.
    #[instrument(skip_all, fields(graph_id = %graph_id))]
    pub async fn analyze(&self, graph_id: uuid::Uuid) -> Vec<OptimizationSuggestion> {
        // Use the existing StateGraphMemory pattern detection
        let detected = self.state.detect_patterns().await;
        let pattern_count = detected.len();
        let suggestions = self.build_suggestions(graph_id, detected).await;

        debug!(
            graph_id = %graph_id,
            patterns = pattern_count,
            suggestions = suggestions.len(),
            "Optimizer analysis complete"
        );

        suggestions
    }

    /// Convert detected patterns to optimization suggestions.
    async fn build_suggestions(
        &self,
        graph_id: uuid::Uuid,
        patterns: Vec<state::DetectedPattern>,
    ) -> Vec<OptimizationSuggestion> {
        let mut suggestions = Vec::new();

        for pattern in patterns {
            // Skip low-confidence patterns
            let confidence: PatternConfidence = pattern.confidence.into();
            if confidence < PatternConfidence::Medium {
                continue;
            }

            if let Some(suggestion) = self.suggestion_from_pattern(graph_id, &pattern, confidence).await {
                suggestions.push(suggestion);
            }
        }

        suggestions
    }

    async fn suggestion_from_pattern(
        &self,
        graph_id: uuid::Uuid,
        pattern: &state::DetectedPattern,
        confidence: PatternConfidence,
    ) -> Option<OptimizationSuggestion> {
        let node_id = pattern.node_id.clone();
        let node_name = self.state.get_node_metrics(&node_id).await
            .and_then(|m| m.node_name.clone())
            .unwrap_or_else(|| node_id.clone());

        let uuid_val = uuid::Uuid::parse_str(&node_id).ok();
        let current_timeout = if let Some(uuid_val) = uuid_val {
            self.state.get_node_timeout(uuid_val).await
        } else {
            30_000
        };

        let action = match pattern.pattern_type {
            state::PatternType::HighTimeoutRate => {
                // Increase timeout by 50%
                let new_timeout = ((current_timeout as f64) * 1.5) as u64;
                OptimizationAction::AdjustTimeout {
                    current_ms: current_timeout,
                    new_ms: new_timeout,
                }
            }
            state::PatternType::StableHighSuccess => {
                // Stable high success — suggest enabling caching
                OptimizationAction::EnableCaching
            }
            state::PatternType::HighLatencyVariance => {
                // High variance — suggest investigating upstream
                OptimizationAction::SimplifyPath {
                    remove_node: node_id.clone(),
                }
            }
            state::PatternType::LowSuccessRate => {
                // Low success — suggest retry adjustments
                OptimizationAction::AdjustRetry {
                    max_attempts: 5,
                    backoff_multiplier: 2.0,
                }
            }
        };

        let description = format!(
            "[{:?}] {} for node {} ({}): {}",
            confidence,
            action,
            node_name,
            node_id,
            pattern.detail
        );

        Some(OptimizationSuggestion {
            id: uuid::Uuid::new_v4(),
            node_id,
            node_name,
            action,
            expected_impact: Some(0.15),
            difficulty: "low",
            description,
            graph_id,
            confidence,
        })
    }

    /// Emit a suggestion to the Go backend for review.
    //
    // The Go backend reviews the suggestion and applies it if approved.
    // This keeps a human-in-the-loop for all production graph mutations.
    pub async fn emit_suggestion(&self, suggestion: &OptimizationSuggestion) {
        info!(
            suggestion_id = %suggestion.id,
            node_id = %suggestion.node_id,
            action = %suggestion.action,
            "Emitting optimization suggestion to Go backend"
        );

        let go_endpoint = std::env::var("ORCHESTRATOR_OPTIMIZE_URL")
            .unwrap_or_else(|_| "http://localhost:8080/api/optimizations".to_string());

        let payload = serde_json::json!({
            "id": suggestion.id.to_string(),
            "graph_id": suggestion.graph_id.to_string(),
            "node_id": suggestion.node_id,
            "pattern_confidence": suggestion.confidence.label(),
            "action": suggestion.action.to_string(),
            "expected_impact": suggestion.expected_impact,
            "difficulty": suggestion.difficulty,
            "description": suggestion.description,
        });

        match reqwest::Client::new()
            .post(&go_endpoint)
            .json(&payload)
            .timeout(Duration::from_secs(5))
            .send()
            .await
        {
            Ok(resp) if resp.status().is_success() => {
                info!(suggestion_id = %suggestion.id, "Optimization suggestion emitted successfully");
            }
            Ok(resp) => {
                tracing::warn!(
                    suggestion_id = %suggestion.id,
                    status = %resp.status().as_u16(),
                    "Optimization suggestion emit returned non-success"
                );
            }
            Err(e) => {
                tracing::warn!(suggestion_id = %suggestion.id, error = %e, "Failed to emit optimization suggestion");
            }
        }
    }
}