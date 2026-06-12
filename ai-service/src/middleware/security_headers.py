"""Security headers middleware for FlyMind AI Service.

Adds security-related HTTP headers to all responses.
"""

import logging
from typing import Callable

from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import Response

logger = logging.getLogger(__name__)


class SecurityHeadersMiddleware(BaseHTTPMiddleware):
    """Middleware to add security headers to all responses.

    Headers added:
    - X-Content-Type-Options: nosniff
    - X-Frame-Options: DENY
    - X-XSS-Protection: 1; mode=block
    - Strict-Transport-Security: max-age=31536000; includeSubDomains
    - Content-Security-Policy: default-src 'self'
    - Referrer-Policy: strict-origin-when-cross-origin
    - Permissions-Policy: accelerometer=(), camera=(), microphone=(), geolocation=()
    """

    # Content-Security-Policy for AI service
    CSP_DEFAULT = "default-src 'self'; " \
                  "script-src 'self' 'unsafe-inline'; " \
                  "style-src 'self' 'unsafe-inline'; " \
                  "img-src 'self' data: https:; " \
                  "connect-src 'self' https://api.openai.com https://api.anthropic.com https://api.fireworks.ai https://api.groq.com https://api.deepinfra.com https://api.together.xyz https://openrouter.ai; " \
                  "frame-ancestors 'none'; " \
                  "form-action 'self'; " \
                  "base-uri 'self';"

    # Permissions-Policy to restrict sensitive features
    PERMISSIONS_POLICY = "accelerometer=(), camera=(), microphone=(), geolocation=(), payment=()"

    async def dispatch(self, request: Request, call_next: Callable) -> Response:
        response = await call_next(request)

        # Skip security headers for streaming responses
        content_type = response.headers.get("content-type", "")
        if "text/event-stream" in content_type:
            return response

        # Add security headers
        response.headers["X-Content-Type-Options"] = "nosniff"
        response.headers["X-Frame-Options"] = "DENY"
        response.headers["X-XSS-Protection"] = "1; mode=block"
        response.headers["Strict-Transport-Security"] = "max-age=31536000; includeSubDomains; preload"
        response.headers["Content-Security-Policy"] = self.CSP_DEFAULT
        response.headers["Referrer-Policy"] = "strict-origin-when-cross-origin"
        response.headers["Permissions-Policy"] = self.PERMISSIONS_POLICY

        # Remove server identification
        response.headers.pop("server", None)

        return response


def get_security_headers_middleware() -> type[SecurityHeadersMiddleware]:
    """Get the security headers middleware class.

    Returns:
        SecurityHeadersMiddleware class
    """
    return SecurityHeadersMiddleware
