//! Enterprise security hardening features.
//!
//! This module provides advanced security features for enterprise deployments,
//! including enhanced input validation, sandboxing improvements, and security best practices.

use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::RwLock;
use serde::{Deserialize, Serialize};
use regex::Regex;

use crate::config::Config;
use crate::logging::StructuredLogger;

/// Enterprise security configuration
#[derive(Debug, Clone)]
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
        let mut enforcer = Self {
            config,
            rate_limits: Arc::new(RwLock::new(HashMap::new())),
            security_patterns: Arc::new(RwLock::new(HashMap::new())),
            logger,
            audit_log: Arc::new(RwLock::new(Vec::new())),
        };

        enforcer.initialize_security_patterns();
        enforcer
    }

    /// Initialize security patterns for threat detection
    fn initialize_security_patterns(&mut self) {
        let patterns = vec![
            ("sql_injection", r"(?i)(union\s+select|select\s+.*\s+from|drop\s+table|insert\s+into|update\s+.*\s+set)", "SQL Injection", "high"),
            ("xss", r"<script[^>]*>.*?</script>|<iframe[^>]*>.*?</iframe>|<object[^>]*>.*?</object>", "Cross-Site Scripting", "high"),
            ("command_injection", r"[;&|`$()]", "Command Injection", "critical"),
            ("path_traversal", r"\.\./|\.\.\\", "Path Traversal", "high"),
            ("suspicious_functions", r"(?i)(eval|exec|system|shell_exec|popen)", "Dangerous Function Call", "medium"),
        ];

        let rt = tokio::runtime::Runtime::new().unwrap();
        rt.block_on(async {
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
        });
    }

    /// Validate input data
    pub async fn validate_input(&self, input: &str, function_key: &str) -> ValidationResult {
        // Check input size
        if input.len() > self.config.max_input_size {
            let reason = format!("Input size {} exceeds maximum allowed size {}", input.len(), self.config.max_input_size);
            self.log_audit_entry(function_key, "input_validation", "blocked", &reason, None, None).await;
            return ValidationResult::Invalid(reason);
        }

        // Check for suspicious patterns
        let patterns = self.security_patterns.read().await;
        let mut suspicious_patterns = Vec::new();

        for (pattern_name, pattern) in patterns.iter() {
            if pattern.pattern.is_match(input) {
                suspicious_patterns.push(pattern_name.clone());
            }
        }

        if suspicious_patterns.len() >= self.config.suspicious_pattern_threshold {
            let reason = format!("Multiple suspicious patterns detected: {:?}", suspicious_patterns);
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

        // Clean up old requests outside the window
        let window_start = now - Duration::from_secs(self.config.rate_limit_window_secs);
        data.requests.retain(|&time| time > window_start);

        // Check if under limit
        if data.requests.len() >= self.config.max_requests_per_window {
            self.log_audit_entry(function_key, "rate_limit", "blocked", "Rate limit exceeded", None, None).await;
            return false;
        }

        // Add current request
        data.requests.push(now);
        true
    }

    /// Enhanced sandboxing validation
    pub async fn validate_sandboxing(&self, function_key: &str, capabilities: &[String]) -> ValidationResult {
        // Check for dangerous capability combinations
        let dangerous_combinations = vec![
            (vec!["fetch", "storage"], "Network and storage access together"),
            (vec!["crypto", "external_api"], "Crypto and external API access together"),
        ];

        for (combo, reason) in dangerous_combinations {
            if combo.iter().all(|cap| capabilities.contains(&cap.to_string())) {
                let violation_reason = format!("Dangerous capability combination: {}", reason);
                self.log_audit_entry(function_key, "sandboxing", "blocked", &violation_reason, None, None).await;
                return ValidationResult::Invalid(violation_reason);
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

        // Log to structured logger
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
            data.requests.retain(|&time| time > cutoff);
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

        // Test SQL injection detection
        let result = enforcer.validate_input("SELECT * FROM users", "test@1.0.0").await;
        match result {
            ValidationResult::Suspicious(_) => {},
            _ => panic!("Expected suspicious input for SQL injection"),
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
