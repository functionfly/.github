//! Capability discovery for UCL

use serde::{Deserialize, Serialize};

use super::{Capability, CapabilityCategory};
use crate::core::PrismResult;

/// Discovery query parameters
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DiscoveryQuery {
    /// Natural language query
    pub query: String,
    /// Required category
    pub category: Option<CapabilityCategory>,
    /// Required tags (all must match)
    pub required_tags: Vec<String>,
    /// Maximum latency in milliseconds
    pub max_latency_ms: Option<u32>,
    /// Minimum throughput in RPS
    pub min_throughput_rps: Option<u32>,
    /// Whether GPU is required
    pub gpu_required: bool,
    /// Minimum trust score (0.0 to 1.0)
    pub min_trust_score: f32,
    /// Preferred runtimes
    pub preferred_runtimes: Vec<String>,
}

/// Result of a capability discovery
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DiscoveryResult {
    /// Matching capabilities sorted by relevance
    pub matches: Vec<CapabilityMatch>,
    /// Total number of matches found
    pub total_found: usize,
    /// Discovery query ID
    pub query_id: String,
}

/// A capability match with relevance scoring
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CapabilityMatch {
    pub capability: Capability,
    /// Relevance score (0.0 to 1.0, higher is better)
    pub relevance_score: f32,
    /// Tags that matched the query
    pub matched_tags: Vec<String>,
    /// Human-readable reason for the match
    pub recommendation_reason: String,
}

/// Capability discovery service
pub struct CapabilityDiscovery {
    registry: super::CapabilityRegistry,
}

impl CapabilityDiscovery {
    pub fn new(registry: super::CapabilityRegistry) -> Self {
        Self { registry }
    }

    /// Discover capabilities matching a query
    pub async fn discover(&self, query: DiscoveryQuery) -> PrismResult<DiscoveryResult> {
        let query_id = uuid::Uuid::new_v4().to_string();
        let mut matches = Vec::new();

        // Get all capabilities
        let all_caps = self.registry.list_all().await;

        for cap in all_caps {
            let (score, matched_tags) = self.score_capability(&cap, &query);

            if score > 0.0 {
                matches.push(CapabilityMatch {
                    capability: cap,
                    relevance_score: score,
                    matched_tags,
                    recommendation_reason: Self::generate_reason(&query),
                });
            }
        }

        // Sort by relevance score
        matches.sort_by(|a, b| b.relevance_score.partial_cmp(&a.relevance_score).unwrap());

        let total_found = matches.len();

        Ok(DiscoveryResult {
            matches,
            total_found,
            query_id,
        })
    }

    /// Score how well a capability matches a query
    fn score_capability(&self, cap: &Capability, query: &DiscoveryQuery) -> (f32, Vec<String>) {
        let mut score: f32 = 0.0;
        let mut matched_tags = Vec::new();

        // Category match
        if let Some(ref cat) = query.category {
            if cap.category == *cat {
                score += 0.2;
            }
        }

        // Tag matching
        for tag in &query.required_tags {
            if cap.metadata.tags.values().any(|t| t == tag) {
                score += 0.15;
                matched_tags.push(tag.clone());
            }
        }

        // GPU requirement
        if query.gpu_required && cap.requires_gpu() {
            score += 0.2;
        } else if query.gpu_required {
            return (0.0, Vec::new()); // GPU required but not available
        }

        // Latency requirement
        if let Some(max_lat) = query.max_latency_ms {
            if cap.performance.avg_latency_ms <= max_lat {
                score += 0.15;
            } else {
                score -= 0.1;
            }
        }

        // Throughput requirement
        if let Some(min_rps) = query.min_throughput_rps {
            if cap.performance.throughput_rps >= min_rps {
                score += 0.1;
            }
        }

        // Trust score
        if cap.trust.score >= query.min_trust_score {
            score += 0.1;
        } else {
            return (0.0, Vec::new()); // Trust requirement not met
        }

        // Runtime preference
        if !query.preferred_runtimes.is_empty() {
            if cap.runtimes.iter().any(|r| query.preferred_runtimes.contains(r)) {
                score += 0.1;
            }
        }

        // Name/description keyword match
        let query_lower = query.query.to_lowercase();
        if cap.name.to_lowercase().contains(&query_lower)
            || cap.metadata.description.to_lowercase().contains(&query_lower)
        {
            score += 0.2;
        }

        let final_score = if score > 1.0 { 1.0 } else if score < 0.0 { 0.0 } else { score };
        (final_score, matched_tags)
    }

    /// Generate a human-readable reason for the match
    fn generate_reason(query: &DiscoveryQuery) -> String {
        if query.gpu_required {
            "Provides GPU acceleration".to_string()
        } else if query.max_latency_ms.is_some() {
            format!("Meets latency requirement of {}ms", query.max_latency_ms.unwrap())
        } else {
            "Matches requested capabilities".to_string()
        }
    }
}