//! seccomp-BPF profile for the functionfly-local runtime process.
//!
//! When enabled (`--enable-seccomp`), this module applies a minimal syscall
//! allowlist after the runtime has finished initialization (HTTP server bound,
//! WASM engine ready) but before serving requests.  This limits the blast
//! radius if a Wasmtime sandbox escape vulnerability is exploited.
//!
//! # Security model
//!
//! The allowlist is designed for a network server that:
//! - Accepts TCP connections (accept4)
//! - Reads/writes sockets (read, write, close, etc.)
//! - Uses epoll for async I/O
//! - Allocates memory (mmap, munmap, mprotect, brk)
//! - Manages threads (clone, futex, exit_group)
//!
//! ## Hardening measures
//!
//! - **Architecture validation**: The BPF filter verifies `seccomp_data.arch`
//!   matches the expected audit architecture before processing syscall numbers.
//!   This prevents ABI confusion attacks where a malicious 32-bit syscall number
//!   could be misinterpreted on a 64-bit kernel.
//!
//! - **NO_NEW_PRIVS**: Before installing the filter, `prctl(PR_SET_NO_NEW_PRIVS)`
//!   is set.  This is required for unprivileged seccomp filter installation and
//!   prevents the process (and children) from gaining privileges via execve of
//!   setuid binaries.
//!
//! - **CLONE_NEWUSER restriction**: The `clone` syscall is filtered to reject
//!   the `CLONE_NEWUSER` flag, preventing user namespace creation which could
//!   be used for container escape attacks.
//!
//! - **Three operating modes**:
//!   - *Permissive*: Denied syscalls return ENOSYS (graceful degradation)
//!   - *Monitor*: Denied syscalls are logged via kernel audit and return ENOSYS
//!   - *Strict*: Denied syscalls kill the process (SIGSYS → KILL_PROCESS)
//!
//! # Platform support
//!
//! seccomp-BPF is Linux-only (kernel ≥ 3.17).  KILL_PROCESS and LOG actions
//! require kernel ≥ 4.14.  On non-Linux platforms this module is a no-op.

use anyhow::Context;

/// Syscall numbers for x86_64 Linux.
#[cfg(target_arch = "x86_64")]
mod syscall_nr {
    pub const READ: i64 = 0;
    pub const WRITE: i64 = 1;
    pub const CLOSE: i64 = 3;
    pub const FSTAT: i64 = 5;
    pub const LSEEK: i64 = 8;
    pub const MMAP: i64 = 9;
    pub const MPROTECT: i64 = 10;
    pub const MUNMAP: i64 = 11;
    pub const BRK: i64 = 12;
    pub const PREAD64: i64 = 17;
    pub const PWRITE64: i64 = 18;
    pub const READV: i64 = 19;
    pub const WRITEV: i64 = 20;
    pub const DUP3: i64 = 292;
    pub const SOCKET: i64 = 41;
    pub const CONNECT: i64 = 42;
    pub const ACCEPT4: i64 = 288;
    pub const SENDTO: i64 = 44;
    pub const RECVFROM: i64 = 45;
    pub const SENDMSG: i64 = 46;
    pub const RECVMSG: i64 = 47;
    pub const SHUTDOWN: i64 = 48;
    pub const BIND: i64 = 49;
    pub const LISTEN: i64 = 50;
    pub const GETSOCKNAME: i64 = 51;
    pub const GETPEERNAME: i64 = 52;
    pub const SETSOCKOPT: i64 = 54;
    pub const GETSOCKOPT: i64 = 55;
    pub const CLONE: i64 = 56;
    pub const EXIT: i64 = 60;
    pub const WAIT4: i64 = 61;
    pub const FCNTL: i64 = 72;
    pub const RENAME: i64 = 82;
    pub const MKDIR: i64 = 83;
    pub const UNLINK: i64 = 87;
    pub const READLINK: i64 = 89;
    pub const FCHMOD: i64 = 91;
    pub const GETUID: i64 = 102;
    pub const GETGID: i64 = 104;
    pub const GETEUID: i64 = 107;
    pub const GETEGID: i64 = 108;
    pub const GETTID: i64 = 186;
    pub const FUTEX: i64 = 202;
    pub const SCHED_GETAFFINITY: i64 = 204;
    pub const SCHED_YIELD: i64 = 24;
    pub const CLOCK_GETTIME: i64 = 228;
    pub const NANOSLEEP: i64 = 35;
    pub const EPOLL_CREATE1: i64 = 291;
    pub const EPOLL_CTL: i64 = 233;
    pub const EPOLL_PWAIT: i64 = 281;
    pub const EVENTFD2: i64 = 290;
    pub const PIPE2: i64 = 293;
    pub const SIGNALFD4: i64 = 289;
    pub const TIMERFD_CREATE: i64 = 283;
    pub const TIMERFD_SETTIME: i64 = 286;
    pub const TIMERFD_GETTIME: i64 = 287;
    pub const SIGALTSTACK: i64 = 131;
    pub const RT_SIGACTION: i64 = 13;
    pub const RT_SIGPROCMASK: i64 = 14;
    pub const RT_SIGRETURN: i64 = 15;
    pub const GETRANDOM: i64 = 318;
    pub const EXIT_GROUP: i64 = 231;
    pub const OPENAT: i64 = 257;
    pub const NEWFSTATAT: i64 = 262;
    pub const FSTATAT: i64 = 262;
    pub const GETDENTS64: i64 = 217;
    pub const MADVISE: i64 = 28;
    pub const PRCTL: i64 = 157;
    pub const ARCH_PRCTL: i64 = 158;
    pub const STATX: i64 = 332;

    #[allow(dead_code)]
    pub const IOCTL: i64 = 16;
    #[allow(dead_code)]
    pub const PIPE: i64 = 22;
    #[allow(dead_code)]
    pub const SELECT: i64 = 23;
    #[allow(dead_code)]
    pub const DUP: i64 = 32;
    #[allow(dead_code)]
    pub const DUP2: i64 = 33;
    #[allow(dead_code)]
    pub const ACCEPT: i64 = 43;
    #[allow(dead_code)]
    pub const FORK: i64 = 57;
    #[allow(dead_code)]
    pub const VFORK: i64 = 58;
    #[allow(dead_code)]
    pub const KILL: i64 = 62;
    #[allow(dead_code)]
    pub const FLOCK: i64 = 73;
    #[allow(dead_code)]
    pub const FSYNC: i64 = 74;
    #[allow(dead_code)]
    pub const FDATASYNC: i64 = 75;
    #[allow(dead_code)]
    pub const TRUNCATE: i64 = 76;
    #[allow(dead_code)]
    pub const FTRUNCATE: i64 = 77;
    #[allow(dead_code)]
    pub const GETCWD: i64 = 79;
    #[allow(dead_code)]
    pub const CHDIR: i64 = 80;
    #[allow(dead_code)]
    pub const CHMOD: i64 = 90;
    #[allow(dead_code)]
    pub const CHOWN: i64 = 92;
    #[allow(dead_code)]
    pub const FCHOWN: i64 = 93;
    #[allow(dead_code)]
    pub const RESTART_SYSCALL: i64 = 219;
    #[allow(dead_code)]
    pub const EPOLL_WAIT: i64 = 232;
    #[allow(dead_code)]
    pub const PREADV: i64 = 295;
    #[allow(dead_code)]
    pub const PWRITEV: i64 = 296;
    #[allow(dead_code)]
    pub const MEMFD_CREATE: i64 = 319;
    #[allow(dead_code)]
    pub const LSTAT: i64 = 6;
    #[allow(dead_code)]
    pub const STAT: i64 = 4;
    #[allow(dead_code)]
    pub const OPEN: i64 = 2;
    #[allow(dead_code)]
    pub const ACCESS: i64 = 21;
    #[allow(dead_code)]
    pub const CLONE3: i64 = 435;
}

/// Syscall numbers for aarch64 Linux.
#[cfg(target_arch = "aarch64")]
mod syscall_nr {
    pub const READ: i64 = 63;
    pub const WRITE: i64 = 64;
    pub const CLOSE: i64 = 57;
    pub const FSTAT: i64 = 80;
    pub const LSEEK: i64 = 62;
    pub const MMAP: i64 = 222;
    pub const MPROTECT: i64 = 226;
    pub const MUNMAP: i64 = 215;
    pub const BRK: i64 = 214;
    pub const IOCTL: i64 = 29;
    pub const PREAD64: i64 = 67;
    pub const PWRITE64: i64 = 68;
    pub const READV: i64 = 65;
    pub const WRITEV: i64 = 66;
    pub const PIPE2: i64 = 59;
    pub const DUP3: i64 = 24;
    pub const SOCKET: i64 = 198;
    pub const CONNECT: i64 = 203;
    pub const ACCEPT4: i64 = 242;
    pub const SENDTO: i64 = 206;
    pub const RECVFROM: i64 = 207;
    pub const SENDMSG: i64 = 211;
    pub const RECVMSG: i64 = 212;
    pub const SHUTDOWN: i64 = 210;
    pub const BIND: i64 = 200;
    pub const LISTEN: i64 = 201;
    pub const GETSOCKNAME: i64 = 204;
    pub const GETPEERNAME: i64 = 205;
    pub const SETSOCKOPT: i64 = 208;
    pub const GETSOCKOPT: i64 = 209;
    pub const CLONE: i64 = 220;
    pub const EXIT: i64 = 93;
    pub const EXIT_GROUP: i64 = 94;
    pub const WAIT4: i64 = 260;
    pub const KILL: i64 = 129;
    pub const FCNTL: i64 = 25;
    pub const FLOCK: i64 = 43;
    pub const FSYNC: i64 = 82;
    pub const FDATASYNC: i64 = 83;
    pub const TRUNCATE: i64 = 45;
    pub const FTRUNCATE: i64 = 46;
    pub const GETCWD: i64 = 17;
    pub const CHDIR: i64 = 49;
    pub const RENAMEAT2: i64 = 276;
    pub const MKDIRAT: i64 = 34;
    pub const UNLINKAT: i64 = 35;
    pub const READLINKAT: i64 = 78;
    pub const FCHMOD: i64 = 52;
    pub const FCHOWN: i64 = 55;
    pub const GETUID: i64 = 174;
    pub const GETGID: i64 = 176;
    pub const GETEUID: i64 = 175;
    pub const GETEGID: i64 = 177;
    pub const GETTID: i64 = 178;
    pub const FUTEX: i64 = 98;
    pub const SCHED_GETAFFINITY: i64 = 123;
    pub const SCHED_YIELD: i64 = 124;
    pub const CLOCK_GETTIME: i64 = 113;
    pub const NANOSLEEP: i64 = 101;
    pub const EPOLL_CREATE1: i64 = 20;
    pub const EPOLL_CTL: i64 = 21;
    pub const EPOLL_PWAIT: i64 = 22;
    pub const EVENTFD2: i64 = 19;
    pub const SIGNALFD4: i64 = 74;
    pub const TIMERFD_CREATE: i64 = 85;
    pub const TIMERFD_SETTIME: i64 = 86;
    pub const TIMERFD_GETTIME: i64 = 87;
    pub const SIGALTSTACK: i64 = 132;
    pub const RT_SIGACTION: i64 = 134;
    pub const RT_SIGPROCMASK: i64 = 135;
    pub const RT_SIGRETURN: i64 = 139;
    pub const GETRANDOM: i64 = 278;
    pub const MEMFD_CREATE: i64 = 279;
    pub const OPENAT: i64 = 56;
    pub const NEWFSTATAT: i64 = 79;
    pub const GETDENTS64: i64 = 61;
    pub const MADVISE: i64 = 233;
    pub const RENAMEAT: i64 = 38;
    pub const STATX: i64 = 291;
    pub const PRCTL: i64 = 167;
    pub const RENAME: i64 = 0;
    pub const MKDIR: i64 = 0;
    pub const UNLINK: i64 = 0;
    pub const READLINK: i64 = 0;
    pub const FCHMOD: i64 = 52;

    #[allow(dead_code)]
    pub const CLONE3: i64 = 435;
}

use syscall_nr::*;

// ── Seccomp operating mode ──────────────────────────────────────────────────

/// Controls what happens when a syscall not in the allowlist is invoked.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum SeccompMode {
    /// Denied syscalls return ENOSYS (errno 38).  The caller receives a
    /// "function not implemented" error and can fall back or exit gracefully.
    /// Suitable for development and testing.
    Permissive,
    /// Like Permissive, but denied syscalls are also logged via the kernel
    /// audit subsystem (`SECCOMP_RET_LOG`).  Use this to discover which
    /// syscalls the runtime actually needs before switching to Strict.
    Monitor,
    /// Denied syscalls kill the process immediately (`SECCOMP_RET_KILL_PROCESS`).
    /// The process receives SIGSYS and cannot catch or ignore it.
    /// **Required for production.**
    Strict,
}

impl SeccompMode {
    /// Return the seccomp return-action for denied syscalls in this mode.
    fn default_action(&self) -> u32 {
        match self {
            SeccompMode::Permissive => SECCOMP_RET_ERRNO | 38, // ENOSYS
            SeccompMode::Monitor => SECCOMP_RET_LOG | 38,      // LOG + ENOSYS
            SeccompMode::Strict => SECCOMP_RET_KILL_PROCESS,
        }
    }

    /// Whether the kernel audit log should record denied syscalls.
    fn should_log(&self) -> bool {
        matches!(self, SeccompMode::Monitor | SeccompMode::Strict)
    }
}

// ── Linux seccomp constants ─────────────────────────────────────────────────

#[cfg(target_os = "linux")]
const BPF_LD: u16 = 0x00;
#[cfg(target_os = "linux")]
const BPF_W: u16 = 0x00;
#[cfg(target_os = "linux")]
const BPF_ABS: u16 = 0x20;
#[cfg(target_os = "linux")]
const BPF_JMP: u16 = 0x05;
#[cfg(target_os = "linux")]
const BPF_JEQ: u16 = 0x10;
#[cfg(target_os = "linux")]
const BPF_JSET: u16 = 0x40;
#[cfg(target_os = "linux")]
const BPF_K: u16 = 0x00;
#[cfg(target_os = "linux")]
const BPF_RET: u16 = 0x06;

#[cfg(target_os = "linux")]
const SECCOMP_RET_ALLOW: u32 = 0x7fff0000;
#[cfg(target_os = "linux")]
const SECCOMP_RET_ERRNO: u32 = 0x00050000;
#[cfg(target_os = "linux")]
const SECCOMP_RET_LOG: u32 = 0x7ffc0000;
#[cfg(target_os = "linux")]
const SECCOMP_RET_KILL_PROCESS: u32 = 0x80000000;

#[cfg(target_os = "linux")]
const SECCOMP_SET_MODE_FILTER: libc::c_ulong = 1;

/// Flag for `SECCOMP_SET_MODE_FILTER`: log all filter actions via audit.
/// Requires Linux ≥ 4.14.
#[cfg(target_os = "linux")]
const SECCOMP_FILTER_FLAG_LOG: libc::c_ulong = 2;

/// `prctl` constants.
#[cfg(target_os = "linux")]
const PR_SET_NO_NEW_PRIVS: libc::c_int = 38;

/// Audit architecture values (from `linux/audit.h`).
/// x86_64 = AUDIT_ARCH_X86_64
#[cfg(target_arch = "x86_64")]
const AUDIT_ARCH: u32 = 0xC000003E;
/// aarch64 = AUDIT_ARCH_AARCH64
#[cfg(target_arch = "aarch64")]
const AUDIT_ARCH: u32 = 0xC00000B7;

/// `CLONE_NEWUSER` flag — reject in `clone` to prevent user namespace creation.
const CLONE_NEWUSER: u32 = 0x10000000;

// ── BPF structures ──────────────────────────────────────────────────────────

/// BPF instruction (matches Linux kernel's `struct sock_filter`).
#[cfg(target_os = "linux")]
#[repr(C)]
#[derive(Copy, Clone)]
struct sock_filter {
    code: u16,
    jt: u8,
    jf: u8,
    k: u32,
}

/// BPF program header (matches Linux kernel's `struct sock_fprog`).
#[cfg(target_os = "linux")]
#[repr(C)]
struct sock_fprog {
    len: u16,
    filter: *const sock_filter,
}

// ── BPF instruction helpers ─────────────────────────────────────────────────

#[cfg(target_os = "linux")]
fn stmt(code: u16, k: u32) -> sock_filter {
    sock_filter { code, jt: 0, jf: 0, k }
}

#[cfg(target_os = "linux")]
fn jmp(code: u16, k: u32, jt: u8, jf: u8) -> sock_filter {
    sock_filter { code, jt, jf, k }
}

// ── Public API ──────────────────────────────────────────────────────────────

/// Apply a seccomp-BPF profile that only allows the syscalls needed by the
/// runtime after initialization.
///
/// # Safety
///
/// This must be called **after** all initialization is complete (HTTP server
/// bound, WASM engine ready, threads spawned).  Calling it too early will
/// cause the process to crash when an unlisted syscall is attempted.
///
/// # Parameters
///
/// * `strict` - When true, uses `SeccompMode::Strict` (KILL_PROCESS default).
///   When false, uses `SeccompMode::Permissive` (ENOSYS default).
///   Also controls whether filter installation failure is fatal.
///
/// # Platform
///
/// This is a no-op on non-Linux platforms.
pub fn apply_seccomp_profile(strict: bool) -> anyhow::Result<()> {
    let mode = if strict {
        SeccompMode::Strict
    } else {
        SeccompMode::Permissive
    };
    apply_seccomp_profile_with_mode(mode, strict)
}

/// Apply a seccomp-BPF profile with an explicit operating mode.
///
/// # Parameters
///
/// * `mode` - The seccomp operating mode (Permissive, Monitor, or Strict).
/// * `fail_hard` - If true, returns an error when the filter cannot be applied.
///   If false, logs a warning and continues without seccomp.
pub fn apply_seccomp_profile_with_mode(mode: SeccompMode, fail_hard: bool) -> anyhow::Result<()> {
    #[cfg(target_os = "linux")]
    {
        apply_seccomp_linux(mode, fail_hard)
    }

    #[cfg(not(target_os = "linux"))]
    {
        let _ = (mode, fail_hard);
        tracing::debug!("seccomp: not supported on this platform, skipping");
        Ok(())
    }
}

/// Upgrade an existing seccomp filter to strict mode (KILL_PROCESS default).
///
/// This installs a **second** BPF filter on top of any existing one.  The kernel
/// evaluates all installed filters and applies the most restrictive result, so
/// this effectively upgrades the default action from ENOSYS (or LOG) to
/// KILL_PROCESS for any syscall not in the allowlist.
///
/// This is defense-in-depth: call it after `apply_seccomp_profile(false)` to
/// harden a running process, or after `apply_seccomp_profile(true)` as a
/// belt-and-suspenders measure.
///
/// # Safety
///
/// After calling this, any disallowed syscall will terminate the process via
/// SIGSYS.  Ensure the allowlist is complete before calling.
///
/// # Errors
///
/// Returns an error if the filter cannot be installed (e.g., kernel too old,
/// missing `CAP_SYS_ADMIN`, or `NO_NEW_PRIVS` not set).
pub fn enable_strict_mode() -> anyhow::Result<()> {
    #[cfg(target_os = "linux")]
    {
        tracing::info!("seccomp: upgrading to strict mode (KILL_PROCESS default)");
        apply_seccomp_linux_inner(SeccompMode::Strict, true, true)
            .context("seccomp: failed to upgrade to strict mode")?;
        tracing::info!("seccomp: strict mode active — disallowed syscalls will kill the process");
        Ok(())
    }

    #[cfg(not(target_os = "linux"))]
    {
        tracing::debug!("seccomp: strict mode not supported on this platform");
        Ok(())
    }
}

// ── Linux implementation ────────────────────────────────────────────────────

#[cfg(target_os = "linux")]
fn apply_seccomp_linux(mode: SeccompMode, fail_hard: bool) -> anyhow::Result<()> {
    apply_seccomp_linux_inner(mode, fail_hard, false)
}

#[cfg(target_os = "linux")]
fn apply_seccomp_linux_inner(
    mode: SeccompMode,
    fail_hard: bool,
    is_upgrade: bool,
) -> anyhow::Result<()> {
    use std::io;

    // ── Step 1: Set NO_NEW_PRIVS ────────────────────────────────────────────
    // Required for unprivileged seccomp filter installation.  Without this,
    // prctl(PR_SET_SECCOMP) and seccomp() return EPERM unless the process has
    // CAP_SYS_ADMIN.  NO_NEW_PRIVS also prevents the process (and children)
    // from gaining privileges via setuid binaries.
    if !is_upgrade {
        set_no_new_privs().context("seccomp: failed to set NO_NEW_PRIVS")?;
    }

    // ── Step 2: Build the BPF program ───────────────────────────────────────
    let allowed = build_allowlist();
    let bpf = build_bpf_program(&allowed, mode.default_action());

    // ── Step 3: Install the filter ──────────────────────────────────────────
    let prog = sock_fprog {
        len: bpf.len() as u16,
        filter: bpf.as_ptr(),
    };

    let flags = if mode.should_log() {
        SECCOMP_FILTER_FLAG_LOG
    } else {
        0
    };

    // SAFETY: We're calling the seccomp syscall with a valid BPF program.
    // The program is on the stack and valid for the duration of this call.
    // NO_NEW_PRIVS has been set (or we're in an upgrade path where it was
    // already set).
    let ret = unsafe {
        libc::syscall(
            libc::SYS_seccomp,
            SECCOMP_SET_MODE_FILTER,
            flags,
            &prog as *const sock_fprog,
        )
    };

    if ret != 0 {
        let err = io::Error::last_os_error();
        let errno = err.raw_os_error().unwrap_or(0);
        if fail_hard {
            return Err(anyhow::anyhow!(
                "seccomp: failed to apply BPF filter (mode {:?}): {} (errno {}). \
                 Ensure the kernel supports seccomp (≥ 3.17) and KILL_PROCESS (≥ 4.14). \
                 For KILL_PROCESS/LOG modes, kernel ≥ 4.14 is required.",
                mode,
                err,
                errno
            ));
        }
        tracing::warn!(
            "seccomp: failed to apply BPF filter (mode {:?}): {} (errno {}). \
             Continuing without seccomp protection.",
            mode,
            err,
            errno
        );
        return Ok(());
    }

    tracing::info!(
        "seccomp: BPF filter applied successfully (mode={:?}, {} allowed syscalls, log={})",
        mode,
        allowed.len(),
        mode.should_log()
    );
    Ok(())
}

/// Set `PR_SET_NO_NEW_PRIVS` on the current process.
///
/// This is required before installing seccomp filters without `CAP_SYS_ADMIN`.
/// Once set, it cannot be cleared and prevents privilege escalation via execve
/// of setuid binaries.
#[cfg(target_os = "linux")]
fn set_no_new_privs() -> anyhow::Result<()> {
    // SAFETY: prctl(PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0) is always safe.
    // It sets a process attribute that prevents privilege escalation.
    let ret = unsafe { libc::prctl(PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0) };
    if ret != 0 {
        let err = std::io::Error::last_os_error();
        return Err(anyhow::anyhow!(
            "prctl(PR_SET_NO_NEW_PRIVS) failed: {} (errno {})",
            err,
            err.raw_os_error().unwrap_or(0)
        ));
    }
    tracing::debug!("seccomp: NO_NEW_PRIVS set successfully");
    Ok(())
}

/// Return the syscall allowlist for the runtime.
#[cfg(target_os = "linux")]
fn build_allowlist() -> Vec<i64> {
    let mut allowed: Vec<i64> = vec![
        // I/O
        READ, WRITE, CLOSE, PREAD64, PWRITE64, READV, WRITEV,
        // File ops
        FSTAT, LSEEK, OPENAT, NEWFSTATAT, GETDENTS64,
        // Memory
        MMAP, MPROTECT, MUNMAP, BRK, MADVISE,
        // Networking
        SOCKET, CONNECT, ACCEPT4, SENDTO, RECVFROM, SENDMSG, RECVMSG,
        SHUTDOWN, BIND, LISTEN, GETSOCKNAME, GETPEERNAME, SETSOCKOPT, GETSOCKOPT,
        // Polling
        EPOLL_CREATE1, EPOLL_CTL, EPOLL_PWAIT, EVENTFD2, PIPE2,
        TIMERFD_CREATE, TIMERFD_SETTIME, TIMERFD_GETTIME,
        // Signal handling
        SIGALTSTACK, RT_SIGACTION, RT_SIGPROCMASK, RT_SIGRETURN, SIGNALFD4,
        // Process / thread
        CLONE, EXIT, EXIT_GROUP, WAIT4, GETTID, GETUID, GETGID, GETEUID, GETEGID,
        // Scheduling
        FUTEX, SCHED_YIELD, SCHED_GETAFFINITY,
        // Time
        CLOCK_GETTIME, NANOSLEEP,
        // Misc
        FCNTL, DUP3, GETRANDOM,
        // File operations
        RENAME, MKDIR, UNLINK, READLINK, FCHMOD,
        // Thread / memory control — required by libc/libpthread and the dynamic linker.
        PRCTL,
        STATX,
    ];

    // ARCH_PRCTL is x86_64-only; the dynamic linker uses it for segment layout.
    #[cfg(target_arch = "x86_64")]
    allowed.push(ARCH_PRCTL);

    allowed
}

/// Build a BPF program that validates architecture, checks clone flags, and
/// allows listed syscalls.
///
/// The program structure is:
///
/// ```text
/// 1. Load seccomp_data.arch (offset 4)
/// 2. If arch != AUDIT_ARCH → KILL_PROCESS  (ABI confusion prevention)
/// 3. Load seccomp_data.nr  (offset 0)
/// 4. If nr == CLONE:
///      Load args[0] (offset 16)
///      If CLONE_NEWUSER bit set → KILL_PROCESS
///      → ALLOW (clone without NEWUSER)
/// 5. For each allowed syscall:
///      If nr == SYSCALL → ALLOW
/// 6. Return default_action (ENOSYS / LOG / KILL_PROCESS)
/// ```
#[cfg(target_os = "linux")]
fn build_bpf_program(allowed: &[i64], default_action: u32) -> Vec<sock_filter> {
    let mut bpf = Vec::<sock_filter>::new();

    // ── Architecture validation ─────────────────────────────────────────────
    // Load seccomp_data.arch (offset 4 in struct seccomp_data).
    bpf.push(stmt(BPF_LD | BPF_W | BPF_ABS, 4));
    // If arch matches, skip 1 instruction (jump to syscall number check).
    bpf.push(jmp(BPF_JMP | BPF_JEQ | BPF_K, AUDIT_ARCH, 1, 0));
    // Wrong architecture — kill the process to prevent ABI confusion.
    bpf.push(stmt(BPF_RET | BPF_K, SECCOMP_RET_KILL_PROCESS));

    // ── Load syscall number ─────────────────────────────────────────────────
    // seccomp_data.nr is at offset 0.
    bpf.push(stmt(BPF_LD | BPF_W | BPF_ABS, 0));

    // ── CLONE_NEWUSER restriction ───────────────────────────────────────────
    // If the syscall is clone, check the flags argument for CLONE_NEWUSER.
    // This prevents user namespace creation which can be used for container
    // escape attacks.
    let clone_idx = allowed.iter().position(|&s| s == CLONE);
    if let Some(_clone_pos) = clone_idx {
        // Jump over clone check if syscall is not clone.
        // We need to know the offset of the clone check block (5 instructions).
        bpf.push(jmp(BPF_JMP | BPF_JEQ | BPF_K, CLONE as u32, 0, 5));
        // Syscall IS clone — load args[0] (flags), offset 16 in seccomp_data.
        bpf.push(stmt(BPF_LD | BPF_W | BPF_ABS, 16));
        // If CLONE_NEWUSER bit is set, kill.  JSET: if (A & K) != 0 → jt, else → jf.
        bpf.push(jmp(BPF_JMP | BPF_JSET | BPF_K, CLONE_NEWUSER, 0, 1));
        bpf.push(stmt(BPF_RET | BPF_K, SECCOMP_RET_KILL_PROCESS));
        // Reload syscall number for subsequent checks.
        bpf.push(stmt(BPF_LD | BPF_W | BPF_ABS, 0));
        // Allow clone without CLONE_NEWUSER.
        bpf.push(stmt(BPF_RET | BPF_K, SECCOMP_RET_ALLOW));
    }

    // ── Allowlist checks ────────────────────────────────────────────────────
    for syscall in allowed {
        if *syscall == CLONE {
            // Already handled above with argument filtering.
            continue;
        }
        bpf.push(jmp(BPF_JMP | BPF_JEQ | BPF_K, *syscall as u32, 0, 1));
        bpf.push(stmt(BPF_RET | BPF_K, SECCOMP_RET_ALLOW));
    }

    // ── Default action ──────────────────────────────────────────────────────
    bpf.push(stmt(BPF_RET | BPF_K, default_action));

    bpf
}

// ── Tests ───────────────────────────────────────────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_seccomp_mode_default_action() {
        assert_eq!(
            SeccompMode::Permissive.default_action(),
            SECCOMP_RET_ERRNO | 38
        );
        assert_eq!(
            SeccompMode::Monitor.default_action(),
            SECCOMP_RET_LOG | 38
        );
        assert_eq!(SeccompMode::Strict.default_action(), SECCOMP_RET_KILL_PROCESS);
    }

    #[test]
    fn test_seccomp_mode_should_log() {
        assert!(!SeccompMode::Permissive.should_log());
        assert!(SeccompMode::Monitor.should_log());
        assert!(SeccompMode::Strict.should_log());
    }

    #[cfg(target_os = "linux")]
    #[test]
    fn test_bpf_program_starts_with_arch_check() {
        let allowed = build_allowlist();
        let bpf = build_bpf_program(&allowed, SECCOMP_RET_ERRNO | 38);

        // First instruction: load arch
        assert_eq!(bpf[0].code, BPF_LD | BPF_W | BPF_ABS);
        assert_eq!(bpf[0].k, 4); // offset of arch in seccomp_data

        // Second instruction: compare arch
        assert_eq!(bpf[1].code, BPF_JMP | BPF_JEQ | BPF_K);
        assert_eq!(bpf[1].k, AUDIT_ARCH);
        assert_eq!(bpf[1].jt, 1);
        assert_eq!(bpf[1].jf, 0);

        // Third instruction: kill on wrong arch
        assert_eq!(bpf[2].code, BPF_RET | BPF_K);
        assert_eq!(bpf[2].k, SECCOMP_RET_KILL_PROCESS);
    }

    #[cfg(target_os = "linux")]
    #[test]
    fn test_bpf_program_ends_with_default_action() {
        let allowed = build_allowlist();
        let default = SECCOMP_RET_ERRNO | 38;
        let bpf = build_bpf_program(&allowed, default);

        let last = bpf.last().unwrap();
        assert_eq!(last.code, BPF_RET | BPF_K);
        assert_eq!(last.k, default);
    }

    #[cfg(target_os = "linux")]
    #[test]
    fn test_bpf_strict_uses_kill_process() {
        let allowed = build_allowlist();
        let bpf = build_bpf_program(&allowed, SECCOMP_RET_KILL_PROCESS);

        let last = bpf.last().unwrap();
        assert_eq!(last.k, SECCOMP_RET_KILL_PROCESS);
    }

    #[cfg(target_os = "linux")]
    #[test]
    fn test_bpf_program_contains_clone_newuser_check() {
        let allowed = build_allowlist();
        let bpf = build_bpf_program(&allowed, SECCOMP_RET_ERRNO | 38);

        // Find the CLONE_NEWUSER constant in the program.
        let has_clone_newuser = bpf.iter().any(|f| f.k == CLONE_NEWUSER);
        assert!(has_clone_newuser, "BPF program should check CLONE_NEWUSER");

        // Find the kill action after CLONE_NEWUSER check.
        let clone_check_start = bpf
            .iter()
            .position(|f| f.code == BPF_JMP | BPF_JEQ | BPF_K && f.k == CLONE as u32)
            .expect("should find clone syscall check");

        // After the clone jump, the next instructions should check args[0].
        assert_eq!(bpf[clone_check_start + 1].code, BPF_LD | BPF_W | BPF_ABS);
        assert_eq!(bpf[clone_check_start + 1].k, 16); // args[0] offset
    }

    #[cfg(target_os = "linux")]
    #[test]
    fn test_bpf_program_size() {
        let allowed = build_allowlist();
        let bpf = build_bpf_program(&allowed, SECCOMP_RET_ERRNO | 38);

        // Program structure:
        // - 3 instructions for architecture validation (LD, JEQ, KILL)
        // - 1 instruction to load syscall number (LD nr)
        // - 6 instructions for clone/CLONE_NEWUSER block
        // - 2 instructions per non-clone allowed syscall (JEQ + ALLOW)
        // - 1 instruction for the default action
        let clone_count = allowed.iter().filter(|&&s| s == CLONE).count();
        let expected = 3 + 1 + 6 + 2 * (allowed.len() - clone_count) + 1;
        assert_eq!(bpf.len(), expected);
    }

    #[cfg(target_os = "linux")]
    #[test]
    fn test_bpf_instructions_valid() {
        let allowed = build_allowlist();
        let bpf = build_bpf_program(&allowed, SECCOMP_RET_ERRNO | 38);

        for (i, instr) in bpf.iter().enumerate() {
            // All instructions should have valid code fields.
            assert!(instr.code <= 0xff, "instruction {} has invalid code: {:#x}", i, instr.code);
            // jt and jf should not exceed program length.
            assert!(
                (instr.jt as usize) < bpf.len(),
                "instruction {} jt {} exceeds program length {}",
                i, instr.jt, bpf.len()
            );
            assert!(
                (instr.jf as usize) < bpf.len(),
                "instruction {} jf {} exceeds program length {}",
                i, instr.jf, bpf.len()
            );
        }
    }

    #[test]
    fn test_allowlist_no_duplicates() {
        let allowed = build_allowlist();
        let mut sorted = allowed.clone();
        sorted.sort();
        sorted.dedup();
        assert_eq!(allowed.len(), sorted.len(), "allowlist contains duplicates");
    }

    #[test]
    fn test_allowlist_contains_essential_syscalls() {
        let allowed = build_allowlist();
        let essential = [READ, WRITE, CLOSE, MMAP, MUNMAP, BRK, FUTEX, EXIT_GROUP, ACCEPT4, EPOLL_PWAIT];
        for syscall in essential {
            assert!(
                allowed.contains(&syscall),
                "essential syscall {} missing from allowlist",
                syscall
            );
        }
    }
}
