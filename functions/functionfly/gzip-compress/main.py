import gzip
import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    level = event.get("level", 9)

    if data is None:
        return {"ok": False, "error": "data is required"}

    try:
        level = int(level)
        if not 0 <= level <= 9:
            return {"ok": False, "error": "level must be between 0 and 9"}
    except (TypeError, ValueError):
        return {"ok": False, "error": "level must be an integer 0-9"}

    try:
        if isinstance(data, str):
            raw = data.encode("utf-8")
        elif isinstance(data, list):
            import json
            raw = json.dumps(data).encode("utf-8")
        elif isinstance(data, dict):
            import json
            raw = json.dumps(data).encode("utf-8")
        else:
            raw = str(data).encode("utf-8")

        compressed = gzip.compress(raw, compresslevel=level)
        encoded = base64.b64encode(compressed).decode("utf-8")

        return {
            "ok": True,
            "result": encoded,
            "original_size": len(raw),
            "compressed_size": len(compressed),
            "ratio": round(len(compressed) / len(raw), 4) if raw else 0,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
