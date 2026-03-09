import base64


def handler(event):
    """
    Decode a Base64 string to plain text.

    Input:
        - encoded: Base64 string to decode (required)
        - url_safe: If true, input is URL-safe alphabet (default: false)

    Returns:
        - ok: True on success
        - text: Decoded string (UTF-8)
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        encoded = event.get("encoded", event.get("data", ""))
        url_safe = event.get("url_safe", False)
    else:
        encoded = str(event) if event is not None else ""
        url_safe = False

    if not encoded or not str(encoded).strip():
        return {"ok": False, "error": "Input 'encoded' is required and cannot be empty"}

    try:
        raw = encoded if isinstance(encoded, bytes) else encoded.encode("ascii")
        if url_safe:
            # Restore padding if stripped
            pad = 4 - (len(raw) % 4)
            if pad != 4:
                raw += b"=" * pad
            data = base64.urlsafe_b64decode(raw)
        else:
            data = base64.b64decode(raw)
        text = data.decode("utf-8")
        return {"ok": True, "text": text}
    except Exception as e:
        return {"ok": False, "error": str(e)}
