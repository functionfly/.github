"""Context collector for debugging service.

Gathers relevant context for error analysis.
"""

import json
import logging
from typing import Optional, Dict, Any, List

import redis.asyncio as redis

from ...config import settings
from ...integrations.orchestrator.client import get_orchestrator_client

logger = logging.getLogger(__name__)


class ContextCollector:
    """Collects context for debugging and error analysis."""

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

    async def collect_context(
        self,
        function_id: str,
        error_message: str,
        stack_trace: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Collect context for error analysis.

        Args:
            function_id: The function ID
            error_message: The error message
            stack_trace: Optional stack trace

        Returns:
            Context dictionary
        """
        context = {
            "function_id": function_id,
            "error_message": error_message,
            "stack_trace": stack_trace,
        }

        # Get function information
        func_info = await self._orchestrator.get_function(function_id)
        if func_info:
            context["function_info"] = {
                "name": func_info.get("name"),
                "runtime": func_info.get("runtime"),
                "memory_mb": func_info.get("memory_mb"),
                "timeout_seconds": func_info.get("timeout_seconds"),
            }

        # Get recent executions
        recent_executions = await self._get_recent_executions(function_id)
        context["recent_executions"] = recent_executions

        # Get recent logs
        recent_logs = await self._get_recent_logs(function_id)
        context["recent_logs"] = recent_logs

        # Get historical errors
        historical_errors = await self._get_historical_errors(function_id)
        context["historical_errors"] = historical_errors

        return context

    async def _get_recent_executions(
        self,
        function_id: str,
        limit: int = 5,
    ) -> List[Dict[str, Any]]:
        """Get recent executions of a function.

        Args:
            function_id: The function ID
            limit: Number of executions to retrieve

        Returns:
            List of recent executions
        """
        try:
            redis_client = await self.get_redis()
            if redis_client:
                key = f"debug:executions:{function_id}"
                executions = await redis_client.lrange(key, 0, limit - 1)
                return [json.loads(e) for e in executions if e]
        except Exception as e:
            logger.warning(f"Failed to get recent executions: {e}")
        return []

    async def _get_recent_logs(
        self,
        function_id: str,
        limit: int = 10,
    ) -> List[str]:
        """Get recent logs for a function.

        Args:
            function_id: The function ID
            limit: Number of logs to retrieve

        Returns:
            List of recent log entries
        """
        try:
            redis_client = await self.get_redis()
            if redis_client:
                key = f"debug:logs:{function_id}"
                logs = await redis_client.lrange(key, 0, limit - 1)
                return logs
        except Exception as e:
            logger.warning(f"Failed to get recent logs: {e}")
        return []

    async def _get_historical_errors(
        self,
        function_id: str,
        limit: int = 5,
    ) -> List[Dict[str, Any]]:
        """Get historical errors for a function.

        Args:
            function_id: The function ID
            limit: Number of errors to retrieve

        Returns:
            List of historical errors
        """
        try:
            redis_client = await self.get_redis()
            if redis_client:
                key = f"debug:errors:{function_id}"
                errors = await redis_client.lrange(key, 0, limit - 1)
                return [json.loads(e) for e in errors if e]
        except Exception as e:
            logger.warning(f"Failed to get historical errors: {e}")
        return []

    async def store_error_for_history(
        self,
        function_id: str,
        error_data: Dict[str, Any],
    ) -> bool:
        """Store error for historical tracking.

        Args:
            function_id: The function ID
            error_data: Error data to store

        Returns:
            True if stored successfully
        """
        try:
            redis_client = await self.get_redis()
            if redis_client:
                key = f"debug:errors:{function_id}"
                await redis_client.lpush(key, json.dumps(error_data))
                # Keep only last 100 errors
                await redis_client.ltrim(key, 0, 99)
                return True
        except Exception as e:
            logger.error(f"Failed to store error: {e}")
        return False

    async def store_execution_for_debug(
        self,
        function_id: str,
        execution_data: Dict[str, Any],
    ) -> bool:
        """Store execution data for debugging.

        Args:
            function_id: The function ID
            execution_data: Execution data to store

        Returns:
            True if stored successfully
        """
        try:
            redis_client = await self.get_redis()
            if redis_client:
                key = f"debug:executions:{function_id}"
                await redis_client.lpush(key, json.dumps(execution_data))
                # Keep only last 50 executions
                await redis_client.ltrim(key, 0, 49)
                return True
        except Exception as e:
            logger.error(f"Failed to store execution: {e}")
        return False

    async def store_log_for_debug(
        self,
        function_id: str,
        log_entry: str,
    ) -> bool:
        """Store log entry for debugging.

        Args:
            function_id: The function ID
            log_entry: Log entry to store

        Returns:
            True if stored successfully
        """
        try:
            redis_client = await self.get_redis()
            if redis_client:
                key = f"debug:logs:{function_id}"
                await redis_client.lpush(key, log_entry)
                # Keep only last 100 logs
                await redis_client.ltrim(key, 0, 99)
                return True
        except Exception as e:
            logger.error(f"Failed to store log: {e}")
        return False

    async def close(self):
        """Close Redis connection."""
        if self._redis:
            await self._redis.close()
            self._redis = None


# Global instance
_context_collector: Optional[ContextCollector] = None


def get_context_collector() -> ContextCollector:
    """Get the global context collector instance.

    Returns:
        The ContextCollector instance
    """
    global _context_collector
    if _context_collector is None:
        _context_collector = ContextCollector()
    return _context_collector
