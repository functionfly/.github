"""Distributed tracing for FlyMind AI Service.

This module provides OpenTelemetry tracing (optional).
"""

import functools
import os
import time
import uuid
from contextlib import contextmanager
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import Any, Callable, Dict, Generator, List, Optional
import logging

logger = logging.getLogger(__name__)

# Try to import OpenTelemetry, but make it optional
try:
    from opentelemetry import trace
    from opentelemetry.sdk.trace import TracerProvider
    from opentelemetry.sdk.trace.export import BatchSpanProcessor
    from opentelemetry.sdk.resources import Resource
    from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
    from opentelemetry.trace import Status, StatusCode
    OPENTELEMETRY_AVAILABLE = True
except ImportError:
    OPENTELEMETRY_AVAILABLE = False


class SpanKind(str, Enum):
    """Span kinds."""
    INTERNAL = "internal"
    SERVER = "server"
    CLIENT = "client"
    PRODUCER = "producer"
    CONSUMER = "consumer"


@dataclass
class Span:
    """A trace span (for when OpenTelemetry is not available)."""
    name: str
    trace_id: str
    span_id: str
    parent_id: Optional[str] = None
    kind: SpanKind = SpanKind.INTERNAL
    start_time: datetime = field(default_factory=datetime.utcnow)
    end_time: Optional[datetime] = None
    attributes: Dict[str, Any] = field(default_factory=dict)
    status: str = "ok"
    events: List[Dict[str, Any]] = field(default_factory=list)

    def set_attribute(self, key: str, value: Any) -> None:
        """Set a span attribute."""
        self.attributes[key] = value

    def add_event(self, name: str, attributes: Optional[Dict[str, Any]] = None) -> None:
        """Add an event to the span."""
        self.events.append({
            "name": name,
            "timestamp": datetime.utcnow().isoformat(),
            "attributes": attributes or {},
        })

    def set_status(self, status: str) -> None:
        """Set span status."""
        self.status = status

    def end(self) -> None:
        """End the span."""
        self.end_time = datetime.utcnow()


@dataclass
class TracingConfig:
    """Configuration for tracing."""
    service_name: str = "flymind-ai"
    enabled: bool = True
    otlp_endpoint: Optional[str] = None
    sample_rate: float = 1.0  # 0.0 to 1.0
    include_attributes: bool = True


class Tracer:
    """Tracer for creating spans."""

    def __init__(self, config: TracingConfig):
        """Initialize the tracer.

        Args:
            config: Tracing configuration
        """
        self._config = config
        self._logger = logging.getLogger(__name__)
        self._spans: List[Span] = []
        self._current_span: Optional[Span] = None

        # Initialize OpenTelemetry if available and configured
        self._otel_tracer = None
        if OPENTELEMETRY_AVAILABLE and config.enabled:
            try:
                self._init_opentelemetry()
            except Exception as e:
                self._logger.warning(f"Failed to initialize OpenTelemetry: {e}")

    def _init_opentelemetry(self) -> None:
        """Initialize OpenTelemetry."""
        resource = Resource.create({
            "service.name": self._config.service_name,
        })

        provider = TracerProvider(resource=resource)

        if self._config.otlp_endpoint:
            exporter = OTLPSpanExporter(endpoint=self._config.otlp_endpoint)
            provider.add_span_processor(BatchSpanProcessor(exporter))

        trace.set_tracer_provider(provider)
        self._otel_tracer = trace.get_tracer(self._config.service_name)

        self._logger.info("OpenTelemetry tracing initialized")

    @contextmanager
    def start_span(
        self,
        name: str,
        kind: SpanKind = SpanKind.INTERNAL,
        attributes: Optional[Dict[str, Any]] = None,
    ) -> Generator[Span, None, None]:
        """Start a new span.

        Args:
            name: Span name
            kind: Span kind
            attributes: Initial attributes

        Yields:
            Span
        """
        trace_id = uuid.uuid4().hex[:16]
        span_id = uuid.uuid4().hex[:8]

        parent_id = None
        if self._current_span:
            parent_id = self._current_span.span_id

        span = Span(
            name=name,
            trace_id=trace_id,
            span_id=span_id,
            parent_id=parent_id,
            kind=kind,
        )

        if attributes:
            for key, value in attributes.items():
                span.set_attribute(key, value)

        # Set as current
        old_span = self._current_span
        self._current_span = span

        # Use OpenTelemetry if available
        if self._otel_tracer:
            otel_span = self._otel_tracer.start_span(name)
            otel_span.set_attributes(attributes or {})

        try:
            yield span
        except Exception as e:
            span.set_status("error")
            span.set_attribute("error", True)
            span.set_attribute("error.message", str(e))
            raise
        finally:
            span.end()
            self._current_span = old_span

            if self._otel_tracer:
                otel_span.end()

    def trace(
        self,
        name: str,
        kind: SpanKind = SpanKind.INTERNAL,
    ) -> Callable:
        """Decorator to trace a function.

        Args:
            name: Span name
            kind: Span kind

        Returns:
            Decorator function
        """
        def decorator(func: Callable) -> Callable:
            @functools.wraps(func)
            def wrapper(*args, **kwargs):
                with self.start_span(name, kind):
                    return func(*args, **kwargs)
            return wrapper
        return decorator

    def get_current_span(self) -> Optional[Span]:
        """Get the current span."""
        return self._current_span


# Global tracer
_tracer: Optional[Tracer] = None
_tracing_config: Optional[TracingConfig] = None


def setup_tracing(config: Optional[TracingConfig] = None) -> Tracer:
    """Setup tracing.

    Args:
        config: Tracing configuration

    Returns:
        Configured Tracer
    """
    global _tracer, _tracing_config

    if config is None:
        config = TracingConfig(
            service_name=os.environ.get("SERVICE_NAME", "flymind-ai"),
            enabled=os.environ.get("TRACING_ENABLED", "true").lower() == "true",
            otlp_endpoint=os.environ.get("OTLP_ENDPOINT"),
            sample_rate=float(os.environ.get("TRACING_SAMPLE_RATE", "1.0")),
        )

    _tracing_config = config
    _tracer = Tracer(config)

    return _tracer


def get_tracer() -> Tracer:
    """Get the global tracer.

    Returns:
        Tracer instance
    """
    global _tracer

    if _tracer is None:
        _tracer = setup_tracing()

    return _tracer


def trace_function(
    name: Optional[str] = None,
    kind: SpanKind = SpanKind.INTERNAL,
) -> Callable:
    """Decorator to trace a function.

    Args:
        name: Span name (defaults to function name)
        kind: Span kind

    Returns:
        Decorator function
    """
    tracer = get_tracer()

    def decorator(func: Callable) -> Callable:
        span_name = name or func.__name__

        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            with tracer.start_span(span_name, kind):
                return func(*args, **kwargs)
        return wrapper
    return decorator
