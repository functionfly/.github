//! FunctionFly Package (.ffpkg) handling

use std::path::Path;
use serde::{Deserialize, Serialize};
use hex;
use ed25519_dalek::VerifyingKey;

/// A FunctionFly Universal WASM Package (.ffpkg)
#[allow(dead_code)]
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Package {
    pub manifest_version: String,
    pub metadata: PackageMetadata,
    pub modules: Vec<PackageModule>,
    pub resources: Vec<PackageResource>,
    pub signature: Option<PackageSignature>,
}

/// Package metadata
#[allow(dead_code)]
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PackageMetadata {
    pub name: String,
    pub version: String,
    pub description: String,
    pub runtime: String,
    pub languages: Vec<String>,
    pub capabilities: Vec<String>,
}

/// A module within a package
#[allow(dead_code)]
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PackageModule {
    pub module_id: String,
    pub name: String,
    pub language: String,
    pub bytecode: Vec<u8>,
    pub entry_point: String,
}

/// A resource within a package
#[allow(dead_code)]
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PackageResource {
    pub resource_id: String,
    pub name: String,
    pub resource_type: String,
    pub content: Vec<u8>,
}

/// Package signature for verification
#[allow(dead_code)]
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PackageSignature {
    pub algorithm: String,
    pub signature: Vec<u8>,
    pub public_key: String,
    pub signed_at: i64,
}

/// Package builder for creating .ffpkg files
#[allow(dead_code)]
pub struct PackageBuilder {
    package: Package,
}

impl PackageBuilder {
    #[allow(dead_code)]
    pub fn new(name: &str, version: &str) -> Self {
        let package = Package {
            manifest_version: "1.0.0".to_string(),
            metadata: PackageMetadata {
                name: name.to_string(),
                version: version.to_string(),
                description: String::new(),
                runtime: "wasm".to_string(),
                languages: Vec::new(),
                capabilities: Vec::new(),
            },
            modules: Vec::new(),
            resources: Vec::new(),
            signature: None,
        };

        Self { package }
    }

    #[allow(dead_code)]
    pub fn with_description(mut self, desc: &str) -> Self {
        self.package.metadata.description = desc.to_string();
        self
    }

    #[allow(dead_code)]
    pub fn with_runtime(mut self, runtime: &str) -> Self {
        self.package.metadata.runtime = runtime.to_string();
        self
    }

    /// Add a programming language to the package metadata
    #[allow(dead_code)]
    pub fn with_language(mut self, lang: &str) -> Self {
        self.package.metadata.languages.push(lang.to_string());
        self
    }

    /// Add a capability to the package metadata
    #[allow(dead_code)]
    pub fn with_capability(mut self, cap: &str) -> Self {
        self.package.metadata.capabilities.push(cap.to_string());
        self
    }

    #[allow(dead_code)]
    pub fn add_module(mut self, module: PackageModule) -> Self {
        self.package.modules.push(module);
        self
    }

    /// Add a resource file to the package (e.g., config, assets)
    #[allow(dead_code)]
    pub fn add_resource(mut self, resource: PackageResource) -> Self {
        self.package.resources.push(resource);
        self
    }

    #[allow(dead_code)]
    pub fn build(self) -> Package {
        self.package
    }

    /// Write package to a file
    #[allow(dead_code)]
    pub fn write(&self, path: &Path) -> std::io::Result<()> {
        let data = serde_json::to_vec_pretty(&self.package)
            .map_err(|e| std::io::Error::new(std::io::ErrorKind::InvalidData, e))?;
        std::fs::write(path, data)
    }
}

impl Package {
    /// Load a package from a file
    #[allow(dead_code)]
    pub fn load(path: &Path) -> std::io::Result<Self> {
        let data = std::fs::read(path)?;
        serde_json::from_slice(&data)
            .map_err(|e| std::io::Error::new(std::io::ErrorKind::InvalidData, e))
    }

    /// Verify package signature
    #[allow(dead_code)]
    pub fn verify(&self) -> Result<bool, String> {
        // Signature verification using Ed25519
        let signature = self.signature.as_ref()
            .ok_or("Package has no signature")?;

        if signature.algorithm != "ed25519" {
            return Err(format!("Unsupported signature algorithm: {}", signature.algorithm));
        }

        // Parse public key from hex
        let public_key_bytes = hex::decode(&signature.public_key)
            .map_err(|e| format!("Invalid public key hex: {}", e))?;

        if public_key_bytes.len() != 32 {
            return Err("Invalid public key length".to_string());
        }

        // Convert Vec<u8> to [u8; 32] for VerifyingKey::from_bytes
        let mut key_array = [0u8; 32];
        key_array.copy_from_slice(&public_key_bytes);
        let verifying_key = VerifyingKey::from_bytes(&key_array)
            .map_err(|e| format!("Invalid Ed25519 public key: {}", e))?;

        // Serialize package (without signature) for verification
        let mut pkg_for_verify = self.clone();
        pkg_for_verify.signature = None;
        let pkg_bytes = serde_json::to_vec(&pkg_for_verify)
            .map_err(|e| format!("Failed to serialize package: {}", e))?;

        // Parse signature bytes
        if signature.signature.len() != 64 {
            return Err("Invalid signature length".to_string());
        }

        let sig_bytes: [u8; 64] = signature.signature[..64].try_into()
            .map_err(|_| "Signature bytes conversion failed")?;

        let sig = ed25519_dalek::Signature::from_bytes(&sig_bytes);

        // Verify signature using ed25519-dalek 2.x API
        Ok(verifying_key.verify_strict(&pkg_bytes, &sig).is_ok())
    }

    /// Validate package structure
    #[allow(dead_code)]
    pub fn validate(&self) -> Result<(), String> {
        if self.manifest_version.is_empty() {
            return Err("Manifest version is required".to_string());
        }
        if self.metadata.name.is_empty() {
            return Err("Package name is required".to_string());
        }
        if self.modules.is_empty() {
            return Err("Package must contain at least one module".to_string());
        }

        for module in &self.modules {
            if module.module_id.is_empty() {
                return Err("Module ID is required".to_string());
            }
            // Validate WASM magic number
            if module.bytecode.len() < 4 {
                return Err(format!("Module {} has invalid bytecode", module.module_id));
            }
            if &module.bytecode[0..4] != b"\0asm" {
                return Err(format!("Module {} is not valid WASM", module.module_id));
            }
        }

        Ok(())
    }
}