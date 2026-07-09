"""
HTTP API Function
A RESTful API with built-in routing for CRUD operations.
"""

import json
import re
from datetime import datetime, timezone
from typing import Any, Optional


class Database:
    """Database abstraction layer supporting D1 and traditional SQL."""

    def __init__(self, env: Any):
        self.env = env
        self._db = env.get("DB")

    async def init(self) -> None:
        """Initialize database schema if using D1."""
        if self._db is None:
            return

        await self._db.exec("""
            CREATE TABLE IF NOT EXISTS users (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                email TEXT UNIQUE NOT NULL,
                created_at TEXT NOT NULL,
                updated_at TEXT
            )
        """)

    async def list_users(self) -> list[dict]:
        """Fetch all users."""
        if self._db is None:
            return list(get_users_from_memory().values())

        result = await self._db.prepare(
            "SELECT id, name, email, created_at, updated_at FROM users ORDER BY id"
        ).all()
        return [dict(row) for row in result.results]

    async def get_user(self, user_id: int) -> Optional[dict]:
        """Fetch a user by ID."""
        if self._db is None:
            return get_users_from_memory().get(user_id)

        result = await self._db.prepare(
            "SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?"
        ).bind(user_id).first()
        return dict(result) if result else None

    async def create_user(self, name: str, email: str) -> dict:
        """Create a new user."""
        now = datetime.now(timezone.utc).isoformat()

        if self._db is None:
            return create_user_in_memory(name, email, now)

        result = await self._db.prepare(
            "INSERT INTO users (name, email, created_at) VALUES (?, ?, ?) RETURNING id, name, email, created_at, updated_at"
        ).bind(name, email, now).run()
        return dict(result.results[0]) if result.results else None

    async def update_user(self, user_id: int, name: str, email: str) -> Optional[dict]:
        """Update an existing user."""
        now = datetime.now(timezone.utc).isoformat()

        if self._db is None:
            return update_user_in_memory(user_id, name, email, now)

        result = await self._db.prepare(
            "UPDATE users SET name = ?, email = ?, updated_at = ? WHERE id = ? RETURNING id, name, email, created_at, updated_at"
        ).bind(name, email, now, user_id).run()
        return dict(result.results[0]) if result.results else None

    async def delete_user(self, user_id: int) -> bool:
        """Delete a user by ID."""
        if self._db is None:
            return delete_user_from_memory(user_id)

        result = await self._db.prepare(
            "DELETE FROM users WHERE id = ?"
        ).bind(user_id).run()
        return result.success


_memory_db: dict = {}
_memory_counter: int = 1


def get_users_from_memory() -> dict:
    global _memory_db
    return _memory_db


def create_user_in_memory(name: str, email: str, created_at: str) -> dict:
    global _memory_counter, _memory_db
    user = {"id": _memory_counter, "name": name, "email": email, "created_at": created_at}
    _memory_db[_memory_counter] = user
    _memory_counter += 1
    return user


def update_user_in_memory(user_id: int, name: str, email: str, updated_at: str) -> Optional[dict]:
    user = _memory_db.get(user_id)
    if not user:
        return None
    user["name"] = name
    user["email"] = email
    user["updated_at"] = updated_at
    return user


def delete_user_from_memory(user_id: int) -> bool:
    if user_id in _memory_db:
        del _memory_db[user_id]
        return True
    return False


def validate_email(email: str) -> bool:
    pattern = r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"
    return bool(re.match(pattern, email))


def parse_path(url: str) -> tuple[str, Optional[str]]:
    """Parse URL path and optional ID."""
    match = re.match(r"^/users(?:/(\d+))?$", url)
    if match:
        return ("users", match.group(1))
    if url == "/health":
        return ("health", None)
    if url.startswith("/users/"):
        parts = url.split("/")
        if len(parts) >= 3 and parts[2].isdigit():
            return ("user", parts[2])
    return ("unknown", None)


async def fetch(request, env, ctx) -> tuple[str, dict]:
    """
    Handle incoming HTTP requests with RESTful routing.

    Routes:
        GET    /users      - List all users
        GET    /users/{id} - Get user by ID
        POST   /users      - Create a new user
        PUT    /users/{id} - Update user
        DELETE /users/{id} - Delete user
        GET    /health     - Health check

    Environment:
        DB: D1 database binding (optional, falls back to in-memory)
    """
    db = Database(env)
    await db.init()

    url = request.url.path if hasattr(request.url, "path") else request.url.split("?")[0]
    method = request.method

    route, param = parse_path(url)

    if route == "health":
        return json.dumps({
            "status": "healthy",
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "service": "http-api",
            "storage": "database" if db._db else "memory"
        }), {"headers": {"Content-Type": "application/json"}}

    if route == "users":
        if method == "GET":
            users = await db.list_users()
            return json.dumps({"users": users, "count": len(users)}), {
                "headers": {"Content-Type": "application/json"}
            }

        if method == "POST":
            try:
                body = await request.json()
            except (ValueError, TypeError):
                return error_response(400, "Invalid JSON payload")

            name = body.get("name", "").strip()
            email = body.get("email", "").strip()

            if not name:
                return error_response(400, "Name is required")
            if not email:
                return error_response(400, "Email is required")
            if not validate_email(email):
                return error_response(400, "Invalid email format")

            user = await db.create_user(name, email)
            return json.dumps(user), {
                "status": "201",
                "headers": {"Content-Type": "application/json"}
            }

    if route == "user":
        user_id = int(param)

        if method == "GET":
            user = await db.get_user(user_id)
            if not user:
                return error_response(404, "User not found")
            return json.dumps(user), {"headers": {"Content-Type": "application/json"}}

        if method == "PUT":
            user = await db.get_user(user_id)
            if not user:
                return error_response(404, "User not found")

            try:
                body = await request.json()
            except (ValueError, TypeError):
                return error_response(400, "Invalid JSON payload")

            name = body.get("name", "").strip() or None
            email = body.get("email", "").strip() or None

            if email and not validate_email(email):
                return error_response(400, "Invalid email format")

            updated = await db.update_user(
                user_id,
                name or user["name"],
                email or user["email"]
            )
            return json.dumps(updated), {"headers": {"Content-Type": "application/json"}}

        if method == "DELETE":
            deleted = await db.delete_user(user_id)
            if not deleted:
                return error_response(404, "User not found")
            return "", {"status": "204"}

    return error_response(404, "Not found")


def error_response(status: int, message: str) -> tuple[str, dict]:
    """Create an error response tuple."""
    return json.dumps({"error": message}), {
        "status": str(status),
        "headers": {"Content-Type": "application/json"}
    }
