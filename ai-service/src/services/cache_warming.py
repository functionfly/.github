"""Cache warming service for ML intelligence layer.

This module provides cache warming functionality to pre-populate Redis cache
with ML state when the service starts, improving initial query performance.
"""

import asyncio
import logging
from datetime import datetime, timedelta
from typing import Optional

logger = logging.getLogger(__name__)


class CacheWarmingService:
    """Service to warm ML caches on startup.

    This helps ML services perform better immediately after startup by
    pre-loading commonly used data from persistent storage into Redis.
    """

    def __init__(
        self,
        redis_client,
        cost_anomaly_repo=None,
        routing_repo=None,
        recommendations_repo=None,
        prewarming_repo=None,
        database_url: Optional[str] = None,
        orchestrator_url: Optional[str] = None,
        orchestrator_api_key: Optional[str] = None,
    ):
        """Initialize the cache warming service.

        Args:
            redis_client: Redis client instance
            cost_anomaly_repo: Cost anomaly repository (optional)
            routing_repo: Routing repository (optional)
            recommendations_repo: Recommendations repository (optional)
            prewarming_repo: Prewarming repository (optional)
            database_url: PostgreSQL connection string for direct DB access
            orchestrator_url: Orchestrator API URL for tenant lookup
            orchestrator_api_key: API key for orchestrator authentication
        """
        self.redis = redis_client
        self.cost_anomaly = cost_anomaly_repo
        self.routing = routing_repo
        self.recommendations = recommendations_repo
        self.prewarming = prewarming_repo
        self._database_url = database_url
        self._orchestrator_url = orchestrator_url
        self._orchestrator_api_key = orchestrator_api_key

        self._warming_in_progress = False
        self._last_warming_time: Optional[datetime] = None
        self._warming_errors: list[dict] = []

    @property
    def is_warming(self) -> bool:
        """Check if cache warming is currently in progress."""
        return self._warming_in_progress

    async def get_active_tenants(self, limit: int = 100) -> list[str]:
        """Fetch active tenant IDs for cache warming.

        Queries the orchestrator API or database for tenants that have
        active API usage within the configured lookback window.

        Args:
            limit: Maximum number of tenants to return (prioritized by recent activity)

        Returns:
            List of active tenant ID strings
        """
        if self._orchestrator_url:
            return await self._get_tenants_via_orchestrator(limit)
        elif self._database_url:
            return await self._get_tenants_via_database(limit)
        else:
            logger.warning("No orchestrator URL or database URL configured for tenant lookup")
            return []

    async def _get_tenants_via_orchestrator(self, limit: int) -> list[str]:
        """Query orchestrator API for active tenants."""
        try:
            import httpx

            async with httpx.AsyncClient(timeout=10.0) as client:
                headers = {}
                if self._orchestrator_api_key:
                    headers["Authorization"] = f"Bearer {self._orchestrator_api_key}"

                response = await client.get(
                    f"{self._orchestrator_url}/v1/admin/tenants",
                    headers=headers,
                    params={"limit": limit, "active_only": "true"},
                )

                if response.status_code == 200:
                    data = response.json()
                    tenants = data.get("data", data.get("tenants", []))
                    tenant_ids = [t.get("id") or t.get("tenant_id") for t in tenants if t.get("id") or t.get("tenant_id")]
                    logger.info(f"Fetched {len(tenant_ids)} tenants from orchestrator")
                    return tenant_ids
                else:
                    logger.warning(f"Orchestrator returned {response.status_code}, falling back to database")
                    return await self._get_tenants_via_database(limit)

        except ImportError:
            logger.warning("httpx not available, falling back to database query")
            return await self._get_tenants_via_database(limit)
        except Exception as e:
            logger.error(f"Failed to fetch tenants from orchestrator: {e}")
            return await self._get_tenants_via_database(limit)

    async def _get_tenants_via_database(self, limit: int) -> list[str]:
        """Query PostgreSQL directly for active tenants."""
        try:
            import asyncpg

            conn = await asyncpg.connect(self._database_url)
            try:
                rows = await conn.fetch("""
                    SELECT DISTINCT t.id::text
                    FROM tenants t
                    INNER JOIN api_keys ak ON ak.tenant_id = t.id
                    WHERE t.deleted_at IS NULL
                      AND ak.revoked = false
                      AND (
                          ak.last_used_at IS NOT NULL
                          AND ak.last_used_at > NOW() - INTERVAL '30 days'
                      )
                    ORDER BY t.created_at DESC
                    LIMIT $1
                """, limit)

                tenant_ids = [row["id"] for row in rows]
                logger.info(f"Fetched {len(tenant_ids)} active tenants from database")
                return tenant_ids
            finally:
                await conn.close()

        except ImportError:
            logger.warning("asyncpg not installed, cannot query database directly for tenants")
            return []
        except Exception as e:
            logger.error(f"Failed to fetch tenants from database: {e}")
            return []

    async def warm_all(self, tenants: list[str]) -> dict:
        """Warm cache for all ML services for given tenants.

        Args:
            tenants: List of tenant IDs to warm caches for

        Returns:
            Dictionary with warming results for each service
        """
        if self._warming_in_progress:
            logger.warning("Cache warming already in progress, skipping")
            return {"status": "skipped", "reason": "warming_in_progress"}

        self._warming_in_progress = True
        self._warming_errors = []

        results = {
            "cost_anomaly": {"keys_loaded": 0, "errors": []},
            "routing": {"keys_loaded": 0, "errors": []},
            "recommendations": {"keys_loaded": 0, "errors": []},
            "prewarming": {"keys_loaded": 0, "errors": []},
            "tenants_processed": 0,
            "tenants_failed": 0,
        }

        start_time = datetime.utcnow()
        logger.info(f"Starting cache warming for {len(tenants)} tenants")

        for tenant in tenants:
            try:
                # Warm cost anomaly
                if self.cost_anomaly:
                    cost_keys = await self._warm_cost_anomaly(tenant)
                    results["cost_anomaly"]["keys_loaded"] += cost_keys

                # Warm routing
                if self.routing:
                    routing_keys = await self._warm_routing(tenant)
                    results["routing"]["keys_loaded"] += routing_keys

                # Warm recommendations
                if self.recommendations:
                    rec_keys = await self._warm_recommendations(tenant)
                    results["recommendations"]["keys_loaded"] += rec_keys

                # Warm prewarming (Holt-Winters state)
                if self.prewarming:
                    prewarm_keys = await self._warm_prewarming(tenant)
                    results["prewarming"]["keys_loaded"] += prewarm_keys

                results["tenants_processed"] += 1

            except Exception as e:
                logger.error(f"Failed to warm cache for tenant {tenant}: {e}")
                results["tenants_failed"] += 1
                self._warming_errors.append({"tenant": tenant, "error": str(e)})
                results["cost_anomaly"]["errors"].append({"tenant": tenant, "error": str(e)})

        end_time = datetime.utcnow()
        duration = (end_time - start_time).total_seconds()
        self._last_warming_time = end_time
        self._warming_in_progress = False

        results["duration_seconds"] = duration
        results["status"] = "completed"

        logger.info(
            f"Cache warming complete: {results['tenants_processed']} tenants, "
            f"{results['cost_anomaly']['keys_loaded'] + results['routing']['keys_loaded'] + results['recommendations']['keys_loaded'] + results['prewarming']['keys_loaded']} total keys loaded, "
            f"duration: {duration:.2f}s, errors: {len(self._warming_errors)}"
        )

        return results

    async def _warm_cost_anomaly(self, tenant: str) -> int:
        """Warm cost anomaly statistics from database.

        Args:
            tenant: Tenant ID

        Returns:
            Number of keys loaded
        """
        if not self.cost_anomaly:
            return 0

        try:
            # Check if repository has warm_cache method
            if hasattr(self.cost_anomaly, 'warm_cache'):
                return await self.cost_anomaly.warm_cache(tenant)

            # Otherwise, try to get recent stats and cache them
            stats_key = f"ml:cost_anomaly:stats:{tenant}"
            existing = await self.redis.get(stats_key)

            if not existing:
                # Try to load from database and cache
                logger.debug(f"No cached cost anomaly stats for tenant {tenant}")
                return 0

            return 1

        except Exception as e:
            logger.warning(f"Failed to warm cost anomaly cache for {tenant}: {e}")
            return 0

    async def _warm_routing(self, tenant: str) -> int:
        """Warm Thompson Sampling routing distributions.

        Args:
            tenant: Tenant ID

        Returns:
            Number of keys loaded
        """
        if not self.routing:
            return 0

        try:
            if hasattr(self.routing, 'warm_cache'):
                return await self.routing.warm_cache(tenant)

            routing_key = f"ml:routing:distributions:{tenant}"
            existing = await self.redis.get(routing_key)

            if not existing:
                logger.debug(f"No cached routing distributions for tenant {tenant}")
                return 0

            return 1

        except Exception as e:
            logger.warning(f"Failed to warm routing cache for {tenant}: {e}")
            return 0

    async def _warm_recommendations(self, tenant: str) -> int:
        """Warm ALS recommendations data.

        Args:
            tenant: Tenant ID

        Returns:
            Number of keys loaded
        """
        if not self.recommendations:
            return 0

        try:
            if hasattr(self.recommendations, 'warm_cache'):
                return await self.recommendations.warm_cache(tenant)

            rec_key = f"ml:recommendations:model:{tenant}"
            existing = await self.redis.get(rec_key)

            if not existing:
                logger.debug(f"No cached recommendation model for tenant {tenant}")
                return 0

            return 1

        except Exception as e:
            logger.warning(f"Failed to warm recommendations cache for {tenant}: {e}")
            return 0

    async def _warm_prewarming(self, tenant: str) -> int:
        """Warm Holt-Winters forecasting state.

        Args:
            tenant: Tenant ID

        Returns:
            Number of keys loaded
        """
        if not self.prewarming:
            return 0

        try:
            if hasattr(self.prewarming, 'warm_cache'):
                return await self.prewarming.warm_cache(tenant)

            prewarm_key = f"ml:prewarming:state:{tenant}"
            existing = await self.redis.get(prewarm_key)

            if not existing:
                logger.debug(f"No cached prewarming state for tenant {tenant}")
                return 0

            return 1

        except Exception as e:
            logger.warning(f"Failed to warm prewarming cache for {tenant}: {e}")
            return 0

    def get_status(self) -> dict:
        """Get cache warming status.

        Returns:
            Dictionary with warming status information
        """
        return {
            "is_warming": self._warming_in_progress,
            "last_warming_time": self._last_warming_time.isoformat() if self._last_warming_time else None,
            "errors": self._warming_errors,
        }


# Global instance
_cache_warming_service: Optional[CacheWarmingService] = None


def get_cache_warming_service() -> Optional[CacheWarmingService]:
    """Get the global cache warming service instance.

    Returns:
        CacheWarmingService instance or None if not initialized
    """
    return _cache_warming_service


async def init_cache_warming_service(
    redis_client=None,
    cost_anomaly_repo=None,
    routing_repo=None,
    recommendations_repo=None,
    prewarming_repo=None,
    database_url: Optional[str] = None,
    orchestrator_url: Optional[str] = None,
    orchestrator_api_key: Optional[str] = None,
) -> CacheWarmingService:
    """Initialize the global cache warming service.

    Args:
        redis_client: Redis client instance (optional, will fetch if not provided)
        cost_anomaly_repo: Cost anomaly repository (optional)
        routing_repo: Routing repository (optional)
        recommendations_repo: Recommendations repository (optional)
        prewarming_repo: Prewarming repository (optional)
        database_url: PostgreSQL connection string for direct DB access (optional)
        orchestrator_url: Orchestrator API URL for tenant lookup (optional)
        orchestrator_api_key: API key for orchestrator authentication (optional)

    Returns:
        CacheWarmingService instance
    """
    global _cache_warming_service

    if redis_client is None:
        from ..services.redis_client import get_redis_client
        redis_client = get_redis_client()

    if cost_anomaly_repo is None:
        try:
            from ..services.cost_anomaly import get_cost_anomaly_service
            cost_anomaly_repo = get_cost_anomaly_service()
        except Exception:
            pass

    if routing_repo is None:
        try:
            from ..services.thompson_routing import get_routing_service
            routing_repo = get_routing_service()
        except Exception:
            pass

    if recommendations_repo is None:
        try:
            from ..services.recommendations import get_recommendations_service
            recommendations_repo = get_recommendations_service()
        except Exception:
            pass

    if prewarming_repo is None:
        try:
            from ..services.prewarming import get_prewarming_service
            prewarming_repo = get_prewarming_service()
        except Exception:
            pass

    _cache_warming_service = CacheWarmingService(
        redis_client=redis_client,
        cost_anomaly_repo=cost_anomaly_repo,
        routing_repo=routing_repo,
        recommendations_repo=recommendations_repo,
        prewarming_repo=prewarming_repo,
        database_url=database_url,
        orchestrator_url=orchestrator_url,
        orchestrator_api_key=orchestrator_api_key,
    )

    return _cache_warming_service


async def warm_caches_for_active_tenants(
    tenants: list[str],
    timeout_seconds: float = 300.0,
) -> dict:
    """Warm caches for a list of tenants.

    Args:
        tenants: List of tenant IDs
        timeout_seconds: Maximum time to spend warming caches

    Returns:
        Dictionary with warming results
    """
    service = get_cache_warming_service()
    if not service:
        return {"status": "not_initialized"}

    try:
        # Run with timeout
        results = await asyncio.wait_for(
            service.warm_all(tenants),
            timeout=timeout_seconds
        )
        return results
    except asyncio.TimeoutError:
        logger.warning(f"Cache warming timed out after {timeout_seconds}s")
        return {
            "status": "timeout",
            "timeout_seconds": timeout_seconds,
            "partial_results": service.get_status(),
        }
    except Exception as e:
        logger.error(f"Cache warming failed: {e}")
        return {"status": "error", "error": str(e)}
