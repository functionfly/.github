//! YARA scan integration for validating WASM artifacts before execution.
//!
//! Before a new WASM artifact is stored or executed for the first time, this
//! module calls the existing [`deploy/yara/yara_service.py`] HTTP service to
//! scan the bytes for known malicious patterns.
//!
//! # Phase 2 implementation
//!
//! Addresses the gap identified in `plans/SANDBOX_EXECUTION_LAYER.md`:
//! > WASM module is not verified before loading (no YARA/signature check) — Low
//! > Integrate the existing `deploy/yara/` service to scan WASM bytes before
//! > instantiation.
//!
//! # Integration
//!
//! The YARA service exposes a simple HTTP API:
//! ```
//! POST /scan
//! Content-Type: application/octet-stream
//! Body: <raw WASM bytes>
//!
//! Response 200: { "matched": false, "rules": [] }
//! Response 200: { "matched": true,  "rules": ["rule_name", ...] }
//! Response 500: { "error": "..." }
//! ```
//!
//! If the service is unavailable and `fail_open` is `true` (default), the scan
//! is skipped and execution proceeds.  Set `fail_open = false` to enforce
//! strict scanning (execution is blocked if the service is unreachable).

use std::time::Duration;
use serde::{Deserialize, Serialize};

/// Result of a YARA scan.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct YaraScanResult {
    /// Whether any YARA rules matched.
    pub matched: bool,
    /// Names of the rules that matched (empty if `matched` is false).
    pub rules: Vec<String>,
    /// Whether the scan was skipped (e.g. service unavailable with fail_open).
    pub skipped: bool,
    /// Optional error message.
    pub error: Option<String>,
}

impl YaraScanResult {
    /// A clean scan result (no matches).
    pub fn clean() -> Self {
        Self {
            matched: false,
            rules: Vec::new(),
            skipped: false,
            error: None,
        }
    }

    /// A skipped scan result.
    pub fn skipped(reason: impl Into<String>) -> Self {
        Self {
            matched: false,
            rules: Vec::new(),
            skipped: true,
            error: Some(reason.into()),
        }
    }

    /// Whether execution should be blocked based on this result.
    pub fn should_block(&self) -> bool {
        self.matched
    }
}

/// Configuration for the YARA scanner.
#[derive(Debug, Clone)]
pub struct YaraScannerConfig {
    /// URL of the YARA service (e.g. `http://localhost:5000`).
    pub service_url: String,
    /// Request timeout.
    pub timeout: Duration,
    /// If `true`, allow execution when the YARA service is unreachable.
    /// If `false`, block execution when the service is unreachable.
    pub fail_open: bool,
    /// Whether scanning is enabled at all.
    pub enabled: bool,
}

impl Default for YaraScannerConfig {
    fn default() -> Self {
        Self {
            service_url: "http://localhost:5000".to_string(),
            timeout: Duration::from_secs(5),
            fail_open: true,
            enabled: false, // Disabled by default; opt-in
        }
    }
}

/// YARA scanner client.
pub struct YaraScanner {
    config: YaraScannerConfig,
    client: reqwest::Client,
}

impl YaraScanner {
    /// Create a new YARA scanner.
    pub fn new(config: YaraScannerConfig) -> Result<Self, reqwest::Error> {
        let client = reqwest::Client::builder()
            .timeout(config.timeout)
            .build()?;
        Ok(Self { config, client })
    }

    /// Scan `wasm_bytes` using the YARA service.
    ///
    /// Returns a `YaraScanResult` describing whether any rules matched.
    /// Never panics — errors are captured in the result.
    pub async fn scan(&self, wasm_bytes: &[u8]) -> YaraScanResult {
        if !self.config.enabled {
            return YaraScanResult::clean();
        }

        let url = format!("{}/scan", self.config.service_url.trim_end_matches('/'));

        match self
            .client
            .post(&url)
            .header("Content-Type", "application/octet-stream")
            .body(wasm_bytes.to_vec())
            .send()
            .await
        {
            Ok(response) => {
                if response.status().is_success() {
                    match response.text().await {
                        Ok(body) => match serde_json::from_str::<YaraScanResult>(&body) {
                            Ok(result) => {
                                if result.matched {
                                    tracing::warn!(
                                        "YaraScanner: WASM artifact matched rules: {:?}",
                                        result.rules
                                    );
                                } else {
                                    tracing::debug!("YaraScanner: clean scan");
                                }
                                result
                            }
                            Err(e) => {
                                let msg = format!("Failed to parse YARA response: {}", e);
                                tracing::warn!("YaraScanner: {}", msg);
                                if self.config.fail_open {
                                    YaraScanResult::skipped(msg)
                                } else {
                                    YaraScanResult {
                                        matched: true,
                                        rules: vec!["parse_error".to_string()],
                                        skipped: false,
                                        error: Some(msg),
                                    }
                                }
                            }
                        },
                        Err(e) => {
                            let msg = format!("Failed to read YARA response body: {}", e);
                            tracing::warn!("YaraScanner: {}", msg);
                            if self.config.fail_open {
                                YaraScanResult::skipped(msg)
                            } else {
                                YaraScanResult {
                                    matched: true,
                                    rules: vec!["read_error".to_string()],
                                    skipped: false,
                                    error: Some(msg),
                                }
                            }
                        }
                    }
                } else {
                    let msg = format!("YARA service returned HTTP {}", response.status());
                    tracing::warn!("YaraScanner: {}", msg);
                    if self.config.fail_open {
                        YaraScanResult::skipped(msg)
                    } else {
                        YaraScanResult {
                            matched: true,
                            rules: vec!["service_error".to_string()],
                            skipped: false,
                            error: Some(msg),
                        }
                    }
                }
            }
            Err(e) => {
                let msg = format!("YARA service unreachable: {}", e);
                tracing::warn!("YaraScanner: {}", msg);
                if self.config.fail_open {
                    YaraScanResult::skipped(msg)
                } else {
                    // fail_closed: block execution
                    YaraScanResult {
                        matched: true,
                        rules: vec!["service_unavailable".to_string()],
                        skipped: false,
                        error: Some(msg),
                    }
                }
            }
        }
    }

    /// Convenience method: scan and return an error if the artifact should be blocked.
    pub async fn scan_or_block(&self, wasm_bytes: &[u8]) -> anyhow::Result<()> {
        let result = self.scan(wasm_bytes).await;
        if result.should_block() {
            Err(anyhow::anyhow!(
                "WASM artifact blocked by YARA scanner: matched rules {:?}",
                result.rules
            ))
        } else {
            Ok(())
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_scan_result_clean() {
        let r = YaraScanResult::clean();
        assert!(!r.matched);
        assert!(!r.should_block());
        assert!(!r.skipped);
    }

    #[test]
    fn test_scan_result_skipped() {
        let r = YaraScanResult::skipped("service down");
        assert!(!r.matched);
        assert!(!r.should_block());
        assert!(r.skipped);
    }

    #[test]
    fn test_scan_result_matched_blocks() {
        let r = YaraScanResult {
            matched: true,
            rules: vec!["malware_rule".to_string()],
            skipped: false,
            error: None,
        };
        assert!(r.should_block());
    }

    #[tokio::test]
    async fn test_scanner_disabled_returns_clean() {
        let config = YaraScannerConfig {
            enabled: false,
            ..Default::default()
        };
        let scanner = YaraScanner::new(config).expect("Failed to create scanner");
        let result = scanner.scan(b"\x00asm\x01\x00\x00\x00").await;
        assert!(!result.matched);
        assert!(!result.skipped);
    }

    #[tokio::test]
    async fn test_scanner_fail_open_when_service_down() {
        let config = YaraScannerConfig {
            service_url: "http://127.0.0.1:19999".to_string(), // nothing listening
            enabled: true,
            fail_open: true,
            timeout: Duration::from_millis(100),
        };
        let scanner = YaraScanner::new(config).expect("Failed to create scanner");
        let result = scanner.scan(b"\x00asm\x01\x00\x00\x00").await;
        // fail_open: should be skipped, not blocked
        assert!(result.skipped);
        assert!(!result.should_block());
    }

    #[tokio::test]
    async fn test_scanner_fail_closed_when_service_down() {
        let config = YaraScannerConfig {
            service_url: "http://127.0.0.1:19999".to_string(), // nothing listening
            enabled: true,
            fail_open: false,
            timeout: Duration::from_millis(100),
        };
        let scanner = YaraScanner::new(config).expect("Failed to create scanner");
        let result = scanner.scan(b"\x00asm\x01\x00\x00\x00").await;
        // fail_closed: should block
        assert!(result.should_block());
    }
}
