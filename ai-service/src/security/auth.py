"""API key validation for FlyMind AI Service.

This module provides API key authentication and validation using the Go orchestrator's
PostgreSQL-backed key storage. The previous in-memory storage that lost keys on restart
has been replaced with persistent storage via the orchestrator API.

IMPORTANT: The old APIKeyValidator with _init_default_keys() should NOT be used in
production. Use PostgresBackedAPIKeyValidator instead.
"""

import hashlib
import logging
import os
import secrets
import time
from dataclasses import dataclass
from datetime import datetime, timedelta
from enum import Enum
from typing import Dict, List, Optional, Union
import threading

from fastapi import Header, HTTPException, Depends, status
from fastapi.security import APIKeyHeader

from .postgres_key_validator import (
    PostgresBackedAPIKeyValidator,
    InMemoryFallbackValidator,
    APIKeyInfo as PostgresAPIKeyInfo,
    KeyStatus,
    KeyScope,
    create_validator,
)

logger = logging.getLogger(__name__)

# FastAPI security scheme
api_key_header = APIKeyHeader(name="X-API-Key", auto_error=False)


# Re-export KeyStatus and KeyScope for backward compatibility
__all__ = [
    "KeyStatus",
    "KeyScope",
    "APIKeyInfo",
    "get_api_key_validator",
    "require_api_key",
    "require_api_key_with_scope",
]


@dataclass
class APIKeyInfo:
    """Information about an API key.

    This class provides backward compatibility for code that expects
    the old in-memory APIKeyInfo format. The actual validation is done
    by PostgresBackedAPIKeyValidator.
    """
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


class PostgresValidatorWrapper:
    """Wrapper around PostgresBackedAPIKeyValidator that provides sync interface.

    The underlying PostgresBackedAPIKeyValidator is async, but FastAPI dependencies
    can be sync when needed by using a simple wrapper that runs the async code
    in a thread pool.
    """

    def __init__(
        self,
        orchestrator_url: Optional[str] = None,
        orchestrator_api_key: Optional[str] = None,
    ):
        self._validator: Optional[PostgresBackedAPIKeyValidator] = None
        self._fallback: Optional[InMemoryFallbackValidator] = None
        self._is_fallback = False
        self._orchestrator_url = orchestrator_url
        self._orchestrator_api_key = orchestrator_api_key
        self._initialized = False

        # Check for development mode
        self._dev_mode = os.getenv("DEVELOPMENT", "false").lower() == "true"
        self._force_fallback = os.getenv("AI_SERVICE_USE_IN_MEMORY_KEY_VALIDATOR", "false").lower() == "true"

    def initialize(self) -> None:
        """Initialize the validator. Call this at startup."""
        if self._initialized:
            return

        if self._force_fallback or self._dev_mode:
            logger.warning(
                "=" * 60
            )
            logger.warning("INSECURE MODE: Using in-memory API key validation!")
            logger.warning("Keys will be lost on restart and no key revocation works.")
            if self._dev_mode:
                logger.warning("This is expected in DEVELOPMENT mode.")
            else:
                logger.warning("Set AI_SERVICE_USE_IN_MEMORY_KEY_VALIDATOR=false to disable.")
            logger.warning("DO NOT use this in production!")
            logger.warning("=" * 60)

            self._fallback = InMemoryFallbackValidator()
            self._is_fallback = True
        else:
            self._validator = PostgresBackedAPIKeyValidator(
                orchestrator_url=self._orchestrator_url or "http://localhost:8080",
                orchestrator_api_key=self._orchestrator_api_key,
            )

        self._initialized = True

    async def _async_initialize(self) -> None:
        """Async initialization for use at startup."""
        if self._is_fallback or self._force_fallback:
            self._fallback = InMemoryFallbackValidator()
            self._is_fallback = True
            return

        self._validator = PostgresBackedAPIKeyValidator(
            orchestrator_url=self._orchestrator_url or "http://localhost:8080",
            orchestrator_api_key=self._orchestrator_api_key,
        )
        await self._validator.initialize()

    def validate_key(self, key: str) -> Optional[APIKeyInfo]:
        """Validate an API key (sync interface for FastAPI dependencies)."""
        if not self._initialized:
            self.initialize()

        import asyncio

        if self._is_fallback and self._fallback:
            # Run async validation in thread pool
            return asyncio.run(self._fallback.validate_key(key))

        if self._validator:
            return asyncio.run(self._validator.validate_key(key))

        return None

    def get_stats(self) -> Dict[str, int]:
        """Get validator statistics."""
        if self._is_fallback and self._fallback:
            return {"mode": "in_memory_fallback"}

        if self._validator:
            return self._validator.get_stats()

        return {"mode": "uninitialized"}


# Global validator instance
_validator_instance: Optional[PostgresValidatorWrapper] = None


def get_api_key_validator() -> PostgresValidatorWrapper:
    """Get the global API key validator.

    Returns:
        PostgresValidatorWrapper instance (uses orchestrator-backed storage in production)
    """
    global _validator_instance
    if _validator_instance is None:
        from ..config import settings

        _validator_instance = PostgresValidatorWrapper(
            orchestrator_url=settings.orchestrator_url,
            orchestrator_api_key=settings.orchestrator_api_key,
        )
        _validator_instance.initialize()

    return _validator_instance


async def initialize_api_key_validator() -> None:
    """Initialize the API key validator asynchronously. Call this at startup."""
    global _validator_instance

    from ..config import settings

    wrapper = PostgresValidatorWrapper(
        orchestrator_url=settings.orchestrator_url,
        orchestrator_api_key=settings.orchestrator_api_key,
    )
    await wrapper._async_initialize()
    _validator_instance = wrapper


# Backward compatibility alias - the old APIKeyValidator is deprecated
# and should not be used in production
APIKeyValidator = PostgresValidatorWrapper


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

    validator = get_api_key_validator()
    info = validator.validate_key(x_api_key)

    if not info:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid or expired API key",
            headers={"WWW-Authenticate": "ApiKey"},
        )

    return info


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
        info = validator.validate_key(x_api_key)

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