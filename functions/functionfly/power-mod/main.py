def handler(event):
    base = event.get("base") if isinstance(event, dict) else None
    exp = event.get("exp") if isinstance(event, dict) else None
    mod = event.get("mod") if isinstance(event, dict) else None
    if base is None or exp is None or mod is None:
        return {"ok": False, "error": "base, exp, and mod are required"}
    try:
        base, exp, mod = int(base), int(exp), int(mod)
        if mod <= 0:
            return {"ok": False, "error": "mod must be positive"}
        return {"ok": True, "result": pow(base, exp, mod)}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
