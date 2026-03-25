"""
Thin CrewAI adapter built on top of shared FunctionFly client.
"""

from __future__ import annotations

from typing import Any, Dict, List, Optional

from ..agent_client import AgentClient
from ..agent_types import TrustPolicy, TrustedFunction


class CrewAIAdapter:
    """
    Returns CrewAI-friendly tool descriptors.
    """

    def __init__(self, client: AgentClient):
        self.client = client

    def build_tools(
        self,
        policy: TrustPolicy,
        query: str,
        category: Optional[str] = None,
        limit: int = 20,
    ) -> List[Dict[str, Any]]:
        trusted_functions = self.client.discover_trusted_functions(
            policy=policy,
            query=query,
            category=category,
            limit=limit,
        )
        return [self._to_crewai_tool(t, policy) for t in trusted_functions]

    def _to_crewai_tool(self, trusted: TrustedFunction, policy: TrustPolicy) -> Dict[str, Any]:
        def _call(**kwargs: Any) -> Dict[str, Any]:
            envelope = self.client.execute_trusted_tool(trusted, policy, kwargs)
            return {
                "ok": envelope.ok,
                "data": envelope.data,
                "error": envelope.error,
                "cached": envelope.cached,
                "duration_ms": envelope.duration_ms,
                "execution_id": envelope.execution_id,
                "metadata": {
                    "tool_id": envelope.metadata.tool_id,
                    "policy_hash": envelope.metadata.policy_hash,
                    "version": envelope.version,
                },
            }

        input_schema = trusted.tool_schema.get("input_schema") or trusted.tool_schema.get("parameters") or {}
        return {
            "name": f"{trusted.author}_{trusted.name}",
            "description": trusted.description,
            "args_schema": input_schema,
            "func": _call,
            "metadata": {
                "author": trusted.author,
                "name": trusted.name,
                "version": trusted.version,
                "trust_score": trusted.trust_score,
                "trust_level": trusted.trust_level,
                "verified": trusted.verified,
            },
        }
