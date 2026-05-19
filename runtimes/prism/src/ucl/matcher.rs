//! Capability matcher for UCL


use super::{Capability, DiscoveryQuery};

/// Match score between a query and a capability
#[derive(Debug, Clone)]
pub struct MatchScore {
    /// Total score (0.0 to 1.0)
    pub total: f32,
    /// Category score
    pub category: f32,
    /// Performance score
    pub performance: f32,
    /// Trust score
    pub trust: f32,
    /// Tag match score
    pub tags: f32,
    /// Runtime compatibility score
    pub runtime: f32,
}

/// Capability matcher for scoring and ranking capabilities
pub struct CapabilityMatcher;

impl CapabilityMatcher {
    /// Calculate a comprehensive match score
    pub fn score(capability: &Capability, query: &DiscoveryQuery) -> MatchScore {
        let category_score = Self::score_category(capability, query);
        let performance_score = Self::score_performance(capability, query);
        let trust_score = Self::score_trust(capability, query);
        let tags_score = Self::score_tags(capability, query);
        let runtime_score = Self::score_runtime(capability, query);

        let total = category_score * 0.25
            + performance_score * 0.25
            + trust_score * 0.2
            + tags_score * 0.15
            + runtime_score * 0.15;

        MatchScore {
            total,
            category: category_score,
            performance: performance_score,
            trust: trust_score,
            tags: tags_score,
            runtime: runtime_score,
        }
    }

    /// Score category match
    fn score_category(cap: &Capability, query: &DiscoveryQuery) -> f32 {
        match query.category {
            Some(ref cat) if cat == &cap.category => 1.0,
            Some(_) => 0.0,
            None => 0.5, // No preference, neutral score
        }
    }

    /// Score performance match
    fn score_performance(cap: &Capability, query: &DiscoveryQuery) -> f32 {
        let mut score = 0.5; // Base score

        if let Some(max_lat) = query.max_latency_ms {
            if cap.performance.avg_latency_ms <= max_lat {
                score += 0.3;
            } else if cap.performance.p99_latency_ms <= max_lat {
                score += 0.1;
            } else {
                return 0.0;
            }
        }

        if let Some(min_rps) = query.min_throughput_rps {
            if cap.performance.throughput_rps >= min_rps {
                score += 0.2;
            } else {
                score -= 0.2;
            }
        }

        let final_score = if score > 1.0 { 1.0 } else if score < 0.0 { 0.0 } else { score };
        final_score
    }

    /// Score trust level match
    fn score_trust(cap: &Capability, query: &DiscoveryQuery) -> f32 {
        if cap.trust.score >= query.min_trust_score {
            cap.trust.score
        } else {
            0.0
        }
    }

    /// Score tag match
    fn score_tags(cap: &Capability, query: &DiscoveryQuery) -> f32 {
        if query.required_tags.is_empty() {
            return 0.5;
        }

        let match_count = query.required_tags.iter()
            .filter(|tag| cap.metadata.tags.values().any(|v| v == *tag))
            .count();

        match_count as f32 / query.required_tags.len() as f32
    }

    /// Score runtime compatibility
    fn score_runtime(cap: &Capability, query: &DiscoveryQuery) -> f32 {
        if query.preferred_runtimes.is_empty() {
            return 0.5;
        }

        if cap.runtimes.iter().any(|r| query.preferred_runtimes.contains(r)) {
            1.0
        } else {
            0.0
        }
    }

    /// Check if a capability is acceptable (meets minimum thresholds)
    pub fn is_acceptable(capability: &Capability, query: &DiscoveryQuery) -> bool {
        let score = Self::score(capability, query);
        score.total >= 0.5 && score.trust > 0.0
    }

    /// Rank capabilities by match score
    pub fn rank(capabilities: &[Capability], query: &DiscoveryQuery) -> Vec<(Capability, MatchScore)> {
        let mut ranked: Vec<_> = capabilities
            .iter()
            .filter_map(|cap| {
                let score = Self::score(cap, query);
                if Self::is_acceptable(cap, query) {
                    Some((cap.clone(), score))
                } else {
                    None
                }
            })
            .collect();

        ranked.sort_by(|a, b| b.1.total.partial_cmp(&a.1.total).unwrap());
        ranked
    }
}