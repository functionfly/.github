"""ARQ worker for FlyMind AI Service background tasks.

This module provides Redis-backed background job processing using ARQ
(Asynchronous Redis Queue) for reliable audit log flushing and other
background tasks that need persistence guarantees.
"""

import json
import logging
from datetime import datetime
from typing import Any, Dict, List, Optional

import psycopg2
from psycopg2.extras import execute_batch
from arq import Actor
from arq.connections import RedisSettings

logger = logging.getLogger(__name__)


class AuditWorker(Actor):
    """ARQ worker for processing audit log flush jobs.

    Uses PostgreSQL batch inserts for efficient audit log persistence.
    Jobs are retried automatically on failure with dead-letter queue support.
    """

    max_retries = 3
    retry_delay = 10  # seconds

    async def process_audit_batch(
        self,
        events: List[Dict[str, Any]],
    ) -> Dict[str, Any]:
        """Process a batch of audit events.

        Args:
            events: List of audit event dictionaries

        Returns:
            Dictionary with processing results
        """
        if not events:
            return {"status": "empty_batch", "processed": 0}

        db_url = self._get_db_url()
        if not db_url:
            logger.error("No database URL configured for audit batch processing")
            return {"status": "no_db_url", "processed": 0}

        try:
            conn = psycopg2.connect(db_url)
            with conn.cursor() as cur:
                insert_sql = """
                    INSERT INTO ai_audit_logs (
                        timestamp, tenant_id, user_id, api_key_id, operation,
                        function_id, model, dimensions, text_hash, success,
                        status, latency_ms, error_message, client_ip, request_id,
                        token_count, cost_usd, query_hash, chunks_retrieved,
                        sources, cache_hit, metadata
                    ) VALUES (
                        %(timestamp)s, %(tenant_id)s, %(user_id)s, %(api_key_id)s, %(operation)s,
                        %(function_id)s, %(model)s, %(dimensions)s, %(text_hash)s, %(success)s,
                        %(status)s, %(latency_ms)s, %(error_message)s, %(client_ip)s, %(request_id)s,
                        %(token_count)s, %(cost_usd)s, %(query_hash)s, %(chunks_retrieved)s,
                        %(sources)s, %(cache_hit)s, %(metadata)s
                    )
                """
                execute_batch(cur, insert_sql, events)
                conn.commit()

            logger.info(f"Audit batch processed: {len(events)} events")
            return {"status": "success", "processed": len(events)}

        except Exception as e:
            logger.error(f"Audit batch processing failed: {e}")
            raise

        finally:
            if 'conn' in locals():
                conn.close()

    def _get_db_url(self) -> Optional[str]:
        """Get database URL from environment."""
        import os
        return os.getenv("AUDIT_DATABASE_URL") or os.getenv("DATABASE_URL")


async def run_worker():
    """Run the ARQ worker as a standalone process."""
    from arq import run

    redis_url = "redis://localhost:6379"
    import os
    if os.getenv("REDIS_ADDR"):
        redis_url = f"redis://{os.getenv('REDIS_ADDR')}"

    await run_worker(
        {
            "REDIS": redis_url,
            "DATABASE_URL": os.getenv("AUDIT_DATABASE_URL") or os.getenv("DATABASE_URL", ""),
        }
    )


# ARQ worker configuration
async def worker_settings(ctx: dict) -> dict:
    """ARQ worker settings.

    Args:
        ctx: ARQ context dictionary

    Returns:
        Dictionary with worker configuration
    """
    import os

    redis_addr = os.getenv("REDIS_ADDR", "localhost:6379")
    redis_password = os.getenv("REDIS_PASSWORD", "")

    if redis_password:
        redis_url = f"redis://:{redis_password}@{redis_addr}/0"
    else:
        redis_url = f"redis://{redis_addr}/0"

    return {
        "redis_settings": RedisSettings.from_url(redis_url),
        "database_url": os.getenv("AUDIT_DATABASE_URL") or os.getenv("DATABASE_URL"),
        "max_jobs": 10,
        "job_timeout": 60,
    }


# Queue functions that can be enqueued
async def enqueue_audit_batch(
    ctx: dict,
    events: List[Dict[str, Any]],
) -> None:
    """Enqueue an audit batch for background processing.

    Args:
        ctx: ARQ context
        events: List of audit event dictionaries
    """
    await ctx["queue"].enqueue_job(
        "process_audit_batch",
        events,
        _job_id=f"audit-{datetime.utcnow().timestamp()}",
    )
