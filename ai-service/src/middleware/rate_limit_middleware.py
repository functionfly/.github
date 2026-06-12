"""Rate limiting middleware for FastAPI."""

import hashlib
import logging
from typing import Callable

from fastapi import Request, HTTPException
from starlette.middleware.base import BaseHTTPMiddleware

from .rate_limiter import get_rate_limiter, RateLimitExceeded

logger = logging.getLogger(__name__)

EXCLUDED_PATHS = {
    "/",
    "/health",
    "/metrics",
    "/docs",
    "/redoc",
    "/openapi.json",
}


class RateLimitMiddleware(BaseHTTPMiddleware):
    """FastAPI middleware for rate limiting.

    Applies rate limiting based on tenant ID extracted from API key.
    Excludes certain paths like health checks and metrics.
    """

    def __init__(self, app, enabled: bool = True):
        super().__init__(app)
        self._enabled = enabled
        self._rate_limiter = None

    async def dispatch(self, request: Request, call_next: Callable) -> Callable:
        if request.url.path in EXCLUDED_PATHS:
            return await call_next(request)

        if not self._enabled:
            return await call_next(request)

        if self._rate_limiter is None:
            self._rate_limiter = get_rate_limiter()

        api_key = request.headers.get("X-API-Key")
        if not api_key:
            return await call_next(request)

        tenant_id = hashlib.sha256(api_key.encode()).hexdigest()[:16]

        try:
            self._rate_limiter.check_limit(tenant_id, cost=1)
            response = await call_next(request)
            usage = self._rate_limiter.get_usage(tenant_id)
            limit = usage.get("limit_minute", 60)
            used = usage.get("minute", 0)
            remaining = max(0, limit - used)
            response.headers["X-RateLimit-Limit-Minute"] = str(limit)
            response.headers["X-RateLimit-Remaining-Minute"] = str(remaining)
            return response
        except RateLimitExceeded as e:
            logger.warning(f"Rate limit exceeded for tenant {tenant_id}: {e}")
            raise HTTPException(
                status_code=429,
                detail=str(e),
                headers={
                    "Retry-After": str(e.retry_after or 60),
                    "X-RateLimit-Limit": str(e.limit),
                    "X-RateLimit-Window": str(e.window_seconds),
                },
            )


def setup_rate_limit_middleware(app, enabled: bool = True) -> None:
    """Add rate limit middleware to the FastAPI app."""
    app.add_middleware(RateLimitMiddleware, enabled=enabled)
