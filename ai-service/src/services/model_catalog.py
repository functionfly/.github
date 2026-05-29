"""FlyMind model catalog API helpers.

Re-exports from providers.model_registry for service-layer imports.
"""

from ..providers.model_registry import CURATED_MODELS, build_model_catalog, model_ids_for_provider

__all__ = ["CURATED_MODELS", "build_model_catalog", "model_ids_for_provider"]
