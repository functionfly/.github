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
    /// Create a new Python runtime
    pub fn new(config: PythonConfig) -> anyhow::Result<Self> {
        // Create a new RustPython interpreter
        let interpreter = Arc::new(vm::Interpreter::without_stdlib(Default::default()));

        Ok(Self {
            interpreter,
            config,
        })
    }

    /// Check if the given bytes contain Python code
    ///
    /// This first checks if the bytes are a valid WASM binary (by checking magic bytes),
    /// and only then checks for Python keywords. This prevents false positives from
    /// FlyPy-generated WASM which contains embedded Python-like strings.
    pub fn is_python_code(code_bytes: &[u8]) -> bool {
        // Check for WASM magic bytes first: 0x00 0x61 0x73 0x6D ("\0asm")
        // If it's a valid WASM binary, it's NOT Python source code
        if code_bytes.len() >= 4 {
            let magic = &code_bytes[0..4];
            if magic == [0x00, 0x61, 0x73, 0x6D] {
                // This is a WASM binary, not Python source code
                return false;
            }
        }

        // Simple check for Python code - look for common Python keywords
        let code = String::from_utf8_lossy(code_bytes);
        code.contains("def ") ||
        code.contains("import ") ||
        code.contains("print(") ||
        code.contains("return ")
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
