use criterion::{criterion_group, criterion_main, Criterion};
use statefabric::cache::{LruCache, RedisCache, RedisConfig};
use statefabric::models::{Event, SourceType};
use statefabric::state::StateManager;
use std::hint::black_box;
use std::sync::Arc;
use tokio::runtime::Runtime;
use uuid::Uuid;

fn bench_lru_cache_operations(c: &mut Criterion) {
    let mut cache = LruCache::new(1000);

    c.bench_function("lru_cache_insert", |b| {
        b.iter(|| {
            let state_id = Uuid::new_v4();
            let data = serde_json::json!({"key": "value", "counter": black_box(42)});
            cache.put(state_id, data);
        })
    });

    c.bench_function("lru_cache_get_hit", |b| {
        let state_id = Uuid::new_v4();
        let data = serde_json::json!({"key": "value", "counter": 42});
        cache.put(state_id, data);

        b.iter(|| {
            let _ = cache.get(&state_id);
        })
    });

    c.bench_function("lru_cache_get_miss", |b| {
        b.iter(|| {
            let state_id = Uuid::new_v4();
            let _ = cache.get(&state_id);
        })
    });
}

fn bench_state_manager_operations(c: &mut Criterion) {
    let rt = Runtime::new().unwrap();

    c.bench_function("state_manager_create_state", |b| {
        let manager = StateManager::new();
        let state_id = Uuid::new_v4();

        b.iter(|| {
            rt.block_on(async {
                let initial_data = serde_json::json!({"counter": 0, "items": []});
                manager.load_snapshot(state_id, initial_data).await;
            });
        })
    });

    c.bench_function("state_manager_get_state", |b| {
        let manager = StateManager::new();
        let state_id = Uuid::new_v4();

        rt.block_on(async {
            let initial_data = serde_json::json!({"counter": 0, "items": []});
            manager.load_snapshot(state_id, initial_data).await;
        });

        b.iter(|| {
            rt.block_on(async {
                let _ = manager.get(state_id).await;
            });
        })
    });

    c.bench_function("state_manager_commit_event", |b| {
        let manager = StateManager::new();
        let state_id = Uuid::new_v4();

        rt.block_on(async {
            let initial_data = serde_json::json!({"counter": 0, "items": []});
            manager.load_snapshot(state_id, initial_data).await;
        });

        let event_data = serde_json::json!({"operation": "increment", "amount": 1});

        b.iter(|| {
            rt.block_on(async {
                let event = Event::set(
                    state_id,
                    "counter".to_string(),
                    event_data.clone(),
                    SourceType::Function,
                    "bench".to_string(),
                );
                let _ = manager.commit_event(event).await;
            });
        })
    });
}

fn bench_redis_cache_operations(c: &mut Criterion) {
    let rt = Runtime::new().unwrap();

    // Skip Redis benchmarks if Redis is not available
    let redis_config = RedisConfig {
        url: "redis://localhost:6379".to_string(),
        connection_timeout: 5,
        default_ttl: 3600,
        max_connections: 10,
        key_prefix: "bench".to_string(),
    };

    let manager = rt.block_on(async {
        match RedisCache::new(redis_config).await {
            Ok(cache) => Some(cache),
            Err(_) => {
                eprintln!("Redis not available, skipping Redis benchmarks");
                None
            }
        }
    });

    if let Some(cache) = manager {
        c.bench_function("redis_cache_set_state", |b| {
            let state_id = Uuid::new_v4();
            let data = serde_json::json!({"key": "value", "counter": black_box(42)});

            b.iter(|| {
                rt.block_on(async {
                    let _ = cache.set_state(&state_id, data.clone(), "1".to_string()).await;
                });
            })
        });

        c.bench_function("redis_cache_get_state_hit", |b| {
            let state_id = Uuid::new_v4();
            let data = serde_json::json!({"key": "value", "counter": 42});

            rt.block_on(async {
                cache.set_state(&state_id, data, "1".to_string()).await.unwrap();
            });

            b.iter(|| {
                rt.block_on(async {
                    let _ = cache.get_state(&state_id).await;
                });
            })
        });

        c.bench_function("redis_cache_get_state_miss", |b| {
            b.iter(|| {
                rt.block_on(async {
                    let state_id = Uuid::new_v4();
                    let _ = cache.get_state(&state_id).await;
                });
            })
        });

        c.bench_function("redis_cache_rate_limit_check", |b| {
            b.iter(|| {
                rt.block_on(async {
                    let identifier = format!("user_{}", black_box(12345));
                    let _ = cache.check_rate_limit(&identifier, 100, 60).await;
                });
            })
        });
    }
}

fn bench_event_processing(c: &mut Criterion) {
    c.bench_function("event_creation_and_serialization", |b| {
        b.iter(|| {
            let event = Event::set(
                Uuid::new_v4(),
                "action".to_string(),
                serde_json::json!({"action": "update", "value": black_box(42)}),
                SourceType::Function,
                "bench".to_string(),
            );

            // Serialize to JSON
            let json = serde_json::to_string(&event).unwrap();
            black_box(json);
        })
    });

    c.bench_function("event_deserialization", |b| {
        let event = Event::set(
            Uuid::new_v4(),
            "action".to_string(),
            serde_json::json!({"action": "update", "value": 42}),
            SourceType::Function,
            "bench".to_string(),
        );
        let event_json = serde_json::to_string(&event).unwrap();

        b.iter(|| {
            let event: Event = serde_json::from_str(&event_json).unwrap();
            black_box(event);
        })
    });
}

fn bench_concurrent_state_operations(c: &mut Criterion) {
    let rt = Runtime::new().unwrap();

    c.bench_function("concurrent_state_reads", |b| {
        let manager = Arc::new(StateManager::new());

        // Pre-populate with states
        rt.block_on(async {
            for i in 0..100 {
                let state_id = Uuid::new_v4();
                let data = serde_json::json!({"id": i, "counter": 0});
                manager.load_snapshot(state_id, data).await;
            }
        });

        b.iter(|| {
            rt.block_on(async {
                let mut handles = vec![];
                for _ in 0..10 {
                    let manager_clone = Arc::clone(&manager);
                    let state_id = Uuid::new_v4(); // Random ID for misses
                    let handle = tokio::spawn(async move {
                        let _ = manager_clone.get(state_id).await;
                    });
                    handles.push(handle);
                }

                for handle in handles {
                    let _ = handle.await;
                }
            });
        })
    });
}

fn bench_memory_usage(c: &mut Criterion) {
    c.bench_function("large_state_creation", |b| {
        let manager = StateManager::new();
        let rt = Runtime::new().unwrap();

        b.iter(|| {
            rt.block_on(async {
                let state_id = Uuid::new_v4();

                // Create a large state object
                let mut large_data = serde_json::Map::new();
                for i in 0..1000 {
                    large_data.insert(format!("key_{}", i), serde_json::json!(format!("value_{}", i)));
                }
                let data = serde_json::Value::Object(large_data);

                manager.load_snapshot(state_id, data).await;
            });
        })
    });

    c.bench_function("lru_cache_memory_pressure", |b| {
        let mut cache = LruCache::new(100); // Small cache to test eviction

        b.iter(|| {
            for i in 0..200 { // More operations than cache size
                let state_id = Uuid::new_v4();
                let data = serde_json::json!({"index": i, "data": "x".repeat(1000)}); // 1KB per entry
                cache.put(state_id, data);
            }
        })
    });
}

criterion_group!(
    benches,
    bench_lru_cache_operations,
    bench_state_manager_operations,
    bench_redis_cache_operations,
    bench_event_processing,
    bench_concurrent_state_operations,
    bench_memory_usage
);
criterion_main!(benches);
