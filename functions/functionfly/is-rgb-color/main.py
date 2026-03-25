import re

RGB_RE = re.compile(
    r'^rgba?\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*(?:,\s*([01]?\.?\d*))?\s*\)$',
    re.IGNORECASE
)


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}
    val = str(value).strip()
    m = RGB_RE.match(val)
    if not m:
        return {"ok": True, "value": value, "result": False}

    r, g, b = int(m.group(1)), int(m.group(2)), int(m.group(3))
    if not (0 <= r <= 255 and 0 <= g <= 255 and 0 <= b <= 255):
        return {"ok": True, "value": value, "result": False}

    alpha = None
    if m.group(4) is not None:
        try:
            alpha = float(m.group(4))
            if not (0 <= alpha <= 1):
                return {"ok": True, "value": value, "result": False}
        except ValueError:
            return {"ok": True, "value": value, "result": False}

    return {"ok": True, "value": value, "result": True, "r": r, "g": g, "b": b, "alpha": alpha}
