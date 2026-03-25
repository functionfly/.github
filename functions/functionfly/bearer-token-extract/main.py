def handler(event):
    value = event.get("value") if isinstance(event, dict) else None

    if not value:
        return {"ok": False, "error": "value is required"}
    if not isinstance(value, str):
        return {"ok": False, "error": "value must be a string"}

    stripped = value.strip()
    lower = stripped.lower()

    if lower.startswith("bearer "):
        token = stripped[7:].strip()
        if not token:
            return {"ok": False, "error": "Bearer token is empty"}
        return {"ok": True, "token": token}

    return {"ok": False, "error": "Authorization header does not use Bearer scheme"}
