"""Model persistence — save/load/version ML models with joblib."""

import hashlib
import json
import logging
import os
from datetime import datetime
from pathlib import Path
from typing import Any, Optional

import joblib

from ...config import settings

logger = logging.getLogger(__name__)


class ModelStore:
    """Manages model serialization, versioning, and loading."""

    def __init__(self, namespace: str):
        self._namespace = namespace
        self._base_dir = Path(settings.ml_model_dir) / namespace
        self._base_dir.mkdir(parents=True, exist_ok=True)
        self._metadata_path = self._base_dir / "metadata.json"

    def save(self, model: Any, version: Optional[str] = None) -> str:
        """Save a model to disk.

        Args:
            model: The model object to serialize
            version: Optional version string (defaults to timestamp)

        Returns:
            Version string of the saved model
        """
        if version is None:
            version = datetime.utcnow().strftime("%Y%m%d_%H%M%S")

        model_path = self._base_dir / f"model_{version}.joblib"
        joblib.dump(model, model_path)

        self._update_metadata(version, model_path)
        logger.info(f"Saved model {self._namespace}/{version}")
        return version

    def load(self, version: Optional[str] = None) -> Optional[Any]:
        """Load a model from disk.

        Args:
            version: Specific version to load (latest if None)

        Returns:
            The loaded model or None if not found
        """
        if version is None:
            version = self._get_latest_version()

        if version is None:
            return None

        model_path = self._base_dir / f"model_{version}.joblib"
        if not model_path.exists():
            logger.warning(f"Model not found: {model_path}")
            return None

        try:
            model = joblib.load(model_path)
            logger.info(f"Loaded model {self._namespace}/{version}")
            return model
        except Exception as e:
            logger.error(f"Failed to load model {self._namespace}/{version}: {e}")
            return None

    def exists(self) -> bool:
        """Check if any model version exists."""
        return self._get_latest_version() is not None

    def _get_latest_version(self) -> Optional[str]:
        """Get the latest model version from metadata."""
        meta = self._read_metadata()
        return meta.get("latest_version")

    def _update_metadata(self, version: str, model_path: Path) -> None:
        """Update the metadata file with a new version."""
        meta = self._read_metadata()
        meta["latest_version"] = version
        meta["versions"] = meta.get("versions", {})
        meta["versions"][version] = {
            "saved_at": datetime.utcnow().isoformat(),
            "size_bytes": model_path.stat().st_size,
        }
        with open(self._metadata_path, "w") as f:
            json.dump(meta, f, indent=2)

    def _read_metadata(self) -> dict:
        """Read metadata from disk."""
        if not self._metadata_path.exists():
            return {}
        try:
            with open(self._metadata_path) as f:
                return json.load(f)
        except Exception:
            return {}
