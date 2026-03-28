"""Tests for AI Gateway API endpoints."""

import base64
import pytest
from fastapi.testclient import TestClient

from src.main import app


@pytest.fixture
def client():
    """Create test client."""
    return TestClient(app)


@pytest.fixture
def valid_encoded_input():
    """Create valid base64-encoded input."""
    return base64.b64encode(b"Hello, World!").decode("utf-8")


class TestHealthEndpoints:
    """Tests for health check endpoints."""

    def test_health_check(self, client):
        """Test /health endpoint returns healthy status."""
        response = client.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
        assert "version" in data
        assert "timestamp" in data

    def test_root_endpoint(self, client):
        """Test root endpoint returns service info."""
        response = client.get("/")
        assert response.status_code == 200
        data = response.json()
        assert data["service"] == "ai-gateway"
        assert "version" in data
        assert data["status"] == "running"


class TestInferenceSchemas:
    """Tests for inference request/response schemas."""

    def test_valid_inference_request(self, valid_encoded_input):
        """Test valid inference request schema."""
        from src.models.schemas import InferenceRequest

        request = InferenceRequest(
            model="onnx://phi-3-mini",
            input=valid_encoded_input,
            parameters={"temperature": 0.7, "max_tokens": 100},
        )
        assert request.model == "onnx://phi-3-mini"
        assert request.input == valid_encoded_input
        assert request.parameters["temperature"] == 0.7

    def test_invalid_base64_input(self):
        """Test invalid base64 input raises validation error."""
        from src.models.schemas import InferenceRequest

        with pytest.raises(ValueError):
            InferenceRequest(
                model="onnx://phi-3-mini",
                input="not-valid-base64!!!",
            )

    def test_inference_parameters_validation(self):
        """Test inference parameters are validated."""
        from src.models.schemas import InferenceParameters

        params = InferenceParameters(
            temperature=1.5,
            max_tokens=1000,
            top_p=0.95,
        )
        assert params.temperature == 1.5
        assert params.max_tokens == 1000

    def test_inference_parameters_out_of_range(self):
        """Test parameters out of range are rejected."""
        from src.models.schemas import InferenceParameters
        from pydantic import ValidationError

        with pytest.raises(ValidationError):
            InferenceParameters(temperature=5.0)  # > 2.0


class TestInferenceResponse:
    """Tests for inference response schema."""

    def test_inference_response_creation(self):
        """Test inference response schema."""
        from src.models.schemas import InferenceResponse, ModelProvider, InferenceBackend

        response = InferenceResponse(
            output=base64.b64encode(b"Test output").decode("utf-8"),
            latency_ms=45.5,
            cost_usd=0.002,
            model="onnx://phi-3-mini",
            provider=ModelProvider.ONNX,
            backend=InferenceBackend.ONNX_RUNTIME,
            request_id="test-123",
        )
        assert response.latency_ms == 45.5
        assert response.cost_usd == 0.002
        assert response.tokens_generated is None


class TestClusterHealth:
    """Tests for cluster health info."""

    def test_cluster_health_info(self):
        """Test cluster health info schema."""
        from src.models.schemas import ClusterHealthInfo, HealthStatus

        health = ClusterHealthInfo(
            cluster_id="us-east-1-a100",
            region="us-east-1",
            status=HealthStatus.HEALTHY,
            healthy_instances=4,
            total_instances=8,
            avg_latency_ms=50.0,
            error_rate=0.01,
        )
        assert health.status == HealthStatus.HEALTHY
        assert health.healthy_instances == 4
        assert health.total_instances == 8


class TestErrorResponse:
    """Tests for error response schema."""

    def test_error_response(self):
        """Test error response schema."""
        from src.models.schemas import ErrorResponse

        error = ErrorResponse(
            error="validation_error",
            message="Invalid input",
            request_id="test-456",
        )
        assert error.error == "validation_error"
        assert error.request_id == "test-456"


class TestModelList:
    """Tests for model listing."""

    def test_model_list_response(self):
        """Test model list response schema."""
        from src.models.schemas import (
            ModelListResponse,
            ModelInfo,
            ModelProvider,
            InferenceBackend,
        )

        models = [
            ModelInfo(
                model_id="onnx://phi-3-mini",
                provider=ModelProvider.ONNX,
                backend=InferenceBackend.ONNX_RUNTIME,
                context_length=4096,
                supported_parameters=["temperature", "max_tokens"],
                is_available=True,
            ),
        ]
        response = ModelListResponse(
            models=models,
            default_model="phi-3-mini",
        )
        assert len(response.models) == 1
        assert response.default_model == "phi-3-mini"
