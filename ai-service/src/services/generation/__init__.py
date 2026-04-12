"""Generation module for cost-optimized AI function generation.

This module provides the complete cost-optimized generation pipeline:
- model_router: Multi-tier model routing
- validation: Syntax/type/security validation
- rag_retrieval: Template and RAG-based retrieval
- cache: Generation caching
- service: Integrated generation service
"""

from .model_router import (
    get_model_router,
    ModelRouter,
    ModelTier,
    RoutingDecision,
    ComplexityAnalyzer,
    ModelConfig,
    TIER_MODELS,
)

from .validation import (
    get_validation_pipeline,
    ValidationPipeline,
    ValidationReport,
    ValidationResult,
    ValidationStage,
)

from .rag_retrieval import (
    get_function_rag_retriever,
    FunctionRAGRetriever,
    FunctionTemplateLibrary,
    RetrievedFunction,
    TemplateMatch,
)

from .cache import (
    get_generation_cache,
    get_cost_tracker,
    GenerationCache,
    GenerationCostTracker,
    CachedGeneration,
)

from .service import (
    get_optimized_generation_service,
    OptimizedGenerationService,
    OptimizedGenerationMetrics,
    GenerationAttempt,
)

__all__ = [
    # Model Router
    "get_model_router",
    "ModelRouter",
    "ModelTier",
    "RoutingDecision",
    "ComplexityAnalyzer",
    "ModelConfig",
    "TIER_MODELS",
    # Validation
    "get_validation_pipeline",
    "ValidationPipeline",
    "ValidationReport",
    "ValidationResult",
    "ValidationStage",
    # RAG Retrieval
    "get_function_rag_retriever",
    "FunctionRAGRetriever",
    "FunctionTemplateLibrary",
    "RetrievedFunction",
    "TemplateMatch",
    # Cache
    "get_generation_cache",
    "get_cost_tracker",
    "GenerationCache",
    "GenerationCostTracker",
    "CachedGeneration",
    # Service
    "get_optimized_generation_service",
    "OptimizedGenerationService",
    "OptimizedGenerationMetrics",
    "GenerationAttempt",
]
