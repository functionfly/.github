# ONNX Models Directory

This directory contains ONNX model files used by the FunctionFly AI inference host functions.

## Security

All model files are verified against SHA256 checksums before use. The `download_model.py` script automatically verifies downloads.

**Important:** Large model files (`.onnx`) are not checked into git. Use the download script or Git LFS.

## Available Models

### alexnet.onnx (✅ Ready to use)

- **Size**: ~244MB
- **Type**: Image classification (1000 classes)
- **Input**: `[1, 3, 224, 224]` - RGB images normalized to [0,1]
- **Output**: `[1, 1000]` - Class probabilities
- **Use case**: General image classification

### Downloading More Models

Use the download script to get additional models. All downloads are verified against SHA256 checksums:

```bash
# Activate the virtual environment first
source onnx-env/bin/activate

# Download models (automatically verified)
python download_model.py alexnet     # Already downloaded
python download_model.py mobilenet   # Lightweight image classification
python download_model.py bert-small   # Text understanding
```

If you have an existing alexnet.onnx, verify its hash:
```bash
sha256sum runtimes/local/models/alexnet.onnx
```

## Model Input/Output Format

Models expect JSON input and return JSON output:

### Input Format

- **1D tensors**: `[1.0, 2.0, 3.0, ...]`
- **2D tensors**: `[[1.0, 2.0], [3.0, 4.0], ...]`
- **Higher dimensions**: Flat array matching the expected shape

### Output Format

- **1D tensors**: `[0.1, 0.9, 0.8, ...]`
- **2D tensors**: `[[0.1, 0.2], [0.9, 0.8], ...]`
- **Multiple outputs**: `{"output_0": [...], "output_1": [...], ...}`

## Example Usage

```javascript
// In your WASM function - Image classification with AlexNet
const imageData = "[0.5, 0.3, 0.8, ...]"; // 224x224x3 normalized pixels as flat array
const result = functionfly.ai("alexnet", imageData);
console.log(result); // JSON array with 1000 class probabilities

// Custom model
const result = functionfly.ai("my_model", "[1.0, 2.0, 3.0]");
console.log(result); // JSON string with inference results
```

## Getting More ONNX Models

### Automated Download

```bash
# Use the download script
source onnx-env/bin/activate
python download_model.py <model_name>

# Available models:
# - alexnet: Image classification (~244MB) ✅
# - mobilenet: Lightweight image classification
# - bert-small: Text understanding
```

### Manual Conversion

Convert models from various frameworks to ONNX:

- **PyTorch**: Use `torch.onnx.export()`
- **TensorFlow**: Use `tf2onnx` or `tensorflow-onnx`
- **Hugging Face**: Many models are available in ONNX format
- **Scikit-learn**: Use `skl2onnx`

### From Hugging Face

```bash
# Install optimum
pip install optimum

# Convert any model to ONNX
optimum-cli export onnx --model microsoft/DialoGPT-small model/
```

## Model Caching

Loaded models are automatically cached in memory to improve performance on subsequent calls.

## Virtual Environment

A Python virtual environment is set up in `onnx-env/` for downloading and managing models without affecting system packages.

```bash
# Activate environment
source onnx-env/bin/activate

# Run download script
python download_model.py alexnet
```
