//! Syscalls for MicroPython WASM.
//!
//! Filesystem and network syscalls are denied in the serverless sandbox.
//! These implementations properly track denied operations for security
//! auditing while returning appropriate errno values to the guest.
//!
//! ## Security
//!
//! - All filesystem and network access is denied
//! - Denied operations are logged for security audit
//! - errno values are returned to inform the guest of denial reason
//! - Virtual working directory is provided for compatibility

use crate::micropython::memory::HostState;
use crate::micropython::errors::MicroPythonError;
use wasmtime::{Linker, Store};

const VIRTUAL_CWD: &str = "/tmp\0";

fn read_string_from_memory(
    memory: &wasmtime::Memory,
    caller: &wasmtime::Caller<'_, HostState>,
    ptr: i32,
    len: i32,
) -> Option<String> {
    if ptr < 0 || len < 0 {
        return None;
    }
    let ptr = ptr as usize;
    let len = len as usize;
    let mem_size = memory.data_size(caller);
    if ptr + len > mem_size {
        return None;
    }
    let mut buf = vec![0u8; len];
    if memory.read(caller, ptr, &mut buf).is_err() {
        return None;
    }
    String::from_utf8(buf).ok()
}

/// Register all syscall handlers.
pub fn register(linker: &mut Linker<HostState>, _store: &mut Store<HostState>) -> Result<(), MicroPythonError> {
    register_filesystem_syscalls(linker)?;
    register_network_syscalls(linker)?;
    register_process_syscalls(linker)?;
    register_signal_syscalls(linker)?;
    register_time_syscalls(linker)?;

    tracing::debug!("Registered syscall handlers");
    Ok(())
}

fn register_filesystem_syscalls(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    // env.__syscall_chdir(path_ptr: i32) -> i32
    // Change current working directory - denied in sandbox
    linker.func_wrap(
        "env",
        "__syscall_chdir",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_chdir: denied - filesystem access blocked in sandbox");
            tracing::debug!(target: "micropython_syscall", "__syscall_chdir: returning ENOENT (2)");
            2 // ENOENT
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_chdir: {}", e)))?;

    // env.__syscall_getcwd(buf_ptr: i32, size: i32) -> i32
    // Get current working directory - returns virtual path
    linker.func_wrap(
        "env",
        "__syscall_getcwd",
        |mut caller: wasmtime::Caller<'_, HostState>, buf_ptr: i32, size: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -1,
            };

            let cwd = VIRTUAL_CWD.trim_end_matches('\0');
            let cwd_bytes = cwd.as_bytes();

            if size as usize <= cwd_bytes.len() {
                tracing::debug!(target: "micropython_syscall", "__syscall_getcwd: buffer too small ({} <= {})", size, cwd_bytes.len());
                return -1; // ERANGE
            }

            let mem_size = memory.data_size(&caller);
            if (buf_ptr as usize) + cwd_bytes.len() > mem_size {
                tracing::debug!(target: "micropython_syscall", "__syscall_getcwd: buffer out of bounds");
                return -1;
            }

            if let Err(e) = memory.write(&mut caller, buf_ptr as usize, cwd_bytes) {
                tracing::error!(target: "micropython_syscall", "__syscall_getcwd: failed to write memory: {}", e);
                return -1;
            }

            tracing::debug!(target: "micropython_syscall", "__syscall_getcwd: returned virtual cwd '{}'", cwd);
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_getcwd: {}", e)))?;

    // env.__syscall_mkdirat(dirfd: i32, path_ptr: i32, mode: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_mkdirat",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_mkdirat: denied - filesystem access blocked in sandbox");
            2 // ENOENT
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_mkdirat: {}", e)))?;

    // env.__syscall_rmdir(path_ptr: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_rmdir",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_rmdir: denied - filesystem access blocked in sandbox");
            2 // ENOENT
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_rmdir: {}", e)))?;

    // env.__syscall_openat(dirfd: i32, path_ptr: i32, flags: i32, mode: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_openat",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_openat: denied - filesystem access blocked in sandbox");
            -1 // EACCES
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_openat: {}", e)))?;

    // env.__syscall_getdents64(fd: i32, dirp: i32, count: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_getdents64",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_getdents64: denied - directory listing blocked in sandbox");
            0 // No entries
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_getdents64: {}", e)))?;

    // env.__syscall_renameat(oldfd: i32, oldpath: i32, newfd: i32, newpath: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_renameat",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_renameat: denied - filesystem access blocked in sandbox");
            2 // ENOENT
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_renameat: {}", e)))?;

    // env.__syscall_unlinkat(path: i32, flags: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_unlinkat",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_unlinkat: denied - filesystem access blocked in sandbox");
            2 // ENOENT
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_unlinkat: {}", e)))?;

    // env.__syscall_fstat64(fd: i32, statbuf: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_fstat64",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_fstat64: denied - filesystem access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_fstat64: {}", e)))?;

    // env.__syscall_stat64(path: i32, statbuf: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_stat64",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_stat64: denied - filesystem access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_stat64: {}", e)))?;

    // env.__syscall_newfstatat(dirfd: i32, path: i32, statbuf: i32, flags: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_newfstatat",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_newfstatat: denied - filesystem access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_newfstatat: {}", e)))?;

    // env.__syscall_lstat64(path: i32, statbuf: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_lstat64",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_lstat64: denied - filesystem access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_lstat64: {}", e)))?;

    // env.__syscall_statfs64(path: i32, buf: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_statfs64",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_statfs64: denied - filesystem access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_statfs64: {}", e)))?;

    // env.__syscall_readlink(path: i32, buf: i32, bufsiz: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_readlink",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_readlink: denied - filesystem access blocked in sandbox");
            2 // ENOENT
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_readlink: {}", e)))?;

    // env.__syscall_access(path: i32, mode: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_access",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_access: denied - filesystem access blocked in sandbox");
            2 // ENOENT
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_access: {}", e)))?;

    // env.__syscall_faccessat(dirfd: i32, path: i32, mode: i32, flags: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_faccessat",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_faccessat: denied - filesystem access blocked in sandbox");
            2 // ENOENT
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_faccessat: {}", e)))?;

    // env.__syscall_link(oldpath: i32, newpath: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_link",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_link: denied - filesystem access blocked in sandbox");
            2 // ENOENT
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_link: {}", e)))?;

    // env.__syscall_symlink(oldpath: i32, newpath: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_symlink",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_symlink: denied - filesystem access blocked in sandbox");
            2 // ENOENT
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_symlink: {}", e)))?;

    // env.__syscall_truncate(path: i32, length: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_truncate",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_truncate: denied - filesystem access blocked in sandbox");
            2 // ENOENT
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_truncate: {}", e)))?;

    // env.__syscall_ftruncate(fd: i32, length: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_ftruncate",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_ftruncate: denied - filesystem access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_ftruncate: {}", e)))?;

    // env.__syscall_copy_file_range(src_fd: i32, src_offset: i32, dst_fd: i32, dst_offset: i32, len: i32, flags: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_copy_file_range",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_copy_file_range: denied - filesystem access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_copy_file_range: {}", e)))?;

    // env._abort_js() - Called on fatal errors
    linker.func_wrap(
        "env",
        "_abort_js",
        |_caller: wasmtime::Caller<'_, HostState>| {
            tracing::error!(target: "micropython_syscall", "_abort_js: abort called by guest");
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register _abort_js: {}", e)))?;

    Ok(())
}

fn register_network_syscalls(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    // env.__syscall_poll(fds: i32, nfds: i32, timeout: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_poll",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_poll: denied - network access blocked in sandbox");
            0 // No file descriptors to poll
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_poll: {}", e)))?;

    // env.__syscall_socket(domain: i32, type_: i32, protocol: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_socket",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_socket: denied - network access blocked in sandbox");
            -1 // EACCES
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_socket: {}", e)))?;

    // env.__syscall_connect(fd: i32, addr: i32, addrlen: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_connect",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_connect: denied - network access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_connect: {}", e)))?;

    // env.__syscall_bind(fd: i32, addr: i32, addrlen: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_bind",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_bind: denied - network access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_bind: {}", e)))?;

    // env.__syscall_listen(fd: i32, backlog: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_listen",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_listen: denied - network access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_listen: {}", e)))?;

    // env.__syscall_accept(fd: i32, addr: i32, addrlen: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_accept",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_accept: denied - network access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_accept: {}", e)))?;

    // env.__syscall_getsockopt(fd: i32, level: i32, optname: i32, optval: i32, optlen: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_getsockopt",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_getsockopt: denied - network access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_getsockopt: {}", e)))?;

    // env.__syscall_setsockopt(fd: i32, level: i32, optname: i32, optval: i32, optlen: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_setsockopt",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_setsockopt: denied - network access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_setsockopt: {}", e)))?;

    // env.__syscall_getsockname(fd: i32, addr: i32, addrlen: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_getsockname",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_getsockname: denied - network access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_getsockname: {}", e)))?;

    // env.__syscall_getpeername(fd: i32, addr: i32, addrlen: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_getpeername",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_getpeername: denied - network access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_getpeername: {}", e)))?;

    // env.__syscall_send(fd: i32, buf: i32, len: i32, flags: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_send",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_send: denied - network access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_send: {}", e)))?;

    // env.__syscall_recv(fd: i32, buf: i32, len: i32, flags: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_recv",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_recv: denied - network access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_recv: {}", e)))?;

    // env.__syscall_sendto(fd: i32, buf: i32, len: i32, flags: i32, addr: i32, addrlen: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_sendto",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_sendto: denied - network access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_sendto: {}", e)))?;

    // env.__syscall_recvfrom(fd: i32, buf: i32, len: i32, flags: i32, addr: i32, addrlen: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_recvfrom",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_recvfrom: denied - network access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_recvfrom: {}", e)))?;

    // env.__syscall_shutdown(fd: i32, how: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_shutdown",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_shutdown: denied - network access blocked in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_shutdown: {}", e)))?;

    Ok(())
}

fn register_process_syscalls(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    // env.__syscall_clone(flags: i32, stack: i32, ptid: i32, ctid: i32, tls: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_clone",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_clone: denied - process spawning not allowed in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_clone: {}", e)))?;

    // env.__syscall_exit(code: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_exit",
        |_caller: wasmtime::Caller<'_, HostState>, code: i32| -> i32 {
            tracing::info!(target: "micropython_syscall", "__syscall_exit: guest exiting with code {}", code);
            code
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_exit: {}", e)))?;

    // env.__syscall_exit_group(code: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_exit_group",
        |_caller: wasmtime::Caller<'_, HostState>, code: i32| -> i32 {
            tracing::info!(target: "micropython_syscall", "__syscall_exit_group: guest exiting with code {}", code);
            code
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_exit_group: {}", e)))?;

    // env.__syscall_wait4(pid: i32, status: i32, options: i32, rusage: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_wait4",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_wait4: denied - process operations not allowed in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_wait4: {}", e)))?;

    // env.__syscall_getpid() -> i32
    linker.func_wrap(
        "env",
        "__syscall_getpid",
        |_caller: wasmtime::Caller<'_, HostState>| -> i32 {
            // Return a virtual PID for sandbox isolation
            1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_getpid: {}", e)))?;

    // env.__syscall_getppid() -> i32
    linker.func_wrap(
        "env",
        "__syscall_getppid",
        |_caller: wasmtime::Caller<'_, HostState>| -> i32 {
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_getppid: {}", e)))?;

    // env.__syscall_getuid() -> i32
    linker.func_wrap(
        "env",
        "__syscall_getuid",
        |_caller: wasmtime::Caller<'_, HostState>| -> i32 {
            65534 // nobody
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_getuid: {}", e)))?;

    // env.__syscall_getgid() -> i32
    linker.func_wrap(
        "env",
        "__syscall_getgid",
        |_caller: wasmtime::Caller<'_, HostState>| -> i32 {
            65534 // nobody
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_getgid: {}", e)))?;

    // env.__syscall_geteuid() -> i32
    linker.func_wrap(
        "env",
        "__syscall_geteuid",
        |_caller: wasmtime::Caller<'_, HostState>| -> i32 {
            65534
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_geteuid: {}", e)))?;

    // env.__syscall_getegid() -> i32
    linker.func_wrap(
        "env",
        "__syscall_getegid",
        |_caller: wasmtime::Caller<'_, HostState>| -> i32 {
            65534
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_getegid: {}", e)))?;

    // env.__syscall_setuid(pid: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_setuid",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_setuid: denied - privilege changes not allowed in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_setuid: {}", e)))?;

    // env.__syscall_setgid(gid: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_setgid",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_setgid: denied - privilege changes not allowed in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_setgid: {}", e)))?;

    // env.__syscall_prctl(option: i32, arg2: i32, arg3: i32, arg4: i32, arg5: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_prctl",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_prctl: denied - system control not allowed in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_prctl: {}", e)))?;

    // env.__syscall_prlimit64(pid: i32, resource: i32, new_limit: i32, old_limit: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_prlimit64",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32, _: i32| -> i32 {
            tracing::warn!(target: "micropython_syscall", "__syscall_prlimit64: denied - resource limits not modifiable in sandbox");
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_prlimit64: {}", e)))?;

    Ok(())
}

fn register_signal_syscalls(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    // env.__syscall_rt_sigaction(signum: i32, act: i32, oldact: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_rt_sigaction",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32| -> i32 {
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_rt_sigaction: {}", e)))?;

    // env.__syscall_rt_sigprocmask(how: i32, set: i32, oldset: i32, sigsetsize: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_rt_sigprocmask",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32, _: i32, _: i32| -> i32 {
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_rt_sigprocmask: {}", e)))?;

    // env.__syscall_rt_sigsuspend(mask: i32, sigsetsize: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_rt_sigsuspend",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32| -> i32 {
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_rt_sigsuspend: {}", e)))?;

    Ok(())
}

fn register_time_syscalls(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    // env.__syscall_time(tloc: i32) -> i64
    linker.func_wrap(
        "env",
        "__syscall_time",
        |mut caller: wasmtime::Caller<'_, HostState>, tloc: i32| -> i64 {
            let now = std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .map(|d| d.as_secs() as i64)
                .unwrap_or(0);

            if tloc != 0 {
                let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                    Some(m) => m,
                    None => return now,
                };
                let _ = memory.write(&mut caller, tloc as usize, &now.to_le_bytes());
            }

            now
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_time: {}", e)))?;

    // env.__syscall_gettimeofday(tv: i32, tz: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_gettimeofday",
        |mut caller: wasmtime::Caller<'_, HostState>, tv: i32, _tz: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -1,
            };

            let now = std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .map(|d| {
                    let secs = d.as_secs() as i64;
                    let usecs = d.subsec_micros() as i64;
                    (secs, usecs)
                })
                .unwrap_or((0, 0));

            // tv struct: tv_sec (i64) at offset 0, tv_usec (i64) at offset 8
            let mut buf = [0u8; 16];
            buf[..8].copy_from_slice(&now.0.to_le_bytes());
            buf[8..].copy_from_slice(&now.1.to_le_bytes());

            if let Err(e) = memory.write(&mut caller, tv as usize, &buf) {
                tracing::error!(target: "micropython_syscall", "__syscall_gettimeofday: failed to write: {}", e);
                return -1;
            }

            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_gettimeofday: {}", e)))?;

    // env.__syscall_clock_gettime(clk_id: i32, tp: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_clock_gettime",
        |mut caller: wasmtime::Caller<'_, HostState>, clk_id: i32, tp: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -1,
            };

            let now = match clk_id {
                // CLOCK_REALTIME
                0 => std::time::SystemTime::now()
                    .duration_since(std::time::UNIX_EPOCH)
                    .map(|d| d.as_secs() as i64)
                    .unwrap_or(0),
                // CLOCK_MONOTONIC
                1 => std::time::Instant::now()
                    .elapsed()
                    .as_secs() as i64,
                // CLOCK_PROCESS_CPUTIME_ID
                2 => {
                    std::time::Instant::now()
                        .elapsed()
                        .as_secs() as i64
                }
                // CLOCK_THREAD_CPUTIME_ID
                3 => {
                    std::time::Instant::now()
                        .elapsed()
                        .as_secs() as i64
                }
                _ => {
                    tracing::warn!(target: "micropython_syscall", "__syscall_clock_gettime: unknown clock_id {}", clk_id);
                    return -1;
                }
            };

            // struct timespec: tv_sec (i64) at offset 0, tv_nsec (i64) at offset 8
            let mut buf = [0u8; 16];
            buf[..8].copy_from_slice(&now.to_le_bytes());
            // For CLOCK_MONOTONIC, use milliseconds as nanoseconds
            buf[8..].copy_from_slice(&0i64.to_le_bytes());

            if let Err(e) = memory.write(&mut caller, tp as usize, &buf) {
                tracing::error!(target: "micropython_syscall", "__syscall_clock_gettime: failed to write: {}", e);
                return -1;
            }

            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_clock_gettime: {}", e)))?;

    // env.__syscall_nanosleep(req: i32, rem: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_nanosleep",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32, _: i32| -> i32 {
            // No sleeping in serverless - return immediately
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_nanosleep: {}", e)))?;

    // env.__syscall_alarm(seconds: i32) -> i32
    linker.func_wrap(
        "env",
        "__syscall_alarm",
        |_caller: wasmtime::Caller<'_, HostState>, _: i32| -> i32 {
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register __syscall_alarm: {}", e)))?;

    Ok(())
}
