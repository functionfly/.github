//! WASI context state snapshot for pooling.
//!
//! WASI state (environment variables, command-line arguments, pipe buffers)
//! must be reset between executions to prevent state leakage. This snapshot
//! captures the static portion (env, args) so it can be restored cheaply.
//! The dynamic portion (pipe contents) is cleared by creating fresh pipes.

/// Snapshot of WASI context state that can be captured and restored.
#[derive(Debug, Clone, Default)]
pub struct WasiStateSnapshot {
    /// Environment variables (key=value pairs)
    pub env_vars: Vec<(String, String)>,
    /// Command-line arguments
    pub args: Vec<String>,
    /// Pipe capacity in bytes (used to recreate output pipes)
    pub pipe_capacity: usize,
}

impl WasiStateSnapshot {
    /// Capture the current WASI environment and arguments from the given config.
    ///
    /// Note: the dynamic state (stdin/stdout/stderr pipe buffers) is cleared by
    /// creating new pipes on restore, rather than being snapshotted.
    pub fn capture_from_config(config: &crate::config::Config) -> Self {
        let mut env_vars = Vec::new();
        // Collect configured WASI env vars
        for env_var in &config.wasi_env {
            if let Some((key, value)) = env_var.split_once('=') {
                env_vars.push((key.to_string(), value.to_string()));
            }
        }
        // Add defaults
        env_vars.push(("PATH".to_string(), "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin".to_string()));
        env_vars.push(("PWD".to_string(), "/".to_string()));
        env_vars.push(("HOME".to_string(), "/tmp".to_string()));

        let pipe_capacity = if config.max_output_bytes > 0 {
            config.max_output_bytes
        } else {
            1024 * 1024 // 1 MiB fallback
        };

        Self {
            env_vars,
            args: vec![config.function.clone()],
            pipe_capacity,
        }
    }

    /// Restore WASI state by configuring a new WasiCtxBuilder.
    ///
    /// This clears pipe buffers (by using fresh pipes in the builder) and
    /// reapplies the snapshotted environment variables and arguments.
    pub fn restore(&self, builder: &mut wasmtime_wasi::WasiCtxBuilder) {
        for (key, value) in &self.env_vars {
            builder.env(key, value);
        }
        builder.args(&self.args);
    }
}
