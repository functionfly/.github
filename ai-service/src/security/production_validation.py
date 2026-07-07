"""Production readiness validation for FlyMind AI Service.

This module enforces security and operational requirements for production deployment.
It validates configuration at startup and blocks the service if critical requirements
are not met.
"""

import logging
import os
import re
from typing import Optional

from ..config import settings

logger = logging.getLogger(__name__)


class ProductionValidationError(Exception):
    """Raised when production validation fails."""
    pass


class ProductionValidator:
    """Validates production readiness requirements.

    Runs comprehensive checks on configuration to ensure the service
    meets security and operational standards for production deployment.
    """

    CLOUD_REDIS_PATTERNS = [
        r"\.upstash\.io$",
        r"\.redis\.cloud$",
        r"\.redislabs\.com$",
        r"\.aws\.amazon\.com/elasticache",
        r"\.gcp\.googleapis\.com/memstore",
        r"\.azure\.com/cache",
        r"clustered\.redis\.com$",
    ]

    def __init__(self):
        self._errors: list[str] = []
        self._warnings: list[str] = []

    def validate(self) -> bool:
        """Run all production validation checks.

        Returns:
            True if all checks pass, False otherwise.

        Raises:
            ProductionValidationError: If critical validation fails and
                ml_require_encryption_in_production is True.
        """
        self._errors.clear()
        self._warnings.clear()

        self._validate_encryption()
        self._validate_redis_tls()
        self._validate_redis_security()
        self._validate_auth_security()
        self._validate_fallback_security()
        self._validate_ml_settings()

        if self._errors:
            logger.error("Production validation failed with errors:")
            for error in self._errors:
                logger.error(f"  - {error}")

        if self._warnings:
            logger.warning("Production validation warnings:")
            for warning in self._warnings:
                logger.warning(f"  - {warning}")

        if self._errors:
            if self._is_production():
                require_encryption = getattr(settings, 'ml_require_encryption_in_production', True)
                if require_encryption:
                    raise ProductionValidationError(
                        f"Production validation failed with {len(self._errors)} error(s). "
                        f"Service cannot start in production mode without fixing these issues."
                    )
            return False

        return True

    def _is_production(self) -> bool:
        """Check if running in production mode."""
        env = os.getenv("ENVIRONMENT", "").lower()
        return env == "production"

    def _validate_encryption(self) -> None:
        """Validate that encryption is properly configured in production."""
        encryption_key = getattr(settings, 'ml_model_encryption_key', None)

        if not encryption_key:
            self._errors.append(
                "ML model encryption key (ML_MODEL_ENCRYPTION_KEY) is not set. "
                "Models will not be encrypted at rest, which is not allowed in production."
            )
        elif not self._is_valid_fernet_key(encryption_key):
            self._errors.append(
                "ML model encryption key (ML_MODEL_ENCRYPTION_KEY) is not a valid Fernet key. "
                "Generate a new key with: python -c \"from cryptography.fernet import Fernet; print(Fernet.generate_key().decode())\""
            )

        cache_encryption_key = getattr(settings, 'redis_cache_encryption_key', None)
        if not cache_encryption_key:
            self._warnings.append(
                "Cache encryption key (REDIS_CACHE_ENCRYPTION_KEY) is not set. "
                "Cache data will not be encrypted."
            )
        elif not self._is_valid_fernet_key(cache_encryption_key):
            self._errors.append(
                "Cache encryption key (REDIS_CACHE_ENCRYPTION_KEY) is not a valid Fernet key."
            )

    def _is_valid_fernet_key(self, key: str) -> bool:
        """Check if a string is a valid Fernet key."""
        if not key:
            return False
        if isinstance(key, bytes):
            key = key.decode('utf-8', errors='ignore')
        if len(key) != 44:
            return False
        return re.match(r'^[A-Za-z0-9_]+$', key) is not None

    def _validate_redis_tls(self) -> None:
        """Validate Redis TLS configuration for cloud Redis."""
        redis_url = getattr(settings, 'redis_url', 'redis://localhost:6379')
        redis_use_tls = getattr(settings, 'redis_use_tls', False)

        if self._is_cloud_redis(redis_url):
            if not redis_use_tls:
                self._errors.append(
                    f"Redis URL '{redis_url}' appears to be a cloud Redis instance "
                    f"but TLS is not enabled (REDIS_USE_TLS=false). "
                    f"TLS is required for cloud Redis connections."
                )
        elif not redis_use_tls and self._is_production():
            self._warnings.append(
                "Redis TLS is not enabled. Enable REDIS_USE_TLS=true for production."
            )

    def _is_cloud_redis(self, redis_url: str) -> bool:
        """Check if Redis URL appears to be a cloud-hosted instance."""
        redis_url_lower = redis_url.lower()
        for pattern in self.CLOUD_REDIS_PATTERNS:
            if re.search(pattern, redis_url_lower):
                return True
        return False

    def _validate_redis_security(self) -> None:
        """Validate Redis security settings."""
        redis_password = getattr(settings, 'redis_password', None)
        redis_url = getattr(settings, 'redis_url', 'redis://localhost:6379')

        if 'localhost' in redis_url or '127.0.0.1' in redis_url:
            if self._is_production():
                self._warnings.append(
                    "Redis is configured to use localhost. "
                    "For production, use a managed Redis service with authentication."
                )
        else:
            if not redis_password:
                self._warnings.append(
                    "Redis password (REDIS_PASSWORD) is not set. "
                    "Managed Redis services should require authentication."
                )

    def _validate_auth_security(self) -> None:
        """Validate authentication security settings."""
        reject_auth = getattr(settings, 'reject_auth_in_degraded_mode', True)
        allow_cached_auth = getattr(settings, 'allow_cached_auth_in_degraded_mode', True)

        if self._is_production():
            if not reject_auth:
                self._errors.append(
                    "reject_auth_in_degraded_mode is False. "
                    "In production, authentication must be rejected when orchestrator is unreachable."
                )

            if allow_cached_auth and reject_auth:
                self._warnings.append(
                    "allow_cached_auth_in_degraded_mode is True. "
                    "Cached keys will be accepted during degraded mode. "
                    "Consider setting to False for higher security."
                )

        orchestrator_url = getattr(settings, 'orchestrator_url', 'http://localhost:8080')
        orchestrator_api_key = getattr(settings, 'orchestrator_api_key', None)

        if 'localhost' in orchestrator_url and self._is_production():
            self._warnings.append(
                "Orchestrator URL uses localhost. "
                "For production, use internal DNS or managed orchestrator service."
            )

        if not orchestrator_api_key:
            self._warnings.append(
                "Orchestrator API key (ORCHESTRATOR_API_KEY) is not set. "
                "Service-to-service authentication may be incomplete."
            )

    def _validate_fallback_security(self) -> None:
        """Validate fallback storage security settings."""
        fallback_to_file = getattr(settings, 'ml_fallback_to_file', True)
        fallback_dir = getattr(settings, 'ml_fallback_file_dir', '/var/lib/flymind/ml-state')

        if fallback_to_file:
            if not self._is_production():
                self._warnings.append(
                    "File-based fallback is enabled. "
                    "Ensure proper permissions are set on fallback directory."
                )

            if fallback_dir.startswith('/tmp') or fallback_dir.startswith('/var/tmp'):
                if self._is_production():
                    self._warnings.append(
                        f"Fallback directory '{fallback_dir}' uses temp storage. "
                        "Data may be lost on system restart. Use persistent storage in production."
                    )

            cache_encryption_key = getattr(settings, 'redis_cache_encryption_key', None)
            if not cache_encryption_key:
                self._warnings.append(
                    "File-based fallback is enabled but cache encryption is not configured. "
                    "Fallback data will be stored in plaintext."
                )

    def _validate_ml_settings(self) -> None:
        """Validate ML-specific settings."""
        ml_enabled = getattr(settings, 'ml_enabled', True)

        if not ml_enabled:
            self._warnings.append(
                "ML services are disabled (ML_ENABLED=false). "
                "ML Intelligence Layer will not be available."
            )

        model_dir = getattr(settings, 'ml_model_dir', '/var/lib/flymind/models')
        if model_dir.startswith('/tmp') or model_dir.startswith('/var/tmp'):
            if self._is_production():
                self._warnings.append(
                    f"Model directory '{model_dir}' uses temp storage. "
                    "Models will be lost on restart. Use persistent storage in production."
                )

        drift_threshold = getattr(settings, 'ml_drift_threshold', 0.15)
        if drift_threshold <= 0 or drift_threshold >= 1:
            self._errors.append(
                f"ML drift threshold ({drift_threshold}) is invalid. "
                "Must be between 0.0 and 1.0."
            )

    def get_validation_report(self) -> dict:
        """Get a detailed validation report.

        Returns:
            Dict with errors, warnings, and validation status.
        """
        return {
            "is_production": self._is_production(),
            "errors": self._errors.copy(),
            "warnings": self._warnings.copy(),
            "passed": len(self._errors) == 0,
        }


_validator: Optional[ProductionValidator] = None


def get_production_validator() -> ProductionValidator:
    """Get the global production validator instance."""
    global _validator
    if _validator is None:
        _validator = ProductionValidator()
    return _validator


def validate_production_config() -> bool:
    """Run production validation and return result.

    This function is called at startup to ensure production requirements are met.

    Returns:
        True if validation passes, False otherwise.

    Raises:
        ProductionValidationError: If production mode is enabled and validation fails.
    """
    validator = get_production_validator()
    return validator.validate()


def get_validation_report() -> dict:
    """Get the current validation report without re-running checks.

    Returns:
        Dict with errors, warnings, and validation status.
    """
    validator = get_production_validator()
    return validator.get_validation_report()
