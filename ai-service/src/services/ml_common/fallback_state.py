"""File-based fallback persistence for ML state when Redis is unavailable.

Provides reliable storage using the filesystem with atomic writes.
Includes optional encryption for sensitive ML state data.
"""

import base64
import hashlib
import json
import logging
import os
import re
import shutil
import tempfile
import threading
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional

from ...config import settings

logger = logging.getLogger(__name__)


class FallbackEncryption:
    """Handles encryption for file-based fallback storage.

    Uses AES-256-GCM for authenticated encryption of stored data.
    """

    def __init__(self):
        self._key: Optional[bytes] = None
        self._aesgcm: Optional[Any] = None
        self._is_enabled = False
        self._initialize()

    def _initialize(self) -> None:
        """Initialize encryption from settings."""
        from cryptography.hazmat.primitives.ciphers.aead import AESGCM

        encryption_key = getattr(settings, 'redis_cache_encryption_key', None)
        if not encryption_key:
            logger.warning(
                "File-based fallback encryption disabled: no encryption key configured. "
                "Set REDIS_CACHE_ENCRYPTION_KEY to enable encryption."
            )
            return

        try:
            if isinstance(encryption_key, str):
                key_bytes = encryption_key.encode()
            else:
                key_bytes = encryption_key

            if len(key_bytes) < 32:
                logger.warning("Encryption key too short, deriving 32 bytes using SHA256")
                key_bytes = hashlib.sha256(key_bytes).digest()

            self._key = key_bytes[:32]
            self._aesgcm = AESGCM(self._key)
            self._is_enabled = True
            logger.info("File-based fallback encryption enabled")
        except Exception as e:
            logger.warning(f"Failed to initialize fallback encryption: {e}")
            self._is_enabled = False

    def is_enabled(self) -> bool:
        """Check if encryption is enabled."""
        return self._is_enabled

    def _generate_nonce(self) -> bytes:
        """Generate a random 12-byte nonce for AES-GCM."""
        import secrets
        return secrets.token_bytes(12)

    def encrypt(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Encrypt data for storage.

        Args:
            data: The dict data to encrypt

        Returns:
            Dict with encryption metadata and ciphertext
        """
        if not self._is_enabled:
            return {"encrypted": False, "data": data}

        try:
            plaintext = json.dumps(data, default=str).encode('utf-8')
            nonce = self._generate_nonce()
            ciphertext = self._aesgcm.encrypt(nonce, plaintext, None)

            return {
                "encrypted": True,
                "nonce": base64.b64encode(nonce).decode('utf-8'),
                "ciphertext": base64.b64encode(ciphertext).decode('utf-8'),
            }
        except Exception as e:
            logger.error(f"Encryption failed: {e}")
            return {"encrypted": False, "data": data, "error": str(e)}

    def decrypt(self, stored: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Decrypt stored data.

        Args:
            stored: The stored encrypted data

        Returns:
            Decrypted dict or None if decryption fails
        """
        if not stored.get("encrypted", False):
            return stored.get("data", stored)

        try:
            nonce = base64.b64decode(stored["nonce"])
            ciphertext = base64.b64decode(stored["ciphertext"])
            plaintext = self._aesgcm.decrypt(nonce, ciphertext, None)
            return json.loads(plaintext.decode('utf-8'))
        except Exception as e:
            logger.error(f"Decryption failed: {e}")
            return None


_fallback_encryption: Optional[FallbackEncryption] = None


def get_fallback_encryption() -> FallbackEncryption:
    """Get the global fallback encryption instance."""
    global _fallback_encryption
    if _fallback_encryption is None:
        _fallback_encryption = FallbackEncryption()
    return _fallback_encryption


class FileBasedMLStore:
    """File-based fallback storage for ML state.

    Uses atomic file operations to ensure data consistency.
    Thread-safe for concurrent access.
    Supports optional encryption for sensitive data.
    """

    def __init__(self, tenant_id: str, namespace: str):
        """Initialize file-based store.

        Args:
            tenant_id: Tenant ID for isolation
            namespace: Service namespace (e.g., 'cost_anomaly', 'thompson')
        """
        self._tenant_id = tenant_id
        self._namespace = namespace
        self._base_dir = Path(settings.ml_fallback_file_dir) / tenant_id / namespace
        self._lock = threading.Lock()
        self._encryption = get_fallback_encryption()
        self._ensure_dir()

    def _ensure_dir(self) -> None:
        """Ensure the base directory exists."""
        try:
            self._base_dir.mkdir(parents=True, exist_ok=True)
            os.chmod(self._base_dir, 0o700)
        except Exception as e:
            logger.error(f"Failed to create fallback directory {self._base_dir}: {e}")

    def _get_file_path(self, key: str) -> Path:
        """Get the file path for a key."""
        safe_key = key.replace(":", "_").replace("/", "_")
        return self._base_dir / f"{safe_key}.json"

    def _atomic_write(self, path: Path, data: Dict[str, Any]) -> bool:
        """Atomically write data to a file.

        Uses a temporary file and rename for atomicity.
        Encrypts data if encryption is enabled.
        """
        try:
            temp_fd, temp_path = tempfile.mkstemp(
                dir=self._base_dir,
                prefix=".tmp_",
                suffix=".json"
            )
            try:
                with os.fdopen(temp_fd, "w") as f:
                    json.dump(data, f, indent=2, default=str)
                os.replace(temp_path, path)
                os.chmod(path, 0o600)
                return True
            except Exception:
                try:
                    os.unlink(temp_path)
                except Exception:
                    pass
                raise
        except Exception as e:
            logger.error(f"Atomic write failed for {path}: {e}")
            return False

    def get(self, key: str) -> Optional[Dict[str, Any]]:
        """Get a value by key.

        Args:
            key: The key to retrieve

        Returns:
            The stored dict or None if not found
        """
        with self._lock:
            path = self._get_file_path(key)
            if not path.exists():
                return None
            try:
                with open(path) as f:
                    stored = json.load(f)
                return self._decrypt_and_extract(stored)
            except Exception as e:
                logger.warning(f"Failed to read {key}: {e}")
                return None

    def _decrypt_and_extract(self, stored: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Decrypt stored data and extract the value.

        Args:
            stored: The stored data dict

        Returns:
            Decrypted value or None
        """
        if not stored.get("encrypted", False):
            return stored.get("value", stored)

        return self._encryption.decrypt(stored)

    def set(self, key: str, value: Dict[str, Any], ttl_seconds: Optional[int] = None) -> bool:
        """Set a value with optional TTL.

        Args:
            key: The key to set
            value: The dict value to store
            ttl_seconds: Optional TTL in seconds

        Returns:
            True if successful
        """
        with self._lock:
            path = self._get_file_path(key)

            stored_data = {
                "stored_at": datetime.utcnow().isoformat(),
                "ttl_seconds": ttl_seconds,
                "expires_at": (
                    datetime.utcnow().timestamp() + ttl_seconds
                    if ttl_seconds else None
                ),
            }

            if self._encryption.is_enabled():
                encrypted = self._encryption.encrypt(value)
                stored_data.update(encrypted)
            else:
                stored_data["encrypted"] = False
                stored_data["data"] = value

            return self._atomic_write(path, stored_data)

    def delete(self, key: str) -> bool:
        """Delete a key.

        Args:
            key: The key to delete

        Returns:
            True if successful or key didn't exist
        """
        with self._lock:
            path = self._get_file_path(key)
            try:
                if path.exists():
                    path.unlink()
                return True
            except Exception as e:
                logger.error(f"Failed to delete {key}: {e}")
                return False

    def list_keys(self, pattern: str = "*") -> List[str]:
        """List keys matching a pattern.

        Args:
            pattern: Glob pattern (default: all keys)

        Returns:
            List of matching keys
        """
        with self._lock:
            try:
                prefix = pattern.replace("*", "")
                keys = []
                for path in self._base_dir.glob("*.json"):
                    key = path.stem.replace("tmp_", "").replace("_", ":")
                    if prefix in key or pattern == "*":
                        keys.append(key)
                return keys
            except Exception as e:
                logger.error(f"Failed to list keys: {e}")
                return []

    def cleanup_expired(self) -> int:
        """Remove expired entries.

        Returns:
            Number of entries removed
        """
        with self._lock:
            removed = 0
            now = datetime.utcnow().timestamp()
            try:
                for path in self._base_dir.glob("*.json"):
                    try:
                        with open(path) as f:
                            data = json.load(f)
                        if data.get("expires_at") and data["expires_at"] < now:
                            path.unlink()
                            removed += 1
                    except Exception:
                        continue
            except Exception as e:
                logger.error(f"Failed to cleanup expired: {e}")
            return removed

    def get_all(self) -> List[Dict[str, Any]]:
        """Get all stored entries.

        Returns:
            List of all stored values
        """
        with self._lock:
            results = []
            try:
                for path in self._base_dir.glob("*.json"):
                    try:
                        with open(path) as f:
                            data = json.load(f)
                        value = self._decrypt_and_extract(data)
                        if value:
                            results.append(value)
                    except Exception:
                        continue
            except Exception as e:
                logger.error(f"Failed to get all: {e}")
            return results

    def clear(self) -> bool:
        """Clear all entries.

        Returns:
            True if successful
        """
        with self._lock:
            try:
                for path in self._base_dir.glob("*.json"):
                    path.unlink()
                return True
            except Exception as e:
                logger.error(f"Failed to clear: {e}")
                return False


class FallbackMLStateManager:
    """Manages file-based fallback for all ML services.

    Coordinates file-based storage across services when Redis is unavailable.
    """

    def __init__(self):
        self._stores: Dict[str, FileBasedMLStore] = {}
        self._lock = threading.Lock()
        self._ensure_base_dir()

    def _ensure_base_dir(self) -> None:
        """Ensure the base fallback directory exists."""
        Path(settings.ml_fallback_file_dir).mkdir(parents=True, exist_ok=True)

    def get_store(self, tenant_id: str, namespace: str) -> FileBasedMLStore:
        """Get a file-based store for a tenant and service.

        Args:
            tenant_id: Tenant ID
            namespace: Service namespace

        Returns:
            FileBasedMLStore instance
        """
        key = f"{tenant_id}:{namespace}"
        with self._lock:
            if key not in self._stores:
                self._stores[key] = FileBasedMLStore(tenant_id, namespace)
            return self._stores[key]

    def cleanup_expired_all(self) -> int:
        """Cleanup expired entries across all stores.

        Returns:
            Total number of entries removed
        """
        total = 0
        with self._lock:
            for store in self._stores.values():
                total += store.cleanup_expired()
        return total


_fallback_manager: Optional[FallbackMLStateManager] = None


def get_fallback_ml_state_manager() -> FallbackMLStateManager:
    """Get the global fallback state manager."""
    global _fallback_manager
    if _fallback_manager is None:
        _fallback_manager = FallbackMLStateManager()
    return _fallback_manager
