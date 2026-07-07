//! Object storage abstraction for events and snapshots

use async_trait::async_trait;
use aws_sdk_s3::{config::{Credentials, Region}, primitives::ByteStream, Client};
use thiserror::Error;

/// Errors that can occur in object storage operations
#[derive(Error, Debug)]
pub enum StorageError {
    #[error("Object not found: {0}")]
    NotFound(String),

    #[error("Upload failed: {0}")]
    UploadFailed(String),

    #[error("Download failed: {0}")]
    DownloadFailed(String),

    #[error("Delete failed: {0}")]
    DeleteFailed(String),

    #[error("Invalid configuration: {0}")]
    InvalidConfig(String),

    #[error("Checksum mismatch: expected {expected}, got {actual}")]
    ChecksumMismatch {
        expected: String,
        actual: String,
    },
}

/// Result type for storage operations
pub type StorageResult<T> = Result<T, StorageError>;

/// Trait for object storage backends
#[async_trait]
pub trait ObjectStore: Send + Sync {
    /// Upload an object
    async fn put(&self, key: &str, data: &[u8], content_type: Option<&str>) -> StorageResult<()>;

    /// Download an object
    async fn get(&self, key: &str) -> StorageResult<Vec<u8>>;

    /// Delete an object
    async fn delete(&self, key: &str) -> StorageResult<()>;

    /// Check if an object exists
    async fn exists(&self, key: &str) -> StorageResult<bool>;

    /// Generate a storage key for an event
    fn event_key(&self, state_id: &uuid::Uuid, event_id: &uuid::Uuid) -> String {
        format!("events/{}/{}.json", state_id, event_id)
    }

    /// Generate a storage key for a snapshot
    fn snapshot_key(&self, state_id: &uuid::Uuid, snapshot_id: &uuid::Uuid) -> String {
        format!("snapshots/{}/{}.json", state_id, snapshot_id)
    }
}

/// Storage backend type
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum StorageBackend {
    /// Cloudflare R2
    R2,
    /// Backblaze B2
    B2,
    /// Wasabi
    Wasabi,
    /// Local filesystem (development)
    Local,
}

impl StorageBackend {
    /// Detect backend from environment
    pub fn from_env() -> Option<Self> {
        // Check for R2
        if std::env::var("R2_ACCOUNT_ID").is_ok() {
            return Some(StorageBackend::R2);
        }

        // Check for B2
        if std::env::var("B2_APPLICATION_KEY_ID").is_ok() {
            return Some(StorageBackend::B2);
        }

        // Check for Wasabi
        if std::env::var("WASABI_ACCESS_KEY").is_ok() {
            return Some(StorageBackend::Wasabi);
        }

        // Check for local development
        if std::env::var("STATEFABRIC_LOCAL_STORAGE").is_ok() {
            return Some(StorageBackend::Local);
        }

        None
    }
}

/// Configuration for object storage
#[derive(Debug, Clone)]
pub struct StorageConfig {
    pub backend: StorageBackend,
    pub bucket: String,
    pub region: Option<String>,
    pub endpoint: Option<String>,
    pub access_key: Option<String>,
    pub secret_key: Option<String>,
    pub path_prefix: Option<String>,
}

impl StorageConfig {
    /// Create config from environment variables
    pub fn from_env() -> Result<Self, StorageError> {
        let backend = StorageBackend::from_env()
            .ok_or_else(|| StorageError::InvalidConfig("No storage backend configured".to_string()))?;

        let (bucket, region, endpoint, access_key, secret_key) = match backend {
            StorageBackend::R2 => {
                let bucket = std::env::var("R2_BUCKET")
                    .map_err(|_| StorageError::InvalidConfig("R2_BUCKET not set".to_string()))?;
                let account_id = std::env::var("R2_ACCOUNT_ID")
                    .map_err(|_| StorageError::InvalidConfig("R2_ACCOUNT_ID not set".to_string()))?;
                let access_key = std::env::var("R2_ACCESS_KEY_ID")
                    .map_err(|_| StorageError::InvalidConfig("R2_ACCESS_KEY_ID not set".to_string()))?;
                let secret_key = std::env::var("R2_SECRET_ACCESS_KEY")
                    .map_err(|_| StorageError::InvalidConfig("R2_SECRET_ACCESS_KEY not set".to_string()))?;

                let endpoint = format!("https://{}.r2.cloudflarestorage.com", account_id);

                (bucket, Some("auto".to_string()), Some(endpoint), Some(access_key), Some(secret_key))
            }
            StorageBackend::B2 => {
                let bucket = std::env::var("B2_BUCKET")
                    .map_err(|_| StorageError::InvalidConfig("B2_BUCKET not set".to_string()))?;
                let key_id = std::env::var("B2_APPLICATION_KEY_ID")
                    .map_err(|_| StorageError::InvalidConfig("B2_APPLICATION_KEY_ID not set".to_string()))?;
                let app_key = std::env::var("B2_APPLICATION_KEY")
                    .map_err(|_| StorageError::InvalidConfig("B2_APPLICATION_KEY not set".to_string()))?;

                (bucket, Some("us-west-000".to_string()), None, Some(key_id), Some(app_key))
            }
            StorageBackend::Wasabi => {
                let bucket = std::env::var("WASABI_BUCKET")
                    .map_err(|_| StorageError::InvalidConfig("WASABI_BUCKET not set".to_string()))?;
                let access_key = std::env::var("WASABI_ACCESS_KEY")
                    .map_err(|_| StorageError::InvalidConfig("WASABI_ACCESS_KEY not set".to_string()))?;
                let secret_key = std::env::var("WASABI_SECRET_KEY")
                    .map_err(|_| StorageError::InvalidConfig("WASABI_SECRET_KEY not set".to_string()))?;

                let endpoint = "https://s3.wasabisys.com".to_string();

                (bucket, Some("us-east-1".to_string()), Some(endpoint), Some(access_key), Some(secret_key))
            }
            StorageBackend::Local => {
                let bucket = std::env::var("STATEFABRIC_LOCAL_BUCKET")
                    .unwrap_or_else(|_| "local".to_string());

                (bucket, None, None, None, None)
            }
        };

        Ok(StorageConfig {
            backend,
            bucket,
            region,
            endpoint,
            access_key,
            secret_key,
            path_prefix: None,
        })
    }
}

/// S3-compatible object storage client
pub struct S3ObjectStore {
    client: Client,
    bucket: String,
}

impl S3ObjectStore {
    /// Create a new S3-compatible client
    pub async fn new(config: &StorageConfig) -> StorageResult<Self> {
        let mut s3_config_builder = aws_config::defaults(aws_config::BehaviorVersion::latest());

        // Set credentials if provided
        if let (Some(access_key), Some(secret_key)) = (&config.access_key, &config.secret_key) {
            s3_config_builder = s3_config_builder.credentials_provider(
                Credentials::new(
                    access_key,
                    secret_key,
                    None,
                    None,
                    "statefabric",
                ),
            );
        }

        // Set region if provided
        if let Some(region) = &config.region {
            s3_config_builder = s3_config_builder.region(Region::new(region.clone()));
        }

        // Set custom endpoint if provided (for R2, Wasabi, etc.)
        let mut s3_config = s3_config_builder.load().await;
        if let Some(endpoint) = &config.endpoint {
            s3_config = s3_config.to_builder()
                .endpoint_url(endpoint)
                .build();
        }

        let client = Client::new(&s3_config);

        Ok(Self {
            client,
            bucket: config.bucket.clone(),
        })
    }
}

#[async_trait]
impl ObjectStore for S3ObjectStore {
    async fn put(&self, key: &str, data: &[u8], content_type: Option<&str>) -> StorageResult<()> {
        let body = ByteStream::from(data.to_vec());

        let mut request = self.client
            .put_object()
            .bucket(&self.bucket)
            .key(key)
            .body(body);

        if let Some(content_type) = content_type {
            request = request.content_type(content_type);
        }

        request
            .send()
            .await
            .map_err(|e| StorageError::UploadFailed(e.to_string()))?;

        Ok(())
    }

    async fn get(&self, key: &str) -> StorageResult<Vec<u8>> {
        let response = self.client
            .get_object()
            .bucket(&self.bucket)
            .key(key)
            .send()
            .await
            .map_err(|e| {
                if e.to_string().contains("NoSuchKey") || e.to_string().contains("404") {
                    StorageError::NotFound(key.to_string())
                } else {
                    StorageError::DownloadFailed(e.to_string())
                }
            })?;

        let data = response
            .body
            .collect()
            .await
            .map_err(|e| StorageError::DownloadFailed(e.to_string()))?;

        Ok(data.into_bytes().to_vec())
    }

    async fn delete(&self, key: &str) -> StorageResult<()> {
        self.client
            .delete_object()
            .bucket(&self.bucket)
            .key(key)
            .send()
            .await
            .map_err(|e| StorageError::DeleteFailed(e.to_string()))?;

        Ok(())
    }

    async fn exists(&self, key: &str) -> StorageResult<bool> {
        match self.client
            .head_object()
            .bucket(&self.bucket)
            .key(key)
            .send()
            .await
        {
            Ok(_) => Ok(true),
            Err(e) => {
                if e.to_string().contains("NoSuchKey") || e.to_string().contains("404") {
                    Ok(false)
                } else {
                    Err(StorageError::DownloadFailed(e.to_string()))
                }
            }
        }
    }
}

/// Cloudflare R2 object storage client
pub type R2ObjectStore = S3ObjectStore;

/// Backblaze B2 object storage client
pub type B2ObjectStore = S3ObjectStore;

/// Wasabi object storage client
pub type WasabiObjectStore = S3ObjectStore;

/// Encrypted object storage wrapper
/// 
/// Wraps any ObjectStore implementation and encrypts data at rest using AES-256-GCM.
/// The encryption key is loaded from the STATEFABRIC_ENCRYPTION_KEY environment variable.
/// If no key is set, encryption is skipped (for development only).
/// 
/// SECURITY NOTE: This provides encryption at rest for blob storage (events, snapshots).
/// The key must be provided via environment variable, never hardcoded.
pub struct EncryptedObjectStore {
    inner: Box<dyn ObjectStore + Send + Sync>,
    encryptor: Option<crate::crypto::ObjectEncryptor>,
    encryption_enabled: bool,
}

impl EncryptedObjectStore {
    /// Create a new encrypted wrapper around an ObjectStore
    pub fn new(inner: Box<dyn ObjectStore + Send + Sync>) -> Self {
        let encryptor = match crate::crypto::ObjectEncryptor::from_env() {
            Ok(e) => {
                tracing::info!("Object storage encryption enabled (AES-256-GCM)");
                Some(e)
            }
            Err(crate::crypto::CryptoError::KeyNotConfigured) => {
                tracing::warn!("Object storage encryption DISABLED: STATEFABRIC_ENCRYPTION_KEY not set");
                None
            }
            Err(e) => {
                tracing::error!("Object storage encryption FAILED to initialize: {}", e);
                None
            }
        };

        let encryption_enabled = encryptor.is_some();

        Self {
            inner,
            encryptor,
            encryption_enabled,
        }
    }

    /// Check if encryption is enabled and active
    pub fn is_encryption_enabled(&self) -> bool {
        self.encryption_enabled
    }
}

#[async_trait]
impl ObjectStore for EncryptedObjectStore {
    async fn put(&self, key: &str, data: &[u8], content_type: Option<&str>) -> StorageResult<()> {
        let data_to_store = if let Some(ref encryptor) = self.encryptor {
            // Encrypt the data with AES-256-GCM
            encryptor.encrypt(data)
                .map_err(|e| StorageError::UploadFailed(format!("Encryption failed: {}", e)))?
        } else {
            data.to_vec()
        };

        // Prepend a magic byte to indicate encryption status (for future compatibility)
        let mut payload = vec![if self.encryption_enabled { 0x01 } else { 0x00 }];
        payload.extend(data_to_store);

        self.inner.put(key, &payload, content_type).await
    }

    async fn get(&self, key: &str) -> StorageResult<Vec<u8>> {
        let payload = self.inner.get(key).await?;

        if payload.is_empty() {
            return Ok(payload);
        }

        // Check encryption magic byte
        let encryption_flag = payload[0];
        let encrypted_data = &payload[1..];

        if encryption_flag == 0x01 {
            // Encrypted data
            if let Some(ref encryptor) = self.encryptor {
                encryptor.decrypt(encrypted_data)
                    .map_err(|e| StorageError::DownloadFailed(format!("Decryption failed: {}", e)))
            } else {
                Err(StorageError::DownloadFailed(
                    "Data is encrypted but decryption key is not available".to_string()
                ))
            }
        } else {
            // Unencrypted data (legacy or development)
            Ok(encrypted_data.to_vec())
        }
    }

    async fn delete(&self, key: &str) -> StorageResult<()> {
        self.inner.delete(key).await
    }

    async fn exists(&self, key: &str) -> StorageResult<bool> {
        self.inner.exists(key).await
    }
}

/// Local filesystem object storage client for development
pub struct LocalObjectStore {
    base_path: std::path::PathBuf,
}

impl LocalObjectStore {
    /// Create a new local filesystem client
    pub fn new(config: &StorageConfig) -> StorageResult<Self> {
        let base_path = std::env::var("STATEFABRIC_LOCAL_STORAGE")
            .unwrap_or_else(|_| "./data".to_string());

        let path = std::path::PathBuf::from(base_path).join(&config.bucket);

        // Create the directory if it doesn't exist
        std::fs::create_dir_all(&path)
            .map_err(|e| StorageError::InvalidConfig(format!("Failed to create storage directory: {}", e)))?;

        Ok(Self {
            base_path: path,
        })
    }
}

#[async_trait]
impl ObjectStore for LocalObjectStore {
    async fn put(&self, key: &str, data: &[u8], _content_type: Option<&str>) -> StorageResult<()> {
        let file_path = self.base_path.join(key);

        // Create parent directories if they don't exist
        if let Some(parent) = file_path.parent() {
            std::fs::create_dir_all(parent)
                .map_err(|e| StorageError::UploadFailed(format!("Failed to create directories: {}", e)))?;
        }

        std::fs::write(&file_path, data)
            .map_err(|e| StorageError::UploadFailed(format!("Failed to write file: {}", e)))?;

        Ok(())
    }

    async fn get(&self, key: &str) -> StorageResult<Vec<u8>> {
        let file_path = self.base_path.join(key);

        if !file_path.exists() {
            return Err(StorageError::NotFound(key.to_string()));
        }

        std::fs::read(&file_path)
            .map_err(|e| StorageError::DownloadFailed(format!("Failed to read file: {}", e)))
    }

    async fn delete(&self, key: &str) -> StorageResult<()> {
        let file_path = self.base_path.join(key);

        if !file_path.exists() {
            return Ok(()); // Idempotent delete
        }

        std::fs::remove_file(&file_path)
            .map_err(|e| StorageError::DeleteFailed(format!("Failed to delete file: {}", e)))?;

        Ok(())
    }

    async fn exists(&self, key: &str) -> StorageResult<bool> {
        let file_path = self.base_path.join(key);
        Ok(file_path.exists())
    }
}

/// Create an object store client based on the configuration
pub async fn create_object_store(config: &StorageConfig) -> StorageResult<Box<dyn ObjectStore>> {
    match config.backend {
        StorageBackend::R2 => {
            let client = R2ObjectStore::new(config).await?;
            Ok(Box::new(client))
        }
        StorageBackend::B2 => {
            let client = B2ObjectStore::new(config).await?;
            Ok(Box::new(client))
        }
        StorageBackend::Wasabi => {
            let client = WasabiObjectStore::new(config).await?;
            Ok(Box::new(client))
        }
        StorageBackend::Local => {
            let client = LocalObjectStore::new(config)?;
            Ok(Box::new(client))
        }
    }
}

