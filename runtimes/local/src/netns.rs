//! Linux network namespace isolation for the runtime process.
//!
//! When enabled (`--enable-net-ns`), this module creates a new network namespace
//! for the runtime process using `unshare(CLONE_NEWNET)`. This provides kernel-level
//! network isolation — the process can only see the loopback interface and any
//! explicitly configured routes.
//!
//! # Requirements
//!
//! - Linux kernel 3.8+
//! - CAP_NET_ADMIN capability (root or `--cap-add NET_ADMIN` in Docker)
//! - iptables/nftables for egress filtering (not implemented here — the namespace
//!   provides isolation by default since only loopback is available)
//!
//! # Platform
//!
//! This is a no-op on non-Linux platforms.

/// Apply network namespace isolation by calling `unshare(CLONE_NEWNET)`.
///
/// After this call, the process only has access to the loopback interface.
/// All external network access requires explicit iptables/nftables rules or
/// veth pair setup, which should be done by the deployment orchestrator.
///
/// # Safety
///
/// This must be called **after** binding the HTTP listener (which requires
/// the host network stack) but **before** serving requests.  The listener
/// socket inherited from before the `unshare` continues to work.
pub fn apply_net_namespace() -> anyhow::Result<()> {
    #[cfg(target_os = "linux")]
    {
        apply_net_namespace_linux()
    }

    #[cfg(not(target_os = "linux"))]
    {
        tracing::debug!("net-ns: not supported on this platform, skipping");
        Ok(())
    }
}

#[cfg(target_os = "linux")]
fn apply_net_namespace_linux() -> anyhow::Result<()> {
    use std::io;

    // CLONE_NEWNET = 0x40000000
    const CLONE_NEWNET: libc::c_int = 0x40_00_00_00;

    let ret = unsafe { libc::unshare(CLONE_NEWNET) };
    if ret != 0 {
        let err = io::Error::last_os_error();
        tracing::warn!(
            "net-ns: failed to create network namespace: {} (errno {}). \
             Ensure the process has CAP_NET_ADMIN.",
            err,
            err.raw_os_error().unwrap_or(0)
        );
        // Don't fail hard — the process may not have CAP_NET_ADMIN in dev mode
        return Ok(());
    }

    tracing::info!("net-ns: process moved to new network namespace (loopback only)");

    // Bring up the loopback interface inside the new namespace.
    // Without this, even localhost connections would fail.
    bring_up_lo()?;

    Ok(())
}

#[cfg(target_os = "linux")]
fn bring_up_lo() -> anyhow::Result<()> {
    use std::process::Command;

    // Use `ip link set lo up` — available on most modern Linux systems.
    let output = Command::new("ip")
        .args(["link", "set", "lo", "up"])
        .output();

    match output {
        Ok(out) if out.status.success() => {
            tracing::debug!("net-ns: loopback interface brought up");
            Ok(())
        }
        Ok(out) => {
            let stderr = String::from_utf8_lossy(&out.stderr);
            tracing::warn!("net-ns: failed to bring up loopback: {}", stderr);
            // Try ioctl-based fallback
            bring_up_lo_ioctl()
        }
        Err(e) => {
            tracing::warn!("net-ns: ip command not available: {}, trying ioctl", e);
            bring_up_lo_ioctl()
        }
    }
}

#[cfg(target_os = "linux")]
fn bring_up_lo_ioctl() -> anyhow::Result<()> {
    use std::io;

    // SIOCSIFFLAGS = 0x8914, IFF_UP = 0x1
    const SIOCSIFFLAGS: libc::c_ulong = 0x8914;
    const IFF_UP: libc::c_short = 0x1;

    // Create a socket to issue the ioctl
    let fd = unsafe { libc::socket(libc::AF_INET, libc::SOCK_DGRAM, 0) };
    if fd < 0 {
        return Err(anyhow::anyhow!("net-ns: failed to create ioctl socket"));
    }

    // struct ifreq { char ifr_name[IFNAMSIZ]; union { ... } ifr_ifru; }
    // We need to set IFF_UP on the "lo" interface.
    let mut ifreq: [u8; 40] = [0; 40]; // IFNAMSIZ = 16, rest is padding

    // Copy "lo" into ifr_name
    ifreq[0] = b'l';
    ifreq[1] = b'o';

    // Set ifr_flags = IFF_UP
    // The flags field is at offset 16 (after ifr_name) as a short
    let flags_bytes = IFF_UP.to_ne_bytes();
    ifreq[16..16 + flags_bytes.len()].copy_from_slice(&flags_bytes);

    let ret = unsafe { libc::ioctl(fd, SIOCSIFFLAGS, ifreq.as_ptr()) };
    unsafe { libc::close(fd) };

    if ret != 0 {
        let err = io::Error::last_os_error();
        tracing::warn!("net-ns: ioctl(SIOCSIFFLAGS) failed: {}", err);
    }

    Ok(())
}
