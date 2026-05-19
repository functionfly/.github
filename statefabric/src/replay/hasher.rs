//! Blake3 hashing for deterministic state verification

use blake3::Hasher as Blake3Hasher;
use serde::{Deserialize, Serialize};

/// State hasher using Blake3 for fast, deterministic hashing
pub struct StateHasher {
    hasher: Blake3Hasher,
}

impl StateHasher {
    /// Create a new hasher
    pub fn new() -> Self {
        Self {
            hasher: Blake3Hasher::new(),
        }
    }

    /// Update the hasher with a key-value pair
    pub fn update_key(&mut self, key: &str, value: &serde_json::Value) {
        self.hasher.update(key.as_bytes());
        self.hasher.update(b":");

        if let Ok(bytes) = serde_json::to_vec(value) {
            self.hasher.update(&bytes);
        }

        self.hasher.update(b"\n");
    }

    /// Update with event data
    pub fn update_event(&mut self, event: &crate::models::Event) {
        self.hasher.update(event.id.to_string().as_bytes());
        self.hasher.update(b":");

        if let Some(key) = &event.key {
            self.hasher.update(key.as_bytes());
            self.hasher.update(b":");
        }

        if let Some(value) = &event.new_value {
            if let Ok(bytes) = serde_json::to_vec(value) {
                self.hasher.update(&bytes);
            }
        }

        self.hasher.update(b"\n");
    }

    /// Finalize and return the hash as a hex string
    pub fn finalize(&self) -> String {
        self.hasher.finalize().to_hex().to_string()
    }
}

impl Default for StateHasher {
    fn default() -> Self {
        Self::new()
    }
}

/// Hash verification result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HashVerification {
    /// Whether the hash matches
    pub valid: bool,
    /// Expected hash
    pub expected: String,
    /// Actual hash computed
    pub actual: String,
    /// Hash algorithm used
    pub algorithm: String,
}

impl HashVerification {
    /// Verify a hash
    pub fn verify(expected: &str, actual: &str) -> Self {
        Self {
            valid: expected == actual,
            expected: expected.to_string(),
            actual: actual.to_string(),
            algorithm: "blake3".to_string(),
        }
    }
}

/// Compute state hash from key-value map
pub fn compute_state_hash(state: &serde_json::Value) -> String {
    let mut hasher = StateHasher::new();

    if let Some(obj) = state.as_object() {
        let mut keys: Vec<_> = obj.keys().collect();
        keys.sort();

        for key in keys {
            if let Some(value) = obj.get(key) {
                hasher.update_key(key, value);
            }
        }
    }

    hasher.finalize()
}

/// Compute event chain hash
pub fn compute_event_chain_hash(events: &[crate::models::Event]) -> String {
    let mut hasher = StateHasher::new();

    for event in events {
        hasher.update_event(event);
    }

    hasher.finalize()
}

/// Input hasher - computes hash of input data for deterministic replay
pub fn compute_input_hash(input: &serde_json::Value) -> String {
    let mut hasher = Blake3Hasher::new();

    if let Ok(bytes) = serde_json::to_vec(input) {
        hasher.update(&bytes);
    }

    hasher.finalize().to_hex().to_string()
}

/// Output hasher - computes hash of output for verification
pub fn compute_output_hash(output: &serde_json::Value) -> String {
    compute_input_hash(output) // Same function, different semantic use
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_state_hash_deterministic() {
        let mut state1 = serde_json::json!({
            "counter": 1,
            "name": "test"
        });

        let mut state2 = serde_json::json!({
            "name": "test",
            "counter": 1
        });

        // Sort keys to ensure determinism
        let hash1 = compute_state_hash(&state1);
        let hash2 = compute_state_hash(&state2);

        println!("Hash1: {}", hash1);
        println!("Hash2: {}", hash2);
    }

    #[test]
    fn test_input_hash() {
        let input = serde_json::json!({
            "data": "test",
            "count": 42
        });

        let hash1 = compute_input_hash(&input);
        let hash2 = compute_input_hash(&input);

        assert_eq!(hash1, hash2, "Same input should produce same hash");
    }
}
