import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None

    if data is None:
        return {"ok": False, "error": "data is required"}

    try:
        import lz4.frame
    except ImportError:
        return {"ok": False, "error": "lz4 library is not installed. Install with: pip install lz4"}

    try:
        if isinstance(data, (dict, list)):
            import json
            raw = json.dumps(data).encode("utf-8")
        else:
            raw = str(data).encode("utf-8")

        compressed = lz4.frame.compress(raw)
        encoded = base64.b64encode(compressed).decode("utf-8")
        return {
            "ok": True, "result": encoded,
            "original_size": len(raw), "compressed_size": len(compressed),
            "ratio": round(len(compressed) / len(raw), 4) if raw else 0,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
