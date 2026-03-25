from urllib.parse import quote, urlencode


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    mode = event.get("mode", "component")
    safe = event.get("safe", "")

    if data is None:
        return {"ok": False, "error": "data is required"}

    try:
        if isinstance(data, dict):
            result = urlencode(data)
        else:
            val = str(data)
            if mode == "full":
                result = quote(val, safe=safe or "/:@!$&'()*+,;=")
            else:
                result = quote(val, safe=safe or "")
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
