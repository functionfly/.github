"""Curated model registry for all FlyMind providers (May 2026).

Single source of truth for provider model IDs, tiers, and catalog metadata.
Used by GET /api/models/catalog and each provider's supported _models list.
"""

from __future__ import annotations

from typing import Any

# tier: frontier | fast | reasoning | code | embedding | local


def _m(
    model_id: str,
    display_name: str,
    provider: str,
    tier: str,
    cost_hint: str,
    capabilities: list[str],
    *,
    context_window: int = 0,
) -> dict[str, Any]:
    return {
        "id": model_id,
        "display_name": display_name,
        "provider": provider,
        "tier": tier,
        "cost_hint": cost_hint,
        "capabilities": capabilities,
        "context_window": context_window,
    }


CURATED_MODELS: list[dict[str, Any]] = [
    # ══════════════════════════════════════════════════════════════════════════
    # OpenRouter — unified gateway (frontier + fast + reasoning + code)
    # ══════════════════════════════════════════════════════════════════════════
    _m("anthropic/claude-opus-4.7", "Claude Opus 4.7", "openrouter", "frontier", "$$$", ["chat", "code", "tools"]),
    _m("anthropic/claude-sonnet-4.6", "Claude Sonnet 4.6", "openrouter", "frontier", "$$", ["chat", "code", "tools"]),
    _m("openai/gpt-5.5", "GPT-5.5", "openrouter", "frontier", "$$$", ["chat", "code", "tools"]),
    _m("openai/gpt-5.4", "GPT-5.4", "openrouter", "frontier", "$$", ["chat", "code", "tools"]),
    _m("google/gemini-3.1-pro", "Gemini 3.1 Pro", "openrouter", "frontier", "$$", ["chat", "code", "tools"]),
    _m("deepseek/deepseek-v4-pro", "DeepSeek V4 Pro", "openrouter", "frontier", "$", ["chat", "code", "tools"]),
    _m("moonshotai/kimi-k2.6", "Kimi K2.6", "openrouter", "frontier", "$$", ["chat", "code", "tools"]),
    _m("x-ai/grok-4", "Grok 4", "openrouter", "frontier", "$$", ["chat", "code", "tools"]),
    _m("anthropic/claude-haiku-4", "Claude Haiku 4", "openrouter", "fast", "$", ["chat", "code", "tools"]),
    _m("openai/gpt-5-mini", "GPT-5 Mini", "openrouter", "fast", "$", ["chat", "code", "tools"]),
    _m("google/gemini-2.5-flash", "Gemini 2.5 Flash", "openrouter", "fast", "$", ["chat", "code", "tools"]),
    _m("deepseek/deepseek-v4-flash", "DeepSeek V4 Flash", "openrouter", "fast", "$", ["chat", "code", "tools"]),
    _m("meta-llama/llama-3.3-70b-instruct", "Llama 3.3 70B Instruct", "openrouter", "fast", "$", ["chat", "code", "tools"]),
    _m("mistralai/mistral-large", "Mistral Large", "openrouter", "fast", "$$", ["chat", "code", "tools"]),
    _m("openai/o3", "OpenAI o3", "openrouter", "reasoning", "$$$", ["chat", "code", "tools"]),
    _m("deepseek/deepseek-r1", "DeepSeek R1", "openrouter", "reasoning", "$", ["chat", "code", "tools"]),
    _m("openai/gpt-oss-120b", "GPT-OSS 120B", "openrouter", "reasoning", "free", ["chat", "code", "tools"]),
    _m("nvidia/nemotron-3-super-120b-a12b:free", "Nemotron 3 Super 120B (free)", "openrouter", "reasoning", "free", ["chat", "code", "tools"]),
    _m("qwen/qwen3-coder", "Qwen3 Coder", "openrouter", "code", "$$", ["chat", "code", "tools"]),
    _m("openai/gpt-5-codex", "GPT-5 Codex", "openrouter", "code", "$$$", ["chat", "code", "tools"]),
    _m("meta-llama/llama-4-maverick-17b-128k-instruct", "Llama 4 Maverick 17B 128k", "openrouter", "code", "$", ["chat", "code", "tools"]),
    # ══════════════════════════════════════════════════════════════════════════
    # Anthropic — direct API
    # ══════════════════════════════════════════════════════════════════════════
    _m("claude-opus-4-20250514", "Claude Opus 4", "anthropic", "frontier", "$$$", ["chat", "code", "tools"], context_window=200000),
    _m("claude-sonnet-4-6", "Claude Sonnet 4.6", "anthropic", "frontier", "$$", ["chat", "code", "tools"], context_window=200000),
    _m("claude-3-5-sonnet-20241022", "Claude 3.5 Sonnet", "anthropic", "fast", "$$", ["chat", "code", "tools"], context_window=200000),
    _m("claude-3-5-haiku-20241022", "Claude 3.5 Haiku", "anthropic", "fast", "$", ["chat", "code", "tools"], context_window=200000),
    _m("claude-3-opus-20240229", "Claude 3 Opus", "anthropic", "frontier", "$$$", ["chat", "code", "tools"], context_window=200000),
    # ══════════════════════════════════════════════════════════════════════════
    # OpenAI — direct API (chat + embeddings)
    # ══════════════════════════════════════════════════════════════════════════
    _m("gpt-5.5", "GPT-5.5", "openai", "frontier", "$$$", ["chat", "code", "tools"]),
    _m("gpt-5.4", "GPT-5.4", "openai", "frontier", "$$", ["chat", "code", "tools"]),
    _m("gpt-5-mini", "GPT-5 Mini", "openai", "fast", "$", ["chat", "code", "tools"]),
    _m("gpt-5-codex", "GPT-5 Codex", "openai", "code", "$$$", ["chat", "code", "tools"]),
    _m("gpt-4.1", "GPT-4.1", "openai", "frontier", "$$", ["chat", "code", "tools"]),
    _m("gpt-4.1-mini", "GPT-4.1 Mini", "openai", "fast", "$", ["chat", "code", "tools"]),
    _m("gpt-4o", "GPT-4o", "openai", "fast", "$$", ["chat", "code", "tools"]),
    _m("gpt-4o-mini", "GPT-4o Mini", "openai", "fast", "$", ["chat", "code", "tools"]),
    _m("o3", "OpenAI o3", "openai", "reasoning", "$$$", ["chat", "code", "tools"]),
    _m("o1", "OpenAI o1", "openai", "reasoning", "$$$", ["chat", "code", "tools"]),
    _m("o1-mini", "OpenAI o1 Mini", "openai", "reasoning", "$$", ["chat", "code", "tools"]),
    _m("text-embedding-3-small", "Text Embedding 3 Small", "openai", "embedding", "$", ["embedding"]),
    _m("text-embedding-3-large", "Text Embedding 3 Large", "openai", "embedding", "$$", ["embedding"]),
    # ══════════════════════════════════════════════════════════════════════════
    # Groq — latency-critical (Fast profile)
    # ══════════════════════════════════════════════════════════════════════════
    _m("llama-4-scout-17b-16e-instruct", "Llama 4 Scout 17B", "groq", "fast", "$", ["chat", "code", "tools"]),
    _m("llama-4-maverick-17b-128k-instruct", "Llama 4 Maverick 17B 128k", "groq", "code", "$", ["chat", "code", "tools"]),
    _m("llama-3.3-70b-versatile", "Llama 3.3 70B Versatile", "groq", "fast", "$", ["chat", "code", "tools"]),
    _m("llama-3.1-8b-instant", "Llama 3.1 8B Instant", "groq", "fast", "$", ["chat", "code", "tools"]),
    _m("qwen-2.5-coder-32b", "Qwen 2.5 Coder 32B", "groq", "code", "$", ["chat", "code", "tools"]),
    _m("qwen-qwq-32b", "Qwen QwQ 32B", "groq", "reasoning", "$", ["chat", "code", "tools"]),
    _m("deepseek-r1-distill-llama-70b", "DeepSeek R1 Distill Llama 70B", "groq", "reasoning", "$", ["chat", "code", "tools"]),
    _m("mixtral-8x7b-32768", "Mixtral 8x7B", "groq", "fast", "$", ["chat", "code", "tools"]),
    # ══════════════════════════════════════════════════════════════════════════
    # Fireworks — structured output / function calling
    # ══════════════════════════════════════════════════════════════════════════
    _m("accounts/fireworks/models/llama-v3p1-405b-instruct", "Llama 3.1 405B", "fireworks", "frontier", "$$$", ["chat", "code", "tools"]),
    _m("accounts/fireworks/models/llama-v3p1-70b-instruct", "Llama 3.1 70B", "fireworks", "fast", "$$", ["chat", "code", "tools"]),
    _m("accounts/fireworks/models/llama-v3p3-70b-instruct", "Llama 3.3 70B", "fireworks", "fast", "$", ["chat", "code", "tools"]),
    _m("accounts/fireworks/models/llama-v3p1-8b-instruct", "Llama 3.1 8B", "fireworks", "fast", "$", ["chat", "code", "tools"]),
    _m("accounts/fireworks/models/qwen2p5-coder-32b-instruct", "Qwen 2.5 Coder 32B", "fireworks", "code", "$", ["chat", "code", "tools"]),
    _m("accounts/fireworks/models/qwen2p5-72b-instruct", "Qwen 2.5 72B", "fireworks", "fast", "$$", ["chat", "code", "tools"]),
    _m("accounts/fireworks/models/deepseek-v3", "DeepSeek V3", "fireworks", "frontier", "$$", ["chat", "code", "tools"]),
    _m("accounts/fireworks/models/deepseek-r1", "DeepSeek R1", "fireworks", "reasoning", "$$", ["chat", "code", "tools"]),
    _m("accounts/fireworks/models/mixtral-8x22b-instruct", "Mixtral 8x22B", "fireworks", "fast", "$$", ["chat", "code", "tools"]),
    # ══════════════════════════════════════════════════════════════════════════
    # DeepInfra — batch / cost-optimized
    # ══════════════════════════════════════════════════════════════════════════
    _m("meta-llama/Llama-3.3-70B-Instruct-Turbo", "Llama 3.3 70B Turbo", "deepinfra", "fast", "$", ["chat", "code", "tools"]),
    _m("meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo", "Llama 3.1 70B Turbo", "deepinfra", "fast", "$", ["chat", "code", "tools"]),
    _m("Qwen/Qwen2.5-72B-Instruct", "Qwen 2.5 72B", "deepinfra", "fast", "$$", ["chat", "code", "tools"]),
    _m("Qwen/Qwen2.5-Coder-32B-Instruct", "Qwen 2.5 Coder 32B", "deepinfra", "code", "$", ["chat", "code", "tools"]),
    _m("deepseek-ai/DeepSeek-V3", "DeepSeek V3", "deepinfra", "frontier", "$$", ["chat", "code", "tools"]),
    _m("deepseek-ai/DeepSeek-R1", "DeepSeek R1", "deepinfra", "reasoning", "$", ["chat", "code", "tools"]),
    _m("mistralai/Mixtral-8x22B-Instruct-v0.1", "Mixtral 8x22B", "deepinfra", "fast", "$$", ["chat", "code", "tools"]),
    _m("BAAI/bge-large-en-v1.5", "BGE Large EN v1.5", "deepinfra", "embedding", "$", ["embedding"]),
    _m("BAAI/bge-base-en-v1.5", "BGE Base EN v1.5", "deepinfra", "embedding", "$", ["embedding"]),
    # ══════════════════════════════════════════════════════════════════════════
    # Together AI — wide catalog / batch
    # ══════════════════════════════════════════════════════════════════════════
    _m("meta-llama/Llama-3.3-70B-Instruct-Turbo", "Llama 3.3 70B Turbo", "together", "fast", "$", ["chat", "code", "tools"]),
    _m("meta-llama/Meta-Llama-3.1-405B-Instruct-Turbo", "Llama 3.1 405B Turbo", "together", "frontier", "$$$", ["chat", "code", "tools"]),
    _m("Qwen/Qwen2.5-72B-Instruct-Turbo", "Qwen 2.5 72B Turbo", "together", "fast", "$$", ["chat", "code", "tools"]),
    _m("Qwen/Qwen2.5-Coder-32B-Instruct", "Qwen 2.5 Coder 32B", "together", "code", "$", ["chat", "code", "tools"]),
    _m("deepseek-ai/DeepSeek-V3", "DeepSeek V3", "together", "frontier", "$$", ["chat", "code", "tools"]),
    _m("deepseek-ai/DeepSeek-R1", "DeepSeek R1", "together", "reasoning", "$", ["chat", "code", "tools"]),
    _m("deepseek-ai/DeepSeek-R1-Distill-Llama-70B", "DeepSeek R1 Distill Llama 70B", "together", "reasoning", "$", ["chat", "code", "tools"]),
    _m("mistralai/Mistral-Small-24B-Instruct-2501", "Mistral Small 24B", "together", "fast", "$", ["chat", "code", "tools"]),
    _m("BAAI/bge-large-en-v1.5", "BGE Large EN v1.5", "together", "embedding", "$", ["embedding"]),
    # ══════════════════════════════════════════════════════════════════════════
    # Xiaomi MiMo — long-context reasoning, agents, flash inference
    # https://platform.xiaomimimo.com
    # ══════════════════════════════════════════════════════════════════════════
    _m(
        "mimo-v2.5-pro",
        "MiMo V2.5 Pro",
        "mimo",
        "frontier",
        "$$$",
        ["chat", "code", "tools"],
        context_window=1_000_000,
    ),
    _m(
        "mimo-v2-pro",
        "MiMo V2 Pro",
        "mimo",
        "frontier",
        "$$$",
        ["chat", "code", "tools"],
        context_window=1_000_000,
    ),
    _m(
        "mimo-v2.5",
        "MiMo V2.5",
        "mimo",
        "frontier",
        "$$",
        ["chat", "code", "tools"],
        context_window=1_000_000,
    ),
    _m(
        "mimo-v2-omni",
        "MiMo V2 Omni",
        "mimo",
        "frontier",
        "$$",
        ["chat", "code", "tools"],
        context_window=256_000,
    ),
    _m(
        "mimo-v2-flash",
        "MiMo V2 Flash",
        "mimo",
        "fast",
        "$",
        ["chat", "code", "tools"],
        context_window=256_000,
    ),
    # ══════════════════════════════════════════════════════════════════════════
    # Ollama — local dev (shown when model is installed)
    # ══════════════════════════════════════════════════════════════════════════
    _m("llama3.3:latest", "Llama 3.3", "ollama", "local", "free", ["chat", "code", "tools"]),
    _m("llama3.2:latest", "Llama 3.2", "ollama", "local", "free", ["chat", "code", "tools"]),
    _m("qwen2.5-coder:latest", "Qwen 2.5 Coder", "ollama", "local", "free", ["chat", "code", "tools"]),
    _m("deepseek-r1:latest", "DeepSeek R1", "ollama", "local", "free", ["chat", "code", "tools"]),
    _m("nomic-embed-text:latest", "Nomic Embed Text", "ollama", "local", "free", ["embedding"]),
    _m("codellama:latest", "Code Llama", "ollama", "local", "free", ["chat", "code", "tools"]),
    # ══════════════════════════════════════════════════════════════════════════
    # StepFun AI — reasoning + coding (platform.stepfun.ai)
    # ══════════════════════════════════════════════════════════════════════════
    _m("step-3.5-flash", "Step 3.5 Flash", "stepfun", "fast", "$", ["chat", "code", "tools"], context_window=32_000),
    _m("step-3", "Step 3", "stepfun", "reasoning", "$$", ["chat", "code", "tools"], context_window=32_000),
    _m("step-2-mini", "Step 2 Mini", "stepfun", "fast", "$$", ["chat", "code", "tools"], context_window=32_000),
    _m("step-2-16k", "Step 2 16k", "stepfun", "frontier", "$$$", ["chat", "code", "tools"], context_window=16_000),
    _m("step-1-8k", "Step 1 8k", "stepfun", "fast", "$", ["chat", "code", "tools"], context_window=8_000),
    _m("step-1-32k", "Step 1 32k", "stepfun", "fast", "$$", ["chat", "code", "tools"], context_window=32_000),
    _m("step-1o-turbo-vision", "Step 1o Turbo Vision", "stepfun", "frontier", "$$$", ["chat", "code", "tools", "vision"]),
    _m("step-1o-vision-32k", "Step 1o Vision 32k", "stepfun", "frontier", "$$$", ["chat", "code", "tools", "vision"], context_window=32_000),
    _m("step-1v-8k", "Step 1v 8k", "stepfun", "fast", "$$", ["chat", "code", "tools", "vision"], context_window=8_000),
    _m("step-1v-32k", "Step 1v 32k", "stepfun", "fast", "$$$", ["chat", "code", "tools", "vision"], context_window=32_000),
    _m("step-r1-v-mini", "Step R1 Vision Mini", "stepfun", "reasoning", "$$", ["chat", "code", "tools", "vision"]),
]

# Providers where catalog entries require the model to exist on the provider
_DYNAMIC_PROVIDERS = frozenset({"ollama"})


def model_ids_for_provider(provider: str) -> list[str]:
    """Return deduplicated model IDs for a provider's _models list."""
    seen: list[str] = []
    for entry in CURATED_MODELS:
        if entry["provider"] != provider:
            continue
        model_id = entry["id"]
        if model_id not in seen:
            seen.append(model_id)
    return seen


def build_model_catalog(provider_manager) -> list[dict[str, Any]]:
    """Return curated catalog for registered providers, plus uncatalogued extras."""
    provider_info = {info.name: info for info in provider_manager.list_providers()}
    registered = set(provider_info.keys())
    available = {name for name, info in provider_info.items() if info.available}

    installed: dict[str, set[str]] = {
        name: set(info.models) for name, info in provider_info.items() if info.available
    }

    deduped: dict[str, dict[str, Any]] = {}

    for entry in CURATED_MODELS:
        provider = entry["provider"]
        if provider not in registered:
            continue
        elif provider in _DYNAMIC_PROVIDERS:
            if provider not in available:
                continue
            model_id = entry["id"]
            names = installed.get(provider, set())
            base = model_id.split(":")[0]
            if model_id not in names and not any(n.split(":")[0] == base for n in names):
                continue

        key = f"{provider}:{entry['id']}"
        item = dict(entry)
        item["provider_available"] = provider in available
        deduped[key] = item

    for info in provider_manager.list_providers():
        if not info.available:
            continue
        for model_id in info.models:
            key = f"{info.name}:{model_id}"
            if key in deduped:
                continue
            model_lower = str(model_id).lower()
            if "embed" in model_lower:
                capabilities = ["embedding"]
                tier = "embedding"
            else:
                capabilities = ["chat", "code"]
                if info.supports_streaming:
                    capabilities.append("tools")
                tier = "local" if info.name == "ollama" else "balanced"

            display_name = str(model_id)
            if "/" in display_name:
                display_name = display_name.rsplit("/", 1)[-1]
            display_name = display_name.replace("-", " ").replace("_", " ")

            deduped[key] = {
                "id": str(model_id),
                "display_name": display_name,
                "provider": info.name,
                "tier": tier,
                "context_window": 0,
                "cost_hint": "free" if info.name == "ollama" else "varies",
                "capabilities": sorted(set(capabilities)),
            }

    return list(deduped.values())
