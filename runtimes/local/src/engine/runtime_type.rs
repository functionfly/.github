//! Runtime type detection and configuration for WASM execution.

/// Runtime type for WASM modules
#[derive(Debug, Clone, PartialEq)]
pub enum RuntimeType {
    /// Standard WebAssembly module (Rust, Go, etc.)
    Wasm,
    /// Python WASM module using RustPython
    Python,
    /// CPython compiled to WASM (full stdlib, no C extensions)
    PythonWasm,
    /// CPython in Firecracker MicroVM (Enterprise tier only)
    PythonMicroVM,
}

impl RuntimeType {
    /// Parse runtime type from string
    pub fn from_str(s: &str) -> Option<Self> {
        match s {
            "wasm" => Some(RuntimeType::Wasm),
            "python" => Some(RuntimeType::Python),
            "python-wasm" => Some(RuntimeType::PythonWasm),
            "python-microvm" => Some(RuntimeType::PythonMicroVM),
            _ => None,
        }
    }

    /// Check if this runtime type requires MicroVM execution
    pub fn requires_microvm(&self) -> bool {
        matches!(self, RuntimeType::PythonMicroVM)
    }

    /// Get the display name for this runtime
    pub fn display_name(&self) -> &'static str {
        match self {
            RuntimeType::Wasm => "WebAssembly",
            RuntimeType::Python => "RustPython",
            RuntimeType::PythonWasm => "CPython-WASM",
            RuntimeType::PythonMicroVM => "CPython (MicroVM)",
        }
    }
}
