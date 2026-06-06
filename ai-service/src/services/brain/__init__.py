"""Brain signal client for FlyMind.
Communicates with the Go orchestrator's brain API to fetch and manage signals."""

import logging
from dataclasses import dataclass, field
from typing import Optional

import httpx

logger = logging.getLogger(__name__)

_brain_client: Optional["BrainClient"] = None


def get_brain_client() -> "BrainClient":
    global _brain_client
    if _brain_client is None:
        _brain_client = BrainClient()
    return _brain_client


@dataclass
class BrainSignal:
    id: str
    tenant_id: str
    connector_slug: str
    signal_type: str
    entity_id: str
    entity_name: str
    fact: str
    importance: int
    source_url: str = ""
    created_at: str = ""
    last_seen_at: str = ""


@dataclass
class BrainContext:
    signals: list[BrainSignal] = field(default_factory=list)
    connector_ids: list[str] = field(default_factory=list)
    memory_used: int = 0
    score_summary: str = ""


class BrainClient:
    """HTTP client for the Go orchestrator's brain API."""

    def __init__(self, base_url: Optional[str] = None):
        from src.config import get_settings
        settings = get_settings()
        self.base_url = base_url or settings.orchestrator_url or "http://localhost:8080"
        self._client: Optional[httpx.AsyncClient] = None

    async def _get_client(self) -> httpx.AsyncClient:
        if self._client is None or self._client.is_closed:
            self._client = httpx.AsyncClient(
                base_url=self.base_url,
                timeout=httpx.Timeout(10.0),
                headers={"Content-Type": "application/json"},
            )
        return self._client

    async def get_brain_context(
        self,
        tenant_id: str,
        query: str,
        intent: str = "general",
        max_hints: int = 10,
        auth_token: Optional[str] = None,
    ) -> Optional[BrainContext]:
        """Fetch brain context signals for a tenant and query."""
        try:
            client = await self._get_client()
            headers = {}
            if auth_token:
                headers["Authorization"] = f"Bearer {auth_token}"

            resp = await client.get(
                "/v1/brain/signals",
                params={"limit": max_hints, "sort": "importance"},
                headers=headers,
            )
            if resp.status_code != 200:
                logger.warning(f"Brain API returned {resp.status_code}")
                return None

            data = resp.json()
            signals = [
                BrainSignal(
                    id=s.get("id", ""),
                    tenant_id=s.get("tenant_id", ""),
                    connector_slug=s.get("connector_slug", ""),
                    signal_type=s.get("signal_type", ""),
                    entity_id=s.get("entity_id", ""),
                    entity_name=s.get("entity_name", ""),
                    fact=s.get("fact", ""),
                    importance=s.get("importance", 1),
                    source_url=s.get("source_url", ""),
                    created_at=s.get("created_at", ""),
                    last_seen_at=s.get("last_seen_at", ""),
                )
                for s in data.get("signals", [])
            ]

            connector_ids = list({s.connector_slug for s in signals})
            facts = [f"- [{s.connector_slug}] {s.fact}" for s in signals]

            return BrainContext(
                signals=signals,
                connector_ids=connector_ids,
                memory_used=data.get("total", len(signals)),
                score_summary="\n".join(facts),
            )
        except Exception as e:
            logger.warning(f"Failed to fetch brain context: {e}")
            return None

    async def get_signals_for_agent(
        self,
        tenant_id: str,
        agent_id: Optional[str] = None,
        max_hints: int = 10,
        auth_token: Optional[str] = None,
    ) -> list[BrainSignal]:
        """Fetch signals scoped to an agent or tenant."""
        ctx = await self.get_brain_context(
            tenant_id=tenant_id,
            query="",
            max_hints=max_hints,
            auth_token=auth_token,
        )
        return ctx.signals if ctx else []

    async def close(self):
        if self._client and not self._client.is_closed:
            await self._client.aclose()
