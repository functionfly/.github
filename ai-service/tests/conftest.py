"""Shared test fixtures and import mocking."""

import sys
from unittest.mock import MagicMock

# Mock the redis_client module to prevent import errors from cache.py
_mock_redis = MagicMock()
_mock_redis.get_redis_client = MagicMock(return_value=MagicMock())
sys.modules["src.services.redis_client"] = _mock_redis
