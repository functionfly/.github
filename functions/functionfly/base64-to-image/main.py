import base64
import re


def handler(event):
    if isinstance(event, dict):
        data = event.get("data", event.get("image", ""))
    else:
        data = ""

    if not data:
        return {"ok": False, "error": "Input 'data' is required"}

    s = str(data).strip()
    m = re.match(r"^data:image/[^;]+;base64,(.+)$", s, re.IGNORECASE)
    if m:
        s = m.group(1)
    try:
        raw = base64.b64decode(s, validate=True)
    except Exception as e:
        return {"ok": False, "error": f"Invalid base64: {e}"}

    encoded = base64.b64encode(raw).decode("ascii")
    return {"ok": True, "image_base64": encoded, "bytes_count": len(raw)}

