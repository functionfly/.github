//! Webhook Notifier Example
//!
//! This example demonstrates how to use the webhook capability in FunctionFly.
//! The function sends HTTP POST requests to a configured webhook URL.

// No external dependencies needed

/// Handler function that gets called when the function is invoked
#[no_mangle]
pub extern "C" fn handler() -> i32 {
    // Get webhook URL from environment
    let url = match get_env_var("WEBHOOK_URL") {
        Some(url) => url,
        None => {
            eprintln!("WEBHOOK_URL environment variable not set");
            return -1;
        }
    };

    // Get notification type from environment
    let notification_type = get_env_var("NOTIFICATION_TYPE").unwrap_or_else(|| "notification".to_string());

    // Create JSON payload
    let payload = format!(
        r#"{{
            "type": "{}",
            "message": "Hello from FunctionFly webhook example!",
            "timestamp": "2024-01-01T00:00:00Z",
            "source": "functionfly-webhook-example"
        }}"#,
        notification_type
    );

    // Prepare headers as JSON string
    let headers = r#"{
        "Content-Type": "application/json",
        "X-FunctionFly-Webhook": "example"
    }"#;

    // Send webhook using the host function
    let result = unsafe {
        webhook_send(
            url.as_ptr() as i32,
            url.len() as i32,
            b"POST".as_ptr() as i32,
            4, // "POST".len()
            payload.as_ptr() as i32,
            payload.len() as i32,
            headers.as_ptr() as i32,
            headers.len() as i32,
        )
    };

    match result {
        0 => {
            println!("Webhook sent successfully");
            0 // Success
        }
        -1 => {
            eprintln!("Invalid URL provided");
            -1
        }
        -2 => {
            eprintln!("Invalid HTTP method");
            -2
        }
        -3 => {
            eprintln!("Invalid payload");
            -3
        }
        -4 => {
            eprintln!("Invalid headers JSON");
            -4
        }
        -5 => {
            eprintln!("Failed to parse headers JSON");
            -5
        }
        -6 => {
            eprintln!("Webhook request failed");
            -6
        }
        _ => {
            eprintln!("Unknown error occurred: {}", result);
            result
        }
    }
}

/// Get an environment variable using WASI
fn get_env_var(name: &str) -> Option<String> {
    // In a real implementation, you'd use WASI environ_get/environ_sizes_get
    // For this example, we'll simulate getting environment variables
    // This is a simplified version - in practice you'd need to implement
    // proper WASI environment variable access

    // For demonstration, we'll return some hardcoded values
    match name {
        "WEBHOOK_URL" => Some("https://httpbin.org/post".to_string()),
        "NOTIFICATION_TYPE" => Some("alert".to_string()),
        _ => None,
    }
}

/// External function declarations for FunctionFly host functions
extern "C" {
    /// Send a webhook request
    /// Returns 0 on success, negative values on error
    fn webhook_send(
        url_ptr: i32,
        url_len: i32,
        method_ptr: i32,
        method_len: i32,
        payload_ptr: i32,
        payload_len: i32,
        headers_ptr: i32,
        headers_len: i32,
    ) -> i32;
}
