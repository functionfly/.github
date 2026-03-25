"""
Types for trust-aware agent SDK integrations.
"""

from __future__ import annotations

from dataclasses import asdict, dataclass, field
import hashlib
import json
from typing import Any, Dict, List, Optional


@dataclass
class TrustPolicy:
    """
    Trust policy for filtering candidate tools.

    Defaults are fail-closed for production use.
    """

    min_trust_score: int = 80
    required_trust_levels: List[str] = field(default_factory=list)
    require_verified: bool = True
    capabilities_allow: List[str] = field(default_factory=list)
    capabilities_deny: List[str] = field(default_factory=list)
    max_egress_domains: List[str] = field(default_factory=list)

    def normalized(self) -> Dict[str, Any]:
        """
        Return deterministic normalized dictionary used for policy hashing.
        """
        data = asdict(self)
        for key in (
            "required_trust_levels",
            "capabilities_allow",
            "capabilities_deny",
            "max_egress_domains",
        ):
            values = data.get(key) or []
            data[key] = sorted({str(v).strip().lower() for v in values if str(v).strip()})
        if data["min_trust_score"] is None:
            data["min_trust_score"] = 0
        return data

    def policy_hash(self) -> str:
        normalized = self.normalized()
        payload = json.dumps(normalized, sort_keys=True, separators=(",", ":"))
        return hashlib.sha256(payload.encode("utf-8")).hexdigest()


@dataclass
class TrustedFunction:
    author: str
    name: str
    version: Optional[str]
    trust_score: float
    trust_level: str
    verified: bool
    description: str = ""
    capabilities: List[str] = field(default_factory=list)
    manifest: Dict[str, Any] = field(default_factory=dict)
    profile: Dict[str, Any] = field(default_factory=dict)
    tool_schema: Dict[str, Any] = field(default_factory=dict)


@dataclass
class ToolExecutionMetadata:
    tool_id: str
    author: str
    name: str
    version: Optional[str]
    policy_hash: str


@dataclass
class ToolExecutionEnvelope:
    ok: bool
    data: Any
    error: Optional[Dict[str, Any]]
    cached: bool
    duration_ms: int
    version: Optional[str]
    execution_id: Optional[str]
    metadata: ToolExecutionMetadata


class AgentClientError(Exception):
    """Base exception for agent integration client errors."""


class AgentHTTPError(AgentClientError):
    """HTTP error from FunctionFly API."""

    def __init__(self, status_code: int, message: str, body: Optional[Dict[str, Any]] = None):
        super().__init__(f"HTTP {status_code}: {message}")
        self.status_code = status_code
        self.body = body or {}


class TrustPolicyError(AgentClientError):
    """Raised when trust policy checks fail for a candidate function."""

    def __init__(self, message: str, reasons: Optional[List[str]] = None):
        super().__init__(message)
        self.reasons = reasons or []
