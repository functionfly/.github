//! Mesh routing for discovering peers and capabilities

use std::collections::{HashMap, HashSet};
use serde::{Deserialize, Serialize};

use crate::mesh::{Peer, MeshPeerId};
use crate::codec::{CborCodec, CodecError};

/// Routing table entry
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RoutingEntry {
    pub peer_id: String, // Base58 encoded PeerId for serialization
    pub capabilities: HashSet<String>,
    pub regions: Vec<String>,
    pub latency_ms: Option<u32>,
    pub score: f32,
}

impl RoutingEntry {
    /// Serialize to CBOR bytes
    pub fn to_cbor(&self) -> Result<Vec<u8>, CodecError> {
        CborCodec::encode(self)
    }

    /// Deserialize from CBOR bytes
    pub fn from_cbor(bytes: &[u8]) -> Result<Self, CodecError> {
        CborCodec::decode(bytes)
    }
}

/// Maintains routing information for the mesh
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RoutingTable {
    entries: HashMap<MeshPeerId, RoutingEntry>,
    capability_index: HashMap<String, Vec<MeshPeerId>>,
    region_index: HashMap<String, Vec<MeshPeerId>>,
}

impl RoutingTable {
    pub fn new() -> Self {
        Self {
            entries: HashMap::new(),
            capability_index: HashMap::new(),
            region_index: HashMap::new(),
        }
    }

    /// Add or update a peer in the routing table
    pub fn upsert_peer(&mut self, peer: &Peer) {
        // Extract regions from peer
        let regions: Vec<String> = peer.region.iter().cloned().collect();

        let entry = RoutingEntry {
            peer_id: peer.peer_id.0.clone(),
            capabilities: peer.advertised_capabilities.iter().cloned().collect(),
            regions: regions.clone(),
            latency_ms: peer.latency_ms,
            score: peer.availability_score(),
        };

        // Remove from old capability indexes if updating
        if let Some(existing) = self.entries.remove(&peer.peer_id) {
            for cap in &existing.capabilities {
                if let Some(peers) = self.capability_index.get_mut(cap) {
                    peers.retain(|id| id != &peer.peer_id);
                }
            }
            for region in &existing.regions {
                if let Some(peers) = self.region_index.get_mut(region) {
                    peers.retain(|id| id != &peer.peer_id);
                }
            }
        }

        self.entries.insert(peer.peer_id.clone(), entry);

        // Update capability index
        for cap in &peer.advertised_capabilities {
            self.capability_index
                .entry(cap.clone())
                .or_default()
                .push(peer.peer_id.clone());
        }

        // Update region index
        for region in &regions {
            self.region_index
                .entry(region.clone())
                .or_default()
                .push(peer.peer_id.clone());
        }
    }

    /// Find peers that have a capability
    pub fn find_by_capability(&self, capability: &str) -> Vec<MeshPeerId> {
        self.capability_index
            .get(capability)
            .cloned()
            .unwrap_or_default()
    }

    /// Find peers in a specific region
    pub fn find_by_region(&self, region: &str) -> Vec<MeshPeerId> {
        self.region_index
            .get(region)
            .cloned()
            .unwrap_or_default()
    }

    /// Get the best peer for a capability
    pub fn best_for_capability(&self, capability: &str) -> Option<MeshPeerId> {
        self.find_by_capability(capability)
            .into_iter()
            .filter_map(|id| self.entries.get(&id))
            .max_by(|a, b| a.score.partial_cmp(&b.score).unwrap())
            .map(|e| MeshPeerId(e.peer_id.clone()))
    }

    /// Get all known peer IDs
    pub fn all_peers(&self) -> Vec<MeshPeerId> {
        self.entries.keys().cloned().collect()
    }

    /// Remove a peer from the routing table
    pub fn remove_peer(&mut self, peer_id: &MeshPeerId) {
        if let Some(entry) = self.entries.remove(peer_id) {
            // Remove from capability index
            for cap in &entry.capabilities {
                if let Some(peers) = self.capability_index.get_mut(cap) {
                    peers.retain(|id| id != peer_id);
                }
            }

            // Remove from region index
            for region in &entry.regions {
                if let Some(peers) = self.region_index.get_mut(region) {
                    peers.retain(|id| id != peer_id);
                }
            }
        }
    }
}

impl Default for RoutingTable {
    fn default() -> Self {
        Self::new()
    }
}