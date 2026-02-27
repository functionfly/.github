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
"""

import os
import sys
import argparse
from pathlib import Path

def download_alexnet():
    """Download AlexNet for image classification"""
    print("📸 Downloading AlexNet for image classification")
    print("Classic convolutional neural network for 1000-class image recognition.")

    try:
        import onnx
        from onnx import hub

        print("Using ONNX hub to download AlexNet...")
        model = hub.load("AlexNet")
        onnx.save(model, "alexnet.onnx")

        print("✅ Download complete!"
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
    print("This model offers excellent reasoning and math capabilities for serverless AI.")
    print("Note: This is a large model. Consider using a smaller model for testing first.")

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
        onnx.save(model, "mobilenetv3-small.onnx")

        print("✅ Download complete!")
        print("📁 Model saved as: mobilenetv3-small.onnx")
        print("💡 Use with: functionfly.ai('mobilenetv3-small', '[normalized_pixel_values]')")
        print("   Input shape: [1, 3, 224, 224] (normalized RGB image)")
        print("   Output: 1000 class probabilities")

    except ImportError:
        print("❌ ONNX not installed. Install with: pip install onnx")
        print("💡 Then run: python -c \"from onnx import hub; model = hub.load('mobilenetv3-small'); import onnx; onnx.save(model, 'mobilenetv3-small.onnx')\"")
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
        onnx.save(model, "bert-tiny.onnx")

        print("✅ Download complete!")
        print("📁 Model saved as: bert-tiny.onnx")
        print("💡 Use with: functionfly.ai('bert-tiny', '[token_ids]')")
        print("   Note: Requires tokenizer preprocessing")
        print("   pip install transformers")
        print("   from transformers import AutoTokenizer; tokenizer = AutoTokenizer.from_pretrained('prajjwal1/bert-tiny')")

    except ImportError:
        print("❌ ONNX not installed. Install with: pip install onnx")
    except Exception as e:
        print(f"❌ Download failed: {e}")
        print("💡 Try: pip install optimum && optimum-cli export onnx --model prajjwal1/bert-tiny bert-tiny/")
        print("💡 Alternative models:")
        print("   - BERT Tiny: https://huggingface.co/onnx-community/bert-tiny")
        print("   - MiniLM: https://huggingface.co/onnx-community/msmarco-distilbert-base-tas-b")

def download_sentence_transformer():
    """Download a lightweight sentence transformer for embeddings"""
    print("🔤 Setting up MiniLM sentence transformer for text embeddings")
    print("This requires converting from PyTorch to ONNX format.")

    print("💡 To create your own MiniLM ONNX model:")
    print("   1. Install dependencies:")
    print("      pip install transformers onnxruntime optimum")
    print("   2. Convert model:")
    print("      optimum-cli export onnx --model sentence-transformers/paraphrase-MiniLM-L3-v2 minilm-embeddings/")
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
        print("ℹ️  No pre-built ONNX MiniLM found. Please convert manually as shown above.")

def main():
    parser = argparse.ArgumentParser(description="Download ONNX models for FunctionFly")
    parser.add_argument(
        "model",
        choices=["alexnet", "phi4-mini", "mobilenet", "bert-small", "sentence-transformer"],
        help="Model to download"
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