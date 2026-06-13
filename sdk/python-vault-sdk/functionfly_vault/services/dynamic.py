from __future__ import annotations

from typing import Any, Dict, Optional


class DynamicTargetsService:
    def __init__(self, client: Any) -> None:
        self._client = client

    def create(self, **fields: Any) -> Dict[str, Any]:
        if not fields.get("admin_password"):
            raise ValueError("admin_password is required")
        return self._client.request("POST", "/v1/vault/dynamic-secret-targets", body=fields)

    def list(self) -> Dict[str, Any]:
        return self._client.request("GET", "/v1/vault/dynamic-secret-targets")

    def delete(self, target_id: str) -> None:
        self._client.request(
            "DELETE", f"/v1/vault/dynamic-secret-targets/{self._client.quote(target_id)}"
        )

    def test(self, target_id: str) -> None:
        self._client.request(
            "POST", f"/v1/vault/dynamic-secret-targets/{self._client.quote(target_id)}/test"
        )


class DynamicCredentialsService:
    def __init__(self, client: Any) -> None:
        self._client = client

    def create(self, **fields: Any) -> Dict[str, Any]:
        if not fields.get("target_id"):
            raise ValueError("target_id is required")
        return self._client.request("POST", "/v1/vault/dynamic-credentials", body=fields)

    def generate(self, credential_id: str, ttl_seconds: Optional[int] = None) -> Dict[str, Any]:
        body: Optional[Dict[str, Any]] = None
        if ttl_seconds is not None:
            body = {"ttl_seconds": ttl_seconds}
        return self._client.request(
            "POST",
            f"/v1/vault/dynamic-credentials/{self._client.quote(credential_id)}/generate",
            body=body,
        )

    def revoke_all(self, credential_id: str) -> None:
        self._client.request(
            "POST", f"/v1/vault/dynamic-credentials/{self._client.quote(credential_id)}/revoke"
        )


class LeasesService:
    def __init__(self, client: Any) -> None:
        self._client = client

    def renew(self, credential_id: str, lease_id: str, ttl_seconds: Optional[int] = None) -> Dict[str, Any]:
        body: Optional[Dict[str, Any]] = None
        if ttl_seconds is not None:
            body = {"ttl_seconds": ttl_seconds}
        return self._client.request(
            "POST",
            f"/v1/vault/dynamic-credentials/{self._client.quote(credential_id)}/leases/{self._client.quote(lease_id)}/renew",
            body=body,
        )

    def revoke(self, credential_id: str, lease_id: str) -> None:
        self._client.request(
            "POST",
            f"/v1/vault/dynamic-credentials/{self._client.quote(credential_id)}/leases/{self._client.quote(lease_id)}/revoke",
        )
