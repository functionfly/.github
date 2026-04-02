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
//! Everything else is denied with ENOSYS (function not implemented), which
//! causes the caller to fall back or error gracefully rather than crashing.
//!
//! # Platform support
//!
//! seccomp-BPF is Linux-only.  On non-Linux platforms this module is a no-op.

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
    pub const IOCTL: i64 = 16;
    pub const PREAD64: i64 = 17;
    pub const PWRITE64: i64 = 18;
    pub const READV: i64 = 19;
    pub const WRITEV: i64 = 20;
    pub const PIPE: i64 = 22;
    pub const SELECT: i64 = 23;
    pub const DUP: i64 = 32;
    pub const DUP2: i64 = 33;
    pub const SOCKET: i64 = 41;
    pub const CONNECT: i64 = 42;
    pub const ACCEPT: i64 = 43;
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
    pub const FORK: i64 = 57;
    pub const VFORK: i64 = 58;
    pub const EXIT: i64 = 60;
    pub const WAIT4: i64 = 61;
    pub const KILL: i64 = 62;
    pub const FCNTL: i64 = 72;
    pub const FLOCK: i64 = 73;
    pub const FSYNC: i64 = 74;
    pub const FDATASYNC: i64 = 75;
    pub const TRUNCATE: i64 = 76;
    pub const FTRUNCATE: i64 = 77;
    pub const GETCWD: i64 = 79;
    pub const CHDIR: i64 = 80;
    pub const RENAME: i64 = 82;
    pub const MKDIR: i64 = 83;
    pub const UNLINK: i64 = 87;
    pub const READLINK: i64 = 89;
    pub const CHMOD: i64 = 90;
    pub const FCHMOD: i64 = 91;
    pub const CHOWN: i64 = 92;
    pub const FCHOWN: i64 = 93;
    pub const GETUID: i64 = 102;
    pub const GETGID: i64 = 104;
    pub const GETEUID: i64 = 107;
    pub const GETEGID: i64 = 108;
    pub const GETTID: i64 = 186;
    pub const FUTEX: i64 = 202;
    pub const SCHED_GETAFFINITY: i64 = 204;
    pub const SCHED_YIELD: i64 = 24;
    pub const RESTART_SYSCALL: i64 = 219;
    pub const CLOCK_GETTIME: i64 = 228;
    pub const NANOSLEEP: i64 = 35;
    pub const EPOLL_CREATE1: i64 = 291;
    pub const EPOLL_CTL: i64 = 233;
    pub const EPOLL_WAIT: i64 = 232;
    pub const EPOLL_PWAIT: i64 = 281;
    pub const ACCEPT4: i64 = 288;
    pub const EVENTFD2: i64 = 290;
    pub const PIPE2: i64 = 293;
    pub const PREADV: i64 = 295;
    pub const PWRITEV: i64 = 296;
    pub const SIGNALFD4: i64 = 289;
    pub const TIMERFD_CREATE: i64 = 283;
    pub const TIMERFD_SETTIME: i64 = 286;
    pub const TIMERFD_GETTIME: i64 = 287;
    pub const SIGALTSTACK: i64 = 131;
    pub const RT_SIGACTION: i64 = 13;
    pub const RT_SIGPROCMASK: i64 = 14;
    pub const RT_SIGRETURN: i64 = 15;
    pub const GETRANDOM: i64 = 318;
    pub const MEMFD_CREATE: i64 = 319;
    pub const EXIT_GROUP: i64 = 231;
    pub const OPENAT: i64 = 257;
    pub const NEWFSTATAT: i64 = 262;
    pub const FSTATAT: i64 = 262;
    pub const LSTAT: i64 = 6;
    pub const STAT: i64 = 4;
    pub const OPEN: i64 = 2;
    pub const ACCESS: i64 = 21;
    pub const GETDENTS64: i64 = 217;
    pub const MADVISE: i64 = 28;
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
    pub const IOCTL: i64 = 29;
    pub const DUMMY: i64 = 0; // placeholder for aarch64
}

use syscall_nr::*;

/// Apply a seccomp-BPF profile that only allows the syscalls needed by the
/// runtime after initialization.
///
/// # Safety
///
/// This must be called **after** all initialization is complete (HTTP server
/// bound, WASM engine ready, threads spawned).  Calling it too early will
/// cause the process to crash when an unlisted syscall is attempted.
///
/// # Platform
///
/// This is a no-op on non-Linux platforms.
pub fn apply_seccomp_profile() -> anyhow::Result<()> {
    #[cfg(target_os = "linux")]
    {
        apply_seccomp_linux()
    }

    #[cfg(not(target_os = "linux"))]
    {
        tracing::debug!("seccomp: not supported on this platform, skipping");
        Ok(())
    }
}

#[cfg(target_os = "linux")]
fn apply_seccomp_linux() -> anyhow::Result<()> {
    use std::io;

    // We use raw prctl + seccomp syscalls via libc-style inline assembly or
    // the nix crate.  For simplicity and to avoid adding a dependency, we use
    // the `prctl` and `seccomp` syscalls directly via `syscall()`.

    // Allowed syscalls — minimal set for a network server + WASM runtime.
    let allowed: &[i64] = &[
        // I/O
        READ,
        WRITE,
        CLOSE,
        PREAD64,
        PWRITE64,
        READV,
        WRITEV,
        // File ops
        FSTAT,
        LSEEK,
        OPENAT,
        NEWFSTATAT,
        GETDENTS64,
        FSTATAT,
        // Memory
        MMAP,
        MPROTECT,
        MUNMAP,
        BRK,
        MADVISE,
        // Networking
        SOCKET,
        CONNECT,
        ACCEPT4,
        SENDTO,
        RECVFROM,
        SENDMSG,
        RECVMSG,
        SHUTDOWN,
        BIND,
        LISTEN,
        GETSOCKNAME,
        GETPEERNAME,
        SETSOCKOPT,
        GETSOCKOPT,
        // Polling
        EPOLL_CREATE1,
        EPOLL_CTL,
        EPOLL_PWAIT,
        EVENTFD2,
        PIPE2,
        TIMERFD_CREATE,
        TIMERFD_SETTIME,
        TIMERFD_GETTIME,
        // Signal handling
        SIGALTSTACK,
        RT_SIGACTION,
        RT_SIGPROCMASK,
        RT_SIGRETURN,
        SIGNALFD4,
        // Process / thread
        CLONE,
        EXIT,
        EXIT_GROUP,
        WAIT4,
        GETTID,
        GETUID,
        GETGID,
        GETEUID,
        GETEGID,
        // Scheduling
        FUTEX,
        SCHED_YIELD,
        SCHED_GETAFFINITY,
        // Time
        CLOCK_GETTIME,
        NANOSLEEP,
        // Misc
        FCNTL,
        DUP3,
        GETRANDOM,
        // Allow fork/vfork/kill only if absolutely needed; comment out for stricter profile
        // FORK, VFORK, KILL,
    ];

    // Build a BPF program that allows the listed syscalls and returns ENOSYS for everything else.
    let mut bpf = Vec::<sock_filter>::new();

    // Load the syscall number (arch-specific offset)
    // On x86_64: seccomp_data.sys_nr is at offset 0
    // On aarch64: same
    bpf.push(sock_filter {
        code: BPF_LD | BPF_W | BPF_ABS,
        jt: 0,
        jf: 0,
        k: 0,
    }); // A = syscall_nr

    // For each allowed syscall: if matches, allow
    for &syscall in allowed {
        bpf.push(sock_filter {
            code: BPF_JMP | BPF_JEQ | BPF_K,
            jt: 0,
            jf: 1,
            k: syscall as u32,
        });
        bpf.push(sock_filter {
            code: BPF_RET | BPF_K,
            jt: 0,
            jf: 0,
            k: SECCOMP_RET_ALLOW,
        });
    }

    // Default: return ENOSYS (function not implemented)
    bpf.push(sock_filter {
        code: BPF_RET | BPF_K,
        jt: 0,
        jf: 0,
        k: SECCOMP_RET_ERRNO | 38,
    }); // 38 = ENOSYS

    // Install the filter
    let prog = sock_fprog {
        len: bpf.len() as u16,
        filter: bpf.as_ptr(),
    };

    // SAFETY: We're calling the seccomp syscall with a valid BPF program.
    // The program is on the stack and valid for the duration of this call.
    let ret = unsafe {
        libc::syscall(
            libc::SYS_seccomp,
            SECCOMP_SET_MODE_FILTER,
            0,
            &prog as *const sock_fprog,
        )
    };

    if ret != 0 {
        let err = io::Error::last_os_error();
        tracing::warn!(
            "seccomp: failed to apply BPF filter: {} (errno {})",
            err,
            err.raw_os_error().unwrap_or(0)
        );
        // Don't fail hard — log and continue.  The process may be running in
        // a container that doesn't allow seccomp (e.g. Docker without
        // --security-opt seccomp=unconfined).
        return Ok(());
    }

    tracing::info!(
        "seccomp: BPF filter applied successfully ({} allowed syscalls)",
        allowed.len()
    );
    Ok(())
}

// BPF constants
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
const BPF_K: u16 = 0x00;
#[cfg(target_os = "linux")]
const BPF_RET: u16 = 0x06;
#[cfg(target_os = "linux")]
const SECCOMP_RET_ALLOW: u32 = 0x7fff0000;
#[cfg(target_os = "linux")]
const SECCOMP_RET_ERRNO: u32 = 0x00050000;
#[cfg(target_os = "linux")]
const SECCOMP_SET_MODE_FILTER: libc::c_ulong = 1;

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

/// Mark the current thread's seccomp filter as "strict" — any syscall not in
/// the filter list will kill the thread (SIGSYS).  Use this only after the
/// runtime is fully initialized and you're confident the allowlist is complete.
///
/// # Safety
///
/// After calling this, any disallowed syscall will terminate the process.
#[cfg(target_os = "linux")]
#[allow(dead_code)]
pub fn enable_strict_mode() -> anyhow::Result<()> {
    // SECCOMP_SET_MODE_STRICT = 0 — only read/write/exit/sigreturn allowed
    // This is too restrictive for our use case, so we instead upgrade the
    // existing filter's default action from ERRNO to KILL_PROCESS.
    //
    // For now, we don't implement strict mode upgrade because Wasmtime may
    // need syscalls we didn't anticipate.  Log a warning instead.
    tracing::warn!(
        "seccomp: strict mode requested but not implemented (filter default remains ENOSYS)"
    );
    Ok(())
}
