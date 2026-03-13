"""Rate limiting for FlyMind AI Service.

This module provides per-tenant rate limiting using token bucket algorithm.
"""

import time
from dataclasses import dataclass, field
from datetime import datetime
from typing import Dict, Optional
import threading
import logging

logger = logging.getLogger(__name__)


class RateLimitExceeded(Exception):
    """Exception raised when rate limit is exceeded."""

    def __init__(
        self,
        message: str,
        tenant_id: str,
        limit: int,
        window_seconds: int,
        retry_after: Optional[int] = None,
    ):
        """Initialize the exception.

        Args:
            message: Error message
            tenant_id: Tenant ID
            limit: Rate limit
            window_seconds: Time window in seconds
            retry_after: Seconds until retry
        """
        super().__init__(message)
        self.tenant_id = tenant_id
        self.limit = limit
        self.window_seconds = window_seconds
        self.retry_after = retry_after


@dataclass
class RateLimitConfig:
    """Configuration for rate limiting."""
    requests_per_minute: int = 60
    requests_per_hour: int = 1000
    requests_per_day: int = 10000
    burst_size: int = 10
    enabled: bool = True


@dataclass
class RateLimitState:
    """State of rate limiting for a tenant."""
    tokens: float
    last_refill: float
    minute_count: int = 0
    hour_count: int = 0
    day_count: int = 0
    minute_reset: float = 0
    hour_reset: float = 0
    day_reset: float = 0


class RateLimiter:
    """Token bucket rate limiter with multiple time windows."""

    def __init__(self):
        """Initialize the rate limiter."""
        self._logger = logging.getLogger(__name__)
        self._lock = threading.Lock()

        # Tenant configurations
        self._configs: Dict[str, RateLimitConfig] = {}

        # Tenant states
        self._states: Dict[str, RateLimitState] = {}

        # Default configuration
        self._default_config = RateLimitConfig()

        # Stats
        self._total_requests = 0
        self._total_rejected = 0

    def set_config(
        self,
        tenant_id: str,
        config: RateLimitConfig,
    ) -> None:
        """Set rate limit configuration for a tenant.

        Args:
            tenant_id: Tenant ID
            config: Rate limit configuration
        """
        with self._lock:
            self._configs[tenant_id] = config

    def get_config(self, tenant_id: str) -> RateLimitConfig:
        """Get rate limit configuration for a tenant.

        Args:
            tenant_id: Tenant ID

        Returns:
            RateLimitConfig
        """
        with self._lock:
            return self._configs.get(tenant_id, self._default_config)

    def check_limit(
        self,
        tenant_id: str,
        cost: int = 1,
    ) -> bool:
        """Check if request is within rate limits.

        Args:
            tenant_id: Tenant ID
            cost: Request cost (for burst calculations)

        Returns:
            True if within limits

        Raises:
            RateLimitExceeded: If limit exceeded
        """
        with self._lock:
            self._total_requests += 1

            config = self.get_config(tenant_id)

            if not config.enabled:
                return True

            # Get or create state
            if tenant_id not in self._states:
                self._states[tenant_id] = RateLimitState(
                    tokens=float(config.burst_size),
                    last_refill=time.time(),
                    minute_reset=time.time() + 60,
                    hour_reset=time.time() + 3600,
                    day_reset=time.time() + 86400,
                )

            state = self._states[tenant_id]
            now = time.time()

            # Refill tokens
            self._refill_tokens(state, config, now)

            # Check minute limit
            if now >= state.minute_reset:
                state.minute_count = 0
                state.minute_reset = now + 60

            if state.minute_count >= config.requests_per_minute:
                self._total_rejected += 1
                retry_after = int(state.minute_reset - now) + 1
                raise RateLimitExceeded(
                    f"Rate limit exceeded: {config.requests_per_minute} requests per minute",
                    tenant_id=tenant_id,
                    limit=config.requests_per_minute,
                    window_seconds=60,
                    retry_after=retry_after,
                )

            # Check hour limit
            if now >= state.hour_reset:
                state.hour_count = 0
                state.hour_reset = now + 3600

            if state.hour_count >= config.requests_per_hour:
                self._total_rejected += 1
                retry_after = int(state.hour_reset - now) + 1
                raise RateLimitExceeded(
                    f"Rate limit exceeded: {config.requests_per_hour} requests per hour",
                    tenant_id=tenant_id,
                    limit=config.requests_per_hour,
                    window_seconds=3600,
                    retry_after=retry_after,
                )

            # Check day limit
            if now >= state.day_reset:
                state.day_count = 0
                state.day_reset = now + 86400

            if state.day_count >= config.requests_per_day:
                self._total_rejected += 1
                retry_after = int(state.day_reset - now) + 1
                raise RateLimitExceeded(
                    f"Rate limit exceeded: {config.requests_per_day} requests per day",
                    tenant_id=tenant_id,
                    limit=config.requests_per_day,
                    window_seconds=86400,
                    retry_after=retry_after,
                )

            # Check token bucket
            if state.tokens < cost:
                self._total_rejected += 1
                raise RateLimitExceeded(
                    f"Rate limit exceeded: burst limit",
                    tenant_id=tenant_id,
                    limit=config.burst_size,
                    window_seconds=60,
                    retry_after=1,
                )

            # Consume tokens and counts
            state.tokens -= cost
            state.minute_count += 1
            state.hour_count += 1
            state.day_count += 1

            return True

    def _refill_tokens(
        self,
        state: RateLimitState,
        config: RateLimitConfig,
        now: float,
    ) -> None:
        """Refill tokens based on elapsed time.

        Args:
            state: Rate limit state
            config: Rate limit config
            now: Current time
        """
        # Calculate time since last refill
        elapsed = now - state.last_refill

        if elapsed > 0:
            # Refill tokens (tokens per second * elapsed)
            refill_rate = config.requests_per_minute / 60.0
            state.tokens = min(
                config.burst_size,
                state.tokens + (refill_rate * elapsed)
            )
            state.last_refill = now

    def get_usage(self, tenant_id: str) -> Dict[str, int]:
        """Get current usage for a tenant.

        Args:
            tenant_id: Tenant ID

        Returns:
            Dictionary with usage statistics
        """
        with self._lock:
            config = self.get_config(tenant_id)
            state = self._states.get(tenant_id)

            if not state:
                return {
                    "minute": 0,
                    "hour": 0,
                    "day": 0,
                    "limit_minute": config.requests_per_minute,
                    "limit_hour": config.requests_per_hour,
                    "limit_day": config.requests_per_day,
                }

            return {
                "minute": state.minute_count,
                "hour": state.hour_count,
                "day": state.day_count,
                "limit_minute": config.requests_per_minute,
                "limit_hour": config.requests_per_hour,
                "limit_day": config.requests_per_day,
                "tokens": int(state.tokens),
            }

    def get_stats(self) -> Dict[str, int]:
        """Get rate limiter statistics.

        Returns:
            Dictionary with stats
        """
        with self._lock:
            return {
                "total_requests": self._total_requests,
                "total_rejected": self._total_rejected,
                "total_tenants": len(self._states),
            }

    def reset(self, tenant_id: str) -> None:
        """Reset rate limits for a tenant.

        Args:
            tenant_id: Tenant ID
        """
        with self._lock:
            if tenant_id in self._states:
                del self._states[tenant_id]

    def reset_all(self) -> None:
        """Reset all rate limits."""
        with self._lock:
            self._states.clear()
            self._total_requests = 0
            self._total_rejected = 0


# Global rate limiter
_rate_limiter: Optional[RateLimiter] = None


def get_rate_limiter() -> RateLimiter:
    """Get the global rate limiter.

    Returns:
        RateLimiter instance
    """
    global _rate_limiter
    if _rate_limiter is None:
        _rate_limiter = RateLimiter()

    return _rate_limiter
