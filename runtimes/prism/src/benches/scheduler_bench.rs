//! Scheduler benchmarks
//!
//! Tests performance of node registration, scheduling decisions, and stats collection.

use criterion::{black_box, criterion_group, Criterion};
use crate::core::{CellId, ExecutionLocation, PlacementHint};
use crate::hypercore::{Scheduler, SchedulerConfig, Node, NodeResources, ScheduleRequest};

fn create_test_node(id: &str, vcpus: u32, memory: u64) -> Node {
    Node::new(
        id,
        ExecutionLocation::Cloud,
        NodeResources::new(vcpus, memory),
    )
}

pub fn bench_scheduler_creation(c: &mut Criterion) {
    c.bench_function("scheduler_creation", |b| {
        b.iter(|| {
            black_box(Scheduler::new(SchedulerConfig::default()))
        });
    });
}

pub fn bench_scheduler_with_custom_config(c: &mut Criterion) {
    c.bench_function("scheduler_custom_config", |b| {
        b.iter(|| {
            let config = SchedulerConfig {
                max_concurrent: 128,
                max_queue_size: 2048,
                rate_limit: 5000,
            };
            black_box(Scheduler::new(config))
        });
    });
}

pub fn bench_scheduler_register_node(c: &mut Criterion) {
    c.bench_function("scheduler_register_node", |b| {
        let runtime = tokio::runtime::Runtime::new().unwrap();
        b.to_async(&runtime).iter(|| async {
            let scheduler = Scheduler::new(SchedulerConfig::default());
            let node = create_test_node("node-1", 4, 8192);
            black_box(scheduler.register_node(node).await)
        });
    });
}

pub fn bench_scheduler_register_multiple_nodes(c: &mut Criterion) {
    c.bench_function("scheduler_register_multiple_nodes", |b| {
        let runtime = tokio::runtime::Runtime::new().unwrap();
        b.iter(|| {
            runtime.block_on(async {
                let scheduler = Scheduler::new(SchedulerConfig::default());
                for i in 0..10 {
                    let node = create_test_node(&format!("node-{}", i), 4, 8192);
                    let _ = scheduler.register_node(node).await;
                }
                black_box(scheduler.list_nodes().await)
            });
        });
    });
}

pub fn bench_scheduler_schedule_empty(c: &mut Criterion) {
    c.bench_function("scheduler_schedule_empty", |b| {
        let runtime = tokio::runtime::Runtime::new().unwrap();
        b.to_async(&runtime).iter(|| async {
            let scheduler = Scheduler::new(SchedulerConfig::default());
            let request = ScheduleRequest {
                cell_id: CellId::new(),
                required_vcpus: 2,
                required_memory: 4096,
                placement_hint: None,
            };
            black_box(scheduler.schedule(request).await)
        });
    });
}

pub fn bench_scheduler_schedule_with_nodes(c: &mut Criterion) {
    c.bench_function("scheduler_schedule_with_nodes", |b| {
        let runtime = tokio::runtime::Runtime::new().unwrap();
        b.iter(|| {
            runtime.block_on(async {
                let scheduler = Scheduler::new(SchedulerConfig::default());
                
                // Register multiple nodes
                for i in 0..5 {
                    let node = create_test_node(&format!("node-{}", i), 4 + i as u32, 8192);
                    let _ = scheduler.register_node(node).await;
                }
                
                let request = ScheduleRequest {
                    cell_id: CellId::new(),
                    required_vcpus: 2,
                    required_memory: 4096,
                    placement_hint: None,
                };
                black_box(scheduler.schedule(request).await)
            });
        });
    });
}

pub fn bench_scheduler_schedule_many_cells(c: &mut Criterion) {
    c.bench_function("scheduler_schedule_many_cells", |b| {
        let runtime = tokio::runtime::Runtime::new().unwrap();
        b.iter(|| {
            runtime.block_on(async {
                let scheduler = Scheduler::new(SchedulerConfig::default());
                
                // Register nodes
                for i in 0..3 {
                    let node = create_test_node(&format!("node-{}", i), 16, 32768);
                    let _ = scheduler.register_node(node).await;
                }
                
                // Schedule many cells
                let mut results = Vec::new();
                for i in 0..50 {
                    let request = ScheduleRequest {
                        cell_id: CellId::new(),
                        required_vcpus: 2,
                        required_memory: 4096,
                        placement_hint: None,
                    };
                    results.push(scheduler.schedule(request).await);
                }
                black_box(results)
            });
        });
    });
}

pub fn bench_scheduler_list_nodes(c: &mut Criterion) {
    c.bench_function("scheduler_list_nodes", |b| {
        let runtime = tokio::runtime::Runtime::new().unwrap();
        b.iter(|| {
            runtime.block_on(async {
                let scheduler = Scheduler::new(SchedulerConfig::default());
                
                // Register nodes
                for i in 0..20 {
                    let node = create_test_node(&format!("node-{}", i), 4, 8192);
                    let _ = scheduler.register_node(node).await;
                }
                
                black_box(scheduler.list_nodes().await)
            });
        });
    });
}

pub fn bench_scheduler_get_stats(c: &mut Criterion) {
    c.bench_function("scheduler_get_stats", |b| {
        let runtime = tokio::runtime::Runtime::new().unwrap();
        b.iter(|| {
            runtime.block_on(async {
                let scheduler = Scheduler::new(SchedulerConfig::default());
                
                // Register nodes
                for i in 0..5 {
                    let node = create_test_node(&format!("node-{}", i), 4, 8192);
                    let _ = scheduler.register_node(node).await;
                }
                
                // Make some scheduling decisions
                for _ in 0..10 {
                    let request = ScheduleRequest {
                        cell_id: CellId::new(),
                        required_vcpus: 2,
                        required_memory: 4096,
                        placement_hint: None,
                    };
                    let _ = scheduler.schedule(request).await;
                }
                
                black_box(scheduler.get_stats().await)
            });
        });
    });
}

pub fn bench_scheduler_placement_history(c: &mut Criterion) {
    c.bench_function("scheduler_placement_history", |b| {
        let runtime = tokio::runtime::Runtime::new().unwrap();
        b.iter(|| {
            runtime.block_on(async {
                let scheduler = Scheduler::new(SchedulerConfig::default());
                
                // Register nodes
                for i in 0..3 {
                    let node = create_test_node(&format!("node-{}", i), 4, 8192);
                    let _ = scheduler.register_node(node).await;
                }
                
                // Make scheduling decisions
                for i in 0..100 {
                    let request = ScheduleRequest {
                        cell_id: CellId::new(),
                        required_vcpus: 2,
                        required_memory: 4096,
                        placement_hint: None,
                    };
                    let _ = scheduler.schedule(request).await;
                }
                
                black_box(scheduler.get_recent_placements(50).await)
            });
        });
    });
}

pub fn bench_scheduler_cancel_placement(c: &mut Criterion) {
    c.bench_function("scheduler_cancel_placement", |b| {
        let runtime = tokio::runtime::Runtime::new().unwrap();
        b.iter(|| {
            runtime.block_on(async {
                let scheduler = Scheduler::new(SchedulerConfig::default());
                
                // Register nodes
                for i in 0..3 {
                    let node = create_test_node(&format!("node-{}", i), 4, 8192);
                    let _ = scheduler.register_node(node).await;
                }
                
                // Schedule a cell
                let cell_id = CellId::new();
                let request = ScheduleRequest {
                    cell_id,
                    required_vcpus: 2,
                    required_memory: 4096,
                    placement_hint: None,
                };
                let _ = scheduler.schedule(request).await;
                
                black_box(scheduler.cancel_placement(&cell_id).await)
            });
        });
    });
}

pub fn bench_placement_hint_creation(c: &mut Criterion) {
    c.bench_function("placement_hint_creation", |b| {
        b.iter(|| {
            black_box(PlacementHint::default())
        });
    });

    c.bench_function("placement_hint_custom", |b| {
        b.iter(|| {
            let hint = PlacementHint {
                preferred_location: ExecutionLocation::Edge,
                preferred_regions: vec!["us-west-2".to_string(), "eu-west-1".to_string()],
                latency_sensitivity: 0.8,
                cost_sensitivity: 0.3,
                gpu_required: true,
                model_affinity: Some("llama2".to_string()),
            };
            black_box(hint)
        });
    });
}

pub fn scheduler_benchmarks() -> impl Fn(&mut Criterion) {
    move |c: &mut Criterion| {
        bench_scheduler_creation(c);
        bench_scheduler_with_custom_config(c);
        bench_scheduler_register_node(c);
        bench_scheduler_register_multiple_nodes(c);
        bench_scheduler_schedule_empty(c);
        bench_scheduler_schedule_with_nodes(c);
        bench_scheduler_schedule_many_cells(c);
        bench_scheduler_list_nodes(c);
        bench_scheduler_get_stats(c);
        bench_scheduler_placement_history(c);
        bench_scheduler_cancel_placement(c);
        bench_placement_hint_creation(c);
    }
}