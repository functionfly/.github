"""RunPod Cluster Client for AI Gateway.

This module provides integration with the RunPod cluster infrastructure
for routing inference requests to optimal GPU instances.
"""

import asyncio
import logging
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional
from urllib.parse import urljoin

import httpx

from ..config import Settings, get_settings
from ..models.schemas import ClusterHealthInfo, HealthStatus

logger = logging.getLogger(__name__)


class ClusterMode(str, Enum):
    """Cluster selection mode."""

    LEAST_LOADED = "least_loaded"
    LATENCY_BASED = "latency_based"
    GEO_AWARE = "geo_aware"
    WEIGHTED_ROUND_ROBIN = "weighted_round_robin"


@dataclass
class ClusterEndpoint:
    """Represents a cluster endpoint for inference."""

    cluster_id: str
    region: str
    gpu_type: str
    endpoint_url: str
    is_healthy: bool = True
    current_load: int = 0
    avg_latency_ms: float = 0.0
    error_rate: float = 0.0


@dataclass
class ClusterStats:
    """Cluster statistics."""

    cluster_id: str
    region: str
    gpu_type: str
    healthy_instances: int
    total_instances: int
    running_instances: int
    idle_instances: int
    failed_instances: int
    total_requests: int
    avg_latency_ms: float
    status: HealthStatus


class ClusterClient:
    """Client for interacting with RunPod cluster manager.

    Integrates with the Go-based ClusterManager via HTTP API
    to route inference requests to optimal GPU clusters.
    """

    def __init__(self, settings: Optional[Settings] = None):
        """Initialize cluster client.

        Args:
            settings: Application settings. If None, uses default.
        """
        self._settings = settings or get_settings()
        self._client: Optional[httpx.AsyncClient] = None
        self._clusters: Dict[str, ClusterEndpoint] = {}
        self._mode = ClusterMode.LEAST_LOADED
        self._preferred_region: Optional[str] = None
        self._retry_config = {
            "max_retries": 3,
            "base_delay": 0.5,
            "max_delay": 10.0,
        }
        self._lock = asyncio.Lock()

    @property
    def base_url(self) -> str:
        """Get cluster manager base URL."""
        return self._settings.RUNPOD_CLUSTER_URL.rstrip("/")

    @property
    def api_key(self) -> str:
        """Get RunPod API key."""
        return self._settings.RUNPOD_API_KEY

    async def _get_client(self) -> httpx.AsyncClient:
        """Get or create HTTP client."""
        if self._client is None:
            self._client = httpx.AsyncClient(
                base_url=self.base_url,
                timeout=httpx.Timeout(
                    connect=10.0,
                    read=self._settings.REQUEST_TIMEOUT_SECONDS,
                    write=30.0,
                    pool=60.0,
                ),
                headers={
                    "Authorization": f"Bearer {self.api_key}",
                    "Content-Type": "application/json",
                },
            )
        return self._client

    async def close(self) -> None:
        """Close HTTP client."""
        if self._client:
            await self._client.aclose()
            self._client = None

    async def _request_with_retry(
        self,
        method: str,
        path: str,
        json: Optional[Dict[str, Any]] = None,
        params: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Make HTTP request with exponential backoff retry.

        Args:
            method: HTTP method
            path: API path
            json: JSON request body
            params: Query parameters

        Returns:
            Response JSON

        Raises:
            httpx.HTTPStatusError: On HTTP error
        """
        client = await self._get_client()
        max_retries = self._retry_config["max_retries"]
        base_delay = self._retry_config["base_delay"]

        for attempt in range(max_retries + 1):
            try:
                response = await client.request(
                    method=method,
                    url=path,
                    json=json,
                    params=params,
                )
                response.raise_for_status()
                return response.json()
            except httpx.HTTPStatusError as e:
                if attempt == max_retries or e.response.status_code < 500:
                    raise
                delay = min(base_delay * (2**attempt), self._retry_config["max_delay"])
                logger.warning(
                    f"Request failed (attempt {attempt + 1}/{max_retries + 1}): "
                    f"{e}. Retrying in {delay:.1f}s..."
                )
                await asyncio.sleep(delay)
            except httpx.RequestError as e:
                if attempt == max_retries:
                    raise
                delay = min(base_delay * (2**attempt), self._retry_config["max_delay"])
                logger.warning(
                    f"Request error (attempt {attempt + 1}/{max_retries + 1}): "
                    f"{e}. Retrying in {delay:.1f}s..."
                )
                await asyncio.sleep(delay)

        raise RuntimeError("Max retries exceeded")

    async def get_clusters(self) -> List[ClusterStats]:
        """Get list of available clusters.

        Returns:
            List of cluster statistics
        """
        try:
            data = await self._request_with_retry("GET", "/api/v1/clusters")
            clusters = []
            for cluster_data in data.get("clusters", []):
                clusters.append(
                    ClusterStats(
                        cluster_id=cluster_data.get("id", ""),
                        region=cluster_data.get("region", ""),
                        gpu_type=cluster_data.get("gpu_type", ""),
                        healthy_instances=cluster_data.get("healthy_instances", 0),
                        total_instances=cluster_data.get("total_instances", 0),
                        running_instances=cluster_data.get("running_instances", 0),
                        idle_instances=cluster_data.get("idle_instances", 0),
                        failed_instances=cluster_data.get("failed_instances", 0),
                        total_requests=cluster_data.get("total_requests", 0),
                        avg_latency_ms=cluster_data.get("avg_latency_ms", 0.0),
                        status=self._map_status(cluster_data.get("status", "")),
                    )
                )
            return clusters
        except Exception as e:
            logger.error(f"Failed to get clusters: {e}")
            return []

    async def get_cluster(self, cluster_id: str) -> Optional[ClusterEndpoint]:
        """Get cluster by ID.

        Args:
            cluster_id: Cluster identifier

        Returns:
            Cluster endpoint or None if not found
        """
        # Check cache first
        if cluster_id in self._clusters:
            return self._clusters[cluster_id]

        try:
            data = await self._request_with_retry(
                "GET", f"/api/v1/clusters/{cluster_id}"
            )
            endpoint = ClusterEndpoint(
                cluster_id=data.get("id", cluster_id),
                region=data.get("region", ""),
                gpu_type=data.get("gpu_type", ""),
                endpoint_url=data.get("endpoint_url", ""),
                is_healthy=data.get("status") == "healthy",
                current_load=data.get("current_load", 0),
                avg_latency_ms=data.get("avg_latency_ms", 0.0),
                error_rate=data.get("error_rate", 0.0),
            )
            async with self._lock:
                self._clusters[cluster_id] = endpoint
            return endpoint
        except Exception as e:
            logger.error(f"Failed to get cluster {cluster_id}: {e}")
            return None

    async def select_cluster(
        self,
        preferred_region: Optional[str] = None,
        model: Optional[str] = None,
    ) -> Optional[ClusterEndpoint]:
        """Select optimal cluster based on load balancing strategy.

        Args:
            preferred_region: Preferred region for inference
            model: Model identifier (for GPU compatibility check)

        Returns:
            Selected cluster endpoint or None if no clusters available
        """
        clusters = await self.get_clusters()
        if not clusters:
            return None

        # Filter healthy clusters
        healthy = [c for c in clusters if c.status == HealthStatus.HEALTHY]
        if not healthy:
            # Try degraded clusters
            healthy = [c for c in clusters if c.status == HealthStatus.DEGRADED]

        if not healthy:
            return None

        # Apply selection strategy
        if self._mode == ClusterMode.LEAST_LOADED:
            # Select cluster with lowest load
            selected = min(
                healthy,
                key=lambda c: c.running_instances / max(c.total_instances, 1),
            )
        elif self._mode == ClusterMode.LATENCY_BASED:
            # Select cluster with lowest latency
            selected = min(healthy, key=lambda c: c.avg_latency_ms)
        elif self._mode == ClusterMode.GEO_AWARE:
            # Prefer region, then lowest load
            if preferred_region:
                regional = [c for c in healthy if c.region == preferred_region]
                if regional:
                    selected = min(
                        regional,
                        key=lambda c: c.running_instances / max(c.total_instances, 1),
                    )
                else:
                    selected = min(
                        healthy,
                        key=lambda c: c.running_instances / max(c.total_instances, 1),
                    )
            else:
                selected = min(
                    healthy,
                    key=lambda c: c.running_instances / max(c.total_instances, 1),
                )
        else:
            # Default: least loaded
            selected = min(
                healthy, key=lambda c: c.running_instances / max(c.total_instances, 1)
            )

        return await self.get_cluster(selected.cluster_id)

    async def route_inference(
        self,
        model: str,
        input_data: str,
        parameters: Optional[Dict[str, Any]] = None,
        preferred_region: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Route inference request to optimal cluster.

        Args:
            model: Model identifier
            input_data: Base64-encoded input data
            parameters: Inference parameters
            preferred_region: Preferred region

        Returns:
            Inference result from cluster
        """
        cluster = await self.select_cluster(preferred_region=preferred_region, model=model)
        if not cluster:
            raise RuntimeError("No available clusters for inference")

        client = await self._get_client()
        try:
            response = await client.post(
                urljoin(cluster.endpoint_url, "/v1/infer"),
                json={
                    "model": model,
                    "input": input_data,
                    "parameters": parameters or {},
                },
                timeout=httpx.Timeout(
                    connect=10.0,
                    read=self._settings.REQUEST_TIMEOUT_SECONDS,
                    write=30.0,
                ),
            )
            response.raise_for_status()
            result = response.json()
            result["region"] = cluster.region
            result["cluster_id"] = cluster.cluster_id
            return result
        except httpx.HTTPStatusError as e:
            logger.error(
                f"Inference request failed on cluster {cluster.cluster_id}: {e}"
            )
            raise
        except Exception as e:
            logger.error(f"Inference request error: {e}")
            raise

    async def health_check(self) -> bool:
        """Check if cluster manager is reachable.

        Returns:
            True if healthy, False otherwise
        """
        try:
            client = await self._get_client()
            response = await client.get(
                "/health",
                timeout=httpx.Timeout(connect=5.0, read=5.0),
            )
            return response.status_code == 200
        except Exception as e:
            logger.error(f"Cluster manager health check failed: {e}")
            return False

    async def get_cluster_health_info(self) -> List[ClusterHealthInfo]:
        """Get health info for all clusters.

        Returns:
            List of cluster health information
        """
        clusters = await self.get_clusters()
        return [
            ClusterHealthInfo(
                cluster_id=c.cluster_id,
                region=c.region,
                status=c.status,
                healthy_instances=c.healthy_instances,
                total_instances=c.total_instances,
                avg_latency_ms=c.avg_latency_ms,
                error_rate=c.error_rate,
            )
            for c in clusters
        ]

    def _map_status(self, status: str) -> HealthStatus:
        """Map cluster status string to HealthStatus enum.

        Args:
            status: Status string from API

        Returns:
            HealthStatus enum value
        """
        status_lower = status.lower()
        if status_lower == "healthy":
            return HealthStatus.HEALTHY
        elif status_lower == "degraded":
            return HealthStatus.DEGRADED
        else:
            return HealthStatus.UNHEALTHY

    def set_mode(self, mode: ClusterMode) -> None:
        """Set cluster selection mode.

        Args:
            mode: New selection mode
        """
        self._mode = mode
        logger.info(f"Cluster selection mode set to {mode.value}")

    def set_preferred_region(self, region: str) -> None:
        """Set preferred region for inference.

        Args:
            region: Preferred region code
        """
        self._preferred_region = region
        logger.info(f"Preferred region set to {region}")

    def clear_cache(self) -> None:
        """Clear cached cluster information."""
        self._clusters.clear()
        logger.info("Cluster cache cleared")


# Global instance
_cluster_client: Optional[ClusterClient] = None


def get_cluster_client() -> ClusterClient:
    """Get global cluster client instance.

    Returns:
        ClusterClient singleton
    """
    global _cluster_client
    if _cluster_client is None:
        _cluster_client = ClusterClient()
    return _cluster_client
