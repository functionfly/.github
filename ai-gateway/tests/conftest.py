"""Pytest configuration and fixtures."""

import base64
import os
import sys

import pytest

# Add src to path for imports
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

# Set test environment variables before importing app
os.environ.setdefault("RUNPOD_API_KEY", "test-key")
os.environ.setdefault("RUNPOD_CLUSTER_URL", "http://localhost:8080")
os.environ.setdefault("DEFAULT_MODEL", "phi-3-mini")


@pytest.fixture
def sample_encoded_input():
    """Create sample base64-encoded input."""
    return base64.b64encode(b"Hello, AI Gateway!").decode("utf-8")


@pytest.fixture
def sample_inference_request(sample_encoded_input):
    """Create sample inference request data."""
    return {
        "model": "onnx://phi-3-mini",
        "input": sample_encoded_input,
        "parameters": {
            "temperature": 0.7,
            "max_tokens": 100,
            "top_p": 0.9,
        },
    }


@pytest.fixture
def mock_settings():
    """Create mock settings for testing."""
    from src.config import Settings

    return Settings(
        HOST="127.0.0.1",
        PORT=8082,
        DEBUG=True,
        RUNPOD_API_KEY="test-key",
        RUNPOD_CLUSTER_URL="http://localhost:8080",
        DEFAULT_MODEL="phi-3-mini",
        MAX_CONTEXT_LENGTH=4096,
        API_KEY_HEADER="X-API-Key",
        TENANT_ID_HEADER="X-Tenant-ID",
        REQUIRED_API_KEY=None,
        MAX_BATCH_SIZE=8,
        BATCH_TIMEOUT_MS=100,
        INFERENCE_BACKEND="onnx",
    )
