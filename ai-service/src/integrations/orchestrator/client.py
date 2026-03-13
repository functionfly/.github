"""HTTP client for Go orchestrator API.

Provides integration with the FunctionFly Go orchestrator
for function execution and management.
"""

import logging
from typing import Optional, Dict, Any
from datetime import datetime

import httpx

from ...config import settings

logger = logging.getLogger(__name__)


class OrchestratorClient:
    """HTTP client for the Go orchestrator API.

    Handles communication with the orchestrator for:
    - Function execution
    - Prewarming triggers
    - Function metadata
    - Execution results
    """

    def __init__(
        self,
        base_url: Optional[str] = None,
        api_key: Optional[str] = None,
        timeout: int = 30,
    ):
        self.base_url = base_url or settings.orchestrator_url
        self.api_key = api_key or settings.orchestrator_api_key
        self.timeout = timeout
        self._client: Optional[httpx.AsyncClient] = None

    async def get_client(self) -> httpx.AsyncClient:
        """Get or create HTTP client."""
        if self._client is None:
            headers = {}
            if self.api_key:
                headers["Authorization"] = f"Bearer {self.api_key}"

            self._client = httpx.AsyncClient(
                base_url=self.base_url,
                timeout=self.timeout,
                headers=headers,
            )
        return self._client

    async def close(self):
        """Close the HTTP client."""
        if self._client:
            await self._client.aclose()
            self._client = None

    async def health_check(self) -> bool:
        """Check if orchestrator is healthy.

        Returns:
            True if orchestrator is available
        """
        try:
            client = await self.get_client()
            response = await client.get("/health")
            return response.status_code == 200
        except Exception as e:
            logger.warning(f"Orchestrator health check failed: {e}")
            return False

    async def get_function(self, function_id: str) -> Optional[Dict[str, Any]]:
        """Get function metadata.

        Args:
            function_id: The function ID

        Returns:
            Function metadata or None if not found
        """
        try:
            client = await self.get_client()
            response = await client.get(f"/api/v1/functions/{function_id}")

            if response.status_code == 404:
                return None

            response.raise_for_status()
            return response.json()
        except Exception as e:
            logger.error(f"Failed to get function: {e}")
            return None

    async def execute_function(
        self,
        function_id: str,
        payload: Dict[str, Any],
        timeout: Optional[int] = None,
    ) -> Optional[Dict[str, Any]]:
        """Execute a function.

        Args:
            function_id: The function ID
            payload: Function input payload
            timeout: Optional timeout override

        Returns:
            Execution result or None on failure
        """
        try:
            client = await self.get_client()

            request_timeout = timeout or self.timeout
            response = await client.post(
                f"/api/v1/functions/{function_id}/execute",
                json=payload,
                timeout=request_timeout,
            )

            response.raise_for_status()
            return response.json()
        except Exception as e:
            logger.error(f"Failed to execute function: {e}")
            return None

    async def trigger_prewarm(
        self,
        function_id: str,
        instances: int = 1,
        edge: Optional[str] = None,
    ) -> bool:
        """Trigger function prewarming.

        Args:
            function_id: The function ID
            instances: Number of instances to warm
            edge: Optional edge provider

        Returns:
            True if successful
        """
        try:
            client = await self.get_client()

            payload = {
                "instances": instances,
            }
            if edge:
                payload["edge"] = edge

            response = await client.post(
                f"/api/v1/functions/{function_id}/prewarm",
                json=payload,
            )

            return response.status_code in (200, 201)
        except Exception as e:
            logger.error(f"Failed to trigger prewarm: {e}")
            return False

    async def get_execution(
        self,
        execution_id: str,
    ) -> Optional[Dict[str, Any]]:
        """Get execution result.

        Args:
            execution_id: The execution ID

        Returns:
            Execution result or None if not found
        """
        try:
            client = await self.get_client()
            response = await client.get(f"/api/v1/executions/{execution_id}")

            if response.status_code == 404:
                return None

            response.raise_for_status()
            return response.json()
        except Exception as e:
            logger.error(f"Failed to get execution: {e}")
            return None

    async def record_execution_metrics(
        self,
        function_id: str,
        latency_ms: float,
        cold_start: bool,
        success: bool,
        error: Optional[str] = None,
    ) -> bool:
        """Record execution metrics.

        Args:
            function_id: The function ID
            latency_ms: Execution latency
            cold_start: Whether this was a cold start
            success: Whether execution succeeded
            error: Optional error message

        Returns:
            True if successful
        """
        try:
            client = await self.get_client()

            payload = {
                "function_id": function_id,
                "latency_ms": latency_ms,
                "cold_start": cold_start,
                "success": success,
                "timestamp": datetime.utcnow().isoformat(),
            }
            if error:
                payload["error"] = error

            response = await client.post(
                "/api/v1/metrics/executions",
                json=payload,
            )

            return response.status_code in (200, 201)
        except Exception as e:
            logger.error(f"Failed to record metrics: {e}")
            return False

    async def get_functions_by_tenant(
        self,
        tenant_id: str,
        limit: int = 50,
    ) -> list[Dict[str, Any]]:
        """Get functions for a tenant.

        Args:
            tenant_id: The tenant ID
            limit: Maximum number of functions

        Returns:
            List of functions
        """
        try:
            client = await self.get_client()
            response = await client.get(
                "/api/v1/functions",
                params={"tenant_id": tenant_id, "limit": limit},
            )

            response.raise_for_status()
            data = response.json()
            return data.get("functions", [])
        except Exception as e:
            logger.error(f"Failed to get functions: {e}")
            return []

    async def update_function_runtime(
        self,
        function_id: str,
        memory_mb: Optional[int] = None,
        timeout_ms: Optional[int] = None,
        network_enabled: Optional[bool] = None,
        runtime: Optional[str] = None,
    ) -> bool:
        """Update runtime configuration for a function.

        Loads current function and merges updates so the orchestrator
        receives a full runtime payload (avoids zeroing omitted fields).

        Args:
            function_id: The function ID
            memory_mb: Memory in MB
            timeout_ms: Timeout in milliseconds
            network_enabled: Whether network is enabled
            runtime: Runtime identifier

        Returns:
            True if successful
        """
        try:
            current = await self.get_function(function_id)
            if not current:
                return False
            default_timeout_ms = current.get("timeout_ms") or (current.get("timeout_seconds") or 30) * 1000
            payload: Dict[str, Any] = {
                "runtime": runtime if runtime is not None else (current.get("runtime") or "python"),
                "memory_mb": memory_mb if memory_mb is not None else (current.get("memory_mb") or 256),
                "timeout_ms": timeout_ms if timeout_ms is not None else default_timeout_ms,
                "network_enabled": network_enabled if network_enabled is not None else current.get("network_enabled", False),
            }
            client = await self.get_client()
            response = await client.put(
                f"/api/v1/functions/{function_id}/runtime",
                json=payload,
            )
            return response.status_code in (200, 201)
        except Exception as e:
            logger.error(f"Failed to update function runtime: {e}")
            return False


# Global orchestrator client instance
_orchestrator_client: Optional[OrchestratorClient] = None


def get_orchestrator_client() -> OrchestratorClient:
    """Get the global orchestrator client instance.

    Returns:
        The OrchestratorClient instance
    """
    global _orchestrator_client
    if _orchestrator_client is None:
        _orchestrator_client = OrchestratorClient()
    return _orchestrator_client
