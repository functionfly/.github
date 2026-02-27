//! Error types for MicroPython WASM execution.

use std::fmt;

/// Errors that can occur during MicroPython WASM execution.
#[derive(Debug, Clone)]
pub enum MicroPythonError {
    /// Failed to load the MicroPython WASM module
    LoadError(String),

    /// Failed to link modules together
    LinkError(String),

    /// Execution failed with a specific error code
    ExecutionError(i32),

    /// Memory allocation or access error
    MemoryError(String),

    /// Execution timeout
    TimeoutError,

    /// Invalid Python code
    InvalidCode(String),

    /// Wrapper generation failed
    WrapperError(String),

    /// I/O error during execution
    IoError(String),

    /// Module instantiation failed
    InstantiationError(String),
}

impl fmt::Display for MicroPythonError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            MicroPythonError::LoadError(msg) => write!(f, "Failed to load MicroPython: {}", msg),
            MicroPythonError::LinkError(msg) => write!(f, "Module linking failed: {}", msg),
            MicroPythonError::ExecutionError(code) => {
                write!(f, "Python execution failed with code: {}", code)
            }
            MicroPythonError::MemoryError(msg) => write!(f, "Memory error: {}", msg),
            MicroPythonError::TimeoutError => write!(f, "Python execution timed out"),
            MicroPythonError::InvalidCode(msg) => write!(f, "Invalid Python code: {}", msg),
            MicroPythonError::WrapperError(msg) => write!(f, "Wrapper generation failed: {}", msg),
            MicroPythonError::IoError(msg) => write!(f, "I/O error: {}", msg),
            MicroPythonError::InstantiationError(msg) => {
                write!(f, "Module instantiation failed: {}", msg)
            }
        }
    }
}

impl std::error::Error for MicroPythonError {}

impl From<anyhow::Error> for MicroPythonError {
    fn from(err: anyhow::Error) -> Self {
        MicroPythonError::LoadError(err.to_string())
    }
}

impl From<std::io::Error> for MicroPythonError {
    fn from(err: std::io::Error) -> Self {
        MicroPythonError::IoError(err.to_string())
    }
}

/// Result type alias for MicroPython operations.
pub type Result<T> = std::result::Result<T, MicroPythonError>;

/// Error codes returned by mp_js_do_exec
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum ExecutionErrorCode {
    /// Success
    Success = 0,
    /// Syntax error in Python code
    SyntaxError = 1,
    /// Runtime error during execution
    RuntimeError = 2,
    /// Memory exhaustion
    OutOfMemory = 3,
    /// Stack overflow
    StackOverflow = 4,
    /// Unknown error
    Unknown = 255,
}

impl ExecutionErrorCode {
    /// Convert an i32 error code to the enum variant.
    pub fn from_i32(code: i32) -> Self {
        match code {
            0 => ExecutionErrorCode::Success,
            1 => ExecutionErrorCode::SyntaxError,
            2 => ExecutionErrorCode::RuntimeError,
            3 => ExecutionErrorCode::OutOfMemory,
            4 => ExecutionErrorCode::StackOverflow,
            _ => ExecutionErrorCode::Unknown,
        }
    }

    /// Get a human-readable description of the error.
    pub fn description(&self) -> &'static str {
        match self {
            ExecutionErrorCode::Success => "Execution succeeded",
            ExecutionErrorCode::SyntaxError => "Python syntax error",
            ExecutionErrorCode::RuntimeError => "Python runtime error",
            ExecutionErrorCode::OutOfMemory => "Out of memory",
            ExecutionErrorCode::StackOverflow => "Stack overflow",
            ExecutionErrorCode::Unknown => "Unknown error",
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_error_display() {
        let err = MicroPythonError::ExecutionError(1);
        assert!(err.to_string().contains("code: 1"));
    }

    #[test]
    fn test_error_code_from_i32() {
        assert_eq!(ExecutionErrorCode::from_i32(0), ExecutionErrorCode::Success);
        assert_eq!(
            ExecutionErrorCode::from_i32(1),
            ExecutionErrorCode::SyntaxError
        );
        assert_eq!(
            ExecutionErrorCode::from_i32(999),
            ExecutionErrorCode::Unknown
        );
    }

    #[test]
    fn test_error_code_description() {
        assert_eq!(
            ExecutionErrorCode::Success.description(),
            "Execution succeeded"
        );
        assert_eq!(
            ExecutionErrorCode::SyntaxError.description(),
            "Python syntax error"
        );
    }
}
