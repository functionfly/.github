import os
import base64


def handler(event):
    key_size = event.get("key_size", 256) if isinstance(event, dict) else 256
    format_ = event.get("format", "hex")

    VALID_SIZES = [128, 192, 256]
    try:
        key_size = int(key_size)
    except (TypeError, ValueError):
        return {"ok": False, "error": "key_size must be an integer"}
    if key_size not in VALID_SIZES:
        return {"ok": False, "error": f"key_size must be one of {VALID_SIZES}"}

    key_bytes = os.urandom(key_size // 8)
    if format_ == "base64":
        result = base64.b64encode(key_bytes).decode("utf-8")
    else:
        result = key_bytes.hex()
    return {"ok": True, "result": result, "key_size": key_size, "format": format_}
