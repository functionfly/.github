"""Pydantic models for cost anomaly detection."""

import uuid
from datetime import datetime
from typing import List, Optional
from pydantic import BaseModel, Field


class CostExecutionMetrics(BaseModel):
    """Metrics for a single function execution."""
    function_id: str
    cost_cents: float
    duration_ms: float
    memory_mb: float
    region: str = "unknown"
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class CostAnomalyResult(BaseModel):
    """Result of a cost anomaly check."""
    is_anomaly: bool
    score: float = Field(ge=0.0, le=1.0)
    anomaly_type: Optional[str] = None  # cost_spike, latency_drift, memory_leak, error_surge
    severity: str = "none"  # none, low, medium, high, critical
    details: str = ""
    function_id: str = ""
    z_score: Optional[float] = None
    threshold: Optional[float] = None
    detected_at: datetime = Field(default_factory=datetime.utcnow)


class CostAnomalySummary(BaseModel):
    """Summary of cost anomalies for a tenant."""
    tenant_id: str
    total_anomalies: int
    anomalies: List[CostAnomalyResult]
    period_hours: int
    generated_at: datetime = Field(default_factory=datetime.utcnow)


class FunctionCostStats(BaseModel):
    """Running statistics for a function's cost."""
    function_id: str
    count: int = 0
    mean: float = 0.0
    m2: float = 0.0  # Welford's online variance accumulator
    min_val: float = float("inf")
    max_val: float = float("-inf")

    @property
    def variance(self) -> float:
        if self.count < 2:
            return 0.0
        return self.m2 / self.count

    @property
    def std(self) -> float:
        return self.variance ** 0.5

    def update(self, value: float) -> None:
        """Update stats using Welford's online algorithm."""
        self.count += 1
        delta = value - self.mean
        self.mean += delta / self.count
        delta2 = value - self.mean
        self.m2 += delta * delta2
        self.min_val = min(self.min_val, value)
        self.max_val = max(self.max_val, value)

    def z_score(self, value: float) -> float:
        if self.std == 0:
            return 0.0
        return (value - self.mean) / self.std
