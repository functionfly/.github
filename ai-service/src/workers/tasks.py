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

    logger.info("Registered default background tasks")


# Aliases for backwards compatibility
CacheWarmingTask = cache_warming_task
AnomalyCheckTask = anomaly_check_task
MetricsCollectionTask = metrics_collection_task
