//! Wrapper module generator for MicroPython linking.
//!
//! This module generates WebAssembly wrapper modules that provide the interface
//! between the host runtime and the MicroPython interpreter.

use super::errors::{MicroPythonError, Result};
use super::memory::MemoryLayout;

/// Configuration for wrapper generation.
#[derive(Debug, Clone)]
pub struct WrapperConfig {
    /// Memory layout configuration
    pub memory: MemoryLayout,
    /// Enable debug logging
    #[allow(dead_code)]
    pub debug: bool,
    /// Maximum output size
    #[allow(dead_code)]
    pub max_output_size: usize,
}

impl Default for WrapperConfig {
    fn default() -> Self {
        Self {
            memory: MemoryLayout::default(),
            debug: false,
            max_output_size: 64 * 1024, // 64KB
        }
    }
}

/// Generator for MicroPython wrapper modules.
pub struct WrapperGenerator {
    config: WrapperConfig,
}

impl WrapperGenerator {
    /// Create a new wrapper generator with default configuration.
    pub fn new() -> Self {
        Self {
            config: WrapperConfig::default(),
        }
    }

    /// Create a new wrapper generator with custom configuration.
    pub fn with_config(config: WrapperConfig) -> Self {
        Self { config }
    }

    /// Generate a wrapper WASM module that embeds the given Python code.
    pub fn generate(&self, python_code: &str) -> Result<Vec<u8>> {
        // Validate Python code size
        let code_bytes = python_code.as_bytes();
        if code_bytes.len() > self.config.memory.code_buffer_size as usize {
            return Err(MicroPythonError::InvalidCode(format!(
                "Python code too large: {} bytes (max: {})",
                code_bytes.len(),
                self.config.memory.code_buffer_size
            )));
        }

        // Generate WAT (WebAssembly Text) format
        let wat = self.generate_wat(python_code)?;

        // Compile WAT to WASM binary
        match wat::parse_str(&wat) {
            Ok(wasm) => Ok(wasm),
            Err(e) => Err(MicroPythonError::WrapperError(format!(
                "WAT compilation failed: {}",
                e
            ))),
        }
    }

    /// Generate WAT (WebAssembly Text) for the wrapper module.
    fn generate_wat(&self, python_code: &str) -> Result<String> {
        let layout = &self.config.memory;
        let code_base = layout.code_buffer_base;
        let output_base = layout.output_buffer_base;
        let output_size = layout.output_buffer_size;
        let dynamic_base = layout.dynamic_base;
        let initial_pages = layout.initial_pages();
        let max_pages = initial_pages * 2; // Allow growth up to 2x

        // Escape the Python code for embedding in WAT data section
        let escaped_code = escape_wat_string(python_code);
        let code_len = escaped_code.len();

        // Build WAT string using format
        let wat = format!(
            r#"(module
  ;; Import host functions for I/O
  (import "host" "log" (func $host_log (param i32 i32)))
  (import "host" "get_input" (func $host_get_input (param i32 i32) (result i32)))
  (import "host" "set_output" (func $host_set_output (param i32 i32)))

  ;; FunctionFly host function imports for Python bridge
  (import "host" "ff_log" (func $ff_log (param i32 i32 i32)))
  (import "host" "ff_get_env" (func $ff_get_env (param i32 i32 i32 i32) (result i32)))
  (import "host" "ff_kv_get" (func $ff_kv_get (param i32 i32 i32 i32) (result i32)))
  (import "host" "ff_kv_set" (func $ff_kv_set (param i32 i32 i32 i32) (result i32)))
  (import "host" "ff_state_get" (func $ff_state_get (param i32 i32 i32 i32) (result i32)))
  (import "host" "ff_state_set" (func $ff_state_set (param i32 i32 i32 i32) (result i32)))
  (import "host" "ff_state_delete" (func $ff_state_delete (param i32 i32) (result i32)))
  (import "host" "ff_state_get_fabric" (func $ff_state_get_fabric (param i32 i32 i32 i32) (result i32)))
  (import "host" "ff_state_create_snapshot" (func $ff_state_create_snapshot (param i32 i32 i32 i32 i32 i32) (result i32)))

  ;; Shared memory exported for MicroPython
  (memory (export "memory") {} {})

  ;; Data section - embed Python code at code_buffer_base
  (data (i32.const {}) "{}")

  ;; Global: allocation pointer (bump allocator)
  (global $alloc_ptr (mut i32) (i32.const {}))

  ;; Global: output written flag
  (global $output_written (mut i32) (i32.const 0))

  ;; malloc - Allocate memory from dynamic area
  (func (export "malloc") (param $size i32) (result i32)
    (local $ptr i32)
    (local $aligned_size i32)
    (local.set $aligned_size
      (i32.and
        (i32.add (local.get $size) (i32.const 7))
        (i32.const -8)))
    (global.get $alloc_ptr)
    (local.set $ptr)
    (global.set $alloc_ptr
      (i32.add (global.get $alloc_ptr) (local.get $aligned_size)))
    (local.get $ptr))

  ;; free - Free memory (no-op for bump allocator)
  (func (export "free") (param $ptr i32))

  ;; mp_js_init - Initialize MicroPython runtime
  (func $mp_js_init (export "mp_js_init") (param $heap_size i32) (result i32)
    (global.set $alloc_ptr (i32.const {}))
    (global.set $output_written (i32.const 0))
    (memory.fill (i32.const {}) (i32.const 0) (i32.const {}))
    (i32.const 0))

  ;; mp_js_do_exec - Execute Python code
  (func $mp_js_do_exec (export "mp_js_do_exec") (param $code_ptr i32) (param $code_len i32) (result i32)
    (i32.const 0))

  ;; mp_js_do_exec_async - Execute Python code asynchronously
  (func (export "mp_js_do_exec_async") (param $code_ptr i32) (param $code_len i32) (result i32)
    (call $mp_js_do_exec (local.get $code_ptr) (local.get $code_len)))

  ;; mp_js_write - Write output from MicroPython
  (func (export "mp_js_write") (param $ptr i32) (param $len i32)
    (call $host_set_output (local.get $ptr) (local.get $len))
    (global.set $output_written (i32.const 1)))

  ;; mp_js_read - Read input into buffer
  (func (export "mp_js_read") (param $ptr i32) (param $max_len i32) (result i32)
    (call $host_get_input (local.get $ptr) (local.get $max_len)))

  ;; mp_js_readline - Read a line of input
  (func (export "mp_js_readline") (param $ptr i32) (param $max_len i32) (result i32)
    (call $host_get_input (local.get $ptr) (local.get $max_len)))

  ;; mp_js_fspath - Convert path to filesystem path
  (func (export "mp_js_fspath") (param $ptr i32) (param $len i32) (result i32)
    (local.get $ptr))

  ;; get_code_base - Get the base address of embedded Python code
  (func (export "get_code_base") (result i32)
    (i32.const {}))

  ;; get_code_len - Get the length of embedded Python code
  (func (export "get_code_len") (result i32)
    (i32.const {}))

  ;; get_output_base - Get the base address of output buffer
  (func (export "get_output_base") (result i32)
    (i32.const {}))

  ;; get_output_size - Get the size of output buffer
  (func (export "get_output_size") (result i32)
    (i32.const {}))

  ;; --- FunctionFly Python Bridge ---
  ;; Host call buffer at fixed memory offset (after dynamic_base + 256KB)
  ;; Layout: [4 bytes: fn_id][4 bytes: arg1_ptr][4 bytes: arg1_len]
  ;;         [4 bytes: arg2_ptr][4 bytes: arg2_len][4 bytes: status]
  ;;         [4 bytes: result_ptr][4 bytes: result_len]

  ;; ff_invoke - Generic host function call via shared memory buffer
  ;; Reads request from HOST_CALL_BUF, dispatches to the appropriate host function
  ;; Returns 0 on success, negative on error
  (func (export "ff_invoke") (param $buf_ptr i32) (result i32)
    (local $fn_id i32)
    (local $arg1_ptr i32)
    (local $arg1_len i32)
    (local $arg2_ptr i32)
    (local $arg2_len i32)
    (local $result_buf_ptr i32)
    (local $result_len_ptr i32)
    (local $status i32)

    ;; Read fn_id from buffer
    (local.set $fn_id
      (i32.load (local.get $buf_ptr)))
    ;; Read arg1_ptr
    (local.set $arg1_ptr
      (i32.load (i32.add (local.get $buf_ptr) (i32.const 4))))
    ;; Read arg1_len
    (local.set $arg1_len
      (i32.load (i32.add (local.get $buf_ptr) (i32.const 8))))
    ;; Read arg2_ptr
    (local.set $arg2_ptr
      (i32.load (i32.add (local.get $buf_ptr) (i32.const 12))))
    ;; Read arg2_len
    (local.set $arg2_len
      (i32.load (i32.add (local.get $buf_ptr) (i32.const 16))))
    ;; Read result buffer ptr (where to write result)
    (local.set $result_buf_ptr
      (i32.load (i32.add (local.get $buf_ptr) (i32.const 24))))
    ;; Read result len ptr (where to write result length)
    (local.set $result_len_ptr
      (i32.load (i32.add (local.get $buf_ptr) (i32.const 28))))

    ;; Dispatch based on fn_id
    (block $done
      ;; fn_id=1: state_get(path_ptr, path_len, val_ptr, val_len_ptr)
      (if (i32.eq (local.get $fn_id) (i32.const 1))
        (then
          (local.set $status
            (call $ff_state_get
              (local.get $arg1_ptr) (local.get $arg1_len)
              (local.get $result_buf_ptr) (local.get $result_len_ptr)))
          (br $done)))

      ;; fn_id=2: state_set(path_ptr, path_len, val_ptr, val_len)
      (if (i32.eq (local.get $fn_id) (i32.const 2))
        (then
          (local.set $status
            (call $ff_state_set
              (local.get $arg1_ptr) (local.get $arg1_len)
              (local.get $arg2_ptr) (local.get $arg2_len)))
          (br $done)))

      ;; fn_id=3: state_delete(path_ptr, path_len)
      (if (i32.eq (local.get $fn_id) (i32.const 3))
        (then
          (local.set $status
            (call $ff_state_delete
              (local.get $arg1_ptr) (local.get $arg1_len)))
          (br $done)))

      ;; fn_id=4: state_get_fabric(fabric_id_ptr, fabric_id_len, resp_ptr, resp_len_ptr)
      (if (i32.eq (local.get $fn_id) (i32.const 4))
        (then
          (local.set $status
            (call $ff_state_get_fabric
              (local.get $arg1_ptr) (local.get $arg1_len)
              (local.get $result_buf_ptr) (local.get $result_len_ptr)))
          (br $done)))

      ;; fn_id=5: state_create_snapshot(path_ptr, path_len, label_ptr, label_len, resp_ptr, resp_len_ptr)
      (if (i32.eq (local.get $fn_id) (i32.const 5))
        (then
          (local.set $status
            (call $ff_state_create_snapshot
              (local.get $arg1_ptr) (local.get $arg1_len)
              (local.get $arg2_ptr) (local.get $arg2_len)
              (local.get $result_buf_ptr) (local.get $result_len_ptr)))
          (br $done)))

      ;; fn_id=6: get_env(name_ptr, name_len, val_ptr, val_len_ptr)
      (if (i32.eq (local.get $fn_id) (i32.const 6))
        (then
          (local.set $status
            (call $ff_get_env
              (local.get $arg1_ptr) (local.get $arg1_len)
              (local.get $result_buf_ptr) (local.get $result_len_ptr)))
          (br $done)))

      ;; fn_id=7: kv_get(key_ptr, key_len, val_ptr, val_len_ptr)
      (if (i32.eq (local.get $fn_id) (i32.const 7))
        (then
          (local.set $status
            (call $ff_kv_get
              (local.get $arg1_ptr) (local.get $arg1_len)
              (local.get $result_buf_ptr) (local.get $result_len_ptr)))
          (br $done)))

      ;; fn_id=8: kv_set(key_ptr, key_len, val_ptr, val_len)
      (if (i32.eq (local.get $fn_id) (i32.const 8))
        (then
          (local.set $status
            (call $ff_kv_set
              (local.get $arg1_ptr) (local.get $arg1_len)
              (local.get $arg2_ptr) (local.get $arg2_len)))
          (br $done)))

      ;; fn_id=9: log(level, msg_ptr, msg_len)
      (if (i32.eq (local.get $fn_id) (i32.const 9))
        (then
          (call $ff_log
            (local.get $arg1_ptr) (local.get $arg2_ptr) (local.get $arg2_len))
          (local.set $status (i32.const 0))
          (br $done)))

      ;; Unknown function ID
      (local.set $status (i32.const -99)))

    ;; Write status to buffer offset 20
    (i32.store (i32.add (local.get $buf_ptr) (i32.const 20)) (local.get $status))

    ;; Return status
    (local.get $status))
)"#,
            initial_pages,
            max_pages,
            code_base,
            escaped_code,
            dynamic_base,
            dynamic_base,
            output_base,
            output_size,
            code_base,
            code_len,
            output_base,
            output_size
        );

        Ok(wat)
    }

    /// Get a reference to the config.
    pub fn config(&self) -> &WrapperConfig {
        &self.config
    }
}

impl Default for WrapperGenerator {
    fn default() -> Self {
        Self::new()
    }
}

/// Escape a string for embedding in WAT data section.
/// Handles special characters that would break WAT parsing.
pub fn escape_wat_string(s: &str) -> String {
    let mut result = String::with_capacity(s.len());
    for c in s.chars() {
        match c {
            '\\' => result.push_str("\\\\"),
            '"' => result.push_str("\\\""),
            '\n' => result.push_str("\\n"),
            '\r' => result.push_str("\\r"),
            '\t' => result.push_str("\\t"),
            c if c.is_ascii_control() => {
                result.push_str(&format!("\\{:02x}", c as u8));
            }
            c => result.push(c),
        }
    }
    result
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_wrapper_generator_new() {
        let gen = WrapperGenerator::new();
        assert!(!gen.config().debug);
    }

    #[test]
    fn test_escape_wat_string() {
        assert_eq!(escape_wat_string("hello"), "hello");
        assert_eq!(escape_wat_string("he\"llo"), "he\\\"llo");
        assert_eq!(escape_wat_string("hello\nworld"), "hello\\nworld");
        assert_eq!(escape_wat_string("a\\b"), "a\\\\b");
    }

    #[test]
    fn test_generate_wrapper() {
        let gen = WrapperGenerator::new();
        let code = "print('hello world')";
        let wasm = gen.generate(code).unwrap();

        // Verify it's a valid WASM module (magic bytes)
        assert_eq!(&wasm[0..4], &[0x00, 0x61, 0x73, 0x6D]);

        // Verify version
        assert_eq!(&wasm[4..8], &[0x01, 0x00, 0x00, 0x00]);
    }

    #[test]
    fn test_generate_wat_valid() {
        let gen = WrapperGenerator::new();
        let code = "def handler(event): return event";
        let result = gen.generate(code);
        assert!(result.is_ok());
    }

    #[test]
    fn test_code_size_limit() {
        let mut config = WrapperConfig::default();
        config.memory.code_buffer_size = 10; // Very small limit

        let gen = WrapperGenerator::with_config(config);
        let code = "print('this is a very long string that exceeds the limit')";

        let result = gen.generate(code);
        assert!(result.is_err());
    }
}
