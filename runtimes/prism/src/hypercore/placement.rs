//! Placement engine for intelligent cell placement

use std::collections::HashMap;
use serde::{Deserialize, Serialize};

use crate::core::{CellId, ExecutionLocation};

/// A placement decision for a cell
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PlacementDecision {
    pub cell_id: CellId,
    pub node_id: NodeId,
    pub location: ExecutionLocation,
    pub config_overrides: HashMap<String, String>,
    pub scheduling_reason: String,
}

/// A node identifier
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct NodeId(pub String);

impl NodeId {
    pub fn new(id: impl Into<String>) -> Self {
        Self(id.into())
    }

    pub fn as_str(&self) -> &str {
        &self.0
    }
}

impl std::fmt::Display for NodeId {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

/// Score breakdown for placement decisions
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PlacementScore {
    /// Latency score (0.0 to 1.0, higher is better)
    pub latency_score: f32,
    /// Cost score (0.0 to 1.0, higher is better)
    pub cost_score: f32,
    /// GPU affinity score (0.0 to 1.0, higher is better)
    pub gpu_score: f32,
    /// AI affinity score (0.0 to 1.0, higher is better)
    pub ai_affinity_score: f32,
    /// Resource availability score (0.0 to 1.0, higher is better)
    pub availability_score: f32,
    /// Total weighted score (0.0 to 1.0, higher is better)
    pub total: f32,
    /// Estimated cost in USD
    pub cost_estimate: f64,
}

impl PlacementScore {
    /// Return the worst possible score
    pub fn worst() -> Self {
        Self {
            latency_score: 0.0,
            cost_score: 0.0,
            gpu_score: 0.0,
            ai_affinity_score: 0.0,
            availability_score: 0.0,
            total: 0.0,
            cost_estimate: f64::MAX,
        }
    }
}

/// Placement engine for making intelligent scheduling decisions
pub struct PlacementEngine;

impl PlacementEngine {
    /// Calculate a composite score for a placement option
    pub fn score_placement(
        latency_weight: f32,
        cost_weight: f32,
        gpu_weight: f32,
        affinity_weight: f32,
        score: &PlacementScore,
    ) -> f32 {
        score.latency_score * latency_weight
            + score.cost_score * cost_weight
            + score.gpu_score * gpu_weight
            + score.ai_affinity_score * affinity_weight
    }

    /// Determine if a placement should be rejected based on minimum thresholds
    pub fn is_acceptable_placement(score: &PlacementScore, min_score: f32) -> bool {
        score.total >= min_score && score.gpu_score >= 0.5
    }
}