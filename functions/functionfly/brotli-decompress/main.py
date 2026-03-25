import base64


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    encoding = event.get("encoding", "utf-8")

    if not data:
        return {"ok": False, "error": "data is required"}

    try:
        import brotli
    except ImportError:
        return {"ok": False, "error": "brotli library is not installed. Install with: pip install brotli"}

    try:
        compressed = base64.b64decode(str(data))
        decompressed = brotli.decompress(compressed)
        try:
            result = decompressed.decode(encoding)
            is_text = True
        except UnicodeDecodeError:
            result = base64.b64encode(decompressed).decode("utf-8")
            is_text = False
        return {"ok": True, "result": result, "is_text": is_text, "decompressed_size": len(decompressed)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
