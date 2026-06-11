"""Atlas gRPC client for high-performance event ingestion.

This module provides a gRPC client for communicating with Atlas Memory Engine.
Currently uses subprocess calls to the Atlas CLI since the Python SDK v0.2
with PyO3 bindings is not yet released.
"""

import asyncio
import json
import logging
import subprocess
import time
from typing import Any, AsyncGenerator, Dict, List, Optional

logger = logging.getLogger(__name__)


class AtlasGRPCClient:
    """High-performance gRPC client for Atlas event ingestion.

    Note: Currently uses subprocess calls to the Atlas CLI.
    Will be updated to use PyO3 bindings when Atlas SDK v0.2 is released.
    """

    def __init__(
        self,
        host: str = "localhost",
        port: int = 50051,
        api_key: str = "",
        tenant_id: str = "default",
        base_url: str = "http://localhost:7447",
    ):
        self.host = host
        self.port = port
        self.api_key = api_key
        self.tenant_id = tenant_id
        self.base_url = base_url
        self._use_cli = True

    async def connect(self) -> bool:
        """Initialize connection to Atlas.

        Returns:
            True if connection successful
        """
        try:
            result = subprocess.run(
                ["atlas", "version"],
                capture_output=True,
                text=True,
                timeout=5,
            )
            if result.returncode == 0:
                logger.info(f"Atlas CLI connected: {result.stdout.strip()}")
                return True
            return False
        except (subprocess.SubprocessError, FileNotFoundError):
            logger.warning("Atlas CLI not found, will use HTTP fallback")
            self._use_cli = False
            return True

    async def create_run(self, metadata: Dict[str, str]) -> str:
        """Create Atlas run, return run_id (ULID).

        Args:
            metadata: Run metadata

        Returns:
            Atlas run ID (ULID)
        """
        if self._use_cli:
            return await self._create_run_cli(metadata)
        return await self._create_run_http(metadata)

    async def _create_run_cli(self, metadata: Dict[str, str]) -> str:
        """Create run using Atlas CLI."""
        meta_json = json.dumps(metadata)
        try:
            result = subprocess.run(
                ["atlas", "run", "create", "--json"],
                input=meta_json,
                capture_output=True,
                text=True,
                timeout=10,
            )
            if result.returncode == 0:
                data = json.loads(result.stdout)
                return data.get("run_id", "")
            logger.error(f"Atlas CLI create_run failed: {result.stderr}")
            return ""
        except subprocess.SubprocessError as e:
            logger.error(f"Atlas CLI error: {e}")
            return ""

    async def _create_run_http(self, metadata: Dict[str, str]) -> str:
        """Create run using HTTP API."""
        import httpx
        try:
            async with httpx.AsyncClient() as client:
                response = await client.post(
                    f"{self.base_url}/v1/runs",
                    json={"metadata": metadata},
                    headers={
                        "Authorization": f"Bearer {self.api_key}",
                        "X-Atlas-Tenant": self.tenant_id,
                    },
                    timeout=10,
                )
                if response.status_code == 201:
                    data = response.json()
                    return data.get("run_id", "")
                return ""
        except Exception as e:
            logger.error(f"Atlas HTTP create_run failed: {e}")
            return ""

    async def append_event(
        self,
        run_id: str,
        kind: str,
        payload: Dict[str, Any],
        system_id: str,
        parent_id: str = "",
    ) -> Optional[Dict[str, Any]]:
        """Append event to run via gRPC/CLI.

        Args:
            run_id: Atlas run ID
            kind: Event kind (INPUT, DECISION, ACTION, RESULT, ERROR)
            payload: Event payload
            system_id: System identifier
            parent_id: Parent event ID

        Returns:
            Event data or None
        """
        if self._use_cli:
            return await self._append_event_cli(run_id, kind, payload, system_id, parent_id)
        return await self._append_event_http(run_id, kind, payload, system_id, parent_id)

    async def _append_event_cli(
        self,
        run_id: str,
        kind: str,
        payload: Dict[str, Any],
        system_id: str,
        parent_id: str,
    ) -> Optional[Dict[str, Any]]:
        """Append event using Atlas CLI."""
        event_data = {
            "run_id": run_id,
            "kind": kind,
            "payload": payload,
            "system_id": system_id,
            "parent": parent_id,
            "timestamp_ns": int(time.time() * 1e9),
        }
        try:
            result = subprocess.run(
                ["atlas", "event", "append", "--json"],
                input=json.dumps(event_data),
                capture_output=True,
                text=True,
                timeout=5,
            )
            if result.returncode == 0:
                return json.loads(result.stdout)
            logger.error(f"Atlas CLI append_event failed: {result.stderr}")
            return None
        except subprocess.SubprocessError as e:
            logger.error(f"Atlas CLI error: {e}")
            return None

    async def _append_event_http(
        self,
        run_id: str,
        kind: str,
        payload: Dict[str, Any],
        system_id: str,
        parent_id: str,
    ) -> Optional[Dict[str, Any]]:
        """Append event using HTTP API."""
        import httpx
        event_data = {
            "kind": kind,
            "payload": payload,
            "system_id": system_id,
            "parent": parent_id,
            "timestamp_ns": int(time.time() * 1e9),
        }
        try:
            async with httpx.AsyncClient() as client:
                response = await client.post(
                    f"{self.base_url}/v1/runs/{run_id}/events",
                    json=event_data,
                    headers={
                        "Authorization": f"Bearer {self.api_key}",
                        "X-Atlas-Tenant": self.tenant_id,
                    },
                    timeout=10,
                )
                if response.status_code == 201:
                    return response.json()
                return None
        except Exception as e:
            logger.error(f"Atlas HTTP append_event failed: {e}")
            return None

    async def replay(
        self,
        run_id: str,
        after_sequence: int = 0,
        limit: int = 1000,
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """Stream events from run.

        Args:
            run_id: Atlas run ID
            after_sequence: Start after this sequence
            limit: Maximum events to return

        Yields:
            Event dictionaries
        """
        if self._use_cli:
            async for event in self._replay_cli(run_id, after_sequence, limit):
                yield event
        else:
            async for event in self._replay_http(run_id, after_sequence, limit):
                yield event

    async def _replay_cli(
        self,
        run_id: str,
        after_sequence: int,
        limit: int,
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """Replay using Atlas CLI."""
        try:
            result = subprocess.run(
                ["atlas", "run", "replay", run_id, "--after", str(after_sequence), "--limit", str(limit), "--json"],
                capture_output=True,
                text=True,
                timeout=30,
            )
            if result.returncode == 0:
                for line in result.stdout.splitlines():
                    if line.strip():
                        try:
                            yield json.loads(line)
                        except json.JSONDecodeError:
                            continue
        except subprocess.SubprocessError as e:
            logger.error(f"Atlas CLI replay error: {e}")

    async def _replay_http(
        self,
        run_id: str,
        after_sequence: int,
        limit: int,
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """Replay using HTTP API."""
        import httpx
        try:
            async with httpx.AsyncClient() as client:
                response = await client.get(
                    f"{self.base_url}/v1/runs/{run_id}/replay",
                    params={"after_sequence": after_sequence, "limit": limit},
                    headers={
                        "Authorization": f"Bearer {self.api_key}",
                        "X-Atlas-Tenant": self.tenant_id,
                    },
                    timeout=30,
                )
                if response.status_code == 200:
                    data = response.json()
                    for event in data.get("events", []):
                        yield event
        except Exception as e:
            logger.error(f"Atlas HTTP replay failed: {e}")

    async def get_stats(self, run_id: str) -> Optional[Dict[str, Any]]:
        """Get run statistics.

        Args:
            run_id: Atlas run ID

        Returns:
            Run statistics or None
        """
        if self._use_cli:
            return await self._get_stats_cli(run_id)
        return await self._get_stats_http(run_id)

    async def _get_stats_cli(self, run_id: str) -> Optional[Dict[str, Any]]:
        """Get stats using Atlas CLI."""
        try:
            result = subprocess.run(
                ["atlas", "run", "stats", run_id, "--json"],
                capture_output=True,
                text=True,
                timeout=10,
            )
            if result.returncode == 0:
                return json.loads(result.stdout)
            return None
        except subprocess.SubprocessError:
            return None

    async def _get_stats_http(self, run_id: str) -> Optional[Dict[str, Any]]:
        """Get stats using HTTP API."""
        import httpx
        try:
            async with httpx.AsyncClient() as client:
                response = await client.get(
                    f"{self.base_url}/v1/runs/{run_id}/stats",
                    headers={
                        "Authorization": f"Bearer {self.api_key}",
                        "X-Atlas-Tenant": self.tenant_id,
                    },
                    timeout=10,
                )
                if response.status_code == 200:
                    return response.json()
                return None
        except Exception:
            return None

    async def end_run(self, run_id: str, status: str = "completed") -> bool:
        """End a run.

        Args:
            run_id: Atlas run ID
            status: Final status

        Returns:
            True if successful
        """
        if self._use_cli:
            return await self._end_run_cli(run_id, status)
        return await self._end_run_http(run_id, status)

    async def _end_run_cli(self, run_id: str, status: str) -> bool:
        """End run using Atlas CLI."""
        try:
            result = subprocess.run(
                ["atlas", "run", "end", run_id, "--status", status],
                capture_output=True,
                text=True,
                timeout=10,
            )
            return result.returncode == 0
        except subprocess.SubprocessError:
            return False

    async def _end_run_http(self, run_id: str, status: str) -> bool:
        """End run using HTTP API."""
        import httpx
        try:
            async with httpx.AsyncClient() as client:
                response = await client.post(
                    f"{self.base_url}/v1/runs/{run_id}/end",
                    json={"status": status},
                    headers={
                        "Authorization": f"Bearer {self.api_key}",
                        "X-Atlas-Tenant": self.tenant_id,
                    },
                    timeout=10,
                )
                return response.status_code == 200
        except Exception:
            return False