import urllib.parse


def handler(event):
    """
    URL-decode a percent-encoded string.

    Input:
        - encoded: URL-encoded string to decode (required)
        - errors: How to handle decode errors: strict, replace, ignore (default: replace)

    Returns:
        - ok: True on success
        - decoded: Decoded string
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        encoded = event.get("encoded", event.get("data", event.get("string", "")))
        errors = event.get("errors", "replace")
    else:
        encoded = str(event) if event is not None else ""
        errors = "replace"

    if encoded is None or (isinstance(encoded, str) and encoded == ""):
        return {"ok": False, "error": "Input 'encoded' is required"}

    try:
        decoded = urllib.parse.unquote(str(encoded), errors=errors)
        return {"ok": True, "decoded": decoded}
    except Exception as e:
        return {"ok": False, "error": str(e)}
