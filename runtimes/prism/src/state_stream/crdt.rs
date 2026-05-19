//! CRDT engine for StateStream Memory Fabric

use std::collections::HashMap;
use serde::{Deserialize, Serialize};

use crate::core::{PrismError, PrismResult};

/// A CRDT operation
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum CrdtOp {
    /// Last-Writer-Wins Register
    Lww(LwwOp),
    /// Grow-only Counter
    GCounter(GCounterOp),
    /// Positive-Negative Counter
    PnCounter(PnCounterOp),
}

/// LWW Register operation
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LwwOp {
    pub node_id: String,
    pub timestamp: i64,
    pub value: Vec<u8>,
}

/// GCounter operation
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GCounterOp {
    pub node_id: String,
    pub delta: u64,
}

/// PN-Counter operation
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PnCounterOp {
    pub node_id: String,
    pub delta: i64,
}

/// CRDT engine that supports multiple CRDT types
pub struct CrdtEngine {
    node_id: String,
    registers: HashMap<String, LwwRegister>,
    counters: HashMap<String, GCounter>,
    pn_counters: HashMap<String, PnCounter>,
    /// Operation history for debugging and auditing
    op_history: Vec<CrdtOpRecord>,
}

/// A record of a CRDT operation for auditing
#[derive(Debug, Clone)]
pub struct CrdtOpRecord {
    pub timestamp: i64,
    pub key: String,
    pub op_type: String,
    pub node_id: String,
}

impl CrdtEngine {
    pub fn new(node_id: impl Into<String>) -> Self {
        Self {
            node_id: node_id.into(),
            registers: HashMap::new(),
            counters: HashMap::new(),
            pn_counters: HashMap::new(),
            op_history: Vec::new(),
        }
    }

    /// Get the node ID for this CRDT engine
    pub fn get_node_id(&self) -> &str {
        &self.node_id
    }

    /// Set the node ID (for migration scenarios)
    pub fn set_node_id(&mut self, node_id: impl Into<String>) {
        self.node_id = node_id.into();
    }

/// Apply a CRDT operation
    pub fn apply(&mut self, key: &str, op: CrdtOp) -> PrismResult<Vec<u8>> {
        // Record operation in history
        let op_type = match op {
            CrdtOp::Lww(_) => "LWW",
            CrdtOp::GCounter(_) => "GCounter",
            CrdtOp::PnCounter(_) => "PnCounter",
        };

        self.op_history.push(CrdtOpRecord {
            timestamp: chrono::Utc::now().timestamp(),
            key: key.to_string(),
            op_type: op_type.to_string(),
            node_id: self.node_id.clone(),
        });

        // Keep history bounded
        if self.op_history.len() > 10000 {
            self.op_history.drain(0..1000);
        }

        match op {
            CrdtOp::Lww(lww) => {
                let reg = self.registers.entry(key.to_string()).or_insert_with(LwwRegister::new);
                reg.apply(lww)
            }
            CrdtOp::GCounter(gc) => {
                let counter = self.counters.entry(key.to_string()).or_insert_with(GCounter::new);
                counter.apply(gc)
            }
            CrdtOp::PnCounter(pn) => {
                let counter = self.pn_counters.entry(key.to_string()).or_insert_with(PnCounter::new);
                counter.apply(pn)
            }
        }
    }

    /// Get the operation history
    pub fn get_history(&self) -> &[CrdtOpRecord] {
        &self.op_history
    }

    /// Get recent operations for a key
    pub fn get_history_for_key(&self, key: &str) -> Vec<&CrdtOpRecord> {
        self.op_history.iter()
            .filter(|r| r.key == key)
            .collect()
    }

    /// Clear the operation history
    pub fn clear_history(&mut self) {
        self.op_history.clear();
    }

    /// Get the current value of a register
    pub fn get_register(&self, key: &str) -> Option<&LwwRegister> {
        self.registers.get(key)
    }

    /// Get the current value of a counter
    pub fn get_counter(&self, key: &str) -> Option<u64> {
        self.counters.get(key).map(|c| c.value())
    }

    /// Get the current value of a PN-counter
    pub fn get_pn_counter(&self, key: &str) -> Option<i64> {
        self.pn_counters.get(key).map(|c| c.value())
    }

    /// Merge another CrdtEngine's state into this one
    pub fn merge(&mut self, other: &CrdtEngine) {
        // Merge registers
        for (key, reg) in &other.registers {
            let self_reg = self.registers.entry(key.clone()).or_insert_with(LwwRegister::new);
            self_reg.merge(reg);
        }

        // Merge G-counters
        for (key, counter) in &other.counters {
            let self_counter = self.counters.entry(key.clone()).or_insert_with(GCounter::new);
            self_counter.merge(counter);
        }

        // Merge PN-counters
        for (key, counter) in &other.pn_counters {
            let self_counter = self.pn_counters.entry(key.clone()).or_insert_with(PnCounter::new);
            self_counter.merge(counter);
        }
    }
}

/// Last-Writer-Wins Register
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LwwRegister {
    /// Per-node timestamps
    values: HashMap<String, (i64, Vec<u8>)>,
}

impl LwwRegister {
    pub fn new() -> Self {
        Self { values: HashMap::new() }
    }

    pub fn apply(&mut self, op: LwwOp) -> PrismResult<Vec<u8>> {
        let current = self.values.get(&op.node_id);

        // Only apply if timestamp is newer
        if let Some(&(ts, _)) = current {
            if op.timestamp <= ts {
                return Err(PrismError::StateStreamError("Stale LWW operation".to_string()));
            }
        }

        self.values.insert(op.node_id, (op.timestamp, op.value.clone()));
        Ok(op.value)
    }

    pub fn merge(&mut self, other: &LwwRegister) {
        for (node_id, (ts, val)) in &other.values {
            if let Some((current_ts, _)) = self.values.get(node_id) {
                if ts > current_ts {
                    self.values.insert(node_id.clone(), (*ts, val.clone()));
                }
            } else {
                self.values.insert(node_id.clone(), (*ts, val.clone()));
            }
        }
    }

    pub fn value(&self) -> Option<&Vec<u8>> {
        self.values.values().max_by_key(|(ts, _)| ts).map(|(_, v)| v)
    }
}

impl Default for LwwRegister {
    fn default() -> Self {
        Self::new()
    }
}

/// Grow-only Counter
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GCounter {
    /// Per-node counts
    counts: HashMap<String, u64>,
}

impl GCounter {
    pub fn new() -> Self {
        Self { counts: HashMap::new() }
    }

    pub fn apply(&mut self, op: GCounterOp) -> PrismResult<Vec<u8>> {
        let current = self.counts.get(&op.node_id).unwrap_or(&0);
        self.counts.insert(op.node_id.clone(), current + op.delta);
        Ok(self.value().to_le_bytes().to_vec())
    }

    pub fn merge(&mut self, other: &GCounter) {
        for (node_id, &count) in &other.counts {
            let entry = self.counts.entry(node_id.clone()).or_insert(0);
            if count > *entry {
                *entry = count;
            }
        }
    }

    pub fn value(&self) -> u64 {
        self.counts.values().sum()
    }
}

impl Default for GCounter {
    fn default() -> Self {
        Self::new()
    }
}

/// Positive-Negative Counter
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PnCounter {
    /// Positive counts
    positive: HashMap<String, u64>,
    /// Negative counts
    negative: HashMap<String, u64>,
}

impl PnCounter {
    pub fn new() -> Self {
        Self {
            positive: HashMap::new(),
            negative: HashMap::new(),
        }
    }

    pub fn apply(&mut self, op: PnCounterOp) -> PrismResult<Vec<u8>> {
        if op.delta >= 0 {
            let current = self.positive.get(&op.node_id).unwrap_or(&0);
            self.positive.insert(op.node_id.clone(), current + op.delta as u64);
        } else {
            let current = self.negative.get(&op.node_id).unwrap_or(&0);
            self.negative.insert(op.node_id.clone(), current + (-op.delta) as u64);
        }
        Ok(self.value().to_le_bytes().to_vec())
    }

    pub fn merge(&mut self, other: &PnCounter) {
        // Merge positive
        for (node_id, &count) in &other.positive {
            let entry = self.positive.entry(node_id.clone()).or_insert(0);
            if count > *entry {
                *entry = count;
            }
        }

        // Merge negative
        for (node_id, &count) in &other.negative {
            let entry = self.negative.entry(node_id.clone()).or_insert(0);
            if count > *entry {
                *entry = count;
            }
        }
    }

    pub fn value(&self) -> i64 {
        let pos: u64 = self.positive.values().sum();
        let neg: u64 = self.negative.values().sum();
        pos as i64 - neg as i64
    }
}

impl Default for PnCounter {
    fn default() -> Self {
        Self::new()
    }
}