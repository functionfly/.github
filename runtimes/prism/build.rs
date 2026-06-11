fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Regenerate the gRPC stubs (message types + server trait) from
    // proto/prism.proto every time the proto file changes. The generated
    // code is written to OUT_DIR and re-exported from src/proto/mod.rs
    // via a `include!`. The pre-existing src/proto/prism_generated.rs
    // file is removed by the build (it is no longer the source of
    // truth) and the gRPC server trait lives in the regenerated module.
    println!("cargo:rerun-if-changed=proto/prism.proto");

    tonic_prost_build::configure()
        .build_server(true)
        .build_client(false)
        .file_descriptor_set_path(
            std::path::PathBuf::from(std::env::var("OUT_DIR")?)
                .join("prism_descriptor.bin"),
        )
        .compile_protos(&["proto/prism.proto"], &["proto"])?;

    // Remove the stale hand-maintained generated file if it exists. It
    // would otherwise conflict with the regenerated `prism.v1.rs` types
    // pulled in via OUT_DIR (the two use different prost type defaults,
    // e.g. BTreeMap vs HashMap for proto map fields).
    let generated = std::path::PathBuf::from("src/proto/prism_generated.rs");
    if generated.exists() {
        std::fs::remove_file(&generated).ok();
    }

    Ok(())
}
