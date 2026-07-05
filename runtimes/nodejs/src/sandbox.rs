//! Sandbox Isolation Primitives
//!
//! This module provides secure isolation for JavaScript execution,
//! including module restrictions, security validation, and resource limits.
//!
//! # Security Model
//!
//! - **Pre-execution validation**: All code is validated before execution
//! - **Static analysis**: Dangerous patterns are detected via regex and string analysis
//! - **Module restrictions**: Blocked/allowed module lists enforced
//! - **Resource limits**: Memory, CPU time, concurrent executions tracked
//! - **Audit logging**: Security-relevant events are logged for compliance
//!
//! # Architecture
//!
//! The sandbox performs pre-execution security checks and resource management.
//! Actual JavaScript execution is performed by the executor (QuickJS via JsContext).
//! This separation ensures security validation happens before any code runs.

use std::sync::atomic::{AtomicU64, AtomicU32, Ordering};
use std::sync::Arc;
use std::collections::HashSet;
use std::time::{Instant, Duration};
use std::panic::{self, PanicInfo};
use std::sync::atomic::AtomicBool;

use parking_lot::RwLock;
use tracing::{info, warn, error, debug, error_span, Instrument};
use regex::Regex;

use crate::{RuntimeError, RuntimeVersion};

/// Maximum concurrent executions allowed per sandbox
const MAX_CONCURRENT_EXECUTIONS: u32 = 100;

/// Maximum code size in bytes (1MB)
/// Maximum size in bytes for a single JS code submission. Hard upper bound
/// enforced by `SandboxConfig::validate`. Public so callers (e.g. `NodeExecutor`)
/// can clamp their own limits to avoid `validate` rejections.
pub const MAX_CODE_SIZE_BYTES: usize = 1_048_576;

/// Maximum input size in bytes (10MB)
const MAX_INPUT_SIZE_BYTES: usize = 10_485_760;

/// Sandbox creation error details
#[derive(Debug, Clone)]
pub struct SandboxCreationError {
    pub reason: String,
    pub config_value: String,
}

impl std::fmt::Display for SandboxCreationError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "Sandbox creation failed: {} (value: {})", self.reason, self.config_value)
    }
}

impl std::error::Error for SandboxCreationError {}

/// Configuration for the sandbox
#[derive(Debug, Clone)]
pub struct SandboxConfig {
    /// Runtime version to use
    pub runtime_version: RuntimeVersion,

    /// Maximum memory in MB
    pub max_memory_mb: u32,

    /// Maximum concurrent executions
    pub max_concurrent_executions: u32,

    /// List of allowed modules (empty = use blocked list)
    pub allowed_modules: Vec<String>,

    /// List of blocked modules (hard-coded restrictions)
    pub blocked_modules: Vec<String>,

    /// Whether network access is enabled
    pub network_enabled: bool,

    /// Environment variables to expose
    pub env_vars: std::collections::HashMap<String, String>,

    /// Maximum code size in bytes
    pub max_code_size_bytes: usize,

    /// Whether to enable strict security mode
    pub strict_mode: bool,
}

impl Default for SandboxConfig {
    fn default() -> Self {
        Self {
            runtime_version: RuntimeVersion::Node20,
            max_memory_mb: 128,
            max_concurrent_executions: MAX_CONCURRENT_EXECUTIONS,
            allowed_modules: vec![],
            blocked_modules: Self::default_blocked_modules(),
            network_enabled: false,
            env_vars: std::collections::HashMap::new(),
            max_code_size_bytes: MAX_CODE_SIZE_BYTES,
            strict_mode: true,
        }
    }
}

impl SandboxConfig {
    /// Default blocked modules for security
    fn default_blocked_modules() -> Vec<String> {
        vec![
            "child_process".to_string(),
            "fs".to_string(),
            "net".to_string(),
            "tls".to_string(),
            "http".to_string(),
            "https".to_string(),
            "dns".to_string(),
            "dgram".to_string(),
            "repl".to_string(),
            "vm".to_string(),
            "worker_threads".to_string(),
            "perf_hooks".to_string(),
            "inspector".to_string(),
            "crypto".to_string(),
            "stream".to_string(),
            "zlib".to_string(),
            "readline".to_string(),
            "cluster".to_string(),
            "module".to_string(),
            "sys".to_string(),
            "pty".to_string(),
            "tty".to_string(),
            "os".to_string(),
            "ppid".to_string(),
            "process".to_string(),
            "assert".to_string(),
        ]
    }

    /// Validate the configuration
    pub fn validate(&self) -> Result<(), SandboxCreationError> {
        if self.max_memory_mb == 0 {
            return Err(SandboxCreationError {
                reason: "Memory limit must be greater than 0".to_string(),
                config_value: format!("{}", self.max_memory_mb),
            });
        }

        if self.max_memory_mb > 2048 {
            return Err(SandboxCreationError {
                reason: "Memory limit cannot exceed 2048MB".to_string(),
                config_value: format!("{}", self.max_memory_mb),
            });
        }

        if self.max_concurrent_executions == 0 {
            return Err(SandboxCreationError {
                reason: "Max concurrent executions must be greater than 0".to_string(),
                config_value: format!("{}", self.max_concurrent_executions),
            });
        }

        if self.max_concurrent_executions > MAX_CONCURRENT_EXECUTIONS {
            return Err(SandboxCreationError {
                reason: format!("Max concurrent executions cannot exceed {}", MAX_CONCURRENT_EXECUTIONS),
                config_value: format!("{}", self.max_concurrent_executions),
            });
        }

        if self.max_code_size_bytes > MAX_CODE_SIZE_BYTES {
            return Err(SandboxCreationError {
                reason: format!("Max code size cannot exceed {} bytes", MAX_CODE_SIZE_BYTES),
                config_value: format!("{}", self.max_code_size_bytes),
            });
        }

        // In strict mode, network must be explicitly enabled
        if self.strict_mode && self.network_enabled {
            debug!("Strict mode enabled with network access allowed");
        }

        Ok(())
    }

    /// Create a production configuration
    pub fn production() -> Self {
        Self {
            runtime_version: RuntimeVersion::Node20,
            max_memory_mb: 128,
            max_concurrent_executions: 50,
            allowed_modules: vec![],
            blocked_modules: Self::default_blocked_modules(),
            network_enabled: false,
            env_vars: std::collections::HashMap::new(),
            max_code_size_bytes: MAX_CODE_SIZE_BYTES,
            strict_mode: true,
        }
    }

    /// Create a development configuration
    pub fn development() -> Self {
        Self {
            runtime_version: RuntimeVersion::Node20,
            max_memory_mb: 256,
            max_concurrent_executions: 100,
            allowed_modules: vec![],
            blocked_modules: Self::default_blocked_modules(),
            network_enabled: true,
            env_vars: std::collections::HashMap::from([
                ("NODE_ENV".to_string(), "development".to_string()),
            ]),
            max_code_size_bytes: MAX_CODE_SIZE_BYTES,
            strict_mode: false,
        }
    }
}

/// Sandbox for isolated execution
///
/// Provides pre-execution security validation and resource management.
/// Actual code execution is delegated to the executor (QuickJS).
pub struct Sandbox {
    config: Arc<SandboxConfig>,
    execution_count: AtomicU64,
    active_executions: AtomicU32,
    total_rejections: AtomicU64,
    successful_executions: AtomicU64,
    failed_executions: AtomicU64,
    last_reset: RwLock<Instant>,
    created_at: Instant,

    // Security: compiled regex patterns (lazily initialized)
    dangerous_patterns: RwLock<Option<Vec<Regex>>>,

    // Security: panic recovery flag
    panic_occurred: AtomicBool,
}

impl Sandbox {
    /// Create a new sandbox with the given configuration
    pub fn new(config: SandboxConfig) -> Result<Self, RuntimeError> {
        // Validate config
        config.validate().map_err(|e| {
            RuntimeError::InvalidInput(e.to_string())
        })?;

        info!(
            "Creating sandbox with runtime: {:?}, memory_limit: {}MB, max_concurrent: {}, strict_mode: {}",
            config.runtime_version,
            config.max_memory_mb,
            config.max_concurrent_executions,
            config.strict_mode
        );

        Ok(Self {
            config: Arc::new(config),
            execution_count: AtomicU64::new(0),
            active_executions: AtomicU32::new(0),
            total_rejections: AtomicU64::new(0),
            successful_executions: AtomicU64::new(0),
            failed_executions: AtomicU64::new(0),
            last_reset: RwLock::new(Instant::now()),
            created_at: Instant::now(),
            dangerous_patterns: RwLock::new(None), // Lazily compiled
            panic_occurred: AtomicBool::new(false),
        })
    }

    /// Initialize the compiled regex patterns
    fn init_dangerous_patterns(&self) {
        let mut guard = self.dangerous_patterns.write();
        if guard.is_some() {
            return;
        }

        // Compile all dangerous patterns into regex for efficiency
        let patterns = vec![
            // Code execution
            r"eval\s*\(",
            r"Function\s*\(",
            r"new\s+Function\s*\(",
            r"eval\s*`",
            r"\beval\b",

            // Shell/command patterns
            r"child_process",
            r"exec\s*\(",
            r"execSync\s*\(",
            r"spawn\s*\(",
            r"spawnSync\s*\(",
            r"fork\s*\(",

            // File system access
            r"fs\s*\.\s*(readFile|writeFile|readFileSync|writeFileSync|unlink|rm|mkdir|rename|copyFile|opendir|readdir|stat|lstat|access|chmod|chown|utimes|realpath|link|symlink|mkdtemp|writeFileSync|readlink)",
            r#"require\s*\(\s*['\"]fs['\"]"#,
            r#"import\s*\(\s*['\"]fs['\"]"#,

            // Process access
            r"__dirname",
            r"__filename",
            r"process\s*\.\s*(cwd|chdir|exit|kill|pid|ppid|title|argv|execArgv|execPath|stdin|stdout|stderr) ",
            r"process\.(exit|kill|cwd|chdir)\s*\(",
            // NOTE: `process.env.*` is intentionally NOT in the dangerous
            // patterns list. Environment access is gated separately by
            // `validate_env_access` which checks each lookup against the
            // sensitive-name blocklist. Blocking the access itself would
            // also block legitimate reads like `process.env.NODE_ENV`.
            r"process\.stdin",
            r"process\.stdout",
            r"process\.stderr",

            // Global scope access
            r"global\s*\.",
            r"globalThis\s*\.\s*(require|process|eval)",
            r"\bGLOBAL\b",

            // Module internals
            r#"require\s*\(\s*['\"]module['\"]"#,
            r#"require\s*\(\s*['\"]child_process['\"]"#,
            r#"require\s*\(\s*['\"]cluster['\"]"#,
            r#"require\s*\(\s*['\"]os['\"]"#,
            r#"require\s*\(\s*['\"]pty['\"]"#,
            r#"require\s*\(\s*['\"]tty['\"]"#,
            r#"import\s*\(\s*['\"]module['\"]"#,
            r#"import\s*\(\s*['\"]child_process['\"]"#,

            // Debugger/profiler access
            r"debugger\b",
            r"--inspect",
            r"--prof",
            r"v8\s*\.",
            r"InspectorSession",
            r"runNNN\b",

            // Native addon access
            r"\.dll\b",
            r"\.so\b",
            r"\.dylib\b",
            r#"require\s*\(\s*['\"](?:native|addon)['\"]"#,

            // Network access when disabled
            r"http\s*\.\s*(get|request|Agent|Server)",
            r"https\s*\.\s*(get|request|Agent|Server)",
            r"net\s*\.\s*(connect|createConnection|Server|Socket|createServer)",
            r"tls\s*\.\s*(connect|createServer|Certificate|credentials)",
            r"dgram\s*\.\s*(createSocket|createServer)",

            // Dynamic code generation
            r"WebAssembly",
            r"\bReflect\.ownKeys\b.*\bconstructor\b",
            r"Object\.getPrototypeOf\s*\(\s*[^)]*\)\s*\.constructor",

            // Template literal injection
            r"\$\{.*eval",
            r"new\s+Promise\s*\(\s*async\s+\(",
        ];

        let compiled: Result<Vec<Regex>, _> = patterns
            .iter()
            .map(|p| Regex::new(p))
            .collect();

        match compiled {
            Ok(regexes) => {
                debug!("Compiled {} dangerous pattern regexes", regexes.len());
                *guard = Some(regexes);
            }
            Err(e) => {
                error!("Failed to compile dangerous pattern regexes: {}", e);
                // Fall back to empty - patterns will be checked via string contains
                *guard = Some(vec![]);
            }
        }
    }

    /// Validate code for security issues before execution
    ///
    /// This performs comprehensive static analysis including:
    /// - Dangerous pattern detection (code execution, shell access, etc.)
    /// - Module restriction validation
    /// - Code size limits
    /// - Network access verification
    pub fn validate_code(&self, code: &str) -> Result<(), RuntimeError> {
        // Initialize patterns if not already done
        self.init_dangerous_patterns();

        // 1. Check code size
        let code_len = code.len();
        if code_len > self.config.max_code_size_bytes {
            self.total_rejections.fetch_add(1, Ordering::Relaxed);
            warn!(
                "Code rejected: size {} exceeds limit {}",
                code_len,
                self.config.max_code_size_bytes
            );
            return Err(RuntimeError::SecurityViolation(format!(
                "Code size {} bytes exceeds maximum allowed {} bytes",
                code_len, self.config.max_code_size_bytes
            )));
        }

        // 2. Check for empty code
        let trimmed = code.trim();
        if trimmed.is_empty() {
            self.total_rejections.fetch_add(1, Ordering::Relaxed);
            return Err(RuntimeError::SecurityViolation(
                "Code cannot be empty".to_string()
            ));
        }

        // 3. Check for binary/non-text characters that could indicate obfuscation
        if self.contains_suspicious_bytes(code) {
            self.total_rejections.fetch_add(1, Ordering::Relaxed);
            warn!("Code rejected: contains suspicious binary patterns");
            return Err(RuntimeError::SecurityViolation(
                "Code contains suspicious binary or obfuscated content".to_string()
            ));
        }

        // 4. Check dangerous patterns via regex
        if let Some(ref patterns) = *self.dangerous_patterns.read() {
            for regex in patterns {
                if regex.is_match(code) {
                    self.total_rejections.fetch_add(1, Ordering::Relaxed);
                    let pattern = regex.as_str();
                    warn!("Blocked dangerous pattern in code: {}", pattern);
                    return Err(RuntimeError::SecurityViolation(format!(
                        "Code contains disallowed pattern: {}", pattern
                    )));
                }
            }
        }

        // 5. Check blocked modules in import/require statements
        self.validate_module_imports(code)?;

        // 6. Check for network access if disabled
        if !self.config.network_enabled {
            self.validate_network_access(code)?;
        }

        // 7. Check environment variable access
        self.validate_env_access(code)?;

        // 8. Validate env vars don't leak sensitive data
        self.validate_env_vars_not_sensitive()?;

        debug!("Code validation passed ({} bytes)", code_len);
        Ok(())
    }

    /// Check for suspicious binary patterns that might indicate obfuscation
    fn contains_suspicious_bytes(&self, code: &str) -> bool {
        // Null bytes in the middle of code
        if code.as_bytes().iter().filter(|&&b| b == 0).count() > 0 {
            return true;
        }

        // Very long lines that might indicate obfuscation (over 10KB)
        for line in code.lines() {
            if line.len() > 10_000 {
                debug!("Suspiciously long line detected: {} chars", line.len());
                return true;
            }
        }

        // High ratio of non-printable characters
        let total = code.chars().count();
        if total > 0 {
            let non_printable = code.chars().filter(|c| !c.is_ascii_graphic() && !c.is_whitespace()).count();
            let ratio = non_printable as f64 / total as f64;
            if ratio > 0.1 {
                debug!("High ratio of non-printable chars: {:.2}%", ratio * 100.0);
                return true;
            }
        }

        false
    }

    /// Validate module imports/requires
    fn validate_module_imports(&self, code: &str) -> Result<(), RuntimeError> {
        // Patterns for detecting module imports
        let import_patterns = [
            (r#"require\s*\(\s*['"]([^'"]+)['"]\s*\)"#, "require"),
            (r#"import\s+\w+\s+from\s+['"]([^'"]+)['"]\s*"#, "import default"),
            (r#"import\s*\(\s*['"]([^'"]+)['"]\s*\)"#, "dynamic import"),
            (r#"import\s+['"]([^'"]+)['"]"#, "import bare"),
        ];

        for (pattern, import_type) in import_patterns {
            let re = match Regex::new(pattern) {
                Ok(r) => r,
                Err(_) => continue,
            };

            for cap in re.captures_iter(code) {
                if let Some(module_name) = cap.get(1) {
                    let module = module_name.as_str();

                    // Check if it's a scoped package
                    let normalized = if module.starts_with('@') {
                        module.split('/').take(2).collect::<Vec<_>>().join("/")
                    } else {
                        module.split('/').next().unwrap_or(module).to_string()
                    };

                    // Check if allowed (if allowlist is not empty)
                    if !self.config.allowed_modules.is_empty() {
                        if !self.config.allowed_modules.iter().any(|m| {
                            m == &normalized || normalized.starts_with(m)
                        }) {
                            self.total_rejections.fetch_add(1, Ordering::Relaxed);
                            warn!("Module '{}' not in allowed list (requested via {})", normalized, import_type);
                            return Err(RuntimeError::SecurityViolation(format!(
                                "Module '{}' is not allowed in this runtime", normalized
                            )));
                        }
                    } else {
                        // Use blocklist
                        if self.config.blocked_modules.contains(&normalized) {
                            self.total_rejections.fetch_add(1, Ordering::Relaxed);
                            warn!("Blocked module '{}' requested via {}", normalized, import_type);
                            return Err(RuntimeError::SecurityViolation(format!(
                                "Module '{}' is not allowed in this runtime", normalized
                            )));
                        }
                    }
                }
            }
        }

        Ok(())
    }

    /// Validate network access patterns if network is disabled
    fn validate_network_access(&self, code: &str) -> Result<(), RuntimeError> {
        let network_patterns = [
            r"fetch\s*\(",
            r"XMLHttpRequest",
            r"ActiveXObject",
            r"WebSocket\s*\(",
            r"EventSource\s*\(",
            r#"require\s*\(\s*['\"]http['\"]"#,
            r#"require\s*\(\s*['\"]https['\"]"#,
            r#"require\s*\(\s*['\"]net['\"]"#,
            r#"require\s*\(\s*['\"]tls['\"]"#,
            r#"require\s*\(\s*['\"]dgram['\"]"#,
            r#"import\s*\(\s*['\"]http['\"]"#,
            r#"import\s*\(\s*['\"]https['\"]"#,
        ];

        for pattern in network_patterns {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(code) {
                    self.total_rejections.fetch_add(1, Ordering::Relaxed);
                    warn!("Network access pattern detected when network is disabled: {}", pattern);
                    return Err(RuntimeError::SecurityViolation(format!(
                        "Network access is not allowed in this runtime"
                    )));
                }
            }
        }

        Ok(())
    }

    /// Validate environment variable access
    fn validate_env_access(&self, code: &str) -> Result<(), RuntimeError> {
        // Patterns that access process.env
        let env_patterns = [
            r"process\.env\s*\.",
            r"process\.env\s*\[",
            r"host_getenv\s*\(",
        ];

        let mut accesses_env = false;
        for pattern in env_patterns {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(code) {
                    accesses_env = true;
                    break;
                }
            }
        }

        if !accesses_env {
            return Ok(());
        }

        // Check if trying to access sensitive env vars. We accept BOTH
        // dot-access (`process.env.API_KEY`) and bracket-access
        // (`process.env['API_KEY']` / `process.env["API_KEY"]`).
        let sensitive_patterns = [
            "PASSWORD", "SECRET", "TOKEN", "API_KEY", "PRIVATE",
            "CREDENTIAL", "AUTH", "CERT", "KEY", "DATABASE_URL",
            "CONNECTION_STRING", "AWS_", "AZURE_", "GCP_", "STRIPE",
        ];

        for sensitive in sensitive_patterns {
            // Dot access: process.env.API_KEY
            let dot_check = format!(
                r#"process\.env\s*\.\s*['\"]?{}{{0,40}}['\"]?"#,
                regex::escape(sensitive)
            );
            // Bracket access: process.env['API_KEY'] or process.env["API_KEY"]
            let bracket_check = format!(
                r#"process\.env\s*\[\s*['\"]{}\b"#,
                regex::escape(sensitive)
            );
            for check in [&dot_check, &bracket_check] {
                if let Ok(check_re) = Regex::new(check) {
                    if check_re.is_match(code) {
                        self.total_rejections.fetch_add(1, Ordering::Relaxed);
                        warn!("Access to sensitive environment variable pattern detected: {}", sensitive);
                        return Err(RuntimeError::SecurityViolation(format!(
                            "Access to sensitive environment variables is not allowed"
                        )));
                    }
                }
            }
        }

        Ok(())
    }

    /// Validate that configured env vars don't contain sensitive data
    fn validate_env_vars_not_sensitive(&self) -> Result<(), RuntimeError> {
        let sensitive_patterns = [
            "PASSWORD", "SECRET", "TOKEN", "API_KEY", "PRIVATE",
            "CREDENTIAL", "AUTH", "CERT", "KEY", "DATABASE_URL",
            "CONNECTION_STRING", "STRIPE", "PRIVATE_KEY", "ACCESS_TOKEN",
        ];

        for (key, value) in &self.config.env_vars {
            for sensitive in &sensitive_patterns {
                if key.to_uppercase().contains(sensitive) && !value.is_empty() {
                    warn!("Sensitive-looking env var '{}' configured - this is not allowed", key);
                    return Err(RuntimeError::SecurityViolation(format!(
                        "Environment variable '{}' appears to contain sensitive data", key
                    )));
                }
            }
        }

        Ok(())
    }

    /// Validate input data size
    pub fn validate_input(&self, input: &serde_json::Value) -> Result<(), RuntimeError> {
        let input_str = serde_json::to_string(input)
            .map_err(|e| RuntimeError::InvalidInput(format!("Failed to serialize input: {}", e)))?;

        if input_str.len() > MAX_INPUT_SIZE_BYTES {
            self.total_rejections.fetch_add(1, Ordering::Relaxed);
            return Err(RuntimeError::SecurityViolation(format!(
                "Input size {} exceeds maximum allowed {} bytes",
                input_str.len(),
                MAX_INPUT_SIZE_BYTES
            )));
        }

        Ok(())
    }

    /// Execute code in the sandbox.
    ///
    /// Validates code and input, enforces concurrent execution limits,
    /// then evaluates the code using the embedded QuickJS engine.
    ///
    /// NOTE: The primary execution path in production is via the executor
    /// (executor.rs) which calls `validate_code()` and `validate_input()`
    /// separately and manages its own QuickJS context. This method is
    /// provided for standalone sandbox usage (e.g. CLI mode, testing).
    pub fn execute(
        &self,
        code: &str,
        input: &serde_json::Value,
    ) -> Result<serde_json::Value, RuntimeError> {
        self.validate_code(code)?;
        self.validate_input(input)?;

        let prev_count = self.execution_count.fetch_add(1, Ordering::Relaxed);
        let prev_active = self.active_executions.fetch_add(1, Ordering::Relaxed);

        if prev_active >= self.config.max_concurrent_executions {
            self.active_executions.fetch_sub(1, Ordering::Relaxed);
            self.total_rejections.fetch_add(1, Ordering::Relaxed);
            return Err(RuntimeError::SecurityViolation(
                "Too many concurrent executions".to_string()
            ));
        }

        debug!(
            "Sandbox execution {} started (active: {}, total: {})",
            prev_count + 1,
            prev_active + 1,
            prev_count + 1
        );

        // Wrap code in an async IIFE that receives input as a global variable,
        // then evaluate using the crate's JsContext (rquickjs wrapper).
        let input_json = serde_json::to_string(input)
            .map_err(|e| RuntimeError::InvalidInput(format!("Failed to serialize input: {}", e)))?;

        let wrapped_code = format!(
            "(async () => {{ const __input = {}; {} }})()",
            input_json, code
        );

        let result = crate::executor::JsContext::new(false, &std::collections::HashMap::new())
            .map_err(|e| RuntimeError::Execution(format!("Failed to create JS context: {}", e)))
            .and_then(|mut ctx| {
                ctx.load_module(&wrapped_code)
                    .map_err(|e| RuntimeError::Execution(format!("JS load error: {}", e)))?;
                ctx.call_handler(&input_json)
                    .map_err(|e| RuntimeError::Execution(format!("JS execution error: {}", e)))
            })
            .and_then(|result_json| {
                serde_json::from_str(&result_json)
                    .map_err(|e| RuntimeError::Execution(format!("Invalid result JSON: {}", e)))
            });

        self.active_executions.fetch_sub(1, Ordering::Relaxed);

        match result {
            Ok(value) => {
                self.successful_executions.fetch_add(1, Ordering::Relaxed);
                Ok(value)
            }
            Err(e) => {
                self.failed_executions.fetch_add(1, Ordering::Relaxed);
                Err(e)
            }
        }
    }

    /// Check if a module is allowed
    pub fn is_module_allowed(&self, module: &str) -> bool {
        if !self.config.allowed_modules.is_empty() {
            return self.config.allowed_modules.iter()
                .any(|m| m == module || module.starts_with(m));
        }

        !self.config.blocked_modules.contains(&module.to_string())
    }

    /// Check if network access is allowed
    pub fn is_network_allowed(&self) -> bool {
        self.config.network_enabled
    }

    /// Check if a specific network pattern is used
    pub fn is_network_pattern_blocked(&self, pattern: &str) -> bool {
        !self.config.network_enabled && [
            "fetch", "XMLHttpRequest", "WebSocket", "EventSource",
            "http.", "https.", "net.", "tls.", "dgram."
        ].iter().any(|p| pattern.contains(p))
    }

    /// Get sandbox statistics
    pub fn stats(&self) -> SandboxStats {
        SandboxStats {
            total_executions: self.execution_count.load(Ordering::Relaxed),
            active_executions: self.active_executions.load(Ordering::Relaxed),
            total_rejections: self.total_rejections.load(Ordering::Relaxed),
            uptime_seconds: self.created_at.elapsed().as_secs(),
            memory_limit_mb: self.config.max_memory_mb,
            max_concurrent_executions: self.config.max_concurrent_executions,
            runtime_version: self.config.runtime_version.clone(),
            strict_mode: self.config.strict_mode,
            network_enabled: self.config.network_enabled,
        }
    }

    /// Health check for the sandbox
    pub async fn health_check(&self) -> bool {
        // Check for panic
        if self.panic_occurred.load(Ordering::Relaxed) {
            warn!("Sandbox panic flag is set");
            return false;
        }

        // Check concurrent executions
        let active = self.active_executions.load(Ordering::Relaxed);
        if active > self.config.max_concurrent_executions {
            warn!("Too many active executions: {} (limit: {})", active, self.config.max_concurrent_executions);
            return false;
        }

        // Check rejection rate (if we've processed many requests)
        let total = self.execution_count.load(Ordering::Relaxed);
        let rejections = self.total_rejections.load(Ordering::Relaxed);
        if total > 100 && rejections as f64 / total as f64 > 0.5 {
            warn!("High rejection rate: {} / {} ({:.1}%)", rejections, total, (rejections as f64 / total as f64) * 100.0);
            // Still return true - high rejection rate is a defense mechanism, not a failure
        }

        true
    }

    /// Reset the sandbox (clears counters, etc.)
    pub fn reset(&self) {
        self.execution_count.store(0, Ordering::Relaxed);
        self.active_executions.store(0, Ordering::Relaxed);
        *self.last_reset.write() = Instant::now();
        self.panic_occurred.store(false, Ordering::Relaxed);
        info!("Sandbox reset");
    }

    /// Record a panic for health monitoring
    pub fn record_panic(&self, panic_info: &PanicInfo) {
        self.panic_occurred.store(true, Ordering::Relaxed);
        error!("Sandbox panic recorded: {:?}", panic_info);
    }

    /// Get sandbox configuration
    pub fn config(&self) -> &SandboxConfig {
        &self.config
    }

    /// Acquire an execution slot, returns guard on success
    pub fn acquire_execution_slot(&self) -> Result<SandboxExecutionGuard, RuntimeError> {
        let current = self.active_executions.fetch_add(1, Ordering::Relaxed);
        if current >= self.config.max_concurrent_executions {
            self.active_executions.fetch_sub(1, Ordering::Relaxed);
            return Err(RuntimeError::SecurityViolation(
                "Too many concurrent executions".to_string()
            ));
        }

        self.execution_count.fetch_add(1, Ordering::Relaxed);
        Ok(SandboxExecutionGuard { sandbox: self })
    }

    /// Release an execution slot
    fn release_execution_slot(&self) {
        self.active_executions.fetch_sub(1, Ordering::Relaxed);
    }
}

/// Guard that automatically releases execution slot when dropped
pub struct SandboxExecutionGuard<'a> {
    sandbox: &'a Sandbox,
}

impl<'a> Drop for SandboxExecutionGuard<'a> {
    fn drop(&mut self) {
        self.sandbox.release_execution_slot();
        debug!("Execution slot released");
    }
}

/// Statistics about the sandbox
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct SandboxStats {
    pub total_executions: u64,
    pub active_executions: u32,
    pub total_rejections: u64,
    pub uptime_seconds: u64,
    pub memory_limit_mb: u32,
    pub max_concurrent_executions: u32,
    pub runtime_version: RuntimeVersion,
    pub strict_mode: bool,
    pub network_enabled: bool,
}

impl SandboxStats {
    /// Get rejection rate as percentage
    pub fn rejection_rate(&self) -> f64 {
        let total = self.total_executions + self.total_rejections;
        if total == 0 {
            return 0.0;
        }
        (self.total_rejections as f64 / total as f64) * 100.0
    }

    /// Get active capacity percentage
    pub fn active_capacity_pct(&self) -> f64 {
        (self.active_executions as f64 / self.max_concurrent_executions as f64) * 100.0
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn create_test_sandbox() -> Sandbox {
        let config = SandboxConfig::default();
        Sandbox::new(config).unwrap()
    }

    #[test]
    fn test_sandbox_creation() {
        let config = SandboxConfig::default();
        let sandbox = Sandbox::new(config);
        assert!(sandbox.is_ok());
    }

    #[test]
    fn test_sandbox_creation_production() {
        let config = SandboxConfig::production();
        let sandbox = Sandbox::new(config);
        assert!(sandbox.is_ok());
    }

    #[test]
    fn test_module_blocking() {
        let sandbox = create_test_sandbox();

        assert!(!sandbox.is_module_allowed("child_process"));
        assert!(!sandbox.is_module_allowed("fs"));
        assert!(!sandbox.is_module_allowed("os"));
        assert!(sandbox.is_module_allowed("json"));
        assert!(sandbox.is_module_allowed("console"));
    }

    #[test]
    fn test_dangerous_code_blocking() {
        let sandbox = create_test_sandbox();

        // eval
        let result = sandbox.validate_code("eval('console.log(1)')");
        assert!(result.is_err());

        // Function constructor
        let result = sandbox.validate_code("Function('console.log(1)')()");
        assert!(result.is_err());

        // process.exit
        let result = sandbox.validate_code("process.exit(0)");
        assert!(result.is_err());

        // child_process
        let result = sandbox.validate_code("require('child_process')");
        assert!(result.is_err());

        // fs access
        let result = sandbox.validate_code("require('fs').readFileSync('/etc/passwd')");
        assert!(result.is_err());

        // __dirname
        let result = sandbox.validate_code("console.log(__dirname)");
        assert!(result.is_err());

        // global
        let result = sandbox.validate_code("global.console.log('test')");
        assert!(result.is_err());

        // debugger
        let result = sandbox.validate_code("debugger;");
        assert!(result.is_err());
    }

    #[test]
    fn test_valid_code_passes() {
        let sandbox = create_test_sandbox();

        let valid_codes = vec![
            r#"export function handler(input) { return { result: input.value }; }"#,
            r#"export default function(input) { return "hello"; }"#,
            r#"const x = 1; const y = 2; export function handler(input) { return x + y; }"#,
            r#"import { json } from 'cirjson'; export function handler(input) { return json.parse(input); }"#,
            r#"export async function handler(input) { return await Promise.resolve(42); }"#,
        ];

        for code in valid_codes {
            let result = sandbox.validate_code(code);
            assert!(result.is_ok(), "Expected {:?} to be valid, got {:?}", code, result);
        }
    }

    #[test]
    fn test_code_size_limit() {
        let sandbox = create_test_sandbox();

        // Create code that exceeds the limit
        let big_code = "x".repeat(MAX_CODE_SIZE_BYTES + 1);
        let result = sandbox.validate_code(&big_code);
        assert!(result.is_err());

        if let Err(RuntimeError::SecurityViolation(msg)) = result {
            assert!(msg.contains("exceeds maximum"));
        } else {
            panic!("Expected SecurityViolation error");
        }
    }

    #[test]
    fn test_network_blocking_when_disabled() {
        let sandbox = create_test_sandbox();
        assert!(!sandbox.is_network_allowed());

        // fetch should be blocked when network is disabled
        let result = sandbox.validate_code("fetch('https://evil.com')");
        assert!(result.is_err());
    }

    #[test]
    fn test_network_allowed_in_config() {
        let config = SandboxConfig::development(); // network_enabled = true
        let sandbox = Sandbox::new(config).unwrap();
        assert!(sandbox.is_network_allowed());
    }

    #[test]
    fn test_sandbox_stats() {
        let sandbox = create_test_sandbox();

        let stats = sandbox.stats();
        assert_eq!(stats.total_executions, 0);
        assert_eq!(stats.active_executions, 0);
        assert_eq!(stats.total_rejections, 0);
    }

    #[test]
    fn test_concurrent_execution_limit() {
        let config = SandboxConfig {
            max_concurrent_executions: 2,
            ..Default::default()
        };
        let sandbox = Sandbox::new(config).unwrap();

        // Acquire two slots
        let guard1 = sandbox.acquire_execution_slot();
        assert!(guard1.is_ok());

        let guard2 = sandbox.acquire_execution_slot();
        assert!(guard2.is_ok());

        // Third should fail
        let guard3 = sandbox.acquire_execution_slot();
        assert!(guard3.is_err());

        drop(guard1);
        drop(guard2);

        // Now we should be able to acquire again
        let guard4 = sandbox.acquire_execution_slot();
        assert!(guard4.is_ok());
    }

    #[test]
    fn test_input_validation() {
        let sandbox = create_test_sandbox();

        // Valid input
        let valid_input = serde_json::json!({"key": "value"});
        assert!(sandbox.validate_input(&valid_input).is_ok());

        // Null input
        let null_input = serde_json::Value::Null;
        assert!(sandbox.validate_input(&null_input).is_ok());

        // Big input
        let big_input = serde_json::json!({"data": "x".repeat(MAX_INPUT_SIZE_BYTES + 1)});
        assert!(sandbox.validate_input(&big_input).is_err());
    }

    #[test]
    fn test_env_var_validation() {
        let sandbox = create_test_sandbox();

        // Valid env var access
        let result = sandbox.validate_code("const x = process.env.NODE_ENV;");
        assert!(result.is_ok());

        // Accessing sensitive env var
        let result = sandbox.validate_code("const key = process.env.API_KEY;");
        assert!(result.is_err());
    }

    #[test]
    fn test_config_validation() {
        // Invalid memory
        let bad_config = SandboxConfig {
            max_memory_mb: 0,
            ..Default::default()
        };
        assert!(bad_config.validate().is_err());

        // Invalid concurrent
        let bad_config = SandboxConfig {
            max_concurrent_executions: 0,
            ..Default::default()
        };
        assert!(bad_config.validate().is_err());

        // Valid
        let good_config = SandboxConfig::default();
        assert!(good_config.validate().is_ok());
    }

    #[test]
    fn test_production_config_has_network_disabled() {
        let config = SandboxConfig::production();
        assert!(!config.network_enabled);
        assert!(config.strict_mode);
    }

    #[test]
    fn test_development_config_has_network_enabled() {
        let config = SandboxConfig::development();
        assert!(config.network_enabled);
        assert!(!config.strict_mode);
    }

    #[test]
    fn test_sandbox_stats_methods() {
        let stats = SandboxStats {
            total_executions: 100,
            active_executions: 10,
            total_rejections: 50,
            uptime_seconds: 3600,
            memory_limit_mb: 128,
            max_concurrent_executions: 100,
            runtime_version: RuntimeVersion::Node20,
            strict_mode: true,
            network_enabled: false,
        };

        assert!((stats.rejection_rate() - 33.33).abs() < 0.1);
        assert!((stats.active_capacity_pct() - 10.0).abs() < 0.1);
    }

    #[test]
    fn test_obfuscation_detection() {
        let sandbox = create_test_sandbox();

        // Very long line
        let long_line = format!("const x = '{}';", "x".repeat(11_000));
        let result = sandbox.validate_code(&long_line);
        assert!(result.is_err());

        // Null bytes
        let with_null = "const x = 'hello\x00world';";
        let result = sandbox.validate_code(with_null);
        assert!(result.is_err());
    }
}
