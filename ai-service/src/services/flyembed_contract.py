"""Contract text builder for FlyEmbed triple-vector embeddings.

Builds structured contract representations from function manifest data
(input/output schemas, types, parameters, error codes, runtime constraints).
"""

import json
import logging
from typing import Any

logger = logging.getLogger(__name__)


class ContractTextBuilder:
    """Builds structured contract representation from function manifest data."""

    def build(self, function_data: dict) -> str:
        """Build contract text from function data.

        Produces a structured representation suitable for contract vector embedding:
            Function: jwt-verify
            Accepts: token: string (required)
            Input Schema: {"type":"object","properties":{...},"required":["token"]}
            Returns: {"type":"object","properties":{...}}
            Runtime: node18
            Timeout: 30000ms
            Deterministic: true
            Capabilities: network

        Args:
            function_data: Dict with manifest, runtime, capabilities, etc.

        Returns:
            Structured contract text string
        """
        manifest = function_data.get("manifest", {})
        if not isinstance(manifest, dict):
            manifest = {}

        parts = []

        # Function name
        name = function_data.get("name", "unknown")
        parts.append(f"Function: {name}")

        # Input schema
        input_schema = manifest.get("input", {})
        if input_schema and isinstance(input_schema, dict):
            self._append_input_section(parts, input_schema)

        # Output schema
        output_schema = manifest.get("output", {})
        if output_schema and isinstance(output_schema, dict):
            self._append_output_section(parts, output_schema)

        # Runtime constraints
        if runtime := function_data.get("runtime"):
            parts.append(f"Runtime: {runtime}")
        if timeout := manifest.get("timeout_ms"):
            parts.append(f"Timeout: {timeout}ms")
        if deterministic := manifest.get("deterministic"):
            parts.append(f"Deterministic: {deterministic}")
        if caps := function_data.get("capabilities"):
            if isinstance(caps, list):
                parts.append(f"Capabilities: {', '.join(caps)}")
        if side_effects := manifest.get("side_effects"):
            parts.append(f"Side Effects: {side_effects}")
        if idempotent := manifest.get("idempotent"):
            parts.append(f"Idempotent: {idempotent}")

        return "\n".join(parts)

    def _append_input_section(self, parts: list[str], input_schema: dict) -> None:
        """Append input schema section to parts list."""
        props = input_schema.get("properties", {})
        required = set(input_schema.get("required", []))

        if props:
            param_lines = []
            for param_name, param_def in props.items():
                if not isinstance(param_def, dict):
                    continue
                param_type = param_def.get("type", "any")
                desc = param_def.get("description", "")
                req_marker = " (required)" if param_name in required else ""
                line = f"{param_name}: {param_type}{req_marker}"
                param_lines.append(line)
                if desc:
                    param_lines.append(f"  {desc}")

            if param_lines:
                # Only add top-level param names for Accepts line
                top_level = [p for p in param_lines if not p.startswith("  ")]
                parts.append(f"Accepts: {', '.join(top_level)}")

            # Full JSON schema (compact)
            try:
                compact = json.dumps(input_schema, separators=(",", ":"))
                parts.append(f"Input Schema: {compact}")
            except (TypeError, ValueError):
                pass

    def _append_output_section(self, parts: list[str], output_schema: dict) -> None:
        """Append output schema section to parts list."""
        try:
            compact = json.dumps(output_schema, separators=(",", ":"))
            parts.append(f"Returns: {compact}")
        except (TypeError, ValueError):
            pass
