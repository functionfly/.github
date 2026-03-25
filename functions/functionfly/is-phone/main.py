import re

# E.164 and common international formats
PHONE_RE = re.compile(r'^\+?[1-9]\d{6,14}$')
LOOSE_RE = re.compile(r'^[\+\-\.\(\)\s\d]{7,20}$')


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    strict = event.get("strict", False)

    if value is None:
        return {"ok": False, "error": "value is required"}

    val = str(value).strip()
    # Normalize: remove spaces, dashes, dots, parens
    normalized = re.sub(r'[\s\-\.\(\)]', '', val)

    if strict:
        result = bool(PHONE_RE.match(normalized))
    else:
        result = bool(LOOSE_RE.match(val)) and len(normalized) >= 7

    return {"ok": True, "value": value, "result": result, "normalized": normalized}
