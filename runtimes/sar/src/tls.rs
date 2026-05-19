use std::path::Path;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum TlsError {
    #[error("Failed to read certificate file: {0}")]
    CertReadError(#[from] std::io::Error),
    #[error("Failed to parse certificate: {0}")]
    CertParseError(String),
    #[error("TLS not configured")]
    NotConfigured,
}

#[derive(Clone, Debug)]
pub struct TlsConfig {
    pub cert_path: String,
    pub key_path: String,
    pub ca_cert_path: Option<String>,
}

impl TlsConfig {
    pub fn from_env() -> Option<Self> {
        let cert_path = std::env::var("SAR_TLS_CERT").ok()?;
        let key_path = std::env::var("SAR_TLS_KEY").ok()?;

        Some(Self {
            cert_path,
            key_path,
            ca_cert_path: std::env::var("SAR_TLS_CA_CERT").ok(),
        })
    }

    pub fn enabled(&self) -> bool {
        Path::new(&self.cert_path).exists() && Path::new(&self.key_path).exists()
    }
}

pub struct NatsTlsConfig {
    pub enabled: bool,
    pub skip_verification: bool,
    pub ca_cert_path: Option<String>,
}

impl NatsTlsConfig {
    pub fn from_env() -> Option<Self> {
        let enabled = std::env::var("SAR_NATS_TLS_ENABLED").is_ok();
        if !enabled {
            return None;
        }

        Some(Self {
            enabled: true,
            skip_verification: std::env::var("SAR_NATS_TLS_SKIP_VERIFICATION").is_ok(),
            ca_cert_path: std::env::var("SAR_NATS_TLS_CA_CERT").ok(),
        })
    }
}