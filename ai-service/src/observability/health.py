"""Health checks for FlyMind AI Service.

This module provides health check functionality for the service
and all its dependencies.
"""

import asyncio
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import Any, Callable, Dict, List, Optional
import logging

logger = logging.getLogger(__name__)


class HealthStatus(str, Enum):
    """Health status levels."""
    HEALTHY = "healthy"
    DEGRADED = "degraded"
    UNHEALTHY = "unhealthy"
    UNKNOWN = "unknown"


@dataclass
class ComponentHealth:
    """Health status of a component."""
    name: str
    status: HealthStatus
    message: Optional[str] = None
    latency_ms: Optional[float] = None
    details: Dict[str, Any] = field(default_factory=dict)
    checked_at: datetime = field(default_factory=datetime.utcnow)

    def is_healthy(self) -> bool:
        """Check if the component is healthy."""
        return self.status == HealthStatus.HEALTHY


class HealthChecker:
    """Health checker for all service components."""

    def __init__(self):
        """Initialize the health checker."""
        self._logger = logging.getLogger(__name__)
        self._checks: Dict[str, Callable] = {}
        self._components: Dict[str, ComponentHealth] = {}

    def register_check(
        self,
        name: str,
        check_fn: Callable[[], bool],
    ) -> None:
        """Register a health check.

        Args:
            name: Check name
            check_fn: Async function that returns True if healthy
        """
        self._checks[name] = check_fn
        self._logger.info(f"Registered health check: {name}")

    async def check(self, component_name: str) -> ComponentHealth:
        """Check the health of a specific component.

        Args:
            component_name: Name of the component

        Returns:
            ComponentHealth
        """
        import time
        start_time = time.time()

        check_fn = self._checks.get(component_name)

        if not check_fn:
            return ComponentHealth(
                name=component_name,
                status=HealthStatus.UNKNOWN,
                message=f"No check registered for {component_name}",
            )

        try:
            # Run the check
            result = check_fn()

            # Handle both sync and async functions
            if asyncio.iscoroutine(result):
                result = await result

            latency_ms = (time.time() - start_time) * 1000

            if result:
                health = ComponentHealth(
                    name=component_name,
                    status=HealthStatus.HEALTHY,
                    latency_ms=latency_ms,
                )
            else:
                health = ComponentHealth(
                    name=component_name,
                    status=HealthStatus.UNHEALTHY,
                    message="Check returned False",
                    latency_ms=latency_ms,
                )

        except Exception as e:
            latency_ms = (time.time() - start_time) * 1000
            health = ComponentHealth(
                name=component_name,
                status=HealthStatus.UNHEALTHY,
                message=f"Check failed: {str(e)}",
                latency_ms=latency_ms,
                details={"error": str(e)},
            )

        self._components[component_name] = health
        return health

    async def check_all(self) -> Dict[str, ComponentHealth]:
        """Check health of all registered components.

        Returns:
            Dictionary of component health statuses
        """
        results = {}

        for component_name in self._checks.keys():
            results[component_name] = await self.check(component_name)

        return results

    def get_overall_status(self) -> HealthStatus:
        """Get overall health status.

        Returns:
            Overall HealthStatus
        """
        if not self._components:
            return HealthStatus.UNKNOWN

        statuses = [c.status for c in self._components.values()]

        if all(s == HealthStatus.HEALTHY for s in statuses):
            return HealthStatus.HEALTHY

        if any(s == HealthStatus.UNHEALTHY for s in statuses):
            return HealthStatus.UNHEALTHY

        return HealthStatus.DEGRADED

    def get_health_summary(self) -> Dict[str, Any]:
        """Get a summary of health status.

        Returns:
            Health summary dictionary
        """
        overall = self.get_overall_status()

        components = {
            name: {
                "status": comp.status.value,
                "message": comp.message,
                "latency_ms": comp.latency_ms,
                "checked_at": comp.checked_at.isoformat(),
            }
            for name, comp in self._components.items()
        }

        return {
            "status": overall.value,
            "components": components,
            "total_components": len(self._components),
            "healthy_components": sum(
                1 for c in self._components.values()
                if c.status == HealthStatus.HEALTHY
            ),
        }


# Default health checks
def create_default_checks(checker: HealthChecker) -> None:
    """Create default health checks.

    Args:
        checker: HealthChecker instance
    """
    # Check if Redis is available
    async def check_redis():
        try:
            import redis
            from ..config import settings

            r = redis.from_url(settings.redis_url)
            return r.ping()
        except Exception:
            return False

    # Check if database is available (asyncpg connect + SELECT 1)
    async def check_database():
        try:
            import asyncpg

            from ..config import settings

            conn = await asyncio.wait_for(
                asyncpg.connect(settings.database_url),
                timeout=2.0,
            )
            try:
                await conn.fetchval("SELECT 1")
                return True
            finally:
                await conn.close()
        except Exception:
            return False

    # Check if providers are available
    async def check_providers():
        try:
            from ..providers.manager import get_provider_manager

            manager = get_provider_manager()
            return len(manager._providers) > 0
        except Exception:
            return False

    checker.register_check("redis", check_redis)
    checker.register_check("database", check_database)
    checker.register_check("providers", check_providers)


# Global health checker
_health_checker: Optional[HealthChecker] = None


def get_health_checker() -> HealthChecker:
    """Get the global health checker.

    Returns:
        HealthChecker instance
    """
    global _health_checker

    if _health_checker is None:
        _health_checker = HealthChecker()
        create_default_checks(_health_checker)

    return _health_checker
