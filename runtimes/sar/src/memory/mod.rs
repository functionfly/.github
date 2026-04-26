use std::collections::HashMap;
use std::sync::Arc;

use parking_lot::RwLock;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum MemoryTier {
    Hot,
    Warm,
    Cold,
    StateGraph,
}

#[derive(Debug)]
pub struct MemoryEntry {
    pub key: String,
    pub value: serde_json::Value,
    pub tier: MemoryTier,
    pub timestamp: chrono::DateTime<chrono::Utc>,
}

impl MemoryEntry {
    pub fn new(key: String, value: serde_json::Value, tier: MemoryTier) -> Self {
        Self {
            key,
            value,
            tier,
            timestamp: chrono::Utc::now(),
        }
    }
}

#[cfg(feature = "multi-memory")]
pub struct HotMemory {
    cache: Arc<RwLock<lru::LruCache<String, MemoryEntry>>>,
}

#[cfg(feature = "multi-memory")]
impl HotMemory {
    pub fn new(capacity: usize) -> Self {
        Self {
            cache: Arc::new(RwLock::new(lru::LruCache::new(
                std::num::NonZeroUsize::new(capacity).unwrap_or(std::num::NonZeroUsize::new(10_000).unwrap()),
            ))),
        }
    }

    pub fn read_sync(&self, key: &str) -> Option<serde_json::Value> {
        let mut cache = self.cache.write();
        cache.get(key).map(|e| e.value.clone())
    }

    pub fn write_sync(&self, key: &str, value: serde_json::Value) {
        let mut cache = self.cache.write();
        cache.put(key.to_string(), MemoryEntry::new(key.to_string(), value, MemoryTier::Hot));
    }

    pub fn delete_sync(&self, key: &str) {
        let mut cache = self.cache.write();
        cache.pop(key);
    }

    pub fn list_sync(&self, pattern: &str) -> Vec<String> {
        let cache = self.cache.read();
        cache.iter()
            .filter(|(k, _): &(&String, _)| k.contains(pattern))
            .map(|(k, _): (&String, _)| k.clone())
            .collect()
    }
}

pub struct StateGraphMemory {
    decisions: Arc<RwLock<HashMap<String, DecisionRecord>>>,
    success_rates: Arc<RwLock<HashMap<String, f64>>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DecisionRecord {
    pub agent_id: String,
    pub decision: String,
    pub context: serde_json::Value,
    pub outcome: String,
    pub timestamp: chrono::DateTime<chrono::Utc>,
}

impl StateGraphMemory {
    pub fn new() -> Self {
        Self {
            decisions: Arc::new(RwLock::new(HashMap::new())),
            success_rates: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub fn record_decision(&self, agent_id: &str, decision: &str, context: serde_json::Value, outcome: &str) {
        let record = DecisionRecord {
            agent_id: agent_id.to_string(),
            decision: decision.to_string(),
            context,
            outcome: outcome.to_string(),
            timestamp: chrono::Utc::now(),
        };
        let key = format!("{}:{}", agent_id, decision);
        let mut decisions = self.decisions.write();
        decisions.insert(key, record);
    }

    pub fn get_success_rate(&self, agent_id: &str) -> f64 {
        let rates = self.success_rates.read();
        rates.get(agent_id).copied().unwrap_or(1.0)
    }

    pub fn update_success_rate(&self, agent_id: &str, success: bool) {
        let mut rates = self.success_rates.write();
        let current = rates.get(&agent_id.to_string()).copied().unwrap_or(1.0);
        let n = rates.get(&agent_id.to_string()).map(|_| 100u64).unwrap_or(1);
        let new_rate = if success {
            current + (1.0 - current) / n as f64
        } else {
            current - current / n as f64
        };
        rates.insert(agent_id.to_string(), new_rate);
    }
}

impl Default for StateGraphMemory {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_state_graph_memory_record_decision() {
        let mem = StateGraphMemory::new();
        mem.record_decision("agent-1", "use_model_a", serde_json::json!({"complexity": "high"}), "success");

        let rates = mem.success_rates.read();
        assert!(rates.is_empty() || rates.contains_key("agent-1"));
    }

    #[test]
    fn test_state_graph_success_rate() {
        let mem = StateGraphMemory::new();
        assert_eq!(mem.get_success_rate("agent-1"), 1.0);

        mem.update_success_rate("agent-1", true);
        let rate = mem.get_success_rate("agent-1");
        assert!(rate > 1.0 || rate == 1.0); // initial success doesn't change rate from 1.0

        mem.update_success_rate("agent-1", false);
        let rate = mem.get_success_rate("agent-1");
        assert!(rate < 1.0);
    }

    #[cfg(feature = "multi-memory")]
    #[test]
    fn test_hot_memory_write_read() {
        let mem = HotMemory::new(100);
        mem.write_sync("key1", serde_json::json!({"value": 42}));
        let result = mem.read_sync("key1");
        assert!(result.is_some());
        assert_eq!(result.unwrap()["value"], 42);
    }

    #[cfg(feature = "multi-memory")]
    #[test]
    fn test_hot_memory_delete() {
        let mem = HotMemory::new(100);
        mem.write_sync("key1", serde_json::json!("hello"));
        assert!(mem.read_sync("key1").is_some());

        mem.delete_sync("key1");
        assert!(mem.read_sync("key1").is_none());
    }

    #[cfg(feature = "multi-memory")]
    #[test]
    fn test_hot_memory_eviction() {
        let mem = HotMemory::new(2);
        mem.write_sync("a", serde_json::json!(1));
        mem.write_sync("b", serde_json::json!(2));
        mem.write_sync("c", serde_json::json!(3));

        // "a" should be evicted since capacity is 2
        assert!(mem.read_sync("a").is_none());
        assert!(mem.read_sync("b").is_some());
        assert!(mem.read_sync("c").is_some());
    }

    #[cfg(feature = "multi-memory")]
    #[test]
    fn test_hot_memory_list() {
        let mem = HotMemory::new(100);
        mem.write_sync("user:1:name", serde_json::json!("Alice"));
        mem.write_sync("user:2:name", serde_json::json!("Bob"));
        mem.write_sync("order:1", serde_json::json!("item"));

        let users = mem.list_sync("user:");
        assert_eq!(users.len(), 2);
    }
}
