//! AI/ML inference host function implementation

use wasmtime_wasi::p1::WasiP1Ctx;
use std::collections::HashMap;
use std::sync::Mutex;
use once_cell::sync::Lazy;
use tract_onnx::prelude::*;

use super::memory_utils;

// Global model cache to avoid reloading models
static MODEL_CACHE: Lazy<Mutex<HashMap<String, Vec<u8>>>> = Lazy::new(|| Mutex::new(HashMap::new()));

/// Add the functionfly.ai function for AI/ML inference
pub fn add_ai_function(
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
) -> anyhow::Result<()> {
    // functionfly.ai(model_ptr: i32, model_len: i32, input_ptr: i32, input_len: i32,
    //                 output_ptr: i32, output_len_ptr: i32) -> i32
    // Returns 0 on success, negative values on error
    linker.func_wrap(
        "functionfly",
        "ai",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              model_ptr: i32,
              model_len: i32,
              input_ptr: i32,
              input_len: i32,
              output_ptr: i32,
              output_len_ptr: i32| -> i32 {
            // Get model name from WASM memory
            let model = match memory_utils::read_string_from_memory(&mut caller, model_ptr, model_len) {
                Ok(m) => m,
                Err(_) => return -1, // Invalid model
            };

            // Get input data from WASM memory
            let input = match memory_utils::read_string_from_memory(&mut caller, input_ptr, input_len) {
                Ok(i) => i,
                Err(_) => return -2, // Invalid input
            };

            // Run AI inference
            let result = run_ai_inference(&model, &input);

            match result {
                Ok(output) => {
                    // Write output back to WASM memory
                    match memory_utils::write_string_to_memory(&mut caller, &output, output_ptr, output_len_ptr) {
                        Ok(_) => 0, // Success
                        Err(_) => -3, // Memory write error
                    }
                }
                Err(_) => -4, // AI inference error
            }
        },
    )?;

    tracing::debug!("Added functionfly.ai host function");
    Ok(())
}

/// Load ONNX model bytes from cache or file
fn load_onnx_model_bytes(model_name: &str) -> anyhow::Result<Vec<u8>> {
    let mut cache = MODEL_CACHE.lock().map_err(|_| anyhow::anyhow!("Failed to lock model cache"))?;

    if let Some(model_bytes) = cache.get(model_name) {
        return Ok(model_bytes.clone());
    }

    // Try to load model from models directory
    let model_path = format!("models/{}.onnx", model_name);
    let model_bytes = std::fs::read(&model_path)
        .map_err(|_| anyhow::anyhow!("Model file not found: {}. Expected at {}", model_name, model_path))?;

    // Cache the model bytes
    cache.insert(model_name.to_string(), model_bytes.clone());

    tracing::info!("Loaded ONNX model: {}", model_name);
    Ok(model_bytes)
}

/// Run inference on an ONNX model
fn run_onnx_inference(model_name: &str, input: &str) -> anyhow::Result<String> {
    let model_bytes = load_onnx_model_bytes(model_name)?;

    // Load and optimize the ONNX model
    let model = tract_onnx::onnx()
        .model_for_read(&mut std::io::Cursor::new(model_bytes))?
        .into_optimized()?
        .into_runnable()?;

    // For simplicity, assume 1D float input for now
    let input_values: Vec<f32> = serde_json::from_str(input)
        .map_err(|_| anyhow::anyhow!("Expected JSON array of numbers for input"))?;

    let input_tensor = Tensor::from_shape(&[input_values.len()], &input_values)
        .map_err(|e| anyhow::anyhow!("Failed to create input tensor: {}", e))?;

    // Run inference
    let outputs = model.run(tvec!(input_tensor.into()))
        .map_err(|e| anyhow::anyhow!("Inference failed: {}", e))?;

    // Format the first output as JSON
    if let Some(first_output) = outputs.get(0) {
        let values: &[f32] = first_output.as_slice()
            .map_err(|e| anyhow::anyhow!("Failed to get output values: {}", e))?;
        serde_json::to_string(values)
            .map_err(|e| anyhow::anyhow!("Failed to serialize output: {}", e))
    } else {
        Err(anyhow::anyhow!("No outputs from model"))
    }
}

/// Run AI inference on a model
pub fn run_ai_inference(
    model: &str,
    input: &str,
) -> anyhow::Result<String> {
    // Try to run ONNX inference first
    match run_onnx_inference(model, input) {
        Ok(result) => Ok(result),
        Err(e) => {
            // Fall back to simple placeholder implementations if ONNX model not found
            tracing::warn!("ONNX inference failed for '{}' ({}), using placeholder", model, e);
            run_placeholder_inference(model, input)
        }
    }
}

/// Fallback placeholder inference for when ONNX models aren't available
fn run_placeholder_inference(
    model: &str,
    input: &str,
) -> anyhow::Result<String> {
    match model {
        "sentiment" => {
            // Simple sentiment analysis placeholder
            if input.contains("good") || input.contains("great") {
                Ok("positive".to_string())
            } else if input.contains("bad") || input.contains("terrible") {
                Ok("negative".to_string())
            } else {
                Ok("neutral".to_string())
            }
        }
        "classify" => {
            // Simple text classification placeholder
            if input.len() > 100 {
                Ok("long_text".to_string())
            } else {
                Ok("short_text".to_string())
            }
        }
        _ => Err(anyhow::anyhow!("Unsupported model: {}", model)),
    }
}