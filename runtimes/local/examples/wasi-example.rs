//! Example WASI-enabled WebAssembly function for FunctionFly
//!
//! This example demonstrates basic WASI functionality:
//! - Environment variable access
//! - Command line arguments
//! - File I/O (if directories are preopened)
//! - Standard output

use std::env;
use std::fs;

fn main() {
    println!("Hello from WASI-enabled WebAssembly!");

    // Access environment variables
    if let Ok(node_env) = env::var("NODE_ENV") {
        println!("Environment NODE_ENV: {}", node_env);
    } else {
        println!("NODE_ENV not set");
    }

    // Access command line arguments
    let args: Vec<String> = env::args().collect();
    println!("Arguments: {:?}", args);

    // Try to read from a preopened directory (if configured)
    if let Ok(entries) = fs::read_dir("/tmp") {
        println!("Contents of /tmp:");
        for entry in entries.take(5) {
            if let Ok(entry) = entry {
                println!("  {}", entry.file_name().to_string_lossy());
            }
        }
    } else {
        println!("Could not read /tmp (directory not preopened?)");
    }

    // Try to write to a file (if write permissions granted)
    if let Ok(mut file) = fs::File::create("/tmp/wasi-test.txt") {
        use std::io::Write;
        if file.write_all(b"Hello from WASI!\n").is_ok() {
            println!("Successfully wrote to /tmp/wasi-test.txt");
        } else {
            println!("Failed to write to file");
        }
    } else {
        println!("Could not create file (write permissions not granted?)");
    }

    println!("WASI example completed!");
}