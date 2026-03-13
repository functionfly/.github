"""Core routing algorithm for intelligent edge selection.

Implements ML-based routing using weighted factors:
- Geographic proximity
- Historical latency data
- Current load/availability
- Function characteristics
"""

import logging
from datetime import datetime
from typing import Dict, List, Optional, Tuple

from ...config import settings
from ...models.schemas import (
    EdgeProvider,
    EdgeStatus,
    RoutingDecision,
    RoutingDecisionRequest,
)
from .collector import LatencyCollector, get_latency_collector
from .models import EdgeMetrics, EdgeScore, calculate_distance, normalize_value

logger = logging.getLogger(__name__)


class RoutingService:
    """ML-based routing service for optimal edge selection.

    Uses weighted factors to score available edges:
    - latency_pct (30%): Lower latency is better
    - load_pct (30%): Lower load is better
    - availability (40%): Higher availability is better

    Falls back to geographically nearest edge when data is insufficient.
    """

    # Default metrics for when no data is available
    DEFAULT_LATENCY_MS = {
        EdgeProvider.CLOUDFLARE: 50.0,
        EdgeProvider.VERCEL: 45.0,
        EdgeProvider.FLY: 40.0,
        EdgeProvider.DENO: 55.0,
        EdgeProvider.FUNCTIONFLY: 35.0,  # FunctionFly edge
    }

    # Default load percentages
    DEFAULT_LOAD_PERCENT = 50.0

    def __init__(self):
        self._latency_collector = get_latency_collector()

        # Routing weights from config
        self._latency_weight = settings.routing_latency_weight
        self._load_weight = settings.routing_load_weight
        self._availability_weight = settings.routing_availability_weight

    async def decide_routing(
        self,
        request: RoutingDecisionRequest,
    ) -> RoutingDecision:
        """Determine optimal edge for a request.

        Args:
            request: Routing decision request with user location and function info

        Returns:
            RoutingDecision with recommended edge and alternatives
        """
        # Get metrics for all edges
        all_metrics = await self._get_all_metrics()

        # Score each edge
        scores = self._score_edges(
            all_metrics,
            user_lat=request.user_latitude,
            user_lon=request.user_longitude,
        )

        if not scores:
            # Fallback to default
            return self._fallback_decision(request.function_id)

        # Sort by score (highest first)
        sorted_scores = sorted(scores, key=lambda s: s.score, reverse=True)

        best = sorted_scores[0]
        alternatives = [s.provider for s in sorted_scores[1:4]]

        # Calculate estimated latency
        latency_estimate = self._estimate_latency(
            best.provider,
            all_metrics.get(best.provider),
            request.user_latitude,
            request.user_longitude,
        )

        return RoutingDecision(
            function_id=request.function_id,
            recommended_edge=best.provider,
            confidence=best.score,
            reasoning=best.reasoning,
            alternatives=alternatives[:3],
            latency_estimate_ms=latency_estimate,
        )

    async def _get_all_metrics(
        self,
    ) -> Dict[EdgeProvider, EdgeMetrics]:
        """Get metrics for all edges.

        Returns:
            Dictionary of edge to metrics
        """
        try:
            return await self._latency_collector.get_all_edge_metrics()
        except Exception as e:
            logger.warning(f"Failed to get edge metrics: {e}")
            return {}

    def _score_edges(
        self,
        metrics: Dict[EdgeProvider, EdgeMetrics],
        user_lat: Optional[float] = None,
        user_lon: Optional[float] = None,
    ) -> List[EdgeScore]:
        """Score all available edges.

        Args:
            metrics: Edge metrics dictionary
            user_lat: User latitude for geo scoring
            user_lon: User longitude for geo scoring

        Returns:
            List of EdgeScore sorted by score
        """
        scores = []

        for edge in EdgeProvider:
            try:
                score = self._score_edge(
                    edge,
                    metrics.get(edge),
                    user_lat,
                    user_lon,
                )
                scores.append(score)
            except Exception as e:
                logger.warning(f"Failed to score edge {edge.value}: {e}")

        return scores

    def _score_edge(
        self,
        edge: EdgeProvider,
        metrics: Optional[EdgeMetrics],
        user_lat: Optional[float],
        user_lon: Optional[float],
    ) -> EdgeScore:
        """Score a single edge.

        Args:
            edge: The edge provider
            metrics: Edge metrics (may be None)
            user_lat: User latitude
            user_lon: User longitude

        Returns:
            EdgeScore for the edge
        """
        # Get latency score (inverse - lower latency = higher score)
        if metrics and metrics.avg_latency_ms > 0:
            latency = metrics.avg_latency_ms
        else:
            latency = self.DEFAULT_LATENCY_MS.get(edge, 50.0)

        # Normalize latency (0-200ms range typical)
        latency_score = 1.0 - normalize_value(latency, 0, 200)

        # Get load score (inverse - lower load = higher score)
        if metrics:
            load = metrics.current_load_percent
        else:
            load = self.DEFAULT_LOAD_PERCENT

        load_score = 1.0 - normalize_value(load, 0, 100)

        # Get availability score
        if metrics:
            availability_score = 1.0 if metrics.available else 0.0
        else:
            availability_score = 0.8  # Default to mostly available

        # If user location provided, adjust score based on geography
        if user_lat is not None and user_lon is not None:
            geo_bonus = self._calculate_geo_bonus(edge, user_lat, user_lon)
            # Apply small geo bonus (max 10%)
            latency_score = min(1.0, latency_score + geo_bonus * 0.1)

        return EdgeScore(
            provider=edge,
            latency_score=latency_score,
            load_score=load_score,
            availability_score=availability_score,
            latency_weight=self._latency_weight,
            load_weight=self._load_weight,
            availability_weight=self._availability_weight,
        )

    def _calculate_geo_bonus(
        self,
        edge: EdgeProvider,
        user_lat: float,
        user_lon: float,
    ) -> float:
        """Calculate geographic proximity bonus.

        Args:
            edge: The edge provider
            user_lat: User latitude
            user_lon: User longitude

        Returns:
            Bonus score (0-1) based on proximity
        """
        # Get edge location
        edge_location = EdgeMetrics.DEFAULT_LOCATIONS.get(edge)
        if not edge_location:
            return 0.0

        # Calculate distance
        distance = calculate_distance(
            user_lat, user_lon,
            edge_location.latitude, edge_location.longitude,
        )

        # Convert distance to score (closer = higher score)
        # Within 100km = max score, >5000km = min score
        return 1.0 - normalize_value(distance, 0, 5000)

    def _estimate_latency(
        self,
        edge: EdgeProvider,
        metrics: Optional[EdgeMetrics],
        user_lat: Optional[float],
        user_lon: Optional[float],
    ) -> float:
        """Estimate latency for a user to an edge.

        Args:
            edge: The edge provider
            metrics: Edge metrics
            user_lat: User latitude
            user_lon: User longitude

        Returns:
            Estimated latency in milliseconds
        """
        base_latency = self.DEFAULT_LATENCY_MS.get(edge, 50.0)

        if metrics and metrics.avg_latency_ms > 0:
            base_latency = metrics.avg_latency_ms

        # Add geographic penalty if user location known
        if user_lat is not None and user_lon is not None:
            geo_bonus = self._calculate_geo_bonus(edge, user_lat, user_lon)
            # Add up to 50ms penalty for far distances
            latency = base_latency + (1.0 - geo_bonus) * 50
        else:
            latency = base_latency

        return round(latency, 2)

    def _fallback_decision(self, function_id: str) -> RoutingDecision:
        """Generate fallback decision when no metrics available.

        Args:
            function_id: The function ID

        Returns:
            RoutingDecision with default edge
        """
        return RoutingDecision(
            function_id=function_id,
            recommended_edge=EdgeProvider.CLOUDFLARE,
            confidence=0.5,
            reasoning="No routing data available, using default edge",
            alternatives=[EdgeProvider.FUNCTIONFLY, EdgeProvider.VERCEL, EdgeProvider.FLY, EdgeProvider.DENO],
            latency_estimate_ms=50.0,
        )

    async def get_edge_statuses(self) -> List[EdgeStatus]:
        """Get status for all available edges.

        Returns:
            List of EdgeStatus for all edges
        """
        metrics = await self._get_all_metrics()

        statuses = []
        for edge in EdgeProvider:
            edge_metrics = metrics.get(edge)

            if edge_metrics:
                statuses.append(EdgeStatus(
                    provider=edge,
                    location=edge_metrics.location,
                    available=edge_metrics.available,
                    current_load_percent=edge_metrics.current_load_percent,
                    avg_latency_ms=edge_metrics.avg_latency_ms,
                    last_check=edge_metrics.last_check,
                ))
            else:
                statuses.append(EdgeStatus(
                    provider=edge,
                    location=EdgeMetrics.DEFAULT_LOCATIONS.get(edge),
                    available=True,
                    current_load_percent=self.DEFAULT_LOAD_PERCENT,
                    avg_latency_ms=self.DEFAULT_LATENCY_MS.get(edge, 50.0),
                    last_check=datetime.utcnow(),
                ))

        return statuses


# Global routing service instance
_routing_service: Optional[RoutingService] = None


def get_routing_service() -> RoutingService:
    """Get the global routing service instance.

    Returns:
        The RoutingService instance
    """
    global _routing_service
    if _routing_service is None:
        _routing_service = RoutingService()
    return _routing_service
