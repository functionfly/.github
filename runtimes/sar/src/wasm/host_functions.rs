use std::collections::HashMap;
use std::sync::Arc;

use wasmtime::Linker;

use super::SandboxState;

pub fn register_host_functions(
    linker: &mut Linker<SandboxState>,
    allowed_env_vars: &[String],
) -> anyhow::Result<()> {
    register_ff_log(linker)?;
    register_ff_get_env(linker, allowed_env_vars)?;
    Ok(())
}

fn register_ff_log(linker: &mut Linker<SandboxState>) -> anyhow::Result<()> {
    linker.func_wrap(
        "env",
        "ff_log",
        |mut caller: wasmtime::Caller<'_, SandboxState>,
         level: i32,
         ptr: i32,
         len: i32| {
            let Some(wasmtime::Extern::Memory(memory)) = caller.get_export("memory") else {
                return;
            };
            let data = memory.data(&caller);
            let start = ptr as usize;
            let end = start + len as usize;
            if end > data.len() {
                return;
            }
            let msg = match std::str::from_utf8(&data[start..end]) {
                Ok(s) => s,
                Err(_) => "<invalid utf8>",
            };
            match level {
                0 => tracing::error!(target: "wasm::sandbox", "{}", msg),
                1 => tracing::warn!(target: "wasm::sandbox", "{}", msg),
                2 => tracing::info!(target: "wasm::sandbox", "{}", msg),
                _ => tracing::debug!(target: "wasm::sandbox", "{}", msg),
            }
        },
    )?;
    Ok(())
}

fn register_ff_get_env(
    linker: &mut Linker<SandboxState>,
    allowed_env_vars: &[String],
) -> anyhow::Result<()> {
    let env_map: Arc<HashMap<String, String>> = Arc::new(
        allowed_env_vars
            .iter()
            .filter_map(|k| std::env::var(k).ok().map(|v| (k.clone(), v)))
            .collect(),
    );

    linker.func_wrap(
        "env",
        "ff_get_env",
        move |mut caller: wasmtime::Caller<'_, SandboxState>,
              key_ptr: i32,
              key_len: i32,
              val_ptr: i32,
              val_cap: i32|
              -> i32 {
            let Some(wasmtime::Extern::Memory(memory)) = caller.get_export("memory") else {
                return -1;
            };

            let key = {
                let data = memory.data(&caller);
                let k_start = key_ptr as usize;
                let k_end = k_start + key_len as usize;
                if k_end > data.len() {
                    return -1;
                }
                match std::str::from_utf8(&data[k_start..k_end]) {
                    Ok(s) => s.to_string(),
                    Err(_) => return -1,
                }
            };

            let Some(value) = env_map.get(&key) else {
                return 0;
            };

            let v_start = val_ptr as usize;
            let copy_len = value.len().min(val_cap as usize);
            if v_start + copy_len > memory.data(&caller).len() {
                return -1;
            }

            let data_mut = memory.data_mut(&mut caller);
            data_mut[v_start..v_start + copy_len]
                .copy_from_slice(&value.as_bytes()[..copy_len]);

            copy_len as i32
        },
    )?;
    Ok(())
}
