"""Cost calculator for optimization service.

Estimates cost savings from optimization recommendations.
"""

import logging
from typing import Optional, Dict, Any

from ...config import settings

logger = logging.getLogger(__name__)


# Cost estimation constants (example values - should be configurable)
COST_PER_INVOCATION = 0.0002  # $0.0002 per invocation
COST_PER_GB_SECOND = 0.00001667  # $0.00001667 per GB-second
COST_PER_MEMORY_MB_MONTH = 0.000625  # $0.000625 per MB-month


class CostCalculator:
    """Calculates cost savings from optimization recommendations."""

    def __init__(self):
        self._cost_per_invocation = COST_PER_INVOCATION
        self._cost_per_gb_second = COST_PER_GB_SECOND
        self._cost_per_memory_mb = COST_PER_MEMORY_MB_MONTH

    async def estimate_savings(
        self,
        function_id: str,
        recommendation_type: str,
        current_value: float = 0,
    ) -> Dict[str, Any]:
        """Estimate savings from a recommendation.

        Args:
            function_id: The function ID
            recommendation_type: Type of recommendation
            current_value: Current metric value

        Returns:
            Estimated savings information
        """
        savings_calculators = {
            "reduce_cold_starts": self._calc_cold_start_savings,
            "reduce_errors": self._calc_error_savings,
            "implement_caching": self._calc_caching_savings,
            "increase_memory": self._calc_memory_savings,
            "optimize_latency": self._calc_latency_savings,
            "optimize_code": self._calc_code_optimization_savings,
            "enable_prewarming": self._calc_prewarming_savings,
            "optimize_memory": self._calc_memory_optimization_savings,
        }

        calculator = savings_calculators.get(recommendation_type)
        if calculator:
            return calculator(current_value)

        return {
            "monthly_savings": 0,
            "description": "Unable to calculate savings",
        }

    def _calc_cold_start_savings(
        self,
        current_cold_start_rate: float,
    ) -> Dict[str, Any]:
        """Calculate savings from reducing cold starts."""
        # Cold starts add ~500ms average latency
        # Each cold start costs ~$0.0001 extra
        current_cold_starts = current_cold_start_rate
        target_cold_starts = 0.1  # Target 10%

        if current_cold_starts <= target_cold_starts:
            return {"monthly_savings": 0, "description": "Already at target"}

        # Estimate monthly savings
        # Assuming 100,000 invocations/month
        monthly_invocations = 100000
        cold_start_diff = current_cold_starts - target_cold_starts
        savings = cold_start_diff * monthly_invocations * 0.0001

        return {
            "monthly_savings": round(savings, 2),
            "description": f"Reduce cold start rate from {current_cold_starts:.1%} to {target_cold_starts:.1%}",
        }

    def _calc_error_savings(
        self,
        current_error_rate: float,
    ) -> Dict[str, Any]:
        """Calculate savings from reducing error rate."""
        # Errors waste compute resources
        # Assume each error costs 2x normal invocation
        monthly_invocations = 100000
        target_error_rate = 0.01  # Target 1%

        if current_error_rate <= target_error_rate:
            return {"monthly_savings": 0, "description": "Already at target"}

        error_diff = current_error_rate - target_error_rate
        wasted_invocations = error_diff * monthly_invocations
        savings = wasted_invocations * self._cost_per_invocation * 2

        return {
            "monthly_savings": round(savings, 2),
            "description": f"Reduce error rate from {current_error_rate:.1%} to {target_error_rate:.1%}",
        }

    def _calc_caching_savings(
        self,
        current_value: float,
    ) -> Dict[str, Any]:
        """Calculate savings from implementing caching."""
        # Caching can reduce invocation costs by 50-80%
        # Assume 60% reduction in compute for cached responses
        monthly_invocations = 100000
        cache_hit_rate = 0.6  # Assume 60% cache hit rate
        avg_compute_cost = 0.0005  # Average compute cost per invocation

        savings = monthly_invocations * cache_hit_rate * avg_compute_cost * 0.5

        return {
            "monthly_savings": round(savings, 2),
            "description": "Implement response caching for frequently called functions",
        }

    def _calc_memory_savings(
        self,
        current_memory_mb: int,
    ) -> Dict[str, Any]:
        """Calculate cost implications of memory changes."""
        # More memory = more cost, but may improve performance
        # Calculate the cost difference
        new_memory = min(current_memory_mb * 2, 4096)

        current_monthly_cost = current_memory_mb * self._cost_per_memory_mb * 30
        new_monthly_cost = new_memory * self._cost_per_memory_mb * 30

        # But improved performance may reduce total cost
        # Assume 20% reduction in invocations due to better performance
        monthly_invocations = 100000
        invocation_savings = monthly_invocations * 0.2 * self._cost_per_invocation

        net_savings = invocation_savings - (new_monthly_cost - current_monthly_cost)

        return {
            "monthly_savings": round(net_savings, 2),
            "description": f"Increase memory from {current_memory_mb}MB to {new_memory}MB",
        }

    def _calc_latency_savings(
        self,
        current_latency_ms: float,
    ) -> Dict[str, Any]:
        """Calculate savings from reducing latency."""
        # Lower latency = less compute time = lower cost
        monthly_invocations = 100000
        target_latency = current_latency_ms * 0.5

        # Assume 1ms compute cost is $0.000001
        latency_savings_per_inv = (current_latency_ms - target_latency) / 1000 * 0.00001
        savings = monthly_invocations * latency_savings_per_inv

        return {
            "monthly_savings": round(savings, 2),
            "description": f"Reduce latency from {current_latency_ms:.0f}ms to {target_latency:.0f}ms",
        }

    def _calc_code_optimization_savings(
        self,
        current_value: float,
    ) -> Dict[str, Any]:
        """Calculate savings from code optimization."""
        # Similar to latency savings but more general
        monthly_invocations = 100000
        avg_compute_cost = 0.0005

        # Assume 30% reduction in compute costs
        savings = monthly_invocations * avg_compute_cost * 0.3

        return {
            "monthly_savings": round(savings, 2),
            "description": "Optimize function code for better performance",
        }

    def _calc_prewarming_savings(
        self,
        current_value: float,
    ) -> Dict[str, Any]:
        """Calculate savings from prewarming."""
        # Prewarming increases baseline cost but reduces cold start costs
        # Net effect depends on invocation pattern
        monthly_invocations = 100000

        # Prewarming adds ~$5/month but saves on cold start penalties
        prewarming_cost = 5
        cold_start_savings = monthly_invocations * 0.3 * 0.0001  # 30% cold start rate

        net_savings = cold_start_savings - prewarming_cost

        return {
            "monthly_savings": round(net_savings, 2),
            "description": "Enable prewarming to reduce cold start latency",
        }

    def _calc_memory_optimization_savings(
        self,
        current_memory_percent: float,
    ) -> Dict[str, Any]:
        """Calculate savings from memory optimization."""
        if current_memory_percent < 80:
            return {"monthly_savings": 0, "description": "Memory usage is optimal"}

        # Assume we can reduce memory allocation by 20%
        # while maintaining performance
        monthly_invocations = 100000
        avg_memory_mb = 512  # Assume average

        memory_savings = avg_memory_mb * 0.2 * self._cost_per_memory_mb * 30

        return {
            "monthly_savings": round(memory_savings, 2),
            "description": "Optimize memory usage to reduce allocation",
        }

    def get_cost_breakdown(
        self,
        function_id: str,
        metrics: Dict[str, Any],
    ) -> Dict[str, Any]:
        """Get detailed cost breakdown for a function.

        Args:
            function_id: The function ID
            metrics: Function metrics

        Returns:
            Cost breakdown
        """
        invocations = metrics.get("invocations_count", 0)
        avg_compute_time = metrics.get("avg_latency_ms", 100) / 1000  # Convert to seconds
        memory_mb = metrics.get("memory_mb", 256)

        # Calculate costs
        invocation_cost = invocations * self._cost_per_invocation
        compute_cost = invocations * avg_compute_time * self._cost_per_gb_second * (memory_mb / 1024)
        memory_cost = memory_mb * self._cost_per_memory_mb * 30

        return {
            "monthly_invocation_cost": round(invocation_cost, 2),
            "monthly_compute_cost": round(compute_cost, 2),
            "monthly_memory_cost": round(memory_cost, 2),
            "monthly_total_cost": round(invocation_cost + compute_cost + memory_cost, 2),
            "currency": "USD",
        }


_cost_calculator: Optional[CostCalculator] = None


def get_cost_calculator() -> CostCalculator:
    """Get the global cost calculator instance."""
    global _cost_calculator
    if _cost_calculator is None:
        _cost_calculator = CostCalculator()
    return _cost_calculator
