"""Tests for RAG template SQL security.

Tests for:
- SafeQueryBuilder identifier validation
- Condition validation
- SQL injection prevention
- Dangerous keyword blocking
"""

import pytest

from src.services.generation.rag_retrieval import SafeQueryBuilder


class TestValidateSqlIdentifier:
    """Tests for SafeQueryBuilder.validate_sql_identifier."""

    def test_valid_simple_table_name(self):
        """Simple alphanumeric table names should be accepted."""
        assert SafeQueryBuilder.validate_sql_identifier("users") == "users"
        assert SafeQueryBuilder.validate_sql_identifier("user_data") == "user_data"
        assert SafeQueryBuilder.validate_sql_identifier("Table123") == "Table123"

    def test_valid_with_underscore(self):
        """Table names with underscores should be accepted."""
        assert SafeQueryBuilder.validate_sql_identifier("user_data_table") == "user_data_table"
        assert SafeQueryBuilder.validate_sql_identifier("_private") == "_private"

    def test_valid_wildcard(self):
        """Wildcard '*' should be accepted when allowed."""
        assert SafeQueryBuilder.validate_sql_identifier("*", allow_wildcard=True) == "*"

    def test_empty_identifier_rejected(self):
        """Empty identifiers should be rejected."""
        with pytest.raises(ValueError, match="Identifier cannot be empty"):
            SafeQueryBuilder.validate_sql_identifier("")

    def test_identifier_with_hyphen_rejected(self):
        """Identifiers with hyphens should be rejected."""
        with pytest.raises(ValueError, match="Invalid identifier"):
            SafeQueryBuilder.validate_sql_identifier("user-data")

    def test_identifier_with_dot_rejected(self):
        """Identifiers with dots should be rejected."""
        with pytest.raises(ValueError, match="Invalid identifier"):
            SafeQueryBuilder.validate_sql_identifier("schema.users")

    def test_identifier_with_space_rejected(self):
        """Identifiers with spaces should be rejected."""
        with pytest.raises(ValueError, match="Invalid identifier"):
            SafeQueryBuilder.validate_sql_identifier("user data")

    def test_identifier_with_special_chars_rejected(self):
        """Identifiers with special characters should be rejected."""
        with pytest.raises(ValueError, match="Invalid identifier"):
            SafeQueryBuilder.validate_sql_identifier("users; DROP TABLE users;--")

    def test_drop_keyword_rejected(self):
        """DROP keyword should be rejected."""
        with pytest.raises(ValueError, match="Forbidden identifier"):
            SafeQueryBuilder.validate_sql_identifier("DROP")

    def test_delete_keyword_rejected(self):
        """DELETE keyword should be rejected."""
        with pytest.raises(ValueError, match="Forbidden identifier"):
            SafeQueryBuilder.validate_sql_identifier("DELETE")

    def test_insert_keyword_rejected(self):
        """INSERT keyword should be rejected."""
        with pytest.raises(ValueError, match="Forbidden identifier"):
            SafeQueryBuilder.validate_sql_identifier("INSERT")

    def test_update_keyword_rejected(self):
        """UPDATE keyword should be rejected."""
        with pytest.raises(ValueError, match="Forbidden identifier"):
            SafeQueryBuilder.validate_sql_identifier("UPDATE")

    def test_truncate_keyword_rejected(self):
        """TRUNCATE keyword should be rejected."""
        with pytest.raises(ValueError, match="Forbidden identifier"):
            SafeQueryBuilder.validate_sql_identifier("TRUNCATE")

    def test_alter_keyword_rejected(self):
        """ALTER keyword should be rejected."""
        with pytest.raises(ValueError, match="Forbidden identifier"):
            SafeQueryBuilder.validate_sql_identifier("ALTER")

    def test_create_keyword_rejected(self):
        """CREATE keyword should be rejected."""
        with pytest.raises(ValueError, match="Forbidden identifier"):
            SafeQueryBuilder.validate_sql_identifier("CREATE")

    def test_grant_keyword_rejected(self):
        """GRANT keyword should be rejected."""
        with pytest.raises(ValueError, match="Forbidden identifier"):
            SafeQueryBuilder.validate_sql_identifier("GRANT")

    def test_exec_keyword_rejected(self):
        """EXEC keyword should be rejected."""
        with pytest.raises(ValueError, match="Forbidden identifier"):
            SafeQueryBuilder.validate_sql_identifier("EXEC")

    def test_sql_comment_double_dash_rejected(self):
        """SQL comment '--' should be rejected."""
        with pytest.raises(ValueError, match="Invalid identifier"):
            SafeQueryBuilder.validate_sql_identifier("users--")

    def test_sql_comment_slash_star_rejected(self):
        """SQL comment '/*' should be rejected."""
        with pytest.raises(ValueError, match="Invalid identifier"):
            SafeQueryBuilder.validate_sql_identifier("users/*")

    def test_semicolon_rejected(self):
        """Semicolons should be rejected."""
        with pytest.raises(ValueError, match="Invalid identifier"):
            SafeQueryBuilder.validate_sql_identifier("users;")

    def test_case_insensitive_keyword_blocking(self):
        """Keyword blocking should be case insensitive."""
        with pytest.raises(ValueError, match="Forbidden identifier"):
            SafeQueryBuilder.validate_sql_identifier("drop")

        with pytest.raises(ValueError, match="Forbidden identifier"):
            SafeQueryBuilder.validate_sql_identifier("Drop")

        with pytest.raises(ValueError, match="Forbidden identifier"):
            SafeQueryBuilder.validate_sql_identifier("DRop")


class TestValidateCondition:
    """Tests for SafeQueryBuilder.validate_condition."""

    def test_valid_simple_equality(self):
        """Simple equality conditions should be accepted."""
        assert SafeQueryBuilder.validate_condition("id = ?") == "id = ?"
        assert SafeQueryBuilder.validate_condition("name = ?") == "name = ?"

    def test_valid_not_equal(self):
        """Not equal conditions should be accepted."""
        assert SafeQueryBuilder.validate_condition("status != ?") == "status != ?"
        assert SafeQueryBuilder.validate_condition("active <> ?") == "active <> ?"

    def test_valid_comparison_operators(self):
        """Comparison operators should be accepted."""
        assert SafeQueryBuilder.validate_condition("age > ?") == "age > ?"
        assert SafeQueryBuilder.validate_condition("age >= ?") == "age >= ?"
        assert SafeQueryBuilder.validate_condition("count < ?") == "count < ?"
        assert SafeQueryBuilder.validate_condition("count <= ?") == "count <= ?"

    def test_valid_like(self):
        """LIKE conditions should be accepted."""
        assert SafeQueryBuilder.validate_condition("name LIKE ?") == "name LIKE ?"
        assert SafeQueryBuilder.validate_condition("email ILIKE ?") == "email ILIKE ?"

    def test_valid_with_underscore_in_column(self):
        """Column names with underscores should be accepted."""
        assert SafeQueryBuilder.validate_condition("user_id = ?") == "user_id = ?"
        assert SafeQueryBuilder.validate_condition("created_at > ?") == "created_at > ?"

    def test_empty_condition_rejected(self):
        """Empty conditions should be rejected."""
        with pytest.raises(ValueError, match="Condition cannot be empty"):
            SafeQueryBuilder.validate_condition("")

    def test_whitespace_only_condition_rejected(self):
        """Whitespace-only conditions should be rejected."""
        with pytest.raises(ValueError, match="Condition cannot be empty"):
            SafeQueryBuilder.validate_condition("   ")

    def test_invalid_format_no_parameter(self):
        """Conditions without '?' parameter marker should be rejected."""
        with pytest.raises(ValueError, match="Invalid condition format"):
            SafeQueryBuilder.validate_condition("id = 1")

    def test_invalid_format_multiple_conditions(self):
        """Multiple conditions (AND/OR) should be rejected."""
        with pytest.raises(ValueError, match="Invalid condition format"):
            SafeQueryBuilder.validate_condition("id = ? AND status = ?")

    def test_invalid_format_raw_sql_injection(self):
        """SQL injection attempts should be rejected."""
        with pytest.raises(ValueError, match="Invalid condition format"):
            SafeQueryBuilder.validate_condition("id = 1; DROP TABLE users;--")

    def test_invalid_format_union_attack(self):
        """UNION-based injection should be rejected."""
        with pytest.raises(ValueError, match="Invalid condition format"):
            SafeQueryBuilder.validate_condition("id = ? UNION SELECT * FROM users")

    def test_whitespace_trimmed(self):
        """Condition whitespace should be trimmed."""
        assert SafeQueryBuilder.validate_condition("  id = ?  ") == "id = ?"


class TestBuildSelect:
    """Tests for SafeQueryBuilder.build_select."""

    def test_select_without_condition(self):
        """SELECT without condition should work."""
        sql, params = SafeQueryBuilder.build_select("users")
        assert sql == "SELECT * FROM users"
        assert params == []

    def test_select_with_valid_condition(self):
        """SELECT with valid condition should work."""
        sql, params = SafeQueryBuilder.build_select("users", condition="id = ?")
        assert sql == "SELECT * FROM users WHERE id = ?"
        assert params == []

    def test_select_with_dangerous_table_rejected(self):
        """SELECT with dangerous table name should be rejected."""
        with pytest.raises(ValueError, match="Forbidden identifier"):
            SafeQueryBuilder.build_select("DROP")

    def test_select_with_invalid_condition_rejected(self):
        """SELECT with invalid condition should be rejected."""
        with pytest.raises(ValueError, match="Invalid condition format"):
            SafeQueryBuilder.build_select("users", condition="id = 1")


class TestBuildInsert:
    """Tests for SafeQueryBuilder.build_insert."""

    def test_insert_with_valid_data(self):
        """INSERT with valid data should work."""
        sql, params = SafeQueryBuilder.build_insert("users", {"name": "John", "email": "john@example.com"})
        assert "INSERT INTO users" in sql
        assert "name" in sql
        assert "email" in sql
        assert "VALUES" in sql
        assert params == ["John", "john@example.com"]

    def test_insert_empty_data_rejected(self):
        """INSERT with empty data should be rejected."""
        with pytest.raises(ValueError, match="INSERT data cannot be empty"):
            SafeQueryBuilder.build_insert("users", {})

    def test_insert_with_dangerous_column_rejected(self):
        """INSERT with dangerous column name should be rejected."""
        with pytest.raises(ValueError, match="Forbidden identifier"):
            SafeQueryBuilder.build_insert("users", {"DROP": "value"})

    def test_insert_with_dangerous_table_rejected(self):
        """INSERT with dangerous table name should be rejected."""
        with pytest.raises(ValueError, match="Forbidden identifier"):
            SafeQueryBuilder.build_insert("DELETE", {"name": "John"})


class TestBuildUpdate:
    """Tests for SafeQueryBuilder.build_update."""

    def test_update_with_valid_data_and_condition(self):
        """UPDATE with valid data and condition should work."""
        sql, params = SafeQueryBuilder.build_update("users", {"name": "Jane"}, "id = ?")
        assert "UPDATE users SET" in sql
        assert "name = %s" in sql
        assert "WHERE id = ?" in sql

    def test_update_empty_data_rejected(self):
        """UPDATE with empty data should be rejected."""
        with pytest.raises(ValueError, match="UPDATE data cannot be empty"):
            SafeQueryBuilder.build_update("users", {}, "id = ?")

    def test_update_with_invalid_condition_rejected(self):
        """UPDATE with invalid condition should be rejected."""
        with pytest.raises(ValueError, match="Invalid condition format"):
            SafeQueryBuilder.build_update("users", {"name": "Jane"}, "id = 1")


class TestBuildDelete:
    """Tests for SafeQueryBuilder.build_delete."""

    def test_delete_with_valid_condition(self):
        """DELETE with valid condition should work."""
        sql, params = SafeQueryBuilder.build_delete("users", "id = ?")
        assert sql == "DELETE FROM users WHERE id = ?"
        assert params == []

    def test_delete_with_invalid_condition_rejected(self):
        """DELETE with invalid condition should be rejected."""
        with pytest.raises(ValueError, match="Invalid condition format"):
            SafeQueryBuilder.build_delete("users", "id = 1")

    def test_delete_with_dangerous_table_rejected(self):
        """DELETE with dangerous table name should be rejected."""
        with pytest.raises(ValueError, match="Forbidden identifier"):
            SafeQueryBuilder.build_delete("DROP", "id = ?")


class TestSqlInjectionPrevention:
    """Tests to verify SQL injection attacks are blocked."""

    def test_injection_via_table_name(self):
        """SQL injection via table name should be blocked."""
        injection_attempts = [
            "users; DROP TABLE users;--",
            "users/*",
            "users--",
            "UNION SELECT * FROM users",
            "users' OR '1'='1",
        ]
        for attempt in injection_attempts:
            with pytest.raises(ValueError):
                SafeQueryBuilder.validate_sql_identifier(attempt)

    def test_injection_via_condition(self):
        """SQL injection via condition should be blocked."""
        injection_attempts = [
            "id = 1; DROP TABLE users;--",
            "id = 1 UNION SELECT * FROM users",
            "name = 'admin'--",
            "id = 1 OR 1=1",
            "id = ?; DELETE FROM users",
        ]
        for attempt in injection_attempts:
            with pytest.raises(ValueError, match="Invalid condition format"):
                SafeQueryBuilder.validate_condition(attempt)

    def test_injection_via_column_name(self):
        """SQL injection via column name should be blocked."""
        with pytest.raises(ValueError):
            SafeQueryBuilder.build_insert("users", {"name; DROP TABLE users;--": "value"})

    def test_classic_or_injection_blocked(self):
        """Classic OR-based injection should be blocked."""
        with pytest.raises(ValueError, match="Invalid condition format"):
            SafeQueryBuilder.validate_condition("id = ? OR 1=1")

    def test_classic_union_injection_blocked(self):
        """Classic UNION-based injection should be blocked."""
        with pytest.raises(ValueError, match="Invalid condition format"):
            SafeQueryBuilder.validate_condition("id = ? UNION ALL SELECT password FROM users")

    def test_comment_based_injection_blocked(self):
        """Comment-based injection should be blocked."""
        with pytest.raises(ValueError):
            SafeQueryBuilder.validate_condition("id = ? -- comment")

    def test_function_based_injection_blocked(self):
        """Function-based injection attempts should be blocked."""
        with pytest.raises(ValueError):
            SafeQueryBuilder.validate_sql_identifier("SLEEP(1)")
