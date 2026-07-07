"""Enhanced time-series forecasting with Holt-Winters exponential smoothing.

Replaces the simple moving average + linear regression with proper
seasonality-aware forecasting.
"""

import json
import logging
from datetime import datetime, timedelta
from typing import List, Optional, Tuple

import numpy as np
import redis.asyncio as redis

from ...config import settings
from ...models.schemas import PredictionRequest
from .models import Prediction, RequestDataPoint

logger = logging.getLogger(__name__)


class HoltWintersForecaster:
    """Holt-Winters triple exponential smoothing for demand forecasting.

    Supports additive seasonality with configurable period (default 24h).
    Falls back to simple exponential smoothing when insufficient data.

    Tenant-isolated: all Redis keys are namespaced by tenant.
    """

    HISTORY_KEY_PREFIX = "ml:prewarming:history:"
    MODEL_KEY_PREFIX = "ml:prewarming:model:"
    HISTORY_EXPIRY_HOURS = 168  # 7 days

    def __init__(self):
        self._redis: Optional[redis.Redis] = None
        self._seasonality = settings.ml_prewarm_seasonality_periods
        self._prediction_window = settings.prewarming_window_minutes
        self._model_cache = {}

    def _make_history_key(self, tenant_id: str, function_id: str) -> str:
        """Create tenant-isolated Redis key for prewarming history."""
        return f"{self.HISTORY_KEY_PREFIX}{tenant_id}:{function_id}"

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

    async def record_request(self, tenant_id: str, function_id: str, count: int = 1) -> None:
        """Record a request for a function with tenant isolation."""
        r = await self.get_redis()
        if not r:
            return

        try:
            key = self._make_history_key(tenant_id, function_id)
            data = json.dumps({"request_count": count, "timestamp": datetime.utcnow().isoformat()})
            score = datetime.utcnow().timestamp()
            await r.zadd(key, {data: score})
            await r.expire(key, self.HISTORY_EXPIRY_HOURS * 3600)
        except Exception as e:
            logger.error(f"Failed to record request: {e}")

    async def _get_hourly_counts(self, tenant_id: str, function_id: str, hours: int = 168) -> np.ndarray:
        """Get request counts aggregated into hourly bins for a tenant-function."""
        r = await self.get_redis()
        if not r:
            return np.zeros(hours)

        try:
            key = self._make_history_key(tenant_id, function_id)
            min_score = (datetime.utcnow() - timedelta(hours=hours)).timestamp()
            items = await r.zrangebyscore(key, min_score, "+inf", withscores=True)

            now = datetime.utcnow().timestamp()
            bins = np.zeros(hours)

            for item_json, score in items:
                try:
                    data = json.loads(item_json)
                    age_hours = int((now - score) / 3600)
                    if 0 <= age_hours < hours:
                        bins[hours - 1 - age_hours] += data.get("request_count", 1)
                except Exception:
                    continue

            return bins
        except Exception as e:
            logger.error(f"Failed to get hourly counts: {e}")
            return np.zeros(hours)

    def _holt_winters_additive(
        self,
        series: np.ndarray,
        season_length: int = 24,
        alpha: float = 0.3,
        beta: float = 0.1,
        gamma: float = 0.3,
        n_forecast: int = 1,
    ) -> Tuple[np.ndarray, float]:
        """Holt-Winters additive seasonal method.

        Args:
            series: Time series data
            season_length: Number of periods in a season
            alpha: Level smoothing parameter
            beta: Trend smoothing parameter
            gamma: Seasonal smoothing parameter
            n_forecast: Number of periods to forecast

        Returns:
            Tuple of (forecast_array, residual_std)
        """
        n = len(series)
        if n < season_length * 2:
            # Not enough data for seasonal model — use simple exponential smoothing
            return self._simple_exp_smoothing(series, alpha, n_forecast)

        # Initialize components
        level = np.mean(series[:season_length])
        trend = (np.mean(series[season_length:2 * season_length]) - np.mean(series[:season_length])) / season_length

        # Initialize seasonal components
        seasonal = np.zeros(season_length)
        for i in range(season_length):
            seasonal[i] = series[i] - level

        # Fit the model
        fitted = np.zeros(n)
        residuals = np.zeros(n)

        for t in range(n):
            if t < season_length:
                fitted[t] = level + trend + seasonal[t % season_length]
            else:
                prev_level = level
                level = alpha * (series[t] - seasonal[t % season_length]) + (1 - alpha) * (level + trend)
                trend = beta * (level - prev_level) + (1 - beta) * trend
                seasonal[t % season_length] = gamma * (series[t] - level) + (1 - gamma) * seasonal[t % season_length]
                fitted[t] = level + trend + seasonal[t % season_length]

            residuals[t] = series[t] - fitted[t]

        # Forecast
        forecasts = np.zeros(n_forecast)
        for h in range(n_forecast):
            forecasts[h] = max(0, level + (h + 1) * trend + seasonal[(n + h) % season_length])

        residual_std = max(np.std(residuals[season_length:]), 1.0)
        return forecasts, residual_std

    def _simple_exp_smoothing(
        self, series: np.ndarray, alpha: float = 0.3, n_forecast: int = 1
    ) -> Tuple[np.ndarray, float]:
        """Simple exponential smoothing fallback."""
        if len(series) == 0:
            return np.zeros(n_forecast), 1.0

        level = series[0]
        residuals = []

        for t in range(1, len(series)):
            prev_level = level
            level = alpha * series[t] + (1 - alpha) * level
            residuals.append(series[t] - prev_level)

        forecasts = np.full(n_forecast, max(0, level))
        residual_std = max(np.std(residuals) if residuals else 1.0, 1.0)
        return forecasts, residual_std

    async def predict(
        self, tenant_id: str, request: PredictionRequest
    ) -> Prediction:
        """Generate a demand prediction using Holt-Winters.

        Args:
            tenant_id: Tenant ID for isolation
            request: Prediction request
        """
        function_id = request.function_id
        window_minutes = request.prediction_window_minutes

        # Get hourly request counts
        hourly = await self._get_hourly_counts(tenant_id, function_id, hours=168)

        total_requests = int(np.sum(hourly))

        if total_requests < 10:
            return self._default_prediction(function_id, window_minutes)

        # Determine if we have enough for seasonal model
        has_seasonal = len(hourly) >= self._seasonality * 2 and total_requests > 50

        if has_seasonal:
            forecasts, residual_std = self._holt_winters_additive(
                hourly, season_length=self._seasonality
            )
            method = "holt_winters_additive"
        else:
            forecasts, residual_std = self._simple_exp_smoothing(hourly)
            method = "simple_exponential_smoothing"

        # Convert hourly forecast to window prediction
        hours_in_window = window_minutes / 60.0
        predicted_requests = max(0, int(np.sum(forecasts[:max(1, int(hours_in_window))])))

        # Calculate confidence based on data quantity and residual quality
        data_hours = np.count_nonzero(hourly)
        if data_hours >= 168:
            base_confidence = 0.85
        elif data_hours >= 72:
            base_confidence = 0.7
        elif data_hours >= 24:
            base_confidence = 0.5
        else:
            base_confidence = 0.3

        # Reduce confidence if residuals are high relative to mean
        mean_val = np.mean(hourly[hourly > 0]) if np.any(hourly > 0) else 1.0
        cv = residual_std / max(mean_val, 0.01)
        confidence = max(0.2, base_confidence * (1.0 - min(cv * 0.1, 0.4)))

        # Determine trend
        if len(forecasts) >= 2:
            trend_slope = forecasts[-1] - forecasts[0]
            if trend_slope > mean_val * 0.2:
                trend = "increasing"
            elif trend_slope < -mean_val * 0.2:
                trend = "decreasing"
            else:
                trend = "stable"
        else:
            trend = "stable"

        now = datetime.utcnow()
        return Prediction(
            function_id=function_id,
            predicted_requests=predicted_requests,
            confidence=round(confidence, 2),
            window_start=now,
            window_end=now + timedelta(minutes=window_minutes),
            trend=trend,
        )

    def _default_prediction(self, function_id: str, window_minutes: int) -> Prediction:
        now = datetime.utcnow()
        return Prediction(
            function_id=function_id,
            predicted_requests=0,
            confidence=0.0,
            window_start=now,
            window_end=now + timedelta(minutes=window_minutes),
            trend="stable",
        )

    async def should_prewarm(self, tenant_id: str, function_id: str, threshold: Optional[int] = None) -> bool:
        """Check if a function should be prewarmed.

        Args:
            tenant_id: Tenant ID for isolation
            function_id: Function to check
            threshold: Optional custom threshold
        """
        threshold = threshold or settings.prewarming_threshold
        request = PredictionRequest(
            function_id=function_id,
            prediction_window_minutes=self._prediction_window,
        )
        prediction = await self.predict(tenant_id, request)
        return prediction.predicted_requests >= threshold

    async def close(self):
        if self._redis:
            await self._redis.close()
            self._redis = None


_forecaster: Optional[HoltWintersForecaster] = None


def get_holt_winters_forecaster() -> HoltWintersForecaster:
    global _forecaster
    if _forecaster is None:
        _forecaster = HoltWintersForecaster()
    return _forecaster
