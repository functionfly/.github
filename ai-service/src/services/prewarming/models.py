"""Data models for prewarming service.

Contains models for predictions, triggers, and forecasting data.
"""

from datetime import datetime, timedelta
from typing import Optional
from pydantic import BaseModel, Field


class Prediction(BaseModel):
    """Prediction for function demand.

    Attributes:
        function_id: The function ID
        predicted_requests: Predicted number of requests in the window
        confidence: Confidence level (0-1)
        window_start: Start of prediction window
        window_end: End of prediction window
        trend: Trend direction: "increasing", "decreasing", "stable"
        generated_at: When the prediction was generated
    """
    function_id: str
    predicted_requests: int
    confidence: float = Field(ge=0.0, le=1.0)
    window_start: datetime
    window_end: datetime
    trend: str  # "increasing", "decreasing", "stable"
    generated_at: datetime = Field(default_factory=datetime.utcnow)


class PrewarmTrigger(BaseModel):
    """Trigger for prewarming a function.

    Attributes:
        function_id: The function ID
        instances: Number of instances to warm
        edge: Optional edge to warm
        triggered_at: When the trigger was created
        status: Current status of the prewarming
    """
    function_id: str
    instances: int = Field(default=1, ge=1, le=10)
    edge: Optional[str] = None
    triggered_at: datetime = Field(default_factory=datetime.utcnow)
    status: str = "pending"  # "pending", "warming", "complete", "failed"


class RequestDataPoint(BaseModel):
    """Historical request data point.

    Attributes:
        function_id: The function ID
        request_count: Number of requests
        timestamp: When the requests occurred
    """
    function_id: str
    request_count: int
    timestamp: datetime


class ForecastResult(BaseModel):
    """Result of a forecast calculation.

    Attributes:
        prediction: The prediction object
        historical_avg: Average historical requests
        trend_slope: Slope of the trend line
        seasonal_factor: Any seasonal patterns detected
    """
    prediction: Prediction
    historical_avg: float
    trend_slope: float
    seasonal_factor: Optional[float] = None


def calculate_simple_moving_average(
    data_points: list[RequestDataPoint],
    window_size: int = 5,
) -> float:
    """Calculate simple moving average.

    Args:
        data_points: List of request data points
        window_size: Number of points to average

    Returns:
        Simple moving average
    """
    if not data_points:
        return 0.0

    recent = data_points[-window_size:]
    total = sum(dp.request_count for dp in recent)
    return total / len(recent)


def calculate_trend(
    data_points: list[RequestDataPoint],
) -> float:
    """Calculate trend slope using linear regression.

    Args:
        data_points: List of request data points (should be ordered by time)

    Returns:
        Slope of trend line (positive = increasing)
    """
    if len(data_points) < 2:
        return 0.0

    # Convert timestamps to relative minutes
    base_time = data_points[0].timestamp
    x_values = [(dp.timestamp - base_time).total_seconds() / 60 for dp in data_points]
    y_values = [dp.request_count for dp in data_points]

    n = len(x_values)
    if n == 0:
        return 0.0

    # Calculate means
    x_mean = sum(x_values) / n
    y_mean = sum(y_values) / n

    # Calculate slope
    numerator = sum((x_values[i] - x_mean) * (y_values[i] - y_mean) for i in range(n))
    denominator = sum((x_values[i] - x_mean) ** 2 for i in range(n))

    if denominator == 0:
        return 0.0

    return numerator / denominator


def determine_trend_direction(slope: float) -> str:
    """Determine trend direction from slope.

    Args:
        slope: The trend slope

    Returns:
        "increasing", "decreasing", or "stable"
    """
    if slope > 0.5:
        return "increasing"
    elif slope < -0.5:
        return "decreasing"
    else:
        return "stable"


def calculate_confidence(
    data_points: list[RequestDataPoint],
    trend_slope: float,
) -> float:
    """Calculate confidence in the prediction.

    Args:
        data_points: Historical data points
        trend_slope: Calculated trend slope

    Returns:
        Confidence score (0-1)
    """
    if len(data_points) < 3:
        return 0.3  # Low confidence with little data
    if len(data_points) < 5:
        return 0.5  # Medium confidence
    if len(data_points) < 10:
        return 0.7  # Good confidence

    # Higher confidence with more data, reduced by volatile trends
    base_confidence = 0.8

    # Reduce confidence if trend is very steep
    trend_factor = min(1.0, abs(trend_slope) / 5.0)

    return max(0.5, base_confidence - trend_factor * 0.2)
