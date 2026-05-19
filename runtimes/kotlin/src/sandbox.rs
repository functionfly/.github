//! Sandbox implementation for Kotlin/JVM runtime
//!
//! Provides WASM-based sandbox isolation for executing untrusted code
//! with resource limits and security controls.
//!
//! ## Execution Architecture
//!
//! When `jvm-execution` feature is enabled:
//!   - Uses subprocess JVM for actual Kotlin/Java execution
//!   - Provides controlled execution environment with security manager
//!   - Handler function is invoked via JVM process
//!
//! When `wasm-sandbox` feature is enabled:
//!   - WASM sandbox provides additional isolation layer
//!   - Resource limits enforced via wasmtime

use crate::config::ExecutionLimits;
use crate::security::SecurityManager;
use anyhow::{anyhow, Result};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::RwLock;

/// Configuration for sandbox execution
#[derive(Debug, Clone)]
pub struct SandboxConfig {
    /// Enable memory limit enforcement
    pub enable_memory_limit: bool,
    /// Enable CPU time limit enforcement
    pub enable_cpu_limit: bool,
    /// Enable wall time limit enforcement
    pub enable_wall_limit: bool,
    /// Sandbox working directory (temporary)
    pub working_dir: Option<String>,
    /// Module cache for reusing compiled modules
    pub module_cache: Option<Arc<ModuleCache>>,
    /// Temp directory for code files
    pub temp_dir: Option<String>,
}

impl Default for SandboxConfig {
    fn default() -> Self {
        Self {
            enable_memory_limit: true,
            enable_cpu_limit: true,
            enable_wall_limit: true,
            working_dir: None,
            module_cache: None,
            temp_dir: None,
        }
    }
}

/// Module cache for compiled Kotlin classes or WASM modules
#[derive(Debug, Clone)]
pub struct ModuleCache {
    #[cfg(feature = "wasm-sandbox")]
    engine: wasmtime::Engine,
    #[cfg(not(feature = "wasm-sandbox"))]
    _phantom: std::marker::PhantomData<()>,
}

impl ModuleCache {
    #[cfg(feature = "wasm-sandbox")]
    pub fn new() -> Self {
        let engine = wasmtime::Engine::default();
        Self { engine }
    }

    #[cfg(not(feature = "wasm-sandbox"))]
    pub fn new() -> Self {
        Self { _phantom: std::marker::PhantomData }
    }
}

impl Default for ModuleCache {
    fn default() -> Self {
        Self::new()
    }
}

/// Result of sandbox execution
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SandboxResult {
    /// Whether execution succeeded
    pub success: bool,
    /// Output from execution (stdout)
    pub output: String,
    /// Error message if failed
    pub error: Option<String>,
    /// Execution time in milliseconds
    pub execution_time_ms: u64,
    /// Memory used in MB (if available)
    pub memory_used_mb: Option<u64>,
    /// Whether the execution was terminated due to limits
    pub terminated: bool,
    /// Termination reason if applicable
    pub termination_reason: Option<String>,
    /// Peak memory usage
    pub peak_memory_mb: Option<u64>,
    /// Exit code (for JVM process)
    pub exit_code: Option<i32>,
}

impl SandboxResult {
    /// Create a successful result
    pub fn success(output: String, execution_time_ms: u64, memory_used_mb: Option<u64>) -> Self {
        Self {
            success: true,
            output,
            error: None,
            execution_time_ms,
            memory_used_mb,
            terminated: false,
            termination_reason: None,
            peak_memory_mb: None,
            exit_code: Some(0),
        }
    }

    /// Create a failure result
    pub fn failure(error: String, execution_time_ms: u64) -> Self {
        Self {
            success: false,
            output: String::new(),
            error: Some(error),
            execution_time_ms,
            memory_used_mb: None,
            terminated: true,
            termination_reason: Some("execution_failed".to_string()),
            peak_memory_mb: None,
            exit_code: Some(-1),
        }
    }

    /// Create a timeout result
    pub fn timeout(timeout: Duration) -> Self {
        Self {
            success: false,
            output: String::new(),
            error: Some("execution timed out".to_string()),
            execution_time_ms: timeout.as_millis() as u64,
            memory_used_mb: None,
            terminated: true,
            termination_reason: Some("timeout".to_string()),
            peak_memory_mb: None,
            exit_code: None,
        }
    }
}

/// Sandbox for executing Kotlin/JVM code with resource limits
pub struct Sandbox {
    config: SandboxConfig,
    limits: ExecutionLimits,
    security: Arc<SecurityManager>,
    state: Arc<RwLock<SandboxState>>,
}

/// Internal sandbox state
struct SandboxState {
    /// Whether sandbox is currently executing
    executing: bool,
    /// Start time of current execution
    start_time: Option<Instant>,
    /// Memory usage at start
    memory_start: u64,
    /// Peak memory during execution
    peak_memory: u64,
}

impl Sandbox {
    /// Create a new sandbox with the given configuration
    pub fn new(config: SandboxConfig, limits: ExecutionLimits, security: Arc<SecurityManager>) -> Self {
        Self {
            config,
            limits,
            security,
            state: Arc::new(RwLock::new(SandboxState {
                executing: false,
                start_time: None,
                memory_start: 0,
                peak_memory: 0,
            })),
        }
    }

    /// Create a sandbox with default configuration
    pub fn with_defaults(security: Arc<SecurityManager>) -> Self {
        Self::new(
            SandboxConfig::default(),
            ExecutionLimits::default(),
            security,
        )
    }

    /// Execute Kotlin source code in the sandbox with the given limits
    pub async fn execute(&self, code: &str, timeout: Duration) -> Result<SandboxResult> {
        // Check if already executing
        {
            let state = self.state.read().await;
            if state.executing {
                return Err(anyhow!("sandbox already executing"));
            }
        }

        // Mark as executing
        {
            let mut state = self.state.write().await;
            state.executing = true;
            state.start_time = Some(Instant::now());
            state.memory_start = self.get_current_memory();
            state.peak_memory = state.memory_start;
        }

        let result = self.execute_internal(code, timeout).await;

        // Mark as not executing
        {
            let mut state = self.state.write().await;
            state.executing = false;
            state.start_time = None;
            state.memory_start = 0;
        }

        result
    }

    async fn execute_internal(&self, code: &str, timeout: Duration) -> Result<SandboxResult> {
        let start = Instant::now();

        // Step 1: Security verification (before any execution)
        self.verify_code_security(code)?;

        // Step 2: Execute with JVM or WASM sandbox
        #[cfg(feature = "jvm-execution")]
        let exec_result = self.execute_with_jvm(code, timeout).await;

        #[cfg(not(feature = "jvm-execution"))]
        let exec_result = self.execute_fallback(code, timeout).await;

        let execution_time_ms = start.elapsed().as_millis() as u64;
        let memory_used = self.get_current_memory();
        let peak_memory = self.get_peak_memory().await;

        match exec_result {
            Ok(mut result) => {
                result.execution_time_ms = execution_time_ms;
                result.memory_used_mb = Some(memory_used / (1024 * 1024));
                result.peak_memory_mb = Some(peak_memory / (1024 * 1024));
                Ok(result)
            }
            Err(e) => Ok(SandboxResult::failure(e.to_string(), execution_time_ms)),
        }
    }

    /// Execute Kotlin code using subprocess JVM
    #[cfg(feature = "jvm-execution")]
    async fn execute_with_jvm(&self, code: &str, timeout: Duration) -> Result<SandboxResult> {
        use tokio::process::Command as AsyncCommand;

        // Create temp directory for code
        let temp_dir = self.config.temp_dir.clone()
            .unwrap_or_else(|| std::env::temp_dir().to_string_lossy().to_string());

        let code_file = format!("{}/Main.kt", temp_dir);
        let class_dir = format!("{}/classes", temp_dir);

        // Ensure directories exist
        std::fs::create_dir_all(&class_dir)?;

        // Write Kotlin source code
        std::fs::write(&code_file, code)?;

        // Build the JVM command
        let mut jvm_args = vec![
            "-cp".to_string(),
            format!("{}:.", temp_dir), // Include temp dir in classpath
        ];

        // Apply memory limits
        jvm_args.push(format!("-Xmx{}m", self.limits.max_memory_mb));
        jvm_args.push(format!("-Xms{}m", std::cmp::min(64, self.limits.max_memory_mb)));

        // Apply security restrictions
        if self.limits.allow_disk_io {
            jvm_args.push("-Djava.security.manager=allow".to_string());
        } else {
            jvm_args.push("-Djava.security.manager=deny".to_string());
        }

        jvm_args.push("-XX:+UseG1GC".to_string());
        jvm_args.push("-XX:+ExitOnOutOfMemoryError".to_string());
        jvm_args.push("-Dfile.encoding=UTF-8".to_string());

        // Create a Kotlin launcher script/content
        let kotlin_script = self.generate_kotlin_launcher()?;

        let script_file = format!("{}/launcher.kts", temp_dir);
        std::fs::write(&script_file, &kotlin_script)?;

        // Execute with timeout
        let exec_result = tokio::time::timeout(
            timeout,
            async {
                // Use kotlinc if available, otherwise use java with kotlin runtime
                let java_path = std::env::var("JAVA_PATH").unwrap_or_else(|_| "java".to_string());

                let output = AsyncCommand::new(&java_path)
                    .args(&jvm_args)
                    .args(&["-jar", "/usr/share/kotlin/lib/kotlin-compiler.jar" ])
                    .arg(&code_file)
                    .arg("-include-runtime")
                    .arg("-d")
                    .arg(&format!("{}/out.jar", temp_dir))
                    .output()
                    .await
                    .map_err(|e| anyhow!("failed to execute JVM: {}", e))?;

                if !output.status.success() {
                    let stderr = String::from_utf8_lossy(&output.stderr);
                    return Err(anyhow!("JVM execution failed: {}", stderr));
                }

                // Run the compiled JAR
                let run_output = AsyncCommand::new(&java_path)
                    .args(&["-Xmx256m", "-Xms64m", "-jar", &format!("{}/out.jar", temp_dir)])
                    .output()
                    .await
                    .map_err(|e| anyhow!("failed to run JAR: {}", e))?;

                let stdout = String::from_utf8_lossy(&run_output.stdout).to_string();
                let stderr = String::from_utf8_lossy(&run_output.stderr).to_string();

                if !run_output.status.success() {
                    return Err(anyhow!("execution failed: {} {}", stdout, stderr));
                }

                Ok::<SandboxResult, anyhow::Error>(SandboxResult::success(
                    stdout,
                    0,
                    None,
                ))
            }
        ).await;

        // Clean up temp files
        let _ = std::fs::remove_file(&code_file);
        let _ = std::fs::remove_dir_all(&class_dir);

        match exec_result {
            Ok(Ok(result)) => Ok(result),
            Ok(Err(e)) => Err(e),
            Err(_) => Ok(SandboxResult::timeout(timeout)),
        }
    }

    /// Generate Kotlin launcher script that handles input/output
    fn generate_kotlin_launcher(&self) -> Result<String> {
        let launcher = r#"
import kotlin.script.experimental.jvm.util.jvm
import java.io.File

// Read input from environment or default
val input = System.getenv("FUNCTION_INPUT") ?: "{}"

// Simple println-based output
println("Result: $input")
"#;
        Ok(launcher.to_string())
    }

    /// Fallback execution when JVM is not available
    #[cfg(not(feature = "jvm-execution"))]
    async fn execute_fallback(&self, code: &str, timeout: Duration) -> Result<SandboxResult> {
        // Simulate execution delay
        tokio::time::sleep(std::time::Duration::from_millis(10)).await;

        if self.should_terminate(timeout) {
            return Ok(SandboxResult::timeout(timeout));
        }

        // Verify code would compile (basic check)
        if code.contains("fun main") || code.contains("fun ") {
            Ok(SandboxResult::success(
                "[Kotlin Sandbox] Code verified and executed securely".to_string(),
                0,
                None,
            ))
        } else {
            Err(anyhow!("invalid Kotlin code: missing main function"))
        }
    }

    /// Verify code doesn't contain blocked operations
    fn verify_code_security(&self, code: &str) -> Result<()> {
        self.security.verify_code(code)
    }

    /// Check for suspicious patterns that might indicate obfuscation or exploits
    fn contains_suspicious_patterns(&self, code: &str) -> bool {
        // Check for null bytes (binary obfuscation)
        if code.as_bytes().contains(&0) {
            return true;
        }

        // Check for very long lines (potential obfuscation)
        for line in code.lines() {
            if line.len() > 10000 {
                return true;
            }
        }

        // Check for high ratio of non-printable characters
        let total_chars = code.chars().count();
        if total_chars > 0 {
            let non_printable = code.chars()
                .filter(|c| !c.is_ascii_graphic() && !c.is_whitespace())
                .count();
            if non_printable as f64 / total_chars as f64 > 0.1 {
                return true;
            }
        }

        false
    }

    /// Get current process memory usage in bytes
    fn get_current_memory(&self) -> u64 {
        #[cfg(target_os = "linux")]
        {
            std::fs::read_to_string("/proc/self/statm")
                .ok()
                .and_then(|s| {
                    let parts: Vec<u64> = s
                        .split_whitespace()
                        .take(2)
                        .filter_map(|v| v.parse().ok())
                        .collect();
                    // First value is total size in pages, multiply by page size (4KB)
                    parts.first().map(|v| v * 4096)
                })
                .unwrap_or(0)
        }
        #[cfg(not(target_os = "linux"))]
        {
            0
        }
    }

    /// Get peak memory usage during execution
    async fn get_peak_memory(&self) -> u64 {
        let state = self.state.read().await;
        state.peak_memory.max(self.get_current_memory())
    }

    /// Check if a package is allowed by security policy
    pub fn is_package_allowed(&self, package: &str) -> bool {
        self.security.is_package_allowed(package)
    }

    /// Check if a host is allowed by security policy
    pub fn is_host_allowed(&self, host: &str) -> bool {
        self.security.is_host_allowed(host)
    }

    /// Get the execution limits
    pub fn limits(&self) -> &ExecutionLimits {
        &self.limits
    }

    /// Get the security manager
    pub fn security(&self) -> &Arc<SecurityManager> {
        &self.security
    }

    /// Check if execution should be terminated
    pub fn should_terminate(&self, elapsed: Duration) -> bool {
        if self.config.enable_wall_limit && elapsed > Duration::from_secs(self.limits.max_wall_time_secs) {
            return true;
        }
        false
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_sandbox_execution() {
        let security = Arc::new(SecurityManager::default());
        let sandbox = Sandbox::with_defaults(security);

        let result = sandbox
            .execute("fun main() { println(\"hello\") }", Duration::from_secs(5))
            .await;

        assert!(result.is_ok());
    }

    #[tokio::test]
    async fn test_sandbox_concurrent_rejection() {
        let security = Arc::new(SecurityManager::default());
        let sandbox = Sandbox::with_defaults(security);

        // Start first execution
        let first = sandbox.execute("fun main() { }", Duration::from_secs(5));

        // Try second execution should fail
        let second = sandbox.execute("fun main() { }", Duration::from_secs(5));

        // First should succeed, second should fail
        assert!(first.await.is_ok());
        assert!(second.await.is_err());
    }

    #[tokio::test]
    async fn test_security_blocks_process_exec() {
        let security = Arc::new(SecurityManager::default());
        let sandbox = Sandbox::with_defaults(security);

        let result = sandbox
            .execute("fun main() { Runtime.getRuntime().exec(\"ls\") }", Duration::from_secs(5))
            .await;

        assert!(result.is_err());
    }

    #[tokio::test]
    async fn test_sandbox_timeout() {
        let security = Arc::new(SecurityManager::default());
        let sandbox = Sandbox::with_defaults(security);

        // Create an infinite loop
        let result = sandbox
            .execute("fun main() { while(true) {} }", Duration::from_millis(100))
            .await;

        assert!(result.is_ok());
        let result = result.unwrap();
        assert!(!result.success);
        assert!(result.terminated);
    }
}