//! Chunked I/O support for MicroPython WASM streaming.
//!
//! Provides support for streaming input and output without loading
//! everything into memory at once. This is important for handling
//! large inputs/outputs in serverless environments.

use crate::micropython::memory::HostState;
use crate::micropython::errors::MicroPythonError;
use wasmtime::{Linker, Store};
use std::sync::{Arc, Mutex};

#[derive(Debug, Clone, Default)]
pub struct StreamingState {
    input_chunks: Vec<Vec<u8>>,
    output_chunks: Vec<Vec<u8>>,
    input_chunk_ptrs: Vec<u32>,
    output_chunk_ptrs: Vec<u32>,
}

impl StreamingState {
    pub fn new() -> Self {
        Self {
            input_chunks: Vec::new(),
            output_chunks: Vec::new(),
            input_chunk_ptrs: Vec::new(),
            output_chunk_ptrs: Vec::new(),
        }
    }

    pub fn init(&mut self) {
        self.input_chunks.clear();
        self.output_chunks.clear();
        self.input_chunk_ptrs.clear();
        self.output_chunk_ptrs.clear();
    }

    pub fn add_input_chunk(&mut self, chunk: Vec<u8>) -> u32 {
        let id = self.input_chunks.len() as u32;
        self.input_chunks.push(chunk);
        id
    }

    pub fn get_input_chunk(&self, id: u32) -> Option<(&[u8], bool)> {
        self.input_chunks.get(id as usize).map(|c| (c.as_slice(), id == self.input_chunks.len() as u32 - 1))
    }

    pub fn add_output_chunk(&mut self, chunk: Vec<u8>) -> u32 {
        let id = self.output_chunks.len() as u32;
        self.output_chunks.push(chunk);
        id
    }

    pub fn get_output_chunk(&self, id: u32) -> Option<&[u8]> {
        self.output_chunks.get(id as usize).map(|c| c.as_slice())
    }

    pub fn get_output_count(&self) -> usize {
        self.output_chunks.len()
    }
}

static STREAMING_STATE: Mutex<Option<StreamingState>> = Mutex::new(None);

/// Register all streaming functions.
pub fn register(linker: &mut Linker<HostState>, _store: &mut Store<HostState>) -> Result<(), MicroPythonError> {
    // env.streaming_init() -> i32
    linker.func_wrap(
        "env",
        "streaming_init",
        |_caller: wasmtime::Caller<'_, HostState>| -> i32 {
            let mut guard = STREAMING_STATE.lock().unwrap();
            if let Some(ref mut state) = *guard {
                state.init();
            } else {
                let mut state = StreamingState::new();
                state.init();
                *guard = Some(state);
            }
            tracing::debug!("Streaming state initialized");
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register streaming_init: {}", e)))?;

    // env.streaming_send_chunk(chunk_id: i32, ptr: i32, len: i32, is_last: i32) -> i32
    linker.func_wrap(
        "env",
        "streaming_send_chunk",
        |mut caller: wasmtime::Caller<'_, HostState>, chunk_id: i32, ptr: i32, len: i32, is_last: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -1,
            };

            let ptr = ptr as usize;
            let len = len as usize;

            let mem_size = memory.data_size(&caller);
            if ptr + len > mem_size {
                tracing::error!("streaming_send_chunk: out of bounds");
                return -1;
            }

            let mut chunk = vec![0u8; len];
            if let Err(e) = memory.read(&caller, ptr, &mut chunk) {
                tracing::error!("streaming_send_chunk: failed to read memory: {}", e);
                return -1;
            }

            let mut guard = STREAMING_STATE.lock().unwrap();
            if let Some(ref mut state) = *guard {
                state.add_input_chunk(chunk);
                tracing::debug!("streaming_send_chunk: added chunk {} ({} bytes, is_last={})", chunk_id, len, is_last != 0);
            }

            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register streaming_send_chunk: {}", e)))?;

    // env.streaming_get_output_chunk(chunk_id: i32) -> ptr i32
    linker.func_wrap(
        "env",
        "streaming_get_output_chunk",
        |mut caller: wasmtime::Caller<'_, HostState>, chunk_id: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return 0,
            };

            let guard = STREAMING_STATE.lock().unwrap();
            let state = match guard.as_ref() {
                Some(s) => s,
                None => return 0,
            };

            let chunk = match state.get_output_chunk(chunk_id as u32) {
                Some(c) => c,
                None => return 0,
            };

            // Write metadata to known location (chunk metadata table at offset 4096)
            let meta_offset = 4096 + (chunk_id as usize) * 16;
            let mem_size = memory.data_size(&caller);
            if meta_offset + 16 > mem_size {
                return 0;
            }

            // Get output pointer for this chunk
            let output_base = caller.data().memory.layout().output_buffer_base as usize;
            let chunk_ptr = output_base + (chunk_id as usize) * 4096;

            // Write chunk to memory
            if chunk_ptr + chunk.len() > mem_size {
                tracing::warn!("streaming_get_output_chunk: chunk {} too large for output buffer", chunk_id);
                return 0;
            }

            if let Err(e) = memory.write(&mut caller, chunk_ptr, chunk) {
                tracing::error!("streaming_get_output_chunk: failed to write: {}", e);
                return 0;
            }

// Write metadata
            let meta_ptr = meta_offset as i32;
            tracing::debug!("streaming_get_output_chunk: chunk {} at ptr {} ({} bytes)", chunk_id, chunk_ptr, chunk.len());

            meta_ptr
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register streaming_get_output_chunk: {}", e)))?;

    // env.streaming_get_input_chunk(chunk_id: i32) -> ptr i32
    linker.func_wrap(
        "env",
        "streaming_get_input_chunk",
        |mut caller: wasmtime::Caller<'_, HostState>, chunk_id: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return 0,
            };

            let guard = STREAMING_STATE.lock().unwrap();
            let state = match guard.as_ref() {
                Some(s) => s,
                None => return 0,
            };

            let (chunk, is_last) = match state.get_input_chunk(chunk_id as u32) {
                Some(c) => c,
                None => return 0,
            };

            // Write metadata to known location (chunk metadata table at offset 8192)
            let meta_offset = 8192 + (chunk_id as usize) * 16;
            let mem_size = memory.data_size(&caller);
            if meta_offset + 16 > mem_size {
                return 0;
            }

            // Get input pointer for this chunk
            let input_base = caller.data().memory.layout().code_buffer_base as usize;
            let chunk_ptr = input_base + (chunk_id as usize) * 4096;

            // Write chunk to memory
            if chunk_ptr + chunk.len() > mem_size {
                tracing::warn!("streaming_get_input_chunk: chunk {} too large", chunk_id);
                return 0;
            }

            if let Err(e) = memory.write(&mut caller, chunk_ptr, chunk) {
                tracing::error!("streaming_get_input_chunk: failed to write: {}", e);
                return 0;
            }

            tracing::debug!("streaming_get_input_chunk: chunk {} at ptr {} ({} bytes, is_last={})", chunk_id, chunk_ptr, chunk.len(), is_last);

            meta_offset as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register streaming_get_input_chunk: {}", e)))?;

    // env.streaming_set_output_ready(chunk_id: i32, ptr: i32, chunk_len: i32) -> i32
    linker.func_wrap(
        "env",
        "streaming_set_output_ready",
        |mut caller: wasmtime::Caller<'_, HostState>, chunk_id: i32, ptr: i32, chunk_len: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -1,
            };

            let ptr = ptr as usize;
            let chunk_len = chunk_len as usize;

            let mem_size = memory.data_size(&caller);
            if ptr + chunk_len > mem_size {
                return -1;
            }

            let mut chunk = vec![0u8; chunk_len];
            if let Err(e) = memory.read(&caller, ptr, &mut chunk) {
                tracing::error!("streaming_set_output_ready: failed to read: {}", e);
                return -1;
            }

            let mut guard = STREAMING_STATE.lock().unwrap();
            if let Some(ref mut state) = *guard {
                state.add_output_chunk(chunk);
                tracing::debug!("streaming_set_output_ready: chunk {} ready ({} bytes)", chunk_id, chunk_len);
            }

            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register streaming_set_output_ready: {}", e)))?;

    // env.streaming_get_next_output_ptr() -> ptr i32
    linker.func_wrap(
        "env",
        "streaming_get_next_output_ptr",
        |caller: wasmtime::Caller<'_, HostState>| -> i32 {
            let guard = STREAMING_STATE.lock().unwrap();
            let state = match guard.as_ref() {
                Some(s) => s,
                None => return 0,
            };

            let next_id = state.get_output_count() as u32;
            let output_base = caller.data().memory.layout().output_buffer_base as u32;
            let next_ptr = output_base + next_id * 4096;

            tracing::debug!("streaming_get_next_output_ptr: chunk {} at ptr {}", next_id, next_ptr);
            next_ptr as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register streaming_get_next_output_ptr: {}", e)))?;

    // env.streaming_chunk_read(chunk_id: i32, dest_ptr: i32, max_len: i32) -> actual_len i32
    linker.func_wrap(
        "env",
        "streaming_chunk_read",
        |mut caller: wasmtime::Caller<'_, HostState>, chunk_id: i32, dest_ptr: i32, max_len: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -1,
            };

            let guard = STREAMING_STATE.lock().unwrap();
            let state = match guard.as_ref() {
                Some(s) => s,
                None => return -1,
            };

            let chunk = match state.get_input_chunk(chunk_id as u32) {
                Some((c, _)) => c,
                None => return -1,
            };

            let dest_ptr = dest_ptr as usize;
            let max_len = max_len as usize;
            let to_copy = chunk.len().min(max_len);

            if let Err(e) = memory.write(&mut caller, dest_ptr, &chunk[..to_copy]) {
                tracing::error!("streaming_chunk_read: failed to write: {}", e);
                return -1;
            }

            tracing::debug!("streaming_chunk_read: read {} bytes from chunk {} into ptr {}", to_copy, chunk_id, dest_ptr);
            to_copy as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register streaming_chunk_read: {}", e)))?;

    tracing::debug!("Registered streaming functions");
    Ok(())
}

/// Get the streaming state for external access.
#[allow(dead_code)]
pub fn get_streaming_state() -> Option<StreamingState> {
    STREAMING_STATE.lock().unwrap().clone()
}