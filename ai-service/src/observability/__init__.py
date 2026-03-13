"""Observability Package for FlyMind AI Service.

This module provides full observability including:
- Prometheus metrics for all services
- Structured logging setup
- Distributed tracing (OpenTelemetry)
- Health check dependencies
"""

from .metrics import (
    MetricsCollector,
    get_metrics_collector,
    RequestMetrics,
    CacheMetrics,
    CostMetrics,
    ProviderMetrics,
)
from .logging import (
    StructuredLogger,
    get_logger,
    setup_logging,
)
from .tracing import (
    TracingConfig,
    get_tracer,
    setup_tracing,
    trace_function,
)
from .health import (
    HealthChecker,
    HealthStatus,
    ComponentHealth,
    get_health_checker,
)

__all__ = [
    "MetricsCollector",
    "get_metrics_collector",
    "RequestMetrics",
    "CacheMetrics",
    "CostMetrics",
    "ProviderMetrics",
    "StructuredLogger",
    "get_logger",
    "setup_logging",
    "TracingConfig",
    "get_tracer",
    "setup_tracing",
    "trace_function",
    "HealthChecker",
    "HealthStatus",
    "ComponentHealth",
    "get_health_checker",
]
