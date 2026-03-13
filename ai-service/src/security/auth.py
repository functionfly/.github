"""API key validation for FlyMind AI Service.

This module provides API key authentication and validation.
"""

import hashlib
import secrets
import time
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum
from typing import Dict, List, Optional
import threading
import logging

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
    rate_limit: int = 60  # requests per minute

    def is_valid(self) -> bool:
        """Check if the key is valid."""
        if self.status != KeyStatus.ACTIVE:
            return False

        if self.expires_at and datetime.utcnow() > self.expires_at:
            return False

        return True


class APIKeyValidator:
    """Validates API keys for authentication."""

    def __init__(self):
        """Initialize the API key validator."""
        self._logger = logging.getLogger(__name__)
        self._lock = threading.Lock()

        # Key storage: key_id -> (hashed_key, info)
        self._keys: Dict[str, tuple[str, APIKeyInfo]] = {}

        # Lookup by full key for fast validation
        self._key_lookup: Dict[str, str] = {}  # hashed_key -> key_id

        # Stats
        self._total_validations = 0
        self._failed_validations = 0

    def create_key(
        self,
        tenant_id: str,
        name: str,
        scopes: List[KeyScope],
        expires_in_days: Optional[int] = None,
        rate_limit: int = 60,
    ) -> tuple[str, APIKeyInfo]:
        """Create a new API key.

        Args:
            tenant_id: Tenant ID
            name: Key name
            scopes: List of scopes
            expires_in_days: Days until expiration (None for no expiration)
            rate_limit: Requests per minute

        Returns:
            Tuple of (full_key, APIKeyInfo)
        """
        with self._lock:
            # Generate key
            key_id = secrets.token_hex(8)
            full_key = f"fly_{secrets.token_hex(32)}"

            # Hash the key for storage
            key_hash = self._hash_key(full_key)

            # Create info
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

            # Store
            self._keys[key_id] = (key_hash, info)
            self._key_lookup[key_hash] = key_id

            self._logger.info(f"Created API key {key_id} for tenant {tenant_id}")

            return full_key, info

    def validate_key(self, key: str) -> Optional[APIKeyInfo]:
        """Validate an API key.

        Args:
            key: The full API key

        Returns:
            APIKeyInfo if valid, None otherwise
        """
        with self._lock:
            self._total_validations += 1

            # Hash the key
            key_hash = self._hash_key(key)

            # Look up
            key_id = self._key_lookup.get(key_hash)
            if not key_id:
                self._failed_validations += 1
                return None

            _, info = self._keys.get(key_id, (None, None))
            if not info:
                self._failed_validations += 1
                return None

            # Check validity
            if not info.is_valid():
                self._failed_validations += 1
                return None

            # Update stats
            info.last_used_at = datetime.utcnow()
            info.request_count += 1

            return info

    def revoke_key(self, key_id: str) -> bool:
        """Revoke an API key.

        Args:
            key_id: The key ID

        Returns:
            True if revoked, False if not found
        """
        with self._lock:
            if key_id not in self._keys:
                return False

            _, info = self._keys[key_id]
            info.status = KeyStatus.REVOKED

            # Remove from lookup
            for hash_val, k_id in list(self._key_lookup.items()):
                if k_id == key_id:
                    del self._key_lookup[hash_val]

            self._logger.info(f"Revoked API key {key_id}")
            return True

    def get_key_info(self, key_id: str) -> Optional[APIKeyInfo]:
        """Get information about an API key.

        Args:
            key_id: The key ID

        Returns:
            APIKeyInfo or None
        """
        with self._lock:
            _, info = self._keys.get(key_id, (None, None))
            return info

    def list_keys(self, tenant_id: str) -> List[APIKeyInfo]:
        """List all keys for a tenant.

        Args:
            tenant_id: Tenant ID

        Returns:
            List of APIKeyInfo
        """
        with self._lock:
            return [
                info for _, info in self._keys.values()
                if info.tenant_id == tenant_id
            ]

    def _hash_key(self, key: str) -> str:
        """Hash an API key for storage.

        Args:
            key: The full key

        Returns:
            Hashed key
        """
        return hashlib.sha256(key.encode()).hexdigest()

    def get_stats(self) -> Dict[str, int]:
        """Get validator statistics.

        Returns:
            Dictionary with stats
        """
        with self._lock:
            return {
                "total_keys": len(self._keys),
                "total_validations": self._total_validations,
                "failed_validations": self._failed_validations,
            }

    # Default keys for development
    def _init_default_keys(self) -> None:
        """Initialize default development keys."""
        # Create a default key for development
        self.create_key(
            tenant_id="default",
            name="Development Key",
            scopes=[KeyScope.FULL],
            rate_limit=100,
        )


# Global validator
_api_key_validator: Optional[APIKeyValidator] = None


def get_api_key_validator() -> APIKeyValidator:
    """Get the global API key validator.

    Returns:
        APIKeyValidator instance
    """
    global _api_key_validator
    if _api_key_validator is None:
        _api_key_validator = APIKeyValidator()
        _api_key_validator._init_default_keys()

    return _api_key_validator
