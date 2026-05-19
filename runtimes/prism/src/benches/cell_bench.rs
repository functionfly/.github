//! Cell lifecycle benchmarks
//!
//! Tests performance of cell creation, status transitions, and migration eligibility.

use criterion::{black_box, criterion_group, Criterion, Benchmark};
use crate::core::{CellId, CellStatus, CellConfig, CellMetadata, ExecutionCell, ExecutionTarget};

pub fn bench_cell_creation(c: &mut Criterion) {
    c.bench_function("cell_creation", |b| {
        b.iter(|| {
            let config = CellConfig::default();
            let metadata = CellMetadata::new("bench-cell", "wasm");
            black_box(ExecutionCell::new("tenant-1", config, metadata))
        });
    });
}

pub fn bench_cell_id_generation(c: &mut Criterion) {
    c.bench_function("cell_id_generation", |b| {
        b.iter(|| {
            black_box(CellId::new())
        });
    });
}

pub fn bench_cell_status_transition(c: &mut Criterion) {
    let config = CellConfig::default();
    let metadata = CellMetadata::new("bench-cell", "wasm");
    let mut cell = ExecutionCell::new("tenant-1", config, metadata);

    c.bench_function("cell_status_transition", |b| {
        b.iter(|| {
            cell.set_status(CellStatus::Running);
            cell.set_status(CellStatus::Waiting);
            cell.set_status(CellStatus::Running);
        });
    });
}

pub fn bench_cell_can_migrate(c: &mut Criterion) {
    let config = CellConfig::default();
    let metadata = CellMetadata::new("bench-cell", "wasm");
    let mut cell = ExecutionCell::new("tenant-1", config, metadata);

    let statuses = [
        CellStatus::Running,
        CellStatus::Waiting,
        CellStatus::Frozen,
        CellStatus::Pending,
        CellStatus::Terminated,
    ];

    c.bench_function("cell_can_migrate", |b| {
        let mut i = 0;
        b.iter(|| {
            cell.set_status(statuses[i % statuses.len()]);
            black_box(cell.can_migrate());
            i += 1;
        });
    });
}

pub fn bench_cell_clone(c: &mut Criterion) {
    let config = CellConfig::default();
    let metadata = CellMetadata::new("bench-cell", "wasm");
    let cell = ExecutionCell::new("tenant-1", config, metadata);

    c.bench_function("cell_clone", |b| {
        b.iter(|| {
            black_box(cell.clone())
        });
    });
}

pub fn bench_cell_config_defaults(c: &mut Criterion) {
    c.bench_function("cell_config_default", |b| {
        b.iter(|| {
            black_box(CellConfig::default())
        });
    });
}

pub fn bench_execution_target_variants(c: &mut Criterion) {
    c.bench_function("execution_target_variants", |b| {
        b.iter(|| {
            black_box(ExecutionTarget::Edge);
            black_box(ExecutionTarget::Browser);
            black_box(ExecutionTarget::Robotic);
            black_box(ExecutionTarget::IoT);
        });
    });
}

pub fn bench_cell_metadata_creation(c: &mut Criterion) {
    c.bench_function("cell_metadata_creation", |b| {
        b.iter(|| {
            black_box(CellMetadata::new("meta-cell", "wasm"))
        });
    });
}

pub fn bench_cell_with_custom_config(c: &mut Criterion) {
    c.bench_function("cell_with_custom_config", |b| {
        let mut config = CellConfig::default();
        config.memory_limit_mb = 512;
        config.timeout_ms = 60_000;
        config.capabilities = vec!["gpu".to_string(), "inference".to_string()];

        b.iter(|| {
            let metadata = CellMetadata::new("custom-cell", "wasm");
            black_box(ExecutionCell::new("tenant-1", config.clone(), metadata))
        });
    });
}

pub fn bench_cell_is_running(c: &mut Criterion) {
    let config = CellConfig::default();
    let metadata = CellMetadata::new("bench-cell", "wasm");
    let mut cell = ExecutionCell::new("tenant-1", config, metadata);

    c.bench_function("cell_is_running", |b| {
        b.iter(|| {
            cell.set_status(CellStatus::Running);
            black_box(cell.is_running());
        });
    });
}

pub fn cell_benchmarks() -> impl Fn(&mut Criterion) {
    move |c: &mut Criterion| {
        bench_cell_creation(c);
        bench_cell_id_generation(c);
        bench_cell_status_transition(c);
        bench_cell_can_migrate(c);
        bench_cell_clone(c);
        bench_cell_config_defaults(c);
        bench_execution_target_variants(c);
        bench_cell_metadata_creation(c);
        bench_cell_with_custom_config(c);
        bench_cell_is_running(c);
    }
}