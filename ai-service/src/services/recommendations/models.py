"""Pydantic models for recommendations."""

from datetime import datetime
from typing import List, Optional
from pydantic import BaseModel, Field


class InteractionEvent(BaseModel):
    """A user-function interaction."""
    user_id: str
    function_id: str
    interaction_type: str  # view, install, execute, rate, search_impression, search_click
    context: Optional[dict] = None
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class RecommendationResult(BaseModel):
    """A single recommendation."""
    function_id: str
    score: float = Field(ge=0.0, le=1.0)
    reason: str = ""
    strategy: str = "collaborative_filtering"


class RecommendationResponse(BaseModel):
    """Response with personalized recommendations."""
    user_id: str
    recommendations: List[RecommendationResult]
    strategy: str
    generated_at: datetime = Field(default_factory=datetime.utcnow)
