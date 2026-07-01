"""ML Intelligence Layer API routes.

Endpoints for:
- Cost anomaly detection
- Enhanced prewarming (Holt-Winters)
- Thompson Sampling routing
- Collaborative filtering recommendations
"""

import logging
from datetime import datetime
from typing import List, Optional

from fastapi import APIRouter, HTTPException, Query, status

from ..config import settings
from ..models.schemas import (
    RoutingDecisionRequest,
    RoutingDecision,
    PredictionRequest,
    Prediction,
    EdgeListResponse,
    EdgeStatus,
    EdgeProvider,
)
from ..security.auth import require_api_key_with_scope, APIKeyInfo, KeyScope

logger = logging.getLogger(__name__)

router = APIRouter()


# =============================================================================
# Cost Anomaly Detection
# =============================================================================


@router.post("/api/ml/anomalies/cost/check")
async def check_cost_anomaly(request: dict):
    """Check a function execution for cost anomalies.

    This is called by the Go backend after each cost allocation batch.
    Uses adaptive per-function Z-score detection.
    """
    if not settings.ml_enabled:
        raise HTTPException(status_code=503, detail="ML services disabled")

    try:
        from ..services.cost_anomaly import get_cost_anomaly_detector
        from ..services.cost_anomaly.models import CostExecutionMetrics

        metrics = CostExecutionMetrics(
            function_id=request.get("function_id", ""),
            cost_cents=float(request.get("cost_cents", 0)),
            duration_ms=float(request.get("duration_ms", 0)),
            memory_mb=float(request.get("memory_mb", 0)),
            region=request.get("region", "unknown"),
        )

        detector = get_cost_anomaly_detector()
        result = await detector.check_execution(metrics)
        return result.model_dump()

    except Exception as e:
        logger.error(f"Cost anomaly check failed: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/api/ml/anomalies/cost/{tenant_id}")
async def get_cost_anomalies(
    tenant_id: str,
    hours: int = Query(24, ge=1, le=168),
    limit: int = Query(50, ge=1, le=200),
):
    """Get cost anomalies for a tenant."""
    if not settings.ml_enabled:
        raise HTTPException(status_code=503, detail="ML services disabled")

    try:
        from ..services.cost_anomaly import get_cost_anomaly_detector

        detector = get_cost_anomaly_detector()
        summary = await detector.get_summary(tenant_id, hours=hours)
        return summary.model_dump()

    except Exception as e:
        logger.error(f"Failed to get cost anomalies: {e}")
        raise HTTPException(status_code=500, detail=str(e))


# =============================================================================
# Enhanced Prewarming (Holt-Winters)
# =============================================================================


@router.post("/api/ml/prewarm/predict", response_model=Prediction)
async def ml_prewarm_predict(request: PredictionRequest):
    """Predict function demand using Holt-Winters exponential smoothing.

    Drop-in replacement for /api/prewarm/predict with seasonality awareness.
    """
    if not settings.ml_enabled:
        raise HTTPException(status_code=503, detail="ML services disabled")

    try:
        from ..services.prewarming.holt_winters import get_holt_winters_forecaster

        forecaster = get_holt_winters_forecaster()
        prediction = await forecaster.predict(request)
        return prediction

    except Exception as e:
        logger.error(f"ML prewarm prediction failed: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/api/ml/prewarm/record")
async def ml_prewarm_record(request: dict):
    """Record a request for ML prewarming tracking."""
    if not settings.ml_enabled:
        return {"status": "disabled"}

    try:
        from ..services.prewarming.holt_winters import get_holt_winters_forecaster

        forecaster = get_holt_winters_forecaster()
        await forecaster.record_request(
            function_id=request.get("function_id", ""),
            count=int(request.get("count", 1)),
        )
        return {"status": "recorded"}

    except Exception as e:
        logger.error(f"ML prewarm record failed: {e}")
        raise HTTPException(status_code=500, detail=str(e))


# =============================================================================
# Thompson Sampling Routing
# =============================================================================


@router.post("/api/ml/route/decide")
async def ml_route_decide(request: RoutingDecisionRequest):
    """Make a routing decision using Thompson Sampling.

    Drop-in replacement for /api/route/decide with adaptive learning.
    """
    if not settings.ml_enabled:
        raise HTTPException(status_code=503, detail="ML services disabled")

    try:
        from ..services.thompson_routing import get_thompson_router

        router_svc = get_thompson_router()
        decision = await router_svc.decide(request)
        return decision

    except Exception as e:
        logger.error(f"ML routing decision failed: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/api/ml/route/outcome")
async def ml_route_outcome(request: dict):
    """Record a routing outcome to update the Thompson Sampling model.

    Called after each edge execution with latency and success data.
    """
    if not settings.ml_enabled:
        return {"status": "disabled"}

    try:
        from ..services.thompson_routing import get_thompson_router
        from ..services.thompson_routing.models import RoutingOutcome

        outcome = RoutingOutcome(
            edge=request.get("edge", ""),
            function_id=request.get("function_id", ""),
            latency_ms=float(request.get("latency_ms", 0)),
            success=bool(request.get("success", True)),
            cost_cents=float(request.get("cost_cents", 0)),
        )

        router_svc = get_thompson_router()
        await router_svc.update(outcome)
        return {"status": "updated"}

    except Exception as e:
        logger.error(f"ML route outcome failed: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/api/ml/route/stats/{function_id}")
async def ml_route_stats(function_id: str):
    """Get Thompson Sampling arm stats for a function."""
    if not settings.ml_enabled:
        raise HTTPException(status_code=503, detail="ML services disabled")

    try:
        from ..services.thompson_routing import get_thompson_router

        router_svc = get_thompson_router()
        arms = await router_svc.get_arm_stats(function_id)

        return {
            "function_id": function_id,
            "arms": {
                k: {
                    "edge": v.edge,
                    "alpha": round(v.alpha, 3),
                    "beta": round(v.beta, 3),
                    "mean": round(v.mean, 3),
                    "total_pulls": v.total_pulls,
                    "success_rate": round(v.success_rate, 3),
                    "avg_latency_ms": round(v.avg_latency_ms, 2),
                }
                for k, v in arms.items()
            },
        }

    except Exception as e:
        logger.error(f"ML route stats failed: {e}")
        raise HTTPException(status_code=500, detail=str(e))


# =============================================================================
# Recommendations
# =============================================================================


@router.get("/api/ml/recommendations/{user_id}")
async def ml_recommendations(
    user_id: str,
    limit: int = Query(20, ge=1, le=100),
    exclude_installed: bool = Query(True),
):
    """Get personalized function recommendations for a user.

    Uses collaborative filtering when sufficient interaction data exists,
    falls back to popularity-based for cold-start users.
    """
    if not settings.ml_enabled:
        raise HTTPException(status_code=503, detail="ML services disabled")

    try:
        from ..services.recommendations import get_recommendation_engine

        engine = get_recommendation_engine()
        result = await engine.recommend(
            user_id=user_id,
            limit=limit,
        )
        return result.model_dump()

    except Exception as e:
        logger.error(f"ML recommendations failed: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/api/ml/recommendations/interactions")
async def ml_record_interaction(request: dict):
    """Record a user-function interaction for recommendation training."""
    if not settings.ml_enabled:
        return {"status": "disabled"}

    try:
        from ..services.recommendations import get_recommendation_engine
        from ..services.recommendations.models import InteractionEvent

        event = InteractionEvent(
            user_id=request.get("user_id", ""),
            function_id=request.get("function_id", ""),
            interaction_type=request.get("interaction_type", "view"),
            context=request.get("context"),
        )

        engine = get_recommendation_engine()
        await engine.record_interaction(event)
        return {"status": "recorded"}

    except Exception as e:
        logger.error(f"ML interaction recording failed: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/api/ml/recommendations/train")
async def ml_train_recommendations():
    """Trigger recommendation model training.

    Typically called by the daily retraining cron.
    """
    if not settings.ml_enabled:
        raise HTTPException(status_code=503, detail="ML services disabled")

    try:
        from ..services.recommendations import get_recommendation_engine

        engine = get_recommendation_engine()
        success = await engine.train()
        return {"status": "trained" if success else "insufficient_data"}

    except Exception as e:
        logger.error(f"ML training failed: {e}")
        raise HTTPException(status_code=500, detail=str(e))


# =============================================================================
# ML Health & Status
# =============================================================================


@router.get("/api/ml/health")
async def ml_health():
    """Health check for all ML services."""
    return {
        "status": "healthy" if settings.ml_enabled else "disabled",
        "services": {
            "cost_anomaly": {"enabled": settings.ml_enabled, "threshold": settings.ml_cost_anomaly_threshold},
            "prewarming": {"enabled": settings.ml_enabled, "seasonality": settings.ml_prewarm_seasonality_periods},
            "routing": {"enabled": settings.ml_enabled, "exploration": settings.ml_routing_exploration},
            "recommendations": {"enabled": settings.ml_enabled, "latent_dims": settings.ml_recommendation_latent_dims},
        },
        "model_dir": settings.ml_model_dir,
        "synthetic_data": settings.ml_synthetic_data_enabled,
    }


@router.post("/api/ml/retrain")
async def ml_retrain_all():
    """Trigger retraining of all ML models.

    Typically called by the daily cron job.
    """
    if not settings.ml_enabled:
        raise HTTPException(status_code=503, detail="ML services disabled")

    results = {}

    try:
        from ..services.recommendations import get_recommendation_engine

        engine = get_recommendation_engine()
        results["recommendations"] = "trained" if await engine.train() else "insufficient_data"
    except Exception as e:
        results["recommendations"] = f"error: {e}"

    return {"status": "completed", "results": results}
