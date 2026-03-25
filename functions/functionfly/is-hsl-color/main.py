import re

HSL_RE = re.compile(
    r'^hsla?\(\s*(\d{1,3})\s*,\s*(\d{1,3})%\s*,\s*(\d{1,3})%\s*(?:,\s*([01]?\.?\d*))?\s*\)$',
    re.IGNORECASE
)


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}
    val = str(value).strip()
    m = HSL_RE.match(val)
    if not m:
        return {"ok": True, "value": value, "result": False}

    h, s, l = int(m.group(1)), int(m.group(2)), int(m.group(3))
    if not (0 <= h <= 360 and 0 <= s <= 100 and 0 <= l <= 100):
        return {"ok": True, "value": value, "result": False}

    alpha = None
    if m.group(4) is not None:
        try:
            alpha = float(m.group(4))
            if not (0 <= alpha <= 1):
                return {"ok": True, "value": value, "result": False}
        except ValueError:
            return {"ok": True, "value": value, "result": False}

    return {"ok": True, "value": value, "result": True, "h": h, "s": s, "l": l, "alpha": alpha}
