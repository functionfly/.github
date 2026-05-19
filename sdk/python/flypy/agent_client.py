"""
Trust-aware FunctionFly client for agent SDK integrations.
"""

from __future__ import annotations

import json
import os
import socket
import time
from typing import Any, Callable, Dict, List, Optional
from urllib.parse import urlencode, urlparse
import urllib.error
import urllib.request

from .agent_policy import evaluate_candidate
from .agent_types import (
    AgentHTTPError,
    ToolExecutionEnvelope,
    ToolExecutionMetadata,
    TrustPolicy,
    TrustPolicyError,
    TrustedFunction,
)


DEFAULT_API_BASE = os.environ.get("FUNCTIONFLY_API_BASE", "")  # Must be set explicitly


def _is_local_host(hostname: str) -> bool:
    host = (hostname or "").lower()
    return host in ("localhost", "127.0.0.1", "::1")


class AgentClient:
    """
    Minimal trust-aware client for discovering and executing FunctionFly tools.
    """

    def __init__(
        self,
        api_base: str = DEFAULT_API_BASE,
        api_key: Optional[str] = None,
        timeout_seconds: float = 10.0,
        max_retries: int = 2,
        retry_backoff_seconds: float = 0.3,
        log_hook: Optional[Callable[[Dict[str, Any]], None]] = None,
    ):
        self.api_base = self._normalize_api_base(api_base)
        self.api_key = api_key or os.environ.get("FUNCTIONFLY_API_KEY")
        self.timeout_seconds = timeout_seconds
        self.max_retries = max(0, int(max_retries))
        self.retry_backoff_seconds = max(0.0, float(retry_backoff_seconds))
        self.log_hook = log_hook

    @staticmethod
    def _normalize_api_base(api_base: str) -> str:
        base = (api_base or "").strip().rstrip("/")
        parsed = urlparse(base)
        if not parsed.scheme or not parsed.netloc:
            raise ValueError("api_base must include scheme and host, e.g. https://api.functionfly.com")
        if parsed.scheme != "https" and not _is_local_host(parsed.hostname or ""):
            raise ValueError("api_base must use https outside localhost")
        return base

    def _log(self, event: Dict[str, Any]) -> None:
        if not self.log_hook:
            return
        redacted = dict(event)
        headers = dict(redacted.get("headers") or {})
        if "Authorization" in headers:
            headers["Authorization"] = "Bearer ***REDACTED***"
        redacted["headers"] = headers
        self.log_hook(redacted)

    def _request(
        self,
        method: str,
        path: str,
        params: Optional[Dict[str, Any]] = None,
        payload: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        url = f"{self.api_base}{path}"
        if params:
            clean_params = {k: v for k, v in params.items() if v is not None}
            if clean_params:
                url = f"{url}?{urlencode(clean_params)}"

        headers = {
            "Accept": "application/json",
            "Content-Type": "application/json",
            "User-Agent": "flypy-agent-client/0.1",
        }
        if self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"

        data = None
        if payload is not None:
            data = json.dumps(payload).encode("utf-8")

        attempts = self.max_retries + 1
        last_error: Optional[Exception] = None

        for attempt in range(1, attempts + 1):
            request = urllib.request.Request(
                url=url,
                data=data,
                headers=headers,
                method=method.upper(),
            )
            self._log({"event": "request", "method": method.upper(), "url": url, "headers": headers, "attempt": attempt})
            try:
                with urllib.request.urlopen(request, timeout=self.timeout_seconds) as response:
                    body = response.read().decode("utf-8")
                    parsed = json.loads(body) if body else {}
                    self._log({"event": "response", "method": method.upper(), "url": url, "status_code": response.status})
                    return parsed
            except urllib.error.HTTPError as err:
                body = err.read().decode("utf-8")
                parsed_body: Dict[str, Any]
                try:
                    parsed_body = json.loads(body) if body else {}
                except json.JSONDecodeError:
                    parsed_body = {"raw": body}

                is_retryable = method.upper() == "GET" and err.code >= 500 and attempt < attempts
                if is_retryable:
                    time.sleep(self.retry_backoff_seconds * attempt)
                    continue
                message = parsed_body.get("error", {}).get("message") or parsed_body.get("error") or str(err)
                raise AgentHTTPError(err.code, str(message), body=parsed_body)
            except (urllib.error.URLError, socket.timeout, TimeoutError) as err:
                last_error = err
                is_retryable = method.upper() == "GET" and attempt < attempts
                if is_retryable:
                    time.sleep(self.retry_backoff_seconds * attempt)
                    continue
                raise AgentHTTPError(0, f"network error: {err}") from err

        raise AgentHTTPError(0, f"request failed after retries: {last_error}")

    def search_registry(
        self,
        query: str,
        category: Optional[str] = None,
        min_rating: Optional[float] = None,
        limit: int = 20,
        offset: int = 0,
    ) -> List[Dict[str, Any]]:
        result = self._request(
            "GET",
            "/v1/registry/search",
            params={
                "q": query,
                "category": category,
                "min_rating": min_rating,
                "limit": limit,
                "offset": offset,
            },
        )
        if isinstance(result, list):
            return result
        return result.get("functions") or result.get("results") or []

    def get_function_profile(self, author: str, name: str, expand_manifest: bool = True) -> Dict[str, Any]:
        params = {"expand": "manifest"} if expand_manifest else None
        return self._request("GET", f"/v1/registry/functions/{author}/{name}", params=params)

    def get_ai_schema(self, author: str, name: str) -> Dict[str, Any]:
        return self._request("GET", f"/fx/{author}/{name}/ai-schema")

    def execute_function(
        self,
        author: str,
        name: str,
        function_input: Dict[str, Any],
        version: Optional[str] = None,
    ) -> Dict[str, Any]:
        target = f"/v1/fx/{author}/{name}"
        if version:
            target = f"{target}@{version}"
        return self._request("POST", target, payload=function_input)

    @staticmethod
    def _extract_function_payload(profile_payload: Dict[str, Any]) -> Dict[str, Any]:
        if "function" in profile_payload and isinstance(profile_payload["function"], dict):
            return profile_payload["function"]
        return profile_payload

    def discover_trusted_functions(
        self,
        policy: TrustPolicy,
        query: str,
        category: Optional[str] = None,
        limit: int = 20,
    ) -> List[TrustedFunction]:
        candidates = self.search_registry(
            query=query,
            category=category,
            min_rating=policy.min_trust_score,
            limit=limit,
        )
        trusted: List[TrustedFunction] = []

        for candidate in candidates:
            author = candidate.get("author")
            name = candidate.get("name")
            if not author or not name:
                continue

            profile_payload = self.get_function_profile(author, name, expand_manifest=True)
            profile = self._extract_function_payload(profile_payload)

            merged = dict(candidate)
            merged.update(profile)
            merged["manifest"] = profile_payload.get("manifest") or profile.get("manifest") or {}
            merged["capabilities"] = (
                profile.get("capabilities")
                or merged["manifest"].get("capabilities")
                or candidate.get("capabilities")
                or []
            )

            allowed, reasons = evaluate_candidate(policy, merged)
            if not allowed:
                # Explicitly block disallowed candidates; do not silently pass them through.
                continue

            tool_schema = self.get_ai_schema(author, name)
            trusted.append(
                TrustedFunction(
                    author=str(author),
                    name=str(name),
                    version=(profile.get("version") or candidate.get("version")),
                    trust_score=float(merged.get("trust_score", 0)),
                    trust_level=str(merged.get("trust_level", "")),
                    verified=bool(merged.get("verified", False)),
                    description=str(profile.get("description") or candidate.get("description") or ""),
                    capabilities=[str(c) for c in merged.get("capabilities", [])],
                    manifest=merged.get("manifest") or {},
                    profile=profile_payload,
                    tool_schema=tool_schema,
                )
            )
            if reasons:
                raise TrustPolicyError("unexpected policy state", reasons=reasons)

        return trusted

    def execute_trusted_tool(
        self,
        trusted_function: TrustedFunction,
        policy: TrustPolicy,
        tool_input: Dict[str, Any],
    ) -> ToolExecutionEnvelope:
        response = self.execute_function(
            trusted_function.author,
            trusted_function.name,
            tool_input,
            version=trusted_function.version,
        )

        meta = ToolExecutionMetadata(
            tool_id=f"{trusted_function.author}/{trusted_function.name}",
            author=trusted_function.author,
            name=trusted_function.name,
            version=trusted_function.version,
            policy_hash=policy.policy_hash(),
        )
        return ToolExecutionEnvelope(
            ok=bool(response.get("ok", True)),
            data=response.get("data"),
            error=response.get("error"),
            cached=bool(response.get("cached", False)),
            duration_ms=int(response.get("duration_ms", 0)),
            version=response.get("version") or trusted_function.version,
            execution_id=response.get("execution_id"),
            metadata=meta,
        )
