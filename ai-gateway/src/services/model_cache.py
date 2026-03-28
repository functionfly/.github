"""Model Cache Service for ONNX models.

This module provides ONNX model loading, caching, and GPU memory management
with LRU eviction policy.
"""

import asyncio
import gc
import logging
import os
from collections import OrderedDict
from dataclasses import dataclass
from datetime import datetime
from typing import Any, Dict, Optional

logger = logging.getLogger(__name__)


@dataclass
class CachedModel:
    """Represents a cached ONNX model."""

    model_id: str
    model_path: str
    session: Any  # onnxruntime.InferenceSession
    loaded_at: datetime
    last_used: datetime
    size_mb: float
    is_warmed: bool = False


class ModelCache:
    """ONNX model cache with LRU eviction.

    Features:
    - LRU (Least Recently Used) eviction policy
    - GPU memory management
    - Model warmup/preloading
    - Automatic cleanup on memory pressure
    """

    def __init__(
        self,
        cache_dir: str = "/tmp/model-cache",
        max_cache_size_mb: float = 10240.0,
        max_models: int = 10,
    ):
        """Initialize model cache.

        Args:
            cache_dir: Directory for model storage
            max_cache_size_mb: Maximum cache size in MB
            max_models: Maximum number of models to cache
        """
        self._cache_dir = cache_dir
        self._max_cache_size_mb = max_cache_size_mb
        self._max_models = max_models
        self._cache: OrderedDict[str, CachedModel] = OrderedDict()
        self._current_size_mb: float = 0.0
        self._lock = asyncio.Lock()
        self._warmup_tasks: Dict[str, asyncio.Task] = {}

        # Ensure cache directory exists
        os.makedirs(cache_dir, exist_ok=True)

    @property
    def cache_dir(self) -> str:
        """Get cache directory path."""
        return self._cache_dir

    async def get_model(self, model_id: str) -> Optional[CachedModel]:
        """Get model from cache.

        Args:
            model_id: Model identifier (e.g., 'phi-3-mini')

        Returns:
            Cached model or None if not found
        """
        async with self._lock:
            if model_id not in self._cache:
                return None

            # Move to end (most recently used)
            self._cache.move_to_end(model_id)

            model = self._cache[model_id]
            model.last_used = datetime.utcnow()
            return model

    async def load_model(
        self,
        model_id: str,
        model_path: Optional[str] = None,
        warmup: bool = True,
    ) -> CachedModel:
        """Load model into cache.

        Args:
            model_id: Model identifier
            model_path: Path to model file (if None, uses cache_dir/model_id.onnx)
            warmup: Whether to warmup the model after loading

        Returns:
            Loaded model

        Raises:
            FileNotFoundError: If model file not found
            RuntimeError: If model loading fails
        """
        # Check if already cached
        cached = await self.get_model(model_id)
        if cached:
            return cached

        # Determine model path
        if model_path is None:
            model_path = os.path.join(self._cache_dir, f"{model_id}.onnx")

        if not os.path.exists(model_path):
            raise FileNotFoundError(f"Model file not found: {model_path}")

        # Get model size
        size_mb = os.path.getsize(model_path) / (1024 * 1024)

        # Check if we need to evict models
        await self._evict_if_needed(size_mb)

        # Load the model
        try:
            session = await self._create_session(model_path)
            model = CachedModel(
                model_id=model_id,
                model_path=model_path,
                session=session,
                loaded_at=datetime.utcnow(),
                last_used=datetime.utcnow(),
                size_mb=size_mb,
                is_warmed=False,
            )

            async with self._lock:
                self._cache[model_id] = model
                self._current_size_mb += size_mb

            logger.info(
                f"Loaded model {model_id} ({size_mb:.2f} MB). "
                f"Cache size: {self._current_size_mb:.2f} MB"
            )

            # Warmup if requested
            if warmup:
                await self.warmup_model(model_id)

            return model

        except Exception as e:
            logger.error(f"Failed to load model {model_id}: {e}")
            raise RuntimeError(f"Model loading failed: {e}") from e

    async def _create_session(self, model_path: str) -> Any:
        """Create ONNX Runtime inference session.

        Args:
            model_path: Path to ONNX model file

        Returns:
            Inference session
        """
        try:
            import onnxruntime as ort

            # Configure session options
            sess_options = ort.SessionOptions()
            sess_options.graph_optimization_level = (
                ort.GraphOptimizationLevel.ORT_ENABLE_ALL
            )
            sess_options.intra_op_num_threads = 4
            sess_options.inter_op_num_threads = 4

            # Create session (try GPU first, fall back to CPU)
            providers = []
            if ort.get_available_providers().__contains__("CUDAExecutionProvider"):
                providers = ["CUDAExecutionProvider", "CPUExecutionProvider"]
            elif ort.get_available_providers().__contains__("CPUExecutionProvider"):
                providers = ["CPUExecutionProvider"]
            else:
                providers = ["CPUExecutionProvider"]

            session = ort.InferenceSession(
                model_path,
                sess_options=sess_options,
                providers=providers,
            )

            logger.info(f"ONNX Runtime providers: {session.get_providers()}")
            return session

        except ImportError:
            logger.warning("onnxruntime not installed, using mock session")
            return MockInferenceSession(model_path)

    async def warmup_model(self, model_id: str) -> None:
        """Warmup model with dummy inference.

        Args:
            model_id: Model identifier
        """
        async with self._lock:
            if model_id not in self._cache:
                logger.warning(f"Cannot warmup {model_id}: not in cache")
                return

            if self._cache[model_id].is_warmed:
                return

            model = self._cache[model_id]

        try:
            # Warmup is done synchronously
            if hasattr(model.session, "run"):
                logger.info(f"Warming up model {model_id}...")
                input_feed = self._build_dummy_input_feed(model.session)
                output_names = [output.name for output in model.session.get_outputs()]
                model.session.run(output_names, input_feed)

            async with self._lock:
                if model_id in self._cache:
                    self._cache[model_id].is_warmed = True

            logger.info(f"Model {model_id} warmup complete")
        except Exception as e:
            async with self._lock:
                if model_id in self._cache:
                    self._cache[model_id].is_warmed = False
            logger.warning(f"Model warmup failed for {model_id}: {e}")

    def _build_dummy_input_feed(self, session: Any) -> Dict[str, Any]:
        """Build dummy inputs from ONNX session input metadata."""
        try:
            import numpy as np
        except ImportError as e:
            raise RuntimeError("numpy is required for model warmup input generation") from e

        input_feed: Dict[str, Any] = {}
        for tensor in session.get_inputs():
            shape = [self._resolve_warmup_dim(dim) for dim in tensor.shape]
            dtype = self._onnx_dtype_to_numpy_dtype(getattr(tensor, "type", None))

            input_feed[tensor.name] = np.zeros(shape, dtype=dtype)

        if not input_feed:
            raise RuntimeError("session has no input tensors")

        return input_feed

    def _resolve_warmup_dim(self, dim: Any) -> int:
        """Resolve an ONNX shape dimension to a concrete positive integer."""
        if isinstance(dim, int) and dim > 0:
            return dim

        if isinstance(dim, str):
            # Common dynamic axes (batch, sequence, etc.) use minimal warmup shape.
            return 1

        # Covers None, 0, negative values, and unknown symbolic dimension objects.
        return 1

    def _onnx_dtype_to_numpy_dtype(self, onnx_type: Optional[str]) -> Any:
        """Map ONNX Runtime type strings to numpy dtypes."""
        try:
            import numpy as np
        except ImportError as e:
            raise RuntimeError("numpy is required for ONNX dtype mapping") from e

        if not onnx_type:
            return np.float32

        normalized = onnx_type.lower()
        mapping = {
            "tensor(float)": np.float32,
            "tensor(float16)": np.float16,
            "tensor(double)": np.float64,
            "tensor(int8)": np.int8,
            "tensor(int16)": np.int16,
            "tensor(int32)": np.int32,
            "tensor(int64)": np.int64,
            "tensor(uint8)": np.uint8,
            "tensor(uint16)": np.uint16,
            "tensor(uint32)": np.uint32,
            "tensor(uint64)": np.uint64,
            "tensor(bool)": np.bool_,
        }

        return mapping.get(normalized, np.float32)

    async def _evict_if_needed(self, required_size_mb: float) -> None:
        """Evict models if cache is full.

        Args:
            required_size_mb: Required size for new model
        """
        # Check size limit
        while (
            self._current_size_mb + required_size_mb > self._max_cache_size_mb
            or len(self._cache) >= self._max_models
        ) and self._cache:
            # Evict least recently used
            oldest_model_id, oldest_model = self._cache.popitem(last=False)
            self._current_size_mb -= oldest_model.size_mb

            # Clear GPU memory
            if hasattr(oldest_model.session, "end_session"):
                oldest_model.session.end_session()
            elif hasattr(oldest_model.session, "__del__"):
                del oldest_model.session

            logger.info(
                f"Evicted model {oldest_model_id} ({oldest_model.size_mb:.2f} MB). "
                f"Cache size: {self._current_size_mb:.2f} MB"
            )

    async def unload_model(self, model_id: str) -> bool:
        """Unload model from cache.

        Args:
            model_id: Model identifier

        Returns:
            True if unloaded, False if not found
        """
        async with self._lock:
            if model_id not in self._cache:
                return False

            model = self._cache.pop(model_id)
            self._current_size_mb -= model.size_mb

            # Clear GPU memory
            if hasattr(model.session, "end_session"):
                model.session.end_session()
            elif hasattr(model.session, "__del__"):
                del model.session

            logger.info(
                f"Unloaded model {model_id}. Cache size: {self._current_size_mb:.2f} MB"
            )
            return True

    async def clear_cache(self) -> None:
        """Clear all cached models."""
        async with self._lock:
            for model in self._cache.values():
                if hasattr(model.session, "end_session"):
                    model.session.end_session()
                elif hasattr(model.session, "__del__"):
                    del model.session

            self._cache.clear()
            self._current_size_mb = 0.0

        # Force garbage collection
        gc.collect()

        logger.info("Model cache cleared")

    async def get_cache_stats(self) -> Dict[str, Any]:
        """Get cache statistics.

        Returns:
            Dictionary with cache stats
        """
        async with self._lock:
            return {
                "cached_models": len(self._cache),
                "current_size_mb": self._current_size_mb,
                "max_size_mb": self._max_cache_size_mb,
                "max_models": self._max_models,
                "utilization_percent": (
                    self._current_size_mb / self._max_cache_size_mb * 100
                    if self._max_cache_size_mb > 0
                    else 0
                ),
                "models": [
                    {
                        "model_id": m.model_id,
                        "size_mb": m.size_mb,
                        "is_warmed": m.is_warmed,
                        "loaded_at": m.loaded_at.isoformat(),
                        "last_used": m.last_used.isoformat(),
                    }
                    for m in self._cache.values()
                ],
            }

    def preload_models(self, model_ids: list[str]) -> None:
        """Start preloading models in background.

        Args:
            model_ids: List of model identifiers to preload
        """
        for model_id in model_ids:
            if model_id not in self._warmup_tasks:
                task = asyncio.create_task(self._preload_task(model_id))
                self._warmup_tasks[model_id] = task
                logger.info(f"Started preloading model {model_id}")

    async def _preload_task(self, model_id: str) -> None:
        """Background preload task.

        Args:
            model_id: Model identifier
        """
        try:
            await self.load_model(model_id, warmup=True)
        except Exception as e:
            logger.error(f"Failed to preload model {model_id}: {e}")
        finally:
            self._warmup_tasks.pop(model_id, None)


class MockInferenceSession:
    """Mock ONNX session for testing without onnxruntime."""

    def __init__(self, model_path: str):
        """Initialize mock session.

        Args:
            model_path: Path to model (not used)
        """
        self.model_path = model_path

    def run(self, output_names: list, input_feed: dict) -> list:
        """Mock inference run.

        Args:
            output_names: Output tensor names
            input_feed: Input tensors

        Returns:
            Mock output tensors
        """
        # Return mock output
        return [b"mock_output"]

    def get_inputs(self) -> list:
        """Get input tensors."""
        return [MockTensor("input", "float32", [1, 512])]

    def get_outputs(self) -> list:
        """Get output tensors."""
        return [MockTensor("output", "float32", [1, 256])]

    def get_providers(self) -> list:
        """Get execution providers."""
        return ["CPUExecutionProvider"]


class MockTensor:
    """Mock tensor info for testing."""

    def __init__(self, name: str, dtype: str, shape: list):
        """Initialize mock tensor.

        Args:
            name: Tensor name
            dtype: Data type
            shape: Tensor shape
        """
        self.name = name
        self.dtype = dtype
        self.shape = shape


# Global instance
_model_cache: Optional[ModelCache] = None


def get_model_cache() -> ModelCache:
    """Get global model cache instance.

    Returns:
        ModelCache singleton
    """
    global _model_cache
    if _model_cache is None:
        from ..config import get_settings

        settings = get_settings()
        _model_cache = ModelCache(
            cache_dir=settings.MODEL_CACHE_DIR,
            max_cache_size_mb=10240.0,
            max_models=10,
        )
    return _model_cache
