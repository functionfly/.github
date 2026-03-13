"""Intelligent Request Routing Service.

This module provides ML-based edge target selection for optimal routing.
"""

from .selector import RoutingService, get_routing_service
from .collector import LatencyCollector, get_latency_collector
from .models import EdgeScore, EdgeMetrics

__all__ = [
    "RoutingService",
    "get_routing_service",
    "LatencyCollector",
    "get_latency_collector",
    "EdgeScore",
    "EdgeMetrics",
]
