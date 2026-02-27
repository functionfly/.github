//! Vsock communication client for MicroVM I/O

use anyhow::{Context, Result};
use std::io::{Read, Write};
use std::time::Duration;
use tokio::time::timeout;
use tracing::{debug, error, info};
use vsock::{VsockAddr, VsockListener, VsockStream};

/// Vsock communication client
pub struct VsockClient {
    /// Vsock CID (Context ID) for the VM
    cid: u32,
    /// Vsock port number
    port: u32,
}

impl VsockClient {
    /// Create a new vsock client for a VM
    pub fn new(cid: u32, port: u32) -> Self {
        Self { cid, port }
    }

    /// Connect to the VM via vsock
    pub async fn connect(&self) -> Result<VsockStream> {
        let addr = VsockAddr::new(self.cid, self.port);
        info!("Connecting to VM at vsock {}:{}", self.cid, self.port);

        // Retry connection with backoff
        let mut delay = Duration::from_millis(100);
        let max_attempts = 10;

        for attempt in 1..=max_attempts {
            match VsockStream::connect(&addr) {
                Ok(stream) => {
                    debug!("Successfully connected to VM on attempt {}", attempt);
                    return Ok(stream);
                }
                Err(e) => {
                    if attempt == max_attempts {
                        error!("Failed to connect to VM after {} attempts: {}", max_attempts, e);
                        return Err(anyhow::anyhow!("Failed to connect to VM: {}", e));
                    }
                    debug!("Connection attempt {} failed, retrying in {:?}", attempt, delay);
                    tokio::time::sleep(delay).await;
                    delay = delay.saturating_mul(2).min(Duration::from_secs(1));
                }
            }
        }

        unreachable!()
    }

    /// Send a command to the VM and receive response
    pub async fn send_command(&self, command: &str) -> Result<String> {
        let mut stream = self.connect().await?;
        debug!("Sending command: {}", command.trim());

        // Send command with newline
        let command_with_nl = format!("{}\n", command);
        stream.write_all(command_with_nl.as_bytes())
            .context("Failed to send command")?;
        stream.flush().context("Failed to flush stream")?;

        // Read response
        let mut buffer = String::new();
        let mut temp_buf = [0u8; 1024];

        // Read until we get a complete JSON response (ends with newline)
        loop {
            let n = stream.read(&mut temp_buf)
                .context("Failed to read response")?;

            if n == 0 {
                break; // EOF
            }

            buffer.push_str(&String::from_utf8_lossy(&temp_buf[..n]));

            // Check if we have a complete line
            if buffer.contains('\n') {
                break;
            }
        }

        let response = buffer.trim().to_string();
        debug!("Received response: {}", response);
        Ok(response)
    }

    /// Send a JSON command and parse JSON response
    pub async fn send_json_command(&self, command: serde_json::Value) -> Result<serde_json::Value> {
        let command_str = serde_json::to_string(&command)
            .context("Failed to serialize command")?;

        let response_str = self.send_command(&command_str).await?;
        let response: serde_json::Value = serde_json::from_str(&response_str)
            .context("Failed to parse JSON response")?;

        Ok(response)
    }

    /// Ping the VM to check if it's ready
    pub async fn ping(&self) -> Result<bool> {
        let command = serde_json::json!({
            "command": "ping"
        });

        match timeout(Duration::from_secs(5), self.send_json_command(command)).await {
            Ok(Ok(response)) => {
                Ok(response.get("status").and_then(|s| s.as_str()) == Some("ok"))
            }
            _ => Ok(false),
        }
    }

    /// Load function code into the VM
    pub async fn load_function(&self, code: &str, handler: &str, packages: &[String]) -> Result<()> {
        let command = serde_json::json!({
            "command": "load",
            "code": code,
            "handler": handler,
            "packages": packages
        });

        let response = self.send_json_command(command).await?;

        if response.get("status").and_then(|s| s.as_str()) == Some("loaded") {
            Ok(())
        } else {
            Err(anyhow::anyhow!("Failed to load function: {:?}", response))
        }
    }

    /// Execute function with input data
    pub async fn execute_function(&self, input: serde_json::Value, packages: &[String]) -> Result<serde_json::Value> {
        let command = serde_json::json!({
            "command": "execute",
            "input": input,
            "packages": packages
        });

        let response = self.send_json_command(command).await?;

        if response.get("success").and_then(|s| s.as_bool()) == Some(true) {
            response.get("result").cloned()
                .ok_or_else(|| anyhow::anyhow!("Missing result in successful response"))
        } else {
            let error = response.get("error")
                .and_then(|e| e.as_str())
                .unwrap_or("Unknown error");
            Err(anyhow::anyhow!("Function execution failed: {}", error))
        }
    }
}

/// Vsock server for listening to VM connections (if needed)
pub struct VsockServer {
    listener: VsockListener,
}

impl VsockServer {
    /// Create a new vsock server
    pub fn new(port: u32) -> Result<Self> {
        let addr = VsockAddr::new(vsock::VMADDR_CID_ANY, port);
        let listener = VsockListener::bind(&addr)
            .context("Failed to bind vsock listener")?;

        info!("Vsock server listening on port {}", port);
        Ok(Self { listener })
    }

    /// Accept a connection
    pub async fn accept(&mut self) -> Result<VsockStream> {
        // Note: VsockListener doesn't have async accept, so we use blocking
        // Clone the listener to avoid lifetime issues
        let mut listener = self.listener.try_clone()
            .context("Failed to clone vsock listener")?;

        tokio::task::spawn_blocking(move || {
            listener.accept()
                .map(|(stream, _addr)| stream)
                .context("Failed to accept vsock connection")
        })
        .await
        .context("Failed to spawn blocking accept task")?
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    #[ignore] // Requires a running VM
    async fn test_vsock_client_creation() {
        let client = VsockClient::new(3, 1234); // CID 3, port 1234
        assert_eq!(client.cid, 3);
        assert_eq!(client.port, 1234);
    }
}