"""Tests for ML API security improvements.

Tests for:
- InteractionType enum validation
- Error sanitization
- Training rate limiting
- Production validation (encryption enforcement)
"""

import pytest
import time
from unittest.mock import MagicMock, AsyncMock, patch
from fastapi import HTTPException

from src.models.schemas import InteractionType, RecommendationInteractionRequest
from src.api.routes_ml import _sanitize_error, TrainingRateLimiter


class TestInteractionType:
    """Tests for InteractionType enum validation."""

    def test_interaction_type_enum_values(self):
        """All expected interaction types should be defined."""
        expected_types = {"view", "install", "execute", "rate", "search_click", "search_impression"}
        actual_types = {t.value for t in InteractionType}
        assert expected_types == actual_types

    def test_recommendation_interaction_request_accepts_valid_type(self):
        """Valid interaction types should be accepted."""
        for interaction_type in InteractionType:
            request = RecommendationInteractionRequest(
                user_id="user123",
                function_id="func456",
                interaction_type=interaction_type,
            )
            assert request.interaction_type == interaction_type

    def test_recommendation_interaction_request_default_is_view(self):
        """Default interaction type should be view."""
        request = RecommendationInteractionRequest(
            user_id="user123",
            function_id="func456",
        )
        assert request.interaction_type == InteractionType.VIEW

    def test_recommendation_interaction_request_rejects_invalid_type(self):
        """Invalid interaction types should be rejected by Pydantic."""
        with pytest.raises(Exception):
            RecommendationInteractionRequest(
                user_id="user123",
                function_id="func456",
                interaction_type="invalid_type",
            )

    def test_recommendation_interaction_request_accepts_string_value(self):
        """String values matching enum should work."""
        request = RecommendationInteractionRequest(
            user_id="user123",
            function_id="func456",
            interaction_type="install",
        )
        assert request.interaction_type == InteractionType.INSTALL


class TestErrorSanitization:
    """Tests for error sanitization in _sanitize_error."""

    def test_http_exception_passes_through(self):
        """HTTPException details should be preserved."""
        e = HTTPException(status_code=400, detail="Invalid function_id")
        result = _sanitize_error(e)
        assert result == "Invalid function_id"

    def test_timeout_error_returns_generic_message(self):
        """TimeoutError should return generic message."""
        e = TimeoutError("Connection timed out after 30 seconds")
        result = _sanitize_error(e)
        assert result == "The operation timed out. Please try again."

    def test_asyncio_timeout_error_returns_generic_message(self):
        """asyncio.TimeoutError should return generic message."""
        e = asyncio.TimeoutError("Task timed out")
        result = _sanitize_error(e)
        assert result == "The operation timed out. Please try again."

    def test_redis_error_returns_cache_message(self):
        """Redis errors should return cache service message."""
        e = Exception("redis ConnectionRefusedError")
        result = _sanitize_error(e)
        assert result == "A cache service is temporarily unavailable"

    def test_database_error_returns_db_message(self):
        """Database errors should return database service message."""
        e = Exception("postgres connection failed")
        result = _sanitize_error(e)
        assert result == "A database service is temporarily unavailable"

    def test_generic_exception_returns_generic_message(self):
        """Generic exceptions should return generic message without leaking details."""
        e = Exception("Some internal error with sensitive data")
        result = _sanitize_error(e)
        assert result == "An internal error occurred"

    def test_long_error_message_returns_generic(self):
        """Long error messages should not leak details."""
        long_error = "x" * 200
        e = Exception(long_error)
        result = _sanitize_error(e)
        assert result == "An internal error occurred"

    def test_sensitive_keywords_in_error_not_leaked(self):
        """Errors containing sensitive keywords should not leak."""
        sensitive_errors = [
            Exception("password is abc123"),
            Exception("token=sk-12345 secret"),
            Exception("api_key=my-secret-key auth failed"),
        ]
        for e in sensitive_errors:
            result = _sanitize_error(e)
            assert "password" not in result.lower()
            assert "token" not in result.lower()
            assert "secret" not in result.lower()
            assert "api_key" not in result.lower()


class TestTrainingRateLimiter:
    """Tests for TrainingRateLimiter."""

    @pytest.fixture
    def limiter(self):
        """Create a fresh limiter for each test."""
        return TrainingRateLimiter()

    @pytest.mark.asyncio
    async def test_first_request_allowed(self, limiter):
        """First training request should be allowed."""
        await limiter.check_training_rate("tenant1")

    @pytest.mark.asyncio
    async def test_multiple_requests_within_limit(self, limiter):
        """Requests within limit should be allowed."""
        with patch("src.api.routes_ml.settings") as mock_settings:
            mock_settings.ml_training_rate_limit_per_hour = 4

            for _ in range(4):
                await limiter.check_training_rate("tenant1")

    @pytest.mark.asyncio
    async def test_requests_exceeding_limit_raises(self, limiter):
        """Requests exceeding limit should raise 429."""
        with patch("src.api.routes_ml.settings") as mock_settings:
            mock_settings.ml_training_rate_limit_per_hour = 2

            await limiter.check_training_rate("tenant1")
            await limiter.check_training_rate("tenant1")

            with pytest.raises(HTTPException) as exc_info:
                await limiter.check_training_rate("tenant1")

            assert exc_info.value.status_code == 429
            assert "Training rate limit exceeded" in exc_info.value.detail
            assert "Retry-After" in exc_info.value.headers

    @pytest.mark.asyncio
    async def test_different_tenants_independent_limits(self, limiter):
        """Different tenants should have independent rate limits."""
        with patch("src.api.routes_ml.settings") as mock_settings:
            mock_settings.ml_training_rate_limit_per_hour = 2

            await limiter.check_training_rate("tenant1")
            await limiter.check_training_rate("tenant1")

            await limiter.check_training_rate("tenant2")

            with pytest.raises(HTTPException) as exc_info:
                await limiter.check_training_rate("tenant1")
            assert exc_info.value.status_code == 429

            await limiter.check_training_rate("tenant2")

    @pytest.mark.asyncio
    async def test_rate_limit_resets_after_window(self, limiter):
        """Rate limit should reset after window expires."""
        with patch("src.api.routes_ml.settings") as mock_settings:
            mock_settings.ml_training_rate_limit_per_hour = 2

            await limiter.check_training_rate("tenant1")
            await limiter.check_training_rate("tenant1")

            with pytest.raises(HTTPException):
                await limiter.check_training_rate("tenant1")

            limiter._counts["tenant1"]["window_start"] = time.time() - 3601

            await limiter.check_training_rate("tenant1")


class TestProductionValidation:
    """Tests for production validation logic."""

    def test_is_production_with_explicit_production_env(self):
        """ENVIRONMENT=production should be detected as production."""
        with patch.dict("os.environ", {"ENVIRONMENT": "production"}):
            from src.services.ml_common.production_validation import ProductionValidator
            validator = ProductionValidator()
            assert validator._is_production_environment() is True

    def test_not_production_with_development_env(self):
        """ENVIRONMENT=development should not be detected as production."""
        with patch.dict("os.environ", {"ENVIRONMENT": "development"}):
            from src.services.ml_common.production_validation import ProductionValidator
            validator = ProductionValidator()
            assert validator._is_production_environment() is False

    def test_not_production_with_test_env(self):
        """ENVIRONMENT=test should not be detected as production."""
        with patch.dict("os.environ", {"ENVIRONMENT": "test"}):
            from src.services.ml_common.production_validation import ProductionValidator
            validator = ProductionValidator()
            assert validator._is_production_environment() is False

    def test_cloud_redis_detected_as_production(self):
        """Redis URL with cloud provider should be detected as production."""
        with patch.dict("os.environ", {"ENVIRONMENT": ""}):
            with patch("src.services.ml_common.production_validation.settings") as mock_settings:
                mock_settings.redis_url = "redis://rediscloud.com:12345"
                from src.services.ml_common.production_validation import ProductionValidator
                validator = ProductionValidator()
                assert validator._is_production_environment() is True

    def test_localhost_redis_not_production(self):
        """Localhost Redis should not be detected as production."""
        with patch.dict("os.environ", {"ENVIRONMENT": ""}):
            with patch("src.services.ml_common.production_validation.settings") as mock_settings:
                mock_settings.redis_url = "redis://localhost:6379"
                from src.services.ml_common.production_validation import ProductionValidator
                validator = ProductionValidator()
                assert validator._is_production_environment() is False


import asyncio
