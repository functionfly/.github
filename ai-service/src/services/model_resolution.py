"""Centralized model selection helpers for AI routes."""

from __future__ import annotations

from typing import Any, Mapping, Optional, Tuple

from ..providers.manager import get_provider_manager


def resolve_model(
    feature: str,
    headers: Mapping[str, str] | None = None,
    body: Mapping[str, Any] | None = None,
) -> Tuple[str, str]:
    """Resolve provider+model using request override -> header -> fallback."""
    headers = headers or {}
    body = body or {}

    provider = (
        body.get("provider")
        or headers.get("x-preferred-provider")
        or headers.get("X-Preferred-Provider")
    )
    model_id = (
        body.get("model")
        or body.get("model_id")
        or headers.get("x-preferred-model")
        or headers.get("X-Preferred-Model")
    )

    if provider and model_id:
        return str(provider), str(model_id)

    provider_manager = get_provider_manager()
    selected_provider = provider_manager.get_provider_for_chat(provider)
    selected_provider_name = provider or selected_provider.name
    selected_model = _default_model_for_provider(selected_provider_name, feature)
    return selected_provider_name, selected_model


def _default_model_for_provider(provider_name: str, feature: str) -> str:
    provider_manager = get_provider_manager()
    provider = provider_manager.get_provider_for_chat(provider_name)
    supported = getattr(provider, "supported_models", {}) or {}
    if feature == "frg":
        preferred = ("code", "chat", "default")
    else:
        preferred = ("chat", "code", "default")
    for key in preferred:
        if key in supported and supported[key]:
            return str(supported[key])
    # Safe fallback used throughout existing code paths.
    return "gpt-4o-mini"
