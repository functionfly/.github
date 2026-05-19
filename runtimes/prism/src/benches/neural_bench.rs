//! Neural optimizer benchmarks
//!
//! Tests performance of the RL-based neural optimizer including
//! Q-learning updates, suggestion generation, and policy management.

use criterion::{black_box, criterion_group, Criterion};
use crate::core::{CellId, ExecutionMetrics};
use crate::neural::{NeuralOptimizer, ExecutionProfile, ExecutionFeatures, ExecutionOutcome};
use crate::neural::optimizer::OptimizationSuggestion;

fn create_test_profile(cell_id: CellId, outcome: ExecutionOutcome) -> ExecutionProfile {
    ExecutionProfile {
        cell_id,
        metrics: ExecutionMetrics::default(),
        features: ExecutionFeatures {
            input_size_bytes: 1024,
            memory_limit_mb: 128,
            vcpus: 2,
            gpu_used: false,
            execution_location: "cloud".to_string(),
            time_of_day: 12.0,
            day_of_week: 3,
        },
        outcome,
    }
}

pub fn bench_neural_optimizer_creation(c: &mut Criterion) {
    c.bench_function("neural_optimizer_creation", |b| {
        b.iter(|| {
            black_box(NeuralOptimizer::new(1000))
        });
    });
}

pub fn bench_neural_optimizer_small_history(c: &mut Criterion) {
    c.bench_function("neural_optimizer_small_history", |b| {
        b.iter(|| {
            let mut optimizer = NeuralOptimizer::new(100);
            for i in 0..10 {
                let profile = create_test_profile(CellId::new(), ExecutionOutcome::Success);
                optimizer.record(profile);
            }
            black_box(optimizer)
        });
    });
}

pub fn bench_neural_optimizer_large_history(c: &mut Criterion) {
    c.bench_function("neural_optimizer_large_history", |b| {
        b.iter(|| {
            let mut optimizer = NeuralOptimizer::new(10000);
            for i in 0..1000 {
                let outcome = if i % 10 == 0 { ExecutionOutcome::Timeout } else { ExecutionOutcome::Success };
                let profile = create_test_profile(CellId::new(), outcome);
                optimizer.record(profile);
            }
            black_box(optimizer)
        });
    });
}

pub fn bench_neural_record_single(c: &mut Criterion) {
    c.bench_function("neural_record_single", |b| {
        b.iter(|| {
            let mut optimizer = NeuralOptimizer::new(1000);
            let profile = create_test_profile(CellId::new(), ExecutionOutcome::Success);
            black_box(optimizer.record(profile))
        });
    });
}

pub fn bench_neural_suggest(c: &mut Criterion) {
    c.bench_function("neural_suggest", |b| {
        let mut optimizer = NeuralOptimizer::new(1000);
        
        // Add some history
        for i in 0..100 {
            let outcome = if i % 20 == 0 { ExecutionOutcome::OOM } else { ExecutionOutcome::Success };
            let profile = create_test_profile(CellId::new(), outcome);
            optimizer.record(profile);
        }
        
        let cell_id = CellId::new();
        b.iter(|| {
            black_box(optimizer.suggest(&cell_id))
        });
    });
}

pub fn bench_neural_suggest_many_cells(c: &mut Criterion) {
    c.bench_function("neural_suggest_many_cells", |b| {
        let mut optimizer = NeuralOptimizer::new(1000);
        
        // Add substantial history
        for _ in 0..500 {
            let profile = create_test_profile(CellId::new(), ExecutionOutcome::Success);
            optimizer.record(profile);
        }
        
        b.iter(|| {
            for _ in 0..100 {
                let cell_id = CellId::new();
                black_box(optimizer.suggest(&cell_id));
            }
        });
    });
}

pub fn bench_neural_get_policy(c: &mut Criterion) {
    c.bench_function("neural_get_policy", |b| {
        let optimizer = NeuralOptimizer::new(1000);
        
        // Add some history
        for i in 0..50 {
            let profile = create_test_profile(CellId::new(), ExecutionOutcome::Success);
            let mut opt = NeuralOptimizer::new(1000);
            opt.record(profile);
        }
        
        b.iter(|| {
            black_box(optimizer.get_policy())
        });
    });
}

pub fn bench_neural_update_q(c: &mut Criterion) {
    c.bench_function("neural_update_q", |b| {
        let mut optimizer = NeuralOptimizer::new(1000);
        
        b.iter(|| {
            let state = "test-cell-123";
            optimizer.update_q(state, "memory", 1.0, state);
        });
    });
}

pub fn bench_neural_many_updates(c: &mut Criterion) {
    c.bench_function("neural_many_updates", |b| {
        let mut optimizer = NeuralOptimizer::new(1000);
        
        b.iter(|| {
            for i in 0..100 {
                let state = format!("cell-{}", i);
                optimizer.update_q(&state, "memory", 1.0, &state);
                optimizer.update_q(&state, "timeout", -0.5, &state);
            }
        });
    });
}

pub fn bench_execution_profile_creation(c: &mut Criterion) {
    c.bench_function("execution_profile_creation", |b| {
        b.iter(|| {
            black_box(create_test_profile(CellId::new(), ExecutionOutcome::Success))
        });
    });
}

pub fn bench_execution_features_variants(c: &mut Criterion) {
    c.bench_function("execution_features_variants", |b| {
        b.iter(|| {
            let f1 = ExecutionFeatures {
                input_size_bytes: 1024,
                memory_limit_mb: 128,
                vcpus: 2,
                gpu_used: false,
                execution_location: "cloud".to_string(),
                time_of_day: 12.0,
                day_of_week: 3,
            };
            let f2 = ExecutionFeatures {
                input_size_bytes: 4096,
                memory_limit_mb: 512,
                vcpus: 8,
                gpu_used: true,
                execution_location: "edge".to_string(),
                time_of_day: 18.5,
                day_of_week: 6,
            };
            black_box((f1, f2))
        });
    });
}

pub fn bench_execution_outcome_variants(c: &mut Criterion) {
    c.bench_function("execution_outcome_variants", |b| {
        b.iter(|| {
            black_box(ExecutionOutcome::Success);
            black_box(ExecutionOutcome::Timeout);
            black_box(ExecutionOutcome::OOM);
            black_box(ExecutionOutcome::Error);
        });
    });
}

pub fn bench_optimization_suggestion(c: &mut Criterion) {
    c.bench_function("optimization_suggestion_creation", |b| {
        b.iter(|| {
            let suggestion = OptimizationSuggestion {
                cell_id: CellId::new(),
                suggested_memory_mb: 256,
                suggested_timeout_ms: 60000,
                suggested_location: "edge".to_string(),
                cache_recommended: true,
                confidence: 0.85,
            };
            black_box(suggestion)
        });
    });
}

pub fn neural_benchmarks() -> impl Fn(&mut Criterion) {
    move |c: &mut Criterion| {
        bench_neural_optimizer_creation(c);
        bench_neural_optimizer_small_history(c);
        bench_neural_optimizer_large_history(c);
        bench_neural_record_single(c);
        bench_neural_suggest(c);
        bench_neural_suggest_many_cells(c);
        bench_neural_get_policy(c);
        bench_neural_update_q(c);
        bench_neural_many_updates(c);
        bench_execution_profile_creation(c);
        bench_execution_features_variants(c);
        bench_execution_outcome_variants(c);
        bench_optimization_suggestion(c);
    }
}