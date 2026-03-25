def handler(event):
    # In Python/JSON context, "undefined" maps to a missing key or null
    has_value = "value" in event
    value = event.get("value")
    # Treat missing key OR null as undefined
    result = not has_value or value is None
    return {"ok": True, "value": value, "result": result, "key_present": has_value}
