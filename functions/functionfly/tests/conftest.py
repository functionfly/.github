"""
Pytest fixtures for functionfly stdlib tests.
Discovers all functions under functions/functionfly and loads manifest + handler.
"""
from __future__ import annotations

import importlib.util
import json
import re
import sys
from pathlib import Path

import pytest

# Repo root: functions/functionfly (this package); parent is functions/
FUNCTIONS_DIR = Path(__file__).resolve().parent.parent
assert FUNCTIONS_DIR.name == "functionfly", "conftest should live under functionfly/"


def _strip_jsonc(raw: str) -> str:
    """Remove // and /* */ comments for JSONC."""
    # Remove single-line // comments
    raw = re.sub(r"//[^\n]*", "", raw)
    # Remove multi-line /* */ comments
    raw = re.sub(r"/\*.*?\*/", "", raw, flags=re.DOTALL)
    return raw.strip()


def load_manifest(function_dir: Path) -> dict | None:
    """Load functionfly.jsonc from function dir. Returns None if missing/invalid."""
    path = function_dir / "functionfly.jsonc"
    if not path.is_file():
        return None
    try:
        raw = path.read_text(encoding="utf-8")
        return json.loads(_strip_jsonc(raw))
    except (json.JSONDecodeError, OSError):
        return None


def load_handler(function_dir: Path, name: str):
    """Load main.handler from function dir. Returns (handler_callable, error_message)."""
    main_py = function_dir / "main.py"
    if not main_py.is_file():
        return None, "main.py not found"
    try:
        spec = importlib.util.spec_from_file_location(
            f"functionfly_{name}".replace("-", "_"), main_py
        )
        mod = importlib.util.module_from_spec(spec)
        sys.modules[spec.name] = mod
        spec.loader.exec_module(mod)
        if not hasattr(mod, "handler"):
            return None, "handler not found"
        return getattr(mod, "handler"), None
    except Exception as e:
        return None, str(e)


def discover_functions():
    """Yield (name, function_dir, manifest, handler) for each function."""
    for child in sorted(FUNCTIONS_DIR.iterdir()):
        if not child.is_dir():
            continue
        manifest_path = child / "functionfly.jsonc"
        if not manifest_path.is_file():
            continue
        name = child.name
        manifest = load_manifest(child)
        if manifest is None:
            continue
        handler, err = load_handler(child, name)
        if err:
            continue
        yield (name, child, manifest, handler)


def get_all_function_names():
    """List of all function names (for parametrization)."""
    return [name for name, _, _, _ in discover_functions()]


@pytest.fixture(scope="session")
def function_list():
    """All (name, manifest, handler) for parametrized tests."""
    return list(discover_functions())


def _get_function_item(name: str):
    """Get (manifest, handler) for a given function name."""
    for n, _dir, manifest, handler in discover_functions():
        if n == name:
            return manifest, handler
    return None, None


@pytest.fixture
def get_function():
    """Fixture that returns a callable: get_function(name) -> (manifest, handler)."""
    return _get_function_item
