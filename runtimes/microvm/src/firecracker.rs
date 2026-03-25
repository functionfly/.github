//! Firecracker API client for MicroVM management

use anyhow::{Context, Result};
use reqwest::Client;
use serde::{Deserialize, Serialize};
#[cfg(unix)]
use std::path::Path;
use std::path::PathBuf;
use std::time::Duration;
use tracing::{debug, info};
use uuid::Uuid;

/// Firecracker VM configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VMConfig {
    /// Number of vCPUs
    pub vcpu_count: u32,
    /// Memory in MiB
    pub mem_size_mib: u32,
    /// Kernel image path
    pub kernel_image: PathBuf,
    /// Initrd path (optional)
    pub initrd: Option<PathBuf>,
    /// Root drive image
    pub root_drive: PathBuf,
    /// Network interface
    pub network_interface: NetworkInterface,
    /// Skip the network-interface PUT to Firecracker (no-network VMs or tap not available).
    /// Controlled by `FIRECRACKER_SKIP_NETWORK=1`.
    #[serde(default)]
    pub skip_network: bool,
}

/// Network interface configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NetworkInterface {
    /// Guest MAC address
    pub guest_mac: String,
    /// Host tap device name
    pub host_dev_name: String,
}

/// Firecracker boot source configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct BootSource {
    pub kernel_image_path: String,
    pub initrd_path: Option<String>,
    pub boot_args: String,
}

/// Firecracker machine configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct MachineConfig {
    pub vcpu_count: u32,
    pub mem_size_mib: u32,
    pub smt: bool,
    pub track_dirty_pages: bool,
}

/// Firecracker drive configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Drive {
    pub drive_id: String,
    pub path_on_host: String,
    pub is_root_device: bool,
    pub is_read_only: bool,
}

/// Firecracker network interface configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct NetworkInterfaceConfig {
    pub iface_id: String,
    pub guest_mac: String,
    pub host_dev_name: String,
}

/// Firecracker balloon configuration for memory limiting
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct BalloonConfig {
    pub amount_mib: u32,
    pub deflate_on_oom: bool,
}

/// VM instance information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VMInstance {
    pub id: Uuid,
    pub state: VMState,
    pub config: VMConfig,
    pub vmm_socket_path: PathBuf,
    pub vsock_path: PathBuf,
    pub cid: u32, // Vsock Context ID for communication
    pub created_at: chrono::DateTime<chrono::Utc>,
}

/// VM state
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum VMState {
    NotStarted,
    Starting,
    Running,
    Stopping,
    Stopped,
    Failed,
}

impl Default for VMState {
    fn default() -> Self {
        VMState::NotStarted
    }
}

/// Firecracker API client
pub struct FirecrackerClient {
    client: Client,
    socket_path: String,
    /// Next available CID for vsock communication (starts at 3)
    next_cid: std::sync::atomic::AtomicU32,
}

impl FirecrackerClient {
    /// Create a new Firecracker API client (HTTP over the given Unix domain socket).
    pub fn new(socket_path: impl Into<String>) -> Result<Self> {
        let socket_path: String = socket_path.into();

        #[cfg(not(unix))]
        {
            anyhow::bail!("Firecracker API client requires Unix domain sockets");
        }

        #[cfg(unix)]
        {
            let client = Client::builder()
                .timeout(Duration::from_secs(30))
                .unix_socket(Path::new(&socket_path))
                .build()
                .context("Failed to build HTTP client")?;

            Ok(Self {
                client,
                socket_path,
                next_cid: std::sync::atomic::AtomicU32::new(3), // Firecracker starts CIDs from 3
            })
        }
    }

    /// Build the request URL for a Firecracker API resource (host is ignored; connection uses `unix_socket`).
    fn api_url(&self, endpoint: &str) -> String {
        let ep = endpoint.trim_start_matches('/');
        format!("http://localhost/{ep}")
    }

    /// Create a new MicroVM instance
    pub async fn create_vm(&self, config: &VMConfig) -> Result<VMInstance> {
        let vm_id = Uuid::new_v4();
        let vmm_socket = format!("/var/run/firecracker/{}.sock", vm_id);
        let vsock_path = format!("/var/run/firecracker/{}.vsock", vm_id);

        // Assign unique CID for vsock communication
        let cid = self.next_cid.fetch_add(1, std::sync::atomic::Ordering::SeqCst);

        info!("Creating VM instance: {} with CID {}", vm_id, cid);
        debug!("Firecracker API socket: {}", self.socket_path);

        // Configure boot source
        self.client
            .put(&self.api_url("boot-source"))
            .json(&BootSource {
                kernel_image_path: config.kernel_image.to_string_lossy().to_string(),
                initrd_path: config.initrd.as_ref().map(|p| p.to_string_lossy().to_string()),
                boot_args: "console=ttyS0 reboot=k panic=1".to_string(),
            })
            .send()
            .await
            .context("Failed to configure boot source")?;

        // Configure machine
        self.client
            .put(&self.api_url("machine-config"))
            .json(&MachineConfig {
                vcpu_count: config.vcpu_count,
                mem_size_mib: config.mem_size_mib,
                smt: false,
                track_dirty_pages: false,
            })
            .send()
            .await
            .context("Failed to configure machine")?;

        // Configure root drive
        self.client
            .put(&self.api_url("drives/rootfs"))
            .json(&Drive {
                drive_id: "rootfs".to_string(),
                path_on_host: config.root_drive.to_string_lossy().to_string(),
                is_root_device: true,
                is_read_only: false,
            })
            .send()
            .await
            .context("Failed to configure drive")?;

        // Configure network (skip if tap is not available or explicitly disabled)
        if !config.skip_network {
            self.client
                .put(&self.api_url("network-interfaces/eth0"))
                .json(&NetworkInterfaceConfig {
                    iface_id: "eth0".to_string(),
                    guest_mac: config.network_interface.guest_mac.clone(),
                    host_dev_name: config.network_interface.host_dev_name.clone(),
                })
                .send()
                .await
                .context("Failed to configure network")?;
            debug!("Network interface configured for VM {}", vm_id);
        } else {
            debug!("Skipping network configuration for VM {} (FIRECRACKER_SKIP_NETWORK)", vm_id);
        }

        debug!("VM instance {} configured successfully", vm_id);

        Ok(VMInstance {
            id: vm_id,
            state: VMState::NotStarted,
            config: config.clone(),
            vmm_socket_path: vmm_socket.into(),
            vsock_path: vsock_path.into(),
            cid,
            created_at: chrono::Utc::now(),
        })
    }

    /// Start a MicroVM instance
    pub async fn start_vm(&self, vm_id: &Uuid) -> Result<()> {
        info!("Starting VM instance: {}", vm_id);

        self.client
            .put(&self.api_url("actions"))
            .json(&serde_json::json!({
                "action_type": "InstanceStart"
            }))
            .send()
            .await
            .context("Failed to start VM")?;

        info!("VM instance {} started", vm_id);
        Ok(())
    }

    /// Stop a MicroVM instance
    pub async fn stop_vm(&self, vm_id: &Uuid) -> Result<()> {
        info!("Stopping VM instance: {}", vm_id);

        self.client
            .put(&self.api_url("actions"))
            .json(&serde_json::json!({
                "action_type": "InstanceStop"
            }))
            .send()
            .await
            .context("Failed to stop VM")?;

        info!("VM instance {} stopped", vm_id);
        Ok(())
    }

    /// Get VM info
    pub async fn get_vm_info(&self) -> Result<serde_json::Value> {
        let response = self.client
            .get(&self.api_url("vm"))
            .send()
            .await
            .context("Failed to get VM info")?;

        let info = response.json().await?;
        Ok(info)
    }

    /// Send stdin to the VM (for execution)
    pub async fn send_stdin(&self, vm_id: &Uuid, data: &str) -> Result<()> {
        // This would use a PTY or vsock for actual stdin
        debug!("Sending {} bytes to VM {}", data.len(), vm_id);
        Ok(())
    }

    /// Configure memory balloon for memory limiting
    pub async fn configure_balloon(&self, amount_mib: u32) -> Result<()> {
        self.client
            .put(&self.api_url("balloon"))
            .json(&BalloonConfig {
                amount_mib,
                deflate_on_oom: true,
            })
            .send()
            .await
            .context("Failed to configure balloon")?;

        Ok(())
    }
}

/// Generate a random MAC address for the VM
pub fn generate_mac() -> String {
    let bytes: [u8; 6] = rand::random();
    format!("02:{:02x}:{:02x}:{:02x}:{:02x}:{:02x}",
            bytes[1], bytes[2], bytes[3], bytes[4], bytes[5])
}

mod rand {
    use std::collections::hash_map::RandomState;
    use std::hash::{BuildHasher, Hasher};

    pub fn random() -> [u8; 6] {
        let state = RandomState::new();
        let mut hasher = state.build_hasher();
        hasher.write_u128(std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_nanos());

        let hash = hasher.finish();
        [
            (hash >> 40) as u8,
            (hash >> 32) as u8,
            (hash >> 24) as u8,
            (hash >> 16) as u8,
            (hash >> 8) as u8,
            hash as u8,
        ]
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_mac_generation() {
        let mac = generate_mac();
        assert!(mac.starts_with("02:"));
        assert_eq!(mac.len(), 17);
    }
}
