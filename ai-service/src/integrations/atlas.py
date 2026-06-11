"""Atlas integration for FlyMind AI Service.

This module provides the main Atlas integration class for recording
AI agent decision-making with Atlas Memory Engine.
"""

import asyncio
import json
import logging
import time
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional

from .atlas_grpc import AtlasGRPCClient
from ..config import settings

logger = logging.getLogger(__name__)


@dataclass
class AtlasConfig:
    """Atlas configuration."""
    enabled: bool = True
    base_url: str = "http://localhost:7447"
    grpc_host: str = "localhost"
    grpc_port: int = 50051
    api_key: str = ""
    agent_id_prefix: str = "flymind"
    default_tenant_id: str = "default"


@dataclass
class CostData:
    """Embedded cost metadata in events."""
    provider: str
    model: str
    input_tokens: int = 0
    output_tokens: int = 0
    cost_usd: float = 0.0
    latency_ms: int = 0


class AtlasIntegration:
    """Main Atlas integration for FlyMind AI Service."""

    def __init__(self, config: Optional[AtlasConfig] = None):
        self.config = config or AtlasConfig(
            enabled=settings.atlas_enabled if hasattr(settings, 'atlas_enabled') else True,
            base_url=settings.atlas_base_url if hasattr(settings, 'atlas_base_url') else "http://localhost:7447",
            grpc_host=settings.atlas_grpc_host if hasattr(settings, 'atlas_grpc_host') else "localhost",
            grpc_port=settings.atlas_grpc_port if hasattr(settings, 'atlas_grpc_port') else 50051,
            api_key=settings.atlas_api_key if hasattr(settings, 'atlas_api_key') else "",
            agent_id_prefix=settings.atlas_agent_id_prefix if hasattr(settings, 'atlas_agent_id_prefix') else "flymind",
        )
        self._grpc_client: Optional[AtlasGRPCClient] = None
        self._current_run_id: Optional[str] = None
        self._current_span_id: Optional[str] = None
        self._tenant_id: str = self.config.default_tenant_id

    async def initialize(self) -> bool:
        """Initialize gRPC connection.

        Returns:
            True if successful
        """
        if not self.config.enabled:
            logger.info("Atlas integration disabled")
            return True

        self._grpc_client = AtlasGRPCClient(
            host=self.config.grpc_host,
            port=self.config.grpc_port,
            api_key=self.config.api_key,
            tenant_id=self._tenant_id,
            base_url=self.config.base_url,
        )
        return await self._grpc_client.connect()

    @property
    def is_enabled(self) -> bool:
        """Check if Atlas is enabled."""
        return self.config.enabled and self._grpc_client is not None

    async def start_run(
        self,
        agent_id: str,
        metadata: Optional[Dict[str, str]] = None,
        span_id: Optional[str] = None,
        parent_span_id: Optional[str] = None,
    ) -> Optional[str]:
        """Start a new observability run.

        Args:
            agent_id: FunctionFly agent ID
            metadata: Additional metadata
            span_id: Optional span ID
            parent_span_id: Optional parent span ID

        Returns:
            Atlas run ID or None
        """
        if not self.is_enabled:
            return None

        meta = {
            "agent_id": agent_id,
            "tenant_id": self._tenant_id,
            "prefix": self.config.agent_id_prefix,
        }
        if metadata:
            meta.update(metadata)

        self._current_run_id = await self._grpc_client.create_run(meta)
        self._current_span_id = span_id or f"span-{int(time.time() * 1000)}"

        return self._current_run_id

    async def start_span(
        self,
        span_id: str,
        metadata: Optional[Dict[str, str]] = None,
    ) -> Optional[str]:
        """Start a nested span within current run.

        Args:
            span_id: Span identifier
            metadata: Additional metadata

        Returns:
            Span ID or None
        """
        if not self.is_enabled:
            return None

        parent_span_id = self._current_span_id
        self._current_span_id = span_id

        await self.record_decision(
            model="system",
            reasoning=f"span_start:{span_id}",
            metadata={
                "parent_span_id": parent_span_id or "",
                **(metadata or {})
            }
        )

        return self._current_span_id

    async def end_span(self, status: str = "completed") -> None:
        """End current span.

        Args:
            status: Span status
        """
        if not self.is_enabled:
            return

        await self.record_result(
            content=f"span_end:{status}",
            metadata={"span_status": status}
        )

    async def end_run(self, status: str = "completed") -> None:
        """End the current observability run.

        Args:
            status: Run status
        """
        if not self.is_enabled or not self._current_run_id:
            return

        if self._current_span_id:
            await self.end_span(status)

        if self._grpc_client:
            await self._grpc_client.end_run(self._current_run_id, status)

        self._current_run_id = None
        self._current_span_id = None

    async def record_input(
        self,
        content: str,
        role: str = "user",
        metadata: Optional[Dict[str, Any]] = None,
    ) -> None:
        """Record INPUT event.

        Args:
            content: Input content
            role: Role (user, system, assistant)
            metadata: Additional metadata
        """
        await self._record_event("INPUT", {
            "role": role,
            "content": content,
            "span_id": self._current_span_id or "",
            **(metadata or {})
        })

    async def record_decision(
        self,
        model: str,
        reasoning: str,
        tool_call: Optional[Dict[str, Any]] = None,
        cost: Optional[CostData] = None,
        metadata: Optional[Dict[str, Any]] = None,
    ) -> None:
        """Record DECISION event with optional cost.

        Args:
            model: Model identifier
            reasoning: Reasoning text
            tool_call: Optional tool call
            cost: Optional cost data
            metadata: Additional metadata
        """
        payload = {
            "model": model,
            "reasoning": reasoning,
            "span_id": self._current_span_id or "",
        }

        if tool_call:
            payload["tool_call"] = tool_call

        if cost:
            payload["cost"] = {
                "provider": cost.provider,
                "model": cost.model,
                "input_tokens": cost.input_tokens,
                "output_tokens": cost.output_tokens,
                "cost_usd": cost.cost_usd,
                "latency_ms": cost.latency_ms,
            }

        if metadata:
            payload.update(metadata)

        system_id = f"{self.config.agent_id_prefix}-{model}"
        await self._record_event("DECISION", payload, system_id=system_id)

    async def record_action(
        self,
        tool_name: str,
        args: Dict[str, Any],
        result: Optional[Dict[str, Any]] = None,
        error: Optional[str] = None,
    ) -> None:
        """Record ACTION event (tool call).

        Args:
            tool_name: Tool name
            args: Tool arguments
            result: Optional result
            error: Optional error
        """
        payload = {
            "tool_name": tool_name,
            "args": args,
            "span_id": self._current_span_id or "",
        }

        if result:
            payload["result"] = result
        if error:
            payload["error"] = error

        await self._record_event("ACTION", payload)

    async def record_result(
        self,
        content: str,
        role: str = "assistant",
        cost: Optional[CostData] = None,
        metadata: Optional[Dict[str, Any]] = None,
    ) -> None:
        """Record RESULT event with optional cost.

        Args:
            content: Result content
            role: Role
            cost: Optional cost data
            metadata: Additional metadata
        """
        payload = {
            "role": role,
            "content": content,
            "span_id": self._current_span_id or "",
        }

        if cost:
            payload["cost"] = {
                "provider": cost.provider,
                "model": cost.model,
                "input_tokens": cost.input_tokens,
                "output_tokens": cost.output_tokens,
                "cost_usd": cost.cost_usd,
                "latency_ms": cost.latency_ms,
            }

        if metadata:
            payload.update(metadata)

        await self._record_event("RESULT", payload)

    async def record_error(
        self,
        error: str,
        context: Optional[Dict[str, Any]] = None,
    ) -> None:
        """Record ERROR event.

        Args:
            error: Error message
            context: Optional context
        """
        await self._record_event("ERROR", {
            "error": error,
            "context": context or {},
            "span_id": self._current_span_id or "",
        })

    async def _record_event(
        self,
        kind: str,
        payload: Dict[str, Any],
        system_id: Optional[str] = None,
        parent_id: str = "",
    ) -> None:
        """Internal: append event to current run.

        Args:
            kind: Event kind
            payload: Event payload
            system_id: System identifier
            parent_id: Parent event ID
        """
        if not self.is_enabled:
            return

        if not self._current_run_id:
            await self.start_run(agent_id=system_id or "unknown")

        system = system_id or f"{self.config.agent_id_prefix}-default"

        if self._grpc_client:
            await self._grpc_client.append_event(
                run_id=self._current_run_id,
                kind=kind,
                payload=payload,
                system_id=system,
                parent_id=parent_id,
            )

    async def create_team_run(
        self,
        team_id: str,
        members: List[str],
        metadata: Optional[Dict[str, str]] = None,
    ) -> Optional[str]:
        """Create a multi-agent team run.

        Args:
            team_id: Team identifier
            members: List of agent IDs
            metadata: Additional metadata

        Returns:
            Atlas run ID or None
        """
        if not self.is_enabled:
            return None

        meta = {
            "team_id": team_id,
            "members": json.dumps(members),
            "type": "team",
            **(metadata or {})
        }

        run_id = await self._grpc_client.create_run(meta)
        return run_id

    async def record_agent_message(
        self,
        from_agent: str,
        to_agent: str,
        message: str,
        run_id: Optional[str] = None,
    ) -> None:
        """Record inter-agent message in team run.

        Args:
            from_agent: Source agent
            to_agent: Target agent
            message: Message content
            run_id: Optional run ID override
        """
        await self._record_event("ACTION", {
            "type": "team_message",
            "from": from_agent,
            "to": to_agent,
            "message": message,
        }, system_id=from_agent)


_atlas_integration: Optional[AtlasIntegration] = None


async def get_atlas_integration() -> AtlasIntegration:
    """Get or create the global Atlas integration instance.

    Returns:
        AtlasIntegration instance
    """
    global _atlas_integration
    if _atlas_integration is None:
        _atlas_integration = AtlasIntegration()
        await _atlas_integration.initialize()
    return _atlas_integration


async def shutdown_atlas() -> None:
    """Shutdown Atlas integration."""
    global _atlas_integration
    if _atlas_integration is not None:
        await _atlas_integration.end_run()
        _atlas_integration = None