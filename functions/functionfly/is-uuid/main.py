import re

UUID_RE = re.compile(
    r'^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$',
    re.IGNORECASE
)
UUID_ANY_RE = re.compile(
    r'^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$',
    re.IGNORECASE
)


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    version = event.get("version")
    strict = event.get("strict", True)

    if value is None:
        return {"ok": False, "error": "value is required"}

    val = str(value).strip()

    if strict:
        match = UUID_RE.match(val)
    else:
        match = UUID_ANY_RE.match(val.replace('-', '') if '-' not in val else val)

    if not match:
        return {"ok": True, "value": value, "result": False, "uuid_version": None}

    # Extract version from character at position 14 (after removing dashes)
    clean = val.replace('-', '')
    detected_version = int(clean[12], 16) if clean[12].isdigit() else None

    if version and detected_version != version:
        return {"ok": True, "value": value, "result": False, "uuid_version": detected_version}

    return {"ok": True, "value": value, "result": True, "uuid_version": detected_version}
