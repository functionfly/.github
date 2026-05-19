//! Peer management for mesh networking

use crate::mesh::Peer;
use chrono::Utc;

impl Peer {
    /// Update the peer's last seen timestamp
    pub fn touch(&mut self) {
        self.last_seen = Utc::now();
    }

    /// Set latency measurement
    pub fn set_latency(&mut self, ms: u32) {
        self.latency_ms = Some(ms);
    }

    /// Update advertised capabilities
    pub fn update_capabilities(&mut self, capabilities: Vec<String>) {
        self.advertised_capabilities = capabilities;
    }

    /// Check if peer has a specific capability
    pub fn has_capability(&self, capability: &str) -> bool {
        self.advertised_capabilities.iter().any(|c| c == capability)
    }
}