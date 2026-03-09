import base64
import re


def handler(event):
    if isinstance(event, dict):
        image = event.get("image", event.get("data", ""))
        data_url = event.get("data_url", False)
    else:
        image = ""
        data_url = False

    if not image:
        return {"ok": False, "error": "Input 'image' is required"}

    s = str(image).strip()
    # Strip data URL prefix if present
    m = re.match(r"^data:image/[^;]+;base64,(.+)$", s, re.IGNORECASE)
    if m:
        s = m.group(1)
    try:
        raw = base64.b64decode(s, validate=True)
        encoded = base64.b64encode(raw).decode("ascii")
    except Exception as e:
        return {"ok": False, "error": f"Invalid base64 image: {e}"}

    if data_url:
        encoded = f"data:image/png;base64,{encoded}"
    return {"ok": True, "encoded": encoded}

