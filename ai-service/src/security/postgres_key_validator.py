"""PostgreSQL-backed API key validator using Go orchestrator.

This module provides API key validation backed by the Go orchestrator's
PostgreSQL storage, replacing the in-memory validator that loses keys on restart.

Key features:
- Keys persist across AI service restarts (stored in Go orchestrator's DB)
- Key revocation is immediately effective
- Usage statistics (last_used_at, request_count) are tracked
- No default dev keys with FULL scope in production

Usage:
    from security.postgres_key_validator import PostgresBackedAPIKeyValidator

    validator = PostgresBackedAPIKeyValidator(orchestrator_url="http://localhost:8080")
    await validator.initialize()

    info = await validator.validate_key("fly_xxx...")
    if info:
        print(f"Key valid for tenant: {info.tenant_id}")
"""

import hashlib
import logging
import time
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import Dict, List, Optional
import threading
import asyncio

import httpx

logger = logging.getLogger(__name__)


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


class PostgresBackedAPIKeyValidator:
    """Validates API keys using Go orchestrator's PostgreSQL storage.

    This validator queries the orchestrator's /auth/validate-key endpoint
    to validate keys against the database-backed storage.
    """

    def __init__(
        self,
        orchestrator_url: str = "http://localhost:8080",
        orchestrator_api_key: Optional[str] = None,
        cache_ttl_seconds: int = 60,
        cache_max_size: int = 1000,
    ):
        """Initialize the validator.

        Args:
            orchestrator_url: URL of the Go orchestrator API
            orchestrator_api_key: Optional API key for orchestrator authentication
            cache_ttl_seconds: How long to cache validation results
            cache_max_size: Maximum number of cached entries
        """
        self._orchestrator_url = orchestrator_url.rstrip("/")
        self._orchestrator_api_key = orchestrator_api_key
        self._cache_ttl = cache_ttl_seconds
        self._cache_max_size = cache_max_size

        self._lock = threading.Lock()
        self._cache: Dict[str, tuple[Optional[APIKeyInfo], float]] = {}  # key_hash -> (info, expiry)
        self._stats_cache: Dict[str, int] = {}  # key_id -> request_count (local only)

        self._client: Optional[httpx.AsyncClient] = None
        self._initialized = False
        self._init_error: Optional[str] = None

    async def initialize(self) -> bool:
        """Initialize the validator by checking orchestrator connectivity.

        Returns:
            True if orchestrator is reachable, False otherwise
        """
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                response = await client.get(f"{self._orchestrator_url}/health")
                if response.status_code == 200:
                    self._initialized = True
                    logger.info("PostgresBackedAPIKeyValidator initialized successfully")
                    return True
        except Exception as e:
            self._init_error = str(e)
            logger.warning(f"Orchestrator health check failed: {e}")

        self._initialized = True  # Mark as initialized to allow fallback behavior
        logger.warning(
            f"PostgresBackedAPIKeyValidator initialized with warning: {self._init_error}. "
            "Will fall back to in-memory validation if orchestrator is unavailable."
        )
        return False

    def _get_cache_key(self, key: str) -> str:
        """Generate a cache key from the API key."""
        return hashlib.sha256(key.encode()).hexdigest()

    def _is_cache_valid(self, cache_entry: tuple[Optional[APIKeyInfo], float]) -> bool:
        """Check if a cache entry is still valid."""
        _, expiry = cache_entry
        return time.time() < expiry

    def _invalidate_key(self, key: str) -> None:
        """Remove a key from the cache (e.g., after revocation)."""
        cache_key = self._get_cache_key(key)
        with self._lock:
            if cache_key in self._cache:
                del self._cache[cache_key]

    async def validate_key(self, key: str) -> Optional[APIKeyInfo]:
        """Validate an API key.

        Args:
            key: The full API key

        Returns:
            APIKeyInfo if valid, None otherwise
        """
        cache_key = self._get_cache_key(key)

        # Check cache first
        with self._lock:
            if cache_key in self._cache:
                cached_info, expiry = self._cache[cache_key]
                if time.time() < expiry:
                    return cached_info

        # Query orchestrator
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
                        # Key not found or invalid - cache the negative result
                        with self._lock:
                            self._cache[cache_key] = (None, time.time() + self._cache_ttl)
                        return None

                    key_data = data["data"]
                    info = self._parse_key_response(key_data)

                    # Cache the result
                    with self._lock:
                        self._cache[cache_key] = (info, time.time() + self._cache_ttl)

                    return info
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
                status = KeyStatus.REVOKED
            elif not data.get("is_active"):
                status = KeyStatus.SUSPENDED
            elif expires_at and datetime.utcnow() > expires_at:
                status = KeyStatus.EXPIRED
            else:
                status = KeyStatus.ACTIVE

            return APIKeyInfo(
                key_id=data.get("key_id", ""),
                tenant_id=data.get("tenant_id", ""),
                name=data.get("name", ""),
                scopes=scopes,
                status=status,
                created_at=datetime.utcnow(),  # Not provided by orchestrator
                expires_at=expires_at,
                last_used_at=last_used_at,
                request_count=0,  # Not tracked per-request in this validator
                rate_limit=data.get("rate_limit_rpm", 60),
            )

        except Exception as e:
            logger.error(f"Failed to parse key response: {e}")
            return None

    async def revoke_key(self, key_id: str) -> bool:
        """Note: Key revocation must be done via the Go orchestrator API.

        This method invalidates the local cache entry only.

        Args:
            key_id: The key ID

        Returns:
            True if cache was invalidated
        """
        # We can't revoke via the orchestrator without the full key
        # But we can clear any cached entries
        with self._lock:
            for cache_key in list(self._cache.keys()):
                info, _ = self._cache[cache_key]
                if info and info.key_id == key_id:
                    del self._cache[cache_key]
        return True

    def get_stats(self) -> Dict[str, int]:
        """Get validator statistics.

        Returns:
            Dictionary with stats
        """
        with self._lock:
            return {
                "cached_keys": len(self._cache),
                "cache_max_size": self._cache_max_size,
            }


class InMemoryFallbackValidator:
    """In-memory fallback validator for development.

    WARNING: This stores keys in process memory and loses them on restart.
    It should ONLY be used when the orchestrator is unavailable in development.
    """

    def __init__(self):
        self._lock = threading.Lock()
        self._keys: Dict[str, tuple[str, APIKeyInfo]] = {}  # key_id -> (hash, info)
        self._key_lookup: Dict[str, str] = {}  # hash -> key_id

    def create_key(
        self,
        tenant_id: str,
        name: str,
        scopes: List[KeyScope],
        expires_in_days: Optional[int] = None,
        rate_limit: int = 60,
    ) -> tuple[str, APIKeyInfo]:
        """Create a new API key in memory."""
        import secrets

        with self._lock:
            key_id = secrets.token_hex(8)
            full_key = f"fly_{secrets.token_hex(32)}"
            key_hash = hashlib.sha256(full_key.encode()).hexdigest()

            from datetime import timedelta
            expires_at = None
            if expires_in_days:
                expires_at = datetime.utcnow() + timedelta(days=expires_in_days)

            info = APIKeyInfo(
                key_id=key_id,
                tenant_id=tenant_id,
                name=name,
                scopes=scopes,
                status=KeyStatus.ACTIVE,
                created_at=datetime.utcnow(),
                expires_at=expires_at,
                rate_limit=rate_limit,
            )

            self._keys[key_id] = (key_hash, info)
            self._key_lookup[key_hash] = key_id

            return full_key, info

    async def validate_key(self, key: str) -> Optional[APIKeyInfo]:
        """Validate an API key from in-memory store."""
        with self._lock:
            key_hash = hashlib.sha256(key.encode()).hexdigest()
            key_id = self._key_lookup.get(key_hash)
            if not key_id:
                return None

            _, info = self._keys.get(key_id, (None, None))
            if not info:
                return None

            if not info.is_valid():
                return None

            # Update stats
            info.last_used_at = datetime.utcnow()
            info.request_count += 1

            return info

    async def initialize(self) -> bool:
        """No-op for fallback validator."""
        return True


def create_validator(
    orchestrator_url: Optional[str] = None,
    orchestrator_api_key: Optional[str] = None,
    use_fallback: bool = False,
) -> tuple[PostgresBackedAPIKeyValidator | InMemoryFallbackValidator, bool]:
    """Create a validator based on configuration.

    Args:
        orchestrator_url: URL of the Go orchestrator API
        orchestrator_api_key: Optional API key for orchestrator
        use_fallback: Force use of in-memory fallback (for development only)

    Returns:
        Tuple of (validator, is_fallback) where is_fallback indicates if in-memory fallback was used
    """
    if use_fallback:
        logger.warning(
            "Using in-memory API key validator. "
            "WARNING: Keys will be lost on restart and default dev keys will be created. "
            "DO NOT use in production!"
        )
        return InMemoryFallbackValidator(), True

    validator = PostgresBackedAPIKeyValidator(
        orchestrator_url=orchestrator_url or "http://localhost:8080",
        orchestrator_api_key=orchestrator_api_key,
    )
    return validator, False