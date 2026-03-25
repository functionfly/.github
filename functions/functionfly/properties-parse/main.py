def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (Java properties string)"}
    try:
        result = {}
        for line in str(data).splitlines():
            line = line.strip()
            if not line or line.startswith("#") or line.startswith("!"):
                continue
            # Handle continuation lines (simplified)
            for sep in ("=", ":", " "):
                if sep in line:
                    key, _, value = line.partition(sep)
                    result[key.strip()] = value.strip()
                    break
        return {"ok": True, "result": result, "count": len(result)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
