"""Atlas OpenTelemetry exporter.

This module exports Atlas events as OpenTelemetry spans for integration
with existing observability infrastructure.
"""

import json
import logging
from dataclasses import dataclass
from typing import Any, Dict, List, Optional

logger = logging.getLogger(__name__)

try:
    from opentelemetry import trace
    from opentelemetry.sdk.trace import TracerProvider
    from opentelemetry.sdk.trace.export import BatchSpanProcessor
    from opentelemetry.sdk.resources import Resource
    from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
    from opentelemetry.trace import SpanKind, Status, StatusCode
    OTEL_AVAILABLE = True
except ImportError:
    OTEL_AVAILABLE = False
    logger.warning("OpenTelemetry not available, Atlas OTEL export disabled")


@dataclass
class AtlasEventConverter:
    """Converts Atlas events to OpenTelemetry spans."""

    service_name: str = "flymind-ai"

    def event_to_span(self, event: Dict[str, Any]) -> Dict[str, Any]:
        """Convert Atlas event to OTLP span.

        Args:
            event: Atlas event dictionary

        Returns:
            Span data dictionary
        """
        kind = event.get("kind", "")
        payload = event.get("payload", {})

        kind_map = {
            "INPUT": "server",
            "DECISION": "internal",
            "ACTION": "client",
            "RESULT": "server",
            "ERROR": "internal",
        }

        span = {
            "name": f"atlas.{kind.lower()}",
            "kind": kind_map.get(kind, "internal"),
            "attributes": {
                "atlas.run_id": event.get("run_id", ""),
                "atlas.event_id": event.get("event_id", ""),
                "atlas.sequence": event.get("sequence", 0),
                "atlas.system_id": event.get("system_id", ""),
            },
            "start_time_unix_nano": event.get("timestamp_ns", 0),
            "end_time_unix_nano": event.get("timestamp_ns", 0) + (payload.get("cost", {}).get("latency_ms", 0) * 1_000_000),
        }

        if "cost" in payload:
            span["attributes"]["cost.usd"] = payload["cost"].get("cost_usd", 0)
            span["attributes"]["cost.input_tokens"] = payload["cost"].get("input_tokens", 0)
            span["attributes"]["cost.output_tokens"] = payload["cost"].get("output_tokens", 0)

        if kind == "DECISION" and "model" in payload:
            span["attributes"]["ai.model"] = payload["model"]
            if "tool_call" in payload:
                span["attributes"]["ai.tool_call"] = payload["tool_call"].get("name", "")

        if kind == "ERROR":
            span["status"] = {"code": "error", "message": payload.get("error", "")}

        return span

    def export_spans(self, events: List[Dict[str, Any]], otlp_endpoint: str) -> None:
        """Export Atlas events as OTLP spans.

        Args:
            events: List of Atlas events
            otlp_endpoint: OTLP endpoint URL
        """
        if not OTEL_AVAILABLE or not otlp_endpoint:
            return

        try:
            exporter = OTLPSpanExporter(endpoint=otlp_endpoint)
            provider = TracerProvider()
            provider.add_span_processor(BatchSpanProcessor(exporter))
            trace.set_tracer_provider(provider)

            tracer = trace.get_tracer(self.service_name)

            for event in events:
                span_data = self.event_to_span(event)

                with tracer.start_as_current_span(
                    span_data["name"],
                    kind=getattr(SpanKind, span_data["kind"].upper(), SpanKind.INTERNAL),
                    attributes=span_data["attributes"],
                ) as span:
                    if "status" in span_data:
                        span.set_status(Status(StatusCode.ERROR, span_data["status"]["message"]))

        except Exception as e:
            logger.error(f"Failed to export spans to OTEL: {e}")


class AtlasOTelExporter:
    """Exports Atlas events as OpenTelemetry spans."""

    def __init__(self, service_name: str = "flymind-ai", otlp_endpoint: Optional[str] = None):
        self.service_name = service_name
        self.otlp_endpoint = otlp_endpoint
        self.converter = AtlasEventConverter(service_name)
        self._tracer = None

        if OTEL_AVAILABLE and otlp_endpoint:
            self._init_otel()

    def _init_otel(self) -> None:
        """Initialize OpenTelemetry with OTLP exporter."""
        try:
            exporter = OTLPSpanExporter(endpoint=self.otlp_endpoint)
            provider = TracerProvider()
            provider.add_span_processor(BatchSpanProcessor(exporter))
            trace.set_tracer_provider(provider)
            self._tracer = trace.get_tracer(self.service_name)
            logger.info(f"Atlas OTEL exporter initialized with endpoint: {self.otlp_endpoint}")
        except Exception as e:
            logger.error(f"Failed to initialize OTEL: {e}")

    def record_event(self, event: Dict[str, Any]) -> None:
        """Record Atlas event as OTEL span.

        Args:
            event: Atlas event dictionary
        """
        if not self._tracer:
            return

        span_data = self.converter.event_to_span(event)

        try:
            with self._tracer.start_as_current_span(
                span_data["name"],
                kind=getattr(SpanKind, span_data["kind"].upper(), SpanKind.INTERNAL),
                attributes=span_data["attributes"],
            ) as span:
                if "status" in span_data:
                    span.set_status(Status(StatusCode.ERROR, span_data["status"]["message"]))
        except Exception as e:
            logger.error(f"Failed to record event: {e}")

    def record_cost(self, run_id: str, cost_data: Dict[str, Any]) -> None:
        """Record cost metrics as OTEL metrics.

        Args:
            run_id: Atlas run ID
            cost_data: Cost data dictionary
        """
        pass

    def export_batch(self, events: List[Dict[str, Any]]) -> None:
        """Export a batch of events.

        Args:
            events: List of Atlas events
        """
        if not self.otlp_endpoint:
            return

        self.converter.export_spans(events, self.otlp_endpoint)


_atlas_otel_exporter: Optional[AtlasOTelExporter] = None


def get_atlas_otel_exporter() -> AtlasOTelExporter:
    """Get or create the global Atlas OTEL exporter.

    Returns:
        AtlasOTelExporter instance
    """
    global _atlas_otel_exporter
    if _atlas_otel_exporter is None:
        from ..config import settings
        otlp_endpoint = getattr(settings, 'atlas_otel_endpoint', None)
        enabled = getattr(settings, 'atlas_otel_exporter_enabled', False)
        if enabled and otlp_endpoint:
            _atlas_otel_exporter = AtlasOTelExporter(
                service_name="flymind-ai",
                otlp_endpoint=otlp_endpoint,
            )
        else:
            _atlas_otel_exporter = AtlasOTelExporter()
    return _atlas_otel_exporter