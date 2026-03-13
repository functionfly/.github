"""Time-series forecasting for demand prediction.

Uses simple moving average and trend detection to predict
incoming request volume.
"""

import json
import logging
from datetime import datetime, timedelta
from typing import List, Optional

import redis.asyncio as redis

from ...config import settings
from ...models.schemas import PredictionRequest
from .models import (
    Prediction,
    RequestDataPoint,
    ForecastResult,
    calculate_simple_moving_average,
    calculate_trend,
    determine_trend_direction,
    calculate_confidence,
)

logger = logging.getLogger(__name__)


class ForecastingService:
    """Forecasting service for demand prediction.

    Uses simple moving average + trend detection to predict
    request volume for the next N minutes.
    """

    # Redis key patterns
    HISTORY_KEY_PREFIX = "prewarming:history:"
    HISTORY_EXPIRY_HOURS = 24

    def __init__(self):
        self._redis: Optional[redis.Redis] = None
        self._prediction_window = settings.prewarming_window_minutes

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

    async def record_request(self, function_id: str, count: int = 1) -> None:
        """Record a request for a function.

        Args:
            function_id: The function ID
            count: Number of requests (default 1)
        """
        data_point = RequestDataPoint(
            function_id=function_id,
            request_count=count,
            timestamp=datetime.utcnow(),
        )

        redis_client = await self.get_redis()
        if not redis_client:
            return

        try:
            key = f"{self.HISTORY_KEY_PREFIX}{function_id}"
            data_json = json.dumps(data_point.model_dump(), default=str)

            # Use sorted set with timestamp as score
            score = data_point.timestamp.timestamp()
            await redis_client.zadd(key, {data_json: score})

            # Set expiry
            await redis_client.expire(key, self.HISTORY_EXPIRY_HOURS * 3600)
        except Exception as e:
            logger.error(f"Failed to record request: {e}")

    async def get_historical_data(
        self,
        function_id: str,
        window_minutes: int = 60,
    ) -> List[RequestDataPoint]:
        """Get historical request data for a function.

        Args:
            function_id: The function ID
            window_minutes: Time window in minutes

        Returns:
            List of request data points
        """
        redis_client = await self.get_redis()
        if not redis_client:
            return []

        try:
            key = f"{self.HISTORY_KEY_PREFIX}{function_id}"
            min_score = (datetime.utcnow() - timedelta(minutes=window_minutes)).timestamp()

            data_json = await redis_client.zrangebyscore(key, min_score, "+inf")

            data_points = []
            for item in data_json:
                try:
                    data = json.loads(item)
                    data_points.append(RequestDataPoint(**data))
                except Exception as e:
                    logger.warning(f"Failed to parse data point: {e}")

            # Sort by timestamp
            return sorted(data_points, key=lambda dp: dp.timestamp)
        except Exception as e:
            logger.error(f"Failed to get historical data: {e}")
            return []

    async def predict(
        self,
        request: PredictionRequest,
    ) -> Prediction:
        """Generate a prediction for function demand.

        Args:
            request: Prediction request with function ID and window

        Returns:
            Prediction object
        """
        window_minutes = request.prediction_window_minutes

        # Get historical data
        history = await self.get_historical_data(
            request.function_id,
            window_minutes=window_minutes * 2,  # Get more history for analysis
        )

        if len(history) < 2:
            # Not enough data - return default prediction
            return self._default_prediction(request.function_id, window_minutes)

        # Calculate moving average
        sma = calculate_simple_moving_average(history)

        # Calculate trend
        trend_slope = calculate_trend(history)

        # Determine trend direction
        trend_direction = determine_trend_direction(trend_slope)

        # Calculate confidence
        confidence = calculate_confidence(history, trend_slope)

        # Predict requests
        # Simple prediction: SMA + (trend * window)
        # This gives us the expected value based on historical average plus trend
        predicted_requests = int(sma + (trend_slope * window_minutes / 60))
        predicted_requests = max(0, predicted_requests)  # Can't be negative

        # Generate prediction
        now = datetime.utcnow()
        window_start = now
        window_end = now + timedelta(minutes=window_minutes)

        return Prediction(
            function_id=request.function_id,
            predicted_requests=predicted_requests,
            confidence=confidence,
            window_start=window_start,
            window_end=window_end,
            trend=trend_direction,
        )

    def _default_prediction(
        self,
        function_id: str,
        window_minutes: int,
    ) -> Prediction:
        """Generate a default prediction when no data available.

        Args:
            function_id: The function ID
            window_minutes: Prediction window in minutes

        Returns:
            Default prediction
        """
        now = datetime.utcnow()

        return Prediction(
            function_id=function_id,
            predicted_requests=0,
            confidence=0.0,
            window_start=now,
            window_end=now + timedelta(minutes=window_minutes),
            trend="stable",
        )

    async def should_prewarm(
        self,
        function_id: str,
        threshold: Optional[int] = None,
    ) -> bool:
        """Determine if a function should be prewarmed.

        Args:
            function_id: The function ID
            threshold: Optional threshold override

        Returns:
            True if prewarming is recommended
        """
        threshold = threshold or settings.prewarming_threshold

        request = PredictionRequest(
            function_id=function_id,
            prediction_window_minutes=self._prediction_window,
        )

        prediction = await self.predict(request)

        return prediction.predicted_requests >= threshold

    async def close(self):
        """Close Redis connection."""
        if self._redis:
            await self._redis.close()
            self._redis = None


# Global forecasting service instance
_forecasting_service: Optional[ForecastingService] = None


def get_forecasting_service() -> ForecastingService:
    """Get the global forecasting service instance.

    Returns:
        The ForecastingService instance
    """
    global _forecasting_service
    if _forecasting_service is None:
        _forecasting_service = ForecastingService()
    return _forecasting_service
