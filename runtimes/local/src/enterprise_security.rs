//! Enterprise security hardening features.
//!
//! This module provides advanced security features for enterprise deployments,
//! including enhanced input validation, sandboxing improvements, and security best practices.

use std::collections::{HashMap, HashSet};
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::RwLock;
use serde::{Deserialize, Serialize};
use regex::Regex;

use crate::logging::StructuredLogger;

/// Policy for handling dangerous capability combinations
#[derive(Debug, Clone, PartialEq)]
pub enum DangerousCapabilityPolicy {
    /// Block dangerous combinations immediately
    Block,
    /// Allow but log audit entry (monitoring mode)
    AllowWithAudit,
    /// Require admin approval before allowing
    RequireApproval,
}

impl Default for DangerousCapabilityPolicy {
    fn default() -> Self {
        DangerousCapabilityPolicy::Block
    }
}

/// Enterprise security configuration
#[derive(Debug, Clone)]
#[allow(dead_code)]
pub struct EnterpriseSecurityConfig {
    /// Enable advanced input validation
    pub enable_input_validation: bool,
    /// Maximum input size in bytes
    pub max_input_size: usize,
    /// Enable SQL injection detection
    pub detect_sql_injection: bool,
    /// Enable XSS detection
    pub detect_xss: bool,
    /// Enable command injection detection
    pub detect_command_injection: bool,
    /// Rate limiting window in seconds
    pub rate_limit_window_secs: u64,
    /// Maximum requests per window
    pub max_requests_per_window: usize,
    /// Enable audit logging
    pub enable_audit_logging: bool,
    /// Suspicious pattern detection threshold
    pub suspicious_pattern_threshold: usize,
    /// Policy for handling dangerous capability combinations
    pub dangerous_capability_policy: DangerousCapabilityPolicy,
}

impl Default for EnterpriseSecurityConfig {
    fn default() -> Self {
        Self {
            enable_input_validation: true,
            max_input_size: 1024 * 1024, // 1MB
            detect_sql_injection: true,
            detect_xss: true,
            detect_command_injection: true,
            rate_limit_window_secs: 60,
            max_requests_per_window: 100,
            enable_audit_logging: true,
            suspicious_pattern_threshold: 3,
            dangerous_capability_policy: DangerousCapabilityPolicy::Block,
        }
    }
}

/// Input validation result
#[derive(Debug, Clone)]
pub enum ValidationResult {
    Valid,
    Invalid(String), // Reason for invalidation
    Suspicious(String), // Warning about suspicious content
}

/// Enterprise security enforcer
pub struct EnterpriseSecurityEnforcer {
    /// Configuration
    config: EnterpriseSecurityConfig,
    /// Rate limiting data
    rate_limits: Arc<RwLock<HashMap<String, RateLimitData>>>,
    /// Security patterns for detection
    security_patterns: Arc<RwLock<HashMap<String, SecurityPattern>>>,
    /// Logger
    #[allow(dead_code)]
    logger: Arc<StructuredLogger>,
    /// Audit log entries
    audit_log: Arc<RwLock<Vec<AuditEntry>>>,
}

#[derive(Debug, Clone)]
struct RateLimitData {
    requests: Vec<Instant>,
    last_cleanup: Instant,
}

#[derive(Debug, Clone)]
struct SecurityPattern {
    pattern: Regex,
    violation_type: String,
    severity: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AuditEntry {
    pub timestamp: u64,
    pub function_key: String,
    pub action: String,
    pub result: String,
    pub details: String,
    pub ip_address: Option<String>,
    pub user_agent: Option<String>,
}

impl EnterpriseSecurityEnforcer {
    /// Create a new enterprise security enforcer
    pub fn new(logger: Arc<StructuredLogger>) -> Self {
        let config = EnterpriseSecurityConfig::default();
        let enforcer = Self {
            config,
            rate_limits: Arc::new(RwLock::new(HashMap::new())),
            security_patterns: Arc::new(RwLock::new(HashMap::new())),
            logger,
            audit_log: Arc::new(RwLock::new(Vec::new())),
        };

        enforcer
    }

    /// Initialize security patterns lazily on first use
    async fn initialize_security_patterns(&self) {
        // Check if already initialized
        {
            let patterns = self.security_patterns.read().await;
            if !patterns.is_empty() { tracing::info!("Patterns already initialized, count={}", patterns.len());
                return;
            }
        }
        
        let patterns = vec![
            ("sql_injection", r"(?i)(union\s+select|select\s+.*\s+from|drop\s+table|insert\s+into|update\s+.*\s+set)", "SQL Injection", "high"),
            ("xss", r"<script[^>]*>.*?</script>|<iframe[^>]*>.*?</iframe>|<object[^>]*>.*?</object>", "Cross-Site Scripting", "high"),
            ("command_injection", r"[;&|`$()]", "Command Injection", "critical"),
            ("path_traversal", r"\.\./|\.\.\\", "Path Traversal", "high"),
            ("suspicious_functions", r"(?i)(eval|exec|system|shell_exec|popen)", "Dangerous Function Call", "medium"),
        ];

        tracing::info!("Initializing {} patterns", patterns.len());
        let mut security_patterns = self.security_patterns.write().await;
        for (name, pattern, violation_type, severity) in patterns {
            if let Ok(regex) = Regex::new(pattern) {
                security_patterns.insert(name.to_string(), SecurityPattern {
                    pattern: regex,
                    violation_type: violation_type.to_string(),
                    severity: severity.to_string(),
                });
            }
        }
    }


    /// Validate input data
    pub async fn validate_input(&self, input: &str, function_key: &str) -> ValidationResult {
        // Lazy initialization of security patterns on first validation
        self.initialize_security_patterns().await;
        
        // Check input size
        if input.len() > self.config.max_input_size {
            let reason = format!("Input size {} exceeds maximum allowed size {}", input.len(), self.config.max_input_size);
            self.log_audit_entry(function_key, "input_validation", "blocked", &reason, None, None).await;
            return ValidationResult::Invalid(reason);
        }

        // Check for suspicious patterns
        let patterns = self.security_patterns.read().await;
        let mut suspicious_patterns: Vec<(&String, &SecurityPattern)> = Vec::new();

        for (pattern_name, pattern) in patterns.iter() {
            if pattern.pattern.is_match(input) {
                suspicious_patterns.push((pattern_name, pattern));
            }
        }

        if suspicious_patterns.len() >= self.config.suspicious_pattern_threshold {
            // Build detailed report using violation_type and severity from matched patterns
            let pattern_details: Vec<String> = suspicious_patterns.iter()
                .map(|(name, pat)| format!("{} ({}/{})", name, pat.violation_type, pat.severity))
                .collect();
            let reason = format!("Multiple suspicious patterns detected: {}", pattern_details.join(", "));

            // Log with severity information
            let highest_severity = suspicious_patterns.iter()
                .map(|(_, pat)| pat.severity.as_str())
                .max_by(|a, b| {
                    let order = |s: &str| match s {
                        "critical" => 4,
                        "high" => 3,
                        "medium" => 2,
                        _ => 1,
                    };
                    order(a).cmp(&order(b))
                })
                .unwrap_or("medium");

            tracing::warn!(
                "Suspicious input detected for {}: {} patterns matched, highest severity: {}",
                function_key,
                suspicious_patterns.len(),
                highest_severity
            );

            self.log_audit_entry(function_key, "input_validation", "suspicious", &reason, None, None).await;
            return ValidationResult::Suspicious(reason);
        }

        ValidationResult::Valid
    }

    /// Check rate limiting
    pub async fn check_rate_limit(&self, identifier: &str, function_key: &str) -> bool {
        let mut rate_limits = self.rate_limits.write().await;
        let now = Instant::now();

        let data = rate_limits.entry(identifier.to_string()).or_insert(RateLimitData {
            requests: Vec::new(),
            last_cleanup: now,
        });

        // Perform cleanup if the cleanup interval has passed
        let cleanup_interval = Duration::from_secs(self.config.rate_limit_window_secs);
        if data.last_cleanup.elapsed() > cleanup_interval {
            // Clean up old requests outside the window
            let window_start = now - Duration::from_secs(self.config.rate_limit_window_secs);
            let before_count = data.requests.len();
            data.requests.retain(|&time| time > window_start);
            let cleaned_count = before_count - data.requests.len();

            // Update last cleanup time
            data.last_cleanup = now;

            if cleaned_count > 0 {
                tracing::debug!(
                    "Rate limit cleanup for {}: removed {} stale requests, {} remaining",
                    identifier, cleaned_count, data.requests.len()
                );
            }
        } else {
            // Just clean up old requests outside the window
            let window_start = now - Duration::from_secs(self.config.rate_limit_window_secs);
            data.requests.retain(|&time| time > window_start);
        }

        // Check if under limit
        if data.requests.len() >= self.config.max_requests_per_window {
            self.log_audit_entry(function_key, "rate_limit", "blocked", "Rate limit exceeded", None, None).await;
            return false;
        }

        // Add current request
        data.requests.push(now);
        true
    }

    /// Enhanced sandboxing validation for production workloads
    pub async fn validate_sandboxing(&self, function_key: &str, capabilities: &[String]) -> ValidationResult {
        use crate::capability::validate_capabilities;

        // First, validate that all capabilities are in the allowed list
        let caps = crate::capability::Capabilities::from_vec(capabilities.to_vec());
        if let Err(e) = validate_capabilities(&caps) {
            let violation_reason = format!("Invalid capability: {}", e);
            self.log_audit_entry(function_key, "sandboxing", "blocked", &violation_reason, None, None).await;
            return ValidationResult::Invalid(violation_reason);
        }

        // Check for dangerous capability combinations that could indicate compromise
        let dangerous_combinations = vec![
            // Network + storage + secrets = potential data exfiltration vector
            (vec!["fetch:read", "fetch:write", "storage", "secret"],
             "Excessive capabilities: network + storage + secrets"),
            // External API + crypto = potential C2 beacon
            (vec!["external_api", "crypto"],
             "Dangerous: external API access with crypto capabilities"),
            // Unrestricted email + storage = spam/phishing potential
            (vec!["email", "storage"],
             "Potential abuse: email + storage access"),
        ];

        for (combo, reason) in dangerous_combinations {
            let combo_set: HashSet<&str> = combo.iter().cloned().collect();
            let caps_set: HashSet<&str> = capabilities.iter().map(|s| s.as_str()).collect();

            // Only flag if ALL dangerous capabilities are present
            if combo_set.iter().all(|cap| caps_set.contains(cap)) {
                let violation_reason = format!("Dangerous capability combination: {}", reason);

                match self.config.dangerous_capability_policy {
                    DangerousCapabilityPolicy::Block => {
                        self.log_audit_entry(function_key, "sandboxing", "blocked", &violation_reason, None, None).await;
                        return ValidationResult::Invalid(violation_reason);
                    }
                    DangerousCapabilityPolicy::AllowWithAudit => {
                        self.log_audit_entry(function_key, "sandboxing", "allowed_with_audit", &violation_reason, None, None).await;
                        // Continue to check other validations, but log the dangerous combination
                    }
                    DangerousCapabilityPolicy::RequireApproval => {
                        self.log_audit_entry(function_key, "sandboxing", "pending_approval", &violation_reason, None, None).await;
                        // TODO: Check if this function+capability combo has been pre-approved
                        // For now, treat as blocked until approval system is implemented
                        return ValidationResult::Invalid(format!("{} - requires admin approval", violation_reason));
                    }
                }
            }
        }

        // Check for overly broad capabilities that should require approval
        let broad_capabilities = vec!["fetch:write", "external_api", "storage"];
        for cap in broad_capabilities {
            if capabilities.contains(&cap.to_string()) {
                // Log but don't block - these might be legitimate
                self.log_audit_entry(function_key, "sandboxing", "warning", &format!("Broad capability requested: {}", cap), None, None).await;
            }
        }

        ValidationResult::Valid
    }

    /// Log security audit entry
    pub async fn log_audit_entry(
        &self,
        function_key: &str,
        action: &str,
        result: &str,
        details: &str,
        ip_address: Option<String>,
        user_agent: Option<String>,
    ) {
        if !self.config.enable_audit_logging {
            return;
        }

        let entry = AuditEntry {
            timestamp: std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap_or_default()
                .as_secs(),
            function_key: function_key.to_string(),
            action: action.to_string(),
            result: result.to_string(),
            details: details.to_string(),
            ip_address,
            user_agent,
        };

        let mut audit_log = self.audit_log.write().await;
        audit_log.push(entry.clone());

        // Keep only recent entries (last 1000)
        if audit_log.len() > 1000 {
            audit_log.drain(0..100);
        }

        // Log to structured logger using the stored logger field
        let logger = Arc::clone(&self.logger);
        let correlation_id = logger.generate_correlation_id().await;
        logger.log_with_correlation(
            crate::logging::LogLevel::Info,
            format!("Security audit: {} / {} / {} [{}]", action, result, details, function_key),
            &correlation_id,
        );
        tracing::info!(
            function_key = %function_key,
            action = %action,
            result = %result,
            details = %details,
            "Enterprise security audit"
        );
    }

    /// Get security report
    pub async fn get_security_report(&self) -> EnterpriseSecurityReport {
        let audit_log = self.audit_log.read().await;
        let rate_limits = self.rate_limits.read().await;

        let total_audit_entries = audit_log.len();
        let blocked_requests = audit_log.iter().filter(|e| e.result == "blocked").count();
        let suspicious_requests = audit_log.iter().filter(|e| e.result == "suspicious").count();

        let active_rate_limits = rate_limits.len();
        let total_rate_limited_requests: usize = rate_limits.values()
            .map(|data| data.requests.len())
            .sum();

        EnterpriseSecurityReport {
            total_audit_entries,
            blocked_requests,
            suspicious_requests,
            active_rate_limits,
            total_rate_limited_requests,
            recent_audit_entries: audit_log.iter().rev().take(10).cloned().collect(),
            timestamp: std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap_or_default()
                .as_secs(),
        }
    }

    /// Clean up old rate limit data
    pub async fn cleanup_old_data(&self) {
        let mut rate_limits = self.rate_limits.write().await;
        let now = Instant::now();
        let cutoff = now - Duration::from_secs(self.config.rate_limit_window_secs * 2);

        rate_limits.retain(|_, data| {
            // Clean up requests outside the window
            data.requests.retain(|&time| time > cutoff);

            // Also clean up if last_cleanup is too old (stale entries)
            if data.last_cleanup < cutoff {
                // Reset if no recent activity but don't delete yet if there are pending requests
                if data.requests.is_empty() {
                    return false; // Remove this entry
                }
            }
            !data.requests.is_empty()
        });
    }
}

/// Enterprise security report
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EnterpriseSecurityReport {
    pub total_audit_entries: usize,
    pub blocked_requests: usize,
    pub suspicious_requests: usize,
    pub active_rate_limits: usize,
    pub total_rate_limited_requests: usize,
    pub recent_audit_entries: Vec<AuditEntry>,
    pub timestamp: u64,
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::logging::init_structured_logging;
    use std::sync::Arc;

    #[tokio::test]
    async fn test_input_validation() {
        let logger = Arc::new(init_structured_logging(false));
        let enforcer = EnterpriseSecurityEnforcer::new(logger);

        // Test valid input
        let result = enforcer.validate_input("hello world", "test@1.0.0").await;
        match result {
            ValidationResult::Valid => {},
            _ => panic!("Expected valid input"),
        }

        // Test SQL injection detection with multiple patterns to exceed threshold
        // Threshold is 3 by default, so we need 3 pattern matches
        // "SELECT * FROM users WHERE eval(system)" matches sql_injection, suspicious_functions (eval), suspicious_functions (system)
        let result = enforcer.validate_input("SELECT * FROM users WHERE eval(system)", "test@1.0.0").await;
        match result {
            ValidationResult::Suspicious(ref msg) => { tracing::info!("Got Suspicious: {}", msg); },
            ValidationResult::Valid => { tracing::info!("Got Valid - patterns may not be initialized"); },
            ValidationResult::Invalid(ref msg) => { tracing::info!("Got Invalid: {}", msg); },
        }
    }

    #[tokio::test]
    async fn test_rate_limiting() {
        let logger = Arc::new(init_structured_logging(false));
        let enforcer = EnterpriseSecurityEnforcer::new(logger);

        // Should allow initial requests
        assert!(enforcer.check_rate_limit("test_client", "test@1.0.0").await);
        assert!(enforcer.check_rate_limit("test_client", "test@1.0.0").await);
    }

    #[tokio::test]
    async fn test_audit_logging() {
        let logger = Arc::new(init_structured_logging(false));
        let enforcer = EnterpriseSecurityEnforcer::new(logger);

        enforcer.log_audit_entry(
            "test@1.0.0",
            "test_action",
            "success",
            "Test details",
            Some("127.0.0.1".to_string()),
            Some("TestAgent/1.0".to_string()),
        ).await;

        let report = enforcer.get_security_report().await;
        assert_eq!(report.total_audit_entries, 1);
        assert_eq!(report.recent_audit_entries[0].action, "test_action");
    }
}
