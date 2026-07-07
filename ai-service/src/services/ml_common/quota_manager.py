"""Per-tenant resource quotas for ML services.

Enforces limits on ML resource usage per tenant to prevent
resource starvation and ensure fair allocation.
"""

import logging
import threading
import time
from dataclasses import dataclass
from datetime import datetime, timedelta
from enum import StrEnum
from typing import Dict, List, Optional

from ...config import settings

logger = logging.getLogger(__name__)


class QuotaType(StrEnum):
    """Types of resource quotas."""
    PREDICTIONS_PER_MINUTE = "predictions_per_minute"
    PREDICTIONS_PER_DAY = "predictions_per_day"
    TRAINING_PER_DAY = "training_per_day"
    STORAGE_MB = "storage_mb"
    CONCURRENT_REQUESTS = "concurrent_requests"


@dataclass
class QuotaLimit:
    """Defines a quota limit."""
    quota_type: QuotaType
    limit: int
    window_seconds: int = 60

    def __post_init__(self):
        if self.window_seconds == 60 and self.quota_type == QuotaType.PREDICTIONS_PER_DAY:
            self.window_seconds = 86400
        elif self.window_seconds == 60 and self.quota_type == QuotaType.TRAINING_PER_DAY:
            self.window_seconds = 86400


@dataclass
class QuotaUsage:
    """Tracks current quota usage."""
    tenant_id: str
    quota_type: QuotaType
    used: int = 0
    window_start: float = 0.0
    last_reset: float = 0.0


class QuotaExceeded(Exception):
    """Raised when quota is exceeded."""

    def __init__(self, quota_type: QuotaType, tenant_id: str, limit: int, retry_after: Optional[int] = None):
        self.quota_type = quota_type
        self.tenant_id = tenant_id
        self.limit = limit
        self.retry_after = retry_after
        super().__init__(
            f"Quota exceeded for tenant {tenant_id}: {quota_type.value} "
            f"(limit: {limit}, retry after: {retry_after}s)"
        )


class PerTenantQuotaManager:
    """Manages per-tenant resource quotas for ML services.

    Tracks usage and enforces limits on:
    - Predictions per minute/day
    - Training requests per day
    - Storage usage
    - Concurrent requests
    """

    DEFAULT_QUOTAS = {
        QuotaType.PREDICTIONS_PER_MINUTE: QuotaLimit(QuotaType.PREDICTIONS_PER_MINUTE, 1000, 60),
        QuotaType.PREDICTIONS_PER_DAY: QuotaLimit(QuotaType.PREDICTIONS_PER_DAY, 100000, 86400),
        QuotaType.TRAINING_PER_DAY: QuotaLimit(QuotaType.TRAINING_PER_DAY, 10, 86400),
        QuotaType.STORAGE_MB: QuotaLimit(QuotaType.STORAGE_MB, 1000, 86400),
        QuotaType.CONCURRENT_REQUESTS: QuotaLimit(QuotaType.CONCURRENT_REQUESTS, 10, 60),
    }

    def __init__(self):
        self._lock = threading.Lock()
        self._usage: Dict[str, Dict[QuotaType, QuotaUsage]] = {}
        self._limits: Dict[str, Dict[QuotaType, QuotaLimit]] = {}
        self._active_requests: Dict[str, int] = {}

    def _get_usage_key(self, tenant_id: str, quota_type: QuotaType) -> str:
        """Generate usage key."""
        return f"{tenant_id}:{quota_type.value}"

    def get_limits(self, tenant_id: str) -> Dict[QuotaType, QuotaLimit]:
        """Get quota limits for a tenant.

        Args:
            tenant_id: Tenant ID

        Returns:
            Dict of quota type to limits
        """
        with self._lock:
            if tenant_id not in self._limits:
                self._limits[tenant_id] = self.DEFAULT_QUOTAS.copy()
            return self._limits[tenant_id]

    def set_limit(self, tenant_id: str, quota_type: QuotaType, limit: int, window_seconds: Optional[int] = None) -> None:
        """Set a custom quota limit for a tenant.

        Args:
            tenant_id: Tenant ID
            quota_type: Type of quota
            limit: New limit
            window_seconds: Optional custom window
        """
        with self._lock:
            if tenant_id not in self._limits:
                self._limits[tenant_id] = self.DEFAULT_QUOTAS.copy()

            base_limit = self.DEFAULT_QUOTAS[quota_type]
            self._limits[tenant_id][quota_type] = QuotaLimit(
                quota_type,
                limit,
                window_seconds or base_limit.window_seconds
            )

    def check_and_increment(
        self,
        tenant_id: str,
        quota_type: QuotaType,
        amount: int = 1,
    ) -> None:
        """Check quota and increment usage if allowed.

        Args:
            tenant_id: Tenant ID
            quota_type: Type of quota
            amount: Amount to increment

        Raises:
            QuotaExceeded: If quota would be exceeded
        """
        limits = self.get_limits(tenant_id)
        if quota_type not in limits:
            return

        limit = limits[quota_type]
        usage = self._get_or_create_usage(tenant_id, quota_type, limit.window_seconds)

        now = time.time()

        with self._lock:
            if now - usage.window_start >= limit.window_seconds:
                usage.used = 0
                usage.window_start = now

            if usage.used + amount > limit.limit:
                retry_after = max(
                    int(limit.window_seconds - (now - usage.window_start)),
                    1
                )
                raise QuotaExceeded(
                    quota_type, tenant_id, limit.limit, retry_after
                )

            usage.used += amount

    def _get_or_create_usage(
        self,
        tenant_id: str,
        quota_type: QuotaType,
        window_seconds: int,
    ) -> QuotaUsage:
        """Get or create usage tracking for a quota.

        Args:
            tenant_id: Tenant ID
            quota_type: Type of quota
            window_seconds: Window duration

        Returns:
            QuotaUsage instance
        """
        key = self._get_usage_key(tenant_id, quota_type)
        now = time.time()

        if tenant_id not in self._usage:
            self._usage[tenant_id] = {}

        if quota_type not in self._usage[tenant_id]:
            self._usage[tenant_id][quota_type] = QuotaUsage(
                tenant_id=tenant_id,
                quota_type=quota_type,
                used=0,
                window_start=now,
                last_reset=now,
            )

        return self._usage[tenant_id][quota_type]

    def decrement(
        self,
        tenant_id: str,
        quota_type: QuotaType,
        amount: int = 1,
    ) -> None:
        """Decrement usage (e.g., when request completes).

        Args:
            tenant_id: Tenant ID
            quota_type: Type of quota
            amount: Amount to decrement
        """
        with self._lock:
            key = self._get_usage_key(tenant_id, quota_type)
            if tenant_id in self._usage and quota_type in self._usage[tenant_id]:
                usage = self._usage[tenant_id][quota_type]
                usage.used = max(0, usage.used - amount)

    def get_usage(
        self,
        tenant_id: str,
        quota_type: QuotaType,
    ) -> tuple[int, int]:
        """Get current usage and limit for a quota.

        Args:
            tenant_id: Tenant ID
            quota_type: Type of quota

        Returns:
            Tuple of (used, limit)
        """
        limits = self.get_limits(tenant_id)
        if quota_type not in limits:
            return 0, 0

        limit = limits[quota_type]
        usage = self._get_or_create_usage(tenant_id, quota_type, limit.window_seconds)

        with self._lock:
            return usage.used, limit.limit

    def reset_usage(self, tenant_id: str, quota_type: Optional[QuotaType] = None) -> None:
        """Reset usage for a tenant.

        Args:
            tenant_id: Tenant ID
            quota_type: Optional specific quota type to reset (all if None)
        """
        with self._lock:
            if tenant_id not in self._usage:
                return

            if quota_type:
                if quota_type in self._usage[tenant_id]:
                    self._usage[tenant_id][quota_type].used = 0
            else:
                for qt in self._usage[tenant_id]:
                    self._usage[tenant_id][qt].used = 0

    def acquire_concurrent_request(self, tenant_id: str) -> bool:
        """Acquire a concurrent request slot.

        Args:
            tenant_id: Tenant ID

        Returns:
            True if acquired, False if at limit
        """
        limits = self.get_limits(tenant_id)
        if QuotaType.CONCURRENT_REQUESTS not in limits:
            return True

        limit = limits[QuotaType.CONCURRENT_REQUESTS].limit

        with self._lock:
            current = self._active_requests.get(tenant_id, 0)
            if current >= limit:
                return False
            self._active_requests[tenant_id] = current + 1
            return True

    def release_concurrent_request(self, tenant_id: str) -> None:
        """Release a concurrent request slot.

        Args:
            tenant_id: Tenant ID
        """
        with self._lock:
            current = self._active_requests.get(tenant_id, 0)
            if current > 0:
                self._active_requests[tenant_id] = current - 1

    def get_active_requests(self, tenant_id: str) -> int:
        """Get number of active requests for a tenant.

        Args:
            tenant_id: Tenant ID

        Returns:
            Number of active requests
        """
        with self._lock:
            return self._active_requests.get(tenant_id, 0)

    def get_quota_status(self, tenant_id: str) -> Dict[str, Dict[str, int]]:
        """Get full quota status for a tenant.

        Args:
            tenant_id: Tenant ID

        Returns:
            Dict with quota status for each type
        """
        limits = self.get_limits(tenant_id)
        status = {}

        for quota_type, limit in limits.items():
            used, _ = self.get_usage(tenant_id, quota_type)
            status[quota_type.value] = {
                "used": used,
                "limit": limit.limit,
                "window_seconds": limit.window_seconds,
                "available": max(0, limit.limit - used),
            }

        status["concurrent_requests"] = {
            "active": self.get_active_requests(tenant_id),
            "limit": limits.get(QuotaType.CONCURRENT_REQUESTS, QuotaLimit(QuotaType.CONCURRENT_REQUESTS, 10)).limit,
        }

        return status


_quota_manager: Optional[PerTenantQuotaManager] = None


def get_quota_manager() -> PerTenantQuotaManager:
    """Get the global quota manager instance."""
    global _quota_manager
    if _quota_manager is None:
        _quota_manager = PerTenantQuotaManager()
    return _quota_manager


def check_ml_quota(
    tenant_id: str,
    quota_type: QuotaType,
    amount: int = 1,
) -> None:
    """Check and increment ML quota.

    Args:
        tenant_id: Tenant ID
        quota_type: Type of quota
        amount: Amount to increment

    Raises:
        QuotaExceeded: If quota exceeded
    """
    manager = get_quota_manager()
    manager.check_and_increment(tenant_id, quota_type, amount)


def record_prediction(tenant_id: str) -> None:
    """Record a prediction for quota tracking.

    Args:
        tenant_id: Tenant ID
    """
    manager = get_quota_manager()
    try:
        manager.check_and_increment(tenant_id, QuotaType.PREDICTIONS_PER_MINUTE)
        manager.check_and_increment(tenant_id, QuotaType.PREDICTIONS_PER_DAY)
    except QuotaExceeded:
        raise


def record_training_request(tenant_id: str) -> None:
    """Record a training request for quota tracking.

    Args:
        tenant_id: Tenant ID

    Raises:
        QuotaExceeded: If quota exceeded
    """
    manager = get_quota_manager()
    manager.check_and_increment(tenant_id, QuotaType.TRAINING_PER_DAY)


def acquire_request_slot(tenant_id: str) -> bool:
    """Acquire a slot for concurrent request.

    Args:
        tenant_id: Tenant ID

    Returns:
        True if acquired, False if at limit
    """
    manager = get_quota_manager()
    return manager.acquire_concurrent_request(tenant_id)


def release_request_slot(tenant_id: str) -> None:
    """Release a concurrent request slot.

    Args:
        tenant_id: Tenant ID
    """
    manager = get_quota_manager()
    manager.release_concurrent_request(tenant_id)
