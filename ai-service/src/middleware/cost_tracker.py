"""Cost tracking and limiting for FlyMind AI Service.

This module tracks API costs per tenant and provides limiting capabilities.
"""

import time
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from typing import Dict, List, Optional
import threading
import logging

logger = logging.getLogger(__name__)


class CostLimitExceeded(Exception):
    """Exception raised when cost limit is exceeded."""

    def __init__(
        self,
        message: str,
        tenant_id: str,
        current_cost: float,
        limit: float,
        period: str,
    ):
        """Initialize the exception.

        Args:
            message: Error message
            tenant_id: Tenant ID
            current_cost: Current period cost
            limit: Cost limit
            period: Time period (hour, day, month)
        """
        super().__init__(message)
        self.tenant_id = tenant_id
        self.current_cost = current_cost
        self.limit = limit
        self.period = period


@dataclass
class CostEntry:
    """A single cost entry."""
    tenant_id: str
    provider: str
    model: str
    input_tokens: int
    output_tokens: int
    cost: float
    timestamp: datetime = field(default_factory=datetime.utcnow)


@dataclass
class CostLimit:
    """Cost limit configuration."""
    limit: float  # in USD
    period: str = "day"  # hour, day, month
    alert_threshold: float = 0.8  # Alert at 80% of limit


class CostTracker:
    """Tracks and limits API costs per tenant."""

    def __init__(self):
        """Initialize the cost tracker."""
        self._logger = logging.getLogger(__name__)
        self._lock = threading.Lock()

        # Cost limits per tenant
        self._limits: Dict[str, CostLimit] = {}

        # Cost entries
        self._entries: List[CostEntry] = []
        self._max_entries = 100000

        # Default limits
        self._default_limit = CostLimit(
            limit=100.0,  # $100 per day
            period="day",
            alert_threshold=0.8,
        )

        # Stats
        self._total_cost = 0.0

    def set_limit(
        self,
        tenant_id: str,
        limit: CostLimit,
    ) -> None:
        """Set cost limit for a tenant.

        Args:
            tenant_id: Tenant ID
            limit: Cost limit
        """
        with self._lock:
            self._limits[tenant_id] = limit

    def get_limit(self, tenant_id: str) -> CostLimit:
        """Get cost limit for a tenant.

        Args:
            tenant_id: Tenant ID

        Returns:
            CostLimit
        """
        with self._lock:
            return self._limits.get(tenant_id, self._default_limit)

    def record_cost(
        self,
        tenant_id: str,
        provider: str,
        model: str,
        input_tokens: int,
        output_tokens: int,
        cost: float,
    ) -> None:
        """Record an API cost.

        Args:
            tenant_id: Tenant ID
            provider: Provider name
            model: Model name
            input_tokens: Number of input tokens
            output_tokens: Number of output tokens
            cost: Cost in USD
        """
        with self._lock:
            # Create entry
            entry = CostEntry(
                tenant_id=tenant_id,
                provider=provider,
                model=model,
                input_tokens=input_tokens,
                output_tokens=output_tokens,
                cost=cost,
            )

            self._entries.append(entry)
            self._total_cost += cost

            # Trim old entries
            if len(self._entries) > self._max_entries:
                self._entries = self._entries[-self._max_entries:]

    def check_limit(
        self,
        tenant_id: str,
        additional_cost: float = 0.0,
    ) -> bool:
        """Check if adding cost would exceed limit.

        Args:
            tenant_id: Tenant ID
            additional_cost: Additional cost to check

        Returns:
            True if within limits

        Raises:
            CostLimitExceeded: If limit exceeded
        """
        with self._lock:
            limit = self.get_limit(tenant_id)
            current_cost = self._get_current_cost(tenant_id, limit.period)

            if current_cost + additional_cost > limit.limit:
                raise CostLimitExceeded(
                    f"Cost limit exceeded: ${current_cost:.2f} / ${limit.limit:.2f} per {limit.period}",
                    tenant_id=tenant_id,
                    current_cost=current_cost,
                    limit=limit.limit,
                    period=limit.period,
                )

            return True

    def _get_current_cost(
        self,
        tenant_id: str,
        period: str,
    ) -> float:
        """Get current cost for a period.

        Args:
            tenant_id: Tenant ID
            period: Period (hour, day, month)

        Returns:
            Current cost
        """
        now = datetime.utcnow()

        if period == "hour":
            cutoff = now - timedelta(hours=1)
        elif period == "day":
            cutoff = now - timedelta(days=1)
        elif period == "month":
            cutoff = now - timedelta(days=30)
        else:
            cutoff = now - timedelta(days=1)

        total = 0.0
        for entry in self._entries:
            if entry.tenant_id == tenant_id and entry.timestamp >= cutoff:
                total += entry.cost

        return total

    def get_usage(self, tenant_id: str) -> Dict[str, float]:
        """Get current usage for a tenant.

        Args:
            tenant_id: Tenant ID

        Returns:
            Dictionary with usage statistics
        """
        with self._lock:
            limit = self.get_limit(tenant_id)

            # Calculate costs for each period
            now = datetime.utcnow()

            hourly = self._calculate_period_cost(tenant_id, now - timedelta(hours=1))
            daily = self._calculate_period_cost(tenant_id, now - timedelta(days=1))
            monthly = self._calculate_period_cost(tenant_id, now - timedelta(days=30))

            return {
                "hourly": hourly,
                "daily": daily,
                "monthly": monthly,
                "limit": limit.limit,
                "period": limit.period,
                "alert_threshold": limit.alert_threshold,
                "alert_at": limit.limit * limit.alert_threshold,
            }

    def _calculate_period_cost(
        self,
        tenant_id: str,
        since: datetime,
    ) -> float:
        """Calculate cost since a given time.

        Args:
            tenant_id: Tenant ID
            since: Start time

        Returns:
            Total cost
        """
        total = 0.0
        for entry in self._entries:
            if entry.tenant_id == tenant_id and entry.timestamp >= since:
                total += entry.cost
        return total

    def get_costs_by_provider(
        self,
        tenant_id: str,
    ) -> Dict[str, float]:
        """Get costs breakdown by provider.

        Args:
            tenant_id: Tenant ID

        Returns:
            Dictionary of provider -> cost
        """
        with self._lock:
            costs: Dict[str, float] = {}

            for entry in self._entries:
                if entry.tenant_id == tenant_id:
                    key = f"{entry.provider}:{entry.model}"
                    costs[key] = costs.get(key, 0.0) + entry.cost

            return costs

    def get_stats(self) -> Dict[str, any]:
        """Get cost tracker statistics.

        Returns:
            Dictionary with stats
        """
        with self._lock:
            return {
                "total_cost": self._total_cost,
                "total_entries": len(self._entries),
                "tracked_tenants": len(self._limits),
            }

    def reset(self, tenant_id: str) -> None:
        """Reset cost tracking for a tenant.

        Args:
            tenant_id: Tenant ID
        """
        with self._lock:
            self._entries = [
                e for e in self._entries
                if e.tenant_id != tenant_id
            ]

    def get_cost_history(
        self,
        tenant_id: str,
        limit: int = 100,
    ) -> List[CostEntry]:
        """Get cost history for a tenant.

        Args:
            tenant_id: Tenant ID
            limit: Maximum entries to return

        Returns:
            List of CostEntry
        """
        with self._lock:
            entries = [
                e for e in reversed(self._entries)
                if e.tenant_id == tenant_id
            ]
            return entries[:limit]


# Global cost tracker
_cost_tracker: Optional[CostTracker] = None


class TokenBudget:
    """Token budget tracking for a tenant."""

    def __init__(self, tokens_per_minute: int = 100000, tokens_per_hour: int = 1000000):
        self.tokens_per_minute = tokens_per_minute
        self.tokens_per_hour = tokens_per_hour
        self.minute_used = 0
        self.hour_used = 0
        self.minute_reset = time.time() + 60
        self.hour_reset = time.time() + 3600

    def check_budget(self, requested_tokens: int) -> bool:
        """Check if requested tokens are within budget."""
        now = time.time()

        # Reset minute counter at boundary
        if now >= self.minute_reset:
            self.minute_used = 0
            self.minute_reset = now + 60

        # Reset hour counter at boundary
        if now >= self.hour_reset:
            self.hour_used = 0
            self.hour_reset = now + 3600

        # Check if adding these tokens would exceed budget
        if self.minute_used + requested_tokens > self.tokens_per_minute:
            return False
        if self.hour_used + requested_tokens > self.tokens_per_hour:
            return False

        return True

    def consume_tokens(self, tokens: int) -> None:
        """Consume tokens from the budget."""
        now = time.time()

        # Reset minute counter at boundary if needed
        if now >= self.minute_reset:
            self.minute_used = 0
            self.minute_reset = now + 60

        # Reset hour counter at boundary if needed
        if now >= self.hour_reset:
            self.hour_used = 0
            self.hour_reset = now + 3600

        self.minute_used += tokens
        self.hour_used += tokens


class TokenBudgetManager:
    """Manages token budgets for all tenants."""
    
    def __init__(self):
        self._budgets: Dict[str, TokenBudget] = {}
        self._lock = threading.Lock()
    
    def get_budget(self, tenant_id: str) -> TokenBudget:
        """Get or create token budget for tenant."""
        with self._lock:
            if tenant_id not in self._budgets:
                self._budgets[tenant_id] = TokenBudget()
            return self._budgets[tenant_id]
    
    def check_token_budget(self, tenant_id: str, requested_tokens: int) -> bool:
        """Check if tenant has sufficient token budget."""
        budget = self.get_budget(tenant_id)
        return budget.check_budget(requested_tokens)
    
    def consume_tokens(self, tenant_id: str, tokens: int) -> None:
        """Consume tokens from tenant's budget."""
        budget = self.get_budget(tenant_id)
        budget.consume_tokens(tokens)


# Global token budget manager
_token_budget_manager: Optional[TokenBudgetManager] = None


def get_token_budget_manager() -> TokenBudgetManager:
    """Get the global token budget manager.
    
    Returns:
        TokenBudgetManager instance
    """
    global _token_budget_manager
    if _token_budget_manager is None:
        _token_budget_manager = TokenBudgetManager()
    return _token_budget_manager


def get_cost_tracker() -> CostTracker:
    """Get the global cost tracker.

    Returns:
        CostTracker instance
    """
    global _cost_tracker
    if _cost_tracker is None:
        _cost_tracker = CostTracker()

    return _cost_tracker
