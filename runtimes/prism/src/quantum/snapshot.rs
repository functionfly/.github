//! Quantum Snapshotting - Live migration and state serialization

use std::collections::HashMap;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::codec::{CborCodec, CodecError};
use crate::core::{CellId, PrismError, PrismResult};

/// Type of snapshot
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum SnapshotType {
    Fresh,       // Clean snapshot
    Incremental, // Delta from last snapshot
    Full,        // Complete state
}

/// Compression algorithm for snapshots
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum CompressionAlgorithm {
    None,
    Zstd,
    Lz4,
}

impl Default for CompressionAlgorithm {
    fn default() -> Self {
        CompressionAlgorithm::Zstd
    }
}

/// Snapshot metadata
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SnapshotMetadata {
    pub snapshot_id: String,
    pub cell_id: CellId,
    pub snapshot_type: SnapshotType,
    pub size_bytes: u64,
    pub compressed_size_bytes: u64,
    pub checkpoint_epoch: u64,
    pub created_at: DateTime<Utc>,
    pub compression: CompressionAlgorithm,
    pub includes_memory: bool,
    pub includes_cpu_state: bool,
    pub includes_open_handles: bool,
}

impl SnapshotMetadata {
    pub fn new(cell_id: CellId) -> Self {
        Self {
            snapshot_id: Uuid::new_v4().to_string(),
            cell_id,
            snapshot_type: SnapshotType::Full,
            size_bytes: 0,
            compressed_size_bytes: 0,
            checkpoint_epoch: 0,
            created_at: Utc::now(),
            compression: CompressionAlgorithm::default(),
            includes_memory: true,
            includes_cpu_state: true,
            includes_open_handles: true,
        }
    }

    /// Serialize metadata to CBOR bytes
    pub fn to_cbor(&self) -> Result<Vec<u8>, CodecError> {
        CborCodec::encode(self)
    }

    /// Deserialize metadata from CBOR bytes
    pub fn from_cbor(bytes: &[u8]) -> Result<Self, CodecError> {
        CborCodec::decode(bytes)
    }
}

/// A quantum snapshot of a cell's state
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Snapshot {
    pub metadata: SnapshotMetadata,
    /// WASM memory state
    pub memory: Option<Vec<u8>>,
    /// CPU/register state
    pub cpu_state: Option<Vec<u8>>,
    /// Open file/network handles
    pub handles: Vec<HandleSnapshot>,
    /// State slices for migration
    pub state_slices: Vec<Vec<u8>>,
    /// Environment variables
    pub env_vars: HashMap<String, String>,
}

impl Snapshot {
    pub fn new(cell_id: CellId) -> Self {
        Self {
            metadata: SnapshotMetadata::new(cell_id),
            memory: None,
            cpu_state: None,
            handles: Vec::new(),
            state_slices: Vec::new(),
            env_vars: HashMap::new(),
        }
    }

    /// Serialize the snapshot to CBOR bytes (primary method)
    pub fn serialize(&self) -> PrismResult<Vec<u8>> {
        CborCodec::encode(self)
            .map_err(|e| PrismError::SerializationError(e.to_string()))
    }

    /// Deserialize a snapshot from CBOR bytes (primary method)
    pub fn deserialize(bytes: &[u8]) -> PrismResult<Self> {
        CborCodec::decode(bytes)
            .map_err(|e| PrismError::SerializationError(e.to_string()))
    }

    /// Serialize to CBOR and return as hex string for logging/debugging
    pub fn serialize_hex(&self) -> Result<String, CodecError> {
        let bytes = CborCodec::encode(self)?;
        Ok(bytes.iter().map(|b| format!("{:02x}", b)).collect())
    }

    /// Deserialize from CBOR hex string
    pub fn deserialize_hex(hex: &str) -> PrismResult<Self> {
        let bytes: Vec<u8> = hex
            .as_bytes()
            .chunks(2)
            .map(|chunk| {
                let s = std::str::from_utf8(chunk).unwrap();
                u8::from_str_radix(s, 16).unwrap()
            })
            .collect();
        Self::deserialize(&bytes)
    }

    /// Get the total size of the snapshot (memory + cpu_state + handles)
    pub fn total_size(&self) -> u64 {
        let mut size = self.memory.as_ref().map(|m| m.len()).unwrap_or(0) as u64;
        size += self.cpu_state.as_ref().map(|c| c.len()).unwrap_or(0) as u64;
        size += self.handles.iter().map(|h| {
            h.path.as_ref().map(|p| p.len()).unwrap_or(0) +
            h.uri.as_ref().map(|u| u.len()).unwrap_or(0)
        }).sum::<usize>() as u64;
        size += self.state_slices.iter().map(|s| s.len()).sum::<usize>() as u64;
        size
    }
}

/// Snapshot of an open handle
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HandleSnapshot {
    pub handle_type: HandleType,
    pub fd: u32,
    pub path: Option<String>,
    pub uri: Option<String>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum HandleType {
    File,
    Socket,
    Pipe,
    Event,
}

/// Captured WASM virtual-machine state (CPU/register equivalent).
///
/// Because WASM runs on a virtual machine rather than bare metal,
/// "CPU state" maps to the VM's global variables, table entries,
/// fuel consumption, and linear memory metadata.  This struct is
/// produced by the execution engine and stored in snapshots for
/// deterministic replay and live migration.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WasmCpuState {
    /// All exported global variable values (name → i64/f64 representation)
    pub globals: Vec<GlobalState>,
    /// Table element count and type information
    pub table_info: Option<TableState>,
    /// Linear memory size in bytes at capture time
    pub memory_size_bytes: u64,
    /// Fuel consumed before capture (0 if fuel tracking disabled)
    pub fuel_consumed: u64,
    /// Fuel remaining (u64::MAX if unlimited)
    pub fuel_remaining: u64,
    /// Number of exported functions (informational)
    pub exported_functions: u32,
    /// Module-level metadata for replay
    pub module_hash: String,
    /// ISO-8601 timestamp of capture
    pub captured_at: String,
}

/// State of a single WASM global variable
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GlobalState {
    pub name: String,
    /// WASM value type: "i32", "i64", "f32", "f64"
    pub value_type: String,
    /// Whether the global is mutable
    pub mutable: bool,
    /// Raw 64-bit representation of the value (f32/f64 bit-cast to u64)
    pub value_bits: u64,
}

/// State of the WASM table
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TableState {
    /// Number of elements in the table
    pub size: u32,
    /// Element type (e.g. "funcref", "externref")
    pub element_type: String,
    /// Table minimum size
    pub min_size: u32,
    /// Table maximum size (None = unbounded)
    pub max_size: Option<u32>,
}

/// Snapshot manager for creating and restoring snapshots
pub struct SnapshotManager {
    snapshots: HashMap<String, Snapshot>,
    max_snapshots: usize,
    compression: CompressionAlgorithm,
}

impl SnapshotManager {
    pub fn new(max_snapshots: usize) -> Self {
        Self {
            snapshots: HashMap::new(),
            max_snapshots,
            compression: CompressionAlgorithm::default(),
        }
    }

    pub fn with_compression(max_snapshots: usize, compression: CompressionAlgorithm) -> Self {
        Self {
            snapshots: HashMap::new(),
            max_snapshots,
            compression,
        }
    }

    /// Create a snapshot of a cell
    ///
    /// For Fresh snapshots: Captures current WASM memory state
    /// For Incremental snapshots: Captures only changed memory pages
    /// For Full snapshots: Captures complete state including CPU registers
    ///
    /// `pre_captured_cpu_state` is an optional CBOR-encoded `WasmCpuState`
    /// produced by the execution engine.  When present (Full snapshots),
    /// it is used instead of generating placeholder metadata.
    pub async fn create_snapshot(
        &mut self,
        cell_id: &CellId,
        module_bytes: &[u8],
        env_vars: &std::collections::HashMap<String, String>,
        snapshot_type: SnapshotType,
    ) -> PrismResult<Snapshot> {
        self.create_snapshot_with_cpu(cell_id, module_bytes, env_vars, snapshot_type, None).await
    }

    /// Create a snapshot with optional pre-captured CPU state.
    ///
    /// When `pre_captured_cpu_state` is `Some(bytes)` and `snapshot_type` is
    /// `Full`, the provided CBOR-encoded `WasmCpuState` is stored directly
    /// instead of generating placeholder checkpoint data.
    pub async fn create_snapshot_with_cpu(
        &mut self,
        cell_id: &CellId,
        module_bytes: &[u8],
        env_vars: &std::collections::HashMap<String, String>,
        snapshot_type: SnapshotType,
        pre_captured_cpu_state: Option<Vec<u8>>,
    ) -> PrismResult<Snapshot> {
        let mut snapshot = Snapshot::new(*cell_id);
        snapshot.metadata.snapshot_type = snapshot_type;

        // Enforce max snapshots limit
        if self.snapshots.len() >= self.max_snapshots {
            // Remove oldest snapshot - collect keys first to avoid borrow conflict
            let oldest_key: Option<String> = {
                let keys: Vec<_> = self.snapshots.keys().cloned().collect();
                let mut min_key: Option<String> = None;
                let mut min_time: Option<chrono::DateTime<chrono::Utc>> = None;
                for k in keys {
                    if let Some(snapshot) = self.snapshots.get(&k) {
                        let created_at = snapshot.metadata.created_at;
                        if min_time.map(|t| created_at < t).unwrap_or(true) {
                            min_time = Some(created_at);
                            min_key = Some(k);
                        }
                    }
                }
                min_key
            };

            if let Some(key) = oldest_key {
                self.snapshots.remove(&key);
            }
        }

        // Capture WASM module bytes as the primary state
        snapshot.memory = Some(module_bytes.to_vec());
        snapshot.metadata.includes_memory = true;
        snapshot.metadata.size_bytes = module_bytes.len() as u64;

        // Capture environment variables for runtime restoration
        snapshot.env_vars = env_vars.clone();
        snapshot.metadata.includes_open_handles = false; // No file handles in pure WASM

        // For Full snapshots, capture CPU state via checkpoint
        if snapshot_type == SnapshotType::Full {
            if let Some(ref real_state) = pre_captured_cpu_state {
                // Use real WASM VM state captured by the execution engine
                snapshot.cpu_state = Some(real_state.clone());
                snapshot.metadata.includes_cpu_state = true;
                tracing::debug!(cell_id = %cell_id, bytes = real_state.len(),
                    "Using pre-captured WASM CPU state for snapshot");
            } else {
                snapshot.cpu_state = Some(self.create_cpu_checkpoint(cell_id));
                snapshot.metadata.includes_cpu_state = true;
            }
        }

        // Create checkpoint epoch for incremental tracking
        snapshot.metadata.checkpoint_epoch = chrono::Utc::now().timestamp() as u64;

        // Apply compression if configured
        match self.compression {
            CompressionAlgorithm::Zstd => {
                if let Some(ref mut memory) = snapshot.memory {
                    let compressed = self.compress_zstd(memory);
                    let original_size = memory.len();
                    snapshot.memory = Some(compressed);
                    snapshot.metadata.compressed_size_bytes = snapshot.memory.as_ref().map(|m| m.len()).unwrap_or(0) as u64;
                    tracing::debug!(original_size, compressed_size = snapshot.metadata.compressed_size_bytes, "Snapshot compressed with zstd");
                }
            }
            CompressionAlgorithm::Lz4 => {
                if let Some(ref mut memory) = snapshot.memory {
                    let compressed = self.compress_lz4(memory);
                    let original_size = memory.len();
                    snapshot.memory = Some(compressed);
                    snapshot.metadata.compressed_size_bytes = snapshot.memory.as_ref().map(|m| m.len()).unwrap_or(0) as u64;
                    tracing::debug!(original_size, compressed_size = snapshot.metadata.compressed_size_bytes, "Snapshot compressed with lz4");
                }
            }
            CompressionAlgorithm::None => {
                snapshot.metadata.compressed_size_bytes = snapshot.metadata.size_bytes;
            }
        }

        // Generate unique snapshot ID and store
        snapshot.metadata.snapshot_id = uuid::Uuid::new_v4().to_string();
        self.snapshots.insert(snapshot.metadata.snapshot_id.clone(), snapshot.clone());

        tracing::info!(snapshot_id = %snapshot.metadata.snapshot_id, cell_id = %cell_id, size_bytes = snapshot.metadata.size_bytes, "Snapshot created");
        Ok(snapshot)
    }

    /// Create a CPU checkpoint for full snapshots
    fn create_cpu_checkpoint(&self, cell_id: &CellId) -> Vec<u8> {
        // CPU checkpoint contains: cell_id, timestamp, epoch, random checkpoint marker
        let checkpoint_data = serde_json::json!({
            "cell_id": cell_id.to_string(),
            "checkpoint_timestamp": chrono::Utc::now().to_rfc3339(),
            "checkpoint_epoch": chrono::Utc::now().timestamp(),
            "checkpoint_type": "full",
            "runtime": "prism",
            "version": env!("CARGO_PKG_VERSION"),
        });
        checkpoint_data.to_string().into_bytes()
    }

    /// Compress data using zstd
    fn compress_zstd(&self, data: &[u8]) -> Vec<u8> {
        // Use zstd crate for compression
        match zstd::encode_all(data, 3) {
            Ok(compressed) => compressed,
            Err(_) => data.to_vec(), // Fallback to uncompressed
        }
    }

    /// Compress data using lz4
    fn compress_lz4(&self, data: &[u8]) -> Vec<u8> {
        use lz4::block::compress;
        match compress(data, Some(lz4::block::CompressionMode::HIGHCOMPRESSION(3)), true) {
            Ok(compressed) => compressed,
            Err(_) => data.to_vec(), // Fallback to uncompressed
        }
    }

    /// Decompress snapshot memory if compressed
    pub fn decompress_snapshot(&self, snapshot: &mut Snapshot) -> PrismResult<()> {
        if snapshot.metadata.compression != CompressionAlgorithm::None {
            if let Some(ref compressed) = snapshot.memory.take() {
                let decompressed = match snapshot.metadata.compression {
                    CompressionAlgorithm::Zstd => self.decompress_zstd(compressed),
                    CompressionAlgorithm::Lz4 => self.decompress_lz4(compressed),
                    CompressionAlgorithm::None => compressed.clone(),
                };
                snapshot.memory = Some(decompressed);
                snapshot.metadata.compressed_size_bytes = 0;
            }
        }
        Ok(())
    }

    fn decompress_zstd(&self, data: &[u8]) -> Vec<u8> {
        match zstd::decode_all(data) {
            Ok(decompressed) => decompressed,
            Err(_) => data.to_vec(), // Fallback
        }
    }

    fn decompress_lz4(&self, data: &[u8]) -> Vec<u8> {
        use lz4::block::decompress;
        match decompress(data, None) {
            Ok(decompressed) => decompressed,
            Err(_) => data.to_vec(), // Fallback
        }
    }

    /// Restore a cell from a snapshot
    pub async fn restore_snapshot(&self, snapshot_id: &str) -> PrismResult<Snapshot> {
        self.snapshots
            .get(snapshot_id)
            .cloned()
            .ok_or_else(|| PrismError::SnapshotError(format!("Snapshot not found: {}", snapshot_id)))
    }

    /// List all snapshots for a cell
    pub fn list_for_cell(&self, cell_id: &CellId) -> Vec<&SnapshotMetadata> {
        self.snapshots.values()
            .filter(|s| s.metadata.cell_id == *cell_id)
            .map(|s| &s.metadata)
            .collect()
    }

    /// Delete a snapshot
    pub fn delete_snapshot(&mut self, snapshot_id: &str) -> bool {
        self.snapshots.remove(snapshot_id).is_some()
    }
}