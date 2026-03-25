import os
import base64


def handler(event):
    size = event.get("size", 16) if isinstance(event, dict) else 16
    format_ = event.get("format", "hex")

    try:
        size = int(size)
        if size < 8 or size > 64:
            return {"ok": False, "error": "size must be between 8 and 64 bytes"}
    except (TypeError, ValueError):
        return {"ok": False, "error": "size must be an integer"}

    salt_bytes = os.urandom(size)
    if format_ == "base64":
        result = base64.b64encode(salt_bytes).decode("utf-8")
    else:
        result = salt_bytes.hex()
    return {"ok": True, "result": result, "size": size, "format": format_}
