from __future__ import annotations

from typing import Any, Dict, List, Optional
from urllib.parse import urlencode


class AuditService:
    def __init__(self, client: Any) -> None:
        self._client = client

    def list(
        self,
        secret_id: Optional[str] = None,
        action: Optional[str] = None,
        actor_id: Optional[str] = None,
        limit: int = 50,
        offset: int = 0,
    ) -> Dict[str, Any]:
        params: Dict[str, Any] = {"limit": limit, "offset": offset}
        if secret_id is not None:
            params["secret_id"] = secret_id
        if action is not None:
            params["action"] = action
        if actor_id is not None:
            params["actor_id"] = actor_id
        return self._client.request("GET", "/v1/vault/audit?" + urlencode(params))
