"""Collaborative filtering recommendations using ALS matrix factorization."""

from .predictor import RecommendationEngine, get_recommendation_engine

__all__ = ["RecommendationEngine", "get_recommendation_engine"]
