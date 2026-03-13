"""Data models for anomaly detection service.

Contains models for metric data, anomaly records, and detection config.
"""

import uuid
from datetime import datetime
from typing import Optional
from pydantic import BaseModel, Field


class MetricData(BaseModel):
    """Metric data point for analysis.

    Attributes:
        function_id: The function ID
        metric_name: Name of the metric (latency_ms, error_rate, cold_start_rate)
        value: The metric value
        timestamp: When the metric was recorded
    """
    function_id: str
    metric_name: str
    value: float
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class AnomalyRecord(BaseModel):
    """Record of a detected anomaly.

    Attributes:
        id: Unique anomaly ID
        function_id: The function ID
        type: Type of anomaly
        severity: Severity level
        detected_at: When the anomaly was detected
        description: Human-readable description
        metric_name: The metric that triggered detection
        metric_value: The metric value at detection
        threshold: The threshold that was exceeded
        z_score: Z-score if calculated
        acknowledged: Whether the anomaly has been acknowledged
        acknowledged_by: Who acknowledged it
        acknowledged_at: When it was acknowledged
    """
    id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    function_id: str
    type: str  # "latency_spike", "error_rate_increase", "cold_start_spike"
    severity: str  # "low", "medium", "high", "critical"
    detected_at: datetime = Field(default_factory=datetime.utcnow)
    description: str
    metric_name: str
    metric_value: float
    threshold: float
    z_score: Optional[float] = None
    acknowledged: bool = False
    acknowledged_by: Optional[str] = None
    acknowledged_at: Optional[datetime] = None


class AnomalyThresholds(BaseModel):
    """Configurable thresholds for anomaly detection.

    Attributes:
        latency_z_score: Z-score threshold for latency (default 3.0)
        error_rate: Error rate threshold (default 0.01 = 1%)
        cold_start_rate: Cold start rate threshold (default 0.10 = 10%)
        window_minutes: Sliding window size in minutes
        check_interval: Check interval in seconds
    """
    latency_z_score: float = 3.0
    error_rate: float = 0.01
    cold_start_rate: float = 0.10
    window_minutes: int = 5
    check_interval: int = 30


class StatisticalSummary(BaseModel):
    """Statistical summary of a metric window.

    Attributes:
        count: Number of samples
        mean: Mean value
        std: Standard deviation
        min: Minimum value
        max: Maximum value
    """
    count: int
    mean: float
    std: float
    min: float
    max: float


def calculate_mean(values: list[float]) -> float:
    """Calculate mean of values.

    Args:
        values: List of numeric values

    Returns:
        Mean value
    """
    if not values:
        return 0.0
    return sum(values) / len(values)


def calculate_std(values: list[float], mean: Optional[float] = None) -> float:
    """Calculate standard deviation.

    Args:
        values: List of numeric values
        mean: Pre-calculated mean (optional)

    Returns:
        Standard deviation
    """
    if len(values) < 2:
        return 0.0

    mean = mean or calculate_mean(values)
    variance = sum((x - mean) ** 2 for x in values) / len(values)
    return variance ** 0.5


def calculate_z_score(value: float, mean: float, std: float) -> float:
    """Calculate Z-score.

    Args:
        value: The value to score
        mean: Mean of the distribution
        std: Standard deviation

    Returns:
        Z-score
    """
    if std == 0:
        return 0.0
    return (value - mean) / std


def determine_severity(z_score: float) -> str:
    """Determine severity based on Z-score.

    Args:
        z_score: The Z-score

    Returns:
        Severity level: "low", "medium", "high", "critical"
    """
    abs_z = abs(z_score)

    if abs_z < 2.5:
        return "low"
    elif abs_z < 3.5:
        return "medium"
    elif abs_z < 5.0:
        return "high"
    else:
        return "critical"
