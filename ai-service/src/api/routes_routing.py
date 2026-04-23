"""Routing, prewarming, and anomaly detection endpoints."""

import logging
from datetime import datetime
from typing import Optional

from fastapi import APIRouter, HTTPException, Query, status

from ..config import settings
from ..models.schemas import (
    RoutingDecisionRequest,
    RoutingDecision,
    EdgeListResponse,
    PredictionRequest,
    Prediction,
    PrewarmTriggerRequest,
    PrewarmStatus,
    AnomalyListResponse,
    AnomalyAcknowledgeRequest,
)
from ..services.routing import get_routing_service
from ..services.prewarming import (
    get_forecasting_service,
    get_prewarming_service,
)
from ..services.anomaly import (
    get_anomaly_detector,
    get_alerting_service,
)

logger = logging.getLogger(__name__)

router = APIRouter()


# =============================================================================
# Intelligent Request Routing
# =============================================================================


@router.post("/api/route/decide", response_model=RoutingDecision)
async def decide_routing(request: RoutingDecisionRequest):
    """Get optimal edge for a request.

    Uses ML-based routing to select the best edge (Cloudflare, Vercel, Fly.io, Deno)
    based on geographic proximity, historical latency, current load, and availability.

    Args:
        request: Routing decision request with function ID and user location

    Returns:
        RoutingDecision with recommended edge and alternatives
    """
    if not settings.routing_enabled:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Routing service is not enabled",
        )

    try:
        routing_service = get_routing_service()
        decision = await routing_service.decide_routing(request)
        return decision
    except Exception as e:
        logger.error(f"Routing decision failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to determine routing",
        )


@router.get("/api/route/edges", response_model=EdgeListResponse)
async def get_edge_statuses():
    """List available edges with status.

    Returns status information for all supported edge providers including
    current latency, load, and availability.

    Returns:
        EdgeListResponse with all edges and their status
    """
    if not settings.routing_enabled:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Routing service is not enabled",
        )

    try:
        routing_service = get_routing_service()
        edges = await routing_service.get_edge_statuses()

        return EdgeListResponse(
            edges=edges,
            total_count=len(edges),
            last_updated=datetime.utcnow(),
        )
    except Exception as e:
        logger.error(f"Failed to get edge statuses: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get edge statuses",
        )


# =============================================================================
# Predictive Cold Start Prewarming
# =============================================================================


@router.post("/api/prewarm/predict", response_model=Prediction)
async def get_prewarming_prediction(request: PredictionRequest):
    """Get prewarming predictions for a function.

    Uses time-series forecasting to predict request volume and
    determine if prewarming is needed.

    Args:
        request: Prediction request with function ID and window

    Returns:
        Prediction with predicted requests and confidence
    """
    if not settings.prewarming_enabled:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Prewarming service is not enabled",
        )

    try:
        forecasting_service = get_forecasting_service()
        prediction = await forecasting_service.predict(request)
        return prediction
    except Exception as e:
        logger.error(f"Prediction failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to generate prediction",
        )


@router.post("/api/prewarm/warm", response_model=PrewarmStatus)
async def trigger_prewarming(request: PrewarmTriggerRequest):
    """Trigger prewarming for a function.

    Proactively warms function instances to reduce cold starts.

    Args:
        request: Prewarm trigger request

    Returns:
        PrewarmStatus with the result
    """
    if not settings.prewarming_enabled:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Prewarming service is not enabled",
        )

    try:
        prewarming_service = get_prewarming_service()
        status_result = await prewarming_service.trigger_prewarming(request)
        return status_result
    except Exception as e:
        logger.error(f"Prewarming trigger failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to trigger prewarming",
        )


# =============================================================================
# Anomaly Detection
# =============================================================================


@router.get("/api/anomalies", response_model=AnomalyListResponse)
async def get_anomalies(
    function_id: Optional[str] = Query(None, description="Filter by function ID"),
    page: int = Query(1, ge=1, description="Page number"),
    page_size: int = Query(20, ge=1, le=100, description="Page size"),
):
    """List detected anomalies.

    Returns anomalies detected by the monitoring system including
    latency spikes, error rate increases, and cold start issues.

    Args:
        function_id: Optional function ID to filter by
        page: Page number
        page_size: Number of items per page

    Returns:
        AnomalyListResponse with detected anomalies
    """
    if not settings.anomaly_detection_enabled:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Anomaly detection is not enabled",
        )

    try:
        detector = get_anomaly_detector()

        # Calculate offset
        offset = (page - 1) * page_size

        # Get anomalies
        anomalies = await detector.get_anomalies(
            function_id=function_id,
            limit=offset + page_size,
        )

        # Paginate
        total_count = len(anomalies)
        anomalies = anomalies[offset : offset + page_size]

        return AnomalyListResponse(
            anomalies=anomalies,
            total_count=total_count,
            page=page,
            page_size=page_size,
        )
    except Exception as e:
        logger.error(f"Failed to get anomalies: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get anomalies",
        )


@router.post("/api/anomalies/acknowledge")
async def acknowledge_anomaly(request: AnomalyAcknowledgeRequest):
    """Acknowledge an anomaly.

    Marks an anomaly as acknowledged to silence alerts.

    Args:
        request: Anomaly acknowledgement request

    Returns:
        Success message
    """
    if not settings.anomaly_detection_enabled:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Anomaly detection is not enabled",
        )

    try:
        detector = get_anomaly_detector()
        success = await detector.acknowledge_anomaly(
            request.anomaly_id,
            request.acknowledged_by,
        )

        if not success:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Anomaly not found",
            )

        return {"message": "Anomaly acknowledged successfully"}
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to acknowledge anomaly: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to acknowledge anomaly",
        )
