"""Context builder for chat service.

Builds context from infrastructure data for chat responses.
"""

import json
import logging
from typing import Optional, Any, Dict, List
from datetime import datetime, timedelta

import redis.asyncio as redis

from ...config import settings
from ...integrations.orchestrator.client import get_orchestrator_client

logger = logging.getLogger(__name__)


class ContextBuilder:
    """Builds context from infrastructure data for chat queries."""

    def __init__(self):
        self._redis: Optional[redis.Redis] = None
        self._orchestrator = get_orchestrator_client()

    async def get_redis(self) -> Optional[redis.Redis]:
        """Get Redis connection."""
        if self._redis is None:
            try:
                self._redis = redis.from_url(
                    settings.redis_url,
                    encoding="utf-8",
                    decode_responses=True,
                )
                await self._redis.ping()
            except Exception as e:
                logger.warning(f"Failed to connect to Redis: {e}")
                self._redis = None
        return self._redis

    async def build_context(
        self,
        user_id: str,
        intent: str,
        query: str,
        tenant_id: Optional[str] = None,
    ) -> str:
        """Build context string for the given query.

        Args:
            user_id: The user ID
            intent: The classified intent
            query: The user's query
            tenant_id: Optional tenant ID; when set, functions are fetched from the orchestrator

        Returns:
            Context string for the LLM
        """
        context_parts = []

        # Get user's functions (from orchestrator when tenant_id is set, else cache-only)
        functions = await self._get_user_functions(user_id, tenant_id=tenant_id)
        if functions:
            context_parts.append(f"Deployed functions: {len(functions)} functions")
            # Add function names
            fn_names = [f.get("name", "unknown") for f in functions[:10]]
            context_parts.append(f"Function names: {', '.join(fn_names)}")
            if len(functions) > 10:
                context_parts.append(f"... and {len(functions) - 10} more")

        # Add intent-specific context
        if intent == "query_intent":
            # Get metrics for functions
            metrics = await self._get_function_metrics(user_id)
            if metrics:
                context_parts.append(f"\nRecent metrics: {metrics}")

        elif intent == "explain_intent":
            # Get latency data
            latency_data = await self._get_latency_data(user_id)
            if latency_data:
                context_parts.append(f"\nLatency analysis: {latency_data}")

        elif intent == "debugging_intent":
            # Get recent errors
            errors = await self._get_recent_errors(user_id)
            if errors:
                context_parts.append(f"\nRecent errors: {errors}")

        elif intent == "optimization_intent":
            # Get resource usage
            resources = await self._get_resource_usage(user_id)
            if resources:
                context_parts.append(f"\nResource usage: {resources}")

        return "\n".join(context_parts) if context_parts else "No specific context available."

    async def _get_user_functions(
        self,
        user_id: str,
        tenant_id: Optional[str] = None,
    ) -> List[Dict[str, Any]]:
        """Get user's deployed functions.

        When tenant_id is provided, fetches from the orchestrator and caches by tenant.
        Otherwise returns cached data only (if any), for backward compatibility.
        """
        try:
            if tenant_id:
                redis_client = await self.get_redis()
                cache_key = f"chat:functions:tenant:{tenant_id}"
                if redis_client:
                    cached = await redis_client.get(cache_key)
                    if cached:
                        return json.loads(cached)
                functions = await self._orchestrator.get_functions_by_tenant(
                    tenant_id, limit=50
                )
                if redis_client and functions:
                    await redis_client.setex(
                        cache_key, 60, json.dumps([self._fn_summary(f) for f in functions])
                    )
                return [self._fn_summary(f) for f in functions]

            # No tenant_id: use legacy cache key by user_id only
            redis_client = await self.get_redis()
            if redis_client:
                cache_key = f"chat:functions:{user_id}"
                cached = await redis_client.get(cache_key)
                if cached:
                    return json.loads(cached)
        except Exception as e:
            logger.warning(f"Failed to get user functions: {e}")
        return []

    @staticmethod
    def _fn_summary(f: Dict[str, Any]) -> Dict[str, Any]:
        """Reduce function payload for context/cache."""
        return {
            "name": f.get("name", "unknown"),
            "id": f.get("id"),
            "runtime": f.get("runtime"),
        }

    async def _get_function_metrics(self, user_id: str) -> Optional[str]:
        """Get function metrics summary."""
        try:
            redis_client = await self.get_redis()
            if redis_client:
                cache_key = f"chat:metrics:{user_id}"
                cached = await redis_client.get(cache_key)
                if cached:
                    return cached
        except Exception as e:
            logger.warning(f"Failed to get function metrics: {e}")
        return None

    async def _get_latency_data(self, user_id: str) -> Optional[str]:
        """Get latency data for explanation queries."""
        try:
            redis_client = await self.get_redis()
            if redis_client:
                cache_key = f"chat:latency:{user_id}"
                cached = await redis_client.get(cache_key)
                if cached:
                    return cached
        except Exception as e:
            logger.warning(f"Failed to get latency data: {e}")
        return None

    async def _get_recent_errors(self, user_id: str) -> Optional[str]:
        """Get recent errors for debugging queries."""
        try:
            redis_client = await self.get_redis()
            if redis_client:
                cache_key = f"chat:errors:{user_id}"
                cached = await redis_client.get(cache_key)
                if cached:
                    return cached
        except Exception as e:
            logger.warning(f"Failed to get recent errors: {e}")
        return None

    async def _get_resource_usage(self, user_id: str) -> Optional[str]:
        """Get resource usage for optimization queries."""
        try:
            redis_client = await self.get_redis()
            if redis_client:
                cache_key = f"chat:resources:{user_id}"
                cached = await redis_client.get(cache_key)
                if cached:
                    return cached
        except Exception as e:
            logger.warning(f"Failed to get resource usage: {e}")
        return None

    async def build_error_context(
        self,
        function_id: str,
        error_message: str,
        stack_trace: Optional[str] = None,
    ) -> str:
        """Build context for error analysis.

        Args:
            function_id: The function ID with the error
            error_message: The error message
            stack_trace: Optional stack trace

        Returns:
            Context string for debugging
        """
        context_parts = [f"Function ID: {function_id}"]
        context_parts.append(f"Error: {error_message}")

        if stack_trace:
            # Extract key info from stack trace
            lines = stack_trace.strip().split("\n")
            if len(lines) > 3:
                context_parts.append(f"Stack trace (last 3 frames): {' | '.join(lines[-4:])}")
            else:
                context_parts.append(f"Stack trace: {stack_trace}")

        # Get recent executions of this function
        try:
            redis_client = await self.get_redis()
            if redis_client:
                cache_key = f"chat:executions:{function_id}"
                recent = await redis_client.get(cache_key)
                if recent:
                    context_parts.append(f"Recent executions: {recent}")
        except Exception as e:
            logger.warning(f"Failed to get recent executions: {e}")

        return "\n".join(context_parts)

    async def build_optimization_context(
        self,
        function_id: str,
    ) -> str:
        """Build context for optimization analysis.

        Args:
            function_id: The function ID to analyze

        Returns:
            Context string for optimization
        """
        context_parts = [f"Function ID: {function_id}"]

        try:
            # Get function config
            function_info = await self._orchestrator.get_function(function_id)
            if function_info:
                context_parts.append(f"Runtime: {function_info.get('runtime', 'unknown')}")
                context_parts.append(f"Memory: {function_info.get('memory_mb', 'unknown')}MB")
                context_parts.append(f"Timeout: {function_info.get('timeout_seconds', 'unknown')}s")

            # Get cached metrics
            redis_client = await self.get_redis()
            if redis_client:
                metrics_key = f"chat:func_metrics:{function_id}"
                metrics = await redis_client.get(metrics_key)
                if metrics:
                    context_parts.append(f"Metrics: {metrics}")
        except Exception as e:
            logger.warning(f"Failed to build optimization context: {e}")

        return "\n".join(context_parts)

    async def close(self):
        """Close Redis connection."""
        if self._redis:
            await self._redis.close()
            self._redis = None


# Global instance
_context_builder: Optional[ContextBuilder] = None


def get_context_builder() -> ContextBuilder:
    """Get the global context builder instance.

    Returns:
        The ContextBuilder instance
    """
    global _context_builder
    if _context_builder is None:
        _context_builder = ContextBuilder()
    return _context_builder
