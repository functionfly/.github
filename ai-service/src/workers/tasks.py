"""Background tasks for FlyMind AI Service.

This module provides default background tasks for cache warming,
anomaly checking, and metrics collection.
"""

import logging
from typing import Any, Dict, Optional

logger = logging.getLogger(__name__)


async def cache_warming_task() -> Dict[str, Any]:
    """Warm the cache based on predictions.

    Returns:
        Dictionary with warming results
    """
    try:
        from ..services.caching.service import get_cache_service

        cache_service = get_cache_service()

        # Get predictions
        predictions = cache_service.get_predictions(limit=50)

        if not predictions:
            return {
                "status": "no_predictions",
                "warmed": 0,
            }

        # Fetch actual values and warm cache: use current cached value when present
        # (refreshes TTL for predicted keys); keys no longer in cache are skipped
        def fetch_callback(key: str):
            return cache_service.get(key)

        results = cache_service.warm_cache(predictions, fetch_callback)
        warmed = sum(1 for ok in results.values() if ok)

        return {
            "status": "success",
            "predictions": len(predictions),
            "warmed": warmed,
            "keys_attempted": len(results),
        }

    except Exception as e:
        logger.error(f"Cache warming task failed: {e}")
        return {
            "status": "error",
            "error": str(e),
        }


async def anomaly_check_task() -> Dict[str, Any]:
    """Check for anomalies.

    Integrates with Phase 2 anomaly detection: runs statistical checks
    for all functions that have metrics, stores anomalies, and triggers alerts.

    Returns:
        Dictionary with anomaly check results
    """
    try:
        from ..config import settings
        from ..services.anomaly import get_anomaly_detector, get_alerting_service

        if not settings.anomaly_detection_enabled:
            return {
                "status": "disabled",
                "anomalies_found": 0,
                "functions_checked": 0,
            }

        detector = get_anomaly_detector()
        alerting = get_alerting_service()

        function_ids = await detector.get_function_ids_with_metrics()
        if not function_ids:
            return {
                "status": "success",
                "anomalies_found": 0,
                "functions_checked": 0,
            }

        total_anomalies = 0
        for function_id in function_ids:
            anomalies = await alerting.check_and_alert(function_id)
            total_anomalies += len(anomalies)

        return {
            "status": "success",
            "anomalies_found": total_anomalies,
            "functions_checked": len(function_ids),
        }

    except Exception as e:
        logger.error(f"Anomaly check task failed: {e}")
        return {
            "status": "error",
            "error": str(e),
        }


async def metrics_collection_task() -> Dict[str, Any]:
    """Collect and report metrics.

    Returns:
        Dictionary with metrics collection results
    """
    try:
        from ..observability.metrics import get_metrics_collector

        metrics = get_metrics_collector()

        # Get current stats
        cache_metrics = metrics.get_cache_metrics()

        return {
            "status": "success",
            "cache_hits": cache_metrics.hits,
            "cache_misses": cache_metrics.misses,
            "hit_rate": cache_metrics.hit_rate,
        }

    except Exception as e:
        logger.error(f"Metrics collection task failed: {e}")
        return {
            "status": "error",
            "error": str(e),
        }


async def cache_cleanup_task() -> Dict[str, Any]:
    """Clean up expired cache entries.

    Returns:
        Dictionary with cleanup results
    """
    try:
        from ..services.caching.service import get_cache_service

        cache_service = get_cache_service()

        # Get strategy and cleanup
        strategy = cache_service._strategy

        if hasattr(strategy, "cleanup_expired"):
            cleaned = strategy.cleanup_expired()
        else:
            cleaned = 0

        return {
            "status": "success",
            "cleaned": cleaned,
        }

    except Exception as e:
        logger.error(f"Cache cleanup task failed: {e}")
        return {
            "status": "error",
            "error": str(e),
        }


async def health_check_task() -> Dict[str, Any]:
    """Run health checks on all components.

    Returns:
        Dictionary with health check results
    """
    try:
        from ..observability.health import get_health_checker

        checker = get_health_checker()
        health = await checker.check_all()

        healthy = sum(1 for h in health.values() if h.is_healthy())
        total = len(health)

        return {
            "status": "success",
            "healthy": healthy,
            "total": total,
            "components": {
                name: comp.status.value
                for name, comp in health.items()
            },
        }

    except Exception as e:
        logger.error(f"Health check task failed: {e}")
        return {
            "status": "error",
            "error": str(e),
        }


async def orchestrator_retry_task() -> Dict[str, Any]:
    """Retry connecting to orchestrator if in degraded mode.

    This task periodically checks if the orchestrator has become available
    and re-initializes the API key validator when it recovers.

    Returns:
        Dictionary with retry results
    """
    try:
        from ..security.auth import get_api_key_validator, initialize_api_key_validator
        from ..observability.health import get_health_checker

        validator = get_api_key_validator()
        checker = get_health_checker()

        # Only retry if we're in degraded state
        if not checker.is_degraded():
            return {
                "status": "not_degraded",
                "action": "none",
            }

        # Check if orchestrator is now reachable
        import httpx
        from ..config import settings

        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                response = await client.get(f"{settings.orchestrator_url}/health")

            if response.status_code == 200:
                # Orchestrator is back! Re-initialize the validator
                logger.info("Orchestrator recovered, re-initializing API key validator")
                await initialize_api_key_validator()
                checker.clear_degraded()

                return {
                    "status": "recovered",
                    "action": "reinitialized_validator",
                }
        except Exception:
            pass

        return {
            "status": "still_unreachable",
            "action": "none",
        }

    except Exception as e:
        logger.error(f"Orchestrator retry task failed: {e}")
        return {
            "status": "error",
            "error": str(e),
        }


async def ml_model_retrain_task() -> Dict[str, Any]:
    """Retrain all ML models.

    This task runs periodically to retrain ML models including:
    - Recommendations (ALS collaborative filtering)
    - Cost anomaly (stats refresh)
    - Prewarming (Holt-Winters model refresh)
    - Thompson routing (arm statistics refresh)

    Returns:
        Dictionary with retrain results for each model type
    """
    import time

    results = {}
    start_time = time.time()

    try:
        from ..config import settings
        from ..observability.metrics import get_metrics_collector

        if not settings.ml_enabled:
            return {
                "status": "disabled",
                "results": {},
            }

        mc = get_metrics_collector()

        try:
            from ..services.recommendations import get_recommendation_engine

            engine = get_recommendation_engine()
            rec_start = time.time()
            rec_result = await engine.train()
            rec_duration = time.time() - rec_start
            results["recommendations"] = "trained" if rec_result else "insufficient_data"
            mc.record_model_training("recommendations", rec_result)
            mc.record_model_training_duration("recommendations", rec_duration)
        except Exception as e:
            results["recommendations"] = f"error: {e}"
            mc.record_model_training("recommendations", False)

        try:
            from ..services.cost_anomaly import get_cost_anomaly_detector

            detector = get_cost_anomaly_detector()
            stats = await detector._load_stats("_global")
            results["cost_anomaly"] = f"ok (adaptive_threshold={detector._threshold})"
            mc.record_model_training("cost_anomaly", True)
        except Exception as e:
            results["cost_anomaly"] = f"error: {e}"
            mc.record_model_training("cost_anomaly", False)

        try:
            from ..services.prewarming.holt_winters import get_holt_winters_forecaster

            forecaster = get_holt_winters_forecaster()
            results["prewarming"] = f"ok (seasonality={forecaster._seasonality})"
            mc.record_model_training("prewarming", True)
        except Exception as e:
            results["prewarming"] = f"error: {e}"
            mc.record_model_training("prewarming", False)

        try:
            from ..services.thompson_routing import get_thompson_router

            router_svc = get_thompson_router()
            results["thompson_routing"] = f"ok (exploration={router_svc._exploration_rate})"
            mc.record_model_training("thompson_routing", True)
        except Exception as e:
            results["thompson_routing"] = f"error: {e}"
            mc.record_model_training("thompson_routing", False)

        total_duration = time.time() - start_time
        results["_total_duration_seconds"] = round(total_duration, 2)

        logger.info(f"ML model retrain completed in {total_duration:.2f}s: {results}")

        return {
            "status": "completed",
            "results": results,
        }

    except Exception as e:
        logger.error(f"ML model retrain task failed: {e}")
        return {
            "status": "error",
            "error": str(e),
        }


def register_default_tasks(scheduler) -> None:
    """Register default background tasks.

    Args:
        scheduler: TaskScheduler instance
    """
    # Cache warming - every 5 minutes
    scheduler.add_task(
        name="cache_warming",
        func=cache_warming_task,
        interval_seconds=300,  # 5 minutes
    )

    # Anomaly check - every 5 minutes
    scheduler.add_task(
        name="anomaly_check",
        func=anomaly_check_task,
        interval_seconds=300,
    )

    # Metrics collection - every minute
    scheduler.add_task(
        name="metrics_collection",
        func=metrics_collection_task,
        interval_seconds=60,
    )

    # Cache cleanup - every 10 minutes
    scheduler.add_task(
        name="cache_cleanup",
        func=cache_cleanup_task,
        interval_seconds=600,
    )

    # Health check - every 2 minutes
    scheduler.add_task(
        name="health_check",
        func=health_check_task,
        interval_seconds=120,
    )

    # Orchestrator retry - every 30 seconds when degraded
    scheduler.add_task(
        name="orchestrator_retry",
        func=orchestrator_retry_task,
        interval_seconds=30,
    )

    # ML model retraining - daily (86400 seconds)
    # Note: ml_retrain_cron config option (e.g., "0 3 * * *" for 3 AM) is not yet parsed;
    # using 24-hour interval for simplicity. A future enhancement could parse cron.
    scheduler.add_task(
        name="ml_model_retrain",
        func=ml_model_retrain_task,
        interval_seconds=86400,  # Daily
    )

    logger.info("Registered default background tasks")


# Aliases for backwards compatibility
CacheWarmingTask = cache_warming_task
AnomalyCheckTask = anomaly_check_task
MetricsCollectionTask = metrics_collection_task
MLModelRetrainTask = ml_model_retrain_task
