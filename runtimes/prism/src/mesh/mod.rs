//! Mesh Networking
//!
//! P2P capability mesh using libp2p for:
//! - peer-to-peer communication
//! - capability discovery via Kademlia DHT
//! - local peer discovery via mDNS
//! - NAT traversal via relay
//!
//! Enables the "function mesh" where functions can:
//! - discover each other
//! - coordinate execution
//! - share state
//! - form temporary clusters

mod network;
mod peer;
mod protocol;
mod routing;

pub use network::{
    MeshNetwork, MeshConfig, MeshEvent, MeshPeerId,
    ConnectionState, MeshError, MeshMessage, Peer,
    MeshSwarm as Swarm, MeshBehaviour,
};
pub use protocol::{MeshProtocol, CapabilityAdvert, StateSync};
pub use routing::{RoutingTable, RoutingEntry};

// Re-export libp2p PeerId for external use
pub use libp2p::PeerId;