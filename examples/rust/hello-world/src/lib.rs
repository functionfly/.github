//! Hello World Example for FunctionFly Rust Runtime
//!
//! This example demonstrates a simple Rust function that can be compiled
//! to WebAssembly and executed on the FunctionFly platform.
//!
//! The function reads input from stdin (passed by the runtime) and returns
//! a JSON response.

use std::io::{self, Read};

/// Main handler function - entry point for FunctionFly
///
/// Returns 0 on success, non-zero on error
#[no_mangle]
pub extern "C" fn handler() -> i32 {
    // Read input from stdin (passed by the FunctionFly runtime)
    let mut input = String::new();
    match io::stdin().read_to_string(&mut input) {
        Ok(_) => {}
        Err(e) => {
            eprintln!("Failed to read input: {}", e);
            return -1;
        }
    };

    // Parse input (expecting JSON) or use default greeting
    let name = if input.trim().is_empty() {
        "World".to_string()
    } else {
        // Try to parse as JSON and extract "name" field
        // For simplicity, we'll just use the raw input
        input.trim().to_string()
    };

    // Create JSON response
    let response = format!(
        r#"{{"message": "Hello, {}! Welcome to FunctionFly Rust runtime."}}"#,
        name
    );

    // Write response to stdout
    println!("{}", response);

    0 // Success
}

/// Alternative handler that returns a response via memory pointer
/// This follows the pointer-based pattern used in webhook-notifier example
#[no_mangle]
pub extern "C" fn handler_with_response(
    input_ptr: i32,
    input_len: i32,
    output_ptr_ptr: i32,
    output_len_ptr: i32,
) -> i32 {
    // Read input from memory using the provided pointer
    // This is a simplified version - in production you'd use the actual
    // WASI memory API to read from the guest's memory

    // For now, just return a simple response
    let response = r#"{"message": "Hello from Rust!"}"#;

    // Write response to the output pointer (simplified)
    // In actual implementation, this would copy to guest memory
    println!("{}", response);

    0 // Success
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_handler_returns_zero() {
        // Handler should return 0 on success
        // Note: This test won't actually work without stdin, but shows the pattern
        assert_eq!(handler(), 0);
    }
}
