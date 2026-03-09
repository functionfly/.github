def handler(event):
    """
    Decode a hexadecimal string to UTF-8 text.

    Input:
        - encoded: Hex string to decode (required)
        - errors: How to handle decode errors: strict, replace, ignore (default: replace)

    Returns:
        - ok: True on success
        - decoded: Decoded string
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        encoded = event.get("encoded", event.get("data", event.get("hex", "")))
        errors = event.get("errors", "replace")
    else:
        encoded = event
        errors = "replace"

    if encoded is None or (isinstance(encoded, str) and not encoded.strip()):
        return {"ok": False, "error": "Input 'encoded' is required"}

    encoded = str(encoded).strip().replace(" ", "").replace("0x", "")
    if len(encoded) % 2 != 0:
        return {"ok": False, "error": "Hex string must have even length"}

    try:
        raw = bytes.fromhex(encoded)
        decoded = raw.decode("utf-8", errors=errors)
        return {"ok": True, "decoded": decoded}
    except ValueError as e:
        return {"ok": False, "error": f"Invalid hex: {e}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
