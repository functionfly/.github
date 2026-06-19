"""Audit logging for FlyMind AI Service.

This module provides comprehensive audit logging for embedding operations,
RAG retrieval, and other sensitive AI service operations.
"""

import hashlib
import json
import logging
import os
import threading
import time
from dataclasses import dataclass, field, asdict
from datetime import datetime
from typing import Dict, List, Optional, Any
from enum import Enum

import psycopg2
from psycopg2.extras import execute_batch
from psycopg2 import pool

logger = logging.getLogger(__name__)


class AuditOperation(str, Enum):
    """Types of audit operations."""
    EMBED_GENERATE = "embed_generate"
    EMBED_BATCH_GENERATE = "embed_batch_generate"
    EMBED_SEARCH = "embed_search"
    EMBED_QUERY = "embed_query"
    RAG_RETRIEVE = "rag_retrieve"


class AuditEventStatus(str, Enum):
    """Status of audit events."""
    SUCCESS = "success"
    FAILURE = "failure"
    BLOCKED = "blocked"


@dataclass
class EmbeddingAuditEvent:
    """Audit event for embedding operations.
    
    Stores comprehensive information about embedding operations for
    security auditing, compliance, and debugging.
    """
    timestamp: datetime
    tenant_id: str
    user_id: Optional[str]
    api_key_id: str
    operation: str
    function_id: Optional[str]
    model: str
    dimensions: int
    text_hash: str
    success: bool
    status: str
    latency_ms: float
    error_message: Optional[str]
    client_ip: Optional[str]
    request_id: Optional[str]
    token_count: Optional[int]
    cost_usd: Optional[float]
    metadata: Dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        """Convert event to dictionary for logging."""
        return {
            "timestamp": self.timestamp.isoformat(),
            "tenant_id": self.tenant_id,
            "user_id": self.user_id,
            "api_key_id": self.api_key_id,
            "operation": self.operation,
            "function_id": self.function_id,
            "model": self.model,
            "dimensions": self.dimensions,
            "text_hash": self.text_hash,
            "success": self.success,
            "status": self.status,
            "latency_ms": self.latency_ms,
            "error_message": self.error_message,
            "client_ip": self.client_ip,
            "request_id": self.request_id,
            "token_count": self.token_count,
            "cost_usd": self.cost_usd,
            "metadata": self.metadata,
        }


@dataclass
class RagAuditEvent:
    """Audit event for RAG operations.
    
    Tracks RAG retrieval for security and debugging.
    """
    timestamp: datetime
    tenant_id: str
    user_id: Optional[str]
    api_key_id: str
    query_hash: str
    chunks_retrieved: int
    sources: List[str]
    latency_ms: float
    cache_hit: bool
    success: bool
    status: str
    error_message: Optional[str]
    client_ip: Optional[str]
    request_id: Optional[str]
    metadata: Dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        """Convert event to dictionary for logging."""
        return {
            "timestamp": self.timestamp.isoformat(),
            "tenant_id": self.tenant_id,
            "user_id": self.user_id,
            "api_key_id": self.api_key_id,
            "query_hash": self.query_hash,
            "chunks_retrieved": self.chunks_retrieved,
            "sources": self.sources,
            "latency_ms": self.latency_ms,
            "cache_hit": self.cache_hit,
            "success": self.success,
            "status": self.status,
            "error_message": self.error_message,
            "client_ip": self.client_ip,
            "request_id": self.request_id,
            "metadata": self.metadata,
        }


class AuditLogger:
    """Centralized audit logging for AI Service operations.

    Uses ARQ (Redis-backed queue) for reliable background persistence.
    Falls back to direct DB writes if ARQ is unavailable.
    Failed writes are stored in a retry queue for later retry.
    """

<<<<<<< Updated upstream
    # Maximum buffer size to prevent memory exhaustion
    MAX_BUFFER_SIZE = 1000
    # Maximum events per flush to prevent DB overload
    MAX_FLUSH_BATCH = 100
=======
    MAX_RETRY_QUEUE_SIZE = 1000
>>>>>>> Stashed changes

    def __init__(self):
        self._logger = logging.getLogger("flymind.audit")
        self._lock = threading.Lock()
        self._buffer: List[Dict[str, Any]] = []
        self._buffer_size = 100
<<<<<<< Updated upstream
        self._buffer_flush_interval = 30  # seconds
        self._last_flush_time = time.time()
=======
        self._retry_queue: List[Dict[str, Any]] = []  # Failed events for retry
        self._retry_count: Dict[str, int] = {}  # event_id -> retry count
>>>>>>> Stashed changes
        self._rq_queue = None
        self._use_rq = False
        self._db_pool = None
        self._db_pool_min = 2
        self._db_pool_max = 10

    def _get_rq_queue(self):
        """Get or create RQ queue for background processing."""
        if self._rq_queue is not None:
            return self._rq_queue

        try:
            import os
            import redis
            from rq import Queue

            redis_addr = os.getenv("REDIS_ADDR", "localhost:6379")
            redis_password = os.getenv("REDIS_PASSWORD", "")

            if redis_password:
                redis_url = f"redis://:{redis_password}@{redis_addr}/0"
            else:
                redis_url = f"redis://{redis_addr}/0"

            r = redis.from_url(redis_url)
            self._rq_queue = Queue("audit", connection=r)
            self._use_rq = True
            self._logger.info("Using RQ queue for audit log persistence")
            return self._rq_queue
        except Exception as e:
            self._logger.warning(f"RQ queue unavailable: {e}. Using direct DB writes.")
            self._use_rq = False
            return None
        
    def _hash_text(self, text: str) -> str:
        """Create SHA256 hash of text for audit logging.
        
        We store hashes instead of actual text to protect sensitive data
        while still enabling deduplication and correlation.
        """
        return hashlib.sha256(text.encode("utf-8")).hexdigest()
    
    def log_embedding_event(
        self,
        tenant_id: str,
        api_key_id: str,
        operation: str,
        model: str,
        dimensions: int,
        text: str,
        success: bool,
        latency_ms: float,
        user_id: Optional[str] = None,
        function_id: Optional[str] = None,
        error_message: Optional[str] = None,
        client_ip: Optional[str] = None,
        request_id: Optional[str] = None,
        token_count: Optional[int] = None,
        cost_usd: Optional[float] = None,
        status: str = "success",
        metadata: Optional[Dict[str, Any]] = None,
    ) -> EmbeddingAuditEvent:
        """Log an embedding operation event.
        
        Args:
            tenant_id: Tenant ID
            api_key_id: API key ID used
            operation: Operation type (embed_generate, embed_search, etc.)
            model: Model used for embedding
            dimensions: Embedding dimensions
            text: The text that was embedded (will be hashed)
            success: Whether the operation succeeded
            latency_ms: Operation latency in milliseconds
            user_id: Optional user ID
            function_id: Optional function ID for FlyEmbed operations
            error_message: Error message if operation failed
            client_ip: Client IP address
            request_id: Request correlation ID
            token_count: Number of tokens processed
            cost_usd: Cost in USD
            status: Event status (success, failure, blocked)
            metadata: Additional metadata
            
        Returns:
            The logged audit event
        """
        event = EmbeddingAuditEvent(
            timestamp=datetime.utcnow(),
            tenant_id=tenant_id,
            user_id=user_id,
            api_key_id=api_key_id,
            operation=operation,
            function_id=function_id,
            model=model,
            dimensions=dimensions,
            text_hash=self._hash_text(text),
            success=success,
            status=status,
            latency_ms=latency_ms,
            error_message=error_message,
            client_ip=client_ip,
            request_id=request_id,
            token_count=token_count,
            cost_usd=cost_usd,
            metadata=metadata or {},
        )
        
        # Log to structured logger
        self._logger.info(
            "Embedding audit event",
            extra={
                "audit_event": "embedding",
                "event_data": event.to_dict(),
            }
        )

        # Buffer for batch database writes with size limit check
        with self._lock:
            if len(self._buffer) >= self.MAX_BUFFER_SIZE:
                self._logger.warning("Audit buffer full, forcing flush")
                self._lock.release()
                self._flush_buffer()
                self._lock.acquire()
            self._buffer.append(event.to_dict())

            # Check if we should flush based on size OR time
            should_flush = (
                len(self._buffer) >= self._buffer_size or
                (time.time() - self._last_flush_time) >= self._buffer_flush_interval
            )
            if should_flush:
                self._lock.release()
                self._flush_buffer()
                self._lock.acquire()

        return event
    
    def log_rag_event(
        self,
        tenant_id: str,
        api_key_id: str,
        query: str,
        chunks_retrieved: int,
        sources: List[str],
        latency_ms: float,
        cache_hit: bool,
        success: bool,
        user_id: Optional[str] = None,
        error_message: Optional[str] = None,
        client_ip: Optional[str] = None,
        request_id: Optional[str] = None,
        status: str = "success",
        metadata: Optional[Dict[str, Any]] = None,
    ) -> RagAuditEvent:
        """Log a RAG retrieval event.
        
        Args:
            tenant_id: Tenant ID
            api_key_id: API key ID used
            query: The query text (will be hashed)
            chunks_retrieved: Number of chunks retrieved
            sources: List of document sources
            latency_ms: Operation latency
            cache_hit: Whether result was from cache
            success: Whether operation succeeded
            user_id: Optional user ID
            error_message: Error message if failed
            client_ip: Client IP address
            request_id: Request correlation ID
            status: Event status
            metadata: Additional metadata
            
        Returns:
            The logged audit event
        """
        event = RagAuditEvent(
            timestamp=datetime.utcnow(),
            tenant_id=tenant_id,
            user_id=user_id,
            api_key_id=api_key_id,
            query_hash=self._hash_text(query),
            chunks_retrieved=chunks_retrieved,
            sources=sources,
            latency_ms=latency_ms,
            cache_hit=cache_hit,
            success=success,
            status=status,
            error_message=error_message,
            client_ip=client_ip,
            request_id=request_id,
            metadata=metadata or {},
        )
        
        # Log to structured logger
        self._logger.info(
            "RAG audit event",
            extra={
                "audit_event": "rag",
                "event_data": event.to_dict(),
            }
        )

        # Buffer for batch database writes with size limit check
        with self._lock:
            if len(self._buffer) >= self.MAX_BUFFER_SIZE:
                self._logger.warning("Audit buffer full, forcing flush")
                self._lock.release()
                self._flush_buffer()
                self._lock.acquire()
            self._buffer.append(event.to_dict())

            # Check if we should flush based on size OR time
            should_flush = (
                len(self._buffer) >= self._buffer_size or
                (time.time() - self._last_flush_time) >= self._buffer_flush_interval
            )
            if should_flush:
                self._lock.release()
                self._flush_buffer()
                self._lock.acquire()

        return event
    
    def _get_db_pool(self):
        """Get or create database connection pool with TLS."""
        if self._db_pool is not None:
            return self._db_pool

        db_url = os.getenv("AUDIT_DATABASE_URL") or os.getenv("DATABASE_URL")
        if not db_url:
            return None

        try:
            self._db_pool = pool.ThreadedConnectionPool(
                self._db_pool_min,
                self._db_pool_max,
                dsn=db_url,
                sslmode=os.getenv("DATABASE_SSLMODE", "require"),
            )
            self._logger.info("Database connection pool created with TLS")
            return self._db_pool
        except Exception as e:
            self._logger.error(f"Failed to create database connection pool: {e}")
            return None

    def _flush_buffer(self) -> None:
        """Flush audit buffer to persistent storage.

        Uses ARQ queue for background processing when available.
        Falls back to direct PostgreSQL batch inserts if RQ is unavailable.
<<<<<<< Updated upstream
        Limits batch size to prevent DB overload.
        Uses connection pooling for efficient database access.
=======
        Includes events from retry queue to retry failed writes.
>>>>>>> Stashed changes
        """
        # Combine buffer with retry queue events
        events_to_flush = list(self._buffer)
        if self._retry_queue:
            # Add retry events that haven't exceeded max retry count
            for event in self._retry_queue[:]:
                event_id = self._get_event_id(event)
                if self._retry_count.get(event_id, 0) < 3:  # Max 3 retries
                    events_to_flush.append(event)

        if not events_to_flush:
            return

        # Take a snapshot of the buffer and clear it to prevent unbounded growth
        # Process in batches to avoid memory exhaustion
        events_to_flush = self._buffer[:self.MAX_FLUSH_BATCH]
        remaining_events = self._buffer[self.MAX_FLUSH_BATCH:]

        with self._lock:
            self._buffer = remaining_events

        # Try RQ first for reliable background processing
        if self._use_rq and self._rq_queue:
            try:
                from datetime import datetime
                from ..workers.rq_worker import process_audit_batch

<<<<<<< Updated upstream
                # Enqueue job for background processing using actual function
                job_id = f"audit-{datetime.utcnow().timestamp()}"
                self._rq_queue.enqueue(
                    process_audit_batch,
=======
                job_id = f"audit-{datetime.utcnow().timestamp()}"
                self._rq_queue.enqueue(
                    "src.workers.rq_worker.process_audit_batch",
>>>>>>> Stashed changes
                    events_to_flush,
                    job_id=job_id,
                )
                self._logger.info(f"Enqueued audit batch to RQ: {len(events_to_flush)} events")
<<<<<<< Updated upstream
                self._last_flush_time = time.time()
=======
                self._buffer.clear()
                # Clear retry queue for successfully enqueued events
                self._retry_queue.clear()
                self._retry_count.clear()
>>>>>>> Stashed changes
                return
            except Exception as e:
                self._logger.warning(f"RQ enqueue failed, falling back to direct DB: {e}")

<<<<<<< Updated upstream
        # Fallback: direct PostgreSQL batch insert with connection pooling
        db_pool = self._get_db_pool()
        if not db_pool:
            self._logger.warning("No database connection pool available for audit logging")
=======
        # Fallback: direct PostgreSQL batch insert
        db_url = os.getenv("AUDIT_DATABASE_URL") or os.getenv("DATABASE_URL")
        if not db_url:
            self._logger.warning("No database URL configured for audit logging")
            # Keep events in retry queue for later
            self._add_to_retry_queue(self._buffer)
            self._buffer.clear()
>>>>>>> Stashed changes
            return

        conn = None
        try:
            conn = db_pool.getconn()
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
                execute_batch(cur, insert_sql, events_to_flush)
                conn.commit()
<<<<<<< Updated upstream
                self._logger.info(f"Audit batch written to DB via pool: {len(events_to_flush)} events")
                self._last_flush_time = time.time()
        except Exception as e:
            self._logger.error(f"Failed to flush audit buffer to DB: {e}")
            with self._lock:
                if len(self._buffer) < self.MAX_BUFFER_SIZE:
                    self._buffer = events_to_flush + self._buffer
                else:
                    self._logger.error("Audit buffer full, dropping oldest events")
        finally:
            if conn:
                db_pool.putconn(conn)
=======
                self._logger.info(f"Audit batch written directly to DB: {len(self._buffer)} events")
                # Clear retry count for successfully written events
                for event in self._buffer:
                    event_id = self._get_event_id(event)
                    self._retry_count.pop(event_id, None)
        except Exception as e:
            self._logger.error(f"Failed to flush audit buffer to DB: {e}")
            # Move failed events to retry queue
            self._add_to_retry_queue(self._buffer)
        finally:
            self._buffer.clear()
            if 'conn' in locals():
                conn.close()

    def _get_event_id(self, event: Dict[str, Any]) -> str:
        """Generate a unique ID for an event for retry tracking."""
        import hashlib
        event_str = json.dumps(event, sort_keys=True, default=str)
        return hashlib.sha256(event_str.encode()).hexdigest()[:16]

    def _add_to_retry_queue(self, events: List[Dict[str, Any]]) -> None:
        """Add failed events to retry queue.

        Args:
            events: List of failed audit events
        """
        for event in events:
            event_id = self._get_event_id(event)
            if event_id not in self._retry_count:
                self._retry_count[event_id] = 0
                self._retry_queue.append(event)
            self._retry_count[event_id] += 1

        # Trim retry queue if too large
        while len(self._retry_queue) > self.MAX_RETRY_QUEUE_SIZE:
            removed = self._retry_queue.pop(0)
            removed_id = self._get_event_id(removed)
            self._retry_count.pop(removed_id, None)

        self._logger.warning(f"Added {len(events)} events to retry queue (total: {len(self._retry_queue)})"
>>>>>>> Stashed changes
    
    def flush(self) -> None:
        """Force flush the audit buffer."""
        with self._lock:
            if self._buffer:
                self._flush_buffer()
    
    def get_stats(self) -> Dict[str, int]:
        """Get audit logger statistics.

        Returns:
            Dictionary with stats
        """
        with self._lock:
            return {
                "buffer_size": len(self._buffer),
                "buffer_capacity": self._buffer_size,
                "retry_queue_size": len(self._retry_queue),
                "retry_events_count": len(self._retry_count),
            }


# Global audit logger instance
_audit_logger: Optional[AuditLogger] = None


def get_audit_logger() -> AuditLogger:
    """Get the global audit logger instance.
    
    Returns:
        AuditLogger instance
    """
    global _audit_logger
    if _audit_logger is None:
        _audit_logger = AuditLogger()
    return _audit_logger


def create_embedding_audit_event(
    text: str,
    api_key_info: Any,
    operation: str,
    model: str,
    dimensions: int,
    success: bool,
    latency_ms: float,
    **kwargs
) -> EmbeddingAuditEvent:
    """Convenience function to create and log an embedding audit event.
    
    Args:
        text: The text that was embedded
        api_key_info: APIKeyInfo object with tenant_id and key_id
        operation: Operation type
        model: Model used
        dimensions: Embedding dimensions
        success: Whether operation succeeded
        latency_ms: Operation latency
        **kwargs: Additional event fields
        
    Returns:
        The logged audit event
    """
    logger = get_audit_logger()
    
    return logger.log_embedding_event(
        tenant_id=getattr(api_key_info, 'tenant_id', 'unknown'),
        api_key_id=getattr(api_key_info, 'key_id', 'unknown'),
        operation=operation,
        model=model,
        dimensions=dimensions,
        text=text,
        success=success,
        latency_ms=latency_ms,
        **kwargs
    )
