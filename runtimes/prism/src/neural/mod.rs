//! Neural Execution Optimization
//!
//! The runtime learns:
//! - optimal execution paths
//! - hot functions
//! - memory behavior
//! - GPU allocation patterns
//! - agent coordination patterns
//!
//! Using reinforcement learning, the runtime literally optimizes itself over time.

pub mod optimizer;
pub mod profile;
pub mod rl;
pub mod feedback;

pub use optimizer::{NeuralOptimizer, OptimizationPolicy, ExecutionProfile, ExecutionFeatures, ExecutionOutcome, OptimizationSuggestion};
pub use profile::ProfileCollector;
pub use rl::{QLearning, Policy};
pub use feedback::{FeedbackLoop, FeedbackEntry};