"""Thompson Sampling edge routing — replaces weighted heuristics."""

from .selector import ThompsonSamplingRouter, get_thompson_router

__all__ = ["ThompsonSamplingRouter", "get_thompson_router"]
