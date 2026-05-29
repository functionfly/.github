#!/usr/bin/env python3
"""Optional live API integration tests for flypy agent client."""

import os

import pytest

from flypy.agent_client import AgentClient


LIVE = os.environ.get("FUNCTIONFLY_INTEGRATION_TESTS") == "true"
API_BASE = os.environ.get("FUNCTIONFLY_API_BASE") or os.environ.get("FUNCTIONFLY_API_URL")
API_KEY = os.environ.get("FUNCTIONFLY_API_KEY")
FUNCTION_ID = os.environ.get("FUNCTIONFLY_TEST_FUNCTION_ID")


pytestmark = pytest.mark.skipif(
    not (LIVE and API_BASE and API_KEY),
    reason="Set FUNCTIONFLY_INTEGRATION_TESTS=true plus API base/key",
)


@pytest.fixture
def client() -> AgentClient:
    return AgentClient(api_base=API_BASE, api_key=API_KEY)


def test_live_registry_search(client: AgentClient):
    results = client.search_registry(query="", limit=5)
    assert isinstance(results, list)


def test_live_get_function_by_id(client: AgentClient):
    if not FUNCTION_ID:
        pytest.skip("Set FUNCTIONFLY_TEST_FUNCTION_ID to a known registry function UUID")

    profile = client.get_function_by_id(FUNCTION_ID)
    assert str(profile.get("id")) == FUNCTION_ID
    assert profile.get("author")
    assert profile.get("name")
