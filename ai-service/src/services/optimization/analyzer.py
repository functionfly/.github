"""Function analyzer for optimization service.

Analyzes function patterns and performance metrics.
"""

import json
import logging
from typing import Optional, Dict, Any, List
from datetime import datetime, timedelta

import redis.asyncio as redis

from ...config import settings
from ...integrations.orchestrator.client import get_orchestrator_client

logger = logging.getLogger(__name__)


class FunctionAnalyzer:
    """Analyzes function patterns for optimization opportunities."""

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

    async def analyze_function(
        self,
        function_id: str,
    ) -> Dict[str, Any]:
        """Analyze a function for optimization opportunities.

        Args:
            function_id: The function ID to analyze

        Returns:
            Analysis results with metrics and patterns
        """
        # Get function info
        func_info = await self._orchestrator.get_function(function_id)
        if not func_info:
            return {"error": "Function not found"}

        # Get metrics from cache/Redis
        metrics = await self._get_function_metrics(function_id)

        # Analyze patterns
        patterns = self._analyze_patterns(metrics, func_info)

        # Identify issues
        issues = self._identify_issues(metrics, func_info)

        return {
            "function_id": function_id,
            "function_name": func_info.get("name"),
            "runtime": func_info.get("runtime"),
            "memory_mb": func_info.get("memory_mb"),
            "timeout_seconds": func_info.get("timeout_seconds"),
            "metrics": metrics,
            "patterns": patterns,
            "issues": issues,
            "analyzed_at": datetime.utcnow().isoformat(),
        }

    async def _get_function_metrics(
        self,
        function_id: str,
    ) -> Dict[str, Any]:
        """Get function metrics from cache."""
        try:
            redis_client = await self.get_redis()
            if redis_client:
                key = f"optimize:metrics:{function_id}"
                cached = await redis_client.get(key)
                if cached:
                    return json.loads(cached)
        except Exception as e:
            logger.warning(f"Failed to get metrics: {e}")

        return self._get_default_metrics()

    def _get_default_metrics(self) -> Dict[str, Any]:
        """Return default metrics when none available."""
        return {
            "avg_latency_ms": 0,
            "p95_latency_ms": 0,
            "p99_latency_ms": 0,
            "invocations_count": 0,
            "error_rate": 0,
            "cold_start_rate": 0,
            "memory_usage_mb": 0,
            "average_memory_percent": 0,
        }

    def _analyze_patterns(
        self,
        metrics: Dict[str, Any],
        func_info: Dict[str, Any],
    ) -> List[Dict[str, Any]]:
        """Analyze function patterns."""
        patterns = []

        invocations = metrics.get("invocations_count", 0)
        if invocations > 1000:
            patterns.append({
                "type": "high_volume",
                "description": "Function is invoked frequently (1000+ times)",
                "potential_impact": "high",
            })

        cold_start_rate = metrics.get("cold_start_rate", 0)
        if cold_start_rate > 0.2:
            patterns.append({
                "type": "high_cold_start",
                "description": f"High cold start rate ({cold_start_rate:.1%})",
                "potential_impact": "high",
            })

        avg_latency = metrics.get("avg_latency_ms", 0)
        if avg_latency > 1000:
            patterns.append({
                "type": "high_latency",
                "description": f"Average latency is {avg_latency:.0f}ms",
                "potential_impact": "medium",
            })

        memory_mb = func_info.get("memory_mb", 0)
        memory_usage = metrics.get("average_memory_percent", 0)
        if memory_usage > 80 and memory_mb < 1024:
            patterns.append({
                "type": "memory_constrained",
                "description": f"Memory usage at {memory_usage:.0f}% of {memory_mb}MB allocation",
                "potential_impact": "medium",
            })

        return patterns

    def _identify_issues(
        self,
        metrics: Dict[str, Any],
        func_info: Dict[str, Any],
    ) -> List[Dict[str, Any]]:
        """Identify specific issues to address."""
        issues = []

        error_rate = metrics.get("error_rate", 0)
        if error_rate > 0.05:
            issues.append({
                "severity": "high",
                "type": "high_error_rate",
                "description": f"Error rate is {error_rate:.1%}",
                "recommendation": "Review error logs and add error handling",
            })

        cold_start_rate = metrics.get("cold_start_rate", 0)
        if cold_start_rate > 0.3:
            issues.append({
                "severity": "medium",
                "type": "excessive_cold_starts",
                "description": f"Cold start rate is {cold_start_rate:.1%}",
                "recommendation": "Consider prewarming or increasing instance count",
            })

        avg_latency = metrics.get("avg_latency_ms", 0)
        memory_mb = func_info.get("memory_mb", 0)
        if avg_latency > 2000 and memory_mb < 512:
            issues.append({
                "severity": "medium",
                "type": "possible_memory_pressure",
                "description": "High latency with low memory allocation",
                "recommendation": "Consider increasing memory for better performance",
            })

        p99_latency = metrics.get("p99_latency_ms", 0)
        avg_latency = metrics.get("avg_latency_ms", 0)
        if p99_latency > avg_latency * 5:
            issues.append({
                "severity": "low",
                "type": "latency_variance",
                "description": "High variance between average and P99 latency",
                "recommendation": "Identify and address outliers in execution",
            })

        return issues

    async def analyze_all_functions(
        self,
        tenant_id: str,
    ) -> List[Dict[str, Any]]:
        """Analyze all functions for a tenant.

        Args:
            tenant_id: The tenant ID

        Returns:
            List of analysis results
        """
        functions = await self._orchestrator.get_functions_by_tenant(
            tenant_id,
            limit=100,
        )

        results = []
        for func in functions:
            function_id = func.get("id")
            if function_id:
                analysis = await self.analyze_function(function_id)
                results.append(analysis)

        return results


_function_analyzer: Optional[FunctionAnalyzer] = None


def get_function_analyzer() -> FunctionAnalyzer:
    """Get the global function analyzer instance."""
    global _function_analyzer
    if _function_analyzer is None:
        _function_analyzer = FunctionAnalyzer()
    return _function_analyzer
