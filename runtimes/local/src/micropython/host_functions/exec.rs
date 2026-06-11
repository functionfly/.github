//! Core execution functions for MicroPython WASM.
//!
//! Implements `mp_js_init` and `mp_js_do_exec` which are the primary
//! entry points for Python code execution in the MicroPython runtime.
//!
//! ## Security
//!
//! - All pointers from WASM are validated before use
//! - Memory bounds are checked before read/write operations
//! - Negative pointers are rejected immediately
//! - Output truncation prevents memory exhaustion attacks

use crate::micropython::memory::HostState;
use crate::micropython::errors::{ExecutionErrorCode, MicroPythonError};
use wasmtime::{Linker, Store};
use std::sync::atomic::{AtomicBool, Ordering};
use rustpython_vm as vm;
use rustpython_vm::AsObject;

static MP_INITIALIZED: AtomicBool = AtomicBool::new(false);

const MAX_OUTPUT_BYTES: usize = 64 * 1024;

pub fn register(linker: &mut Linker<HostState>, _store: &mut Store<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "env",
        "mp_js_init",
        |_caller: wasmtime::Caller<'_, HostState>, _heap_size: i32| -> i32 {
            tracing::debug!("MicroPython mp_js_init called");
            MP_INITIALIZED.store(true, Ordering::SeqCst);
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register mp_js_init: {}", e)))?;

    linker.func_wrap(
        "env",
        "mp_js_do_exec",
        |mut caller: wasmtime::Caller<'_, HostState>, code_ptr: i32, code_len: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => {
                    tracing::error!("mp_js_do_exec: no memory export");
                    return ExecutionErrorCode::Unknown as i32;
                }
            };

            if code_ptr < 0 || code_len < 0 {
                tracing::error!("mp_js_do_exec: negative pointer code_ptr={} code_len={}", code_ptr, code_len);
                return ExecutionErrorCode::Unknown as i32;
            }

            let code_ptr_usize = code_ptr as usize;
            let code_len_usize = code_len as usize;

            let mem_size = memory.data_size(&caller);
            if code_ptr_usize + code_len_usize > mem_size {
                tracing::error!("mp_js_do_exec: out of bounds code_ptr={} code_len={} mem_size={}", code_ptr, code_len, mem_size);
                return ExecutionErrorCode::Unknown as i32;
            }

            let mut code_buf = vec![0u8; code_len_usize];
            if let Err(e) = memory.read(&caller, code_ptr_usize, &mut code_buf) {
                tracing::error!("mp_js_do_exec: failed to read memory: {}", e);
                return ExecutionErrorCode::Unknown as i32;
            }

            let python_source = match String::from_utf8(code_buf) {
                Ok(s) => s,
                Err(e) => {
                    tracing::error!("mp_js_do_exec: invalid UTF-8: {}", e);
                    return ExecutionErrorCode::SyntaxError as i32;
                }
            };

            let input = caller.data().input.clone();
            let max_output = MAX_OUTPUT_BYTES;

            let result = execute_python(&python_source, &input, max_output);

            match result {
                Ok(output) => {
                    let state = caller.data_mut();
                    if let Ok(mut state_output) = state.output.try_write() {
                        *state_output = output;
                    }
                    0
                }
                Err(exec_err) => {
                    tracing::error!("mp_js_do_exec execution error: {:?}", exec_err);
                    exec_err
                }
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register mp_js_do_exec: {}", e)))?;

    tracing::debug!("Registered mp_js_init and mp_js_do_exec with RustPython backend");
    Ok(())
}

fn execute_python(source: &str, input: &str, max_output_bytes: usize) -> Result<String, i32> {
    let stdlib_defs = rustpython_stdlib::stdlib_module_defs(&vm::Context::genesis());

    let interpreter = vm::Interpreter::builder(Default::default())
        .add_native_modules(&stdlib_defs)
        .build();

    interpreter.enter(|vm| -> Result<String, i32> {
        let scope = vm.new_scope_with_builtins();

        let input_value = vm.ctx.new_str(input.to_string());
        if let Err(e) = scope.globals.set_item("input_data", input_value.into(), vm) {
            tracing::error!("Failed to set input_data: {:?}", e);
            return Err(ExecutionErrorCode::RuntimeError as i32);
        }

        let wrapper_code = format!(r#"
def _truncate_output(s, limit):
    if len(s) > limit:
        return s[:limit] + "...[truncated]"
    return s

{}

_handler_result = handler(input_data)
if _handler_result is not None:
    _handler_result
else:
    ""
"#, source);

        let code_obj = match vm.compile(
            &wrapper_code,
            vm::compiler::Mode::Exec,
            r#"<micropython>"#.to_owned(),
        ) {
            Ok(c) => c,
            Err(err) => {
                let syntax_error = vm.new_syntax_error(&err, Some(&wrapper_code));
                let msg = syntax_error.as_object().str(vm)
                    .map(|s| s.to_string())
                    .unwrap_or_else(|_| "Syntax error".to_string());
                tracing::error!("Python syntax error: {}", msg);
                return Err(ExecutionErrorCode::SyntaxError as i32);
            }
        };

        let result = match vm.run_code_obj(code_obj, scope) {
            Ok(r) => r,
            Err(e) => {
                let exc_name = e.as_object().class().name().to_string();
                let exc_msg = e.as_object().str(vm)
                    .map(|s| s.to_string())
                    .unwrap_or_else(|_| "<no message>".to_string());
                tracing::error!("Python runtime error: {}: {}", exc_name, exc_msg);

                let error_code = match exc_name.as_str() {
                    "MemoryError" => ExecutionErrorCode::OutOfMemory,
                    "RecursionError" => ExecutionErrorCode::StackOverflow,
                    _ => ExecutionErrorCode::RuntimeError,
                };
                return Err(error_code as i32);
            }
        };

        let result_str = match result.str(vm) {
            Ok(s) => s.to_string(),
            Err(e) => {
                tracing::error!("Failed to convert result to string: {:?}", e);
                return Err(ExecutionErrorCode::RuntimeError as i32);
            }
        };

        let output = if result_str.len() > max_output_bytes {
            format!("{}... [output truncated {}->{} bytes]",
                &result_str[..max_output_bytes],
                result_str.len(),
                max_output_bytes)
        } else {
            result_str
        };

        Ok(output)
    })
}

#[allow(dead_code)]
pub fn is_initialized() -> bool {
    MP_INITIALIZED.load(Ordering::SeqCst)
}

#[allow(dead_code)]
pub fn reset_initialized() {
    MP_INITIALIZED.store(false, Ordering::SeqCst);
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_initialization_state() {
        reset_initialized();
        assert!(!is_initialized());
    }
}