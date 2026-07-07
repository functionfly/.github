"""ML Service Registry - thread-safe singleton management for ML services.

This module provides a centralized registry for ML services that supports:
- Thread-safe singleton access with proper locking
- Graceful lifecycle management (init/shutdown)
- Health checks for ML service dependencies
- Horizontal scaling via stateless service design

All ML services store state in Redis, making them safe for concurrent access
across multiple instances. The registry ensures proper initialization order
and cleanup on shutdown.
"""

import asyncio
import logging
import threading
from contextlib import asynccontextmanager
from typing import Any, AsyncGenerator, Dict, Optional, Type

logger = logging.getLogger(__name__)


class MLServiceRegistry:
    """Thread-safe registry for ML services.

    Provides singleton access to ML services with proper initialization,
    lifecycle management, and health checks. Services are initialized lazily
    on first access and properly closed on shutdown.
    """

    _instance: Optional["MLServiceRegistry"] = None
    _lock = threading.Lock()

    def __new__(cls) -> "MLServiceRegistry":
        if cls._instance is None:
            with cls._lock:
                if cls._instance is None:
                    cls._instance = super().__new__(cls)
                    cls._instance._initialized = False
        return cls._instance

    def __init__(self):
        if self._initialized:
            return
        self._services: Dict[str, Any] = {}
        self._service_lock = threading.Lock()
        self._init_lock = asyncio.Lock()
        self._initialized = True
        logger.info("MLServiceRegistry initialized")

    def register(self, name: str, service: Any) -> None:
        """Register an ML service instance.

        Args:
            name: Service name (e.g., 'cost_anomaly', 'thompson_routing')
            service: Service instance
        """
        with self._service_lock:
            self._services[name] = service
            logger.info(f"Registered ML service: {name}")

    def get(self, name: str) -> Optional[Any]:
        """Get a registered service by name.

        Args:
            name: Service name

        Returns:
            Service instance or None if not registered
        """
        with self._service_lock:
            return self._services.get(name)

    def get_or_create(
        self,
        name: str,
        factory: callable,
        *args,
        **kwargs
    ) -> Any:
        """Get a service or create it if not exists (thread-safe).

        Args:
            name: Service name
            factory: Factory function to create the service
            *args, **kwargs: Arguments to pass to factory

        Returns:
            Service instance
        """
        with self._service_lock:
            if name not in self._services:
                self._services[name] = factory(*args, **kwargs)
                logger.info(f"Created ML service via factory: {name}")
            return self._services[name]

    async def close_service(self, name: str) -> None:
        """Close a specific ML service.

        Args:
            name: Service name to close
        """
        service = self.get(name)
        if service and hasattr(service, 'close'):
            try:
                if asyncio.iscoroutinefunction(service.close):
                    await service.close()
                else:
                    service.close()
                logger.info(f"Closed ML service: {name}")
            except Exception as e:
                logger.error(f"Error closing ML service {name}: {e}")

    async def close_all(self) -> None:
        """Close all registered ML services."""
        with self._service_lock:
            service_names = list(self._services.keys())

        for name in service_names:
            await self.close_service(name)

        with self._service_lock:
            self._services.clear()
        logger.info("All ML services closed")

    async def health_check(self) -> Dict[str, bool]:
        """Run health check on all ML services.

        Returns:
            Dict mapping service name to health status
        """
        results = {}
        with self._service_lock:
            services = dict(self._services)

        for name, service in services.items():
            try:
                if hasattr(service, 'health_check'):
                    if asyncio.iscoroutinefunction(service.health_check):
                        await service.health_check()
                    else:
                        service.health_check()
                results[name] = True
            except Exception as e:
                logger.warning(f"Health check failed for {name}: {e}")
                results[name] = False

        return results


_registry = MLServiceRegistry()


def get_registry() -> MLServiceRegistry:
    """Get the global ML service registry.

    Returns:
        MLServiceRegistry singleton instance
    """
    return _registry


async def init_ml_services() -> None:
    """Initialize all ML services in the correct order.

    This should be called during application startup.
    Services are initialized lazily on first use, but this
    pre-warms them for faster first requests.
    """
    logger.info("Pre-initializing ML services...")

    from ..cost_anomaly import get_cost_anomaly_detector
    from ..thompson_routing import get_thompson_router
    from ..recommendations import get_recommendation_engine
    from ..prewarming.holt_winters import get_holt_winters_forecaster

    services = [
        ("cost_anomaly", get_cost_anomaly_detector),
        ("thompson_routing", get_thompson_router),
        ("recommendations", get_recommendation_engine),
        ("prewarming", get_holt_winters_forecaster),
    ]

    for name, getter in services:
        try:
            service = getter()
            _registry.register(name, service)
            logger.info(f"Pre-initialized ML service: {name}")
        except Exception as e:
            logger.error(f"Failed to pre-initialize ML service {name}: {e}")


async def shutdown_ml_services() -> None:
    """Shutdown all ML services gracefully.

    This should be called during application shutdown to ensure
    all connections are properly closed.
    """
    logger.info("Shutting down ML services...")
    await _registry.close_all()


@asynccontextmanager
async def ml_services_lifespan() -> AsyncGenerator[None, None]:
    """Async context manager for ML services lifecycle.

    Usage:
        async with ml_services_lifespan():
            # services are initialized
            ...
        # services are shut down
    """
    await init_ml_services()
    try:
        yield
    finally:
        await shutdown_ml_services()
