"""API key validation for FlyMind AI Service.

This module provides API key authentication and validation using the Go orchestrator's
PostgreSQL-backed key storage. Keys persist across service restarts and revocations
are immediately effective.
"""

import hashlib
import logging
import time
from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from typing import Dict, List, Optional
import threading

import httpx

from fastapi import Header, HTTPException, Depends, status
from fastapi.security import APIKeyHeader

logger = logging.getLogger(__name__)

# FastAPI security scheme
api_key_header = APIKeyHeader(name="X-API-Key", auto_error=False)


class AuthRateLimiter:
    """Rate limiter for failed authentication attempts.

    Implements progressive delay to prevent brute force attacks on API keys.
    """

    MAX_LOCKOUT_MINUTES = 15
    PROGRESSIVE_DELAYS = [
        (3, 1),    # 3 failures -> 1 second delay
        (5, 5),    # 5 failures -> 5 second delay
        (10, 30),   # 10 failures -> 30 second delay
        (20, 60),   # 20 failures -> 60 second delay
    ]

    def __init__(self):
        self._lock = threading.Lock()
        self._attempts: Dict[str, Dict] = {}  # key_hash -> {"count": int, "last_attempt": float, "locked_until": float}

    def _get_key_hash(self, api_key: str) -> str:
        """Get a hash of the API key for tracking."""
        return hashlib.sha256(api_key.encode()).hexdigest()[:16]

    def _get_delay(self, failures: int) -> int:
        """Get delay in seconds based on failure count."""
        for threshold, delay in self.PROGRESSIVE_DELAYS:
            if failures < threshold:
                return 0
        return self.PROGRESSIVE_DELAYS[-1][1]

    def record_failure(self, api_key: str) -> int:
        """Record a failed auth attempt and return delay in seconds before retry.

        Args:
            api_key: The API key that failed

        Returns:
            Delay in seconds before retry (0 if not locked out)
        """
        key_hash = self._get_key_hash(api_key)
        now = time.time()

        with self._lock:
            if key_hash not in self._attempts:
                self._attempts[key_hash] = {"count": 0, "last_attempt": now, "locked_until": 0}

            entry = self._attempts[key_hash]

            # Reset if last attempt was > 15 minutes ago
            if now - entry["last_attempt"] > 900:
                entry["count"] = 0

            entry["count"] += 1
            entry["last_attempt"] = now

            delay = self._get_delay(entry["count"])

            if delay > 0:
                entry["locked_until"] = now + delay
                logger.warning(f"Auth rate limit: key hash {key_hash[:8]} locked for {delay}s after {entry['count']} failures")

            return delay

    def record_success(self, api_key: str) -> None:
        """Clear failure count on successful auth.

        Args:
            api_key: The API key that succeeded
        """
        key_hash = self._get_key_hash(api_key)

        with self._lock:
            if key_hash in self._attempts:
                del self._attempts[key_hash]

    def check_lockout(self, api_key: str) -> tuple[bool, int]:
        """Check if key is locked out.

        Args:
            api_key: The API key to check

        Returns:
            Tuple of (is_locked, retry_after_seconds)
        """
        key_hash = self._get_key_hash(api_key)
        now = time.time()

        with self._lock:
            if key_hash not in self._attempts:
                return False, 0

            entry = self._attempts[key_hash]

            if now < entry.get("locked_until", 0):
                return True, int(entry["locked_until"] - now) + 1

            return False, 0

    def get_stats(self) -> Dict[str, int]:
        """Get rate limiter statistics."""
        with self._lock:
            return {
                "tracked_keys": len(self._attempts),
            }


# Global auth rate limiter instance
_auth_rate_limiter: Optional[AuthRateLimiter] = None


def get_auth_rate_limiter() -> AuthRateLimiter:
    """Get the global auth rate limiter instance."""
    global _auth_rate_limiter
    if _auth_rate_limiter is None:
        _auth_rate_limiter = AuthRateLimiter()
    return _auth_rate_limiter


class KeyStatus(str, Enum):
    """API key status."""
    ACTIVE = "active"
    REVOKED = "revoked"
    EXPIRED = "expired"
    SUSPENDED = "suspended"


class KeyScope(str, Enum):
    """API key scopes."""
    READ = "read"
    WRITE = "write"
    ADMIN = "admin"
    FULL = "full"
    # Embedding-specific scopes
    EMBED_READ = "embed:read"
    EMBED_WRITE = "embed:write"
    EMBED_ADMIN = "embed:admin"
    RAG_READ = "rag:read"
    # Chat/composer-specific scopes
    CHAT_WRITE = "chat:write"
    CHAT_READ = "chat:read"


@dataclass
class APIKeyInfo:
    """Information about an API key."""
    key_id: str
    tenant_id: str
    name: str
    scopes: List[KeyScope]
    status: KeyStatus
    created_at: datetime
    expires_at: Optional[datetime] = None
    last_used_at: Optional[datetime] = None
    request_count: int = 0
    rate_limit: int = 60

    def is_valid(self) -> bool:
        """Check if the key is valid."""
        if self.status != KeyStatus.ACTIVE:
            return False

        if self.expires_at and datetime.utcnow() > self.expires_at:
            return False

        return True

    def has_scope(self, scope: KeyScope) -> bool:
        """Check if the key has a specific scope.

        FULL scope grants access to everything.
        """
        if KeyScope.FULL in self.scopes:
            return True
        return scope in self.scopes


class APIKeyValidator:
    """Validates API keys using Go orchestrator's PostgreSQL storage.

    This validator queries the orchestrator's /auth/validate-key endpoint
    to validate keys against the database-backed storage.
    """

    MAX_FAILED_ATTEMPTS = 5
    FAILED_ATTEMPT_WINDOW_SECONDS = 300  # 5 minutes
    FAILED_ATTEMPT_COOLDOWN_SECONDS = 900  # 15 minutes lockout after max attempts

    def __init__(
        self,
        orchestrator_url: str = "http://localhost:8080",
        orchestrator_api_key: Optional[str] = None,
        cache_ttl_seconds: int = 60,
        reject_in_degraded_mode: bool = True,
    ):
        """Initialize the validator.

        Args:
            orchestrator_url: URL of the Go orchestrator API
            orchestrator_api_key: Optional API key for orchestrator authentication
            cache_ttl_seconds: How long to cache validation results
            reject_in_degraded_mode: If True, reject all requests when orchestrator is unreachable
        """
        self._orchestrator_url = orchestrator_url.rstrip("/")
        self._orchestrator_api_key = orchestrator_api_key
        self._cache_ttl = cache_ttl_seconds
        self._reject_in_degraded_mode = reject_in_degraded_mode

        self._lock = threading.Lock()
        self._cache: Dict[str, tuple[Optional[APIKeyInfo], float]] = {}  # key_hash -> (info, expiry)
        self._failed_attempts: Dict[str, tuple[int, float]] = {}  # key_hash -> (attempts, first_failure_time)
        self._is_degraded = False

    def set_degraded(self, degraded: bool) -> None:
        """Set the degraded mode flag.

        Args:
            degraded: True if running in degraded mode (orchestrator unreachable)
        """
        self._is_degraded = degraded
        if degraded:
            logger.warning("API key validator is now in DEGRADED mode")
        else:
            logger.info("API key validator recovered from degraded mode")

    def is_degraded(self) -> bool:
        """Check if validator is in degraded mode."""
        return self._is_degraded

    def _get_cache_key(self, key: str) -> str:
        """Generate a cache key from the API key."""
        return hashlib.sha256(key.encode()).hexdigest()

    async def _validate_via_orchestrator(self, key: str) -> Optional[APIKeyInfo]:
        """Validate a key directly via orchestrator API (no cache)."""
        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                headers = {}
                if self._orchestrator_api_key:
                    headers["Authorization"] = f"Bearer {self._orchestrator_api_key}"

                response = await client.post(
                    f"{self._orchestrator_url}/api/v1/auth/validate-key",
                    json={"api_key": key},
                    headers=headers,
                )

                if response.status_code == 200:
                    data = response.json()
                    if not data.get("data", {}).get("valid", False):
                        return None

                    return self._parse_key_response(data["data"])
                else:
                    logger.warning(f"Unexpected response from orchestrator: {response.status_code}")

        except Exception as e:
            logger.error(f"Failed to validate key via orchestrator: {e}")

        return None

    def _parse_key_response(self, data: dict) -> Optional[APIKeyInfo]:
        """Parse the orchestrator response into APIKeyInfo."""
        try:
            # Parse expiration
            expires_at = None
            if data.get("expires_at"):
                expires_at = datetime.fromisoformat(data["expires_at"].replace("Z", "+00:00"))

            # Parse last used
            last_used_at = None
            if data.get("last_used_at"):
                last_used_at = datetime.fromisoformat(data["last_used_at"].replace("Z", "+00:00"))

            # Parse scopes
            scopes = []
            for scope_str in data.get("scopes", []):
                try:
                    scopes.append(KeyScope(scope_str))
                except ValueError:
                    logger.debug(f"Unknown scope: {scope_str}")

            # Determine status
            if data.get("is_revoked"):
                status_val = KeyStatus.REVOKED
            elif not data.get("is_active"):
                status_val = KeyStatus.SUSPENDED
            elif expires_at and datetime.utcnow() > expires_at:
                status_val = KeyStatus.EXPIRED
            else:
                status_val = KeyStatus.ACTIVE

            return APIKeyInfo(
                key_id=data.get("key_id", ""),
                tenant_id=data.get("tenant_id", ""),
                name=data.get("name", ""),
                scopes=scopes,
                status=status_val,
                created_at=datetime.utcnow(),
                expires_at=expires_at,
                last_used_at=last_used_at,
                request_count=0,
                rate_limit=data.get("rate_limit_rpm", 60),
            )

        except Exception as e:
            logger.error(f"Failed to parse key response: {e}")
            return None

    async def initialize(self) -> bool:
        """Initialize the validator by checking orchestrator connectivity.

        Returns:
            True if orchestrator is reachable
        """
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                response = await client.get(f"{self._orchestrator_url}/health")
                return response.status_code == 200
        except Exception as e:
            logger.error(f"Orchestrator health check failed: {e}")
            return False

    def validate_key_sync(self, key: str) -> Optional[APIKeyInfo]:
        """Validate an API key synchronously (for FastAPI dependencies).

        Uses asyncio.run_until_complete to properly handle async validation
        without nesting event loops. Prefer validate_key_async in async contexts.
        """
        import asyncio

        # In degraded mode with reject enabled, refuse to validate
        if self._is_degraded and self._reject_in_degraded_mode:
            logger.warning("Rejecting auth request in degraded mode")
            return None

        cache_key = self._get_cache_key(key)

        # Check cache first
        with self._lock:
            if cache_key in self._cache:
                cached_info, expiry = self._cache[cache_key]
                if time.time() < expiry:
                    return cached_info

        # Validate via orchestrator using existing event loop
        try:
            loop = asyncio.get_event_loop()
            if loop.is_running():
                # If loop is running, we need to use a different approach
                # Fall back to creating a new loop in a separate thread
                import concurrent.futures
                with concurrent.futures.ThreadPoolExecutor() as executor:
                    future = executor.submit(
                        asyncio.run,
                        self._validate_via_orchestrator(key)
                    )
                    info = future.result()
            else:
                info = loop.run_until_complete(self._validate_via_orchestrator(key))
        except RuntimeError:
            # No event loop, create one
            info = asyncio.run(self._validate_via_orchestrator(key))

        # Cache the result
        with self._lock:
            self._cache[cache_key] = (info, time.time() + self._cache_ttl)

        return info

    async def validate_key_async(self, key: str) -> Optional[APIKeyInfo]:
        """Validate an API key asynchronously.

        Args:
            key: The API key to validate

        Returns:
            APIKeyInfo if valid, None otherwise

        Raises:
            HTTPException: 429 if too many failed attempts
        """
        cache_key = self._get_cache_key(key)

        # Check rate limiting based on failed attempts
        with self._lock:
            if cache_key in self._failed_attempts:
                attempts, first_failure = self._failed_attempts[cache_key]
                elapsed = time.time() - first_failure

                if attempts >= self.MAX_FAILED_ATTEMPTS:
                    if elapsed < self.FAILED_ATTEMPT_COOLDOWN_SECONDS:
                        logger.warning(f"Rate limited: too many failed attempts for key hash {cache_key[:8]}")
                        raise HTTPException(
                            status_code=status.HTTP_429_TOO_MANY_REQUESTS,
                            detail="Too many failed attempts. Please try again later.",
                        )
                    else:
                        del self._failed_attempts[cache_key]

                elif elapsed < self.FAILED_ATTEMPT_WINDOW_SECONDS:
                    logger.warning(f"High failure rate for key hash {cache_key[:8]}: {attempts} failures")
        # Check cache first
        with self._lock:
            if cache_key in self._cache:
                cached_info, expiry = self._cache[cache_key]
                if time.time() < expiry:
                    return cached_info

        # Validate via orchestrator
        info = await self._validate_via_orchestrator(key)

        # Track failed attempts
        if info is None:
            with self._lock:
                if cache_key in self._failed_attempts:
                    attempts, first_failure = self._failed_attempts[cache_key]
                    if time.time() - first_failure > self.FAILED_ATTEMPT_WINDOW_SECONDS:
                        self._failed_attempts[cache_key] = (1, time.time())
                    else:
                        self._failed_attempts[cache_key] = (attempts + 1, first_failure)
                else:
                    self._failed_attempts[cache_key] = (1, time.time())

        # Cache the result
        with self._lock:
            self._cache[cache_key] = (info, time.time() + self._cache_ttl)

        return info

    def get_stats(self) -> Dict[str, int]:
        """Get validator statistics."""
        with self._lock:
            return {
                "cached_keys": len(self._cache),
            }


# Global validator instance
_validator_instance: Optional[APIKeyValidator] = None


def get_api_key_validator() -> APIKeyValidator:
    """Get the global API key validator.

    Returns:
        APIKeyValidator instance
    """
    global _validator_instance
    if _validator_instance is None:
        from ..config import settings

        _validator_instance = APIKeyValidator(
            orchestrator_url=settings.orchestrator_url,
            orchestrator_api_key=settings.orchestrator_api_key,
            reject_in_degraded_mode=getattr(settings, 'reject_auth_in_degraded_mode', True),
        )

    return _validator_instance


async def initialize_api_key_validator() -> bool:
    """Initialize the API key validator. Call this at startup.

    Returns:
        True if orchestrator is reachable, False otherwise.
        When orchestrator is unreachable, the service enters degraded mode.
        In degraded mode with reject_in_degraded_mode=True, all auth requests are rejected.
    """
    global _validator_instance

    from ..config import settings

    _validator_instance = APIKeyValidator(
        orchestrator_url=settings.orchestrator_url,
        orchestrator_api_key=settings.orchestrator_api_key,
        reject_in_degraded_mode=getattr(settings, 'reject_auth_in_degraded_mode', True),
    )

    healthy = await _validator_instance.initialize()
    if not healthy:
        _validator_instance.set_degraded(True)
        logger.warning(
            f"Orchestrator not reachable at {settings.orchestrator_url}. "
            "API key validation entering DEGRADED mode - "
            f"auth requests will be {'REJECTED' if _validator_instance._reject_in_degraded_mode else 'ALLOWED with cached keys'}."
        )
        return False

    logger.info("API key validator initialized (orchestrator-backed)")
    return True


def set_auth_degraded(degraded: bool) -> None:
    """Manually set the auth degraded mode flag.

    Args:
        degraded: True to enter degraded mode, False to recover
    """
    global _validator_instance
    if _validator_instance is not None:
        _validator_instance.set_degraded(degraded)


# FastAPI dependency functions
async def require_api_key(
    x_api_key: Optional[str] = Header(None, alias="X-API-Key"),
) -> APIKeyInfo:
    """FastAPI dependency to require and validate an API key.

    Args:
        x_api_key: The API key from the X-API-Key header

    Returns:
        APIKeyInfo if valid

    Raises:
        HTTPException: 401 if key is missing or invalid
    """
    if not x_api_key:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="API key required. Provide X-API-Key header.",
            headers={"WWW-Authenticate": "ApiKey"},
        )

    # Check if key is locked out due to too many failures
    rate_limiter = get_auth_rate_limiter()
    is_locked, retry_after = rate_limiter.check_lockout(x_api_key)
    if is_locked:
        raise HTTPException(
            status_code=status.HTTP_429_TOO_MANY_REQUESTS,
            detail=f"Too many auth failures. Retry after {retry_after} seconds.",
            headers={"Retry-After": str(retry_after)},
        )

    validator = get_api_key_validator()
    info = validator.validate_key_sync(x_api_key)

    if not info:
        # Record failure
        delay = rate_limiter.record_failure(x_api_key)
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid or expired API key",
            headers={"WWW-Authenticate": "ApiKey"},
        )

    # Clear failures on success
    rate_limiter.record_success(x_api_key)
    return info


def require_api_key_with_scope(scope: KeyScope):
    """Create a FastAPI dependency that requires a specific scope.

    Usage:
        @router.post("/api/embed")
        async def embed(request: EmbeddingRequest, api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.EMBED_WRITE)):
            ...

    Args:
        scope: Required scope

    Returns:
        Dependency function
    """
    async def dependency(
        x_api_key: Optional[str] = Header(None, alias="X-API-Key"),
    ) -> APIKeyInfo:
        if not x_api_key:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="API key required. Provide X-API-Key header.",
                headers={"WWW-Authenticate": "ApiKey"},
            )

        # Check if key is locked out
        rate_limiter = get_auth_rate_limiter()
        is_locked, retry_after = rate_limiter.check_lockout(x_api_key)
        if is_locked:
            raise HTTPException(
                status_code=status.HTTP_429_TOO_MANY_REQUESTS,
                detail=f"Too many auth failures. Retry after {retry_after} seconds.",
                headers={"Retry-After": str(retry_after)},
            )

        validator = get_api_key_validator()
        info = validator.validate_key_sync(x_api_key)

        if not info:
            delay = rate_limiter.record_failure(x_api_key)
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid or expired API key",
                headers={"WWW-Authenticate": "ApiKey"},
            )

        # Clear failures on success
        rate_limiter.record_success(x_api_key)

        if not info.has_scope(scope):
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail=f"Insufficient permissions. Required scope: {scope.value}",
            )

        return info

    return dependency


def require_api_key_with_scope(scope: KeyScope):
    """Create a FastAPI dependency that requires a specific scope.

    Usage:
        @router.post("/api/embed")
        async def embed(request: EmbeddingRequest, api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.EMBED_WRITE))):
            ...

    Args:
        scope: Required scope

    Returns:
        Dependency function
    """
    async def dependency(
        x_api_key: Optional[str] = Header(None, alias="X-API-Key"),
    ) -> APIKeyInfo:
        if not x_api_key:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="API key required. Provide X-API-Key header.",
                headers={"WWW-Authenticate": "ApiKey"},
            )

        validator = get_api_key_validator()
        info = await validator.validate_key_async(x_api_key)

        if not info:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid or expired API key",
                headers={"WWW-Authenticate": "ApiKey"},
            )

        if not info.has_scope(scope):
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail=f"Insufficient permissions. Required scope: {scope.value}",
            )

        return info

    return dependency
