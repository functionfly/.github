//! Mesh networking for Prism using libp2p
//!
//! This module provides real P2P networking via libp2p with:
//! - TCP transport with noise encryption and yamux multiplexing
//! - Kademlia DHT for peer discovery and capability storage
//! - Gossipsub for capability advertisement broadcasting
//! - Ping for connection health checking

use std::collections::HashMap;
use std::num::NonZeroUsize;
use std::sync::Arc;
use std::sync::Mutex;
use std::time::{Duration, Instant};

use tokio::sync::RwLock;
use tokio::sync::broadcast;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

use libp2p::{
    PeerId, Multiaddr, identity::Keypair, Transport,
    core::muxing::StreamMuxerBox,
    core::upgrade::Version,
    swarm::{Swarm, SwarmEvent, Config},
    core::ConnectedPoint,
    multiaddr::Protocol,
};
use futures::StreamExt;

use crate::codec::{CborCodec, CodecError};

// ============================================================================
// Error Types
// ============================================================================

/// Mesh networking errors
#[derive(Debug, thiserror::Error)]
pub enum MeshError {
    #[error("Failed to create transport: {0}")]
    TransportError(String),
    #[error("Failed to parse address: {0}")]
    AddressError(String),
    #[error("Connection failed: {0}")]
    ConnectionError(String),
    #[error("Peer not found: {0}")]
    PeerNotFound(String),
    #[error("Swarm error: {0}")]
    SwarmError(String),
    #[error("Timeout: {0}")]
    Timeout(String),
    #[error("Invalid configuration: {0}")]
    ConfigError(String),
    #[error("Protocol error: {0}")]
    ProtocolError(String),
}

// ============================================================================
// Peer Types
// ============================================================================

/// Peer identifier (string-based wrapper for serialization)
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct MeshPeerId(pub String);

impl MeshPeerId {
    pub fn new(id: impl Into<String>) -> Self {
        Self(id.into())
    }

    pub fn as_str(&self) -> &str {
        &self.0
    }

    /// Serialize to CBOR bytes
    pub fn to_cbor(&self) -> Result<Vec<u8>, CodecError> {
        CborCodec::encode(self)
    }

    /// Deserialize from CBOR bytes
    pub fn from_cbor(bytes: &[u8]) -> Result<Self, CodecError> {
        CborCodec::decode(bytes)
    }

    /// Parse from base58 string
    pub fn from_base58(s: &str) -> Option<Self> {
        s.parse().ok().map(Self)
    }
}

impl std::fmt::Display for MeshPeerId {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl From<PeerId> for MeshPeerId {
    fn from(pid: PeerId) -> Self {
        MeshPeerId(pid.to_base58())
    }
}

impl From<MeshPeerId> for PeerId {
    fn from(mpid: MeshPeerId) -> Self {
        mpid.0.parse().unwrap_or_else(|_| PeerId::random())
    }
}

/// Connection state to a peer
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ConnectionState {
    Disconnected,
    Connecting,
    Connected,
    Handshake,
    Active,
}

/// A peer in the mesh network
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Peer {
    pub peer_id: MeshPeerId,
    pub connection_state: ConnectionState,
    pub address: Option<String>,
    pub region: Option<String>,
    pub advertised_capabilities: Vec<String>,
    pub connected_at: Option<DateTime<Utc>>,
    pub last_seen: DateTime<Utc>,
    pub latency_ms: Option<u32>,
}

impl Peer {
    pub fn new(peer_id: PeerId) -> Self {
        Self {
            peer_id: MeshPeerId(peer_id.to_base58()),
            connection_state: ConnectionState::Disconnected,
            address: None,
            region: None,
            advertised_capabilities: Vec::new(),
            connected_at: None,
            last_seen: Utc::now(),
            latency_ms: None,
        }
    }

    pub fn connect(&mut self, address: &str) {
        self.connection_state = ConnectionState::Connected;
        self.address = Some(address.to_string());
        self.connected_at = Some(Utc::now());
    }

    pub fn disconnect(&mut self) {
        self.connection_state = ConnectionState::Disconnected;
        self.connected_at = None;
    }

    pub fn is_connected(&self) -> bool {
        matches!(self.connection_state, ConnectionState::Connected | ConnectionState::Active)
    }

    /// Calculate availability score based on latency and uptime
    #[allow(clippy::assigning_clones)]
    pub fn availability_score(&self) -> f32 {
        let mut score = 0.5;

        match self.connection_state {
            ConnectionState::Active => score += 0.3,
            ConnectionState::Connected => score += 0.2,
            ConnectionState::Handshake => score += 0.1,
            _ => score -= 0.3,
        }

        if let Some(latency) = self.latency_ms {
            if latency < 50 {
                score += 0.2;
            } else if latency < 200 {
                score += 0.1;
            } else if latency > 500 {
                score -= 0.1;
            }
        }

        let final_score = if score > 1.0 { 1.0 } else if score < 0.0 { 0.0 } else { score };
        final_score
    }
}

// ============================================================================
// Configuration
// ============================================================================

/// Configuration for the mesh network
#[derive(Debug, Clone)]
pub struct MeshConfig {
    pub listen_addr: String,
    pub external_addr: Option<String>,
    pub keypair_path: Option<String>,
    pub bootstrap_nodes: Vec<String>,
    pub enable_relay: bool,
    pub enable_nat_detection: bool,
    pub connection_timeout_secs: u64,
    pub keep_alive_secs: u64,
    pub max_peers: usize,
    pub enable_kad_discovery: bool,
    pub ping_interval_secs: u64,
    /// Gossipsub config
    pub gossipsub_interval_secs: u64,
    /// Kademlia config
    pub kad_parallelism: u32,
    pub kad_query_timeout_secs: u64,
}

impl Default for MeshConfig {
    fn default() -> Self {
        Self {
            listen_addr: "/ip4/0.0.0.0/tcp/0".to_string(),
            external_addr: None,
            keypair_path: None,
            bootstrap_nodes: Vec::new(),
            enable_relay: true,
            enable_nat_detection: true,
            connection_timeout_secs: 30,
            keep_alive_secs: 60,
            max_peers: 1000,
            enable_kad_discovery: true,
            ping_interval_secs: 30,
            gossipsub_interval_secs: 20,
            kad_parallelism: 5,
            kad_query_timeout_secs: 30,
        }
    }
}

// ============================================================================
// Events
// ============================================================================

/// Event emitted by the mesh network
#[derive(Debug, Clone)]
pub enum MeshEvent {
    PeerConnected(MeshPeerId),
    PeerDisconnected(MeshPeerId),
    PeerAdvertisedCapabilities(MeshPeerId, Vec<String>),
    MessageReceived(MeshPeerId, Vec<u8>),
    Listening { addr: Multiaddr },
    DiscoveryComplete,
    PeerHealth { peer_id: MeshPeerId, latency_ms: Option<u32> },
    RoutingUpdated { peer_id: MeshPeerId },
    QueryResult { key: Vec<u8>, value: Option<Vec<u8>> },
}

/// Mesh message types for capability advertisement
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum MeshMessage {
    CapabilityAdv {
        peer_id: String,
        capabilities: Vec<String>,
        timestamp: i64,
    },
    DiscoveryReq { peer_id: String },
    DiscoveryResp {
        peer_id: String,
        capabilities: Vec<String>,
        addresses: Vec<String>,
    },
    Ping { nonce: u64 },
    Pong { nonce: u64 },
    AppMessage { data: Vec<u8> },
}

impl MeshMessage {
    pub fn encode(&self) -> Result<Vec<u8>, CodecError> {
        CborCodec::encode(self)
    }

    pub fn decode(bytes: &[u8]) -> Result<Self, CodecError> {
        CborCodec::decode(bytes)
    }
}

// ============================================================================
// Mesh Behaviour (NetworkBehaviour implementation)
// ============================================================================

/// Protocol constants
const CAPABILITY_TOPIC: &str = "functionfly:capabilities";
const DISCOVERY_TOPIC: &str = "functionfly:discovery";

/// Combined event from all sub-behaviours.
/// This is the event type returned by `MeshBehaviour::poll()`.
#[derive(Debug)]
pub enum MeshBehaviourEvent {
    Kademlia(libp2p::kad::Event),
    Gossipsub(libp2p::gossipsub::Event),
    Ping(libp2p::ping::Event),
}

impl From<libp2p::kad::Event> for MeshBehaviourEvent {
    fn from(event: libp2p::kad::Event) -> Self {
        MeshBehaviourEvent::Kademlia(event)
    }
}

impl From<libp2p::gossipsub::Event> for MeshBehaviourEvent {
    fn from(event: libp2p::gossipsub::Event) -> Self {
        MeshBehaviourEvent::Gossipsub(event)
    }
}

impl From<libp2p::ping::Event> for MeshBehaviourEvent {
    fn from(event: libp2p::ping::Event) -> Self {
        MeshBehaviourEvent::Ping(event)
    }
}

/// The main NetworkBehaviour for mesh networking.
/// Combines Kademlia (DHT), Gossipsub (pubsub), and Ping protocols
/// using the `#[derive(NetworkBehaviour)]` macro which handles handler
/// composition automatically.
#[derive(libp2p::swarm::NetworkBehaviour)]
#[behaviour(to_swarm = "MeshBehaviourEvent")]
pub struct MeshBehaviour {
    /// Kademlia DHT for peer discovery and capability storage
    kad: libp2p::kad::Behaviour<libp2p::kad::store::MemoryStore>,
    /// Gossipsub for capability broadcasting
    gossipsub: libp2p::gossipsub::Behaviour,
    /// Ping protocol for health checks
    ping: libp2p::ping::Behaviour,
}

impl MeshBehaviour {
    /// Create a new MeshBehaviour with proper initialization
    pub fn new(local_key: Keypair, config: &MeshConfig) -> Result<Self, Box<dyn std::error::Error + Send + Sync>> {
        let local_peer_id = PeerId::from(local_key.public());

        // Create Kademlia with memory store
        let mut kad_config = libp2p::kad::Config::new(libp2p::swarm::StreamProtocol::new("/functionfly/kad/1.0.0"));
        kad_config
            .set_parallelism(
                NonZeroUsize::new(config.kad_parallelism as usize)
                    .unwrap_or(NonZeroUsize::new(5).unwrap()),
            )
            .set_query_timeout(Duration::from_secs(config.kad_query_timeout_secs));

        let kad = libp2p::kad::Behaviour::with_config(
            local_peer_id,
            libp2p::kad::store::MemoryStore::new(local_peer_id),
            kad_config,
        );

        // Configure Gossipsub
        let gossipsub_config = libp2p::gossipsub::ConfigBuilder::default()
            .heartbeat_interval(Duration::from_secs(config.gossipsub_interval_secs))
            .history_length(10)
            .max_transmit_size(1024 * 1024) // 1MB
            .validation_mode(libp2p::gossipsub::ValidationMode::Strict)
            .build()
            .map_err(|e| Box::new(std::io::Error::new(std::io::ErrorKind::InvalidInput, e)) as Box<dyn std::error::Error + Send + Sync>)?;

        let gossipsub = libp2p::gossipsub::Behaviour::new(
            libp2p::gossipsub::MessageAuthenticity::Signed(local_key),
            gossipsub_config,
        )
        .map_err(|e| Box::new(std::io::Error::new(std::io::ErrorKind::InvalidInput, e)) as Box<dyn std::error::Error + Send + Sync>)?;

        // Configure Ping
        let ping = libp2p::ping::Behaviour::new(
            libp2p::ping::Config::new()
                .with_interval(Duration::from_secs(config.ping_interval_secs))
                .with_timeout(Duration::from_secs(10)),
        );

        Ok(Self {
            kad,
            gossipsub,
            ping,
        })
    }

    /// Bootstrap to a known peer
    pub fn bootstrap(&mut self, peer_id: PeerId, address: Multiaddr) {
        self.kad.add_address(&peer_id, address);
        let _ = self.kad.bootstrap();
    }

    /// Get all known peer IDs from Kademlia routing table
    pub fn known_peers(&mut self) -> Vec<PeerId> {
        self.kad
            .kbuckets()
            .flat_map(|kbucket| {
                kbucket.iter()
                    .map(|entry| *entry.node.key.preimage())
                    .collect::<Vec<_>>()
            })
            .collect()
    }

    /// Query for a specific key in the DHT
    pub fn query(&mut self, key: Vec<u8>) {
        use libp2p::kad::RecordKey;
        let record_key = RecordKey::new(&key);
        self.kad.get_record(record_key);
    }

    /// Put a record into the DHT
    pub fn put_record(&mut self, key: Vec<u8>, value: Vec<u8>) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        use libp2p::kad::{Record, RecordKey};
        let record = Record::new(RecordKey::new(&key), value);
        self.kad.put_record(record, libp2p::kad::Quorum::Majority)?;
        Ok(())
    }

    /// Store a capability in the DHT
    pub fn store_capability(&mut self, capability: &str, peer_id: PeerId) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        use libp2p::kad::{Record, RecordKey};
        let key = format!("capability:{}", capability);
        let value = serde_json::to_vec(&serde_json::json!({
            "peer_id": peer_id.to_base58(),
            "capability": capability,
            "timestamp": Utc::now().timestamp(),
        }))?;
        let record = Record::new(RecordKey::new(&key), value);
        self.kad.put_record(record, libp2p::kad::Quorum::Majority)?;
        Ok(())
    }

    /// Subscribe to a gossipsub topic
    pub fn subscribe(&mut self, topic: &str) -> Result<bool, Box<dyn std::error::Error + Send + Sync>> {
        let topic = libp2p::gossipsub::IdentTopic::new(topic);
        let result = self.gossipsub.subscribe(&topic)?;
        Ok(result)
    }

    /// Publish to a gossipsub topic
    pub fn publish(&mut self, topic: &str, data: Vec<u8>) -> Result<libp2p::gossipsub::MessageId, Box<dyn std::error::Error + Send + Sync>> {
        let topic = libp2p::gossipsub::IdentTopic::new(topic);
        let msg_id = self.gossipsub.publish(topic, data)?;
        Ok(msg_id)
    }

    /// Advertise capabilities via gossipsub
    pub fn advertise_capabilities(&mut self, capabilities: Vec<String>, local_peer_id: PeerId) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        let message = MeshMessage::CapabilityAdv {
            peer_id: local_peer_id.to_base58(),
            capabilities,
            timestamp: Utc::now().timestamp(),
        };
        let encoded = serde_json::to_vec(&message)?;
        self.publish(CAPABILITY_TOPIC, encoded)?;
        Ok(())
    }

    /// Handle incoming capability advertisement
    pub fn handle_capability_advert(&self, source: PeerId, data: &[u8]) -> Option<Vec<String>> {
        if let Ok(msg) = serde_json::from_slice::<MeshMessage>(data) {
            match msg {
                MeshMessage::CapabilityAdv { capabilities, .. } => {
                    tracing::debug!("Received capability advert from {:?}: {:?}", source, capabilities);
                    return Some(capabilities);
                }
                _ => {}
            }
        }
        None
    }
}

// ============================================================================
// Mesh Network (Main API)
// ============================================================================

/// Mesh networking layer using libp2p
pub struct MeshNetwork {
    config: MeshConfig,
    local_keypair: Keypair,
    local_peer_id: PeerId,
    peers: Arc<RwLock<HashMap<MeshPeerId, Peer>>>,
    local_capabilities: Arc<RwLock<Vec<String>>>,
    event_tx: broadcast::Sender<MeshEvent>,
    event_rx: Arc<RwLock<broadcast::Receiver<MeshEvent>>>,
    swarm: Arc<Mutex<Option<Swarm<MeshBehaviour>>>>,
    shutdown: Arc<RwLock<bool>>,
    routing_table: Arc<RwLock<HashMap<MeshPeerId, Vec<Multiaddr>>>>,
    pending_pings: Arc<RwLock<HashMap<u64, Instant>>>,
}

impl MeshNetwork {
    pub fn new(config: MeshConfig) -> Result<Self, MeshError> {
        let local_keypair = Self::load_or_generate_keypair(&config)?;
        let local_peer_id = PeerId::from(local_keypair.public());
        let (event_tx, event_rx) = broadcast::channel(100);

        Ok(Self {
            config,
            local_keypair,
            local_peer_id,
            peers: Arc::new(RwLock::new(HashMap::new())),
            local_capabilities: Arc::new(RwLock::new(Vec::new())),
            event_tx,
            event_rx: Arc::new(RwLock::new(event_rx)),
            swarm: Arc::new(Mutex::new(None)),
            shutdown: Arc::new(RwLock::new(false)),
            routing_table: Arc::new(RwLock::new(HashMap::new())),
            pending_pings: Arc::new(RwLock::new(HashMap::new())),
        })
    }

    fn load_or_generate_keypair(config: &MeshConfig) -> Result<Keypair, MeshError> {
        let keypair = if let Some(ref key_path) = config.keypair_path {
            let key_bytes = std::fs::read(key_path)
                .map_err(|e| MeshError::ConfigError(format!("Failed to read keypair: {}", e)))?;
            Keypair::from_protobuf_encoding(&key_bytes)
                .map_err(|e| MeshError::ConfigError(format!("Invalid keypair format: {}", e)))?
        } else {
            Keypair::generate_ed25519()
        };
        Ok(keypair)
    }

    /// Start the mesh network with full libp2p stack
    pub async fn start(&self) -> Result<(), MeshError> {
        if *self.shutdown.read().await {
            return Err(MeshError::ConfigError("Network already shut down".into()));
        }

        // Create transport with noise encryption and yamux multiplexing
        let transport = libp2p::tcp::tokio::Transport::default()
            .upgrade(Version::V1Lazy)
            .authenticate(
                libp2p::noise::Config::new(&self.local_keypair)
                    .map_err(|e| MeshError::TransportError(e.to_string()))?,
            )
            .multiplex(libp2p::yamux::Config::default())
            .map(|(peer_id, muxer), _| (peer_id, StreamMuxerBox::new(muxer)))
            .boxed();

        // Create mesh behaviour with Kademlia, Gossipsub, and Ping
        let mut behaviour = MeshBehaviour::new(self.local_keypair.clone(), &self.config)
            .map_err(|e| MeshError::SwarmError(e.to_string()))?;

        // Subscribe to capability topic
        if let Err(e) = behaviour.subscribe(CAPABILITY_TOPIC) {
            tracing::warn!("Failed to subscribe to capability topic: {:?}", e);
        }

        // Subscribe to discovery topic
        if let Err(e) = behaviour.subscribe(DISCOVERY_TOPIC) {
            tracing::warn!("Failed to subscribe to discovery topic: {:?}", e);
        }

        let mut swarm = Swarm::new(
            transport,
            behaviour,
            self.local_peer_id,
            Config::with_tokio_executor(),
        );

        let listen_addr: Multiaddr = self.config.listen_addr.parse()
            .map_err(|e: libp2p::multiaddr::Error| MeshError::AddressError(e.to_string()))?;

        swarm.listen_on(listen_addr)
            .map_err(|e| MeshError::SwarmError(format!("{:?}", e)))?;

        // Dial bootstrap nodes
        for bootstrap_addr in &self.config.bootstrap_nodes {
            if let Ok(addr) = bootstrap_addr.parse::<Multiaddr>() {
                if let Some(peer_id) = addr.iter().find_map(|p| match p {
                    Protocol::P2p(pid) => Some(pid),
                    _ => None,
                }) {
                    let addr_without_p2p: Multiaddr = addr.iter()
                        .filter(|p| !matches!(p, Protocol::P2p(_)))
                        .collect();
                    let addr_clone = addr_without_p2p.clone();
                    if let Err(e) = swarm.dial(addr_without_p2p) {
                        tracing::warn!("Failed to dial bootstrap node: {:?}", e);
                    }
                    let mut routing = self.routing_table.write().await;
                    routing.entry(MeshPeerId::from(peer_id)).or_default().push(addr_clone);
                } else if let Err(e) = swarm.dial(addr) {
                    tracing::warn!("Failed to dial bootstrap node: {:?}", e);
                }
            }
        }

        *self.swarm.lock().unwrap() = Some(swarm);

        // Start event loop
        let event_tx = self.event_tx.clone();
        let peers = self.peers.clone();
        let routing_table = self.routing_table.clone();
        let pending_pings = self.pending_pings.clone();
        let shutdown = self.shutdown.clone();
        let swarm_arc = self.swarm.clone();
        let local_capabilities = self.local_capabilities.clone();

        tokio::spawn(async move {
            let mut _ping_nonce: u64 = 0;

            loop {
                if *shutdown.read().await {
                    break;
                }

                let swarm_opt = swarm_arc.lock().unwrap().take();

                if let Some(mut swarm_instance) = swarm_opt {
                    let event = tokio::select! {
                        biased;
                        event = swarm_instance.next() => event,
                        _ = tokio::time::sleep(Duration::from_millis(100)) => None,
                    };

                    *swarm_arc.lock().unwrap() = Some(swarm_instance);

                    if let Some(event) = event {
                        Self::handle_swarm_event(
                            event,
                            &peers,
                            &routing_table,
                            &pending_pings,
                            &mut _ping_nonce,
                            &event_tx,
                            &local_capabilities,
                        ).await;
                    }
                } else {
                    tokio::time::sleep(Duration::from_millis(100)).await;
                    if swarm_arc.lock().unwrap().is_none() {
                        tracing::warn!("Swarm not available, mesh network may be shut down");
                        break;
                    }
                }
            }
        });

        Ok(())
    }

    async fn handle_swarm_event(
        event: SwarmEvent<MeshBehaviourEvent>,
        peers: &Arc<RwLock<HashMap<MeshPeerId, Peer>>>,
        routing_table: &Arc<RwLock<HashMap<MeshPeerId, Vec<Multiaddr>>>>,
        _pending_pings: &Arc<RwLock<HashMap<u64, Instant>>>,
        _ping_nonce: &mut u64,
        event_tx: &broadcast::Sender<MeshEvent>,
        _local_capabilities: &Arc<RwLock<Vec<String>>>,
    ) {
        match event {
            SwarmEvent::NewListenAddr { address, .. } => {
                tracing::info!("Mesh listening on {:?}", address);
                let _ = event_tx.send(MeshEvent::Listening { addr: address });
            }
            SwarmEvent::ConnectionEstablished { peer_id, endpoint, .. } => {
                let mesh_peer_id = MeshPeerId::from(peer_id);
                let addr = match endpoint {
                    ConnectedPoint::Dialer { address, .. } => Some(address.to_string()),
                    ConnectedPoint::Listener { local_addr, .. } => Some(local_addr.to_string()),
                };

                let mut peers_lock = peers.write().await;
                if let Some(peer) = peers_lock.get_mut(&mesh_peer_id) {
                    peer.connection_state = ConnectionState::Connected;
                    peer.address = addr;
                } else {
                    let mut new_peer = Peer::new(peer_id);
                    new_peer.connection_state = ConnectionState::Connected;
                    new_peer.address = addr;
                    peers_lock.insert(mesh_peer_id.clone(), new_peer);
                }
                let _ = event_tx.send(MeshEvent::PeerConnected(mesh_peer_id));
            }
            SwarmEvent::ConnectionClosed { peer_id, .. } => {
                let mesh_peer_id = MeshPeerId::from(peer_id);
                let mut peers_lock = peers.write().await;
                if let Some(peer) = peers_lock.get_mut(&mesh_peer_id) {
                    peer.disconnect();
                }
                let _ = event_tx.send(MeshEvent::PeerDisconnected(mesh_peer_id));
            }
            SwarmEvent::OutgoingConnectionError { peer_id, error, .. } => {
                tracing::warn!("Outbound connection error to {:?}: {:?}", peer_id, error);
            }
            SwarmEvent::IncomingConnectionError { error, .. } => {
                tracing::warn!("Inbound connection error: {:?}", error);
            }
            SwarmEvent::Dialing { .. } => {
                // Dialing initiated, no action needed
            }
            SwarmEvent::Behaviour(MeshBehaviourEvent::Kademlia(kad_event)) => {
                match kad_event {
                    libp2p::kad::Event::RoutingUpdated { peer, .. } => {
                        let mesh_peer_id = MeshPeerId::from(peer);
                        let mut routing = routing_table.write().await;
                        if !routing.contains_key(&mesh_peer_id) {
                            routing.insert(mesh_peer_id.clone(), Vec::new());
                        }
                        let _ = event_tx.send(MeshEvent::RoutingUpdated { peer_id: mesh_peer_id });
                    }
                    libp2p::kad::Event::OutboundQueryProgressed { result, .. } => {
                        use libp2p::kad::QueryResult;
                        match result {
                            QueryResult::GetRecord(Ok(ok)) => {
                                if let libp2p::kad::GetRecordOk::FoundRecord(peer_record) = ok {
                                    let key = peer_record.record.key.to_vec();
                                    let value = Some(peer_record.record.value);
                                    let _ = event_tx.send(MeshEvent::QueryResult { key, value });
                                }
                            }
                            QueryResult::GetRecord(Err(err)) => {
                                tracing::debug!("GetRecord query failed: {:?}", err);
                            }
                            QueryResult::PutRecord(Ok(ok)) => {
                                tracing::debug!("PutRecord succeeded: {:?}", ok.key);
                            }
                            QueryResult::PutRecord(Err(err)) => {
                                tracing::warn!("PutRecord failed: {:?}", err);
                            }
                            _ => {}
                        }
                    }
                    _ => {}
                }
            }
            SwarmEvent::Behaviour(MeshBehaviourEvent::Gossipsub(gossip_event)) => {
                match gossip_event {
                    libp2p::gossipsub::Event::Message { propagation_source, message, .. } => {
                        let mesh_peer_id = MeshPeerId::from(propagation_source);
                        let _ = event_tx.send(MeshEvent::MessageReceived(mesh_peer_id, message.data));
                    }
                    libp2p::gossipsub::Event::Subscribed { peer_id, topic } => {
                        tracing::debug!("Peer {:?} subscribed to topic {:?}", peer_id, topic);
                    }
                    libp2p::gossipsub::Event::Unsubscribed { peer_id, topic } => {
                        tracing::debug!("Peer {:?} unsubscribed from topic {:?}", peer_id, topic);
                    }
                    libp2p::gossipsub::Event::GossipsubNotSupported { peer_id } => {
                        tracing::warn!("Peer {:?} does not support gossipsub", peer_id);
                    }
                }
            }
            SwarmEvent::Behaviour(MeshBehaviourEvent::Ping(ping_event)) => {
                let mesh_peer_id = MeshPeerId::from(ping_event.peer);
                match ping_event.result {
                    Ok(latency) => {
                        let latency_ms = Some(latency.as_millis() as u32);
                        let mut peers_lock = peers.write().await;
                        if let Some(peer) = peers_lock.get_mut(&mesh_peer_id) {
                            peer.latency_ms = latency_ms;
                            peer.last_seen = Utc::now();
                        }
                        let _ = event_tx.send(MeshEvent::PeerHealth { peer_id: mesh_peer_id, latency_ms });
                    }
                    Err(e) => {
                        tracing::debug!("Ping failed to {:?}: {:?}", mesh_peer_id, e);
                        let mut peers_lock = peers.write().await;
                        if let Some(peer) = peers_lock.get_mut(&mesh_peer_id) {
                            peer.latency_ms = None;
                        }
                        let _ = event_tx.send(MeshEvent::PeerHealth { peer_id: mesh_peer_id, latency_ms: None });
                    }
                }
            }
            _ => {}
        }
    }

    pub async fn subscribe(&self) -> broadcast::Receiver<MeshEvent> {
        self.event_rx.write().await.resubscribe()
    }

    pub async fn dial(&self, address: &str) -> Result<(), MeshError> {
        let mut swarm_guard = self.swarm.lock().unwrap();
        let swarm = swarm_guard.as_mut()
            .ok_or_else(|| MeshError::SwarmError("Swarm not initialized".into()))?;

        let addr: Multiaddr = address.parse()
            .map_err(|e: libp2p::multiaddr::Error| MeshError::AddressError(e.to_string()))?;

        swarm.dial(addr)
            .map_err(|e| MeshError::ConnectionError(e.to_string()))?;

        Ok(())
    }

    pub async fn dial_peer(&self, peer_id: PeerId) -> Result<(), MeshError> {
        let mut swarm_guard = self.swarm.lock().unwrap();
        let swarm = swarm_guard.as_mut()
            .ok_or_else(|| MeshError::SwarmError("Swarm not initialized".into()))?;

        swarm.dial(peer_id)
            .map_err(|e| MeshError::ConnectionError(e.to_string()))?;

        Ok(())
    }

    pub async fn send_to(&self, _peer_id: &MeshPeerId, _message: Vec<u8>) -> Result<(), MeshError> {
        // This would require opening a stream to the peer
        // For now, use broadcast instead
        Err(MeshError::ProtocolError("Direct send not yet implemented. Use broadcast() instead.".into()))
    }

    pub async fn broadcast(&self, message: Vec<u8>) -> Result<(), MeshError> {
        let mut swarm_guard = self.swarm.lock().unwrap();
        let swarm = swarm_guard.as_mut()
            .ok_or_else(|| MeshError::SwarmError("Swarm not initialized".into()))?;

        swarm.behaviour_mut().publish(DISCOVERY_TOPIC, message)
            .map_err(|e| MeshError::ProtocolError(e.to_string()))?;

        Ok(())
    }

    pub async fn advertise_capabilities(&self, capabilities: Vec<String>) {
        let mut local = self.local_capabilities.write().await;
        *local = capabilities.clone();

        let mut swarm_guard = self.swarm.lock().unwrap();
        if let Some(swarm) = swarm_guard.as_mut() {
            if let Err(e) = swarm.behaviour_mut().advertise_capabilities(capabilities, self.local_peer_id) {
                tracing::warn!("Failed to advertise capabilities: {:?}", e);
            }
        }
    }

    /// Store a capability in the DHT for discovery
    pub async fn store_capability(&self, capability: &str) -> Result<(), MeshError> {
        let mut swarm_guard = self.swarm.lock().unwrap();
        let swarm = swarm_guard.as_mut()
            .ok_or_else(|| MeshError::SwarmError("Swarm not initialized".into()))?;

        swarm.behaviour_mut().store_capability(capability, self.local_peer_id)
            .map_err(|e| MeshError::ProtocolError(e.to_string()))?;

        Ok(())
    }

    /// Query for peers with a specific capability via DHT
    pub async fn query_capability(&self, capability: &str) {
        let key = format!("capability:{}", capability).into_bytes();
        let mut swarm_guard = self.swarm.lock().unwrap();
        if let Some(swarm) = swarm_guard.as_mut() {
            swarm.behaviour_mut().query(key);
        }
    }

    pub async fn shutdown(&self) -> Result<(), MeshError> {
        *self.shutdown.write().await = true;

        let mut swarm_guard = self.swarm.lock().unwrap();
        if let Some(mut swarm) = swarm_guard.take() {
            let peer_ids: Vec<PeerId> = swarm.connected_peers().cloned().collect();
            for peer_id in peer_ids {
                let _ = swarm.disconnect_peer_id(peer_id);
            }
        }

        Ok(())
    }

    pub fn local_peer_id(&self) -> &PeerId {
        &self.local_peer_id
    }

    pub fn local_peer_id_str(&self) -> String {
        self.local_peer_id.to_base58()
    }

    pub fn config(&self) -> &MeshConfig {
        &self.config
    }

    pub fn max_peers(&self) -> usize {
        self.config.max_peers
    }

    pub fn connection_timeout(&self) -> Duration {
        Duration::from_secs(self.config.connection_timeout_secs)
    }

    pub fn keep_alive_interval(&self) -> Duration {
        Duration::from_secs(self.config.keep_alive_secs)
    }

    pub fn is_relay_enabled(&self) -> bool {
        self.config.enable_relay
    }

    pub fn is_nat_detection_enabled(&self) -> bool {
        self.config.enable_nat_detection
    }

    pub fn listen_address(&self) -> &str {
        &self.config.listen_addr
    }

    pub fn external_address(&self) -> Option<&str> {
        self.config.external_addr.as_deref()
    }

    pub async fn add_peer(&self, peer: Peer) {
        let mut peers = self.peers.write().await;
        peers.insert(peer.peer_id.clone(), peer);
    }

    pub async fn remove_peer(&self, peer_id: &MeshPeerId) -> bool {
        let mut peers = self.peers.write().await;
        peers.remove(peer_id).is_some()
    }

    pub async fn get_peer(&self, peer_id: &MeshPeerId) -> Option<Peer> {
        let peers = self.peers.read().await;
        peers.get(peer_id).cloned()
    }

    pub async fn connected_peers(&self) -> Vec<Peer> {
        let peers = self.peers.read().await;
        peers.values().filter(|p| p.is_connected()).cloned().collect()
    }

    pub async fn find_peers_with_capability(&self, capability: &str) -> Vec<Peer> {
        let peers = self.peers.read().await;
        let local = self.local_capabilities.read().await;
        peers.values()
            .filter(|p| {
                p.is_connected() &&
                (p.advertised_capabilities.contains(&capability.to_string()) ||
                 local.contains(&capability.to_string()))
            })
            .cloned()
            .collect()
    }

    pub async fn peer_count(&self) -> usize {
        let peers = self.peers.read().await;
        peers.len()
    }

    pub fn is_running(&self) -> bool {
        self.swarm.lock().unwrap().is_some()
    }

    pub async fn get_routing_table(&self) -> HashMap<MeshPeerId, Vec<Multiaddr>> {
        self.routing_table.read().await.clone()
    }

    /// Get all known peers from Kademlia
    pub async fn get_known_peers(&self) -> Vec<PeerId> {
        let mut swarm_guard = self.swarm.lock().unwrap();
        if let Some(swarm) = swarm_guard.as_mut() {
            swarm.behaviour_mut().known_peers()
        } else {
            Vec::new()
        }
    }
}

impl Default for MeshNetwork {
    fn default() -> Self {
        Self::new(MeshConfig::default()).expect("Default config should always be valid")
    }
}

// Re-export for external use
#[allow(unused_imports)]
pub use libp2p::swarm::Swarm as MeshSwarm;
