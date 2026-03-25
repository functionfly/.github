import re


def handler(event):
    etag = event.get("etag") if isinstance(event, dict) else None

    if etag is None:
        return {"ok": False, "error": "etag is required"}
    if not isinstance(etag, str):
        return {"ok": False, "error": "etag must be a string"}

    etag = etag.strip()

    # Weak ETag: W/"value"
    weak_match = re.match(r'^W/"([^"]*)"$', etag)
    if weak_match:
        return {"ok": True, "valid": True, "etag": etag, "weak": True, "value": weak_match.group(1)}

    # Strong ETag: "value"
    strong_match = re.match(r'^"([^"]*)"$', etag)
    if strong_match:
        return {"ok": True, "valid": True, "etag": etag, "weak": False, "value": strong_match.group(1)}

    return {"ok": True, "valid": False, "etag": etag, "error": "Invalid ETag format"}
