"""Recommendation engine for optimization service.

Generates optimization suggestions based on analysis.
"""

import logging
from typing import Optional, Dict, Any, List
from enum import Enum

from .analyzer import FunctionAnalyzer, get_function_analyzer
from .cost_calculator import CostCalculator, get_cost_calculator

logger = logging.getLogger(__name__)


class RecommendationPriority(str, Enum):
    """Priority levels for recommendations."""
    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"


class RecommendationCategory(str, Enum):
    """Categories of optimization recommendations."""
    PERFORMANCE = "performance"
    COST = "cost"
    RELIABILITY = "reliability"
    SCALABILITY = "scalability"


class RecommendationEngine:
    """Generates optimization recommendations."""

    def __init__(self):
        self._analyzer = get_function_analyzer()
        self._cost_calculator = get_cost_calculator()

    async def generate_recommendations(
        self,
        function_id: str,
    ) -> List[Dict[str, Any]]:
        """Generate optimization recommendations for a function.

        Args:
            function_id: The function ID

        Returns:
            List of recommendations with priority and savings
        """
        # Analyze the function
        analysis = await self._analyzer.analyze_function(function_id)

        if "error" in analysis:
            return []

        recommendations = []

        # Generate recommendations based on patterns and issues
        patterns = analysis.get("patterns", [])
        issues = analysis.get("issues", [])

        # Process each issue and generate recommendations
        for issue in issues:
            rec = self._generate_recommendation(issue, analysis)
            if rec:
                recommendations.append(rec)

        # Add general recommendations based on patterns
        for pattern in patterns:
            rec = self._generate_pattern_recommendation(pattern, analysis)
            if rec:
                recommendations.append(rec)

        # Calculate potential savings for each recommendation
        for rec in recommendations:
            savings = await self._cost_calculator.estimate_savings(
                function_id=function_id,
                recommendation_type=rec.get("type"),
                current_value=rec.get("current_value", 0),
            )
            rec["estimated_savings_monthly"] = savings.get("monthly_savings", 0)
            rec["savings_currency"] = "USD"

        # Sort by priority
        priority_order = {
            RecommendationPriority.CRITICAL: 0,
            RecommendationPriority.HIGH: 1,
            RecommendationPriority.MEDIUM: 2,
            RecommendationPriority.LOW: 3,
        }
        recommendations.sort(
            key=lambda x: priority_order.get(x.get("priority"), 99),
        )

        return recommendations

    def _generate_recommendation(
        self,
        issue: Dict[str, Any],
        analysis: Dict[str, Any],
    ) -> Optional[Dict[str, Any]]:
        """Generate a recommendation for an identified issue."""
        issue_type = issue.get("type")
        severity = issue.get("severity", "medium")

        rec_map = {
            "high_error_rate": {
                "type": "reduce_errors",
                "title": "Reduce Function Error Rate",
                "description": issue.get("recommendation", "Add error handling and logging"),
                "category": RecommendationCategory.RELIABILITY,
                "priority": self._severity_to_priority(severity),
                "current_value": analysis.get("metrics", {}).get("error_rate", 0),
                "target_value": 0.01,
                "action": "Review error logs and implement proper error handling",
            },
            "excessive_cold_starts": {
                "type": "reduce_cold_starts",
                "title": "Reduce Cold Start Rate",
                "description": "Enable prewarming to reduce cold start frequency",
                "category": RecommendationCategory.SCALABILITY,
                "priority": RecommendationPriority.HIGH if severity == "medium" else RecommendationPriority.MEDIUM,
                "current_value": analysis.get("metrics", {}).get("cold_start_rate", 0),
                "target_value": 0.1,
                "action": "Enable prewarming or increase minimum instances",
            },
            "possible_memory_pressure": {
                "type": "increase_memory",
                "title": "Increase Memory Allocation",
                "description": "More memory can improve performance without significant cost increase",
                "category": RecommendationCategory.PERFORMANCE,
                "priority": RecommendationPriority.MEDIUM,
                "current_value": analysis.get("memory_mb", 0),
                "target_value": min(analysis.get("memory_mb", 0) * 2, 4096),
                "action": "Increase memory allocation from {}MB to {}MB".format(
                    analysis.get("memory_mb", 0),
                    min(analysis.get("memory_mb", 0) * 2, 4096),
                ),
            },
            "latency_variance": {
                "type": "optimize_latency",
                "title": "Reduce Latency Variance",
                "description": "Identify outliers causing high P99 latency",
                "category": RecommendationCategory.PERFORMANCE,
                "priority": RecommendationPriority.LOW,
                "current_value": analysis.get("metrics", {}).get("p99_latency_ms", 0),
                "target_value": analysis.get("metrics", {}).get("avg_latency_ms", 0) * 2,
                "action": "Profile function execution to identify slow paths",
            },
        }

        return rec_map.get(issue_type)

    def _generate_pattern_recommendation(
        self,
        pattern: Dict[str, Any],
        analysis: Dict[str, Any],
    ) -> Optional[Dict[str, Any]]:
        """Generate a recommendation based on function patterns."""
        pattern_type = pattern.get("type")
        impact = pattern.get("potential_impact", "low")

        rec_map = {
            "high_volume": {
                "type": "implement_caching",
                "title": "Implement Response Caching",
                "description": "Cache responses for repeated requests to reduce costs",
                "category": RecommendationCategory.COST,
                "priority": RecommendationPriority.HIGH if impact == "high" else RecommendationPriority.MEDIUM,
                "current_value": 0,
                "target_value": 1,
                "action": "Add caching layer for idempotent requests",
            },
            "high_cold_start": {
                "type": "enable_prewarming",
                "title": "Enable Function Prewarming",
                "description": "Pre-warm function instances to reduce cold starts",
                "category": RecommendationCategory.SCALABILITY,
                "priority": RecommendationPriority.HIGH,
                "current_value": 0,
                "target_value": 1,
                "action": "Enable prewarming in function configuration",
            },
            "high_latency": {
                "type": "optimize_code",
                "title": "Optimize Function Code",
                "description": "Review and optimize code for better performance",
                "category": RecommendationCategory.PERFORMANCE,
                "priority": RecommendationPriority.MEDIUM,
                "current_value": analysis.get("metrics", {}).get("avg_latency_ms", 0),
                "target_value": analysis.get("metrics", {}).get("avg_latency_ms", 0) * 0.5,
                "action": "Profile function to identify performance bottlenecks",
            },
            "memory_constrained": {
                "type": "optimize_memory",
                "title": "Optimize Memory Usage",
                "description": "Reduce memory footprint or increase allocation",
                "category": RecommendationCategory.PERFORMANCE,
                "priority": RecommendationPriority.MEDIUM,
                "current_value": analysis.get("metrics", {}).get("average_memory_percent", 0),
                "target_value": 60,
                "action": "Review memory usage and optimize data structures",
            },
        }

        return rec_map.get(pattern_type)

    def _severity_to_priority(self, severity: str) -> RecommendationPriority:
        """Convert issue severity to recommendation priority."""
        mapping = {
            "high": RecommendationPriority.CRITICAL,
            "medium": RecommendationPriority.HIGH,
            "low": RecommendationPriority.MEDIUM,
        }
        return mapping.get(severity, RecommendationPriority.LOW)

    async def apply_recommendation(
        self,
        function_id: str,
        recommendation_id: str,
    ) -> Dict[str, Any]:
        """Apply a recommendation to a function via the orchestrator.

        Resolves the recommendation by id (index from list), then calls
        the orchestrator for automatable types (e.g. increase_memory,
        enable_prewarming). Other types return success with a manual-action message.

        Args:
            function_id: The function ID
            recommendation_id: The recommendation ID (index as string, e.g. "0", "1")

        Returns:
            Result dict with success, function_id, recommendation_id, message
        """
        from ...integrations.orchestrator.client import get_orchestrator_client

        recommendations_list = await self.generate_recommendations(function_id)
        idx = None
        if recommendation_id.isdigit():
            i = int(recommendation_id)
            if 0 <= i < len(recommendations_list):
                idx = i
        if idx is None:
            return {
                "success": False,
                "function_id": function_id,
                "recommendation_id": recommendation_id,
                "message": "Recommendation not found or invalid id",
            }

        rec = recommendations_list[idx]
        rec_type = rec.get("type") or ""
        target_value = rec.get("target_value")
        current_value = rec.get("current_value")
        orchestrator = get_orchestrator_client()

        if rec_type == "increase_memory" and (target_value is not None or current_value is not None):
            new_mb = int(target_value if target_value else min((current_value or 0) * 2, 4096))
            if new_mb > 0:
                ok = await orchestrator.update_function_runtime(function_id, memory_mb=new_mb)
                if ok:
                    return {
                        "success": True,
                        "function_id": function_id,
                        "recommendation_id": recommendation_id,
                        "message": f"Memory updated to {new_mb} MB",
                    }
            return {
                "success": False,
                "function_id": function_id,
                "recommendation_id": recommendation_id,
                "message": "Orchestrator rejected runtime update or endpoint unavailable",
            }

        if rec_type == "enable_prewarming":
            ok = await orchestrator.trigger_prewarm(function_id, instances=1)
            return {
                "success": True,
                "function_id": function_id,
                "recommendation_id": recommendation_id,
                "message": "Prewarm triggered" if ok else "Prewarm requested (trigger may be async)",
            }

        # Not automatable: return success with action text for user
        action = rec.get("action", "Review and apply manually.")
        return {
            "success": True,
            "function_id": function_id,
            "recommendation_id": recommendation_id,
            "message": f"Manual action: {action}",
        }


_recommendation_engine: Optional[RecommendationEngine] = None


def get_recommendation_engine() -> RecommendationEngine:
    """Get the global recommendation engine instance."""
    global _recommendation_engine
    if _recommendation_engine is None:
        _recommendation_engine = RecommendationEngine()
    return _recommendation_engine
