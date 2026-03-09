import hashlib


def handler(event):
    """
    Generate MD5 hash of a string or raw bytes (hex or base64).

    Input:
        - data: String or base64-encoded data to hash (required)
        - encoding: Output encoding: hex or base64 (default: hex)

    Returns:
        - ok: True on success
        - hash: Hash string
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        data = event.get("data", event.get("text", event.get("input", "")))
        encoding = event.get("encoding", "hex")
    else:
        data = event
        encoding = "hex"

    if data is None:
        return {"ok": False, "error": "Input 'data' is required"}

    try:
        if isinstance(data, str):
            raw = data.encode("utf-8")
        else:
            raw = bytes(data) if data is not None else b""
        digest = hashlib.md5(raw).digest()
        if encoding == "base64":
            import base64
            out = base64.b64encode(digest).decode("ascii")
        else:
            out = hashlib.md5(raw).hexdigest()
        return {"ok": True, "hash": out}
    except Exception as e:
        return {"ok": False, "error": str(e)}
