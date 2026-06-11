"""Atlas sampling middleware for configurable tracing.

This module provides sampling decision logic for Atlas observability.
"""

import random
from dataclasses import dataclass
from typing import Optional

from ..config import settings

logger = __import__('logging').getLogger(__name__)


@dataclass
class SamplingConfig:
    """Sampling configuration."""
    rate: float = 1.0
    errors_only: bool = False
    head_percent: float = 100.0
    tail_count: int = 10


@dataclass
class SamplingDecision:
    """Sampling decision result."""
    should_trace: bool
    reason: str
    sample_type: str  # 'head', 'tail', 'error', 'full'


def should_sample(
    config: SamplingConfig,
    is_error: bool = False,
    event_count: int = 0,
    total_events: int = 0,
) -> SamplingDecision:
    """Determine if a request should be traced based on sampling config.

    Args:
        config: Sampling configuration
        is_error: Whether the request resulted in an error
        event_count: Current event count
        total_events: Total expected events

    Returns:
        SamplingDecision with trace/no-trace and reason
    """
    if is_error and config.errors_only:
        return SamplingDecision(True, "error tracing enabled", "error")

    if is_error and not config.errors_only:
        return SamplingDecision(True, "errors always traced", "full")

    if config.rate < 1.0:
        if random.random() > config.rate:
            return SamplingDecision(False, f"sample_rate={config.rate}", "head")

    if config.head_percent < 100:
        if total_events > 0:
            threshold = (config.head_percent / 100) * total_events
            if event_count > threshold:
                return SamplingDecision(False, f"head_percent={config.head_percent}", "head")

    return SamplingDecision(True, "full trace", "full")


def get_sampling_config() -> SamplingConfig:
    """Get sampling config from settings.

    Returns:
        SamplingConfig instance
    """
    return SamplingConfig(
        rate=getattr(settings, 'atlas_sample_rate', 1.0),
        errors_only=getattr(settings, 'atlas_trace_errors_only', False),
        head_percent=getattr(settings, 'atlas_sample_head_percent', 100.0),
        tail_count=getattr(settings, 'atlas_sample_tail_count', 10),
    )


class AtlasSamplingMiddleware:
    """Middleware that applies sampling decisions to Atlas tracing."""

    def __init__(self, atlas_integration, config: Optional[SamplingConfig] = None):
        self.atlas = atlas_integration
        self.config = config or get_sampling_config()

    async def should_trace(self, is_error: bool = False, event_count: int = 0) -> bool:
        """Check if current request should be traced.

        Args:
            is_error: Whether request resulted in error
            event_count: Current event count

        Returns:
            True if should trace
        """
        decision = should_sample(self.config, is_error, event_count)
        return decision.should_trace

    async def wrap_trace(self, func, is_error: bool = False):
        """Wrap function with conditional tracing.

        Args:
            func: Function to wrap
            is_error: Whether request resulted in error

        Returns:
            Function result
        """
        decision = should_sample(self.config, is_error)

        if not decision.should_trace:
            return await func()

        return await self.atlas._record_event("ACTION", {
            "type": "sampled_trace",
            "reason": decision.reason,
            "sample_type": decision.sample_type,
        })