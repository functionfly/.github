"""ML Intelligence Layer API routes.

Endpoints for:
- Cost anomaly detection
- Enhanced prewarming (Holt-Winters)
- Thompson Sampling routing
- Collaborative filtering recommendations
"""

import asyncio
import logging
import re

from fastapi import APIRouter, Depends, HTTPException, Query

from ..config import settings
from ..models.schemas import (
    CostAnomalyCheckRequest,
    CostAnomalyCheckResponse,
    InteractionType,
    MLHealthResponse,
    Prediction,
    PredictionRequest,
    PrewarmRecordRequest,
    RecommendationInteractionRequest,
    RecommendationResponse,
    RoutingDecision,
    RoutingDecisionRequest,
    RoutingOutcomeRequest,
)
from ..security.auth import (
    APIKeyInfo,
    KeyScope,
    require_api_key_with_scope,
)

logger = logging.getLogger(__name__)

router = APIRouter()

ML_SERVICE_UNAVAILABLE = "ML services are currently unavailable"
AUTHENTICATION_REQUIRED = "Authentication required. Provide X-API-Key header."
INSUFFICIENT_PERMISSIONS = "Insufficient permissions for ML operations"

ID_PATTERN = re.compile(r"^[a-zA-Z0-9_\-]{1,64}$")


class TrainingRateLimiter:
    """Rate limiter for ML training endpoints.

    Prevents resource abuse by limiting training requests per tenant.
    Uses in-memory tracking with hourly windows.
    """

    def __init__(self):
        self._lock = asyncio.Lock()
        self._counts: dict[str, dict] = {}

    async def check_training_rate(self, tenant_id: str) -> None:
        """Check if tenant is within training rate limits.

        Args:
            tenant_id: The tenant ID to check

        Raises:
            HTTPException: 429 if rate limit exceeded
        """
        limit_per_hour = settings.ml_training_rate_limit_per_hour

        async with self._lock:
            now = time.time()
            if tenant_id not in self._counts:
                self._counts[tenant_id] = {"count": 0, "window_start": now}

            entry = self._counts[tenant_id]
            window_start = entry["window_start"]

            if now - window_start >= 3600:
                entry["count"] = 0
                entry["window_start"] = now

            if entry["count"] >= limit_per_hour:
                retry_after = int(3600 - (now - window_start)) + 1
                logger.warning(f"Training rate limit exceeded for tenant {tenant_id}")
                raise HTTPException(
                    status_code=429,
                    detail=f"Training rate limit exceeded. Maximum {limit_per_hour} training requests per hour. Retry after {retry_after} seconds.",
                    headers={"Retry-After": str(retry_after)},
                )

            entry["count"] += 1


_training_limiter: TrainingRateLimiter | None = None


def _get_training_limiter() -> TrainingRateLimiter:
    """Get the global training rate limiter instance."""
    global _training_limiter
    if _training_limiter is None:
        _training_limiter = TrainingRateLimiter()
    return _training_limiter


def _validate_id(value: str, field_name: str) -> None:
    """Validate an ID parameter to prevent injection attacks.

    Args:
        value: The ID value to validate
        field_name: The name of the field for error messages

    Raises:
        HTTPException: 400 if the ID is invalid
    """
    if not value:
        raise HTTPException(
            status_code=400,
            detail=f"{field_name} is required"
        )
    if not ID_PATTERN.match(value):
        raise HTTPException(
            status_code=400,
            detail=f"{field_name} contains invalid characters. Only alphanumeric, underscore, and hyphen are allowed (max 64 chars)"
        )


def _sanitize_error(e: Exception) -> str:
    """Sanitize error message to prevent internal detail leakage.

    Returns a generic error message regardless of the actual exception.
    Internal details are logged but never exposed to clients.
    """
    error_type = type(e).__name__
    error_str = str(e)

    logger.debug(f"ML service error: type={error_type}, message={error_str}")

    if isinstance(e, HTTPException):
        return e.detail

    if isinstance(e, (TimeoutError, asyncio.TimeoutError)):
        return "The operation timed out. Please try again."

    if "redis" in error_type.lower() or "redis" in error_str.lower():
        return "A cache service is temporarily unavailable"

    if "postgres" in error_type.lower() or "database" in error_str.lower():
        return "A database service is temporarily unavailable"

    return "An internal error occurred"


def _validate_tenant_access(api_key_tenant_id: str, requested_tenant_id: str, api_key: APIKeyInfo) -> None:
    """Validate that the API key's tenant can access the requested tenant's data.

    Args:
        api_key_tenant_id: The tenant_id from the API key
        requested_tenant_id: The tenant_id in the request
        api_key: Full API key info for error messages

    Raises:
        HTTPException: 403 if tenant access is denied
    """
    if api_key_tenant_id != requested_tenant_id:
        logger.warning(
            f"Tenant access denied: API key tenant '{api_key_tenant_id}' "
            f"attempted to access tenant '{requested_tenant_id}' resources"
        )
        raise HTTPException(
            status_code=403,
            detail="Access denied: You do not have permission to access this tenant's data"
        )


def _get_redis_status() -> bool:
    """Check if Redis is connected."""
    try:
        import redis
        r = redis.from_url(settings.redis_url)
        r.ping()
        return True
    except Exception:
        return False


def _get_ml_metrics_collector():
    """Get ML metrics collector,懒加载避免循环导入."""
    from ..observability.metrics import get_metrics_collector
    return get_metrics_collector()


# =============================================================================
# Cost Anomaly Detection
# =============================================================================


@router.post("/api/ml/anomalies/cost/check", response_model=CostAnomalyCheckResponse)
async def check_cost_anomaly(
    request: CostAnomalyCheckRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.ML_WRITE)),
):
    """Check a function execution for cost anomalies.

    This is called by the Go backend after each cost allocation batch.
    Uses adaptive per-function Z-score detection.
    Tenant isolation is enforced - data is stored per-tenant.
    """
    import time

    from ..security.audit import create_ml_audit_event

    if not settings.ml_enabled:
        raise HTTPException(status_code=503, detail=ML_SERVICE_UNAVAILABLE)

    _validate_id(request.function_id, "function_id")

    start_time = time.time()
    success = False
    error_msg = None

    try:
        from ..services.cost_anomaly import get_cost_anomaly_detector
        from ..services.cost_anomaly.models import CostExecutionMetrics

        metrics_data = CostExecutionMetrics(
            function_id=request.function_id,
            cost_cents=request.cost_cents,
            duration_ms=request.duration_ms,
            memory_mb=request.memory_mb,
            region=request.region,
        )

        detector = get_cost_anomaly_detector()
        result = await detector.check_execution(api_key.tenant_id, metrics_data)

        latency_ms = (time.time() - start_time) * 1000
        mc = _get_ml_metrics_collector()
        mc.record_cost_anomaly_check(request.function_id)
        mc.record_cost_anomaly_latency(request.function_id, latency_ms)
        if result.is_anomaly:
            mc.record_cost_anomaly_detection(
                request.function_id,
                result.severity or "unknown",
                result.anomaly_type or "unknown",
            )

        success = True
        return result

    except Exception as e:
        error_msg = _sanitize_error(e)
        logger.error(f"Cost anomaly check failed: {e}")
        raise HTTPException(status_code=500, detail=error_msg) from None

    finally:
        final_latency_ms = (time.time() - start_time) * 1000
        create_ml_audit_event(
            api_key_info=api_key,
            operation="ml_anomaly_check",
            model_type="cost_anomaly",
            success=success,
            latency_ms=final_latency_ms,
            function_id=request.function_id,
            error_message=error_msg,
        )


@router.get("/api/ml/anomalies/cost/{tenant_id}")
async def get_cost_anomalies(
    tenant_id: str,
    hours: int = Query(24, ge=1, le=168),
    limit: int = Query(50, ge=1, le=200),
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.ML_READ)),
):
    """Get cost anomalies for a tenant.

    Returns anomalies only for the tenant associated with the API key.
    Tenant isolation is enforced - cross-tenant access is prohibited.
    """
    import time

    from ..security.audit import create_ml_audit_event

    if not settings.ml_enabled:
        raise HTTPException(status_code=503, detail=ML_SERVICE_UNAVAILABLE)

    _validate_id(tenant_id, "tenant_id")

    _validate_tenant_access(api_key.tenant_id, tenant_id, api_key)

    start_time = time.time()
    success = False
    error_msg = None

    try:
        from ..services.cost_anomaly import get_cost_anomaly_detector

        detector = get_cost_anomaly_detector()
        summary = await detector.get_summary(tenant_id, hours=hours)
        success = True
        return summary.model_dump()

    except Exception as e:
        error_msg = _sanitize_error(e)
        logger.error(f"Failed to get cost anomalies: {e}")
        raise HTTPException(status_code=500, detail=error_msg) from None

    finally:
        latency_ms = (time.time() - start_time) * 1000
        create_ml_audit_event(
            api_key_info=api_key,
            operation="ml_anomaly_check",
            model_type="cost_anomaly",
            success=success,
            latency_ms=latency_ms,
            function_id=tenant_id,
            error_message=error_msg,
        )


# =============================================================================
# Enhanced Prewarming (Holt-Winters)
# =============================================================================


@router.post("/api/ml/prewarm/predict", response_model=Prediction)
async def ml_prewarm_predict(
    request: PredictionRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.ML_READ)),
):
    """Predict function demand using Holt-Winters exponential smoothing.

    Drop-in replacement for /api/prewarm/predict with seasonality awareness.
    Tenant isolation is enforced - predictions are per-tenant.
    """
    import time

    from ..security.audit import create_ml_audit_event

    if not settings.ml_enabled:
        raise HTTPException(status_code=503, detail=ML_SERVICE_UNAVAILABLE)

    _validate_id(request.function_id, "function_id")

    start_time = time.time()
    success = False
    error_msg = None

    try:
        from ..services.prewarming.holt_winters import get_holt_winters_forecaster

        forecaster = get_holt_winters_forecaster()
        prediction = await forecaster.predict(api_key.tenant_id, request)

        mc = _get_ml_metrics_collector()
        mc.record_prewarm_prediction(request.function_id, prediction.confidence)

        success = True
        return prediction

    except Exception as e:
        error_msg = _sanitize_error(e)
        logger.error(f"ML prewarm prediction failed: {e}")
        raise HTTPException(status_code=500, detail=error_msg) from None

    finally:
        latency_ms = (time.time() - start_time) * 1000
        create_ml_audit_event(
            api_key_info=api_key,
            operation="ml_prewarm_predict",
            model_type="prewarming",
            success=success,
            latency_ms=latency_ms,
            function_id=request.function_id,
            error_message=error_msg,
        )


@router.post("/api/ml/prewarm/record")
async def ml_prewarm_record(
    request: PrewarmRecordRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.ML_WRITE)),
):
    """Record a request for ML prewarming tracking.

    Uses tenant isolation - records are stored per-tenant.
    """
    import time

    from ..security.audit import create_ml_audit_event

    if not settings.ml_enabled:
        return {"status": "disabled"}

    _validate_id(request.function_id, "function_id")

    start_time = time.time()
    success = False
    error_msg = None

    try:
        from ..services.prewarming.holt_winters import get_holt_winters_forecaster

        forecaster = get_holt_winters_forecaster()
        await forecaster.record_request(
            api_key.tenant_id,
            function_id=request.function_id,
            count=request.count,
        )
        success = True
        return {"status": "recorded"}

    except Exception as e:
        error_msg = _sanitize_error(e)
        logger.error(f"ML prewarm record failed: {e}")
        raise HTTPException(status_code=500, detail=error_msg) from None

    finally:
        latency_ms = (time.time() - start_time) * 1000
        create_ml_audit_event(
            api_key_info=api_key,
            operation="ml_prewarm_record",
            model_type="prewarming",
            success=success,
            latency_ms=latency_ms,
            function_id=request.function_id,
            error_message=error_msg,
        )


# =============================================================================
# Thompson Sampling Routing
# =============================================================================


@router.post("/api/ml/route/decide", response_model=RoutingDecision)
async def ml_route_decide(
    request: RoutingDecisionRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.ML_READ)),
):
    """Make a routing decision using Thompson Sampling.

    Drop-in replacement for /api/route/decide with adaptive learning.
    Uses tenant isolation - routing decisions are stored per-tenant.
    """
    import time

    from ..security.audit import create_ml_audit_event

    if not settings.ml_enabled:
        raise HTTPException(status_code=503, detail=ML_SERVICE_UNAVAILABLE)

    _validate_id(request.function_id, "function_id")

    start_time = time.time()
    success = False
    error_msg = None

    try:
        from ..services.thompson_routing import get_thompson_router

        router_svc = get_thompson_router()
        decision = await router_svc.decide(api_key.tenant_id, request)

        mc = _get_ml_metrics_collector()
        is_exploration = "Exploration" in decision.reasoning
        mc.record_thompson_decision(
            request.function_id,
            decision.recommended_edge.value,
            is_exploration,
        )

        success = True
        return decision

    except Exception as e:
        error_msg = _sanitize_error(e)
        logger.error(f"ML routing decision failed: {e}")
        raise HTTPException(status_code=500, detail=error_msg) from None

    finally:
        latency_ms = (time.time() - start_time) * 1000
        create_ml_audit_event(
            api_key_info=api_key,
            operation="ml_route_decide",
            model_type="routing",
            success=success,
            latency_ms=latency_ms,
            function_id=request.function_id,
            error_message=error_msg,
        )


@router.post("/api/ml/route/outcome")
async def ml_route_outcome(
    request: RoutingOutcomeRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.ML_WRITE)),
):
    """Record a routing outcome to update the Thompson Sampling model.

    Called after each edge execution with latency and success data.
    Uses tenant isolation - outcomes are stored per-tenant.
    """
    import time

    from ..security.audit import create_ml_audit_event

    if not settings.ml_enabled:
        return {"status": "disabled"}

    _validate_id(request.function_id, "function_id")

    start_time = time.time()
    success = False
    error_msg = None

    try:
        from ..services.thompson_routing import get_thompson_router
        from ..services.thompson_routing.models import RoutingOutcome

        outcome = RoutingOutcome(
            edge=request.edge,
            function_id=request.function_id,
            latency_ms=request.latency_ms,
            success=request.success,
            cost_cents=request.cost_cents,
        )

        router_svc = get_thompson_router()
        await router_svc.update(api_key.tenant_id, outcome)

        mc = _get_ml_metrics_collector()
        mc.record_thompson_arm_pull(request.function_id, request.edge)
        latency_reward = max(0, 1.0 - min(request.latency_ms / 500.0, 1.0))
        success_reward = 1.0 if request.success else 0.0
        cost_reward = max(0, 1.0 - min(request.cost_cents / 1.0, 1.0))
        reward = 0.4 * latency_reward + 0.4 * success_reward + 0.2 * cost_reward
        mc.record_thompson_reward(request.function_id, request.edge, reward)

        success = True
        return {"status": "updated"}

    except Exception as e:
        error_msg = _sanitize_error(e)
        logger.error(f"ML route outcome failed: {e}")
        raise HTTPException(status_code=500, detail=error_msg) from None

    finally:
        latency_ms = (time.time() - start_time) * 1000
        create_ml_audit_event(
            api_key_info=api_key,
            operation="ml_route_outcome",
            model_type="routing",
            success=success,
            latency_ms=latency_ms,
            function_id=request.function_id,
            error_message=error_msg,
        )


@router.get("/api/ml/route/stats/{tenant_id}/{function_id}")
async def ml_route_stats(
    tenant_id: str,
    function_id: str,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.ML_READ)),
):
    """Get Thompson Sampling arm stats for a function.

    Tenant isolation is enforced - returns data only for the API key's tenant.
    """
    import time

    from ..security.audit import create_ml_audit_event

    if not settings.ml_enabled:
        raise HTTPException(status_code=503, detail=ML_SERVICE_UNAVAILABLE)

    _validate_id(tenant_id, "tenant_id")
    _validate_id(function_id, "function_id")

    _validate_tenant_access(api_key.tenant_id, tenant_id, api_key)

    start_time = time.time()
    success = False
    error_msg = None

    try:
        from ..services.thompson_routing import get_thompson_router

        router_svc = get_thompson_router()
        arms = await router_svc.get_arm_stats(tenant_id, function_id)
        success = True

        return {
            "tenant_id": tenant_id,
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
        error_msg = _sanitize_error(e)
        logger.error(f"ML route stats failed: {e}")
        raise HTTPException(status_code=500, detail=error_msg) from None

    finally:
        latency_ms = (time.time() - start_time) * 1000
        create_ml_audit_event(
            api_key_info=api_key,
            operation="ml_route_stats",
            model_type="routing",
            success=success,
            latency_ms=latency_ms,
            function_id=function_id,
            error_message=error_msg,
        )


# =============================================================================
# Recommendations
# =============================================================================


@router.get("/api/ml/recommendations/{tenant_id}/{user_id}", response_model=RecommendationResponse)
async def ml_recommendations(
    tenant_id: str,
    user_id: str,
    limit: int = Query(20, ge=1, le=100),
    exclude_installed: bool = Query(True),
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.ML_READ)),
):
    """Get personalized function recommendations for a user.

    Uses collaborative filtering when sufficient interaction data exists,
    falls back to popularity-based for cold-start users.
    Tenant isolation is enforced - recommendations are per-tenant.
    """
    import time

    from ..security.audit import create_ml_audit_event

    if not settings.ml_enabled:
        raise HTTPException(status_code=503, detail=ML_SERVICE_UNAVAILABLE)

    _validate_id(tenant_id, "tenant_id")
    _validate_id(user_id, "user_id")

    _validate_tenant_access(api_key.tenant_id, tenant_id, api_key)

    start_time = time.time()
    success = False
    error_msg = None

    try:
        from ..services.recommendations import get_recommendation_engine

        engine = get_recommendation_engine()
        result = await engine.recommend(
            tenant_id=tenant_id,
            user_id=user_id,
            limit=limit,
        )

        mc = _get_ml_metrics_collector()
        mc.record_recommendation_served(user_id, result.strategy)

        success = True
        return result

    except Exception as e:
        error_msg = _sanitize_error(e)
        logger.error(f"ML recommendations failed: {e}")
        raise HTTPException(status_code=500, detail=error_msg) from None

    finally:
        latency_ms = (time.time() - start_time) * 1000
        create_ml_audit_event(
            api_key_info=api_key,
            operation="ml_recommendations",
            model_type="recommendations",
            success=success,
            latency_ms=latency_ms,
            user_id=user_id,
            error_message=error_msg,
        )


@router.post("/api/ml/recommendations/interactions")
async def ml_record_interaction(
    request: RecommendationInteractionRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.ML_WRITE)),
):
    """Record a user-function interaction for recommendation training.

    Uses tenant isolation - interactions are stored per-tenant.
    """
    import time

    from ..security.audit import create_ml_audit_event

    if not settings.ml_enabled:
        return {"status": "disabled"}

    _validate_id(request.user_id, "user_id")
    _validate_id(request.function_id, "function_id")

    start_time = time.time()
    success = False
    error_msg = None

    try:
        from ..services.recommendations import get_recommendation_engine
        from ..services.recommendations.models import InteractionEvent

        event = InteractionEvent(
            user_id=request.user_id,
            function_id=request.function_id,
            interaction_type=request.interaction_type,
            context=request.context,
        )

        engine = get_recommendation_engine()
        await engine.record_interaction(api_key.tenant_id, event)

        mc = _get_ml_metrics_collector()
        mc.record_recommendation_interaction(request.interaction_type)

        success = True
        return {"status": "recorded"}

    except Exception as e:
        error_msg = _sanitize_error(e)
        logger.error(f"ML interaction recording failed: {e}")
        raise HTTPException(status_code=500, detail=error_msg) from None

    finally:
        latency_ms = (time.time() - start_time) * 1000
        create_ml_audit_event(
            api_key_info=api_key,
            operation="ml_interaction_record",
            model_type="recommendations",
            success=success,
            latency_ms=latency_ms,
            user_id=request.user_id,
            function_id=request.function_id,
            error_message=error_msg,
        )


@router.post("/api/ml/recommendations/train")
async def ml_train_recommendations(
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.ML_ADMIN)),
):
    """Trigger recommendation model training for the API key's tenant.

    Typically called by the daily retraining cron.
    Has a timeout to prevent indefinite training.
    Trains only for the tenant associated with the API key.
    """
    import asyncio
    import time

    from ..security.audit import create_ml_audit_event

    if not settings.ml_enabled:
        raise HTTPException(status_code=503, detail=ML_SERVICE_UNAVAILABLE)

    await _get_training_limiter().check_training_rate(api_key.tenant_id)

    start_time = time.time()
    success = False
    error_msg = None
    timeout_seconds = settings.ml_training_timeout_seconds

    try:
        from ..services.recommendations import get_recommendation_engine

        engine = get_recommendation_engine()

        try:
            result = await asyncio.wait_for(
                engine.train(api_key.tenant_id),
                timeout=timeout_seconds
            )
        except TimeoutError:
            logger.error(f"ML training timed out after {timeout_seconds}s")
            raise HTTPException(
                status_code=504,
                detail=f"Training timed out after {timeout_seconds} seconds"
            ) from None

        mc = _get_ml_metrics_collector()
        mc.record_model_training("recommendations", result)

        success = result
        return {"status": "trained" if result else "insufficient_data", "tenant_id": api_key.tenant_id}

    except HTTPException:
        raise
    except Exception as e:
        error_msg = _sanitize_error(e)
        logger.error(f"ML training failed: {e}")
        mc = _get_ml_metrics_collector()
        mc.record_model_training("recommendations", False)
        raise HTTPException(status_code=500, detail=error_msg) from None

    finally:
        latency_ms = (time.time() - start_time) * 1000
        mc = _get_ml_metrics_collector()
        mc.record_model_training_duration("recommendations", latency_ms / 1000)
        create_ml_audit_event(
            api_key_info=api_key,
            operation="ml_train",
            model_type="recommendations",
            success=success,
            latency_ms=latency_ms,
            error_message=error_msg,
        )


# =============================================================================
# ML Health & Status
# =============================================================================


@router.get("/api/ml/health", response_model=MLHealthResponse)
async def ml_health(
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.ML_READ)),
):
    """Health check for all ML services."""
    redis_connected = _get_redis_status()
    return MLHealthResponse(
        status="healthy" if settings.ml_enabled and redis_connected else "degraded",
        services={
            "cost_anomaly": {"enabled": settings.ml_enabled, "threshold": settings.ml_cost_anomaly_threshold},
            "prewarming": {"enabled": settings.ml_enabled, "seasonality": settings.ml_prewarm_seasonality_periods},
            "routing": {"enabled": settings.ml_enabled, "exploration": settings.ml_routing_exploration},
            "recommendations": {"enabled": settings.ml_enabled, "latent_dims": settings.ml_recommendation_latent_dims},
        },
        model_dir=settings.ml_model_dir,
        synthetic_data=settings.ml_synthetic_data_enabled,
        redis_connected=redis_connected,
    )


@router.post("/api/ml/retrain")
async def ml_retrain_all(
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.ML_ADMIN)),
):
    """Trigger retraining of all ML models for the API key's tenant.

    Typically called by the daily cron job.
    Trains: recommendations, cost_anomaly, prewarming, thompson_routing.
    Each training operation has a timeout to prevent indefinite blocking.
    Only trains for the tenant associated with the API key.
    """
    import asyncio
    import time

    if not settings.ml_enabled:
        raise HTTPException(status_code=503, detail=ML_SERVICE_UNAVAILABLE)

    await _get_training_limiter().check_training_rate(api_key.tenant_id)

    timeout_seconds = settings.ml_training_timeout_seconds
    results = {}
    start_time = time.time()

    try:
        from ..services.recommendations import get_recommendation_engine

        engine = get_recommendation_engine()
        rec_start = time.time()
        try:
            rec_result = await asyncio.wait_for(
                engine.train(api_key.tenant_id),
                timeout=timeout_seconds
            )
        except TimeoutError:
            logger.error(f"Recommendations training timed out after {timeout_seconds}s")
            rec_result = False
        rec_duration = time.time() - rec_start
        results["recommendations"] = "trained" if rec_result else "timeout_or_insufficient"

        mc = _get_ml_metrics_collector()
        mc.record_model_training("recommendations", rec_result)
        mc.record_model_training_duration("recommendations", rec_duration)
    except Exception as e:
        results["recommendations"] = f"error: {e}"
        mc = _get_ml_metrics_collector()
        mc.record_model_training("recommendations", False)

    try:
        from ..services.cost_anomaly import get_cost_anomaly_detector

        detector = get_cost_anomaly_detector()
        await detector._load_stats("_global")
        results["cost_anomaly"] = f"ok (adaptive_threshold={detector._threshold})"

        mc = _get_ml_metrics_collector()
        mc.record_model_training("cost_anomaly", True)
    except Exception as e:
        results["cost_anomaly"] = f"error: {e}"
        mc = _get_ml_metrics_collector()
        mc.record_model_training("cost_anomaly", False)

    try:
        from ..services.prewarming.holt_winters import get_holt_winters_forecaster

        forecaster = get_holt_winters_forecaster()
        results["prewarming"] = f"ok (seasonality={forecaster._seasonality})"

        mc = _get_ml_metrics_collector()
        mc.record_model_training("prewarming", True)
    except Exception as e:
        results["prewarming"] = f"error: {e}"
        mc = _get_ml_metrics_collector()
        mc.record_model_training("prewarming", False)

    try:
        from ..services.thompson_routing import get_thompson_router

        router_svc = get_thompson_router()
        results["thompson_routing"] = f"ok (exploration={router_svc._exploration_rate})"

        mc = _get_ml_metrics_collector()
        mc.record_model_training("thompson_routing", True)
    except Exception as e:
        results["thompson_routing"] = f"error: {e}"
        mc = _get_ml_metrics_collector()
        mc.record_model_training("thompson_routing", False)

    total_duration = time.time() - start_time
    results["_total_duration_seconds"] = round(total_duration, 2)

    return {"status": "completed", "results": results}
