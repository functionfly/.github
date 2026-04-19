//! Threshold configuration for pattern detection.
//!
//! These thresholds control when patterns are flagged and how confident
//! the optimizer is in each detection.

/// Default thresholds for pattern detection.
pub struct ThresholdConfig {
    /// Timeout rate above which to flag a node.
    pub timeout_rate_threshold: f64,
    /// Minimum sample size before flagging any pattern.
    pub min_sample_size: u32,
    /// Success rate above which to consider a node stable.
    pub success_rate_threshold: f64,
    /// Latency variance ratio (p99 / p50) above which to flag variance.
    pub latency_variance_threshold: f64,
    /// Cost multiplier above average to flag as high cost.
    pub cost_threshold_multiplier: f64,
}

impl Default for ThresholdConfig {
    fn default() -> Self {
        Self {
            timeout_rate_threshold: 0.10,
            min_sample_size: 20,
            success_rate_threshold: 0.95,
            latency_variance_threshold: 2.0,
            cost_threshold_multiplier: 3.0,
        }
    }
}

impl ThresholdConfig {
    /// Aggressive thresholds for production — higher confidence required.
    pub fn production() -> Self {
        Self {
            timeout_rate_threshold: 0.15,
            min_sample_size: 50,
            success_rate_threshold: 0.98,
            latency_variance_threshold: 2.5,
            cost_threshold_multiplier: 4.0,
        }
    }

    /// Conservative thresholds for development — easier to trigger.
    pub fn development() -> Self {
        Self {
            timeout_rate_threshold: 0.05,
            min_sample_size: 5,
            success_rate_threshold: 0.90,
            latency_variance_threshold: 1.5,
            cost_threshold_multiplier: 2.0,
        }
    }
}