//! WASI imports for MicroPython WASM.
//!
//! Implements WASI (WebAssembly System Interface) imports for file I/O.
//! These are capability-gated and sandboxed.

use crate::micropython::memory::HostState;
use crate::micropython::errors::MicroPythonError;
use wasmtime::{Linker, Store};

/// Register WASI imports.
pub fn register(linker: &mut Linker<HostState>, _store: &mut Store<HostState>) -> Result<(), MicroPythonError> {
    register_fd_read(linker)?;
    register_fd_write(linker)?;
    register_fd_close(linker)?;
    register_fd_seek(linker)?;
    register_fd_sync(linker)?;
    register_fd_fdstat_get(linker)?;
    register_path_open(linker)?;
    register_path_create_directory(linker)?;
    register_path_remove_directory(linker)?;
    register_path_rename(linker)?;
    register_path_unlink_file(linker)?;
    register_path_filestat_get(linker)?;
    register_environ_get(linker)?;
    register_environ_sizes_get(linker)?;
    register_clock_res_get(linker)?;
    register_clock_time_get(linker)?;
    register_random_get(linker)?;
    register_poll_oneoff(linker)?;

    tracing::debug!("Registered WASI imports");
    Ok(())
}

fn register_fd_read(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "fd_read",
        |mut caller: wasmtime::Caller<'_, HostState>, fd: i32, iovs_ptr: i32, iovs_len: i32, nread_ptr: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return 1,
            };

            if fd != 0 {
                tracing::debug!("fd_read: denied for fd={}", fd);
                return 1;
            }

            let input = caller.data().input.clone();
            let input_bytes = input.as_bytes();
            let to_write = input_bytes.len().min((iovs_len as usize) * 8);

            let mem_size = memory.data_size(&caller);
            let mut offset = iovs_ptr as usize;

            for i in 0..iovs_len as usize {
                if offset + 8 > mem_size || to_write == 0 {
                    break;
                }

                let (buf_ptr, buf_len) = {
                    let mut hdr = [0u8; 8];
                    if let Err(_) = memory.read(&caller, offset, &mut hdr) {
                        break;
                    }
                    let buf_ptr = i32::from_le_bytes([hdr[0], hdr[1], hdr[2], hdr[3]]) as usize;
                    let buf_len = i32::from_le_bytes([hdr[4], hdr[5], hdr[6], hdr[7]]) as usize;
                    (buf_ptr, buf_len)
                };

                if buf_ptr + buf_len > mem_size {
                    break;
                }

                let write_len = buf_len.min(to_write);
                if let Err(_) = memory.write(&mut caller, buf_ptr, &input_bytes[..write_len]) {
                    break;
                }

                offset += 8;
            }

            if let Err(_) = memory.write(&mut caller, nread_ptr as usize, &(to_write as i32).to_le_bytes()) {
                tracing::error!("fd_read: failed to write nread");
            }

            tracing::debug!("fd_read: read {} bytes from stdin", to_write);
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register fd_read: {}", e)))?;

    Ok(())
}

fn register_fd_write(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "fd_write",
        |mut caller: wasmtime::Caller<'_, HostState>, fd: i32, iovs_ptr: i32, iovs_len: i32, nwritten_ptr: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return 1,
            };

            if fd != 1 && fd != 2 {
                tracing::debug!("fd_write: denied for fd={}", fd);
                return 1;
            }

            let mem_size = memory.data_size(&caller);
            let mut offset = iovs_ptr as usize;
            let mut total_written = 0;

            for _ in 0..iovs_len as usize {
                if offset + 8 > mem_size {
                    break;
                }

                let (buf_ptr, buf_len) = {
                    let mut hdr = [0u8; 8];
                    if let Err(_) = memory.read(&caller, offset, &mut hdr) {
                        break;
                    }
                    let buf_ptr = i32::from_le_bytes([hdr[0], hdr[1], hdr[2], hdr[3]]) as usize;
                    let buf_len = i32::from_le_bytes([hdr[4], hdr[5], hdr[6], hdr[7]]) as usize;
                    (buf_ptr, buf_len)
                };

                if buf_ptr + buf_len > mem_size {
                    break;
                }

                let mut buf = vec![0u8; buf_len];
                if let Err(_) = memory.read(&caller, buf_ptr, &mut buf) {
                    break;
                }

                total_written += buf_len;
                offset += 8;

                if let Ok(s) = String::from_utf8(buf) {
                    match fd {
                        1 => tracing::info!(target: "micropython_stdout", "{}", s),
                        2 => tracing::error!(target: "micropython_stderr", "{}", s),
                        _ => {}
                    }
                }
            }

            if let Err(_) = memory.write(&mut caller, nwritten_ptr as usize, &(total_written as i32).to_le_bytes()) {
                tracing::error!("fd_write: failed to write nwritten");
            }

            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register fd_write: {}", e)))?;

    Ok(())
}

fn register_fd_close(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "fd_close",
        |_caller: wasmtime::Caller<'_, HostState>, _fd: i32| -> i32 {
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register fd_close: {}", e)))?;

    Ok(())
}

fn register_fd_seek(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "fd_seek",
        |_caller: wasmtime::Caller<'_, HostState>, _fd: i32, _offset: i64, _whence: i32, _newoffset_ptr: i32| -> i32 {
            1 // ESPIPE
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register fd_seek: {}", e)))?;

    Ok(())
}

fn register_fd_sync(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "fd_sync",
        |_caller: wasmtime::Caller<'_, HostState>, _fd: i32| -> i32 {
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register fd_sync: {}", e)))?;

    Ok(())
}

fn register_fd_fdstat_get(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "fd_fdstat_get",
        |_caller: wasmtime::Caller<'_, HostState>, _fd: i32, _stat_ptr: i32| -> i32 {
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register fd_fdstat_get: {}", e)))?;

    Ok(())
}

fn register_path_open(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "path_open",
        |_caller: wasmtime::Caller<'_, HostState>, _dirfd: i32, _flags: i32, _path_ptr: i32, _path_len: i32, _oflags: i32, _fdflags: i32, _opened_fd_ptr: i32| -> i32 {
            2 // ENOENT
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register path_open: {}", e)))?;

    Ok(())
}

fn register_path_create_directory(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "path_create_directory",
        |_caller: wasmtime::Caller<'_, HostState>, _dirfd: i32, _path_ptr: i32, _path_len: i32| -> i32 {
            2 // ENOENT
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register path_create_directory: {}", e)))?;

    Ok(())
}

fn register_path_remove_directory(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "path_remove_directory",
        |_caller: wasmtime::Caller<'_, HostState>, _dirfd: i32, _path_ptr: i32, _path_len: i32| -> i32 {
            2 // ENOENT
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register path_remove_directory: {}", e)))?;

    Ok(())
}

fn register_path_rename(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "path_rename",
        |_caller: wasmtime::Caller<'_, HostState>, _old_dirfd: i32, _old_path_ptr: i32, _old_path_len: i32, _new_dirfd: i32, _new_path_ptr: i32, _new_path_len: i32| -> i32 {
            2 // ENOENT
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register path_rename: {}", e)))?;

    Ok(())
}

fn register_path_unlink_file(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "path_unlink_file",
        |_caller: wasmtime::Caller<'_, HostState>, _dirfd: i32, _path_ptr: i32, _path_len: i32| -> i32 {
            2 // ENOENT
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register path_unlink_file: {}", e)))?;

    Ok(())
}

fn register_path_filestat_get(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "path_filestat_get",
        |_caller: wasmtime::Caller<'_, HostState>, _dirfd: i32, _flags: i32, _path_ptr: i32, _path_len: i32, _filestat_ptr: i32| -> i32 {
            2 // ENOENT
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register path_filestat_get: {}", e)))?;

    Ok(())
}

fn register_environ_get(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "environ_get",
        |mut caller: wasmtime::Caller<'_, HostState>, environ_ptr: i32, environ_buf_ptr: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return 1,
            };

            let env_vars: Vec<(String, String)> = if let Some(ref provider) = caller.data().env_provider {
                std::env::vars()
                    .filter(|(k, _)| provider(k).is_some())
                    .collect()
            } else {
                std::env::vars().collect()
            };

            let mem_size = memory.data_size(&caller);
            let mut buf_offset = environ_buf_ptr as usize;

            for (i, (key, value)) in env_vars.iter().enumerate() {
                let entry_offset = environ_ptr as usize + (i * 8);

                let key_ptr = buf_offset as i32;
                if entry_offset + 8 > mem_size {
                    break;
                }
                if let Err(_) = memory.write(&mut caller, entry_offset, &key_ptr.to_le_bytes()) {
                    break;
                }

                let kv_string = format!("{}={}\0", key, value);
                if buf_offset + kv_string.len() > mem_size {
                    break;
                }
                if let Err(_) = memory.write(&mut caller, buf_offset, kv_string.as_bytes()) {
                    break;
                }

                buf_offset += kv_string.len();
            }

            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register environ_get: {}", e)))?;

    Ok(())
}

fn register_environ_sizes_get(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "environ_sizes_get",
        |mut caller: wasmtime::Caller<'_, HostState>, environ_count_ptr: i32, environ_buf_size_ptr: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return 1,
            };

            let env_vars: Vec<(String, String)> = if let Some(ref provider) = caller.data().env_provider {
                std::env::vars()
                    .filter(|(k, _)| provider(k).is_some())
                    .collect()
            } else {
                std::env::vars().collect()
            };

            let count = env_vars.len() as i32;
            let buf_size: i32 = env_vars.iter()
                .map(|(k, v)| k.len() + v.len() + 2)
                .sum::<usize>() as i32;

            if let Err(_) = memory.write(&mut caller, environ_count_ptr as usize, &count.to_le_bytes()) {
                return 1;
            }
            if let Err(_) = memory.write(&mut caller, environ_buf_size_ptr as usize, &buf_size.to_le_bytes()) {
                return 1;
            }

            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register environ_sizes_get: {}", e)))?;

    Ok(())
}

fn register_clock_res_get(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "clock_res_get",
        |_caller: wasmtime::Caller<'_, HostState>, _id: i32, _resolution_ptr: i32| -> i32 {
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register clock_res_get: {}", e)))?;

    Ok(())
}

fn register_clock_time_get(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "clock_time_get",
        |mut caller: wasmtime::Caller<'_, HostState>, _id: i32, _precision: i64, time_ptr: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return 1,
            };

            let now = std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .map(|d| d.as_nanos() as i64)
                .unwrap_or(0);

            if let Err(_) = memory.write(&mut caller, time_ptr as usize, &now.to_le_bytes()) {
                return 1;
            }

            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register clock_time_get: {}", e)))?;

    Ok(())
}

fn register_random_get(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "random_get",
        |mut caller: wasmtime::Caller<'_, HostState>, buf_ptr: i32, buf_len: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return 1,
            };

            let buf_ptr = buf_ptr as usize;
            let buf_len = buf_len as usize;

            let mem_size = memory.data_size(&caller);
            if buf_ptr + buf_len > mem_size {
                return 1;
            }

            let mut buf = vec![0u8; buf_len];
            use std::collections::hash_map::DefaultHasher;
            use std::hash::{Hash, Hasher};
            use std::time::Instant;

            for i in 0..buf_len {
                let mut hasher = DefaultHasher::new();
                Instant::now().hash(&mut hasher);
                std::thread::current().id().hash(&mut hasher);
                (i as u64).hash(&mut hasher);
                buf[i] = (hasher.finish() >> (i % 8)) as u8;
            }

            if let Err(_) = memory.write(&mut caller, buf_ptr, &buf) {
                return 1;
            }

            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register random_get: {}", e)))?;

    Ok(())
}

fn register_poll_oneoff(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "wasi_snapshot_preview1",
        "poll_oneoff",
        |_caller: wasmtime::Caller<'_, HostState>, _in_ptr: i32, _out_ptr: i32, _nsubscriptions: i32, _nevents_ptr: i32| -> i32 {
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register poll_oneoff: {}", e)))?;

    Ok(())
}