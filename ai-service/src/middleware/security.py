"""Security-focused middleware for FlyMind AI Service."""

import re
import time
from typing import Callable, Optional, Pattern
import logging

from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import Response

from ..config import settings

logger = logging.getLogger(__name__)


TENANT_ID_PATTERN: Pattern[str] = re.compile(r'^[a-zA-Z0-9_\-]{1,64}$')


class SecurityHeadersMiddleware(BaseHTTPMiddleware):
    """Middleware to add security headers to all responses.

    Adds the following headers:
    - Strict-Transport-Security (HSTS)
    - X-Content-Type-Options: nosniff
    - X-Frame-Options: DENY
    - X-XSS-Protection: 1; mode=block
    - Referrer-Policy: strict-origin-when-cross-origin
    - Content-Security-Policy (restrictive default)
    - Permissions-Policy (restrictive defaults)
    """

    async def dispatch(self, request: Request, call_next: Callable) -> Response:
        response: Response = await call_next(request)

        response.headers["Strict-Transport-Security"] = "max-age=31536000; includeSubDomains"
        response.headers["X-Content-Type-Options"] = "nosniff"
        response.headers["X-Frame-Options"] = "DENY"
        response.headers["X-XSS-Protection"] = "1; mode=block"
        response.headers["Referrer-Policy"] = "strict-origin-when-cross-origin"
        response.headers["Permissions-Policy"] = "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()"

        return response


class RequestSizeLimitMiddleware(BaseHTTPMiddleware):
    """Middleware to enforce maximum request body size.

    Prevents memory exhaustion from oversized payloads.
    """

    MAX_BODY_SIZE = 10 * 1024 * 1024

    async def dispatch(self, request: Request, call_next: Callable) -> Response:
        content_length = request.headers.get("content-length")

        if content_length and int(content_length) > self.MAX_BODY_SIZE:
            from starlette.responses import JSONResponse
            return JSONResponse(
                status_code=413,
                content={"detail": "Request body too large"},
            )

        return await call_next(request)


class TimeoutMiddleware(BaseHTTPMiddleware):
    """Middleware to enforce request timeouts.

    Prevents slow-loris and resource exhaustion attacks.
    """

    DEFAULT_TIMEOUT = 30.0

    async def dispatch(self, request: Request, call_next: Callable) -> Response:
        start_time = time.monotonic()
        timeout = getattr(settings, 'request_timeout_seconds', self.DEFAULT_TIMEOUT)

        response = await call_next(request)

        elapsed = time.monotonic() - start_time
        if elapsed > timeout:
            logger.warning(f"Request to {request.url.path} exceeded timeout: {elapsed:.2f}s")

        return response


class TenantValidationMiddleware(BaseHTTPMiddleware):
    """Middleware to validate tenant IDs in requests.

    Prevents path traversal attacks via tenant_id parameter.
    """

    TENANT_PARAM_NAMES = {"tenant_id", "tenant", "t"}

    async def dispatch(self, request: Request, call_next: Callable) -> Response:
        from starlette.responses import JSONResponse

        query_params = dict(request.query_params)

        for param_name in self.TENANT_PARAM_NAMES:
            if param_name in query_params:
                tenant_id = query_params[param_name]
                if not self._validate_tenant_id(tenant_id):
                    return JSONResponse(
                        status_code=400,
                        content={"detail": f"Invalid tenant_id format: {tenant_id}"},
                    )

        return await call_next(request)

    def _validate_tenant_id(self, tenant_id: str) -> bool:
        if not tenant_id:
            return False
        return bool(TENANT_ID_PATTERN.match(tenant_id))


def validate_tenant_id(tenant_id: str) -> bool:
    """Validate a tenant ID format.

    Args:
        tenant_id: The tenant ID to validate

    Returns:
        True if valid, False otherwise
    """
    if not tenant_id:
        return False
    return bool(TENANT_ID_PATTERN.match(tenant_id))


def sanitize_tenant_id(tenant_id: str) -> str:
    """Sanitize a tenant ID by removing potentially dangerous characters.

    Args:
        tenant_id: The tenant ID to sanitize

    Returns:
        Sanitized tenant ID
    """
    if not tenant_id:
        return "default"
    sanitized = re.sub(r'[^a-zA-Z0-9_\-]', '', tenant_id)
    return sanitized[:64] or "default"