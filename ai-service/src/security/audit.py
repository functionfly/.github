"""Audit logging for FlyMind AI Service.

This module provides comprehensive audit logging for embedding operations,
RAG retrieval, and other sensitive AI service operations.
"""

import hashlib
import json
import logging
import os
import threading
from dataclasses import dataclass, field, asdict
from datetime import datetime
from typing import Dict, List, Optional, Any
from enum import Enum

import psycopg2
from psycopg2.extras import execute_batch

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
    """Centralized audit logging for AI Service operations."""
    
    def __init__(self):
        self._logger = logging.getLogger("flymind.audit")
        self._lock = threading.Lock()
        self._buffer: List[Dict[str, Any]] = []
        self._buffer_size = 100
        
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
        
        # Buffer for batch database writes
        with self._lock:
            self._buffer.append(event.to_dict())
            if len(self._buffer) >= self._buffer_size:
                self._flush_buffer()
        
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
        
        # Buffer for batch database writes
        with self._lock:
            self._buffer.append(event.to_dict())
            if len(self._buffer) >= self._buffer_size:
                self._flush_buffer()
        
        return event
    
    def _flush_buffer(self) -> None:
        """Flush audit buffer to persistent storage.
        
        In production, batch inserts to PostgreSQL. Falls back to
        structured logging if DB is unavailable.
        """
        if not self._buffer:
            return

        db_url = os.getenv("AUDIT_DATABASE_URL") or os.getenv("DATABASE_URL")
        if not db_url:
            self._buffer.clear()
            return

        try:
            conn = psycopg2.connect(db_url)
            with conn.cursor() as cur:
                # Batch insert to ai_audit_logs table
                # Assumes table: ai_audit_logs(
                #   id uuid primary key default gen_random_uuid(),
                #   timestamp timestamptz,
                #   tenant_id text,
                #   user_id text,
                #   api_key_id text,
                #   operation text,
                #   function_id text,
                #   model text,
                #   dimensions int,
                #   text_hash text,
                #   success bool,
                #   status text,
                #   latency_ms float,
                #   error_message text,
                #   client_ip inet,
                #   request_id text,
                #   token_count int,
                #   cost_usd decimal(10,6),
                #   query_hash text,
                #   chunks_retrieved int,
                #   sources text[],
                #   cache_hit bool,
                #   metadata jsonb,
                #   created_at timestamptz default now()
                # )
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
                execute_batch(cur, insert_sql, self._buffer)
                conn.commit()
        except Exception as e:
            logger.error(f"Failed to flush audit buffer to DB: {e}")
            # Events remain in structured logs; don't re-raise
        finally:
            self._buffer.clear()
            if 'conn' in locals():
                conn.close()
    
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
