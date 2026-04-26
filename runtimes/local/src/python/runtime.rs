//! Python runtime wrapper using RustPython VM.
//!
//! This module provides the runtime integration for executing Python code
//! using the RustPython virtual machine, which offers full Python 3.11+
//! compatibility and better performance than Micropython.

use anyhow::Context;
use rustpython_vm as vm;
use tokio::sync::oneshot;

/// Message sent to the Python runtime worker
struct PythonExecutionRequest {
    code: String,
    input: String,
    response_tx: oneshot::Sender<anyhow::Result<String>>,
}

fn execute_on_interpreter(
    interpreter: &vm::Interpreter,
    python_code: &str,
    input: &str,
    max_output_bytes: usize,
) -> anyhow::Result<String> {
    interpreter.enter(|vm| -> anyhow::Result<String> {
        let scope = vm.new_scope_with_builtins();

        let input_value = vm.ctx.new_str(input.to_string());
        scope.globals.set_item("input_data", input_value.into(), vm)
            .map_err(|e| anyhow::anyhow!("Failed to set input variable: {:?}", e))?;

        let wrapper_code = format!(r#"
import json

try:
    _parsed_input = json.loads(input_data)
except (json.JSONDecodeError, TypeError):
    _parsed_input = input_data

input_data = _parsed_input

def _truncate_output(s, limit):
    if len(s) > limit:
        return s[:limit] + "...[truncated]"
    return s

{}
"#, python_code);

        let code_obj = vm
            .compile(&wrapper_code, vm::compiler::Mode::Exec, r#"<string>.py"#.to_owned())
            .map_err(|err| vm.new_syntax_error(&err, Some(&wrapper_code)))
            .map_err(|e| anyhow::anyhow!("Failed to compile Python code: {:?}", e))?;

        let result = vm.run_code_obj(code_obj, scope)
            .map_err(|e| anyhow::anyhow!("Failed to execute Python code: {:?}", e))?;

        let result_str = result.str(vm)
            .map_err(|e| anyhow::anyhow!("Failed to convert result to string: {:?}", e))?;
        let result_truncated = if result_str.to_string().len() > max_output_bytes {
            format!("{}... [output truncated {}->{} bytes]",
                &result_str.to_string()[..max_output_bytes],
                result_str.to_string().len(),
                max_output_bytes)
        } else {
            result_str.to_string()
        };
        Ok(result_truncated)
    })
}

/// Python runtime wrapper using RustPython VM.
///
/// Interpreters are created fresh for each execution to ensure thread safety.
/// Async execution uses a dedicated worker thread via channels for true interpreter
/// reuse when needed, while sync execution creates interpreters on-demand.
#[derive(Clone)]
pub struct PythonRuntime {
    /// Runtime configuration (reserved for future use)
    #[allow(dead_code)]
    config: PythonConfig,
    /// Channel sender for async execution via dedicated runtime
    execution_tx: Option<tokio::sync::mpsc::Sender<PythonExecutionRequest>>,
}

impl PythonRuntime {
    pub fn new(config: PythonConfig) -> anyhow::Result<Self> {
        let (execution_tx, mut execution_rx) = tokio::sync::mpsc::channel::<PythonExecutionRequest>(100);

        std::thread::Builder::new()
            .name("python-runtime-worker".to_string())
            .spawn(move || {
                let interpreter = vm::Interpreter::without_stdlib(Default::default());
                let rt = tokio::runtime::Builder::new_current_thread()
                    .enable_all()
                    .build()
                    .expect("Failed to create Python runtime");

                rt.block_on(async move {
                    while let Some(req) = execution_rx.recv().await {
                        let PythonExecutionRequest { code, input, response_tx } = req;
                        let result = execute_on_interpreter(&interpreter, &code, &input, 64 * 1024);
                        let _ = response_tx.send(result);
                    }
                });
            })
            .context("Failed to spawn Python runtime worker thread")?;

        Ok(Self {
            config,
            execution_tx: Some(execution_tx),
        })
    }

    /// Check if the given bytes contain Python source code.
    ///
    /// Detection strategy (in order):
    /// 1. Reject WASM binaries immediately via magic bytes (`\0asm`).
    /// 2. Require the content to be valid UTF-8 text.
    /// 3. Count how many Python-specific patterns are present and require at
    ///    least 2 matches to reduce false positives from non-Python text that
    ///    happens to contain a single keyword.
    ///
    /// Recognised patterns:
    /// - `def ` / `async def ` — function definitions
    /// - `import ` / `from ` — module imports
    /// - `class ` — class definitions
    /// - `print(` — built-in print call
    /// - `return ` — return statement
    /// - `#!` shebang (e.g. `#!/usr/bin/env python`)
    /// - `handler(` — FunctionFly handler convention
    pub fn is_python_code(code_bytes: &[u8]) -> bool {
        // 1. Reject WASM binaries: magic bytes 0x00 0x61 0x73 0x6D ("\0asm")
        if code_bytes.len() >= 4 && code_bytes[0..4] == [0x00, 0x61, 0x73, 0x6D] {
            return false;
        }

        // 2. Must be valid UTF-8 text
        let code = match std::str::from_utf8(code_bytes) {
            Ok(s) => s,
            Err(_) => return false, // Binary data that isn't WASM — not Python
        };

        // 3. Count Python-specific pattern matches; require at least 2.
        let patterns: &[&str] = &[
            "def ",
            "async def ",
            "import ",
            "from ",
            "class ",
            "print(",
            "return ",
            "#!",
            "handler(",
        ];

        let match_count = patterns.iter().filter(|&&p| code.contains(p)).count();
        match_count >= 2
    }

    pub fn execute_sync(
        &self,
        python_code: &str,
        input: &str,
    ) -> anyhow::Result<String> {
        let interpreter = vm::Interpreter::without_stdlib(Default::default());
        let max_output = self.config.max_output_bytes;
        interpreter.enter(|vm| -> anyhow::Result<String> {
            let scope = vm.new_scope_with_builtins();

            let input_value = vm.ctx.new_str(input.to_string());
            scope.globals.set_item("input_data", input_value.into(), vm)
                .map_err(|e| anyhow::anyhow!("Failed to set input variable: {:?}", e))?;

            let wrapper_code = format!(r#"
import json

try:
    _parsed_input = json.loads(input_data)
except (json.JSONDecodeError, TypeError):
    _parsed_input = input_data

input_data = _parsed_input

def _truncate_output(s, limit):
    if len(s) > limit:
        return s[:limit] + "...[truncated]"
    return s

{}
"#, python_code);

            let code_obj = vm
                .compile(&wrapper_code, vm::compiler::Mode::Exec, r#"<string>.py"#.to_owned())
                .map_err(|err| vm.new_syntax_error(&err, Some(&wrapper_code)))
                .map_err(|e| anyhow::anyhow!("Failed to compile Python code: {:?}", e))?;

            let result = vm.run_code_obj(code_obj, scope)
                .map_err(|e| anyhow::anyhow!("Failed to execute Python code: {:?}", e))?;

            let result_str = result.str(vm)
                .map_err(|e| anyhow::anyhow!("Failed to convert result to string: {:?}", e))?;
            let result_truncated = if result_str.to_string().len() > max_output {
                format!("{}... [output truncated {}->{} bytes]",
                    &result_str.to_string()[..max_output],
                    result_str.to_string().len(),
                    max_output)
            } else {
                result_str.to_string()
            };
            Ok(result_truncated)
        })
    }

    /// Execute Python code using the RustPython VM (async version with true interpreter reuse).
    ///
    /// Uses the dedicated single-threaded runtime via a channel to achieve true interpreter
    /// reuse across async calls. The interpreter lives for the lifetime of the worker thread,
    /// avoiding the overhead of creating a fresh interpreter per call.
    pub async fn execute(
        &self,
        python_code: &str,
        input: &str,
    ) -> anyhow::Result<String> {
        if let Some(ref tx) = self.execution_tx {
            let (response_tx, response_rx) = oneshot::channel();

            tx.send(PythonExecutionRequest {
                code: python_code.to_string(),
                input: input.to_string(),
                response_tx,
            })
            .await
            .context("Failed to send execution request to Python runtime")?;

            response_rx
                .await
                .context("Python runtime worker failed")?
        } else {
            tracing::warn!("Python runtime channel not available, falling back to fresh interpreter");
            let code = python_code.to_string();
            let input_data = input.to_string();
            let max_output = self.config.max_output_bytes;

            let handle = tokio::task::spawn_blocking(move || -> anyhow::Result<String> {
                let interpreter = vm::Interpreter::without_stdlib(Default::default());
                interpreter.enter(|vm| -> anyhow::Result<String> {
                    let scope = vm.new_scope_with_builtins();
                    let input_value = vm.ctx.new_str(input_data);
                    scope.globals.set_item("input_data", input_value.into(), vm)
                        .map_err(|e| anyhow::anyhow!("Failed to set input variable: {:?}", e))?;

                    let wrapper_code = format!(r#"
import json

try:
    _parsed_input = json.loads(input_data)
except (json.JSONDecodeError, TypeError):
    _parsed_input = input_data

input_data = _parsed_input

def _truncate_output(s, limit):
    if len(s) > limit:
        return s[:limit] + "...[truncated]"
    return s

{}
"#, code);

                    let code_obj = vm
                        .compile(&wrapper_code, vm::compiler::Mode::Exec, r#"<string>.py"#.to_owned())
                        .map_err(|err| vm.new_syntax_error(&err, Some(&wrapper_code)))
                        .map_err(|e| anyhow::anyhow!("Failed to compile Python code: {:?}", e))?;

                    let result = vm.run_code_obj(code_obj, scope)
                        .map_err(|e| anyhow::anyhow!("Failed to execute Python code: {:?}", e))?;

                    let result_str = result.str(vm)
                        .map_err(|e| anyhow::anyhow!("Failed to convert result to string: {:?}", e))?;
                    let result_truncated = if result_str.to_string().len() > max_output {
                        format!("{}... [output truncated {}->{} bytes]",
                            &result_str.to_string()[..max_output],
                            result_str.to_string().len(),
                            max_output)
                    } else {
                        result_str.to_string()
                    };
                    Ok(result_truncated)
                })
            });

            handle
                .await
                .context("Python runtime task panicked")?
        }
    }
}

/// Configuration for Python runtime
#[derive(Debug, Clone)]
pub struct PythonConfig {
    pub memory_limit: usize,
    pub timeout_ms: u64,
    pub debug: bool,
    pub python_version: String,
    pub runtime_version: String,
    pub max_output_bytes: usize,
    pub stack_size_limit: usize,
}

impl Default for PythonConfig {
    fn default() -> Self {
        Self {
            memory_limit: 128 * 1024 * 1024,
            timeout_ms: 5000,
            debug: false,
            python_version: "3.11".to_string(),
            runtime_version: "rustpython-0.4".to_string(),
            max_output_bytes: 64 * 1024,
            stack_size_limit: 8 * 1024 * 1024,
        }
    }
}

impl From<crate::config::Config> for PythonConfig {
    fn from(config: crate::config::Config) -> Self {
        Self {
            memory_limit: (config.memory_mb as usize) * 1024 * 1024,
            timeout_ms: config.timeout_ms,
            debug: config.python_debug,
            python_version: "3.11".to_string(),
            runtime_version: "rustpython-0.4".to_string(),
            max_output_bytes: config.max_output_bytes,
            stack_size_limit: 8 * 1024 * 1024, // 8MB default stack limit
        }
    }
}

impl PythonConfig {
    pub fn with_timeout(mut self, timeout_ms: u64) -> Self {
        self.timeout_ms = timeout_ms;
        self
    }

    pub fn with_memory_limit(mut self, memory_limit: usize) -> Self {
        self.memory_limit = memory_limit;
        self
    }

    pub fn fuel_for_timeout(&self) -> u64 {
        let timeout_ms = self.timeout_ms;
        let fuel_per_ms = 100_000;
        timeout_ms.saturating_mul(fuel_per_ms)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_python_config_defaults() {
        let config = PythonConfig::default();
        assert_eq!(config.memory_limit, 128 * 1024 * 1024);
        assert_eq!(config.timeout_ms, 5000);
        assert!(!config.debug);
        assert_eq!(config.python_version, "3.11");
        assert_eq!(config.runtime_version, "rustpython-0.4");
    }

    #[test]
    fn test_is_python_code() {
        // Valid Python code
        let python_code = r#"def hello():
    return 'world'"#;
        assert!(PythonRuntime::is_python_code(python_code.as_bytes()));

        // Valid Python with import
        let python_import = r#"import sys
print(sys.version)"#;
        assert!(PythonRuntime::is_python_code(python_import.as_bytes()));

        // Non-Python code
        let non_python = r#"console.log('hello');"#;
        assert!(!PythonRuntime::is_python_code(non_python.as_bytes()));
    }

    #[tokio::test]
    async fn test_runtime_creation() {
        let config = PythonConfig::default();
        let runtime = PythonRuntime::new(config);
        assert!(runtime.is_ok());
    }
}
