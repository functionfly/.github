"""Pydantic models for AI Gateway API requests and responses."""

from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional, Union

from pydantic import BaseModel, Field, field_validator


class ModelProvider(str, Enum):
    """Model provider types."""

    RUNPOD = "runpod"
    ONNX = "onnx"
    OPENAI = "openai"
    OPENROUTER = "openrouter"
    OLLAMA = "ollama"


class InferenceBackend(str, Enum):
    """Inference backend types."""

    RUNPOD_API = "runpod_api"
    RUNPOD_CLUSTER = "runpod_cluster"
    ONNX_RUNTIME = "onnx_runtime"
    OPENAI_API = "openai_api"
    OPENROUTER_API = "openrouter_api"
    OLLAMA_API = "ollama_api"


class InferenceRequest(BaseModel):
    """Request model for inference endpoint."""

    model: str = Field(
        ...,
        description="Model identifier (e.g., 'onnx://phi-3-mini', 'openai://gpt-4')",
        examples=["onnx://phi-3-mini", "openai://gpt-4", "runpod://phi-3-mini"],
    )
    input: str = Field(
        ...,
        description="Base64-encoded input data",
        examples=["SGVsbG8gV29ybGQh"],  # "Hello World!" base64
    )
    parameters: Optional[Dict[str, Any]] = Field(
        default=None,
        description="Inference parameters (temperature, max_tokens, etc.)",
        examples=[{"temperature": 0.7, "max_tokens": 100, "top_p": 0.9}],
    )
    tenant_id: Optional[str] = Field(
        default=None,
        description="Tenant ID for multi-tenant isolation",
        examples=["tenant-123"],
    )
    stream: bool = Field(
        default=False,
        description="Enable streaming response",
    )
    prefer_region: Optional[str] = Field(
        default=None,
        description="Preferred GPU region for inference",
        examples=["us-east-1", "eu-west-1"],
    )

    @field_validator("model")
    @classmethod
    def validate_model(cls, v: str) -> str:
        """Validate model identifier format."""
        if not v or len(v.strip()) == 0:
            raise ValueError("Model identifier cannot be empty")
        return v.strip()

    @field_validator("input")
    @classmethod
    def validate_input(cls, v: str) -> str:
        """Validate base64 input."""
        import base64

        try:
            base64.b64decode(v, validate=True)
        except Exception as e:
            raise ValueError(f"Invalid base64 encoding: {e}")
        return v


class StreamInferenceRequest(InferenceRequest):
    """Request model for streaming inference."""

    stream: bool = Field(default=True, description="Must be True for streaming")


class InferenceParameters(BaseModel):
    """Inference parameters model."""

    temperature: Optional[float] = Field(
        default=0.7,
        ge=0.0,
        le=2.0,
        description="Sampling temperature",
    )
    max_tokens: Optional[int] = Field(
        default=256,
        ge=1,
        le=8192,
        description="Maximum tokens to generate",
    )
    top_p: Optional[float] = Field(
        default=0.9,
        ge=0.0,
        le=1.0,
        description="Nucleus sampling probability",
    )
    top_k: Optional[int] = Field(
        default=50,
        ge=1,
        description="Top-k sampling parameter",
    )
    repeat_penalty: Optional[float] = Field(
        default=1.1,
        ge=1.0,
        le=2.0,
        description="Repetition penalty",
    )
    stop: Optional[List[str]] = Field(
        default=None,
        description="Stop sequences",
    )


class InferenceResponse(BaseModel):
    """Response model for inference endpoint."""

    output: str = Field(
        ...,
        description="Base64-encoded output data",
    )
    latency_ms: float = Field(
        ...,
        ge=0,
        description="Inference latency in milliseconds",
    )
    cost_usd: float = Field(
        ...,
        ge=0,
        description="Estimated cost in USD",
    )
    model: str = Field(
        ...,
        description="Model used for inference",
    )
    provider: ModelProvider = Field(
        ...,
        description="Model provider",
    )
    backend: InferenceBackend = Field(
        ...,
        description="Inference backend used",
    )
    tokens_generated: Optional[int] = Field(
        default=None,
        ge=0,
        description="Number of tokens generated",
    )
    region: Optional[str] = Field(
        default=None,
        description="Region used for inference",
    )
    request_id: str = Field(
        ...,
        description="Unique request identifier",
    )
    timestamp: datetime = Field(
        default_factory=datetime.utcnow,
        description="Response timestamp",
    )


class BatchInferenceRequest(BaseModel):
    """Request model for batch inference."""

    requests: List[InferenceRequest] = Field(
        ...,
        min_length=1,
        max_length=8,
        description="List of inference requests (max 8)",
    )


class BatchInferenceResponse(BaseModel):
    """Response model for batch inference."""

    outputs: List[InferenceResponse] = Field(
        ...,
        description="List of inference responses",
    )
    total_latency_ms: float = Field(
        ...,
        ge=0,
        description="Total batch processing latency in milliseconds",
    )
    total_cost_usd: float = Field(
        ...,
        ge=0,
        description="Total estimated cost in USD",
    )
    batch_size: int = Field(
        ...,
        ge=1,
        description="Number of requests in batch",
    )


class HealthStatus(str, Enum):
    """Health status values."""

    HEALTHY = "healthy"
    DEGRADED = "degraded"
    UNHEALTHY = "unhealthy"


class HealthResponse(BaseModel):
    """Response model for health endpoint."""

    status: HealthStatus = Field(
        ...,
        description="Service health status",
    )
    version: str = Field(
        ...,
        description="Service version",
    )
    timestamp: datetime = Field(
        default_factory=datetime.utcnow,
        description="Health check timestamp",
    )


class ClusterHealthInfo(BaseModel):
    """Health information for a cluster."""

    cluster_id: str = Field(..., description="Cluster identifier")
    region: str = Field(..., description="Cluster region")
    status: HealthStatus = Field(..., description="Cluster status")
    healthy_instances: int = Field(..., ge=0, description="Healthy instance count")
    total_instances: int = Field(..., ge=0, description="Total instance count")
    avg_latency_ms: float = Field(..., ge=0, description="Average latency in ms")
    error_rate: float = Field(..., ge=0, le=1, description="Error rate (0-1)")


class ReadinessResponse(BaseModel):
    """Response model for readiness endpoint."""

    status: HealthStatus = Field(..., description="Service readiness status")
    clusters: List[ClusterHealthInfo] = Field(
        default_factory=list, description="Cluster health information"
    )
    total_clusters: int = Field(..., ge=0, description="Total number of clusters")
    healthy_clusters: int = Field(..., ge=0, description="Number of healthy clusters")
    version: str = Field(..., description="Service version")
    timestamp: datetime = Field(
        default_factory=datetime.utcnow, description="Readiness check timestamp"
    )


class MetricsLabels(BaseModel):
    """Prometheus metric labels."""

    model: Optional[str] = None
    provider: Optional[str] = None
    backend: Optional[str] = None
    region: Optional[str] = None
    tenant_id: Optional[str] = None
    status_code: Optional[str] = None


class MetricsResponse(BaseModel):
    """Response model for Prometheus metrics."""

    metrics: List[str] = Field(
        ...,
        description="Prometheus-formatted metrics",
    )
    content_type: str = Field(
        default="text/plain; charset=utf-8",
        description="Content-Type header value",
    )


class ErrorResponse(BaseModel):
    """Error response model."""

    error: str = Field(..., description="Error type")
    message: str = Field(..., description="Error message")
    request_id: Optional[str] = Field(default=None, description="Request identifier")
    details: Optional[Dict[str, Any]] = Field(
        default=None, description="Additional error details"
    )
    timestamp: datetime = Field(
        default_factory=datetime.utcnow, description="Error timestamp"
    )


class TokenUsage(BaseModel):
    """Token usage information."""

    prompt_tokens: int = Field(..., ge=0, description="Tokens in prompt")
    completion_tokens: int = Field(..., ge=0, description="Tokens in completion")
    total_tokens: int = Field(..., ge=0, description="Total tokens")


class ModelInfo(BaseModel):
    """Information about a supported model."""

    model_id: str = Field(..., description="Model identifier")
    provider: ModelProvider = Field(..., description="Model provider")
    backend: InferenceBackend = Field(..., description="Inference backend")
    context_length: int = Field(..., ge=1, description="Maximum context length")
    supported_parameters: List[str] = Field(
        ..., description="Supported inference parameters"
    )
    is_available: bool = Field(..., description="Model availability status")


class ModelListResponse(BaseModel):
    """Response model for listing available models."""

    models: List[ModelInfo] = Field(..., description="List of available models")
    default_model: str = Field(..., description="Default model identifier")
