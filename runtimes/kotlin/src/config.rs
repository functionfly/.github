//! Runtime configuration for Kotlin/JVM execution

use serde::{Deserialize, Serialize};
use std::time::Duration;

/// JVM configuration for Kotlin execution
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct JvmConfig {
    /// Path to the JVM binary (java/kotlin)
    pub java_path: String,
    /// JVM options/flags
    pub jvm_options: Vec<String>,
    /// Working directory for execution
    pub working_dir: Option<String>,
    /// Classpath for the application
    pub classpath: Vec<String>,
    /// Main class to execute
    pub main_class: String,
    /// Enable JVM security manager
    pub enable_security_manager: bool,
    /// Custom properties
    pub system_properties: Vec<(String, String)>,
}

impl Default for JvmConfig {
    fn default() -> Self {
        Self {
            java_path: "java".to_string(),
            jvm_options: vec![
                "-Xmx256m".to_string(),
                "-Xms64m".to_string(),
                "-XX:+UseG1GC".to_string(),
                "-XX:+ExitOnOutOfMemoryError".to_string(),
                "-Djava.security.manager=allow".to_string(),
                "-Dfile.encoding=UTF-8".to_string(),
            ],
            working_dir: None,
            classpath: vec![],
            main_class: "kotlin.KotlinMain".to_string(),
            enable_security_manager: true,
            system_properties: vec![
                ("kotlin.diagnostics.enabled".to_string(), "false".to_string()),
                ("kotlin.incremental".to_string(), "false".to_string()),
            ],
        }
    }
}

/// Execution limits for a single function execution
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct ExecutionLimits {
    /// Maximum memory in MB
    pub max_memory_mb: u64,
    /// Maximum CPU time in seconds
    pub max_cpu_time_secs: u64,
    /// Maximum wall time (including I/O)
    pub max_wall_time_secs: u64,
    /// Maximum output size in bytes
    pub max_output_bytes: usize,
    /// Maximum number of classes that can be loaded
    pub max_classes: usize,
    /// Maximum total network request size
    pub max_network_request_bytes: usize,
    /// Enable disk I/O (sandboxed)
    pub allow_disk_io: bool,
    /// Allow network access
    pub allow_net: bool,
    /// Maximum file size for reading
    pub max_file_size_bytes: usize,
    /// Maximum number of threads
    pub max_threads: usize,
}

impl Default for ExecutionLimits {
    fn default() -> Self {
        Self {
            max_memory_mb: 256,
            max_cpu_time_secs: 10,
            max_wall_time_secs: 30,
            max_output_bytes: 1024 * 1024, // 1MB
            max_classes: 1000,
            max_network_request_bytes: 10 * 1024 * 1024, // 10MB
            allow_disk_io: false,
            allow_net: true,
            max_file_size_bytes: 10 * 1024 * 1024, // 10MB
            max_threads: 4,
        }
    }
}

/// Main runtime configuration
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct RuntimeConfig {
    /// Execution limits
    pub limits: ExecutionLimits,
    /// Security policy
    pub security: SecurityPolicy,
    /// JVM configuration
    pub jvm: JvmConfig,
    /// Enable WASM sandbox
    pub use_sandbox: bool,
    /// Maximum concurrent executions
    pub max_concurrent: usize,
    /// Execution timeout
    pub default_timeout: Duration,
    /// Enable metrics collection
    pub enable_metrics: bool,
    /// Rate limit for requests
    pub requests_per_minute: Option<u32>,
}

impl Default for RuntimeConfig {
    fn default() -> Self {
        Self {
            limits: ExecutionLimits::default(),
            security: SecurityPolicy::default(),
            jvm: JvmConfig::default(),
            use_sandbox: true,
            max_concurrent: 100,
            default_timeout: Duration::from_secs(30),
            enable_metrics: true,
            requests_per_minute: Some(1000),
        }
    }
}

/// Security policy for the runtime
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct SecurityPolicy {
    /// Blocked package patterns
    pub blocked_packages: Vec<String>,
    /// Allowed hosts for network requests (empty = all allowed if allow_net)
    pub allowed_hosts: Vec<String>,
    /// Blocked hosts (takes precedence over allowed_hosts)
    pub blocked_hosts: Vec<String>,
    /// Environment variables to expose (empty = none)
    pub env_whitelist: Vec<String>,
    /// Enable secure sandbox mode
    pub sandbox_enabled: bool,
    /// Allow ProcessBuilder/System exec
    pub allow_process_exec: bool,
    /// Allow Reflection
    pub allow_reflection: bool,
    /// Allow JNI/native code
    pub allow_jni: bool,
    /// Maximum depth for reflection
    pub max_reflection_depth: usize,
    /// Blocked class patterns
    pub blocked_classes: Vec<String>,
}

impl Default for SecurityPolicy {
    fn default() -> Self {
        Self {
            blocked_packages: vec![
                "java.lang.Process".to_string(),
                "java.lang.System".to_string(),
                "java.lang.Runtime".to_string(),
                "java.lang.reflect".to_string(),
                "java.io.File".to_string(),
                "java.nio.file".to_string(),
                "java.net.Socket".to_string(),
                "java.net.ServerSocket".to_string(),
                "java.net.URL".to_string(),
                "java.lang.ClassLoader".to_string(),
                "sun.misc".to_string(),
                "jdk.internal".to_string(),
            ],
            allowed_hosts: vec![],
            blocked_hosts: vec![
                "169.254.169.254".to_string(), // AWS metadata
                "metadata.google.internal".to_string(), // GCP metadata
                "metadata.azure.com".to_string(), // Azure metadata
                "100.100.100.200".to_string(), // Alibaba Cloud metadata
            ],
            env_whitelist: vec![],
            sandbox_enabled: true,
            allow_process_exec: false,
            allow_reflection: false,
            allow_jni: false,
            max_reflection_depth: 5,
            blocked_classes: vec![
                "java.lang.Thread".to_string(),
                "java.lang.ThreadGroup".to_string(),
                "java.lang.ProcessBuilder".to_string(),
                "java.lang.System".to_string(),
                "java.lang.Runtime".to_string(),
                "java.lang.Class".to_string(),
                "java.lang.ClassLoader".to_string(),
            ],
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_config() {
        let config = RuntimeConfig::default();
        assert!(config.use_sandbox);
        assert_eq!(config.max_concurrent, 100);
    }

    #[test]
    fn test_default_limits() {
        let limits = ExecutionLimits::default();
        assert_eq!(limits.max_memory_mb, 256);
        assert_eq!(limits.max_cpu_time_secs, 10);
    }

    #[test]
    fn test_default_security_policy() {
        let policy = SecurityPolicy::default();
        assert!(policy.sandbox_enabled);
        assert!(!policy.allow_process_exec);
        assert!(policy.blocked_packages.contains(&"java.lang.Process".to_string()));
    }

    #[test]
    fn test_jvm_config_default() {
        let jvm = JvmConfig::default();
        assert_eq!(jvm.java_path, "java");
        assert!(jvm.enable_security_manager);
    }
}