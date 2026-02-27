//! Memory utility functions for reading from and writing to WebAssembly memory

use wasmtime_wasi::p1::WasiP1Ctx;

/// Helper function to read a string from WASM memory
pub fn read_string_from_memory(
    caller: &mut wasmtime::Caller<WasiP1Ctx>,
    ptr: i32,
    len: i32,
) -> anyhow::Result<String> {
    if let Some(memory) = caller.get_export("memory").and_then(|e| e.into_memory()) {
        let data = memory.data(caller);
        let start = ptr as usize;
        let end = start + len as usize;

        if end > data.len() {
            return Err(anyhow::anyhow!("String read out of bounds"));
        }

        let bytes = &data[start..end];
        String::from_utf8(bytes.to_vec())
            .map_err(|e| anyhow::anyhow!("Invalid UTF-8 string: {}", e))
    } else {
        Err(anyhow::anyhow!("No memory export found"))
    }
}

/// Helper function to write a string to WASM memory
pub fn write_string_to_memory(
    caller: &mut wasmtime::Caller<WasiP1Ctx>,
    value: &str,
    value_ptr: i32,
    value_len_ptr: i32,
) -> anyhow::Result<()> {
    if let Some(memory) = caller.get_export("memory").and_then(|e| e.into_memory()) {
        let bytes = value.as_bytes();
        let len = bytes.len();

        // Write the length first
        {
            let data = memory.data_mut(&mut *caller);
            let len_slice = &mut data[value_len_ptr as usize..value_len_ptr as usize + 4];
            len_slice.copy_from_slice(&(len as u32).to_le_bytes());
        }

        // Then write the string data
        let data = memory.data_mut(&mut *caller);
        let value_slice = &mut data[value_ptr as usize..value_ptr as usize + len];
        value_slice.copy_from_slice(bytes);

        Ok(())
    } else {
        Err(anyhow::anyhow!("No memory export found"))
    }
}

/// Helper function to read bytes from WASM memory
pub fn read_bytes_from_memory(
    caller: &mut wasmtime::Caller<WasiP1Ctx>,
    ptr: i32,
    len: i32,
) -> anyhow::Result<Vec<u8>> {
    if let Some(memory) = caller.get_export("memory").and_then(|e| e.into_memory()) {
        let data = memory.data(caller);
        let start = ptr as usize;
        let end = start + len as usize;

        if end > data.len() {
            return Err(anyhow::anyhow!("Bytes read out of bounds"));
        }

        Ok(data[start..end].to_vec())
    } else {
        Err(anyhow::anyhow!("No memory export found"))
    }
}

/// Helper function to write bytes to WASM memory
pub fn write_bytes_to_memory(
    caller: &mut wasmtime::Caller<WasiP1Ctx>,
    value: &[u8],
    value_ptr: i32,
    value_len_ptr: i32,
) -> anyhow::Result<()> {
    if let Some(memory) = caller.get_export("memory").and_then(|e| e.into_memory()) {
        let len = value.len();

        // Write the length first
        {
            let data = memory.data_mut(&mut *caller);
            let len_slice = &mut data[value_len_ptr as usize..value_len_ptr as usize + 4];
            len_slice.copy_from_slice(&(len as u32).to_le_bytes());
        }

        // Then write the data
        let data = memory.data_mut(&mut *caller);
        let value_slice = &mut data[value_ptr as usize..value_ptr as usize + len];
        value_slice.copy_from_slice(value);

        Ok(())
    } else {
        Err(anyhow::anyhow!("No memory export found"))
    }
}