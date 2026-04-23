"""Tests for API key authentication and validation."""

import pytest
from datetime import datetime, timedelta

from src.security.auth import (
    APIKeyValidator,
    APIKeyInfo,
    KeyStatus,
    KeyScope,
    get_api_key_validator,
)


class TestAPIKeyInfo:
    """Tests for APIKeyInfo dataclass."""

    def test_is_valid_active_key(self):
        """Active key without expiration should be valid."""
        info = APIKeyInfo(
            key_id="test",
            tenant_id="tenant1",
            name="Test Key",
            scopes=[KeyScope.FULL],
            status=KeyStatus.ACTIVE,
            created_at=datetime.utcnow(),
        )
        assert info.is_valid() is True

    def test_is_invalid_revoked_key(self):
        """Revoked key should not be valid."""
        info = APIKeyInfo(
            key_id="test",
            tenant_id="tenant1",
            name="Test Key",
            scopes=[KeyScope.FULL],
            status=KeyStatus.REVOKED,
            created_at=datetime.utcnow(),
        )
        assert info.is_valid() is False

    def test_is_invalid_expired_key(self):
        """Expired key should not be valid."""
        info = APIKeyInfo(
            key_id="test",
            tenant_id="tenant1",
            name="Test Key",
            scopes=[KeyScope.FULL],
            status=KeyStatus.ACTIVE,
            created_at=datetime.utcnow(),
            expires_at=datetime.utcnow() - timedelta(days=1),
        )
        assert info.is_valid() is False

    def test_is_valid_not_yet_expired(self):
        """Key with future expiration should still be valid."""
        info = APIKeyInfo(
            key_id="test",
            tenant_id="tenant1",
            name="Test Key",
            scopes=[KeyScope.FULL],
            status=KeyStatus.ACTIVE,
            created_at=datetime.utcnow(),
            expires_at=datetime.utcnow() + timedelta(days=1),
        )
        assert info.is_valid() is True

    def test_has_scope_exact(self):
        """Should check for exact scope match."""
        info = APIKeyInfo(
            key_id="test",
            tenant_id="tenant1",
            name="Test Key",
            scopes=[KeyScope.READ, KeyScope.WRITE],
            status=KeyStatus.ACTIVE,
            created_at=datetime.utcnow(),
        )
        assert info.has_scope(KeyScope.READ) is True
        assert info.has_scope(KeyScope.WRITE) is True
        assert info.has_scope(KeyScope.ADMIN) is False

    def test_has_scope_full_grants_all(self):
        """FULL scope should grant access to everything."""
        info = APIKeyInfo(
            key_id="test",
            tenant_id="tenant1",
            name="Test Key",
            scopes=[KeyScope.FULL],
            status=KeyStatus.ACTIVE,
            created_at=datetime.utcnow(),
        )
        assert info.has_scope(KeyScope.READ) is True
        assert info.has_scope(KeyScope.WRITE) is True
        assert info.has_scope(KeyScope.ADMIN) is True
        assert info.has_scope(KeyScope.EMBED_WRITE) is True
        assert info.has_scope(KeyScope.CHAT_WRITE) is True


class TestAPIKeyValidator:
    """Tests for APIKeyValidator."""

    def test_create_key(self):
        """Should create a new API key."""
        validator = APIKeyValidator()
        full_key, info = validator.create_key(
            tenant_id="tenant1",
            name="Test Key",
            scopes=[KeyScope.FULL],
        )

        assert full_key.startswith("fly_")
        assert info.tenant_id == "tenant1"
        assert info.name == "Test Key"
        assert info.status == KeyStatus.ACTIVE
        assert KeyScope.FULL in info.scopes

    def test_validate_key_valid(self):
        """Should validate a valid key."""
        validator = APIKeyValidator()
        full_key, info = validator.create_key(
            tenant_id="tenant1",
            name="Test Key",
            scopes=[KeyScope.READ],
        )

        validated = validator.validate_key(full_key)
        assert validated is not None
        assert validated.key_id == info.key_id

    def test_validate_key_invalid(self):
        """Should return None for invalid key."""
        validator = APIKeyValidator()
        result = validator.validate_key("fly_invalid_key")
        assert result is None

    def test_validate_key_revoked(self):
        """Should return None for revoked key."""
        validator = APIKeyValidator()
        full_key, info = validator.create_key(
            tenant_id="tenant1",
            name="Test Key",
            scopes=[KeyScope.READ],
        )

        validator.revoke_key(info.key_id)
        result = validator.validate_key(full_key)
        assert result is None

    def test_validate_key_expired(self):
        """Should return None for expired key."""
        validator = APIKeyValidator()
        full_key, info = validator.create_key(
            tenant_id="tenant1",
            name="Test Key",
            scopes=[KeyScope.READ],
            expires_in_days=-1,  # Already expired
        )

        result = validator.validate_key(full_key)
        assert result is None

    def test_revoke_key(self):
        """Should revoke an existing key."""
        validator = APIKeyValidator()
        _, info = validator.create_key(
            tenant_id="tenant1",
            name="Test Key",
            scopes=[KeyScope.READ],
        )

        assert validator.revoke_key(info.key_id) is True
        retrieved = validator.get_key_info(info.key_id)
        assert retrieved.status == KeyStatus.REVOKED

    def test_revoke_nonexistent_key(self):
        """Revoke should return False for nonexistent key."""
        validator = APIKeyValidator()
        assert validator.revoke_key("nonexistent") is False

    def test_get_key_info(self):
        """Should return key info for existing key."""
        validator = APIKeyValidator()
        _, info = validator.create_key(
            tenant_id="tenant1",
            name="Test Key",
            scopes=[KeyScope.READ],
        )

        retrieved = validator.get_key_info(info.key_id)
        assert retrieved is not None
        assert retrieved.name == "Test Key"

    def test_get_key_info_nonexistent(self):
        """Should return None for nonexistent key."""
        validator = APIKeyValidator()
        assert validator.get_key_info("nonexistent") is None

    def test_list_keys_for_tenant(self):
        """Should list all keys for a tenant."""
        validator = APIKeyValidator()
        validator.create_key(tenant_id="t1", name="Key1", scopes=[KeyScope.READ])
        validator.create_key(tenant_id="t1", name="Key2", scopes=[KeyScope.WRITE])
        validator.create_key(tenant_id="t2", name="Key3", scopes=[KeyScope.READ])

        t1_keys = validator.list_keys("t1")
        assert len(t1_keys) == 2
        assert all(k.tenant_id == "t1" for k in t1_keys)

    def test_list_keys_empty_tenant(self):
        """Should return empty list for tenant with no keys."""
        validator = APIKeyValidator()
        keys = validator.list_keys("nonexistent")
        assert keys == []

    def test_get_stats(self):
        """Should return validator statistics."""
        validator = APIKeyValidator()
        validator.create_key(tenant_id="t1", name="Key1", scopes=[KeyScope.READ])

        stats = validator.get_stats()
        assert stats["total_keys"] == 1
        assert stats["total_validations"] == 0
        assert stats["failed_validations"] == 0

    def test_validation_updates_stats(self):
        """Validation should update statistics."""
        validator = APIKeyValidator()
        full_key, _ = validator.create_key(tenant_id="t1", name="Key1", scopes=[KeyScope.READ])

        validator.validate_key(full_key)
        validator.validate_key("fly_invalid")

        stats = validator.get_stats()
        assert stats["total_validations"] == 2
        assert stats["failed_validations"] == 1

    def test_validation_updates_request_count(self):
        """Validation should update request count on key info."""
        validator = APIKeyValidator()
        full_key, info = validator.create_key(tenant_id="t1", name="Key1", scopes=[KeyScope.READ])

        validator.validate_key(full_key)
        validator.validate_key(full_key)

        updated_info = validator.get_key_info(info.key_id)
        assert updated_info.request_count == 2

    def test_key_expiration_in_days(self):
        """Should set correct expiration from expires_in_days."""
        validator = APIKeyValidator()
        _, info = validator.create_key(
            tenant_id="t1",
            name="Key1",
            scopes=[KeyScope.READ],
            expires_in_days=30,
        )

        assert info.expires_at is not None
        assert info.expires_at > datetime.utcnow()
        assert info.expires_at < datetime.utcnow() + timedelta(days=31)

    def test_custom_rate_limit(self):
        """Should accept custom rate limit."""
        validator = APIKeyValidator()
        _, info = validator.create_key(
            tenant_id="t1",
            name="Key1",
            scopes=[KeyScope.READ],
            rate_limit=120,
        )

        assert info.rate_limit == 120


class TestGetAPIKeyValidator:
    """Tests for get_api_key_validator singleton."""

    def test_returns_same_instance(self):
        """Should return the same instance on multiple calls."""
        import src.security.auth as auth_module

        auth_module._api_key_validator = None

        v1 = get_api_key_validator()
        v2 = get_api_key_validator()
        assert v1 is v2

    def test_default_key_created(self):
        """Should create a default development key."""
        import src.security.auth as auth_module

        auth_module._api_key_validator = None

        validator = get_api_key_validator()
        keys = validator.list_keys("default")
        assert len(keys) >= 1


class TestKeyScope:
    """Tests for KeyScope enum."""

    def test_scopes_exist(self):
        """Expected scopes should be defined."""
        expected = [
            "READ",
            "WRITE",
            "ADMIN",
            "FULL",
            "EMBED_READ",
            "EMBED_WRITE",
            "EMBED_ADMIN",
            "RAG_READ",
            "CHAT_WRITE",
            "CHAT_READ",
        ]
        for scope_name in expected:
            assert hasattr(KeyScope, scope_name)

    def test_scope_values(self):
        """Scope values should match expected strings."""
        assert KeyScope.READ.value == "read"
        assert KeyScope.EMBED_WRITE.value == "embed:write"
        assert KeyScope.CHAT_WRITE.value == "chat:write"
