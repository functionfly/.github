//! Enterprise package caching and management.
//!
//! This module provides advanced package caching capabilities for enterprise deployments,
//! including dependency resolution caching, package download caching, and bytecode caching.

use anyhow::Context;
use std::collections::{HashMap, HashSet};
use std::fs;
use std::path::PathBuf;
use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::RwLock;

use crate::cache::ResultCache;
use crate::host_functions::fetch;

/// Max size for a single package download (100 MiB)
pub const MAX_PACKAGE_DOWNLOAD_BYTES: usize = 100 * 1024 * 1024;
/// Timeout for package HTTP download
pub const PACKAGE_DOWNLOAD_TIMEOUT: Duration = Duration::from_secs(60);

/// Parse the package name from a requirement string (e.g. "requests>=2.28" -> "requests").
/// Strips version specifiers (==, !=, >=, <=, ~=, <, >).
pub fn parse_requirement_name(requirement: &str) -> String {
    let s = requirement.trim();
    let specifiers = ["~=", "==", "!=", ">=", "<=", "<", ">"];
    let mut min_idx = s.len();
    for spec in &specifiers {
        if let Some(idx) = s.find(spec) {
            min_idx = min_idx.min(idx);
        }
    }
    s[..min_idx].trim().to_string()
}

/// Enterprise package manager with caching
#[allow(dead_code)]
pub struct PackageManager {
    /// Package cache
    cache: Arc<RwLock<ResultCache>>,
    /// Cache directory for large packages
    cache_dir: PathBuf,
    /// Maximum cache size in bytes
    max_cache_size: usize,
    /// Current cache size
    current_cache_size: Arc<RwLock<usize>>,
    /// Allowed hosts for package downloads (network whitelist)
    network_whitelist: HashSet<String>,
    /// If true, only whitelist hosts are allowed; if false, any public host is allowed
    strict_network_whitelist: bool,
    /// Number of package/dependency cache hits
    cache_hits: AtomicU64,
    /// Number of package/dependency cache misses
    cache_misses: AtomicU64,
}

impl PackageManager {
    /// Create a new package manager.
    ///
    /// `network_whitelist`: allowed hosts for downloads (e.g. `files.pythonhosted.org`, `*.pypi.org`).
    /// `strict_network_whitelist`: if true, only whitelist hosts are allowed; if false, any public host is allowed.
    pub fn new(
        cache: Arc<RwLock<ResultCache>>,
        cache_dir: PathBuf,
        max_cache_size_mb: usize,
        network_whitelist: Vec<String>,
        strict_network_whitelist: bool,
    ) -> anyhow::Result<Self> {
        let max_cache_size = max_cache_size_mb * 1024 * 1024;

        // Ensure cache directory exists
        fs::create_dir_all(&cache_dir)?;

        let network_whitelist: HashSet<String> = network_whitelist.into_iter().collect();

        let manager = Self {
            cache,
            cache_dir,
            max_cache_size,
            current_cache_size: Arc::new(RwLock::new(0)),
            network_whitelist,
            strict_network_whitelist,
            cache_hits: AtomicU64::new(0),
            cache_misses: AtomicU64::new(0),
        };

        // Initialize cache size
        manager.update_cache_size()?;

        Ok(manager)
    }

    /// Download and cache a package
    #[allow(dead_code)]
    pub async fn download_package(&self, package_name: &str, version: &str, download_url: &str) -> anyhow::Result<Vec<u8>> {
        // Check cache first
        let mut cache = self.cache.write().await;
        if let Some(cached_data) = cache.get_package(package_name, version) {
            self.cache_hits.fetch_add(1, Ordering::Relaxed);
            tracing::debug!("Package cache hit for {}@{}", package_name, version);
            return Ok(cached_data);
        }
        drop(cache);
        self.cache_misses.fetch_add(1, Ordering::Relaxed);

        // Download package
        tracing::info!("Downloading package {}@{} from {}", package_name, version, download_url);
        let package_data = self.download_from_url(download_url).await?;

        // Cache the package
        let mut cache = self.cache.write().await;
        cache.set_package(package_name, version, &package_data);

        // Also store in filesystem cache for large packages
        if package_data.len() > 1024 * 1024 { // > 1MB
            self.store_large_package(package_name, version, &package_data).await?;
        }

        Ok(package_data)
    }

    /// Resolve dependencies and cache the result
    #[allow(dead_code)]
    pub async fn resolve_dependencies(&self, requirements: &[String]) -> anyhow::Result<HashMap<String, String>> {
        let requirements_hash = ResultCache::hash_requirements(requirements);

        // Check cache first
        let mut cache = self.cache.write().await;
        if let Some(cached_resolution) = cache.get_dependency_resolution(&requirements_hash) {
            if let Ok(resolution) = serde_json::from_str(&cached_resolution) {
                self.cache_hits.fetch_add(1, Ordering::Relaxed);
                tracing::debug!("Dependency resolution cache hit");
                return Ok(resolution);
            }
        }
        drop(cache);
        self.cache_misses.fetch_add(1, Ordering::Relaxed);

        // Perform dependency resolution
        let resolution = self.perform_dependency_resolution(requirements).await?;

        // Cache the result
        let resolution_json = serde_json::to_string(&resolution)?;
        let mut cache = self.cache.write().await;
        cache.set_dependency_resolution(&requirements_hash, resolution_json);

        Ok(resolution)
    }

    /// Get cached bytecode for Python source
    #[allow(dead_code)]
    pub async fn get_cached_bytecode(&self, source_hash: &str) -> Option<Vec<u8>> {
        let mut cache = self.cache.write().await;
        cache.get_python_wasm(source_hash)
    }

    /// Cache Python bytecode
    #[allow(dead_code)]
    pub async fn set_cached_bytecode(&self, source_hash: &str, bytecode: &[u8]) {
        let mut cache = self.cache.write().await;
        cache.set_python_wasm(source_hash, bytecode);
    }

    /// Clean up old cache entries to stay within size limits
    pub async fn cleanup_cache(&self) -> anyhow::Result<()> {
        let current_size = *self.current_cache_size.read().await;

        if current_size > self.max_cache_size {
            tracing::info!("Package cache size {}MB exceeds limit {}MB, cleaning up",
                          current_size / (1024 * 1024), self.max_cache_size / (1024 * 1024));

            // Remove oldest files first
            self.remove_old_cache_files().await?;
            self.update_cache_size()?;
        }

        Ok(())
    }

    /// Download package from URL with network whitelist enforcement.
    async fn download_from_url(&self, url: &str) -> anyhow::Result<Vec<u8>> {
        if !fetch::is_network_request_allowed(
            url,
            &self.network_whitelist,
            self.strict_network_whitelist,
        ) {
            anyhow::bail!(
                "Package download denied: URL not allowed by network whitelist (host not in whitelist or private/loopback)"
            );
        }

        let client = reqwest::Client::builder()
            .timeout(PACKAGE_DOWNLOAD_TIMEOUT)
            .build()
            .context("build HTTP client for package download")?;

        let resp = client
            .get(url)
            .send()
            .await
            .context("package download request failed")?;

        let status = resp.status();
        if !status.is_success() {
            anyhow::bail!("Package download failed: HTTP {}", status);
        }

        let content_len = resp.content_length().unwrap_or(0) as usize;
        if content_len > MAX_PACKAGE_DOWNLOAD_BYTES {
            anyhow::bail!(
                "Package too large ({} bytes, max {} bytes)",
                content_len,
                MAX_PACKAGE_DOWNLOAD_BYTES
            );
        }

        let bytes = resp.bytes().await.context("read package response body")?;
        if bytes.len() > MAX_PACKAGE_DOWNLOAD_BYTES {
            anyhow::bail!(
                "Package too large ({} bytes, max {} bytes)",
                bytes.len(),
                MAX_PACKAGE_DOWNLOAD_BYTES
            );
        }

        Ok(bytes.to_vec())
    }

    /// Perform dependency resolution via PyPI JSON API.
    /// Each requirement is parsed to a package name (version specifiers stripped); then the latest
    /// version is fetched from https://pypi.org/pypi/<name>/json. Respects network whitelist.
    async fn perform_dependency_resolution(&self, requirements: &[String]) -> anyhow::Result<HashMap<String, String>> {
        let mut resolution = HashMap::new();
        for req in requirements {
            let name = parse_requirement_name(req);
            if name.is_empty() {
                continue;
            }
            let name_lower = name.to_lowercase();
            if resolution.contains_key(&name_lower) {
                continue;
            }
            let version = self.fetch_pypi_latest_version(&name).await?;
            resolution.insert(name_lower, version);
        }
        Ok(resolution)
    }

    /// Fetch latest version for a package from PyPI JSON API.
    async fn fetch_pypi_latest_version(&self, package_name: &str) -> anyhow::Result<String> {
        let url = format!("https://pypi.org/pypi/{}/json", package_name);
        let body = self.download_from_url(&url).await?;
        let json: serde_json::Value = serde_json::from_slice(&body).context("parse PyPI JSON")?;
        let version = json
            .get("info")
            .and_then(|info| info.get("version"))
            .and_then(|v| v.as_str())
            .ok_or_else(|| anyhow::anyhow!("PyPI response missing info.version for {}", package_name))?;
        Ok(version.to_string())
    }

    /// Store large package in filesystem cache
    async fn store_large_package(&self, package_name: &str, version: &str, data: &[u8]) -> anyhow::Result<()> {
        let filename = format!("{}-{}.pkg", package_name, version);
        let filepath = self.cache_dir.join(filename);

        tokio::fs::write(&filepath, data).await?;
        tracing::debug!("Stored large package {}@{} in filesystem cache", package_name, version);

        Ok(())
    }

    /// Remove old cache files to free up space
    async fn remove_old_cache_files(&self) -> anyhow::Result<()> {
        // Get all package files sorted by modification time
        let mut entries = Vec::new();
        let mut read_dir = tokio::fs::read_dir(&self.cache_dir).await?;

        while let Some(entry) = read_dir.next_entry().await? {
            if entry.file_type().await?.is_file() {
                let metadata = entry.metadata().await?;
                entries.push((entry.path(), metadata.modified()?));
            }
        }

        // Sort by modification time (oldest first)
        entries.sort_by_key(|(_, mtime)| *mtime);

        // Remove oldest files until we're under the limit
        let mut current_size = *self.current_cache_size.read().await;
        for (path, _) in entries {
            if current_size <= self.max_cache_size {
                break;
            }

            if let Ok(metadata) = tokio::fs::metadata(&path).await {
                let file_size = metadata.len() as usize;
                if tokio::fs::remove_file(&path).await.is_ok() {
                    current_size = current_size.saturating_sub(file_size);
                    tracing::debug!("Removed old cache file: {:?}", path);
                }
            }
        }

        // Update current size
        *self.current_cache_size.write().await = current_size;

        Ok(())
    }

    /// Update the current cache size
    fn update_cache_size(&self) -> anyhow::Result<()> {
        let mut total_size = 0usize;

        for entry in fs::read_dir(&self.cache_dir)? {
            let entry = entry?;
            if entry.file_type()?.is_file() {
                total_size += entry.metadata()?.len() as usize;
            }
        }

        // Update the shared current size
        let rt = tokio::runtime::Runtime::new()?;
        rt.block_on(async {
            *self.current_cache_size.write().await = total_size;
        });

        Ok(())
    }

    /// Get cache statistics
    pub async fn get_cache_stats(&self) -> PackageCacheStats {
        let cache = self.cache.read().await;
        let result_stats = cache.stats();
        let current_size = *self.current_cache_size.read().await;
        let hits = self.cache_hits.load(Ordering::Relaxed);
        let misses = self.cache_misses.load(Ordering::Relaxed);
        let total = hits + misses;
        let cache_hit_ratio = if total > 0 {
            hits as f64 / total as f64
        } else {
            0.0
        };

        PackageCacheStats {
            result_cache_entries: result_stats.entries,
            filesystem_cache_size_mb: current_size / (1024 * 1024),
            max_cache_size_mb: self.max_cache_size / (1024 * 1024),
            cache_hit_ratio,
        }
    }
}

/// Package cache statistics
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct PackageCacheStats {
    pub result_cache_entries: usize,
    pub filesystem_cache_size_mb: usize,
    pub max_cache_size_mb: usize,
    pub cache_hit_ratio: f64,
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::sync::Arc;
    use tokio::sync::RwLock;

    #[tokio::test]
    async fn test_package_manager_creation() {
        let cache = Arc::new(RwLock::new(ResultCache::new(3600)));
        let temp_dir = tempfile::tempdir().unwrap();
        let manager = PackageManager::new(
            cache,
            temp_dir.path().to_path_buf(),
            100,
            vec![],
            false,
        )
        .unwrap();
        assert!(manager.cache_dir.exists());
    }

    #[tokio::test]
    async fn test_cache_stats() {
        let cache = Arc::new(RwLock::new(ResultCache::new(3600)));
        let temp_dir = tempfile::tempdir().unwrap();
        let manager = PackageManager::new(
            cache,
            temp_dir.path().to_path_buf(),
            100,
            vec![],
            false,
        )
        .unwrap();

        let stats = manager.get_cache_stats().await;
        assert_eq!(stats.max_cache_size_mb, 100);
        assert_eq!(stats.result_cache_entries, 0);
    }
}
