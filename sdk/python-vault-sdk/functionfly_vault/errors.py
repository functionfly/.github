"""VaultAPIError is the structured error raised when the FunctionFly
vault API returns a non-2xx response or a network error occurs."""

from __future__ import annotations

from typing import Any, Dict, Optional


class VaultAPIError(Exception):
    """Raised when the vault API returns a non-2xx status or a network
    error occurs. Attributes mirror the structured API error response.
    """

    def __init__(
        self,
        status: int,
        code: str,
        message: str,
        details: Optional[Dict[str, Any]] = None,
    ) -> None:
        super().__init__(f"{code}: {message}")
        self.status = status
        self.code = code
        self.message = message
        self.details = details or {}

    def __repr__(self) -> str:
        return f"VaultAPIError(status={self.status}, code={self.code!r}, message={self.message!r})"
