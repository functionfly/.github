def deep_merge(base, override):
    """Deep merge two dicts, override takes precedence."""
    result = dict(base)
    for key, value in override.items():
        if key in result and isinstance(result[key], dict) and isinstance(value, dict):
            result[key] = deep_merge(result[key], value)
        else:
            result[key] = value
    return result


def handler(event):
    """Deep merge multiple configuration objects."""
    try:
        configs = event.get("configs")
        if not configs:
            return {"ok": False, "error": "configs is required"}

        strategy = event.get("strategy", "deep")

        if strategy == "shallow":
            result = {}
            for cfg in configs:
                result.update(cfg)
        elif strategy == "replace":
            result = configs[-1] if configs else {}
        else:  # deep
            result = {}
            for cfg in configs:
                result = deep_merge(result, cfg)

        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
