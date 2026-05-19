//! Feedback loop for continuous optimization

use std::collections::VecDeque;
use serde::{Deserialize, Serialize};

use crate::core::CellId;
use crate::neural::optimizer::{ExecutionOutcome, OptimizationSuggestion};

/// Feedback entry for the optimization loop
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FeedbackEntry {
    pub cell_id: CellId,
    pub suggestion: OptimizationSuggestion,
    pub actual_outcome: ExecutionOutcome,
    pub improvement_score: f32,
    pub timestamp: i64,
}

/// Feedback loop for continuous improvement
pub struct FeedbackLoop {
    entries: VecDeque<FeedbackEntry>,
    max_entries: usize,
}

impl FeedbackLoop {
    pub fn new(max_entries: usize) -> Self {
        Self {
            entries: VecDeque::new(),
            max_entries,
        }
    }

    /// Record feedback for a suggestion
    pub fn record(&mut self, entry: FeedbackEntry) {
        self.entries.push_back(entry);

        while self.entries.len() > self.max_entries {
            self.entries.pop_front();
        }
    }

    /// Get recent feedback entries for a cell
    pub fn get_recent(&self, cell_id: &CellId, count: usize) -> Vec<&FeedbackEntry> {
        self.entries
            .iter()
            .filter(|e| e.cell_id == *cell_id)
            .rev()
            .take(count)
            .collect()
    }

    /// Calculate improvement score from feedback
    pub fn calculate_improvement(&self, cell_id: &CellId) -> f32 {
        let recent = self.get_recent(cell_id, 10);
        if recent.is_empty() {
            return 0.0;
        }

        let total: f32 = recent.iter().map(|e| e.improvement_score).sum();
        total / recent.len() as f32
    }

    /// Get entries with positive improvement
    pub fn positive_feedback(&self) -> Vec<&FeedbackEntry> {
        self.entries
            .iter()
            .filter(|e| e.improvement_score > 0.0)
            .collect()
    }

    /// Get entries with negative feedback (for debugging)
    pub fn negative_feedback(&self) -> Vec<&FeedbackEntry> {
        self.entries
            .iter()
            .filter(|e| e.improvement_score < 0.0)
            .collect()
    }
}

impl Default for FeedbackLoop {
    fn default() -> Self {
        Self::new(1000)
    }
}