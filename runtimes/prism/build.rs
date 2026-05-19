fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Proto files are pre-generated to src/proto/prism_generated.rs
    // This build script exists to trigger re-runs when proto changes
    println!("cargo:rerun-if-changed=proto/prism.proto");
    println!("cargo:rerun-if-changed=src/proto/prism_generated.rs");
    Ok(())
}