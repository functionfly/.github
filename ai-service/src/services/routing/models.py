"""Data models for routing service.

Contains models for edge metrics, scores, and routing decisions.
"""

import math
from datetime import datetime
from typing import Optional

from ...models.schemas import EdgeProvider, EdgeLocation


class EdgeMetrics:
    """Metrics for an edge provider.

    Attributes:
        provider: The edge provider
        location: Geographic location
        avg_latency_ms: Average latency in milliseconds
        current_load_percent: Current load percentage (0-100)
        available: Whether the edge is available
        last_check: Timestamp of last health check
        sample_count: Number of latency samples collected
    """

    # Default edge locations
    DEFAULT_LOCATIONS = {
        EdgeProvider.CLOUDFLARE: EdgeLocation(
            region="global", country="US", latitude=37.0902, longitude=-95.7129
        ),
        EdgeProvider.VERCEL: EdgeLocation(
            region="global", country="US", latitude=37.0902, longitude=-95.7129
        ),
        EdgeProvider.FLY: EdgeLocation(
            region="global", country="US", latitude=37.0902, longitude=-95.7129
        ),
        EdgeProvider.DENO: EdgeLocation(
            region="global", country="US", latitude=37.0902, longitude=-95.7129
        ),
        EdgeProvider.FUNCTIONFLY: EdgeLocation(
            region="global", country="US", latitude=37.0902, longitude=-95.7129
        ),
    }

    def __init__(
        self,
        provider: EdgeProvider,
        avg_latency_ms: float = 0.0,
        current_load_percent: float = 0.0,
        available: bool = True,
        last_check: Optional[datetime] = None,
        sample_count: int = 0,
    ):
        self.provider = provider
        self.location = self.DEFAULT_LOCATIONS.get(provider)
        self.avg_latency_ms = avg_latency_ms
        self.current_load_percent = current_load_percent
        self.available = available
        self.last_check = last_check or datetime.utcnow()
        self.sample_count = sample_count

    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "provider": self.provider.value,
            "location": self.location.model_dump() if self.location else None,
            "avg_latency_ms": self.avg_latency_ms,
            "current_load_percent": self.current_load_percent,
            "available": self.available,
            "last_check": self.last_check.isoformat(),
            "sample_count": self.sample_count,
        }


class EdgeScore:
    """Calculated score for an edge provider.

    The score is calculated using weighted factors:
    - latency_pct (30%): Lower latency is better
    - load_pct (30%): Lower load is better
    - availability (40%): Higher availability is better

    Attributes:
        provider: The edge provider
        score: Final score (0-1, higher is better)
        latency_score: Score component for latency
        load_score: Score component for load
        availability_score: Score component for availability
        reasoning: Human-readable explanation of the score
    """

    def __init__(
        self,
        provider: EdgeProvider,
        latency_score: float,
        load_score: float,
        availability_score: float,
        latency_weight: float = 0.30,
        load_weight: float = 0.30,
        availability_weight: float = 0.40,
    ):
        self.provider = provider
        self.latency_score = latency_score
        self.load_score = load_score
        self.availability_score = availability_score

        # Calculate weighted final score
        self.score = (
            latency_score * latency_weight
            + load_score * load_weight
            + availability_score * availability_weight
        )

        # Generate reasoning
        self.reasoning = self._generate_reasoning(latency_weight, load_weight, availability_weight)

    def _generate_reasoning(
        self,
        latency_weight: float,
        load_weight: float,
        availability_weight: float,
    ) -> str:
        """Generate human-readable reasoning for the score."""
        reasons = []

        # Check latency contribution
        if self.latency_score >= 0.8:
            reasons.append("excellent latency")
        elif self.latency_score >= 0.6:
            reasons.append("good latency")
        elif self.latency_score >= 0.4:
            reasons.append("moderate latency")
        else:
            reasons.append("high latency")

        # Check load contribution
        if self.load_score >= 0.8:
            reasons.append("low load")
        elif self.load_score >= 0.6:
            reasons.append("moderate load")
        elif self.load_score >= 0.4:
            reasons.append("elevated load")
        else:
            reasons.append("high load")

        # Check availability contribution
        if self.availability_score >= 0.9:
            reasons.append("highly available")
        elif self.availability_score >= 0.7:
            reasons.append("mostly available")
        else:
            reasons.append("reduced availability")

        return f"Selected based on {', '.join(reasons[:2])}"

    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "provider": self.provider.value,
            "score": round(self.score, 3),
            "latency_score": round(self.latency_score, 3),
            "load_score": round(self.load_score, 3),
            "availability_score": round(self.availability_score, 3),
            "reasoning": self.reasoning,
        }


def calculate_distance(
    lat1: float, lon1: float,
    lat2: float, lon2: float
) -> float:
    """Calculate distance between two coordinates using Haversine formula.

    Args:
        lat1: Latitude of first point
        lon1: Longitude of first point
        lat2: Latitude of second point
        lon2: Longitude of second point

    Returns:
        Distance in kilometers
    """
    R = 6371  # Earth's radius in km

    lat1_rad = math.radians(lat1)
    lat2_rad = math.radians(lat2)
    delta_lat = math.radians(lat2 - lat1)
    delta_lon = math.radians(lon2 - lon1)

    a = (
        math.sin(delta_lat / 2) ** 2
        + math.cos(lat1_rad) * math.cos(lat2_rad) * math.sin(delta_lon / 2) ** 2
    )
    c = 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a))

    return R * c


def normalize_value(value: float, min_val: float, max_val: float) -> float:
    """Normalize a value to 0-1 range.

    Args:
        value: The value to normalize
        min_val: Minimum possible value
        max_val: Maximum possible value

    Returns:
        Normalized value between 0 and 1
    """
    if max_val == min_val:
        return 0.5

    normalized = (value - min_val) / (max_val - min_val)
    return max(0.0, min(1.0, normalized))


def exponential_decay(age_seconds: float, half_life_seconds: float = 300) -> float:
    """Calculate exponential decay factor based on age.

    Args:
        age_seconds: Age of the data point in seconds
        half_life_seconds: Half-life for decay (default 5 minutes)

    Returns:
        Decay factor between 0 and 1
    """
    return math.exp(-0.693 * age_seconds / half_life_seconds)
