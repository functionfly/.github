from __future__ import annotations

from typing import Any, Dict, List, Optional


class TokensService:
    def __init__(self, client: Any) -> None:
        self._client = client

    def create(self, secret_id: str, expires_in_hours: int = 24, scopes: Optional[List[str]] = None, name: Optional[str] = None) -> Dict[str, Any]:
        if not secret_id:
            raise ValueError("secret_id is required")
        body: Dict[str, Any] = {
            "secret_id": secret_id,
            "expires_in_hours": expires_in_hours,
        }
        if scopes is not None:
            body["scopes"] = scopes
        if name:
            body["name"] = name
        return self._client.request(
            "POST", f"/v1/vault/secrets/{self._client.quote(secret_id)}/tokens", body=body
        )

    def list(self, secret_id: str) -> Dict[str, Any]:
        return self._client.request(
            "GET", f"/v1/vault/secrets/{self._client.quote(secret_id)}/tokens"
        )

    def revoke(self, token_id: str) -> None:
        self._client.request("DELETE", f"/v1/vault/tokens/{self._client.quote(token_id)}")
