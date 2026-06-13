"""VaultClient is the entry point for the FunctionFly vault Python SDK.

The SDK targets Python 3.9+ and uses only the standard library
(``urllib``) so it has no third-party runtime dependencies. The
caller is responsible for client-side encryption — the SDK ships the
zero-knowledge ciphertext to the API verbatim.
"""

from __future__ import annotations

import json
import urllib.error
import urllib.parse
import urllib.request
from typing import Any, Dict, Optional

from .errors import VaultAPIError
from .services import (
    AuditService,
    DynamicCredentialsService,
    DynamicTargetsService,
    LeasesService,
    SecretsService,
    TokensService,
)

__all__ = ["VaultClient"]


DEFAULT_BASE_URL = "https://api.functionfly.com"
SDK_VERSION = "0.1.0"
DEFAULT_TIMEOUT = 30.0


class VaultClient:
    """A thin HTTP wrapper around the FunctionFly vault REST API."""

    def __init__(
        self,
        token: str,
        base_url: str = DEFAULT_BASE_URL,
        timeout: float = DEFAULT_TIMEOUT,
        user_agent: Optional[str] = None,
    ) -> None:
        if not token:
            raise ValueError("token is required")
        self.token = token
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self.user_agent = user_agent or f"functionfly-vault-python/{SDK_VERSION}"

        self.secrets = SecretsService(self)
        self.tokens = TokensService(self)
        self.dynamic_credentials = DynamicCredentialsService(self)
        self.dynamic_targets = DynamicTargetsService(self)
        self.leases = LeasesService(self)
        self.audit = AuditService(self)

    def request(
        self,
        method: str,
        path: str,
        body: Optional[Dict[str, Any]] = None,
    ) -> Any:
        """Perform an HTTP request and return the decoded JSON body.

        Raises :class:`VaultAPIError` on non-2xx responses or network
        errors.
        """
        url = self.base_url + path
        data: Optional[bytes] = None
        headers = {
            "Accept": "application/json",
            "Authorization": f"Bearer {self.token}",
            "User-Agent": self.user_agent,
        }
        if body is not None:
            data = json.dumps(body).encode("utf-8")
            headers["Content-Type"] = "application/json"
        req = urllib.request.Request(url, data=data, method=method, headers=headers)
        try:
            with urllib.request.urlopen(req, timeout=self.timeout) as resp:
                payload = resp.read()
        except urllib.error.HTTPError as exc:
            raw = exc.read().decode("utf-8", errors="replace")
            try:
                err_body = json.loads(raw) if raw else {}
            except json.JSONDecodeError:
                err_body = {}
            message = (
                err_body.get("message")
                or err_body.get("error")
                or f"HTTP {exc.code}"
            )
            code = err_body.get("code") or "HTTP_ERROR"
            raise VaultAPIError(
                status=exc.code,
                code=code,
                message=message,
                details=err_body,
            ) from exc
        except urllib.error.URLError as exc:
            raise VaultAPIError(
                status=0,
                code="network_error",
                message=str(exc.reason),
            ) from exc

        if not payload:
            return None
        try:
            return json.loads(payload.decode("utf-8"))
        except json.JSONDecodeError as exc:
            raise VaultAPIError(
                status=200,
                code="decode_error",
                message=f"could not parse response: {exc}",
            ) from exc

    @staticmethod
    def quote(s: str) -> str:
        return urllib.parse.quote(s, safe="")
