"""
Cron Job Function
A scheduled function that runs periodically.
Use cases: Data cleanup, report generation, cache refresh, etc.
"""

import json
import asyncio
from datetime import datetime, timezone, timedelta
from typing import Any


async def fetch(request, env, ctx) -> tuple[str, dict]:
    """
    Handle scheduled execution.

    This function runs automatically based on the schedule defined
    in functionfly.jsonc. It receives a trigger event with timing info.

    Schedule presets:
        "*/5 * * * *"    - Every 5 minutes
        "0 * * * *"      - Every hour
        "0 0 * * *"      - Every day at midnight
        "0 0 * * 1-5"    - Weekdays at midnight
    """
    body = {}
    try:
        body = await request.json()
    except (ValueError, TypeError):
        pass

    trigger = body.get("trigger", "manual")
    timestamp = body.get("timestamp", datetime.now(timezone.utc).isoformat())

    tasks = {}

    cleanup_result = await cleanup_old_records(env)
    tasks["cleanup"] = cleanup_result

    cache_result = await refresh_caches(env)
    tasks["cache"] = cache_result

    health_result = await check_system_health(env)
    tasks["health"] = health_result

    return json.dumps({
        "status": "success",
        "trigger": trigger,
        "executed_at": datetime.now(timezone.utc).isoformat(),
        "message": "Scheduled job completed successfully",
        "tasks": tasks
    }), {
        "headers": {
            "Content-Type": "application/json",
            "X-FunctionFly-Template": "cron-job"
        }
    }


async def cleanup_old_records(env: Any) -> dict:
    """
    Remove records older than retention period.
    Reads RETENTION_DAYS from environment (default: 90 days).
    """
    retention_days = int(env.get("RETENTION_DAYS", "90"))
    cutoff = datetime.now(timezone.utc) - timedelta(days=retention_days)

    deleted_count = 0
    try:
        db = env.get("DB")
        if db is None:
            return {"status": "skipped", "reason": "No database binding"}

        async with db.cursor() as cursor:
            await cursor.execute(
                "DELETE FROM execution_logs WHERE created_at < %s",
                (cutoff,)
            )
            deleted_count = cursor.rowcount
    except Exception as e:
        return {"status": "failed", "error": str(e)}

    return {
        "status": "completed",
        "deleted": deleted_count,
        "retention_days": retention_days,
        "cutoff": cutoff.isoformat()
    }


async def refresh_caches(env: Any) -> dict:
    """
    Refresh frequently accessed data from database into cache.
    """
    refreshed = []
    try:
        cache = env.get("CACHE")
        db = env.get("DB")

        if cache is None or db is None:
            return {"status": "skipped", "reason": "No cache or database binding"}

        async with db.cursor() as cursor:
            await cursor.execute(
                "SELECT key, value FROM config WHERE is_cached = true"
            )
            rows = await cursor.fetchall()

        for key, value in rows:
            await cache.put(key, value, expiration_ttl=3600)
            refreshed.append(key)

    except Exception as e:
        return {"status": "failed", "error": str(e)}

    return {
        "status": "completed",
        "entries_refreshed": len(refreshed),
        "keys": refreshed
    }


async def check_system_health(env: Any) -> dict:
    """
    Perform health checks on dependencies and record metrics.
    """
    checks = {}

    try:
        db = env.get("DB")
        if db is not None:
            try:
                async with db.cursor() as cursor:
                    await cursor.execute("SELECT 1")
                checks["database"] = {"status": "healthy"}
            except Exception as e:
                checks["database"] = {"status": "unhealthy", "error": str(e)}
        else:
            checks["database"] = {"status": "not_configured"}
    except Exception as e:
        checks["database"] = {"status": "error", "error": str(e)}

    try:
        cache = env.get("CACHE")
        if cache is not None:
            try:
                await cache.get("health_check_probe")
                checks["cache"] = {"status": "healthy"}
            except Exception as e:
                checks["cache"] = {"status": "unhealthy", "error": str(e)}
        else:
            checks["cache"] = {"status": "not_configured"}
    except Exception as e:
        checks["cache"] = {"status": "error", "error": str(e)}

    overall = "healthy" if all(
        c.get("status") in ("healthy", "not_configured") for c in checks.values()
    ) else "unhealthy"

    return {
        "status": overall,
        "checks": checks,
        "checked_at": datetime.now(timezone.utc).isoformat()
    }


async def handle_manual(request, env, ctx) -> tuple[str, dict]:
    """Handle manual invocation for testing."""
    return await fetch(request, env, ctx)
