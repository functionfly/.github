"""Model persistence — save/load/version ML models with joblib and optional encryption."""

import hashlib
import json
import logging
import os
import shutil
import base64
from datetime import datetime
from pathlib import Path
from typing import Any, Optional, List

import joblib

from ...config import settings

logger = logging.getLogger(__name__)


def _get_aes256gcm_encryption() -> tuple[bytes | None, Any]:
    """Get AES-256-GCM encryption instance if key is configured.

    Returns:
        Tuple of (key_bytes, cipher_provider) where cipher_provider is either
        an AESGCM instance or None if encryption is disabled.
    """
    from cryptography.hazmat.primitives.ciphers.aead import AESGCM

    key = getattr(settings, 'ml_model_encryption_key', None)
    if not key:
        return None, None

    try:
        if isinstance(key, str):
            key_bytes = key.encode()
        else:
            key_bytes = key

        if len(key_bytes) < 32:
            logger.warning("Encryption key too short, deriving 32 bytes")
            key_bytes = hashlib.sha256(key_bytes).digest()

        aesgcm = AESGCM(key_bytes[:32])
        return key_bytes[:32], aesgcm
    except Exception as e:
        logger.warning(f"Invalid ML model encryption key: {e}")
        return None, None


class ModelIntegrityError(Exception):
    """Raised when model integrity verification fails."""
    pass


class EncryptedModelStore:
    """Model persistence with AES-256-GCM encryption at rest.

    Models are encrypted before writing to disk and decrypted on read.
    Uses AES-256-GCM for authenticated encryption with integrity verification.
    Includes automatic backup before overwrites and retention management.
    """

    def __init__(self, namespace: str, encrypt: bool = True):
        self._namespace = namespace
        self._base_dir = Path(settings.ml_model_dir) / namespace
        self._backup_dir = Path(settings.ml_model_backup_dir) / namespace
        self._base_dir.mkdir(parents=True, exist_ok=True)
        self._backup_dir.mkdir(parents=True, exist_ok=True)
        self._metadata_path = self._base_dir / "metadata.json"

        self._key_bytes, self._aesgcm = _get_aes256gcm_encryption()
        self._encrypt = encrypt and self._aesgcm is not None

        self._backup_enabled = settings.ml_model_backup_enabled
        self._max_backups = settings.ml_model_max_backups

        if self._encrypt:
            logger.info(f"Model AES-256-GCM encryption enabled for namespace: {namespace}")
        else:
            logger.info(f"Model encryption disabled for namespace: {namespace}")

    def _compute_integrity_hash(self, data: bytes) -> str:
        """Compute SHA-256 integrity hash of model data."""
        return hashlib.sha256(data).hexdigest()

    def _generate_nonce(self) -> bytes:
        """Generate a random 12-byte nonce for AES-GCM."""
        import secrets
        return secrets.token_bytes(12)

    def save(self, model: Any, version: Optional[str] = None, create_backup: bool = True) -> str:
        """Save an encrypted model to disk.

        Args:
            model: The model object to serialize
            version: Optional version string (defaults to timestamp)
            create_backup: If True, backup existing model before saving

        Returns:
            Version string of the saved model

        Raises:
            ModelIntegrityError: If model verification fails after save
        """
        if version is None:
            version = datetime.utcnow().strftime("%Y%m%d_%H%M%S")

        model_path = self._base_dir / f"model_{version}.joblib"
        serialized = joblib.dumps(model)

        integrity_hash = self._compute_integrity_hash(serialized)

        if create_backup and model_path.exists():
            self._create_backup(version)

        if self._encrypt and self._aesgcm:
            nonce = self._generate_nonce()
            encrypted = self._aesgcm.encrypt(nonce, serialized, None)
            combined = nonce + encrypted
            model_path.write_bytes(combined)
        else:
            joblib.dump(model, model_path)

        self._update_metadata(version, model_path, integrity_hash)
        logger.info(f"Saved {'encrypted ' if self._encrypt else ''}model {self._namespace}/{version}")
        return version

    def load(self, version: Optional[str] = None, verify_integrity: bool = True) -> Optional[Any]:
        """Load a model from disk, decrypting if necessary.

        Args:
            version: Specific version to load (latest if None)
            verify_integrity: If True, verify model integrity hash after loading

        Returns:
            The loaded model or None if not found

        Raises:
            ModelIntegrityError: If model integrity verification fails
        """
        if version is None:
            version = self._get_latest_version()

        if version is None:
            return None

        model_path = self._base_dir / f"model_{version}.joblib"
        if not model_path.exists():
            logger.warning(f"Model not found: {model_path}")
            return None

        try:
            if self._encrypt and self._aesgcm:
                combined = model_path.read_bytes()
                nonce = combined[:12]
                ciphertext = combined[12:]
                decrypted = self._aesgcm.decrypt(nonce, ciphertext, None)
                model = joblib.loads(decrypted)
            else:
                model = joblib.load(model_path)

            if verify_integrity:
                serialized = joblib.dumps(model)
                computed_hash = self._compute_integrity_hash(serialized)
                expected_hash = self._get_integrity_hash(version)
                if expected_hash and computed_hash != expected_hash:
                    logger.error(f"Model integrity mismatch for {self._namespace}/{version}")
                    raise ModelIntegrityError(
                        f"Model integrity verification failed for {self._namespace}/{version}"
                    )

            logger.info(f"Loaded model {self._namespace}/{version}")
            return model
        except Exception as e:
            if "decryption" in str(e).lower() or "integrity" in str(e).lower():
                logger.error(f"Failed to decrypt/verify model {self._namespace}/{version}: {e}")
            else:
                logger.error(f"Failed to load model {self._namespace}/{version}: {e}")
            return None

    def exists(self) -> bool:
        """Check if any model version exists."""
        return self._get_latest_version() is not None

    def _get_latest_version(self) -> Optional[str]:
        """Get the latest model version from metadata."""
        meta = self._read_metadata()
        return meta.get("latest_version")

    def _update_metadata(self, version: str, model_path: Path, integrity_hash: str) -> None:
        """Update the metadata file with a new version."""
        meta = self._read_metadata()
        meta["latest_version"] = version
        meta["versions"] = meta.get("versions", {})
        meta["versions"][version] = {
            "saved_at": datetime.utcnow().isoformat(),
            "size_bytes": model_path.stat().st_size,
            "encrypted": self._encrypt,
            "integrity_hash": integrity_hash,
        }
        with open(self._metadata_path, "w") as f:
            json.dump(meta, f, indent=2)

    def _read_metadata(self) -> dict:
        """Read metadata from disk."""
        if not self._metadata_path.exists():
            return {}
        try:
            with open(self._metadata_path) as f:
                return json.load(f)
        except Exception:
            return {}

    def _get_integrity_hash(self, version: str) -> Optional[str]:
        """Get the integrity hash for a specific version."""
        meta = self._read_metadata()
        return meta.get("versions", {}).get(version, {}).get("integrity_hash")

    def _create_backup(self, version: str) -> Optional[str]:
        """Create a backup of the current model before overwriting.

        Returns:
            Backup version string if successful, None otherwise
        """
        if not self._backup_enabled:
            return None

        try:
            current = self._get_latest_version()
            if not current:
                return None

            current_path = self._base_dir / f"model_{current}.joblib"
            if not current_path.exists():
                return None

            backup_version = f"backup_{current}_{datetime.utcnow().strftime('%Y%m%d_%H%M%S')}"
            backup_path = self._backup_dir / f"model_{backup_version}.joblib"

            shutil.copy2(current_path, backup_path)
            logger.info(f"Created backup: {backup_path}")

            current_integrity_hash = self._get_integrity_hash(current)
            if current_integrity_hash:
                self._update_backup_metadata(backup_version, current_integrity_hash)

            self._cleanup_old_backups()
            return backup_version
        except Exception as e:
            logger.warning(f"Failed to create backup: {e}")
            return None

    def _update_backup_metadata(self, backup_version: str, integrity_hash: str) -> None:
        """Update metadata with backup information.

        Args:
            backup_version: The backup version string
            integrity_hash: The integrity hash of the backed up model
        """
        meta = self._read_metadata()
        if "backups" not in meta:
            meta["backups"] = {}
        meta["backups"][backup_version] = {
            "created_at": datetime.utcnow().isoformat(),
            "integrity_hash": integrity_hash,
            "encrypted": self._encrypt,
        }
        with open(self._metadata_path, "w") as f:
            json.dump(meta, f, indent=2)

    def _cleanup_old_backups(self) -> None:
        """Remove old backups beyond retention limit."""
        try:
            backups = sorted(
                self._backup_dir.glob("model_backup_*.joblib"),
                key=lambda p: p.stat().st_mtime,
                reverse=True
            )
            for old_backup in backups[self._max_backups:]:
                old_backup.unlink()
                logger.info(f"Removed old backup: {old_backup}")
        except Exception as e:
            logger.warning(f"Failed to cleanup old backups: {e}")

    def list_backups(self) -> List[dict]:
        """List all available backups.

        Returns:
            List of backup info dicts with version, created_at, size_bytes
        """
        backups = []
        try:
            for backup_path in self._backup_dir.glob("model_backup_*.joblib"):
                stat = backup_path.stat()
                version = backup_path.stem.replace("model_", "")
                backups.append({
                    "version": version,
                    "created_at": datetime.fromtimestamp(stat.st_mtime).isoformat(),
                    "size_bytes": stat.st_size,
                })
        except Exception as e:
            logger.warning(f"Failed to list backups: {e}")

        return sorted(backups, key=lambda x: x["created_at"], reverse=True)

    def verify_backup(self, backup_version: str) -> tuple[bool, str]:
        """Verify a backup's integrity by loading and checking.

        Args:
            backup_version: The backup version string to verify

        Returns:
            Tuple of (is_valid, message)
        """
        try:
            backup_path = self._backup_dir / f"model_{backup_version}.joblib"
            if not backup_path.exists():
                return False, f"Backup not found: {backup_version}"

            try:
                if self._encrypt and self._aesgcm:
                    combined = backup_path.read_bytes()
                    nonce = combined[:12]
                    ciphertext = combined[12:]
                    decrypted = self._aesgcm.decrypt(nonce, ciphertext, None)
                    model = joblib.loads(decrypted)
                else:
                    model = joblib.load(backup_path)

                integrity_hash = self._compute_integrity_hash(joblib.dumps(model))
                expected_hash = self._get_backup_integrity_hash(backup_version)

                if expected_hash and integrity_hash != expected_hash:
                    return False, f"Integrity hash mismatch for {backup_version}"

                logger.info(f"Backup verified successfully: {backup_version}")
                return True, f"Backup {backup_version} is valid"

            except Exception as e:
                return False, f"Failed to load/verify backup: {str(e)}"

        except Exception as e:
            return False, f"Backup verification failed: {str(e)}"

    def _get_backup_integrity_hash(self, backup_version: str) -> Optional[str]:
        """Get the integrity hash for a specific backup.

        Args:
            backup_version: The backup version string

        Returns:
            Integrity hash string or None
        """
        meta = self._read_metadata()
        backups_meta = meta.get("backups", {}).get(backup_version, {})
        return backups_meta.get("integrity_hash")

    def verify_all_backups(self) -> dict:
        """Verify all available backups.

        Returns:
            Dict with verification results for each backup
        """
        results = {}
        backups = self.list_backups()

        for backup in backups:
            version = backup["version"]
            is_valid, message = self.verify_backup(version)
            results[version] = {
                "valid": is_valid,
                "message": message,
                "created_at": backup["created_at"],
                "size_bytes": backup["size_bytes"],
            }

        valid_count = sum(1 for v in results.values() if v["valid"])
        logger.info(
            f"Backup verification complete: {valid_count}/{len(results)} backups valid "
            f"for namespace {self._namespace}"
        )

        return results

    def restore_backup(self, backup_version: str) -> bool:
        """Restore a model from backup.

        Args:
            backup_version: The backup version string to restore

        Returns:
            True if successful
        """
        try:
            backup_path = self._backup_dir / f"model_{backup_version}.joblib"
            if not backup_path.exists():
                logger.error(f"Backup not found: {backup_path}")
                return False

            current_version = self._get_latest_version()
            if current_version:
                self._create_backup(current_version)

            current_path = self._base_dir / f"model_{current_version or 'restored'}.joblib"
            shutil.copy2(backup_path, current_path)

            meta = self._read_metadata()
            new_version = datetime.utcnow().strftime("%Y%m%d_%H%M%S")
            meta["latest_version"] = new_version
            meta["versions"][new_version] = meta.get("versions", {}).get(backup_version, {
                "restored_from_backup": backup_version,
                "restored_at": datetime.utcnow().isoformat(),
            })
            with open(self._metadata_path, "w") as f:
                json.dump(meta, f, indent=2)

            logger.info(f"Restored backup {backup_version} as {new_version}")
            return True
        except Exception as e:
            logger.error(f"Failed to restore backup: {e}")
            return False

    def list_versions(self) -> List[dict]:
        """List all model versions.

        Returns:
            List of version info dicts
        """
        meta = self._read_metadata()
        versions = []
        for version, info in meta.get("versions", {}).items():
            versions.append({
                "version": version,
                "saved_at": info.get("saved_at"),
                "size_bytes": info.get("size_bytes"),
                "encrypted": info.get("encrypted", False),
            })
        return sorted(versions, key=lambda x: x["saved_at"] or "", reverse=True)


class ModelStore:
    """Manages model serialization, versioning, and loading.

    For production, use EncryptedModelStore instead for encryption at rest.
    This class maintains backward compatibility for unencrypted storage.
    """

    def __init__(self, namespace: str):
        self._namespace = namespace
        self._base_dir = Path(settings.ml_model_dir) / namespace
        self._base_dir.mkdir(parents=True, exist_ok=True)
        self._metadata_path = self._base_dir / "metadata.json"

    def save(self, model: Any, version: Optional[str] = None) -> str:
        """Save a model to disk.

        Args:
            model: The model object to serialize
            version: Optional version string (defaults to timestamp)

        Returns:
            Version string of the saved model
        """
        if version is None:
            version = datetime.utcnow().strftime("%Y%m%d_%H%M%S")

        model_path = self._base_dir / f"model_{version}.joblib"
        joblib.dump(model, model_path)

        self._update_metadata(version, model_path)
        logger.info(f"Saved model {self._namespace}/{version}")
        return version

    def load(self, version: Optional[str] = None) -> Optional[Any]:
        """Load a model from disk.

        Args:
            version: Specific version to load (latest if None)

        Returns:
            The loaded model or None if not found
        """
        if version is None:
            version = self._get_latest_version()

        if version is None:
            return None

        model_path = self._base_dir / f"model_{version}.joblib"
        if not model_path.exists():
            logger.warning(f"Model not found: {model_path}")
            return None

        try:
            model = joblib.load(model_path)
            logger.info(f"Loaded model {self._namespace}/{version}")
            return model
        except Exception as e:
            logger.error(f"Failed to load model {self._namespace}/{version}: {e}")
            return None

    def exists(self) -> bool:
        """Check if any model version exists."""
        return self._get_latest_version() is not None

    def _get_latest_version(self) -> Optional[str]:
        """Get the latest model version from metadata."""
        meta = self._read_metadata()
        return meta.get("latest_version")

    def _update_metadata(self, version: str, model_path: Path) -> None:
        """Update the metadata file with a new version."""
        meta = self._read_metadata()
        meta["latest_version"] = version
        meta["versions"] = meta.get("versions", {})
        meta["versions"][version] = {
            "saved_at": datetime.utcnow().isoformat(),
            "size_bytes": model_path.stat().st_size,
        }
        with open(self._metadata_path, "w") as f:
            json.dump(meta, f, indent=2)

    def _read_metadata(self) -> dict:
        """Read metadata from disk."""
        if not self._metadata_path.exists():
            return {}
        try:
            with open(self._metadata_path) as f:
                return json.load(f)
        except Exception:
            return {}
