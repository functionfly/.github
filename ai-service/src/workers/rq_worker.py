"""RQ worker for FlyMind AI Service background tasks.

This module provides Redis-backed background job processing using RQ
(Redis Queue) for reliable audit log flushing and other background tasks
that need persistence guarantees.
"""

import logging
from datetime import datetime
from typing import Any, Dict, List

import psycopg2
from psycopg2.extras import execute_batch
from redis import Redis
from rq import Worker, Queue, connection
from rq.decorators import job

logger = logging.getLogger(__name__)


def get_redis_connection() -> Redis:
    """Get Redis connection from environment."""
    import os
    redis_addr = os.getenv("REDIS_ADDR", "localhost:6379")
    redis_password = os.getenv("REDIS_PASSWORD", "")

    if redis_password:
        redis_url = f"redis://:{redis_password}@{redis_addr}/0"
    else:
        redis_url = f"redis://{redis_addr}/0"

    return Redis.from_url(redis_url)


def get_db_url() -> str:
    """Get database URL from environment."""
    import os
    return os.getenv("AUDIT_DATABASE_URL") or os.getenv("DATABASE_URL", "")


@job("audit", connection_func=get_redis_connection, timeout=60)
def process_audit_batch(events: List[Dict[str, Any]]) -> Dict[str, Any]:
    """Process a batch of audit events.

    Args:
        events: List of audit event dictionaries

    Returns:
        Dictionary with processing results
    """
    if not events:
        return {"status": "empty_batch", "processed": 0}

    db_url = get_db_url()
    if not db_url:
        logger.error("No database URL configured for audit batch processing")
        return {"status": "no_db_url", "processed": 0}

    conn = None
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
        if conn:
            conn.close()


def run_worker():
    """Run the RQ worker as a standalone process."""
    import os
    from rq import Worker, Queue, Connection

    redis_conn = get_redis_connection()

    with Connection(redis_conn):
        worker = Worker(["audit"])
        worker.work()


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    run_worker()