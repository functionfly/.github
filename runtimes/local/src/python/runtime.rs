//! Python runtime wrapper using RustPython VM.
//!
//! This module provides the runtime integration for executing Python code
//! using the RustPython virtual machine, which offers full Python 3.11+
//! compatibility and better performance than Micropython.

use anyhow::Context;
use rustpython_vm as vm;
use std::sync::Arc;

/// Python runtime wrapper using RustPython VM
pub struct PythonRuntime {
    /// RustPython interpreter instance
    interpreter: Arc<vm::Interpreter>,
    /// Runtime configuration
    config: PythonConfig,
}

impl PythonRuntime {
    /// Create a new Python runtime.
    ///
    /// Note: `self.interpreter` is stored for potential future reuse, but
    /// `execute_sync` / `execute` currently create a fresh interpreter per
    /// invocation for isolation (each call gets a clean global scope with no
    /// state leakage between executions). RustPython's `Interpreter` is not
    /// `Send`, so it cannot be shared across threads without additional
    /// synchronisation. If single-threaded reuse is desired in the future,
    /// `execute_sync` can be updated to call `self.interpreter.enter(...)`.
    pub fn new(config: PythonConfig) -> anyhow::Result<Self> {
        // Create a new RustPython interpreter (stored for potential future reuse)
        let interpreter = Arc::new(vm::Interpreter::without_stdlib(Default::default()));

        Ok(Self {
            interpreter,
            config,
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
        if code_bytes.len() >= 4 && &code_bytes[0..4] == [0x00, 0x61, 0x73, 0x6D] {
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

    /// Execute Python code using the RustPython VM (synchronous version for blocking tasks)
    pub fn execute_sync(
        &self,
        python_code: &str,
        input: &str,
    ) -> anyhow::Result<String> {
        // Create a new interpreter for this execution since VM is not Send
        let interpreter = vm::Interpreter::without_stdlib(Default::default());
        interpreter.enter(|vm| -> anyhow::Result<String> {
            // Create a scope for execution
            let scope = vm.new_scope_with_builtins();

            // Set up input as a global variable in the scope (as string)
            let input_value = vm.ctx.new_str(input.to_string());
            scope.globals.set_item("input_data", input_value.into(), vm)
                .map_err(|e| anyhow::anyhow!("Failed to set input variable: {:?}", e))?;

            // Create a wrapper that parses JSON input and calls the handler
            // This wraps the user code to provide JSON parsing
            let wrapper_code = format!(r#"
import json

# Try to parse input_data as JSON, fall back to string
try:
    _parsed_input = json.loads(input_data)
except (json.JSONDecodeError, TypeError):
    _parsed_input = input_data

# Make parsed input available
input_data = _parsed_input

# User's code follows
{}
"#, python_code);

            // Execute the Python code
            let code_obj = vm
                .compile(&wrapper_code, vm::compiler::Mode::Exec, r#"<string>.py"#.to_owned())
                .map_err(|err| vm.new_syntax_error(&err, Some(&wrapper_code)))
                .map_err(|e| anyhow::anyhow!("Failed to compile Python code: {:?}", e))?;

            // Execute the code
            let result = vm.run_code_obj(code_obj, scope)
                .map_err(|e| anyhow::anyhow!("Failed to execute Python code: {:?}", e))?;

            // Get the result as a string
            let result_str = result.str(vm)
                .map_err(|e| anyhow::anyhow!("Failed to convert result to string: {:?}", e))?;
            Ok(result_str.to_string())
        })
    }

    /// Execute Python code using the RustPython VM (async version that wraps sync execution)
    pub async fn execute(
        &self,
        python_code: &str,
        input: &str,
    ) -> anyhow::Result<String> {
        // Clone the code and input for the blocking task
        let code = python_code.to_string();
        let input_data = input.to_string();

        // Execute in a blocking task since RustPython is not async
        tokio::task::spawn_blocking(move || -> anyhow::Result<String> {
            let interpreter = vm::Interpreter::without_stdlib(Default::default());
            interpreter.enter(|vm| -> anyhow::Result<String> {
                // Create a scope for execution
                let scope = vm.new_scope_with_builtins();

                // Set up input as a global variable in the scope (as string)
                let input_value = vm.ctx.new_str(input_data.clone());
                scope.globals.set_item("input_data", input_value.into(), vm)
                    .map_err(|e| anyhow::anyhow!("Failed to set input variable: {:?}", e))?;

                // Create a wrapper that parses JSON input and calls the handler
                // This wraps the user code to provide JSON parsing
                let wrapper_code = format!(r#"
import json

# Try to parse input_data as JSON, fall back to string
try:
    _parsed_input = json.loads(input_data)
except (json.JSONDecodeError, TypeError):
    _parsed_input = input_data

# Make parsed input available
input_data = _parsed_input

# User's code follows
{}
"#, code);

                // Execute the Python code
                let code_obj = vm
                    .compile(&wrapper_code, vm::compiler::Mode::Exec, r#"<string>.py"#.to_owned())
                    .map_err(|err| vm.new_syntax_error(&err, Some(&wrapper_code)))
                    .map_err(|e| anyhow::anyhow!("Failed to compile Python code: {:?}", e))?;

                // Execute the code
                let result = vm.run_code_obj(code_obj, scope)
                    .map_err(|e| anyhow::anyhow!("Failed to execute Python code: {:?}", e))?;

                // Get the result as a string
                let result_str = result.str(vm)
                    .map_err(|e| anyhow::anyhow!("Failed to convert result to string: {:?}", e))?;
                Ok(result_str.to_string())
            })
        })
        .await
        .context(r#"Failed to execute Python code in blocking task"#)?
    }
}

/// Configuration for Python runtime
#[derive(Debug, Clone)]
pub struct PythonConfig {
    /// Memory limit in bytes
    pub memory_limit: usize,
    /// Execution timeout in milliseconds
    pub timeout_ms: u64,
    /// Enable debugging
    pub debug: bool,
    /// Python version
    pub python_version: String,
    /// Runtime version
    pub runtime_version: String,
}

impl Default for PythonConfig {
    fn default() -> Self {
        Self {
            memory_limit: 128 * 1024 * 1024, // 128MB
            timeout_ms: 5000, // 5 seconds
            debug: false,
            python_version: "3.11".to_string(),
            runtime_version: "rustpython-0.4".to_string(),
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
        }
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
