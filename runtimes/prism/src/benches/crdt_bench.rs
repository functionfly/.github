//! CRDT benchmarks
//!
//! Tests performance of CRDT operations (LWW, GCounter, PnCounter) and merging.

use criterion::{black_box, criterion_group, Criterion};
use crate::state_stream::crdt::{CrdtEngine, CrdtOp, LwwOp, GCounterOp, PnCounterOp};

fn create_test_engine() -> CrdtEngine {
    CrdtEngine::new("bench-node")
}

pub fn bench_crdt_engine_creation(c: &mut Criterion) {
    c.bench_function("crdt_engine_creation", |b| {
        b.iter(|| {
            black_box(CrdtEngine::new("bench-node"))
        });
    });
}

pub fn bench_crdt_engine_creation_with_id(c: &mut Criterion) {
    c.bench_function("crdt_engine_creation_with_id", |b| {
        b.iter(|| {
            black_box(CrdtEngine::new("node-12345678901234567890"))
        });
    });
}

pub fn bench_lww_register_write(c: &mut Criterion) {
    c.bench_function("lww_register_write", |b| {
        b.iter(|| {
            let mut engine = create_test_engine();
            let op = CrdtOp::Lww(LwwOp {
                node_id: "node-1".to_string(),
                timestamp: chrono::Utc::now().timestamp_millis(),
                value: b"test-value".to_vec(),
            });
            black_box(engine.apply("key-1", op).unwrap())
        });
    });
}

pub fn bench_lww_register_many_writes(c: &mut Criterion) {
    c.bench_function("lww_register_many_writes", |b| {
        b.iter(|| {
            let mut engine = create_test_engine();
            let base_ts = chrono::Utc::now().timestamp_millis();
            
            for i in 0..100 {
                let op = CrdtOp::Lww(LwwOp {
                    node_id: format!("node-{}", i % 5),
                    timestamp: base_ts + i as i64,
                    value: format!("value-{}", i).into_bytes(),
                });
                let _ = engine.apply(&format!("key-{}", i % 20), op);
            }
            black_box(engine)
        });
    });
}

pub fn bench_gcounter_increment(c: &mut Criterion) {
    c.bench_function("gcounter_increment", |b| {
        b.iter(|| {
            let mut engine = create_test_engine();
            let op = CrdtOp::GCounter(GCounterOp {
                node_id: "node-1".to_string(),
                delta: 1,
            });
            black_box(engine.apply("counter-1", op).unwrap())
        });
    });
}

pub fn bench_gcounter_many_increments(c: &mut Criterion) {
    c.bench_function("gcounter_many_increments", |b| {
        b.iter(|| {
            let mut engine = create_test_engine();
            
            for i in 0..1000 {
                let op = CrdtOp::GCounter(GCounterOp {
                    node_id: "node-1".to_string(),
                    delta: i as u64,
                });
                let _ = engine.apply("counter-1", op);
            }
            black_box(engine.get_counter("counter-1"))
        });
    });
}

pub fn bench_pncounter_operations(c: &mut Criterion) {
    c.bench_function("pncounter_operations", |b| {
        b.iter(|| {
            let mut engine = create_test_engine();
            
            // Increment
            let inc_op = CrdtOp::PnCounter(PnCounterOp {
                node_id: "node-1".to_string(),
                delta: 10,
            });
            let _ = engine.apply("pn-counter", inc_op);
            
            // Decrement
            let dec_op = CrdtOp::PnCounter(PnCounterOp {
                node_id: "node-1".to_string(),
                delta: -3,
            });
            black_box(engine.apply("pn-counter", dec_op).unwrap())
        });
    });
}

pub fn bench_crdt_merge(c: &mut Criterion) {
    c.bench_function("crdt_merge", |b| {
        b.iter(|| {
            let mut engine1 = CrdtEngine::new("node-1");
            let mut engine2 = CrdtEngine::new("node-2");
            
            // Add data to engine1
            for i in 0..50 {
                let op = CrdtOp::Lww(LwwOp {
                    node_id: "node-1".to_string(),
                    timestamp: i as i64,
                    value: format!("value-{}", i).into_bytes(),
                });
                let _ = engine1.apply(&format!("key-{}", i), op);
            }
            
            // Add data to engine2 (overlapping keys)
            for i in 25..75 {
                let op = CrdtOp::Lww(LwwOp {
                    node_id: "node-2".to_string(),
                    timestamp: i as i64 + 100, // Higher timestamps
                    value: format!("value-{}", i).into_bytes(),
                });
                let _ = engine2.apply(&format!("key-{}", i), op);
            }
            
            // Merge engine2 into engine1
            black_box(engine1.merge(&engine2))
        });
    });
}

pub fn bench_crdt_history(c: &mut Criterion) {
    c.bench_function("crdt_history", |b| {
        b.iter(|| {
            let mut engine = create_test_engine();
            
            for i in 0..100 {
                let op = CrdtOp::Lww(LwwOp {
                    node_id: format!("node-{}", i % 3),
                    timestamp: i as i64,
                    value: format!("value-{}", i).into_bytes(),
                });
                let _ = engine.apply("key-1", op);
            }
            
            black_box(engine.get_history())
        });
    });
}

pub fn bench_crdt_history_for_key(c: &mut Criterion) {
    c.bench_function("crdt_history_for_key", |b| {
        b.iter(|| {
            let mut engine = create_test_engine();
            
            for i in 0..200 {
                let op = CrdtOp::Lww(LwwOp {
                    node_id: format!("node-{}", i % 5),
                    timestamp: i as i64,
                    value: format!("value-{}", i).into_bytes(),
                });
                let key = format!("key-{}", i % 10);
                let _ = engine.apply(&key, op);
            }
            
            black_box(engine.get_history_for_key("key-5"))
        });
    });
}

pub fn bench_crdt_multiple_registers(c: &mut Criterion) {
    c.bench_function("crdt_multiple_registers", |b| {
        b.iter(|| {
            let mut engine = create_test_engine();
            
            for i in 0..50 {
                let op = CrdtOp::Lww(LwwOp {
                    node_id: "node-1".to_string(),
                    timestamp: i as i64,
                    value: format!("value-for-key-{}", i).into_bytes(),
                });
                let _ = engine.apply(&format!("register-{}", i), op);
            }
            
            // Access all registers
            for i in 0..50 {
                black_box(engine.get_register(&format!("register-{}", i)));
            }
        });
    });
}

pub fn bench_crdt_multiple_counters(c: &mut Criterion) {
    c.bench_function("crdt_multiple_counters", |b| {
        b.iter(|| {
            let mut engine = create_test_engine();
            
            for i in 0..100 {
                let op = CrdtOp::GCounter(GCounterOp {
                    node_id: format!("node-{}", i % 5),
                    delta: i as u64,
                });
                let _ = engine.apply(&format!("counter-{}", i % 20), op);
            }
            
            // Get all counter values
            for i in 0..20 {
                black_box(engine.get_counter(&format!("counter-{}", i)));
            }
        });
    });
}

pub fn bench_lww_register_read(c: &mut Criterion) {
    c.bench_function("lww_register_read", |b| {
        b.iter(|| {
            let mut engine = create_test_engine();
            
            // Write a value
            let op = CrdtOp::Lww(LwwOp {
                node_id: "node-1".to_string(),
                timestamp: chrono::Utc::now().timestamp_millis(),
                value: b"important-data".to_vec(),
            });
            let _ = engine.apply("important-key", op);
            
            // Read it
            black_box(engine.get_register("important-key"))
        });
    });
}

pub fn crdt_benchmarks() -> impl Fn(&mut Criterion) {
    move |c: &mut Criterion| {
        bench_crdt_engine_creation(c);
        bench_crdt_engine_creation_with_id(c);
        bench_lww_register_write(c);
        bench_lww_register_many_writes(c);
        bench_gcounter_increment(c);
        bench_gcounter_many_increments(c);
        bench_pncounter_operations(c);
        bench_crdt_merge(c);
        bench_crdt_history(c);
        bench_crdt_history_for_key(c);
        bench_crdt_multiple_registers(c);
        bench_crdt_multiple_counters(c);
        bench_lww_register_read(c);
    }
}