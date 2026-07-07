"""Distributed locking for ML training jobs.

Prevents concurrent training jobs for the same tenant, ensuring
idempotency and preventing resource contention.
"""

import asyncio
import logging
import time
import uuid
from contextlib import asynccontextmanager
from dataclasses import dataclass
from typing import AsyncGenerator, Optional

from ...config import settings

logger = logging.getLogger(__name__)


@dataclass
class LockResult:
    """Result of a lock acquisition attempt."""
    acquired: bool
    lock_id: Optional[str] = None
    held_by: Optional[str] = None
    ttl_seconds: Optional[int] = None
    retry_after: Optional[float] = None


class DistributedTrainingLock:
    """Distributed lock for ML training operations.

    Uses Redis for distributed locking to prevent concurrent training
    jobs for the same tenant across multiple service instances.
    """

    LOCK_PREFIX = "flymind:training:lock:"

    def __init__(self, redis_client):
        """Initialize with Redis client.

        Args:
            redis_client: Redis client instance
        """
        self._redis = redis_client
        self._lock_timeout = 600
        self._retry_delay = 5.0
        self._max_retries = 3

    async def acquire(
        self,
        tenant_id: str,
        service: str,
        lock_timeout: Optional[int] = None,
        idempotency_key: Optional[str] = None,
    ) -> LockResult:
        """Attempt to acquire a training lock.

        Args:
            tenant_id: Tenant ID requesting the lock
            service: ML service name (e.g., 'recommendations', 'cost_anomaly')
            lock_timeout: Optional custom timeout in seconds
            idempotency_key: Optional idempotency key to prevent duplicate training

        Returns:
            LockResult with acquisition status
        """
        if self._redis is None:
            logger.warning("Redis not available, using local lock")
            return LockResult(acquired=True, lock_id="local")

        lock_key = f"{self.LOCK_PREFIX}{tenant_id}:{service}"
        lock_value = str(uuid.uuid4())
        timeout = lock_timeout or self._lock_timeout

        try:
            acquired = await self._redis.set(
                lock_key,
                lock_value,
                ex=timeout,
            )

            if acquired:
                logger.info(f"Acquired training lock for {tenant_id}/{service}: {lock_value}")
                return LockResult(
                    acquired=True,
                    lock_id=lock_value,
                    ttl_seconds=timeout,
                )

            existing = await self._redis.get(lock_key)
            if existing and idempotency_key:
                if existing == idempotency_key:
                    logger.info(
                        f"Training lock already held with same idempotency key "
                        f"for {tenant_id}/{service}"
                    )
                    return LockResult(
                        acquired=True,
                        lock_id=existing,
                        held_by="self",
                        ttl_seconds=timeout,
                    )

            ttl = await self._redis.ttl(lock_key)
            retry_after = max(ttl, 5) if ttl > 0 else self._retry_delay

            logger.info(
                f"Training lock held by another process for {tenant_id}/{service}, "
                f"retry after {retry_after}s"
            )
            return LockResult(
                acquired=False,
                held_by=existing,
                ttl_seconds=ttl if ttl > 0 else None,
                retry_after=retry_after,
            )

        except Exception as e:
            logger.error(f"Failed to acquire training lock: {e}")
            return LockResult(acquired=False)

    async def release(
        self,
        tenant_id: str,
        service: str,
        lock_id: Optional[str] = None,
    ) -> bool:
        """Release a training lock.

        Args:
            tenant_id: Tenant ID
            service: ML service name
            lock_id: Optional lock ID to verify ownership

        Returns:
            True if released successfully
        """
        if self._redis is None:
            return True

        lock_key = f"{self.LOCK_PREFIX}{tenant_id}:{service}"

        try:
            if lock_id:
                existing = await self._redis.get(lock_key)
                if existing and existing != lock_id:
                    logger.warning(
                        f"Cannot release lock for {tenant_id}/{service}: "
                        f"lock_id mismatch (got {lock_id}, held by {existing})"
                    )
                    return False

            deleted = await self._redis.delete(lock_key)
            if deleted:
                logger.info(f"Released training lock for {tenant_id}/{service}")
            return bool(deleted)

        except Exception as e:
            logger.error(f"Failed to release training lock: {e}")
            return False

    async def extend(
        self,
        tenant_id: str,
        service: str,
        lock_id: str,
        additional_seconds: int = 300,
    ) -> bool:
        """Extend a lock's TTL.

        Args:
            tenant_id: Tenant ID
            service: ML service name
            lock_id: Lock ID to verify ownership
            additional_seconds: Seconds to extend

        Returns:
            True if extended successfully
        """
        if self._redis is None:
            return True

        lock_key = f"{self.LOCK_PREFIX}{tenant_id}:{service}"

        try:
            existing = await self._redis.get(lock_key)
            if not existing or existing != lock_id:
                logger.warning(
                    f"Cannot extend lock for {tenant_id}/{service}: "
                    f"lock_id mismatch"
                )
                return False

            current_ttl = await self._redis.ttl(lock_key)
            new_ttl = max(current_ttl, 0) + additional_seconds

            await self._redis.expire(lock_key, new_ttl)
            logger.info(f"Extended training lock for {tenant_id}/{service} by {additional_seconds}s")
            return True

        except Exception as e:
            logger.error(f"Failed to extend training lock: {e}")
            return False

    @asynccontextmanager
    async def hold(
        self,
        tenant_id: str,
        service: str,
        idempotency_key: Optional[str] = None,
        lock_timeout: Optional[int] = None,
    ) -> AsyncGenerator[Optional[str], None]:
        """Context manager to hold a training lock.

        Usage:
            async with training_lock.hold(tenant_id, "recommendations") as lock_id:
                if lock_id:
                    await train_model()
                else:
                    raise HTTPException(status_code=409, detail="Training already in progress")

        Args:
            tenant_id: Tenant ID
            service: ML service name
            idempotency_key: Optional idempotency key
            lock_timeout: Optional custom timeout

        Yields:
            Lock ID if acquired, None otherwise
        """
        result = await self.acquire(
            tenant_id, service,
            lock_timeout=lock_timeout,
            idempotency_key=idempotency_key,
        )

        try:
            yield result.lock_id if result.acquired else None
        finally:
            if result.acquired and result.lock_id and result.lock_id != "local":
                await self.release(tenant_id, service, result.lock_id)

    def get_lock_status(self, tenant_id: str, service: str) -> dict:
        """Get current lock status.

        Args:
            tenant_id: Tenant ID
            service: ML service name

        Returns:
            Dict with lock status
        """
        lock_key = f"{self.LOCK_PREFIX}{tenant_id}:{service}"

        try:
            if self._redis is None:
                return {"locked": False, "reason": "redis_unavailable"}

            existing = asyncio.create_task(self._redis.get(lock_key))
            ttl = asyncio.create_task(self._redis.ttl(lock_key))

            import asyncio
            existing_val = asyncio.run(existing)
            ttl_val = asyncio.run(ttl)

            if existing_val:
                return {
                    "locked": True,
                    "lock_id": existing_val,
                    "ttl_seconds": ttl_val if ttl_val > 0 else None,
                }
            return {"locked": False}

        except Exception as e:
            logger.error(f"Failed to get lock status: {e}")
            return {"locked": False, "error": str(e)}


_training_lock: Optional[DistributedTrainingLock] = None


async def get_training_lock() -> DistributedTrainingLock:
    """Get the global training lock instance."""
    global _training_lock
    if _training_lock is None:
        from ..redis_client import get_redis_client
        redis_client = get_redis_client()
        _training_lock = DistributedTrainingLock(redis_client)
    return _training_lock


class TrainingJobIdempotencyTracker:
    """Tracks training job IDs to ensure idempotency.

    Prevents duplicate training jobs from being accepted within
    a configurable time window.
    """

    JOB_PREFIX = "flymind:training:job:"

    def __init__(self, redis_client):
        self._redis = redis_client
        self._job_ttl = 7200

    async def is_duplicate(
        self,
        tenant_id: str,
        service: str,
        job_id: str,
    ) -> bool:
        """Check if a job is a duplicate.

        Args:
            tenant_id: Tenant ID
            service: ML service name
            job_id: Unique job identifier

        Returns:
            True if this job was already processed
        """
        if self._redis is None:
            return False

        key = f"{self.JOB_PREFIX}{tenant_id}:{service}:{job_id}"

        try:
            exists = await self._redis.get(key)
            return bool(exists)
        except Exception:
            return False

    async def mark_completed(
        self,
        tenant_id: str,
        service: str,
        job_id: str,
        result: bool = True,
    ) -> None:
        """Mark a job as completed.

        Args:
            tenant_id: Tenant ID
            service: ML service name
            job_id: Unique job identifier
            result: Whether training succeeded
        """
        if self._redis is None:
            return

        key = f"{self.JOB_PREFIX}{tenant_id}:{service}:{job_id}"

        try:
            await self._redis.set(
                key,
                "completed" if result else "failed",
                ex=self._job_ttl,
            )
        except Exception as e:
            logger.error(f"Failed to mark job completed: {e}")


_idempotency_tracker: Optional[TrainingJobIdempotencyTracker] = None


async def get_idempotency_tracker() -> TrainingJobIdempotencyTracker:
    """Get the global idempotency tracker instance."""
    global _idempotency_tracker
    if _idempotency_tracker is None:
        from ..redis_client import get_redis_client
        redis_client = get_redis_client()
        _idempotency_tracker = TrainingJobIdempotencyTracker(redis_client)
    return _idempotency_tracker
