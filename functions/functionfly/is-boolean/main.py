BOOL_TRUE = {"true", "1", "yes", "on"}
BOOL_FALSE = {"false", "0", "no", "off"}


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    strict = event.get("strict", False)

    if value is None:
        return {"ok": False, "error": "value is required"}

    if isinstance(value, bool):
        return {"ok": True, "value": value, "result": True, "parsed": value}

    if strict:
        return {"ok": True, "value": value, "result": isinstance(value, bool), "parsed": value if isinstance(value, bool) else None}

    val_str = str(value).strip().lower()
    if val_str in BOOL_TRUE:
        return {"ok": True, "value": value, "result": True, "parsed": True}
    if val_str in BOOL_FALSE:
        return {"ok": True, "value": value, "result": True, "parsed": False}

    return {"ok": True, "value": value, "result": False, "parsed": None}
