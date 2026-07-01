"""Predictive Cold Start Prewarming Service.

This module provides time-series forecasting for demand prediction
and proactive function instance prewarming.

The default forecaster is now Holt-Winters exponential smoothing,
which provides seasonality-aware predictions.
"""

from .forecaster import ForecastingService, get_forecasting_service
from .holt_winters import HoltWintersForecaster, get_holt_winters_forecaster
from .warmer import PrewarmingService, get_prewarming_service
from .models import Prediction, PrewarmTrigger

__all__ = [
    "ForecastingService",
    "get_forecasting_service",
    "HoltWintersForecaster",
    "get_holt_winters_forecaster",
    "PrewarmingService",
    "get_prewarming_service",
    "Prediction",
    "PrewarmTrigger",
]
