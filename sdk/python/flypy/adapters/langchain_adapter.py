"""
LangChain adapter for FunctionFly trusted tools.
"""

from __future__ import annotations

from typing import Any, Callable, Dict, List, Optional, Tuple, Type

try:
    from pydantic import BaseModel, Field, create_model
except Exception:  # pragma: no cover
    BaseModel = object  # type: ignore
    Field = None  # type: ignore
    create_model = None  # type: ignore

from ..agent_client import AgentClient
from ..agent_types import ToolExecutionEnvelope, TrustPolicy, TrustedFunction


PYTHON_TYPE_MAP = {
    "string": str,
    "number": float,
    "integer": int,
    "boolean": bool,
    "array": list,
    "object": dict,
}


def _sanitize_model_name(name: str) -> str:
    return "".join([ch if ch.isalnum() else "_" for ch in name]).strip("_") or "ToolInput"


def json_schema_to_pydantic_model(tool_name: str, schema: Dict[str, Any]) -> Optional[Type[BaseModel]]:
    if not create_model:
        return None
    properties = schema.get("properties") or {}
    required = set(schema.get("required") or [])
    fields: Dict[str, Tuple[Any, Any]] = {}

    for prop, prop_schema in properties.items():
        prop_type = PYTHON_TYPE_MAP.get(str(prop_schema.get("type", "string")).lower(), Any)
        description = str(prop_schema.get("description") or "")
        if prop in required:
            default = Field(..., description=description) if Field else ...
        else:
            default = Field(None, description=description) if Field else None
        fields[prop] = (prop_type, default)

    model_name = f"{_sanitize_model_name(tool_name).title()}Input"
    return create_model(model_name, **fields)  # type: ignore[arg-type]


class LangChainAdapter:
    """
    Produces LangChain-compatible tools from FunctionFly trusted functions.
    """

    def __init__(self, client: AgentClient):
        self.client = client

    def _build_tool_callable(
        self,
        trusted_function: TrustedFunction,
        policy: TrustPolicy,
    ) -> Callable[..., Dict[str, Any]]:
        def _tool_callable(**kwargs: Any) -> Dict[str, Any]:
            envelope: ToolExecutionEnvelope = self.client.execute_trusted_tool(
                trusted_function=trusted_function,
                policy=policy,
                tool_input=kwargs,
            )
            return {
                "ok": envelope.ok,
                "data": envelope.data,
                "error": envelope.error,
                "cached": envelope.cached,
                "duration_ms": envelope.duration_ms,
                "version": envelope.version,
                "execution_id": envelope.execution_id,
                "metadata": {
                    "tool_id": envelope.metadata.tool_id,
                    "policy_hash": envelope.metadata.policy_hash,
                    "author": envelope.metadata.author,
                    "name": envelope.metadata.name,
                    "version": envelope.metadata.version,
                },
            }

        return _tool_callable

    def _langchain_structured_tool(
        self,
        trusted_function: TrustedFunction,
        policy: TrustPolicy,
    ) -> Optional[Any]:
        try:
            from langchain_core.tools import StructuredTool
        except Exception:
            return None

        schema = trusted_function.tool_schema.get("input_schema") or trusted_function.tool_schema.get("parameters") or {}
        args_schema = json_schema_to_pydantic_model(
            f"{trusted_function.author}_{trusted_function.name}",
            schema if isinstance(schema, dict) else {},
        )
        fn = self._build_tool_callable(trusted_function, policy)
        return StructuredTool.from_function(
            func=fn,
            name=f"{trusted_function.author}_{trusted_function.name}",
            description=trusted_function.description or f"Execute {trusted_function.author}/{trusted_function.name}",
            args_schema=args_schema,
        )

    def build_tools(
        self,
        policy: TrustPolicy,
        query: str,
        category: Optional[str] = None,
        limit: int = 20,
    ) -> List[Any]:
        trusted_functions = self.client.discover_trusted_functions(
            policy=policy,
            query=query,
            category=category,
            limit=limit,
        )
        tools: List[Any] = []
        for trusted in trusted_functions:
            structured = self._langchain_structured_tool(trusted, policy)
            if structured is not None:
                tools.append(structured)
                continue

            # Fallback: plain tool descriptor for environments without langchain installed.
            tools.append(
                {
                    "name": f"{trusted.author}_{trusted.name}",
                    "description": trusted.description,
                    "input_schema": trusted.tool_schema.get("input_schema") or {},
                    "callable": self._build_tool_callable(trusted, policy),
                    "metadata": {
                        "author": trusted.author,
                        "name": trusted.name,
                        "version": trusted.version,
                        "trust_score": trusted.trust_score,
                        "trust_level": trusted.trust_level,
                        "verified": trusted.verified,
                    },
                }
            )
        return tools

    def execute_tool(
        self,
        trusted_function: TrustedFunction,
        policy: TrustPolicy,
        tool_input: Dict[str, Any],
    ) -> ToolExecutionEnvelope:
        return self.client.execute_trusted_tool(trusted_function, policy, tool_input)
