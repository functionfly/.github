#!/usr/bin/env python3
"""
Tests for FunctionFly agent SDK integrations.
"""

import json
from unittest.mock import patch

from flypy import AgentClient, LangChainAdapter, TrustPolicy
from flypy.agent_policy import evaluate_candidate


class _MockHTTPResponse:
    def __init__(self, payload, status=200):
        self._payload = payload
        self.status = status

    def read(self):
        return json.dumps(self._payload).encode("utf-8")

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc, tb):
        return False


def test_policy_filter_blocks_unverified_and_low_score():
    policy = TrustPolicy(min_trust_score=80, require_verified=True, required_trust_levels=["high"])
    candidate = {
        "verified": False,
        "trust_score": 55,
        "trust_level": "medium",
        "capabilities": ["http_get"],
    }
    allowed, reasons = evaluate_candidate(policy, candidate)
    assert allowed is False
    assert any("not verified" in reason for reason in reasons)
    assert any("below minimum" in reason for reason in reasons)


def test_policy_hash_is_deterministic():
    p1 = TrustPolicy(required_trust_levels=["high", "verified"], capabilities_allow=["http_get", "http_post"])
    p2 = TrustPolicy(required_trust_levels=["verified", "high"], capabilities_allow=["http_post", "http_get"])
    assert p1.policy_hash() == p2.policy_hash()


def test_langchain_adapter_fallback_tool_shape():
    class _DummyClient:
        def discover_trusted_functions(self, policy, query, category=None, limit=20):
            return [
                type(
                    "TF",
                    (),
                    {
                        "author": "demo",
                        "name": "echo",
                        "version": "1.0.0",
                        "trust_score": 92.0,
                        "trust_level": "high",
                        "verified": True,
                        "description": "Echo text",
                        "tool_schema": {"input_schema": {"type": "object", "properties": {"text": {"type": "string"}}}},
                    },
                )()
            ]

        def execute_trusted_tool(self, trusted_function, policy, tool_input):
            return type(
                "Env",
                (),
                {
                    "ok": True,
                    "data": {"echo": tool_input.get("text")},
                    "error": None,
                    "cached": False,
                    "duration_ms": 5,
                    "version": "1.0.0",
                    "execution_id": "exec_123",
                    "metadata": type(
                        "Meta",
                        (),
                        {
                            "tool_id": "demo/echo",
                            "policy_hash": policy.policy_hash(),
                            "author": "demo",
                            "name": "echo",
                            "version": "1.0.0",
                        },
                    )(),
                },
            )()

    adapter = LangChainAdapter(_DummyClient())
    tools = adapter.build_tools(TrustPolicy(), query="echo")
    assert len(tools) == 1
    if isinstance(tools[0], dict):
        result = tools[0]["callable"](text="hello")
    else:
        result = tools[0].invoke({"text": "hello"})
    assert result["ok"] is True
    assert result["data"]["echo"] == "hello"
    assert "policy_hash" in result["metadata"]


def test_e2e_smoke_discover_filter_schema_execute():
    search_payload = [
        {"author": "demo", "name": "echo", "trust_score": 90, "trust_level": "high", "verified": True},
        {"author": "bad", "name": "unsafe", "trust_score": 10, "trust_level": "low", "verified": False},
    ]
    profile_payload_demo = {
        "function": {
            "author": "demo",
            "name": "echo",
            "version": "1.2.3",
            "description": "Echo",
            "trust_score": 90,
            "trust_level": "high",
            "verified": True,
            "capabilities": ["http_get"],
        },
        "manifest": {"capabilities": ["http_get"]},
    }
    profile_payload_bad = {
        "function": {
            "author": "bad",
            "name": "unsafe",
            "version": "0.0.1",
            "description": "Unsafe",
            "trust_score": 10,
            "trust_level": "low",
            "verified": False,
            "capabilities": ["secrets_read"],
        },
        "manifest": {"capabilities": ["secrets_read"]},
    }
    schema_payload = {"input_schema": {"type": "object", "properties": {"text": {"type": "string"}}}}
    execute_payload = {"ok": True, "data": {"result": "HELLO"}, "cached": False, "duration_ms": 8, "version": "1.2.3", "execution_id": "exec_456"}

    def _mock_urlopen(request, timeout=10):
        url = request.full_url
        method = request.get_method()
        if url.endswith("/v1/registry/search?q=echo&min_rating=80&limit=20&offset=0"):
            return _MockHTTPResponse(search_payload)
        if url.endswith("/v1/registry/functions/demo/echo?expand=manifest"):
            return _MockHTTPResponse(profile_payload_demo)
        if url.endswith("/v1/registry/functions/bad/unsafe?expand=manifest"):
            return _MockHTTPResponse(profile_payload_bad)
        if url.endswith("/fx/demo/echo/ai-schema"):
            return _MockHTTPResponse(schema_payload)
        if url.endswith("/v1/fx/demo/echo@1.2.3") and method == "POST":
            return _MockHTTPResponse(execute_payload)
        raise AssertionError(f"Unexpected request: {method} {url}")

    with patch("urllib.request.urlopen", side_effect=_mock_urlopen):
        client = AgentClient(api_base="http://localhost:8080", api_key="test-key")
        policy = TrustPolicy()
        trusted = client.discover_trusted_functions(policy=policy, query="echo")
        assert len(trusted) == 1
        envelope = client.execute_trusted_tool(trusted[0], policy=policy, tool_input={"text": "hello"})
        assert envelope.ok is True
        assert envelope.metadata.tool_id == "demo/echo"
        assert envelope.metadata.policy_hash == policy.policy_hash()
        assert envelope.version == "1.2.3"
