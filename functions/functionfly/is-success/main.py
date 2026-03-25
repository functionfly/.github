def handler(event):
    status_code = event.get("status_code") if isinstance(event, dict) else None

    if status_code is None:
        return {"ok": False, "error": "status_code is required"}
    try:
        code = int(status_code)
    except (TypeError, ValueError):
        return {"ok": False, "error": "status_code must be an integer"}

    result = 200 <= code <= 299
    return {"ok": True, "status_code": code, "result": result}
