from __future__ import annotations

from typing import Any, Dict, List, Optional


class SecretsService:
    def __init__(self, client: Any) -> None:
        self._client = client

    def create(self, **fields: Any) -> Dict[str, Any]:
        secret_type = fields.get("secret_type")
        if not secret_type:
            raise ValueError("secret_type is required")
        return self._client.request("POST", "/v1/vault/secrets", body=fields)

    def get(self, secret_id: str) -> Dict[str, Any]:
        return self._client.request("GET", f"/v1/vault/secrets/{self._client.quote(secret_id)}")

    def update(self, secret_id: str, **fields: Any) -> Dict[str, Any]:
        return self._client.request(
            "PATCH", f"/v1/vault/secrets/{self._client.quote(secret_id)}", body=fields
        )

    def rotate(self, secret_id: str, **fields: Any) -> Dict[str, Any]:
        return self._client.request(
            "PATCH", f"/v1/vault/secrets/{self._client.quote(secret_id)}/rotate", body=fields
        )

    def delete(self, secret_id: str) -> None:
        self._client.request("DELETE", f"/v1/vault/secrets/{self._client.quote(secret_id)}")

    def list(self, limit: int = 50, offset: int = 0, secret_type: Optional[str] = None) -> Dict[str, Any]:
        qs: List[str] = [f"limit={limit}", f"offset={offset}"]
        if secret_type:
            qs.append(f"secret_type={secret_type}")
        return self._client.request("GET", "/v1/vault/secrets?" + "&".join(qs))
