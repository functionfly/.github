"""Metrics collection for FlyMind AI Service.

This module provides Prometheus metrics for all services:
- Request latency (p50, p95, p99)
- Provider API costs
- Cache hit/miss rates
- Anomaly detection alerts
- Chat session metrics
- Search query performance
"""

import time
import os
from collections import defaultdict
from dataclasses import dataclass, field
from datetime import datetime
from typing import Any, Dict, List, Optional
import logging
import threading

logger = logging.getLogger(__name__)

# Try to import prometheus_client, but make it optional
try:
    from prometheus_client import Counter, Gauge, Histogram, Summary, CollectorRegistry, generate_latest
    PROMETHEUS_AVAILABLE = True
except ImportError:
    PROMETHEUS_AVAILABLE = False
    # Create dummy classes
    class Counter:
        def __init__(self, *args, **kwargs): pass
        def inc(self, *args, **kwargs): pass
        def labels(self, *args, **kwargs): return self

    class Gauge:
        def __init__(self, *args, **kwargs): pass
        def inc(self, *args, **kwargs): pass
        def dec(self, *args, **kwargs): pass
        def set(self, *args, **kwargs): pass
        def labels(self, *args, **kwargs): return self

    class Histogram:
        def __init__(self, *args, **kwargs): pass
        def observe(self, *args, **kwargs): pass
        def labels(self, *args, **kwargs): return self

    class Summary:
        def __init__(self, *args, **kwargs): pass
        def observe(self, *args, **kwargs): pass
        def labels(self, *args, **kwargs): return self

    class CollectorRegistry:
        pass

    def generate_latest(registry=None): return b""


@dataclass
class RequestMetrics:
    """Request metrics data."""
    endpoint: str
    method: str
    latency_ms: float
    status_code: int
    tenant_id: Optional[str] = None
    timestamp: datetime = field(default_factory=datetime.utcnow)


@dataclass
class CacheMetrics:
    """Cache metrics data."""
    hits: int = 0
    misses: int = 0
    sets: int = 0
    deletes: int = 0
    evictions: int = 0
    hit_rate: float = 0.0


@dataclass
class CostMetrics:
    """Cost metrics data."""
    provider: str
    model: str
    input_tokens: int
    output_tokens: int
    total_tokens: int
    cost_usd: float
    tenant_id: Optional[str] = None
    timestamp: datetime = field(default_factory=datetime.utcnow)


@dataclass
class ProviderMetrics:
    """Provider metrics data."""
    provider: str
    requests: int
    errors: int
    latency_ms: float
    avg_latency_ms: float = 0.0
    success_rate: float = 0.0


class MetricsCollector:
    """Collects and exposes Prometheus metrics."""

    def __init__(self, service_name: str = "flymind-ai"):
        """Initialize the metrics collector.

        Args:
            service_name: Name of the service
        """
        self._service_name = service_name
        self._logger = logging.getLogger(__name__)
        self._lock = threading.Lock()

        # In-memory metrics storage (for when Prometheus is not available)
        self._request_latencies: Dict[str, List[float]] = defaultdict(list)
        self._request_counts: Dict[str, int] = defaultdict(int)
        self._cache_hits = 0
        self._cache_misses = 0
        self._costs: List[CostMetrics] = []
        self._provider_metrics: Dict[str, ProviderMetrics] = {}
        self._chat_sessions = 0
        self._search_queries = 0

        # Initialize Prometheus metrics if available
        if PROMETHEUS_AVAILABLE:
            self._init_prometheus_metrics()
        else:
            self._logger.warning("Prometheus client not available, using in-memory metrics")
            self._prom_metrics = {}

    def _init_prometheus_metrics(self) -> None:
        """Initialize Prometheus metrics."""
        # Request metrics
        self._prom_metrics["requests_total"] = Counter(
            f"{self._service_name}_requests_total",
            "Total number of requests",
            ["endpoint", "method", "status"]
        )

        self._prom_metrics["request_latency"] = Histogram(
            f"{self._service_name}_request_latency_seconds",
            "Request latency in seconds",
            ["endpoint", "method"],
            buckets=[0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0]
        )

        # Cache metrics
        self._prom_metrics["cache_hits"] = Counter(
            f"{self._service_name}_cache_hits_total",
            "Total number of cache hits"
        )

        self._prom_metrics["cache_misses"] = Counter(
            f"{self._service_name}_cache_misses_total",
            "Total number of cache misses"
        )

        # Cost metrics
        self._prom_metrics["api_costs"] = Counter(
            f"{self._service_name}_api_costs_total",
            "Total API costs in USD",
            ["provider", "model", "tenant"]
        )

        self._prom_metrics["token_usage"] = Counter(
            f"{self._service_name}_token_usage_total",
            "Total token usage",
            ["provider", "model", "type", "tenant"]
        )

        # Provider metrics
        self._prom_metrics["provider_requests"] = Counter(
            f"{self._service_name}_provider_requests_total",
            "Total provider requests",
            ["provider", "status"]
        )

        self._prom_metrics["provider_latency"] = Histogram(
            f"{self._service_name}_provider_latency_seconds",
            "Provider request latency",
            ["provider"],
            buckets=[0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0]
        )

        # Chat metrics
        self._prom_metrics["chat_sessions"] = Counter(
            f"{self._service_name}_chat_sessions_total",
            "Total number of chat sessions",
            ["tenant"]
        )

        self._prom_metrics["chat_messages"] = Counter(
            f"{self._service_name}_chat_messages_total",
            "Total number of chat messages",
            ["tenant", "intent"]
        )

        # Search metrics
        self._prom_metrics["search_queries"] = Counter(
            f"{self._service_name}_search_queries_total",
            "Total number of search queries",
            ["tenant"]
        )

        self._prom_metrics["search_latency"] = Histogram(
            f"{self._service_name}_search_latency_seconds",
            "Search query latency",
            buckets=[0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0]
        )

        # Anomaly metrics
        self._prom_metrics["anomalies_detected"] = Counter(
            f"{self._service_name}_anomalies_detected_total",
            "Total number of anomalies detected",
            ["type", "severity"]
        )

        # Active connections
        self._prom_metrics["active_requests"] = Gauge(
            f"{self._service_name}_active_requests",
            "Number of active requests"
        )

    def record_request(self, metrics: RequestMetrics) -> None:
        """Record a request metric.

        Args:
            metrics: Request metrics
        """
        with self._lock:
            key = f"{metrics.method}:{metrics.endpoint}"
            self._request_latencies[key].append(metrics.latency_ms)
            self._request_counts[key] += 1

            # Keep only recent latencies
            if len(self._request_latencies[key]) > 1000:
                self._request_latencies[key] = self._request_latencies[key][-1000:]

            # Record to Prometheus if available
            if PROMETHEUS_AVAILABLE:
                self._prom_metrics["requests_total"].labels(
                    endpoint=metrics.endpoint,
                    method=metrics.method,
                    status=str(metrics.status_code)
                ).inc()

                self._prom_metrics["request_latency"].labels(
                    endpoint=metrics.endpoint,
                    method=metrics.method
                ).observe(metrics.latency_ms / 1000)  # Convert to seconds

    def record_cache_hit(self) -> None:
        """Record a cache hit."""
        with self._lock:
            self._cache_hits += 1

            if PROMETHEUS_AVAILABLE:
                self._prom_metrics["cache_hits"].inc()

    def record_cache_miss(self) -> None:
        """Record a cache miss."""
        with self._lock:
            self._cache_misses += 1

            if PROMETHEUS_AVAILABLE:
                self._prom_metrics["cache_misses"].inc()

    def record_cache_set(self) -> None:
        """Record a cache set."""
        pass  # Currently not tracked separately

    def record_cache_delete(self) -> None:
        """Record a cache delete."""
        pass  # Currently not tracked separately

    def record_cost(self, metrics: CostMetrics) -> None:
        """Record API cost.

        Args:
            metrics: Cost metrics
        """
        with self._lock:
            self._costs.append(metrics)

            # Keep only recent costs
            if len(self._costs) > 10000:
                self._costs = self._costs[-5000:]

            if PROMETHEUS_AVAILABLE:
                self._prom_metrics["api_costs"].labels(
                    provider=metrics.provider,
                    model=metrics.model,
                    tenant=metrics.tenant_id or "unknown"
                ).inc(metrics.cost_usd)

                self._prom_metrics["token_usage"].labels(
                    provider=metrics.provider,
                    model=metrics.model,
                    type="input",
                    tenant=metrics.tenant_id or "unknown"
                ).inc(metrics.input_tokens)

                self._prom_metrics["token_usage"].labels(
                    provider=metrics.provider,
                    model=metrics.model,
                    type="output",
                    tenant=metrics.tenant_id or "unknown"
                ).inc(metrics.output_tokens)

    def record_provider_request(
        self,
        provider: str,
        latency_ms: float,
        success: bool,
    ) -> None:
        """Record a provider request.

        Args:
            provider: Provider name
            latency_ms: Request latency
            success: Whether the request was successful
        """
        with self._lock:
            if provider not in self._provider_metrics:
                self._provider_metrics[provider] = ProviderMetrics(
                    provider=provider,
                    requests=0,
                    errors=0,
                    latency_ms=0.0,
                )

            m = self._provider_metrics[provider]
            m.requests += 1
            if not success:
                m.errors += 1
            m.latency_ms += latency_ms

            # Calculate running average
            m.avg_latency_ms = m.latency_ms / m.requests
            m.success_rate = (m.requests - m.errors) / m.requests

            if PROMETHEUS_AVAILABLE:
                status = "success" if success else "error"
                self._prom_metrics["provider_requests"].labels(
                    provider=provider,
                    status=status
                ).inc()

                self._prom_metrics["provider_latency"].labels(
                    provider=provider
                ).observe(latency_ms / 1000)

    def record_chat_session(self, tenant_id: Optional[str] = None) -> None:
        """Record a new chat session.

        Args:
            tenant_id: Tenant ID
        """
        with self._lock:
            self._chat_sessions += 1

            if PROMETHEUS_AVAILABLE:
                self._prom_metrics["chat_sessions"].labels(
                    tenant=tenant_id or "unknown"
                ).inc()

    def record_chat_message(
        self,
        tenant_id: Optional[str] = None,
        intent: str = "unknown",
    ) -> None:
        """Record a chat message.

        Args:
            tenant_id: Tenant ID
            intent: Message intent
        """
        if PROMETHEUS_AVAILABLE:
            self._prom_metrics["chat_messages"].labels(
                tenant=tenant_id or "unknown",
                intent=intent
            ).inc()

    def record_search_query(
        self,
        tenant_id: Optional[str] = None,
        latency_ms: float = 0.0,
    ) -> None:
        """Record a search query.

        Args:
            tenant_id: Tenant ID
            latency_ms: Query latency
        """
        with self._lock:
            self._search_queries += 1

            if PROMETHEUS_AVAILABLE:
                self._prom_metrics["search_queries"].labels(
                    tenant=tenant_id or "unknown"
                ).inc()

                self._prom_metrics["search_latency"].observe(latency_ms / 1000)

    def record_anomaly(
        self,
        anomaly_type: str,
        severity: str,
    ) -> None:
        """Record an anomaly detection.

        Args:
            anomaly_type: Type of anomaly
            severity: Severity level
        """
        if PROMETHEUS_AVAILABLE:
            self._prom_metrics["anomalies_detected"].labels(
                type=anomaly_type,
                severity=severity
            ).inc()

    def increment_active_requests(self) -> None:
        """Increment active request counter."""
        if PROMETHEUS_AVAILABLE:
            self._prom_metrics["active_requests"].inc()

    def decrement_active_requests(self) -> None:
        """Decrement active request counter."""
        if PROMETHEUS_AVAILABLE:
            self._prom_metrics["active_requests"].dec()

    def get_cache_metrics(self) -> CacheMetrics:
        """Get current cache metrics.

        Returns:
            CacheMetrics
        """
        with self._lock:
            total = self._cache_hits + self._cache_misses
            hit_rate = self._cache_hits / total if total > 0 else 0.0

            return CacheMetrics(
                hits=self._cache_hits,
                misses=self._cache_misses,
                hit_rate=hit_rate,
            )

    def get_cost_metrics(
        self,
        provider: Optional[str] = None,
        tenant_id: Optional[str] = None,
    ) -> List[CostMetrics]:
        """Get cost metrics.

        Args:
            provider: Optional provider filter
            tenant_id: Optional tenant filter

        Returns:
            List of CostMetrics
        """
        with self._lock:
            metrics = self._costs

            if provider:
                metrics = [m for m in metrics if m.provider == provider]
            if tenant_id:
                metrics = [m for m in metrics if m.tenant_id == tenant_id]

            return metrics

    def get_total_cost(self) -> float:
        """Get total cost in USD.

        Returns:
            Total cost
        """
        with self._lock:
            return sum(m.cost_usd for m in self._costs)

    def get_provider_metrics(self) -> Dict[str, ProviderMetrics]:
        """Get provider metrics.

        Returns:
            Dictionary of provider metrics
        """
        with self._lock:
            return self._provider_metrics.copy()

    def get_request_latency_percentiles(
        self,
        endpoint: Optional[str] = None,
    ) -> Dict[str, float]:
        """Get request latency percentiles.

        Args:
            endpoint: Optional endpoint filter

        Returns:
            Dictionary with p50, p95, p99
        """
        with self._lock:
            latencies = []

            if endpoint:
                for key, values in self._request_latencies.items():
                    if endpoint in key:
                        latencies.extend(values)
            else:
                for values in self._request_latencies.values():
                    latencies.extend(values)

            if not latencies:
                return {"p50": 0, "p95": 0, "p99": 0}

            latencies.sort()
            n = len(latencies)

            return {
                "p50": latencies[int(n * 0.5)],
                "p95": latencies[int(n * 0.95)],
                "p99": latencies[int(n * 0.99)],
            }

    def get_metrics_text(self) -> str:
        """Get Prometheus metrics as text.

        Returns:
            Metrics in Prometheus format
        """
        if PROMETHEUS_AVAILABLE:
            return generate_latest().decode("utf-8")

        # Fallback to simple text format
        lines = [
            "# Fallback metrics (Prometheus not available)",
            f"flymind_ai_cache_hits_total {self._cache_hits}",
            f"flymind_ai_cache_misses_total {self._cache_misses}",
            f"flymind_ai_chat_sessions_total {self._chat_sessions}",
            f"flymind_ai_search_queries_total {self._search_queries}",
        ]
        return "\n".join(lines)


# Global metrics collector
_metrics_collector: Optional[MetricsCollector] = None


def get_metrics_collector() -> MetricsCollector:
    """Get the global metrics collector instance.

    Returns:
        MetricsCollector instance
    """
    global _metrics_collector
    if _metrics_collector is None:
        service_name = os.environ.get("SERVICE_NAME", "flymind-ai")
        _metrics_collector = MetricsCollector(service_name)

    return _metrics_collector
