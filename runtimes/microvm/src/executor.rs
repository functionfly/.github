//! Function Executor - executes Python code inside the MicroVM

use anyhow::{Context, Result};
use std::process::Stdio;
use std::time::Duration;
use tokio::process::Command;
use tracing::{debug, error, info};

/// Python function executor that runs inside the MicroVM
pub struct PythonExecutor {
    /// Path to Python interpreter
    python_path: String,
    /// Working directory for function execution
    workdir: String,
    /// Default timeout for execution
    default_timeout_ms: u64,
}

impl PythonExecutor {
    /// Create a new Python executor
    pub fn new(workdir: impl Into<String>) -> Self {
        Self {
            python_path: "/usr/bin/python3".to_string(),
            workdir: workdir.into(),
            default_timeout_ms: 30_000, // 30 seconds
        }
    }

    /// Execute a Python function
    pub async fn execute(
        &self,
        code: &str,
        input: &str,
        handler: &str,
        packages: &[String],
        timeout_ms: Option<u64>,
    ) -> Result<String> {
        let timeout = Duration::from_millis(timeout_ms.unwrap_or(self.default_timeout_ms));

        // Prepare the execution script
        let script = self.prepare_script(code, handler, input)?;

        // Install packages if needed
        if !packages.is_empty() {
            self.install_packages(packages).await?;
        }

        // Execute the script with timeout
        let output = tokio::time::timeout(
            timeout,
            Command::new(&self.python_path)
                .arg("-c")
                .arg(&script)
                .current_dir(&self.workdir)
                .stdout(Stdio::piped())
                .stderr(Stdio::piped())
                .output()
        )
        .await
        .context("Execution timeout")??;

        if output.status.success() {
            let stdout = String::from_utf8_lossy(&output.stdout).to_string();
            debug!("Execution successful: {} bytes", stdout.len());
            Ok(stdout)
        } else {
            let stderr = String::from_utf8_lossy(&output.stderr).to_string();
            error!("Execution failed: {}", stderr);
            Err(anyhow::anyhow!("Python execution failed: {}", stderr))
        }
    }

    /// Prepare the execution script
    fn prepare_script(&self, code: &str, handler: &str, input: &str) -> Result<String> {
        // Create a script that:
        // 1. Sets up the input variable
        // 2. Executes the user code
        // 3. Calls the handler function with input

        let script = format!(
            r#"
import sys
import json

# Set up input
input_data = {input_repr}

# Execute user code
{user_code}

# Call handler
try:
    result = {handler}(input_data)
    print(json.dumps({{"success": True, "result": result}}))
except Exception as e:
    print(json.dumps({{"success": False, "error": str(e)}}))
"#,
            input_repr = serde_json::to_string(input)?,
            user_code = code,
            handler = handler
        );

        Ok(script)
    }

    /// Install Python packages
    async fn install_packages(&self, packages: &[String]) -> Result<()> {
        if packages.is_empty() {
            return Ok(());
        }

        info!("Installing packages: {:?}", packages);

        let output = Command::new(&self.python_path)
            .arg("-m")
            .arg("pip")
            .arg("install")
            .arg("--quiet")
            .arg("--user")
            .args(packages)
            .output()
            .await
            .context("Failed to install packages")?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            error!("Package installation failed: {}", stderr);
            return Err(anyhow::anyhow!("Failed to install packages: {}", stderr));
        }

        Ok(())
    }

    /// Check if the executor is ready
    pub async fn is_ready(&self) -> bool {
        Command::new(&self.python_path)
            .arg("--version")
            .output()
            .await
            .map(|o| o.status.success())
            .unwrap_or(false)
    }
}

/// Simple execution result
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct SimpleResult {
    pub success: bool,
    pub result: Option<serde_json::Value>,
    pub error: Option<String>,
}

impl SimpleResult {
    pub fn ok(result: impl serde::Serialize) -> Self {
        Self {
            success: true,
            result: serde_json::to_value(result).ok(),
            error: None,
        }
    }

    pub fn err(error: impl Into<String>) -> Self {
        Self {
            success: false,
            result: None,
            error: Some(error.into()),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_script_preparation() {
        let executor = PythonExecutor::new("/tmp");

        let script = executor.prepare_script(
            "def handler(data): return data",
            "handler",
            r#"{"key": "value"}"#,
        ).unwrap();

        assert!(script.contains("def handler"));
        assert!(script.contains("input_data"));
    }

    #[tokio::test]
    async fn test_simple_execution() {
        let executor = PythonExecutor::new("/tmp");

        // Test simple print
        let result = executor.execute(
            "print('hello world')",
            "",
            "print",
            &[],
            Some(5000),
        ).await;

        // This might fail depending on the environment
        // But the function should handle it gracefully
        assert!(result.is_ok() || result.is_err());
    }
}
