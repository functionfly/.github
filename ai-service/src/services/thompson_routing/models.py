"""Pydantic models for Thompson Sampling routing."""

from datetime import datetime
from typing import Dict, List, Optional
from pydantic import BaseModel, Field


class ArmState(BaseModel):
    """Beta distribution parameters for a single arm (edge provider)."""
    edge: str
    alpha: float = 1.0  # successes
    beta: float = 1.0   # failures
    total_pulls: int = 0
    total_reward: float = 0.0
    avg_latency_ms: float = 0.0
    success_count: int = 0

    @property
    def mean(self) -> float:
        return self.alpha / (self.alpha + self.beta)

    @property
    def avg_reward(self) -> float:
        if self.total_pulls == 0:
            return 0.0
        return self.total_reward / self.total_pulls

    @property
    def success_rate(self) -> float:
        if self.total_pulls == 0:
            return 0.0
        return self.success_count / self.total_pulls


class RoutingOutcome(BaseModel):
    """Outcome of a routed execution — used to update the bandit."""
    edge: str
    function_id: str
    latency_ms: float
    success: bool
    cost_cents: float = 0.0
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class ThompsonDecision(BaseModel):
    """A routing decision made by Thompson Sampling."""
    recommended_edge: str
    confidence: float
    sampled_values: Dict[str, float]
    alternatives: List[str]
    latency_estimate_ms: float
    is_exploration: bool = False
    reasoning: str = ""
