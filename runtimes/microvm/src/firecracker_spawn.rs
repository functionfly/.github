//! Spawn a Firecracker process with a dedicated API Unix socket (one process per MicroVM).

use anyhow::{Context, Result};
use std::path::{Path, PathBuf};
use std::process::{Child, Command, Stdio};
use std::time::Duration;
use tracing::{debug, warn};

/// Path to the Firecracker binary (`FIRECRACKER_BINARY` or `firecracker` on `PATH`).
pub fn firecracker_binary() -> String {
    std::env::var("FIRECRACKER_BINARY").unwrap_or_else(|_| "firecracker".to_string())
}

/// Start Firecracker listening on `api_sock`. The socket file is created by Firecracker when it starts.
pub fn spawn_firecracker(api_sock: &Path) -> Result<Child> {
    let bin = firecracker_binary();
    // Remove stale socket from a previous crash
    if api_sock.exists() {
        let _ = std::fs::remove_file(api_sock);
    }
    let child = Command::new(&bin)
        .arg("--api-sock")
        .arg(api_sock)
        .stdin(Stdio::null())
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .spawn()
        .with_context(|| format!("failed to spawn Firecracker ({bin}); install firecracker or set FIRECRACKER_BINARY"))?;
    debug!("Spawned Firecracker api_sock={}", api_sock.display());
    Ok(child)
}

/// Wait until the API socket exists (Firecracker creates it).
pub async fn wait_for_api_socket(path: &Path, timeout_ms: u64) -> Result<()> {
    let deadline = Duration::from_millis(timeout_ms);
    let start = std::time::Instant::now();
    while start.elapsed() < deadline {
        if path.exists() {
            return Ok(());
        }
        tokio::time::sleep(Duration::from_millis(50)).await;
    }
    anyhow::bail!(
        "Firecracker API socket did not appear within {}ms: {}",
        timeout_ms,
        path.display()
    )
}

/// Allocate a unique API socket path under the system temp directory.
pub fn allocate_api_socket_path() -> PathBuf {
    let id = uuid::Uuid::new_v4();
    std::env::temp_dir().join(format!("functionfly-fc-{id}.sock"))
}

/// Best-effort terminate Firecracker child.
pub fn kill_firecracker(mut child: Child) {
    if let Err(e) = child.kill() {
        warn!("Failed to kill Firecracker process: {}", e);
    }
    let _ = child.wait();
}
