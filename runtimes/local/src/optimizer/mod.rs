//! Self-optimization engine for SAR graph execution.
//!
//! Implements threshold-based graph rewriting driven by per-node execution metrics.
//! Uses `StateGraphMemory` (from `memory::state`) as the source of truth for
//! per-node success rates, latency, and cost history.
//!
//! ## Architecture
//!
//! The optimizer uses the existing `StateGraphMemory::detect_patterns()` method
//! which already implements:
//! | Pattern | Threshold | Action |
//! |---------|-----------|--------|
//! | High timeout rate | `>10%` timeouts | Increase node timeout |
//! | Stable high success | `>95%` success, `>100` calls | Suggest node consolidation |
//! | High latency variance | `p95 > 3x avg` | Suggest parallel execution |
//!
//! This module wraps those patterns with:
//! - `PatternConfidence` calculation based on sample size and metric value
//! - `OptimizationAction` generation from detected patterns
//! - HTTP emission to the Go backend for review
//!
//! ## Graph Mutation (Phase 7)
//!
//! The `mutation` submodule implements the actual graph rewriting operations:
//! - Timeout adjustments (high confidence)
//! - Path simplification (remove dead branches)
//! - Caching node insertion
//! - Retry policy adjustments
//! - Model switching
//!
//! All mutations support rollback and canary testing.
//!
//! ## Safety
//!
//! This module NEVER automatically mutates a production graph. All optimizations
//! are emitted as `OptimizationSuggestion` records that the Go backend can review
//! and approve before applying. This preserves the "human-in-the-loop" principle.

pub mod config;
pub mod optimizer;
pub mod mutation;

pub use config::ThresholdConfig;
pub use optimizer::{GraphOptimizer, OptimizationAction, OptimizationSuggestion, PatternConfidence};
pub use mutation::{GraphMutator, CanaryTester, MutationResult, AppliedMutation, MutationType, GraphBackup};
