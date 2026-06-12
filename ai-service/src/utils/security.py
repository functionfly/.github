"""Security utilities for FlyMind AI Service."""

import re
import logging
from typing import Optional

logger = logging.getLogger(__name__)


SAFE_ERROR_MESSAGE = "An internal error occurred. Please try again later."


def sanitize_error_message(error: Exception, include_details: bool = False) -> str:
    """Sanitize error messages for client responses.

    Prevents information leakage through error messages while
    still providing useful feedback in non-production environments.

    Args:
        error: The exception that occurred
        include_details: If True, include error details (only in development)

    Returns:
        A safe error message suitable for client responses
    """
    if include_details:
        error_str = str(error)
        if len(error_str) > 200:
            error_str = error_str[:200] + "..."
        return error_str
    return SAFE_ERROR_MESSAGE


def sanitize_filename(filename: str) -> str:
    """Sanitize a filename to prevent path traversal.

    Args:
        filename: The filename to sanitize

    Returns:
        A safe filename
    """
    filename = re.sub(r'[^a-zA-Z0-9._-]', '_', filename)
    if len(filename) > 255:
        filename = filename[:255]
    return filename


def sanitize_id(id_value: str, max_length: int = 64) -> str:
    """Sanitize an ID value to prevent injection attacks.

    Args:
        id_value: The ID to sanitize
        max_length: Maximum allowed length

    Returns:
        A safe ID
    """
    if not id_value:
        return ""
    sanitized = re.sub(r'[^a-zA-Z0-9_\-]', '', id_value)
    return sanitized[:max_length]


def is_valid_uuid(value: str) -> bool:
    """Check if a string is a valid UUID format.

    Args:
        value: The string to check

    Returns:
        True if valid UUID format
    """
    uuid_pattern = re.compile(
        r'^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$',
        re.IGNORECASE
    )
    return bool(uuid_pattern.match(value))


def is_valid_tenant_id(value: str) -> bool:
    """Check if a tenant ID has a valid format.

    Args:
        value: The tenant ID to check

    Returns:
        True if valid format
    """
    if not value or len(value) > 64:
        return False
    return bool(re.match(r'^[a-zA-Z0-9_\-]+$', value))