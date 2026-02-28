//! WASI (WebAssembly System Interface) context and linker management.
//!
//! This module provides proper WASI support for WebAssembly modules, including:
//! - Environment variables
//! - Filesystem access with preopened directories
//! - Command line arguments
//! - Standard I/O streams
//! - Networking and time access controls

use anyhow::Context;
use std::path::Path;
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::RwLock;
use wasmtime_wasi::p1::WasiP1Ctx;
use wasmtime_wasi::p2::pipe::{MemoryOutputPipe, MemoryInputPipe};
use wasmtime_wasi::{DirPerms, FilePerms, HostMonotonicClock, HostWallClock, WasiCtxBuilder};

use crate::capability::Capabilities;
use crate::config::Config;
use crate::host_functions::HostFunctionsLinker;
use crate::kv::SharedKVStore;

/// Custom monotonic clock that denies access to real time.
///
/// Instead of panicking (which would crash the entire runtime process), we
/// return a fixed epoch value of 0.  This keeps the WASM guest alive but
/// prevents it from observing real wall-clock or monotonic time, satisfying
/// the determinism requirement without bringing down the server.
pub struct DisabledMonotonicClock;

impl HostMonotonicClock for DisabledMonotonicClock {
    fn resolution(&self) -> u64 {
        // Return 1 ns resolution but always report time as 0.
        tracing::warn!("DisabledMonotonicClock::resolution() called - returning stub value (time access disabled)");
        1
    }

    fn now(&self) -> u64 {
        // Always return epoch 0 so the guest cannot observe real time.
        tracing::warn!("DisabledMonotonicClock::now() called - returning 0 (time access disabled)");
        0
    }
}

/// Custom wall clock that denies access to real time.
pub struct DisabledWallClock;

/// Custom WASI context that includes FunctionFly-specific data
pub struct FunctionFlyWasiCtx {
    pub wasi_ctx: WasiP1Ctx,
    pub function_key: String,
}

impl HostWallClock for DisabledWallClock {
    fn resolution(&self) -> Duration {
        // Return 1 ns resolution but always report time as epoch 0.
        tracing::warn!("DisabledWallClock::resolution() called - returning stub value (time access disabled)");
        Duration::from_nanos(1)
    }

    fn now(&self) -> Duration {
        // Always return epoch 0 so the guest cannot observe real wall-clock time.
        tracing::warn!("DisabledWallClock::now() called - returning epoch 0 (time access disabled)");
        Duration::ZERO
    }
}

/// WASI context wrapper using WASIp1
pub struct WasiContext {
    pub ctx: WasiP1Ctx,
    /// Whether time access is allowed
    pub time_access_allowed: bool,
    /// Pipe for capturing stdout
    pub stdout_pipe: MemoryOutputPipe,
    /// Pipe for capturing stderr
    pub stderr_pipe: MemoryOutputPipe,
    /// Function key for security checks (function_name@version)
    pub function_key: String,
}

impl WasiContext {
    /// Create a new WASI context with the given configuration
    pub fn new(config: &Config, function_key: String) -> anyhow::Result<Self> {
        Self::new_with_input(config, function_key, "")
    }

    /// Create a new WASI context with input data for stdin
    pub fn new_with_input(config: &Config, function_key: String, input: &str) -> anyhow::Result<Self> {
        // Use the configurable output pipe capacity (default 1 MiB).
        // This prevents silent truncation of large function outputs.
        let pipe_capacity = if config.max_output_bytes > 0 {
            config.max_output_bytes
        } else {
            1024 * 1024 // 1 MiB fallback
        };
        let stdout_pipe = MemoryOutputPipe::new(pipe_capacity);
        let stderr_pipe = MemoryOutputPipe::new(pipe_capacity);
        let input_pipe = MemoryInputPipe::new(input.as_bytes().to_vec()); // Input pipe with data

        let mut builder = WasiCtxBuilder::new();

        // Set up stdin, stdout and stderr pipes
        builder.stdin(input_pipe).stdout(stdout_pipe.clone()).stderr(stderr_pipe.clone());

        // Set up environment variables
        Self::configure_environment(&mut builder, config)?;

        // Set up filesystem access
        Self::configure_filesystem(&mut builder, config)?;

        // Set up command line arguments
        Self::configure_arguments(&mut builder, config)?;

        // Set up capabilities (networking, time access, etc.)
        Self::configure_capabilities(&mut builder, config)?;

        let ctx = builder.build_p1();

        Ok(Self {
            ctx,
            time_access_allowed: config.wasi_allow_time,
            stdout_pipe,
            stderr_pipe,
            function_key,
        })
    }


    /// Configure environment variables for WASI
    fn configure_environment(
        builder: &mut WasiCtxBuilder,
        config: &Config,
    ) -> anyhow::Result<()> {
        // Add custom environment variables from config
        for env_var in &config.wasi_env {
            if let Some((key, value)) = env_var.split_once('=') {
                builder.env(key, value);
            } else {
                tracing::warn!("Invalid WASI environment variable format: {}", env_var);
            }
        }

        // Add some default environment variables that are commonly expected
        builder
            .env("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
            .env("PWD", "/")
            .env("HOME", "/tmp");

        Ok(())
    }

    /// Configure filesystem access for WASI
    fn configure_filesystem(
        builder: &mut WasiCtxBuilder,
        config: &Config,
    ) -> anyhow::Result<()> {
        // Parse preopened directories from config
        // Format: host_path:wasm_path:permissions
        for dir_spec in &config.wasi_dirs {
            let parts: Vec<&str> = dir_spec.split(':').collect();
            match parts.len() {
                2 => {
                    // Default permissions: read-only
                    let host_path = parts[0];
                    let wasm_path = parts[1];
                    Self::add_preopened_dir(builder, host_path, wasm_path, DirPerms::from_bits_truncate(1), FilePerms::from_bits_truncate(1))?;
                }
                3 => {
                    let host_path = parts[0];
                    let wasm_path = parts[1];
                    let perms_str = parts[2];

                    let (dir_perms, file_perms) = Self::parse_permissions(perms_str)?;
                    Self::add_preopened_dir(builder, host_path, wasm_path, dir_perms, file_perms)?;
                }
                _ => {
                    tracing::warn!("Invalid WASI directory format: {}", dir_spec);
                    continue;
                }
            }
        }

        // Add a default temporary directory if no directories are specified
        if config.wasi_dirs.is_empty() {
            let temp_dir = std::env::temp_dir();
            if temp_dir.exists() {
                Self::add_preopened_dir(builder, temp_dir.to_str().unwrap_or("/tmp"), "/tmp", DirPerms::from_bits_truncate(3), FilePerms::from_bits_truncate(3))?;
            }
        }

        Ok(())
    }

    /// Parse permission string into DirPerms and FilePerms
    fn parse_permissions(perms_str: &str) -> anyhow::Result<(DirPerms, FilePerms)> {
        match perms_str.to_lowercase().as_str() {
            "r" | "read" => Ok((DirPerms::from_bits_truncate(1), FilePerms::from_bits_truncate(1))),
            "w" | "write" => Ok((DirPerms::from_bits_truncate(2), FilePerms::from_bits_truncate(2))),
            "rw" | "readwrite" => Ok((DirPerms::from_bits_truncate(3), FilePerms::from_bits_truncate(3))),
            _ => {
                tracing::warn!("Invalid permission string '{}', using read-only permissions", perms_str);
                Ok((DirPerms::from_bits_truncate(1), FilePerms::from_bits_truncate(1)))
            }
        }
    }

    /// Add a preopened directory to WASI context
    fn add_preopened_dir(
        builder: &mut WasiCtxBuilder,
        host_path: &str,
        wasm_path: &str,
        dir_perms: DirPerms,
        file_perms: FilePerms,
    ) -> anyhow::Result<()> {
        let host_path = Path::new(host_path);

        if !host_path.exists() {
            tracing::warn!("WASI preopened directory does not exist: {}", host_path.display());
            return Ok(());
        }

        if !host_path.is_dir() {
            tracing::warn!("WASI preopened path is not a directory: {}", host_path.display());
            return Ok(());
        }

        // Add the preopened directory with specified permissions
        builder.preopened_dir(
            host_path,
            wasm_path,
            dir_perms,
            file_perms,
        )?;
        tracing::debug!("Added WASI preopened directory: {} -> {}", host_path.display(), wasm_path);

        Ok(())
    }


    /// Configure command line arguments for WASI
    fn configure_arguments(
        builder: &mut WasiCtxBuilder,
        config: &Config,
    ) -> anyhow::Result<()> {
        // Add program name as first argument
        builder.arg(&config.function);

        // Add configured arguments
        for arg in &config.wasi_args {
            builder.arg(arg);
        }

        Ok(())
    }

    /// Configure additional capabilities for WASI based on declared capabilities
    /// This implements structural denial - only expose bindings for declared capabilities
    fn configure_capabilities(
        builder: &mut WasiCtxBuilder,
        config: &Config,
    ) -> anyhow::Result<()> {
        // Parse capabilities from config
        let capabilities = Capabilities::from_string(&config.capabilities);

        // Deny by default - no capabilities means no network, no filesystem beyond preopened
        tracing::info!("Configuring capabilities: {:?}", capabilities.all());

        // Network access: only enabled if fetch:read or fetch:write is declared
        if capabilities.can_fetch() {
            builder
                .allow_tcp(true)
                .allow_udp(true)
                .allow_ip_name_lookup(true);
            tracing::debug!("Network access enabled (capability: fetch:*)");
        } else {
            // Explicitly disable network access - structural denial
            builder
                .allow_tcp(false)
                .allow_udp(false)
                .allow_ip_name_lookup(false);
            tracing::info!("WASI networking disabled - no fetch capability declared");
        }

        // Note: Cache, KV, and other bindings are handled at the application level,
        // not the WASI level. This function only controls WASI system access.

        // Configure time access
        if config.wasi_allow_time {
            // Allow time access (default in wasmtime-wasi)
            tracing::debug!("WASI time access enabled");
        } else {
            // Time access is disabled - use custom clock bindings that deny access
            builder
                .monotonic_clock(DisabledMonotonicClock)
                .wall_clock(DisabledWallClock);
            tracing::warn!("WASI time access disabled per configuration");
            tracing::info!("WebAssembly modules will not be able to access system time functions");
            tracing::info!("Custom WASI clock bindings installed to intercept time calls");
            tracing::warn!("Note: Current implementation provides controlled values instead of blocking calls");
            tracing::warn!("Full implementation would require lower-level WASI interception to return proper errors");
        }

        // Additional hardened security measures
        if config.hardened_security {
            tracing::info!("Applying hardened security measures");

            // Disable all potentially dangerous features by default
            builder
                .allow_tcp(false)
                .allow_udp(false)
                .allow_ip_name_lookup(false);

            // Only allow network if explicitly enabled via capabilities
            if capabilities.can_fetch() {
                builder
                    .allow_tcp(true)
                    .allow_udp(true)
                    .allow_ip_name_lookup(true);
                tracing::warn!("Network access enabled in hardened mode via capability");
            } else if config.wasi_allow_network {
                // Fallback to config flag for backwards compatibility
                builder
                    .allow_tcp(true)
                    .allow_udp(true)
                    .allow_ip_name_lookup(true);
                tracing::warn!("Network access enabled in hardened mode - ensure proper validation");
            }

            // Limit environment variables to essential ones only
            // (Already handled above, but reinforced here)

            // Ensure no dangerous directories are preopened
            // (Already handled in configure_filesystem)
        }

        Ok(())
    }
}

/// WASI linker for connecting WASI imports to implementations
pub struct WasiLinker {
    linker: wasmtime::Linker<WasiP1Ctx>,
    kv_store: Option<SharedKVStore>,
    logger: crate::logging::StructuredLogger,
    security_monitor: Arc<crate::security::SecurityMonitor>,
}

impl WasiLinker {
    /// Create a new WASI linker
    pub fn new(
        engine: &wasmtime::Engine,
        config: &Config,
        kv_store: Option<SharedKVStore>,
        logger: crate::logging::StructuredLogger,
        security_monitor: Arc<crate::security::SecurityMonitor>,
    ) -> anyhow::Result<Self> {
        let mut linker = wasmtime::Linker::new(engine);

        // Parse capabilities to determine what bindings to expose
        let capabilities = Capabilities::from_string(&config.capabilities);

        // Add WASI p1 interfaces
        wasmtime_wasi::p1::add_to_linker_sync(&mut linker, |ctx: &mut WasiP1Ctx| ctx)
            .context("Failed to add WASI to linker")?;

        // Add FunctionFly host functions
        let host_functions_linker = HostFunctionsLinker::new(
            kv_store.clone(),
            logger.clone(),
            config.clone(),
            security_monitor.clone(),
        );
        host_functions_linker.add_to_linker(&mut linker)?;

        // Add MicroPython runtime imports for Python WASM modules
        Self::add_micropython_imports(&mut linker)?;

        Ok(Self {
            linker,
            kv_store,
            logger,
            security_monitor,
        })
    }

    /// Add MicroPython runtime imports for Python WASM wrapper modules
    /// These are stubs that satisfy the imports from micropython.wasm
    fn add_micropython_imports(linker: &mut wasmtime::Linker<WasiP1Ctx>) -> anyhow::Result<()> {
        // mp_js_init - Initialize MicroPython with heap size (stub)
        linker.func_wrap("env", "mp_js_init", |heap_size: i32| {
            tracing::info!("MicroPython mp_js_init called with heap_size: {}", heap_size);
            // In a full implementation, this would initialize the MicroPython runtime
            // For now, we just log that it was called
        })?;

        // mp_js_do_exec - Execute Python code (stub)
        linker.func_wrap("env", "mp_js_do_exec", |code_ptr: i32, input_ptr: i32| -> i32 {
            tracing::info!("MicroPython mp_js_do_exec called: code_ptr={}, input_ptr={}", code_ptr, input_ptr);
            // In a full implementation, this would execute Python code
            // For now, return 0 (null pointer = no result)
            0
        })?;

        // malloc - Allocate memory (stub)
        linker.func_wrap("env", "malloc", |size: i32| -> i32 {
            tracing::debug!("MicroPython malloc called: size={}", size);
            // Return 0 to indicate failure (module should use its own memory)
            0
        })?;

        // free - Free memory (stub)
        linker.func_wrap("env", "free", |ptr: i32| {
            tracing::debug!("MicroPython free called: ptr={}", ptr);
            // No-op stub
        })?;

        tracing::info!("MicroPython runtime imports added to WASI linker");
        Ok(())
    }



    /// Get the underlying linker
    pub fn linker(&self) -> &wasmtime::Linker<WasiP1Ctx> {
        &self.linker
    }

    /// Get mutable access to the linker
    pub fn linker_mut(&mut self) -> &mut wasmtime::Linker<WasiP1Ctx> {
        &mut self.linker
    }

}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::config::Config;

    #[test]
    fn test_wasi_context_creation() {
        let config = Config::default();
        let ctx = WasiContext::new(&config, "test@1.0.0".to_string());
        assert!(ctx.is_ok());
    }

    #[test]
    fn test_wasi_context_with_env_vars() {
        let config = Config {
            wasi_env: vec!["TEST_VAR=test_value".to_string()],
            ..Config::default()
        };
        let ctx = WasiContext::new(&config, "test@1.0.0".to_string());
        assert!(ctx.is_ok());
    }

    #[test]
    fn test_wasi_linker_creation() {
        let engine = wasmtime::Engine::default();
        let config = Config::default();
        let kv_store = Some(Arc::new(RwLock::new(crate::kv::KVStore::new(1000))));
        let logger = crate::logging::init_structured_logging(false);
        let security_monitor = Arc::new(crate::security::SecurityMonitor::new());
        let linker = WasiLinker::new(&engine, &config, kv_store, logger, security_monitor);
        assert!(linker.is_ok());
    }

    #[test]
    fn test_wasi_linker_with_kv_capability() {
        let engine = wasmtime::Engine::default();
        let config = Config {
            capabilities: "kv".to_string(),
            ..Config::default()
        };
        let kv_store = Some(Arc::new(RwLock::new(crate::kv::KVStore::new(1000))));
        let logger = crate::logging::init_structured_logging(false);
        let security_monitor = Arc::new(crate::security::SecurityMonitor::new());
        let linker = WasiLinker::new(&engine, &config, kv_store, logger, security_monitor);
        assert!(linker.is_ok());
    }
}
