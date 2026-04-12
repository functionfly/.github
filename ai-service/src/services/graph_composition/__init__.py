"""Graph Composition Service for AI-powered backend generation.

Part of the "Backend as a Graph" vision - generates graph topologies from natural language.
"""

from .service import (
    GraphCompositionService,
    get_graph_composition_service,
    CompositionAttempt,
)

__all__ = [
    "GraphCompositionService",
    "get_graph_composition_service",
    "CompositionAttempt",
]
