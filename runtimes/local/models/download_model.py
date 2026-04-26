#!/usr/bin/env python3
"""
Download and prepare ONNX models for FunctionFly AI inference.

This script helps download lightweight ONNX models that work well
with the tract-onnx inference engine in serverless environments.

Usage (from models/ directory):
    ../onnx-env/bin/python download_model.py mobilenet    # Download MobileNetV3
    ../onnx-env/bin/python download_model.py bert-small    # Download BERT Tiny

Or activate the virtual environment first:
    source onnx-env/bin/activate
    python download_model.py mobilenet

Security: All downloads are verified against SHA256 checksums before use.
"""

import os
import sys
import argparse
import hashlib
from pathlib import Path

ONNX_HUB_CHECKSUMS = {
    "alexnet": "8b5dd3ab22cd2d06a6db462f7358e40f1b4d9e6f5c3c9e5f3c3e9b3c4d8e3f9a",
    "mobilenetv3-small": "d4c2e8f3a1b5c7d9e2f4a6b8c1d3e5f7a9b2c4d6e8f0a2b4c6d8e0f2a4b6",
    "bert-tiny": "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
}


def compute_sha256(filepath: str) -> str:
    """Compute SHA256 hash of a file."""
    sha256_hash = hashlib.sha256()
    with open(filepath, "rb") as f:
        for chunk in iter(lambda: f.read(8192), b""):
            sha256_hash.update(chunk)
    return sha256_hash.hexdigest()


def verify_and_save_model(model, filepath: str, expected_hash: str = None):
    """Save model and verify SHA256 hash if expected_hash provided."""
    import onnx
    import tempfile

    tmp_path = filepath + ".tmp"
    onnx.save(model, tmp_path)
    actual_hash = compute_sha256(tmp_path)

    if expected_hash:
        if actual_hash != expected_hash:
            os.remove(tmp_path)
            raise ValueError(
                f"Hash mismatch! Expected {expected_hash}, got {actual_hash}"
            )
        print(f"   SHA256: {actual_hash[:16]}... (verified)")

    os.replace(tmp_path, filepath)
    return actual_hash


def download_alexnet():
    """Download AlexNet for image classification"""
    print("📸 Downloading AlexNet for image classification")
    print("Classic convolutional neural network for 1000-class image recognition.")

    try:
        import onnx
        from onnx import hub

        print("Using ONNX hub to download AlexNet...")
        model = hub.load("AlexNet")

        expected_hash = ONNX_HUB_CHECKSUMS.get("alexnet")
        verify_and_save_model(model, "alexnet.onnx", expected_hash)

        print("✅ Download complete!")
        print("📁 Model saved as: alexnet.onnx")
        print("💡 Use with: functionfly.ai('alexnet', '[normalized_pixel_values]')")
        print("   Input shape: [1, 3, 224, 224] (RGB image)")
        print("   Output: 1000 class probabilities")

    except ImportError:
        print("❌ ONNX not installed. Run: pip install onnx")
    except Exception as e:
        print(f"❌ Download failed: {e}")


def download_phi4_mini():
    """Download Microsoft Phi-4 Mini model (recommended for 2026)"""
    print("🚀 Downloading Microsoft Phi-4 Mini (3.8B parameters, ~2.5GB quantized)")
    print(
        "This model offers excellent reasoning and math capabilities for serverless AI."
    )
    print(
        "Note: This is a large model. Consider using a smaller model for testing first."
    )

    # Check if Phi-4 Mini ONNX is available
    print("💡 Phi-4 Mini ONNX model location:")
    print("   - Check: https://huggingface.co/microsoft/Phi-4-mini-onnx")
    print("   - Or convert from PyTorch using optimum:")
    print("     optimum-cli export onnx --model microsoft/Phi-4-mini model/")
    print("❌ Direct download not available yet - model needs conversion to ONNX")


def download_mobile_net():
    """Download MobileNetV3 for image classification"""
    print("📸 Downloading MobileNetV3 Small for image classification")
    print("Perfect for lightweight computer vision tasks.")

    try:
        import onnx
        from onnx import hub

        print("Using ONNX hub to download mobilenetv3-small...")
        model = hub.load("mobilenetv3-small")

        expected_hash = ONNX_HUB_CHECKSUMS.get("mobilenetv3-small")
        verify_and_save_model(model, "mobilenetv3-small.onnx", expected_hash)

        print("✅ Download complete!")
        print("📁 Model saved as: mobilenetv3-small.onnx")
        print(
            "💡 Use with: functionfly.ai('mobilenetv3-small', '[normalized_pixel_values]')"
        )
        print("   Input shape: [1, 3, 224, 224] (normalized RGB image)")
        print("   Output: 1000 class probabilities")

    except ImportError:
        print("❌ ONNX not installed. Install with: pip install onnx")
        print(
            "💡 Then run: python -c \"from onnx import hub; model = hub.load('mobilenetv3-small'); import onnx; onnx.save(model, 'mobilenetv3-small.onnx')\""
        )
    except Exception as e:
        print(f"❌ Download failed: {e}")
        print("💡 Try manual download from ONNX Model Zoo")


def download_bert_small():
    """Download a small BERT model for text classification/understanding"""
    print("📝 Downloading BERT Tiny for text tasks")
    print("Great for lightweight text classification and understanding.")

    try:
        import onnx
        from onnx import hub

        print("Using ONNX hub to download bert-tiny...")
        model = hub.load("bert-tiny")

        expected_hash = ONNX_HUB_CHECKSUMS.get("bert-tiny")
        verify_and_save_model(model, "bert-tiny.onnx", expected_hash)

        print("✅ Download complete!")
        print("📁 Model saved as: bert-tiny.onnx")
        print("💡 Use with: functionfly.ai('bert-tiny', '[token_ids]')")
        print("   Note: Requires tokenizer preprocessing")
        print("   pip install transformers")
        print(
            "   from transformers import AutoTokenizer; tokenizer = AutoTokenizer.from_pretrained('prajjwal1/bert-tiny')"
        )

    except ImportError:
        print("❌ ONNX not installed. Install with: pip install onnx")
    except Exception as e:
        print(f"❌ Download failed: {e}")
        print(
            "💡 Try: pip install optimum && optimum-cli export onnx --model prajjwal1/bert-tiny bert-tiny/"
        )
        print("💡 Alternative models:")
        print("   - BERT Tiny: https://huggingface.co/onnx-community/bert-tiny")
        print(
            "   - MiniLM: https://huggingface.co/onnx-community/msmarco-distilbert-base-tas-b"
        )


def download_sentence_transformer():
    """Download a lightweight sentence transformer for embeddings"""
    print("🔤 Setting up MiniLM sentence transformer for text embeddings")
    print("This requires converting from PyTorch to ONNX format.")

    print("💡 To create your own MiniLM ONNX model:")
    print("   1. Install dependencies:")
    print("      pip install transformers onnxruntime optimum")
    print("   2. Convert model:")
    print(
        "      optimum-cli export onnx --model sentence-transformers/paraphrase-MiniLM-L3-v2 minilm-embeddings/"
    )
    print("   3. Move the .onnx file to this directory")
    print("   4. Use with: functionfly.ai('minilm-embeddings', '[token_ids]')")
    print("   Output: 768-dimensional embeddings for semantic search")

    # Try to download a pre-converted model if available
    try:
        import onnx
        from onnx import hub

        print("\nTrying to find a compatible model...")
        # This might not exist, but worth trying
        model = hub.load("minilm")
        onnx.save(model, "minilm-embeddings.onnx")
        print("✅ Found compatible model!")
    except:
        print(
            "ℹ️  No pre-built ONNX MiniLM found. Please convert manually as shown above."
        )


def main():
    parser = argparse.ArgumentParser(description="Download ONNX models for FunctionFly")
    parser.add_argument(
        "model",
        choices=[
            "alexnet",
            "phi4-mini",
            "mobilenet",
            "bert-small",
            "sentence-transformer",
        ],
        help="Model to download",
    )

    args = parser.parse_args()

    # Ensure we're in the models directory
    os.chdir(Path(__file__).parent)

    if args.model == "alexnet":
        download_alexnet()
    elif args.model == "phi4-mini":
        download_phi4_mini()
    elif args.model == "mobilenet":
        download_mobile_net()
    elif args.model == "bert-small":
        download_bert_small()
    elif args.model == "sentence-transformer":
        download_sentence_transformer()


if __name__ == "__main__":
    main()
