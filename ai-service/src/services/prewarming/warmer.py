"""Prewarming trigger service.

Triggers prewarming based on predictions from the forecaster.
"""

import json
import logging
from datetime import datetime
from typing import Optional

import redis.asyncio as redis

from ...config import settings
from ...integrations.orchestrator.client import get_orchestrator_client
from ...models.schemas import PrewarmTriggerRequest, PrewarmStatus
from .forecaster import ForecastingService, get_forecasting_service

logger = logging.getLogger(__name__)


class PrewarmingService:
    """Service for triggering function prewarming.

    Uses predictions from the forecasting service to determine
    when to trigger prewarming.
    """

    # Redis key patterns
    TRIGGER_KEY_PREFIX = "prewarming:trigger:"
    TRIGGER_EXPIRY_SECONDS = 3600  # 1 hour

    def __init__(self):
        self._redis: Optional[redis.Redis] = None
        self._forecasting_service = get_forecasting_service()
        self._orchestrator = get_orchestrator_client()
        self._threshold = settings.prewarming_threshold

    async def get_redis(self) -> Optional[redis.Redis]:
        """Get Redis connection."""
        if self._redis is None:
            try:
                self._redis = redis.from_url(
                    settings.redis_url,
                    encoding="utf-8",
                    decode_responses=True,
                )
                await self._redis.ping()
            except Exception as e:
                logger.warning(f"Failed to connect to Redis: {e}")
                self._redis = None
        return self._redis

    async def trigger_prewarming(
        self,
        request: PrewarmTriggerRequest,
    ) -> PrewarmStatus:
        """Trigger prewarming for a function.

        Args:
            request: Prewarm trigger request

        Returns:
            PrewarmStatus with the result
        """
        trigger_id = f"{request.function_id}:{datetime.utcnow().timestamp()}"

        status = PrewarmStatus(
            function_id=request.function_id,
            instances_requested=request.instances,
            instances_warmed=0,
            status="pending",
            triggered_at=datetime.utcnow(),
        )

        try:
            # Update status to warming
            status.status = "warming"

            # Call orchestrator API to trigger actual prewarming
            edge_str = request.edge.value if request.edge else None
            ok = await self._orchestrator.trigger_prewarm(
                request.function_id,
                instances=request.instances,
                edge=edge_str,
            )
            if ok:
                instances_warmed = request.instances
                status.status = "complete"
            else:
                instances_warmed = 0
                status.status = "failed"

            status.instances_warmed = instances_warmed
            status.completed_at = datetime.utcnow()

            # Store the trigger status
            await self._store_trigger_status(trigger_id, status)

            logger.info(
                f"Prewarming triggered for {request.function_id}: "
                f"{instances_warmed}/{request.instances} instances"
            )

        except Exception as e:
            logger.error(f"Failed to trigger prewarming: {e}")
            status.status = "failed"
            await self._store_trigger_status(trigger_id, status)

        return status

    async def _store_trigger_status(
        self,
        trigger_id: str,
        status: PrewarmStatus,
    ) -> None:
        """Store trigger status in Redis.

        Args:
            trigger_id: Unique trigger ID
            status: PrewarmStatus to store
        """
        redis_client = await self.get_redis()
        if not redis_client:
            return

        try:
            key = f"{self.TRIGGER_KEY_PREFIX}{trigger_id}"
            status_json = json.dumps(status.model_dump(), default=str)
            await redis_client.setex(key, self.TRIGGER_EXPIRY_SECONDS, status_json)
        except Exception as e:
            logger.error(f"Failed to store trigger status: {e}")

    async def get_trigger_status(self, trigger_id: str) -> Optional[PrewarmStatus]:
        """Get status of a prewarming trigger.

        Args:
            trigger_id: The trigger ID

        Returns:
            PrewarmStatus if found, None otherwise
        """
        redis_client = await self.get_redis()
        if not redis_client:
            return None

        try:
            key = f"{self.TRIGGER_KEY_PREFIX}{trigger_id}"
            status_json = await redis_client.get(key)

            if status_json:
                data = json.loads(status_json)
                return PrewarmStatus(**data)
        except Exception as e:
            logger.error(f"Failed to get trigger status: {e}")

        return None

    async def auto_prewarm_if_needed(self, function_id: str) -> Optional[PrewarmStatus]:
        """Automatically trigger prewarming if needed based on predictions.

        Args:
            function_id: The function ID

        Returns:
            PrewarmStatus if triggered, None otherwise
        """
        # Check if prewarming is needed
        should_prewarm = await self._forecasting_service.should_prewarm(
            function_id,
            threshold=self._threshold,
        )

        if not should_prewarm:
            return None

        # Trigger prewarming
        request = PrewarmTriggerRequest(
            function_id=function_id,
            instances=settings.prewarming_instances_default,
        )

        return await self.trigger_prewarming(request)

    async def close(self):
        """Close Redis connection."""
        if self._redis:
            await self._redis.close()
            self._redis = None


# Global prewarming service instance
_prewarming_service: Optional[PrewarmingService] = None


def get_prewarming_service() -> PrewarmingService:
    """Get the global prewarming service instance.

    Returns:
        The PrewarmingService instance
    """
    global _prewarming_service
    if _prewarming_service is None:
        _prewarming_service = PrewarmingService()
    return _prewarming_service
