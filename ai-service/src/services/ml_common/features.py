"""Feature extraction from Redis and Postgres for ML models."""

import json
import logging
from datetime import datetime, timedelta
from typing import List, Optional

import numpy as np
import redis.asyncio as redis

from ...config import settings

logger = logging.getLogger(__name__)


class FeatureExtractor:
    """Extracts features from Redis sorted sets for ML models."""

    def __init__(self):
        self._redis: Optional[redis.Redis] = None

    async def get_redis(self) -> Optional[redis.Redis]:
        if self._redis is None:
            try:
                self._redis = redis.from_url(
                    settings.redis_url, encoding="utf-8", decode_responses=True
                )
                await self._redis.ping()
            except Exception as e:
                logger.warning(f"Redis connection failed: {e}")
                self._redis = None
        return self._redis

    async def get_time_series(
        self, key_prefix: str, entity_id: str, window_hours: int = 168
    ) -> List[float]:
        """Get time-series values from a Redis sorted set.

        Args:
            key_prefix: Redis key prefix
            entity_id: Entity identifier
            window_hours: Time window in hours

        Returns:
            List of values sorted by timestamp
        """
        r = await self.get_redis()
        if not r:
            return []

        try:
            key = f"{key_prefix}{entity_id}"
            min_score = (datetime.utcnow() - timedelta(hours=window_hours)).timestamp()
            items = await r.zrangebyscore(key, min_score, "+inf")

            values = []
            for item in items:
                try:
                    data = json.loads(item)
                    values.append(float(data.get("value", data.get("request_count", 0))))
                except (json.JSONDecodeError, TypeError, ValueError):
                    continue
            return values
        except Exception as e:
            logger.error(f"Failed to get time series for {key_prefix}{entity_id}: {e}")
            return []

    async def get_time_series_with_timestamps(
        self, key_prefix: str, entity_id: str, window_hours: int = 168
    ) -> List[tuple]:
        """Get time-series values with timestamps from Redis.

        Returns:
            List of (timestamp_float, value) tuples sorted by time
        """
        r = await self.get_redis()
        if not r:
            return []

        try:
            key = f"{key_prefix}{entity_id}"
            min_score = (datetime.utcnow() - timedelta(hours=window_hours)).timestamp()
            items = await r.zrangebyscore(key, min_score, "+inf", withscores=True)

            result = []
            for item_json, score in items:
                try:
                    data = json.loads(item_json)
                    value = float(data.get("value", data.get("request_count", 0)))
                    result.append((score, value))
                except (json.JSONDecodeError, TypeError, ValueError):
                    continue
            return sorted(result, key=lambda x: x[0])
        except Exception as e:
            logger.error(f"Failed to get time series with timestamps: {e}")
            return []

    @staticmethod
    def to_hourly_bins(timestamped_values: List[tuple], hours: int = 168) -> np.ndarray:
        """Convert timestamped values to hourly bins.

        Args:
            timestamped_values: List of (timestamp_float, value) tuples
            hours: Number of hours of history

        Returns:
            numpy array of hourly aggregated values
        """
        if not timestamped_values:
            return np.zeros(hours)

        now = datetime.utcnow().timestamp()
        bins = np.zeros(hours)

        for ts, val in timestamped_values:
            age_hours = int((now - ts) / 3600)
            if 0 <= age_hours < hours:
                bins[hours - 1 - age_hours] += val

        return bins
