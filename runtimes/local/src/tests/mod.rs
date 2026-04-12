//! Comprehensive test suite for the MicroVM Runtime
//!
//! This test module covers:
//! - AOT Cache (compilation, cache hit/miss, LRU eviction, disk persistence)
//! - Memory Limiter (memory growth denial, thread-local safety)
//! - Instance Pooling (acquire, prewarm, stats, concurrency limiting)
//! - Pool Manager (pool creation, per-function management)
//! - WASI State Snapshot (capture, restore)
//! - Execution (handler result reading)
//! - Configuration (validation, fuel computation)

#[cfg(test)]
mod tests {
    use std::sync::Arc;
    use std::time::Duration;

    use wasmtime::Module;

    // Engine modules (using public re-exports)
    use crate::engine::{AotCache};
    use crate::engine::{install_memory_limiter, FunctionMemoryLimiter};
    use crate::config::Config;

    // Pool modules (using public re-exports)
    use crate::pool::{WasmInstancePool, PoolManager, PooledWasmInstance, WasiStateSnapshot, WasmPoolStats};

    // Test utilities
    fn create_test_config() -> Config {
        Config {
            aot_cache_enabled: true,
            aot_cache_size_mb: 10, // Small cache for testing eviction
            aot_cache_dir: "".to_string(), // In-memory only for most tests
            function: "test".to_string(),
            version: "1.0.0".to_string(),
            wasi_env: vec![
                "TEST_VAR=test_value".to_string(),
                "FOO=bar".to_string(),
            ],
            max_output_bytes: 1024 * 1024,
            ..Config::default()
        }
    }

    // ==========================================================================
    // AOT Cache Tests
    // ==========================================================================

    mod aot_cache_tests {
        use super::*;

        #[test]
        fn test_aot_cache_new_creates_empty_cache() {
            let cache = AotCache::new();
            // Verify internal state via public methods
            let config = create_test_config();
            let engine = wasmtime::Engine::default();
            let wasm_bytes = create_minimal_wasm_module();

            // First compilation should miss cache
            let result1 = cache.get_or_compile_module(&engine, &wasm_bytes, &config);
            assert!(result1.is_ok(), "First compilation should succeed");

            // Second compilation should hit cache
            let result2 = cache.get_or_compile_module(&engine, &wasm_bytes, &config);
            assert!(result2.is_ok(), "Second compilation should succeed");
        }

        #[test]
        fn test_aot_cache_default_impl() {
            let cache: AotCache = Default::default();
            // Should be equivalent to new()
            let engine = wasmtime::Engine::default();
            let config = create_test_config();
            let wasm_bytes = create_minimal_wasm_module();

            let result = cache.get_or_compile_module(&engine, &wasm_bytes, &config);
            assert!(result.is_ok());
        }

        #[tokio::test]
        async fn test_aot_cache_compile_and_cache() {
            let cache = AotCache::new();
            let config = create_test_config();
            let engine = wasmtime::Engine::default();
            let wasm_bytes = create_minimal_wasm_module();

            let hash = "test_hash_12345";

            // Compile and cache
            let result = cache.compile_and_cache(&engine, &wasm_bytes, hash, &config);
            assert!(result.is_ok(), "Compilation should succeed");

            let compiled = result.unwrap();
            assert!(!compiled.is_empty(), "Compiled bytes should not be empty");
        }

        #[tokio::test]
        async fn test_aot_cache_memory_hit() {
            let cache = AotCache::new();
            let mut config = create_test_config();
            config.aot_cache_enabled = true;
            let engine = wasmtime::Engine::default();
            let wasm_bytes = create_minimal_wasm_module();

            // Compute hash the same way the cache does
            use std::collections::hash_map::DefaultHasher;
            use std::hash::{Hash, Hasher};
            let mut hasher = DefaultHasher::new();
            wasm_bytes.hash(&mut hasher);
            let hash = format!("{:016x}", hasher.finish());

            // First: compile and cache
            let _ = cache.compile_and_cache(&engine, &wasm_bytes, &hash, &config).unwrap();

            // Second: load from cache should hit
            let cached = cache.load_precompiled(&engine, &hash, &config);
            assert!(cached.is_ok(), "Cache load should succeed");
            assert!(cached.unwrap().is_some(), "Cache should have the module");
        }

        #[tokio::test]
        async fn test_aot_cache_miss_returns_none() {
            let cache = AotCache::new();
            let config = create_test_config();
            let engine = wasmtime::Engine::default();

            // Try to load a non-existent hash
            let result = cache.load_precompiled(&engine, "nonexistent_hash", &config);
            assert!(result.is_ok(), "Cache miss should not error");
            assert!(result.unwrap().is_none(), "Cache miss should return None");
        }

        #[tokio::test]
        async fn test_aot_cache_disabled_returns_none() {
            let cache = AotCache::new();
            let mut config = create_test_config();
            config.aot_cache_enabled = false;
            let engine = wasmtime::Engine::default();
            let wasm_bytes = create_minimal_wasm_module();

            // Compile with cache disabled
            let hash = "disabled_test";
            let _ = cache.compile_and_cache(&engine, &wasm_bytes, hash, &config);

            // Load should return None when disabled
            let cached = cache.load_precompiled(&engine, hash, &config);
            assert!(cached.is_ok());
            assert!(cached.unwrap().is_none());
        }

        #[tokio::test]
        async fn test_aot_cache_lru_eviction() {
            let cache = AotCache::new();
            let mut config = create_test_config();
            // Small cache: the serialized compiled modules are small,
            // so we set a very small limit to force eviction quickly
            config.aot_cache_size_mb = 0; // Zero limit means only 1 item fits (first eviction removes oldest)
            config.aot_cache_dir = "".to_string(); // In-memory only
            let engine = wasmtime::Engine::default();

            // Create three different minimal modules (different content = different hashes)
            let mut wasm1 = create_minimal_wasm_module();
            wasm1.push(0x00); // Add a custom section with just "1"
            wasm1.push(0x02); // section size
            wasm1.push(0x01); // name length
            wasm1.push(0x31); // '1'

            let mut wasm2 = create_minimal_wasm_module();
            wasm2.push(0x00); // Add a custom section with just "2"
            wasm2.push(0x02); // section size
            wasm2.push(0x01); // name length
            wasm2.push(0x32); // '2'

            let mut wasm3 = create_minimal_wasm_module();
            wasm3.push(0x00); // Add a custom section with just "3"
            wasm3.push(0x02); // section size
            wasm3.push(0x01); // name length
            wasm3.push(0x33); // '3'

            // Compile first module
            let hash1 = "module_1";
            cache.compile_and_cache(&engine, &wasm1, hash1, &config).unwrap();

            // Verify first is cached
            assert!(cache.load_precompiled(&engine, hash1, &config).unwrap().is_some());

            // Compile second module - should evict first (due to zero limit)
            let hash2 = "module_2";
            cache.compile_and_cache(&engine, &wasm2, hash2, &config).unwrap();

            // First should be evicted, second should remain
            assert!(cache.load_precompiled(&engine, hash1, &config).unwrap().is_none());
            assert!(cache.load_precompiled(&engine, hash2, &config).unwrap().is_some());

            // Compile third module - should evict second
            let hash3 = "module_3";
            cache.compile_and_cache(&engine, &wasm3, hash3, &config).unwrap();

            // Second should be evicted, third should remain
            assert!(cache.load_precompiled(&engine, hash2, &config).unwrap().is_none());
            assert!(cache.load_precompiled(&engine, hash3, &config).unwrap().is_some());
        }

        #[tokio::test]
        async fn test_aot_cache_disk_persistence() {
            let temp_dir = tempfile::tempdir().unwrap();
            let cache = AotCache::new();
            let mut config = create_test_config();
            config.aot_cache_enabled = true;
            config.aot_cache_dir = temp_dir.path().to_str().unwrap().to_string();
            let engine = wasmtime::Engine::default();
            let wasm_bytes = create_minimal_wasm_module();

            // Compute hash
            use std::collections::hash_map::DefaultHasher;
            use std::hash::{Hash, Hasher};
            let mut hasher = DefaultHasher::new();
            wasm_bytes.hash(&mut hasher);
            let hash = format!("{:016x}", hasher.finish());

            // Compile and cache (should write to disk)
            cache.compile_and_cache(&engine, &wasm_bytes, &hash, &config).unwrap();

            // Verify file exists
            let cache_file = temp_dir.path().join(format!("{}.cwasm", hash));
            assert!(cache_file.exists(), "Cache file should be written to disk");

            // Clear memory cache by creating new cache instance
            let cache2 = AotCache::new();

            // Load should hit disk cache
            let cached = cache2.load_precompiled(&engine, &hash, &config);
            assert!(cached.is_ok());
            assert!(cached.unwrap().is_some(), "Should load from disk cache");
        }

        #[test]
        fn test_get_or_compile_module_flow() {
            let cache = AotCache::new();
            let config = create_test_config();
            let engine = wasmtime::Engine::default();
            let wasm_bytes = create_minimal_wasm_module();

            // First call should compile
            let result1 = cache.get_or_compile_module(&engine, &wasm_bytes, &config);
            assert!(result1.is_ok());

            // Second call with same bytes should use cache
            let result2 = cache.get_or_compile_module(&engine, &wasm_bytes, &config);
            assert!(result2.is_ok());
        }
    }

    // ==========================================================================
    // Memory Limiter Tests
    // ==========================================================================

    mod memory_limiter_tests {
        use super::*;
        use wasmtime::ResourceLimiter;

        #[test]
        fn test_memory_limiter_new() {
            let mut limiter = FunctionMemoryLimiter::new(128); // 128 MB
            // Test via the ResourceLimiter trait

            // Should allow growth within limit
            let allowed = limiter.memory_growing(0, 64 * 1024 * 1024, None);
            assert!(allowed.is_ok());
            assert!(allowed.unwrap(), "Should allow growth within limit");
        }

        #[test]
        fn test_memory_limiter_denies_excessive_growth() {
            let mut limiter = FunctionMemoryLimiter::new(1); // 1 MB

            // Should deny growth exceeding limit
            let allowed = limiter.memory_growing(0, 2 * 1024 * 1024, None); // 2 MB desired
            assert!(allowed.is_ok());
            assert!(!allowed.unwrap(), "Should deny growth exceeding limit");
        }

        #[test]
        fn test_memory_limiter_allows_exact_limit() {
            let mut limiter = FunctionMemoryLimiter::new(10); // 10 MB

            // Should allow growth exactly at limit
            let allowed = limiter.memory_growing(0, 10 * 1024 * 1024, None);
            assert!(allowed.is_ok());
            assert!(allowed.unwrap(), "Should allow growth at exact limit");
        }

        #[test]
        fn test_memory_limiter_table_growing_always_allowed() {
            let mut limiter = FunctionMemoryLimiter::new(1);

            // Table growing should always be allowed (not limited)
            let allowed = limiter.table_growing(0, 1000, None);
            assert!(allowed.is_ok());
            assert!(allowed.unwrap(), "Table growing should always be allowed");
        }

        #[test]
        fn test_install_memory_limiter_guard() {
            // Install limiter for current thread
            let guard = install_memory_limiter(64);

            // Guard should clear limiter on drop
            drop(guard);

            // After drop, limiter should be cleared (accessing it would panic)
            // We can't directly test this, but we verify the guard drops cleanly
        }

        #[test]
        fn test_limiter_guard_clears_on_drop() {
            // Create multiple guards to verify they don't interfere
            let guard1 = install_memory_limiter(32);
            drop(guard1);

            let guard2 = install_memory_limiter(64);
            drop(guard2);

            // Both guards should have cleaned up properly
        }

        #[test]
        fn test_memory_limiter_different_sizes() {
            // Test various memory limits
            let limits = vec![1, 10, 64, 128, 512, 1024];

            for limit_mb in limits {
                let mut limiter = FunctionMemoryLimiter::new(limit_mb);
                let limit_bytes = limit_mb as usize * 1024 * 1024;

                // Should allow up to limit
                let allowed = limiter.memory_growing(0, limit_bytes, None);
                assert!(allowed.is_ok());
                assert!(allowed.unwrap(), "Should allow up to limit for {} MB", limit_mb);

                // Should deny beyond limit
                let denied = limiter.memory_growing(0, limit_bytes + 1024, None);
                assert!(denied.is_ok());
                assert!(!denied.unwrap(), "Should deny beyond limit for {} MB", limit_mb);
            }
        }
    }

    // ==========================================================================
    // Pool Manager Tests
    // ==========================================================================

    mod pool_manager_tests {
        use super::*;

        #[tokio::test]
        async fn test_pool_manager_new() {
            let manager = PoolManager::new(10, 4);

            // Verify initial state via stats
            let stats = manager.stats().await;
            assert!(stats.is_empty(), "New manager should have no pools");
        }

        #[tokio::test]
        async fn test_pool_manager_get_or_create_pool() {
            let manager = PoolManager::new(5, 2);

            // Get pool for function
            let pool1 = manager.get_or_create_pool("func@1.0.0").await;
            let pool2 = manager.get_or_create_pool("func@1.0.0").await;

            // Should be the same pool (Arc equality)
            assert!(Arc::ptr_eq(&pool1, &pool2), "Should return same pool for same function");

            // Different function should get different pool
            let pool3 = manager.get_or_create_pool("other@1.0.0").await;
            assert!(!Arc::ptr_eq(&pool1, &pool3), "Different functions should have different pools");
        }

        #[tokio::test]
        async fn test_pool_manager_acquire_no_prewarm() {
            let manager = PoolManager::new(5, 2);

            // Acquire without pre-warming should fail
            let result = manager.acquire("unwarmed@1.0.0").await;
            assert!(result.is_err(), "Acquire should fail without pre-warmed instances");
        }

        #[tokio::test]
        async fn test_pool_manager_prewarm_and_acquire() {
            let manager = PoolManager::new(5, 2);
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);
            let wasi_ctx = create_test_wasi_ctx();

            let function_key = "test_func@1.0.0";

            // Pre-warm the pool
            manager.prewarm_instance(function_key, module, wasi_ctx).await;

            // Now acquire should succeed
            let guard = manager.acquire(function_key).await;
            assert!(guard.is_ok(), "Acquire should succeed after pre-warming");

            // Drop guard to return instance
            drop(guard);
        }

        #[tokio::test]
        async fn test_pool_manager_stats() {
            let manager = PoolManager::new(5, 2);
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);

            // Pre-warm two functions (need separate wasi_ctx for each)
            manager.prewarm_instance("func1@1.0.0", module.clone(), create_test_wasi_ctx()).await;
            manager.prewarm_instance("func2@1.0.0", module, create_test_wasi_ctx()).await;

            let stats = manager.stats().await;
            assert_eq!(stats.len(), 2, "Should have stats for 2 pools");
        }

        #[tokio::test]
        async fn test_pool_manager_pool_stats() {
            let manager = PoolManager::new(5, 2);
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);
            let wasi_ctx = create_test_wasi_ctx();

            let function_key = "test@1.0.0";
            manager.prewarm_instance(function_key, module, wasi_ctx).await;

            let stats = manager.pool_stats(function_key).await;
            assert!(stats.is_some(), "Should have stats for warmed function");

            let stats = stats.unwrap();
            assert_eq!(stats.function_key, function_key);
            assert_eq!(stats.idle_count, 1);
        }

        #[tokio::test]
        async fn test_pool_manager_is_warmed() {
            let manager = PoolManager::new(5, 2);
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);
            let wasi_ctx = create_test_wasi_ctx();

            let function_key = "warmed@1.0.0";

            // Initially not warmed
            assert!(!manager.is_warmed(function_key).await);

            // Pre-warm
            manager.prewarm_instance(function_key, module, wasi_ctx).await;

            // Now warmed
            assert!(manager.is_warmed(function_key).await);
        }

        #[tokio::test]
        async fn test_pool_manager_warmed_function_count() {
            let manager = PoolManager::new(5, 2);
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);

            // Initially 0
            assert_eq!(manager.warmed_function_count().await, 0);

            // Warm 3 functions (create new wasi_ctx for each)
            for i in 0..3 {
                manager.prewarm_instance(&format!("func{}@1.0.0", i), module.clone(), create_test_wasi_ctx()).await;
            }

            assert_eq!(manager.warmed_function_count().await, 3);
        }

        #[tokio::test]
        async fn test_pool_manager_clear() {
            let manager = PoolManager::new(5, 2);
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);

            // Pre-warm
            manager.prewarm_instance("test@1.0.0", module, create_test_wasi_ctx()).await;
            assert_eq!(manager.warmed_function_count().await, 1);

            // Clear
            manager.clear().await;

            // All pools removed
            assert_eq!(manager.warmed_function_count().await, 0);
        }

        #[tokio::test]
        async fn test_pool_manager_remove_pool() {
            let manager = PoolManager::new(5, 2);
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);

            // Pre-warm two functions (create separate wasi_ctx for each)
            manager.prewarm_instance("keep@1.0.0", module.clone(), create_test_wasi_ctx()).await;
            manager.prewarm_instance("remove@1.0.0", module, create_test_wasi_ctx()).await;

            assert_eq!(manager.warmed_function_count().await, 2);

            // Remove one
            manager.remove_pool("remove@1.0.0").await;

            assert_eq!(manager.warmed_function_count().await, 1);
            assert!(manager.is_warmed("keep@1.0.0").await);
            assert!(!manager.is_warmed("remove@1.0.0").await);
        }

        #[tokio::test]
        async fn test_pool_manager_with_pipe_capacity() {
            let manager = PoolManager::new(5, 2)
                .with_pipe_capacity(512 * 1024); // 512KB

            // Just verify it creates without error
            let _pool = manager.get_or_create_pool("test@1.0.0").await;
        }

        #[tokio::test]
        async fn test_pool_manager_concurrent_acquire() {
            let manager = Arc::new(PoolManager::new(2, 2)); // Max 2 concurrent, max 2 idle
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);

            let function_key = "concurrent@1.0.0";

            // Pre-warm 2 instances (each needs separate wasi_ctx)
            manager.prewarm_instance(function_key, module.clone(), create_test_wasi_ctx()).await;
            manager.prewarm_instance(function_key, module, create_test_wasi_ctx()).await;

            // Acquire both
            let guard1 = manager.acquire(function_key).await.unwrap();
            let guard2 = manager.acquire(function_key).await.unwrap();

            // Third acquire should block (but we can't test that without timeout)
            // Instead, just verify we have both guards
            assert_eq!(guard1.instance().reuse_count, 0);
            assert_eq!(guard2.instance().reuse_count, 0);

            drop(guard1);
            drop(guard2);
        }
    }

    // ==========================================================================
    // Wasm Instance Pool Tests
    // ==========================================================================

    mod wasm_instance_pool_tests {
        use super::*;

        #[test]
        fn test_wasm_instance_pool_new() {
            let pool = WasmInstancePool::new("test@1.0.0".to_string(), 10, 4);

            // Verify function key
            assert_eq!(pool.function_key(), "test@1.0.0");
        }

        #[test]
        fn test_wasm_instance_pool_new_ensures_min_max_concurrent() {
            // Zero max_concurrent should be normalized to 1
            let pool = WasmInstancePool::new("test@1.0.0".to_string(), 0, 4);

            // Can't directly verify, but test it doesn't panic
            assert_eq!(pool.function_key(), "test@1.0.0");
        }

        #[test]
        fn test_wasm_instance_pool_for_function() {
            let pool = WasmInstancePool::new("original@1.0.0".to_string(), 10, 4);
            let new_pool = pool.for_function("new@1.0.0".to_string());

            assert_eq!(new_pool.function_key(), "new@1.0.0");
        }

        #[tokio::test]
        async fn test_wasm_instance_pool_stats_empty() {
            let pool = WasmInstancePool::new("test@1.0.0".to_string(), 10, 4);
            let stats = pool.stats().await;

            assert_eq!(stats.function_key, "test@1.0.0");
            assert_eq!(stats.idle_count, 0);
            assert_eq!(stats.max_idle, 4);
            assert_eq!(stats.max_concurrent, 10);
            assert_eq!(stats.available_permits, 10);
        }

        #[tokio::test]
        async fn test_wasm_instance_pool_prewarm() {
            let pool = Arc::new(WasmInstancePool::new("test@1.0.0".to_string(), 10, 4));
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);
            let wasi_ctx = create_test_wasi_ctx();

            let instance = PooledWasmInstance::new(module, wasi_ctx, "test@1.0.0".to_string(), 1024 * 1024);

            pool.prewarm(instance).await;

            let stats = pool.stats().await;
            assert_eq!(stats.idle_count, 1);
        }

        #[tokio::test]
        async fn test_wasm_instance_pool_prewarm_with() {
            let pool = Arc::new(WasmInstancePool::new("test@1.0.0".to_string(), 10, 4));
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);
            let wasi_ctx = create_test_wasi_ctx();

            pool.prewarm_with(module, wasi_ctx, 1024 * 1024).await;

            let stats = pool.stats().await;
            assert_eq!(stats.idle_count, 1);
        }

        #[tokio::test]
        async fn test_wasm_instance_pool_prewarm_respects_max_idle() {
            let pool = Arc::new(WasmInstancePool::new("test@1.0.0".to_string(), 10, 2));
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);

            // Try to pre-warm more than max_idle (create new wasi_ctx for each)
            for _i in 0..5 {
                let inst = PooledWasmInstance::new(
                    module.clone(),
                    create_test_wasi_ctx(),
                    "test@1.0.0".to_string(),
                    1024 * 1024
                );
                pool.prewarm(inst).await;
            }

            let stats = pool.stats().await;
            assert_eq!(stats.idle_count, 2, "Should not exceed max_idle");
        }

        #[tokio::test]
        async fn test_wasm_instance_pool_acquire_and_return() {
            let pool = Arc::new(WasmInstancePool::new("test@1.0.0".to_string(), 10, 4));
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);
            let wasi_ctx = create_test_wasi_ctx();

            // Pre-warm
            pool.prewarm_with(module, wasi_ctx, 1024 * 1024).await;

            // Acquire
            let guard = WasmInstancePool::acquire(pool.clone()).await;
            assert!(guard.is_ok());

            // Stats during acquisition
            let stats = pool.stats().await;
            assert_eq!(stats.idle_count, 0, "Instance should be checked out");
            assert_eq!(stats.available_permits, 9, "One permit should be used");

            // Drop guard - instance should return to pool
            drop(guard.unwrap());

            // Give a moment for the drop to complete
            tokio::time::sleep(Duration::from_millis(10)).await;

            let stats = pool.stats().await;
            assert_eq!(stats.idle_count, 1, "Instance should be back in pool");
        }

        #[tokio::test]
        async fn test_wasm_instance_pool_acquire_no_idle() {
            let pool = Arc::new(WasmInstancePool::new("test@1.0.0".to_string(), 10, 4));

            // Try to acquire without pre-warming
            let result = WasmInstancePool::acquire(pool.clone()).await;
            assert!(result.is_err(), "Should fail without pre-warmed instances");
        }

        #[tokio::test]
        async fn test_wasm_instance_pool_is_warmed() {
            let pool = Arc::new(WasmInstancePool::new("test@1.0.0".to_string(), 10, 4));

            assert!(!pool.is_warmed().await, "Empty pool should not be warmed");

            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);
            let wasi_ctx = create_test_wasi_ctx();

            pool.prewarm_with(module, wasi_ctx, 1024 * 1024).await;

            assert!(pool.is_warmed().await, "Pool with instance should be warmed");
        }

        #[tokio::test]
        async fn test_wasm_instance_pool_blocking_stats() {
            let pool = WasmInstancePool::new("test@1.0.0".to_string(), 10, 4);

            let stats = pool.blocking_stats();
            assert_eq!(stats.function_key, "test@1.0.0");
            assert_eq!(stats.idle_count, 0);
            assert_eq!(stats.max_idle, 4);
        }

        #[tokio::test]
        async fn test_wasm_instance_pool_is_warmed_blocking() {
            let pool = WasmInstancePool::new("test@1.0.0".to_string(), 10, 4);

            // Empty pool
            assert!(!pool.is_warmed_blocking());

            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);
            let wasi_ctx = create_test_wasi_ctx();

            pool.prewarm_with(module, wasi_ctx, 1024 * 1024).await;

            assert!(pool.is_warmed_blocking());
        }

        #[test]
        fn test_wasm_instance_pool_send_sync() {
            // Verify Send + Sync traits
            fn assert_send_sync<T: Send + Sync>() {}
            assert_send_sync::<WasmInstancePool>();
        }
    }

    // ==========================================================================
    // Pooled Wasm Instance Tests
    // ==========================================================================

    mod pooled_wasm_instance_tests {
        use super::*;

        #[test]
        fn test_pooled_wasm_instance_new() {
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);
            let wasi_ctx = create_test_wasi_ctx();

            let instance = PooledWasmInstance::new(
                module,
                wasi_ctx,
                "test@1.0.0".to_string(),
                1024 * 1024
            );

            assert_eq!(instance.function_key, "test@1.0.0");
            assert_eq!(instance.reuse_count, 0);
            assert!(instance.memory_estimate > 0);
        }

        #[test]
        fn test_pooled_wasm_instance_create_store() {
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);
            let wasi_ctx = create_test_wasi_ctx();

            let instance = PooledWasmInstance::new(
                module,
                wasi_ctx,
                "test@1.0.0".to_string(),
                1024 * 1024
            );

            let store = instance.create_store(&engine);
            // Store should be valid (just verify it doesn't panic)
            drop(store);
        }

        #[test]
        fn test_pooled_wasm_instance_reset_for_execution() {
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);
            let wasi_ctx = create_test_wasi_ctx();

            let mut instance = PooledWasmInstance::new(
                module,
                wasi_ctx,
                "test@1.0.0".to_string(),
                1024 * 1024
            );

            assert_eq!(instance.reuse_count, 0);

            // Reset with input
            instance.reset_for_execution("test input data");

            // Reuse count should increment
            assert_eq!(instance.reuse_count, 1);
        }

        #[test]
        fn test_pooled_wasm_instance_multiple_resets() {
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);
            let wasi_ctx = create_test_wasi_ctx();

            let mut instance = PooledWasmInstance::new(
                module,
                wasi_ctx,
                "test@1.0.0".to_string(),
                1024 * 1024
            );

            // Multiple resets
            for i in 1..=5 {
                instance.reset_for_execution(&format!("input {}", i));
                assert_eq!(instance.reuse_count, i);
            }
        }

        #[test]
        fn test_pooled_wasm_instance_send() {
            // Verify Send trait (required for pool usage)
            fn assert_send<T: Send>() {}
            assert_send::<PooledWasmInstance>();
        }

        #[tokio::test]
        async fn test_pooled_wasm_instance_guard_basic() {
            let pool = Arc::new(WasmInstancePool::new("test@1.0.0".to_string(), 10, 4));
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);
            let wasi_ctx = create_test_wasi_ctx();

            pool.prewarm_with(module, wasi_ctx, 1024 * 1024).await;

            let mut guard = WasmInstancePool::acquire(pool.clone()).await.unwrap();

            // Access instance
            let _instance_ref = guard.instance();

            // Mutable access
            let _instance_mut = guard.instance_mut();
        }

        #[tokio::test]
        async fn test_pooled_wasm_instance_guard_mark_dirty() {
            let pool = Arc::new(WasmInstancePool::new("test@1.0.0".to_string(), 10, 4));
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);
            let wasi_ctx = create_test_wasi_ctx();

            pool.prewarm_with(module, wasi_ctx, 1024 * 1024).await;

            let mut guard = WasmInstancePool::acquire(pool.clone()).await.unwrap();
            guard.mark_dirty();

            // Drop dirty guard - instance should NOT return to pool
            drop(guard);

            tokio::time::sleep(Duration::from_millis(10)).await;

            let stats = pool.stats().await;
            assert_eq!(stats.idle_count, 0, "Dirty instance should not return to pool");
        }

        #[tokio::test]
        async fn test_pooled_wasm_instance_guard_take() {
            let pool = Arc::new(WasmInstancePool::new("test@1.0.0".to_string(), 10, 4));
            let engine = wasmtime::Engine::default();
            let module = create_test_wasm_module(&engine);
            let wasi_ctx = create_test_wasi_ctx();

            pool.prewarm_with(module, wasi_ctx, 1024 * 1024).await;

            let guard = WasmInstancePool::acquire(pool.clone()).await.unwrap();

            // Take ownership
            let instance = guard.take();
            assert_eq!(instance.function_key, "test@1.0.0");

            // Pool should be empty (instance taken, not returned)
            let stats = pool.stats().await;
            assert_eq!(stats.idle_count, 0);
        }
    }

    // ==========================================================================
    // Wasi State Snapshot Tests
    // ==========================================================================

    mod wasi_state_snapshot_tests {
        use super::*;

        #[test]
        fn test_wasi_state_snapshot_default() {
            let snapshot = WasiStateSnapshot::default();
            assert!(snapshot.env_vars.is_empty());
            assert!(snapshot.args.is_empty());
            assert_eq!(snapshot.pipe_capacity, 0);
        }

        #[test]
        fn test_wasi_state_snapshot_capture_from_config() {
            let config = create_test_config();
            let snapshot = WasiStateSnapshot::capture_from_config(&config);

            // Should capture env vars from config
            assert!(!snapshot.env_vars.is_empty());

            // Should have captured TEST_VAR
            let test_var = snapshot.env_vars.iter()
                .find(|(k, _)| k == "TEST_VAR");
            assert!(test_var.is_some());
            assert_eq!(test_var.unwrap().1, "test_value");

            // Should have default PATH
            let path_var = snapshot.env_vars.iter()
                .find(|(k, _)| k == "PATH");
            assert!(path_var.is_some());

            // Should have function as arg
            assert!(!snapshot.args.is_empty());
            assert_eq!(snapshot.args[0], "test");

            // Pipe capacity should match max_output_bytes
            assert_eq!(snapshot.pipe_capacity, config.max_output_bytes);
        }

        #[test]
        fn test_wasi_state_snapshot_capture_custom_max_output() {
            let mut config = create_test_config();
            config.max_output_bytes = 512 * 1024; // 512KB

            let snapshot = WasiStateSnapshot::capture_from_config(&config);
            assert_eq!(snapshot.pipe_capacity, 512 * 1024);
        }

        #[test]
        fn test_wasi_state_snapshot_capture_zero_max_output() {
            let mut config = create_test_config();
            config.max_output_bytes = 0;

            let snapshot = WasiStateSnapshot::capture_from_config(&config);
            // Should use default 1 MiB
            assert_eq!(snapshot.pipe_capacity, 1024 * 1024);
        }

        #[test]
        fn test_wasi_state_snapshot_restore() {
            let mut builder = wasmtime_wasi::WasiCtxBuilder::new();

            let snapshot = WasiStateSnapshot {
                env_vars: vec![
                    ("KEY1".to_string(), "value1".to_string()),
                    ("KEY2".to_string(), "value2".to_string()),
                ],
                args: vec!["arg1".to_string(), "arg2".to_string()],
                pipe_capacity: 1024,
            };

            // Restore should not panic
            snapshot.restore(&mut builder);

            // Build to verify it worked
            let _ctx = builder.build_p1();
        }

        #[test]
        fn test_wasi_state_snapshot_clone() {
            let snapshot = WasiStateSnapshot {
                env_vars: vec![("KEY".to_string(), "value".to_string())],
                args: vec!["arg".to_string()],
                pipe_capacity: 1024,
            };

            let cloned = snapshot.clone();
            assert_eq!(cloned.env_vars, snapshot.env_vars);
            assert_eq!(cloned.args, snapshot.args);
            assert_eq!(cloned.pipe_capacity, snapshot.pipe_capacity);
        }

        #[test]
        fn test_wasi_state_snapshot_debug() {
            let snapshot = WasiStateSnapshot {
                env_vars: vec![("KEY".to_string(), "value".to_string())],
                args: vec!["arg".to_string()],
                pipe_capacity: 1024,
            };

            let debug_str = format!("{:?}", snapshot);
            assert!(debug_str.contains("WasiStateSnapshot"));
        }
    }

    // ==========================================================================
    // Wasm Pool Stats Tests
    // ==========================================================================

    mod wasm_pool_stats_tests {
        use super::*;

        #[test]
        fn test_wasm_pool_stats_display() {
            let stats = WasmPoolStats {
                function_key: "test@1.0.0".to_string(),
                idle_count: 2,
                max_idle: 4,
                available_permits: 8,
                max_concurrent: 10,
            };

            let display_str = format!("{}", stats);
            assert!(display_str.contains("test@1.0.0"));
            assert!(display_str.contains("idle=2/4"));
            assert!(display_str.contains("permits=8/10"));
        }

        #[test]
        fn test_wasm_pool_stats_clone() {
            let stats = WasmPoolStats {
                function_key: "test@1.0.0".to_string(),
                idle_count: 2,
                max_idle: 4,
                available_permits: 8,
                max_concurrent: 10,
            };

            let cloned = stats.clone();
            assert_eq!(cloned.function_key, stats.function_key);
            assert_eq!(cloned.idle_count, stats.idle_count);
            assert_eq!(cloned.max_idle, stats.max_idle);
            assert_eq!(cloned.available_permits, stats.available_permits);
            assert_eq!(cloned.max_concurrent, stats.max_concurrent);
        }

        #[test]
        fn test_wasm_pool_stats_debug() {
            let stats = WasmPoolStats {
                function_key: "test@1.0.0".to_string(),
                idle_count: 2,
                max_idle: 4,
                available_permits: 8,
                max_concurrent: 10,
            };

            let debug_str = format!("{:?}", stats);
            assert!(debug_str.contains("WasmPoolStats"));
            assert!(debug_str.contains("test@1.0.0"));
        }

        #[test]
        fn test_wasm_pool_stats_zero_values() {
            let stats = WasmPoolStats {
                function_key: "empty@1.0.0".to_string(),
                idle_count: 0,
                max_idle: 0,
                available_permits: 0,
                max_concurrent: 1,
            };

            let display_str = format!("{}", stats);
            assert!(display_str.contains("idle=0/0"));
            assert!(display_str.contains("permits=0/1"));
        }
    }

    // ==========================================================================
    // Config Tests
    // ==========================================================================

    mod config_tests {
        use super::*;
        use crate::budget::BudgetTier;

        #[test]
        fn test_config_function_key() {
            let config = Config {
                function: "myfunc".to_string(),
                version: "2.1.0".to_string(),
                ..Config::default()
            };

            assert_eq!(config.function_key(), "myfunc@2.1.0");
        }

        #[test]
        fn test_config_default() {
            let config = Config::default();

            // Verify key defaults
            assert_eq!(config.port, 8787);
            assert_eq!(config.function, "function");
            assert_eq!(config.version, "1.0.0");
            assert_eq!(config.memory_mb, 128);
            assert_eq!(config.timeout_ms, 5000);
            assert!(config.aot_cache_enabled);
            assert_eq!(config.aot_cache_size_mb, 512);
            assert_eq!(config.wasm_pool_max_concurrent, 10);
            assert_eq!(config.wasm_pool_max_idle, 4);
        }

        #[test]
        fn test_config_get_budget_tier() {
            let ultra_low = Config { tier: "ultra-low".to_string(), ..Config::default() };
            assert!(matches!(ultra_low.get_budget_tier(), BudgetTier::UltraLow));

            let low = Config { tier: "low".to_string(), ..Config::default() };
            assert!(matches!(low.get_budget_tier(), BudgetTier::Low));

            let medium = Config { tier: "medium".to_string(), ..Config::default() };
            assert!(matches!(medium.get_budget_tier(), BudgetTier::Medium));

            let high = Config { tier: "high".to_string(), ..Config::default() };
            assert!(matches!(high.get_budget_tier(), BudgetTier::High));

            // Unknown tier defaults to UltraLow
            let unknown = Config { tier: "unknown".to_string(), ..Config::default() };
            assert!(matches!(unknown.get_budget_tier(), BudgetTier::UltraLow));
        }

        #[test]
        fn test_config_get_budget_tier_case_insensitive() {
            let ultra = Config { tier: "ULTRA-LOW".to_string(), ..Config::default() };
            assert!(matches!(ultra.get_budget_tier(), BudgetTier::UltraLow));

            let low = Config { tier: "LOW".to_string(), ..Config::default() };
            assert!(matches!(low.get_budget_tier(), BudgetTier::Low));
        }

        #[test]
        fn test_config_validate_success() {
            let config = Config::default();
            assert!(config.validate().is_ok());
        }

        #[test]
        fn test_config_validate_enterprise_requires_medium_tier() {
            let low = Config {
                enterprise_enabled: true,
                tier: "low".to_string(),
                ..Config::default()
            };
            assert!(low.validate().is_err());

            let ultra = Config {
                enterprise_enabled: true,
                tier: "ultra-low".to_string(),
                ..Config::default()
            };
            assert!(ultra.validate().is_err());
        }

        #[test]
        fn test_config_validate_enterprise_allowed_medium_high() {
            let medium = Config {
                enterprise_enabled: true,
                tier: "medium".to_string(),
                ..Config::default()
            };
            assert!(medium.validate().is_ok());

            let high = Config {
                enterprise_enabled: true,
                tier: "high".to_string(),
                ..Config::default()
            };
            assert!(high.validate().is_ok());
        }

        #[test]
        fn test_config_validate_memory_exceeds_tier() {
            // Low tier max memory is small
            let low = Config {
                tier: "low".to_string(),
                memory_mb: 10000, // 10GB
                ..Config::default()
            };
            assert!(low.validate().is_err());
        }

        #[test]
        fn test_config_supports_microvm() {
            let non_enterprise = Config::default();
            assert!(!non_enterprise.supports_microvm());

            let enterprise_low = Config {
                enterprise_enabled: true,
                tier: "low".to_string(),
                ..Config::default()
            };
            assert!(!enterprise_low.supports_microvm());

            let enterprise_medium = Config {
                enterprise_enabled: true,
                tier: "medium".to_string(),
                ..Config::default()
            };
            assert!(enterprise_medium.supports_microvm());

            let enterprise_high = Config {
                enterprise_enabled: true,
                tier: "high".to_string(),
                ..Config::default()
            };
            assert!(enterprise_high.supports_microvm());
        }

        #[test]
        fn test_config_supports_cpython_wasm() {
            let disabled = Config::default();
            assert!(!disabled.supports_cpython_wasm());

            let enabled_no_path = Config {
                use_cpython_wasm: true,
                cpython_wasm_path: "".to_string(),
                ..Config::default()
            };
            assert!(!enabled_no_path.supports_cpython_wasm());

            let enabled = Config {
                use_cpython_wasm: true,
                cpython_wasm_path: "/path/to/cpython.wasm".to_string(),
                ..Config::default()
            };
            assert!(enabled.supports_cpython_wasm());
        }

        #[test]
        fn test_config_fuel_for_timeout() {
            let config = Config {
                timeout_ms: 1000,
                fuel_per_ms: 10000,
                ..Config::default()
            };

            assert_eq!(config.fuel_for_timeout(), 10_000_000);
        }

        #[test]
        fn test_config_fuel_for_timeout_saturating() {
            let config = Config {
                timeout_ms: u64::MAX,
                fuel_per_ms: u64::MAX,
                ..Config::default()
            };

            // Should saturate, not overflow
            let fuel = config.fuel_for_timeout();
            assert_eq!(fuel, u64::MAX);
        }
    }

    // ==========================================================================
    // Helper Functions
    // ==========================================================================

    /// Create a minimal valid WASM module (empty with just a start function)
    fn create_minimal_wasm_module() -> Vec<u8> {
        // Minimal WASM: (module (func (export "add")))
        // Binary representation matching micropython/loader.rs test module:
        vec![
            0x00, 0x61, 0x73, 0x6d, // magic
            0x01, 0x00, 0x00, 0x00, // version
            0x01, 0x04, 0x01, 0x60, 0x00, 0x00, // type section: 1 type, () -> ()
            0x03, 0x02, 0x01, 0x00, // func section: 1 function, type index 0
            0x07, 0x07, 0x01, 0x03, 0x61, 0x64, 0x64, 0x00, 0x00, // export "add" (3 chars = 0x03) as func 0
            0x0a, 0x04, 0x01, 0x02, 0x00, 0x0b, // code section: 1 function, empty body
        ]
    }

    /// Create a WASM module with specific approximate size
    fn create_wasm_module_with_size(target_bytes: usize) -> Vec<u8> {
        // Base minimal module is ~37 bytes
        // To create different-sized modules, we add custom sections with padding
        let base = create_minimal_wasm_module();
        let padding_needed = target_bytes.saturating_sub(base.len());

        if padding_needed == 0 {
            return base;
        }

        // Create multiple custom sections to reach target size
        // Each custom section: id=0, size, name_len, name, data...
        let mut module = base;
        let custom_name = "p";
        let name_len = 1;
        let section_overhead = 3 + name_len; // id + size_byte + name_len + name
        let data_per_section = 65535usize; // Max single-byte LEB128 section size

        let mut remaining = padding_needed;

        while remaining > 0 {
            let data_len = std::cmp::min(remaining.saturating_sub(section_overhead), data_per_section);
            if data_len == 0 {
                break;
            }

            module.push(0x00); // custom section id
            module.push((name_len + data_len + 1) as u8); // section size
            module.push(name_len as u8); // name length
            module.extend_from_slice(custom_name.as_bytes()); // name
            module.extend(vec![0x00; data_len]); // padding data

            remaining = remaining.saturating_sub(section_overhead + data_len);
        }

        module
    }

    /// Create a test WASM module using wasmtime
    fn create_test_wasm_module(engine: &wasmtime::Engine) -> wasmtime::Module {
        let wasm_bytes = create_minimal_wasm_module();
        Module::new(engine, &wasm_bytes).expect("Should compile valid WASM")
    }

    /// Create a basic WASI context for testing
    fn create_test_wasi_ctx() -> wasmtime_wasi::p1::WasiP1Ctx {
        let mut builder = wasmtime_wasi::WasiCtxBuilder::new();
        builder.build_p1()
    }
}
