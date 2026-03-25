import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    level = event.get("level", 3)

    if data is None:
        return {"ok": False, "error": "data is required"}

    try:
        import zstandard as zstd
    except ImportError:
        return {"ok": False, "error": "zstandard library is not installed. Install with: pip install zstandard"}

    try:
        if isinstance(data, (dict, list)):
            import json
            raw = json.dumps(data).encode("utf-8")
        else:
            raw = str(data).encode("utf-8")

        cctx = zstd.ZstdCompressor(level=int(level))
        compressed = cctx.compress(raw)
        encoded = base64.b64encode(compressed).decode("utf-8")
        return {
            "ok": True, "result": encoded,
            "original_size": len(raw), "compressed_size": len(compressed),
            "ratio": round(len(compressed) / len(raw), 4) if raw else 0,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
