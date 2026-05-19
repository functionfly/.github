//! Criterion benchmark harness for Prism Runtime
//!
//! Run benchmarks with:
//!   cargo bench --package functionfly-prism
//!
//! Generate HTML reports:
//!   cargo bench --package functionfly-prism -- --profile-time

use criterion::Criterion;

fn main() {
    let mut criterion = Criterion::default()
        .configure_from_args()
        .sample_size(100)
        .measurement_time(std::time::Duration::from_secs(10));

    // Import and run all benchmarks
    prism_runtime::benches::run_benchmarks(&mut criterion);
}