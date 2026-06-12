"""Encryption utilities for FlyMind AI Service.

Provides Fernet-based encryption for sensitive data at rest.
"""

import base64
import logging
from typing import Optional

from cryptography.fernet import Fernet

logger = logging.getLogger(__name__)


class CacheEncryption:
    """Fernet-based encryption for cache data."""

    def __init__(self, encryption_key: Optional[str] = None):
        """Initialize encryption with a base64-encoded Fernet key.

        Args:
            encryption_key: Base64-encoded Fernet key. If None, encryption is disabled.
        """
        self._fernet: Optional[Fernet] = None
        if encryption_key:
            try:
                self._fernet = Fernet(encryption_key.encode())
                logger.info("Cache encryption initialized")
            except Exception as e:
                logger.error(f"Failed to initialize cache encryption: {e}")
                self._fernet = None

    @property
    def is_enabled(self) -> bool:
        """Check if encryption is enabled."""
        return self._fernet is not None

    def encrypt(self, data: str) -> str:
        """Encrypt a string and return base64-encoded ciphertext.

        Args:
            data: Plain text string to encrypt

        Returns:
            Base64-encoded encrypted string

        Raises:
            ValueError: If encryption is not enabled
        """
        if not self._fernet:
            raise ValueError("Encryption not enabled")

        encrypted = self._fernet.encrypt(data.encode())
        return base64.urlsafe_b64encode(encrypted).decode()

    def decrypt(self, encrypted_data: str) -> str:
        """Decrypt base64-encoded ciphertext back to plain text.

        Args:
            encrypted_data: Base64-encoded encrypted string

        Returns:
            Decrypted plain text string

        Raises:
            ValueError: If encryption is not enabled
        """
        if not self._fernet:
            raise ValueError("Encryption not enabled")

        try:
            data = base64.urlsafe_b64decode(encrypted_data.encode())
            decrypted = self._fernet.decrypt(data)
            return decrypted.decode()
        except Exception as e:
            logger.error(f"Decryption failed: {e}")
            raise ValueError("Failed to decrypt data")


def generate_fernet_key() -> str:
    """Generate a new Fernet key for cache encryption.

    Returns:
        Base64-encoded Fernet key
    """
    return Fernet.generate_key().decode()


# Global encryption instance
_cache_encryption: Optional[CacheEncryption] = None


def get_cache_encryption() -> CacheEncryption:
    """Get the global cache encryption instance.

    Returns:
        CacheEncryption instance
    """
    global _cache_encryption
    if _cache_encryption is None:
        from ..config import settings
        _cache_encryption = CacheEncryption(settings.redis_cache_encryption_key)
    return _cache_encryption
