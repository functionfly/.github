"""Production readiness validation for FlyMind ML services.

Validates that critical security and operational settings are properly configured
before the service starts in production.
"""

import logging
import os
from typing import List, Optional

from ..config import settings

logger = logging.getLogger(__name__)


class ValidationError(Exception):
    """Raised when validation fails."""
    pass


class ProductionValidator:
    """Validates production readiness of ML services."""

    def __init__(self):
        self._errors: List[str] = []
        self._warnings: List[str] = []

    def validate(self) -> bool:
        """Run all validations.

        Returns:
            True if all validations pass

        Raises:
            ValidationError: If critical validations fail in production
        """
        self._errors.clear()
        self._warnings.clear()

        self._validate_environment()
        self._validate_encryption()
        self._validate_redis()
        self._validate_directories()

        if self._errors:
            logger.error("Production validation failed:")
            for error in self._errors:
                logger.error(f"  - {error}")

        if self._warnings:
            logger.warning("Production validation warnings:")
            for warning in self._warnings:
                logger.warning(f"  - {warning}")

        if self._errors:
            is_prod = self._is_production_environment()
            if is_prod:
                raise ValidationError(
                    f"Production validation failed with {len(self._errors)} error(s): "
                    + "; ".join(self._errors)
                )
            else:
                logger.warning(
                    f"Validation errors found but not in production mode: {'; '.join(self._errors)}"
                )
                return True

        return True

    def _is_production_environment(self) -> bool:
        """Determine if we're running in a production environment.

        Checks multiple signals to determine if this is production:
        1. ENVIRONMENT=production env var is set
        2. ENVIRONMENT is not "development" or "test" AND other production signals present
        3. Redis URL points to a cloud provider (not localhost)

        Returns:
            True if production environment detected
        """
        environment = os.getenv("ENVIRONMENT", "").lower()

        if environment == "production":
            return True

        if environment in ("development", "test", ""):
            return False

        redis_url = getattr(settings, "redis_url", "") or ""

        cloud_redis_indicators = [
            "rediscloud.com",
            "redislabs.com",
            "aws.elasticache",
            "aws.redis",
            "dataplane.redis",
            ".cache.amazonaws.com",
        ]

        if any(indicator in redis_url.lower() for indicator in cloud_redis_indicators):
            return True

        return False

    def _validate_environment(self) -> None:
        """Validate environment settings."""
        environment = os.getenv("ENVIRONMENT", "").lower()
        is_prod = self._is_production_environment()

        if environment == "production":
            logger.info("Running in PRODUCTION mode - strict validation enabled")
        elif environment == "":
            if is_prod:
                logger.info("Production environment detected - strict validation enabled")
            else:
                self._warnings.append(
                    "ENVIRONMENT not set - defaulting to development mode validation"
                )
        else:
            if is_prod:
                logger.info(f"Production environment detected ({environment}) - strict validation enabled")
            else:
                self._warnings.append(f"Running in {environment} mode")

    def _validate_encryption(self) -> None:
        """Validate encryption configuration."""
        is_prod = self._is_production_environment()
        encryption_key = getattr(settings, "ml_model_encryption_key", None)
        require_encryption = getattr(settings, "ml_require_encryption_in_production", True)

        if is_prod and require_encryption:
            if not encryption_key:
                self._errors.append(
                    "ml_model_encryption_key is required in production - "
                    "generate with: python -c \"from cryptography.fernet import Fernet; print(Fernet.generate_key().decode())\""
                )
            elif len(encryption_key) < 32:
                self._errors.append(
                    "ml_model_encryption_key must be at least 32 bytes when decoded"
                )
        elif not encryption_key:
            self._warnings.append(
                "ml_model_encryption_key not set - models will not be encrypted at rest"
            )

    def _validate_redis(self) -> None:
        """Validate Redis configuration."""
        redis_url = getattr(settings, "redis_url", None)
        redis_use_tls = getattr(settings, "redis_use_tls", False)
        environment = os.getenv("ENVIRONMENT", "").lower()

        if not redis_url:
            self._errors.append("redis_url is required")
            return

        if environment == "production" and not redis_use_tls:
            is_localhost = "localhost" in redis_url or "127.0.0.1" in redis_url
            if not is_localhost:
                self._errors.append(
                    "redis_use_tls is False - Redis traffic is not encrypted in production"
                )
            else:
                self._warnings.append(
                    "Redis TLS disabled but localhost is used - not network-exposed; this is acceptable for single-node deployments"
                )

        if "localhost" in redis_url or "127.0.0.1" in redis_url:
            if environment == "production":
                self._errors.append(
                    "Redis URL points to localhost - not suitable for production"
                )
            else:
                self._warnings.append(
                    "Redis URL points to localhost - not suitable for production deployment"
                )

    def _validate_directories(self) -> None:
        """Validate required directories."""
        from pathlib import Path

        model_dir = Path(getattr(settings, "ml_model_dir", "/var/lib/flymind/models"))
        backup_dir = Path(getattr(settings, "ml_model_backup_dir", "/var/lib/flymind/backups"))
        fallback_dir = Path(getattr(settings, "ml_fallback_file_dir", "/var/lib/flymind/ml-state"))
        environment = os.getenv("ENVIRONMENT", "").lower()

        for directory, name in [
            (model_dir, "ml_model_dir"),
            (backup_dir, "ml_model_backup_dir"),
            (fallback_dir, "ml_fallback_file_dir"),
        ]:
            if environment == "production":
                if not directory.exists():
                    try:
                        directory.mkdir(parents=True, exist_ok=True)
                        logger.info(f"Created directory {directory} for {name}")
                    except Exception as e:
                        self._errors.append(
                            f"Cannot create {name} directory {directory}: {e}"
                        )
                elif not os.access(directory, os.W_OK):
                    self._errors.append(f"{name} directory {directory} is not writable")
            else:
                if not directory.exists():
                    try:
                        directory.mkdir(parents=True, exist_ok=True)
                    except Exception:
                        pass

    def get_report(self) -> dict:
        """Get validation report.

        Returns:
            Dict with errors, warnings, and validation status
        """
        return {
            "passed": len(self._errors) == 0,
            "errors": self._errors.copy(),
            "warnings": self._warnings.copy(),
        }


_validator: Optional[ProductionValidator] = None


def get_production_validator() -> ProductionValidator:
    """Get the global production validator."""
    global _validator
    if _validator is None:
        _validator = ProductionValidator()
    return _validator


def validate_production_readiness() -> bool:
    """Run production validation.

    Returns:
        True if validation passes

    Raises:
        ValidationError: In production if validation fails
    """
    validator = get_production_validator()
    return validator.validate()
